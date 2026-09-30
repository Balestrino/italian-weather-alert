package ocr

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

// ArtifactIdentity includes the exact provider request, not just a page hash.
// Renderer and configuration identities must change when their behavior changes.
type ArtifactIdentity struct {
	Scope, Model, Configuration, Renderer, InputSHA256, RequestSHA256 string
}

var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func (i ArtifactIdentity) Key() (string, error) {
	if i.Scope == "" || i.Model == "" || i.Configuration == "" || i.Renderer == "" || !hashPattern.MatchString(i.InputSHA256) || !hashPattern.MatchString(i.RequestSHA256) {
		return "", ErrInvalid
	}
	encoded, _ := json.Marshal(i)
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:]), nil
}

type Artifact struct {
	Key, Token   string
	Generation   int64
	LeaseExpires time.Time
	Ready, Owned bool
	Page         PageResult
}

// ClaimArtifact serializes equal work across processes. Expired owners are
// fenced by both token and generation. Ready results never acquire ownership.
func (s *Store) ClaimArtifact(ctx context.Context, identity ArtifactIdentity, now time.Time, lease time.Duration) (Artifact, error) {
	key, err := identity.Key()
	if err != nil || now.IsZero() || lease < time.Second || lease > time.Hour {
		return Artifact{}, ErrInvalid
	}
	var random [24]byte
	if _, err = rand.Read(random[:]); err != nil {
		return Artifact{}, err
	}
	token := hex.EncodeToString(random[:])
	body, _ := json.Marshal(identity)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Artifact{}, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO ocr_artifacts(artifact_key,identity,state,owner_token,lease_expires_at,created_at) VALUES($1,$2,'pending',$3,$4,$5) ON CONFLICT DO NOTHING`, key, body, token, now.Add(lease), now)
	if err != nil {
		return Artifact{}, err
	}
	var a Artifact
	var state string
	var result []byte
	err = tx.QueryRow(ctx, `SELECT artifact_key,owner_token,generation,lease_expires_at,state,result FROM ocr_artifacts WHERE artifact_key=$1 FOR UPDATE`, key).Scan(&a.Key, &a.Token, &a.Generation, &a.LeaseExpires, &state, &result)
	if err != nil {
		return Artifact{}, err
	}
	if state == "ready" {
		a.Ready = true
		if json.Unmarshal(result, &a.Page) != nil {
			return Artifact{}, ErrConflict
		}
	} else if a.Token == token {
		a.Owned = true
	} else if !a.LeaseExpires.After(now) {
		a.Token = token
		a.Generation++
		a.LeaseExpires = now.Add(lease)
		a.Owned = true
		_, err = tx.Exec(ctx, `UPDATE ocr_artifacts SET owner_token=$2,generation=$3,lease_expires_at=$4 WHERE artifact_key=$1`, key, a.Token, a.Generation, a.LeaseExpires)
		if err != nil {
			return Artifact{}, err
		}
	}
	if !a.Owned {
		a.Token = ""
	}
	return a, tx.Commit(ctx)
}

func (s *Store) RenewArtifact(ctx context.Context, a Artifact, now time.Time, lease time.Duration) error {
	if !a.Owned || now.IsZero() || lease < time.Second || lease > time.Hour {
		return ErrInvalid
	}
	tag, err := s.pool.Exec(ctx, `UPDATE ocr_artifacts SET lease_expires_at=$4 WHERE artifact_key=$1 AND owner_token=$2 AND generation=$3 AND state='pending' AND lease_expires_at>$5`, a.Key, a.Token, a.Generation, now.Add(lease), now)
	if err == nil && tag.RowsAffected() != 1 {
		return ErrConflict
	}
	return err
}

func (s *Store) CompleteArtifact(ctx context.Context, a Artifact, page PageResult, now time.Time) error {
	if !a.Owned || now.IsZero() || (page.Status != "complete" && page.Status != "unreadable") || (page.Status == "complete") != (page.ExtractedText != "") {
		return ErrInvalid
	}
	hash := sha256.Sum256([]byte(page.ExtractedText))
	if page.OutputSHA256 != hex.EncodeToString(hash[:]) {
		return ErrInvalid
	}
	body, _ := json.Marshal(page)
	tag, err := s.pool.Exec(ctx, `UPDATE ocr_artifacts SET state='ready',result=$4,completed_at=$5 WHERE artifact_key=$1 AND owner_token=$2 AND generation=$3 AND state='pending' AND lease_expires_at>$5 AND identity->>'InputSHA256'=$6`, a.Key, a.Token, a.Generation, body, now, page.InputSHA256)
	if err == nil && tag.RowsAffected() != 1 {
		return ErrConflict
	}
	return err
}

// AssociateArtifact maps immutable content to this version's evidence location.
// Reuse never copies paid usage into the new page or attempt.
func (s *Store) AssociateArtifact(ctx context.Context, key string, page PageResult, reused bool) error {
	if page.RunID < 1 || page.DocumentVersionID < 1 || page.PageNumber < 1 || page.ResourceURL == "" || page.CreatedAt.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var body []byte
	var configuration string
	err = tx.QueryRow(ctx, `SELECT result,identity->>'Configuration' FROM ocr_artifacts WHERE artifact_key=$1 AND state='ready'`, key).Scan(&body, &configuration)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	var stored PageResult
	if json.Unmarshal(body, &stored) != nil {
		return ErrConflict
	}
	var valid bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM processing_runs WHERE id=$1 AND document_version_id=$2 AND configuration_version_id=$3)`, page.RunID, page.DocumentVersionID, configuration).Scan(&valid)
	if err != nil {
		return err
	}
	if !valid {
		return ErrConflict
	}
	stored.RunID, stored.DocumentVersionID, stored.PageNumber, stored.ResourceURL, stored.CreatedAt = page.RunID, page.DocumentVersionID, page.PageNumber, page.ResourceURL, page.CreatedAt
	if reused {
		stored.InputTokens = nil
		stored.OutputTokens = nil
		stored.CacheReadTokens = nil
	}
	if err = putPage(ctx, tx, stored); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO ocr_artifact_associations(run_id,page_number,artifact_key,reused,created_at) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, page.RunID, page.PageNumber, key, reused, page.CreatedAt)
	if err != nil {
		return err
	}
	var existing string
	if err = tx.QueryRow(ctx, `SELECT artifact_key FROM ocr_artifact_associations WHERE run_id=$1 AND page_number=$2`, page.RunID, page.PageNumber).Scan(&existing); err != nil {
		return err
	}
	if existing != key {
		return ErrConflict
	}
	return tx.Commit(ctx)
}
