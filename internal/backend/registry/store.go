package registry

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func (s *Store) CreateAuthority(ctx context.Context, a Authority) error {
	if strings.TrimSpace(a.ID) == "" || strings.TrimSpace(a.Name) == "" || !validURL(a.OfficialURL) {
		return ErrInvalid
	}
	_, err := s.pool.Exec(ctx, "INSERT INTO registry_authorities(id,name,official_url) VALUES ($1,$2,$3)", a.ID, a.Name, a.OfficialURL)
	return err
}
func (s *Store) CreateChannel(ctx context.Context, c Channel) error {
	if strings.TrimSpace(c.ID) == "" || strings.TrimSpace(c.Platform) == "" || !validURL(c.URL) {
		return ErrInvalid
	}
	_, err := s.pool.Exec(ctx, "INSERT INTO registry_channels(id,publisher_id,platform,url,external) VALUES ($1,$2,$3,$4,$5)", c.ID, c.PublisherID, c.Platform, c.URL, c.External)
	return err
}
func (s *Store) CreateSource(ctx context.Context, src Source, c Configuration, actor string) error {
	c.defaults()
	if src.ID == "" || src.Territory == "" || strings.TrimSpace(actor) == "" || !c.valid() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES ($1,$2,$3,$4,$5)`, src.ID, src.AuthorityID, src.ChannelID, src.ProductID, src.Territory)
	if err != nil {
		return err
	}
	if err = insertConfig(ctx, tx, src.ID, 1, c, actor); err != nil {
		return err
	}
	if err = territory.Associate(ctx, tx, src.ID, src.ProductID, src.Territory, c.ProcessingProfile, actor); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func insertConfig(ctx context.Context, tx pgx.Tx, id string, revision int, c Configuration, actor string) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO registry_configurations(source_id,revision,body,actor) VALUES ($1,$2,$3,$4)`, id, revision, b, actor); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "UPDATE registry_sources SET latest_revision=$2 WHERE id=$1", id, revision)
	return err
}

// AppendConfiguration preserves the applied revision, its collection and public
// status. A proposed revision does not change active collection boundaries.
func (s *Store) AppendConfiguration(ctx context.Context, id string, expected int, c Configuration, actor string) (int, error) {
	c.defaults()
	if !c.valid() || strings.TrimSpace(actor) == "" {
		return 0, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	st, err := state(ctx, tx, id, true)
	if err != nil {
		return 0, err
	}
	if st.LatestRevision != expected {
		return 0, ErrConflict
	}
	revision := expected + 1
	if err = insertConfig(ctx, tx, id, revision, c, actor); err != nil {
		return 0, err
	}
	return revision, tx.Commit(ctx)
}

type querier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func state(ctx context.Context, q querier, id string, lock bool) (State, error) {
	var st State
	query := `SELECT id,authority_id,channel_id,product_id,territory,latest_revision,active_revision,collection_enabled,public_enabled,interpretation_suspended_at,EXISTS(SELECT 1 FROM registry_events e WHERE e.source_id=registry_sources.id AND e.revision=registry_sources.active_revision AND e.kind='acceptance' AND NOT EXISTS(SELECT 1 FROM registry_regressions r WHERE r.source_id=e.source_id AND r.revision=e.revision AND NOT r.passed AND r.recorded_at>=e.created_at)) FROM registry_sources WHERE id=$1`
	if lock {
		query += " FOR UPDATE"
	}
	err := q.QueryRow(ctx, query, id).Scan(&st.Source.ID, &st.Source.AuthorityID, &st.Source.ChannelID, &st.Source.ProductID, &st.Source.Territory, &st.LatestRevision, &st.ActiveRevision, &st.CollectionEnabled, &st.PublicEnabled, &st.InterpretationSuspendedAt, &st.Accepted)
	if errors.Is(err, pgx.ErrNoRows) {
		return st, ErrNotFound
	}
	return st, err
}
func (s *Store) State(ctx context.Context, id string) (State, error) {
	return state(ctx, s.pool, id, false)
}
func version(ctx context.Context, q querier, id string, rev int) (Version, error) {
	v := Version{SourceID: id, Revision: rev}
	var b []byte
	err := q.QueryRow(ctx, "SELECT body,actor,created_at FROM registry_configurations WHERE source_id=$1 AND revision=$2", id, rev).Scan(&b, &v.Actor, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrNotFound
	}
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(b, &v.Configuration)
	return v, err
}
func (s *Store) Version(ctx context.Context, id string, rev int) (Version, error) {
	return version(ctx, s.pool, id, rev)
}
func record(ctx context.Context, tx pgx.Tx, id string, rev int, kind, actor string, evidence any) error {
	b, err := json.Marshal(evidence)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "INSERT INTO registry_events(source_id,revision,kind,actor,evidence,created_at) VALUES ($1,$2,$3,$4,$5,clock_timestamp())", id, rev, kind, actor, b)
	return err
}
func exists(ctx context.Context, tx pgx.Tx, id string, rev int, kind string) (bool, error) {
	var yes bool
	err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM registry_events WHERE source_id=$1 AND revision=$2 AND kind=$3)", id, rev, kind).Scan(&yes)
	return yes, err
}

// RecordPreview records successful acquisition evidence from the engine. The
// administrative HTTP interface executes that engine; it cannot submit success.
func (s *Store) RecordPreview(ctx context.Context, id string, rev int, actor string, e Evidence) error {
	if !validEvidence(e) {
		return ErrInvalid
	}
	return s.change(ctx, id, rev, actor, func(tx pgx.Tx, st State, c Configuration) error {
		if !c.Policy.CollectionPermitted {
			return ErrPrerequisite
		}
		return record(ctx, tx, id, rev, "preview", actor, e)
	})
}
func (s *Store) change(ctx context.Context, id string, rev int, actor string, fn func(pgx.Tx, State, Configuration) error) error {
	if strings.TrimSpace(actor) == "" {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	st, err := state(ctx, tx, id, true)
	if err != nil {
		return err
	}
	v, err := version(ctx, tx, id, rev)
	if err != nil {
		return err
	}
	if err = fn(tx, st, v.Configuration); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *Store) EnableCollection(ctx context.Context, id string, rev int, actor string) error {
	return s.change(ctx, id, rev, actor, func(tx pgx.Tx, st State, c Configuration) error {
		if err := territory.CheckConfiguration(ctx, tx, id, c.ProcessingProfile, c.regionalKind()); err != nil {
			return err
		}
		if rev != st.LatestRevision && (st.ActiveRevision == nil || *st.ActiveRevision != rev) {
			return ErrConflict
		}
		preview, err := exists(ctx, tx, id, rev, "preview")
		if err != nil {
			return err
		}
		if !preview || !c.Policy.CollectionPermitted {
			return ErrPrerequisite
		}
		// Switching configurations requires separate acceptance and public activation.
		same := st.ActiveRevision != nil && *st.ActiveRevision == rev
		if _, err = tx.Exec(ctx, "UPDATE registry_sources SET active_revision=$2,collection_enabled=true,public_enabled=public_enabled AND $3 WHERE id=$1", id, rev, same); err != nil {
			return err
		}
		return record(ctx, tx, id, rev, "collection_enabled", actor, struct{}{})
	})
}
func sameSections(a, b []string) bool {
	x := slices.Clone(a)
	y := slices.Clone(b)
	slices.Sort(x)
	slices.Sort(y)
	return slices.Equal(x, y)
}
func (s *Store) Accept(ctx context.Context, id string, rev int, actor string, a Acceptance) error {
	if !validEvidence(a.Report) || a.PeriodStart.IsZero() || a.PeriodEnd.Sub(a.PeriodStart) < 7*24*time.Hour || a.PeriodEnd.After(a.Report.ObservedAt) || !a.ExtractionVerified || !a.UpdatesVerified || !a.AttachmentsVerified || !a.ScannedAttachmentsVerified || !a.HistoryVerified || !a.FailureBehaviorVerified || !a.InterfacesEquivalent || a.KnownOmissions != 0 || a.UnsupportedAssertions != 0 || !validCoverageStatus(a.CoverageStatus, a.CoverageLimitations) {
		return ErrInvalid
	}
	return s.change(ctx, id, rev, actor, func(tx pgx.Tx, st State, c Configuration) error {
		if regionalProduct(st.Source.ProductID) && !sameStrings(a.RiskCoverage, ToscanaRisks) {
			return ErrInvalid
		}
		if st.Source.ProductID == "municipal" && len(a.RiskCoverage) != 0 {
			return ErrInvalid
		}
		ready, err := regressionReady(ctx, tx, id, rev)
		if err != nil {
			return err
		}
		if !ready {
			return ErrPrerequisite
		}
		if st.ActiveRevision == nil || *st.ActiveRevision != rev || !st.CollectionEnabled {
			return ErrPrerequisite
		}
		if !sameSections(a.Sections, c.Sections) || len(c.Unresolved) > 0 || !c.Policy.PublicationPermitted || c.Provenance == nil {
			return ErrPrerequisite
		}
		var external bool
		var official string
		if err := tx.QueryRow(ctx, `SELECT c.external,a.official_url FROM registry_channels c JOIN registry_sources s ON s.channel_id=c.id JOIN registry_authorities a ON a.id=s.authority_id WHERE s.id=$1`, id).Scan(&external, &official); err != nil {
			return err
		}
		officialURL, _ := url.Parse(official)
		if external {
			r := c.Referral
			if r == nil || r.ProductID != st.Source.ProductID || r.Territory != st.Source.Territory || r.Destination != c.URL || !sameSections(r.Sections, c.Sections) {
				return ErrPrerequisite
			}
			referring, _ := url.Parse(r.Evidence.URL)
			if referring.Hostname() != officialURL.Hostname() {
				return ErrPrerequisite
			}
		} else {
			origin, _ := url.Parse(c.Provenance.URL)
			if origin.Hostname() != officialURL.Hostname() {
				return ErrPrerequisite
			}
		}
		return record(ctx, tx, id, rev, "acceptance", actor, a)
	})
}
func (s *Store) EnablePublic(ctx context.Context, id string, rev int, actor string) error {
	return s.change(ctx, id, rev, actor, func(tx pgx.Tx, st State, c Configuration) error {
		var allowed bool
		if err := tx.QueryRow(ctx, `SELECT registry_source_in_public_scope($1)`, id).Scan(&allowed); err != nil {
			return err
		}
		if !allowed {
			return territory.ErrPublicScope
		}
		if st.ActiveRevision == nil || *st.ActiveRevision != rev {
			return ErrConflict
		}
		ready, err := regressionReady(ctx, tx, id, rev)
		if err != nil {
			return err
		}
		if !ready {
			return ErrPrerequisite
		}
		var eligible bool
		if err := tx.QueryRow(ctx, "SELECT public_eligible FROM registry_products WHERE id=$1", st.Source.ProductID).Scan(&eligible); err != nil {
			return err
		}
		var accepted bool
		err = tx.QueryRow(ctx, "SELECT "+acceptedSQL, id, rev).Scan(&accepted)
		if err != nil {
			return err
		}
		var releaseAccepted bool
		if err = tx.QueryRow(ctx, "SELECT "+releaseAcceptedSQL, id, rev).Scan(&releaseAccepted); err != nil {
			return err
		}
		if !eligible || !accepted || !releaseAccepted || !c.Policy.PublicationPermitted || st.InterpretationSuspendedAt != nil {
			return ErrPrerequisite
		}
		if _, err = tx.Exec(ctx, "UPDATE registry_sources SET public_enabled=true WHERE id=$1", id); err != nil {
			return err
		}
		return record(ctx, tx, id, rev, "public_enabled", actor, struct{}{})
	})
}

// Suspension records a separate decision; stopping collection does not erase
// published history or imply that the underlying measures expired.
func (s *Store) Disable(ctx context.Context, id string, rev int, actor string, public bool) error {
	return s.change(ctx, id, rev, actor, func(tx pgx.Tx, st State, c Configuration) error {
		if st.ActiveRevision == nil || *st.ActiveRevision != rev {
			return ErrConflict
		}
		column, kind := "collection_enabled", "collection_disabled"
		if public {
			column, kind = "public_enabled", "public_disabled"
		}
		if _, err := tx.Exec(ctx, "UPDATE registry_sources SET "+column+"=false WHERE id=$1", id); err != nil {
			return err
		}
		return record(ctx, tx, id, rev, kind, actor, struct{}{})
	})
}
func (s *Store) Events(ctx context.Context, id string) ([]Event, error) {
	rows, err := s.pool.Query(ctx, "SELECT revision,kind,actor,created_at,evidence FROM registry_events WHERE source_id=$1 ORDER BY id", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []Event{}
	for rows.Next() {
		var e Event
		if err = rows.Scan(&e.Revision, &e.Kind, &e.Actor, &e.CreatedAt, &e.Evidence); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *Store) Authority(ctx context.Context, id string) (Authority, error) {
	var a Authority
	err := s.pool.QueryRow(ctx, "SELECT id,name,official_url FROM registry_authorities WHERE id=$1", id).Scan(&a.ID, &a.Name, &a.OfficialURL)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return a, err
}
func (s *Store) Channel(ctx context.Context, id string) (Channel, error) {
	var c Channel
	err := s.pool.QueryRow(ctx, "SELECT id,publisher_id,platform,url,external FROM registry_channels WHERE id=$1", id).Scan(&c.ID, &c.PublisherID, &c.Platform, &c.URL, &c.External)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return c, err
}

func (s *Store) CheckTerritorialExecution(ctx context.Context, id string, revision int) error {
	v, err := s.Version(ctx, id, revision)
	if err != nil {
		return err
	}
	if err = territory.CheckConfiguration(ctx, s.pool, id, v.Configuration.ProcessingProfile, v.Configuration.regionalKind()); err != nil {
		return err
	}
	return territory.CheckExecution(ctx, s.pool, id)
}
