package documents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Objects must verify the stored bytes before Ensure returns. Calls are
// idempotent; objects are addressed by SHA-256, never by a mutable source URL.
type Objects interface {
	Ensure(context.Context, string, []byte) error
	Read(context.Context, string, int64) ([]byte, error)
}

type Store struct {
	pool    *pgxpool.Pool
	objects Objects
}

func New(pool *pgxpool.Pool, objects Objects) *Store { return &Store{pool: pool, objects: objects} }

// Stage commits a replayable acquisition before any object write. The caller
// must reuse the same ID/input after an uncertain database commit.
func (s *Store) Stage(ctx context.Context, input Acquisition) error {
	a, _, hash, _, err := prepare(input)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(a)
	if err != nil {
		return ErrInvalid
	}
	requestHash := digest(payload)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Serialize ID collisions before creating documents or checking policies.
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,730023))", a.ID); err != nil {
		return err
	}
	var old string
	err = tx.QueryRow(ctx, "SELECT request_hash FROM retained_acquisitions WHERE id=$1", a.ID).Scan(&old)
	if err == nil {
		if old != requestHash {
			return ErrConflict
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	for _, r := range a.Resources {
		var raw []byte
		if err = tx.QueryRow(ctx, "SELECT body FROM registry_configurations WHERE source_id=$1 AND revision=$2", r.SourceID, r.Configuration).Scan(&raw); err != nil {
			return err
		}
		var c registry.Configuration
		if json.Unmarshal(raw, &c) != nil {
			return ErrPolicy
		}
		if r.Missing == "" && (!c.Policy.CollectionPermitted || !c.Policy.RetentionPermitted || c.Policy.Evidence == nil) {
			return ErrPolicy
		}
	}
	var documentID int64
	if err = tx.QueryRow(ctx, `INSERT INTO retained_documents(source_id,official_url) VALUES ($1,$2) ON CONFLICT(source_id,official_url) DO UPDATE SET official_url=EXCLUDED.official_url RETURNING id`, a.SourceID, a.URL).Scan(&documentID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO retained_acquisitions(id,document_id,source_id,configuration,request_hash,content_hash,payload) VALUES ($1,$2,$3,$4,$5,$6,$7)`, a.ID, documentID, a.SourceID, a.Configuration, requestHash, hash, payload)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) Retain(ctx context.Context, a Acquisition) (Version, error) {
	if err := s.Stage(ctx, a); err != nil {
		return Version{}, err
	}
	return s.Recover(ctx, a.ID)
}

// Pending lists durable acquisitions, including failures of an earlier process.
// Workers (task 2.4) can call Recover; no upstream re-fetch is necessary.
func (s *Store) Pending(ctx context.Context, limit int) ([]string, error) {
	if limit < 1 || limit > 1000 {
		return nil, ErrInvalid
	}
	rows, err := s.pool.Query(ctx, "SELECT id FROM retained_acquisitions WHERE version_id IS NULL ORDER BY acquired_at,id LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Recover can be called concurrently or after a process/database/object-store
// restart. A failed transfer leaves all staged bytes available for retry.
func (s *Store) Recover(ctx context.Context, id string) (Version, error) {
	var payload []byte
	var versionID *int64
	err := s.pool.QueryRow(ctx, "SELECT payload,version_id FROM retained_acquisitions WHERE id=$1", id).Scan(&payload, &versionID)
	if err != nil {
		return Version{}, err
	}
	if versionID != nil {
		return s.Version(ctx, *versionID)
	}
	var a Acquisition
	if json.Unmarshal(payload, &a) != nil {
		return Version{}, ErrInvalid
	}
	_, refs, hash, complete, err := prepare(a)
	if err != nil {
		return Version{}, err
	}
	// A response differing only in a known transport-only marker shares the
	// retained version. Keep the first raw original and avoid storing an
	// unreferenced object for each later equivalent response.
	var equivalentID int64
	err = s.pool.QueryRow(ctx, `UPDATE retained_acquisitions a SET version_id=v.id,payload=NULL
 FROM retained_versions v WHERE a.id=$1 AND a.version_id IS NULL
 AND v.document_id=a.document_id AND v.content_hash=$2 RETURNING v.id`, id, hash).Scan(&equivalentID)
	if err == nil {
		return s.Version(ctx, equivalentID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Version{}, err
	}
	for _, r := range a.Resources {
		if r.Missing == "" {
			if err = s.objects.Ensure(ctx, digest(r.Bytes), r.Bytes); err != nil {
				return Version{}, ErrStorage
			}
		}
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Version{}, err
	}
	defer tx.Rollback(ctx)
	var documentID int64
	if err = tx.QueryRow(ctx, "SELECT document_id FROM retained_acquisitions WHERE id=$1", id).Scan(&documentID); err != nil {
		return Version{}, err
	}
	// Serializes versions for one URL and finalization of concurrent retries.
	if _, err = tx.Exec(ctx, "SELECT id FROM retained_documents WHERE id=$1 FOR UPDATE", documentID); err != nil {
		return Version{}, err
	}
	if err = tx.QueryRow(ctx, "SELECT version_id FROM retained_acquisitions WHERE id=$1", id).Scan(&versionID); err != nil {
		return Version{}, err
	}
	if versionID != nil {
		if err = tx.Commit(ctx); err != nil {
			return Version{}, err
		}
		return s.Version(ctx, *versionID)
	}
	var vID int64
	err = tx.QueryRow(ctx, `SELECT id FROM retained_versions WHERE document_id=$1 AND content_hash=$2`, documentID, hash).Scan(&vID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, `INSERT INTO retained_versions(document_id,content_hash,issuer_id,first_acquired_at,complete,metadata)
   SELECT $1,$2,$3,min(acquired_at),$4,$5 FROM retained_acquisitions WHERE document_id=$1 AND content_hash=$2 RETURNING id`, documentID, hash, a.IssuerID, complete, a.Metadata).Scan(&vID)
		if err != nil {
			return Version{}, err
		}
		for _, r := range refs {
			var objectHash *string
			if r.Hash != "" {
				objectHash = &r.Hash
				if _, err = tx.Exec(ctx, "INSERT INTO retained_objects(hash,object_key,byte_size) VALUES($1,$2,$3) ON CONFLICT(hash) DO NOTHING", r.Hash, objectKey(r.Hash), r.Size); err != nil {
					return Version{}, err
				}
			}
			if _, err = tx.Exec(ctx, `INSERT INTO retained_resources(version_id,url,role,required,source_id,configuration,media_type,object_hash,missing) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, vID, r.URL, r.Role, r.Required, r.SourceID, r.Configuration, r.MediaType, objectHash, r.Missing); err != nil {
				return Version{}, err
			}
		}
	} else if err != nil {
		return Version{}, err
	}
	if _, err = tx.Exec(ctx, "UPDATE retained_acquisitions SET version_id=$2,payload=NULL WHERE id=$1", id, vID); err != nil {
		return Version{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Version{}, err
	}
	return s.Version(ctx, vID)
}

func (s *Store) Version(ctx context.Context, id int64) (Version, error) {
	var v Version
	err := s.pool.QueryRow(ctx, "SELECT id,document_id,content_hash,first_acquired_at,complete,metadata,issuer_id FROM retained_versions WHERE id=$1", id).Scan(&v.ID, &v.DocumentID, &v.Hash, &v.FirstAcquiredAt, &v.Complete, &v.Metadata, &v.IssuerID)
	if err != nil {
		return v, err
	}
	rows, err := s.pool.Query(ctx, `SELECT r.url,r.role,r.required,r.source_id,r.configuration,r.media_type,COALESCE(r.object_hash,''),COALESCE(o.byte_size,0),r.missing,c.body FROM retained_resources r LEFT JOIN retained_objects o ON o.hash=r.object_hash JOIN registry_sources src ON src.id=r.source_id JOIN registry_configurations c ON c.source_id=r.source_id AND c.revision=COALESCE(src.active_revision,r.configuration) WHERE version_id=$1 ORDER BY r.url`, id)
	if err != nil {
		return v, err
	}
	defer rows.Close()
	for rows.Next() {
		var r Reference
		var configuration []byte
		if err = rows.Scan(&r.URL, &r.Role, &r.Required, &r.SourceID, &r.Configuration, &r.MediaType, &r.Hash, &r.Size, &r.Missing, &configuration); err != nil {
			return v, err
		}
		var cfg registry.Configuration
		if json.Unmarshal(configuration, &cfg) != nil {
			return v, ErrInvalid
		}
		decision := cfg.InferenceEligibility.Decide(r.URL, r.Hash, r.Role, r.Missing)
		r.Inference = &decision
		v.Resources = append(v.Resources, r)
	}
	return v, rows.Err()
}

// Read is internal-only; public copy authorization is a separate application
// concern. Both length and hash are checked on every read.
func (s *Store) Read(ctx context.Context, versionID int64, resourceURL string) ([]byte, error) {
	var hash string
	var size int64
	err := s.pool.QueryRow(ctx, `SELECT o.hash,o.byte_size FROM retained_resources r JOIN retained_objects o ON o.hash=r.object_hash WHERE version_id=$1 AND url=$2`, versionID, resourceURL).Scan(&hash, &size)
	if err != nil {
		return nil, err
	}
	b, err := s.objects.Read(ctx, hash, size)
	if err != nil || int64(len(b)) != size || digest(b) != hash {
		return nil, fmt.Errorf("read retained evidence: %w", ErrStorage)
	}
	return b, nil
}
