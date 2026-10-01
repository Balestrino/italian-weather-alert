package acquisition

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type revisionTracker interface {
	RecordUnavailable(context.Context, string, int, string, int, time.Time) error
	ClearUnavailable(context.Context, string, int, string, time.Time) error
	Remember(context.Context, string, int, []DiscoveredDocument, time.Time) (DiscoveryResult, error)
	Plan(context.Context, string, int, time.Time, int) (RevisionPlan, error)
	RecordVersion(context.Context, string, string, int64, time.Time) (bool, error)
	CompleteBootstrap(context.Context, string, int, time.Time) error
}

type TrackingStore struct{ pool *pgxpool.Pool }

func NewTrackingStore(pool *pgxpool.Pool) *TrackingStore { return &TrackingStore{pool: pool} }

type PlannedResource struct {
	URL      string
	Required bool
}

type PlannedDocument struct {
	UnavailableStatus int
	NextAttemptAt     *time.Time
	URL               string
	PublicationDate   *time.Time
	Resources         []PlannedResource
}

type RevisionPlan struct {
	Bootstrap bool
	Documents []PlannedDocument
}

type DiscoveryResult struct {
	NewDocuments      int
	NewestPublication *time.Time
}

func (s *TrackingStore) Remember(ctx context.Context, sourceID string, revision int, documents []DiscoveredDocument, observedAt time.Time) (DiscoveryResult, error) {
	result := DiscoveryResult{}
	if sourceID == "" || revision < 1 || observedAt.IsZero() {
		return result, ErrInvalidConfiguration
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	for _, document := range documents {
		if document.URL == "" {
			return result, ErrInvalidConfiguration
		}
		var inserted string
		err = tx.QueryRow(ctx, `INSERT INTO acquisition_targets(source_id,url,configuration,source_publication_date,first_discovered_at,last_seen_at)
 VALUES($1,$2,$3,$4,$5,$5) ON CONFLICT(source_id,url) DO NOTHING RETURNING url`, sourceID, document.URL, revision, dateValue(document.PublicationDate), observedAt.UTC()).Scan(&inserted)
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = tx.Exec(ctx, `UPDATE acquisition_targets SET configuration=$3,
 source_publication_date=COALESCE($4::date,source_publication_date),last_seen_at=$5 WHERE source_id=$1 AND url=$2`, sourceID, document.URL, revision, dateValue(document.PublicationDate), observedAt.UTC())
		} else if err == nil {
			result.NewDocuments++
			if document.PublicationDate != nil && (result.NewestPublication == nil || document.PublicationDate.After(*result.NewestPublication)) {
				published := document.PublicationDate.UTC()
				result.NewestPublication = &published
			}
		}
		if err != nil {
			return result, err
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO acquisition_bootstrap(source_id,configuration) VALUES($1,$2)
 ON CONFLICT(source_id) DO UPDATE SET configuration=EXCLUDED.configuration,completed_at=CASE WHEN acquisition_bootstrap.configuration=EXCLUDED.configuration THEN acquisition_bootstrap.completed_at ELSE NULL END`, sourceID, revision)
	if err != nil {
		return result, err
	}
	return result, tx.Commit(ctx)
}

func (s *TrackingStore) AddExplicitReference(ctx context.Context, sourceID string, revision int, raw string, publicationDate *time.Time, observedAt time.Time) error {
	if _, err := s.Remember(ctx, sourceID, revision, []DiscoveredDocument{{URL: raw, PublicationDate: publicationDate}}, observedAt); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, "UPDATE acquisition_targets SET explicit_reference=true WHERE source_id=$1 AND url=$2", sourceID, raw)
	return err
}

func (s *TrackingStore) SetReviewState(ctx context.Context, sourceID, raw, relevance, measure string) error {
	if (relevance != "unknown" && relevance != "irrelevant" && relevance != "relevant") || (measure != "none" && measure != "ongoing" && measure != "unresolved" && measure != "ceased") {
		return ErrInvalidConfiguration
	}
	tag, err := s.pool.Exec(ctx, "UPDATE acquisition_targets SET relevance_state=$3,measure_state=$4 WHERE source_id=$1 AND url=$2", sourceID, raw, relevance, measure)
	if err == nil && tag.RowsAffected() == 0 {
		err = registryNotFound()
	}
	return err
}

func (s *TrackingStore) AddDependency(ctx context.Context, sourceID, parentURL, resourceURL string, required bool, observedAt time.Time) error {
	if sourceID == "" || parentURL == "" || resourceURL == "" || observedAt.IsZero() {
		return ErrInvalidConfiguration
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO acquisition_dependencies(source_id,parent_url,resource_url,required,first_recorded_at)
 VALUES($1,$2,$3,$4,$5) ON CONFLICT(source_id,parent_url,resource_url) DO UPDATE SET required=EXCLUDED.required`, sourceID, parentURL, resourceURL, required, observedAt.UTC())
	return err
}

func (s *TrackingStore) Plan(ctx context.Context, sourceID string, revision int, now time.Time, bootstrapDays int) (RevisionPlan, error) {
	if sourceID == "" || revision < 1 || now.IsZero() || bootstrapDays < 1 {
		return RevisionPlan{}, ErrInvalidConfiguration
	}
	plan := RevisionPlan{}
	err := s.pool.QueryRow(ctx, "SELECT completed_at IS NULL FROM acquisition_bootstrap WHERE source_id=$1 AND configuration=$2", sourceID, revision).Scan(&plan.Bootstrap)
	if errors.Is(err, pgx.ErrNoRows) {
		return plan, ErrInvalidConfiguration
	}
	if err != nil {
		return plan, err
	}
	cutoff := now.UTC().AddDate(0, 0, -bootstrapDays)
	rows, err := s.pool.Query(ctx, `SELECT t.url,t.source_publication_date,COALESCE(u.http_status,0),u.next_attempt_at FROM acquisition_targets t
 LEFT JOIN acquisition_unavailable_targets u ON u.source_id=t.source_id AND u.configuration=t.configuration AND u.url=t.url AND u.recovered_at IS NULL
 LEFT JOIN LATERAL (
   SELECT decision,observed_last_seen_at FROM acquisition_target_dispositions d
   WHERE d.source_id=t.source_id AND d.configuration=t.configuration AND d.url=t.url
   ORDER BY d.id DESC LIMIT 1
 ) disposition ON true
 WHERE t.source_id=$1 AND t.configuration=$2
   AND (t.source_publication_date IS NULL OR t.source_publication_date >= $3::date OR t.measure_state IN ('ongoing','unresolved') OR ($4 AND t.explicit_reference))
   AND (disposition.decision IS DISTINCT FROM 'exclude_stale_unavailable' OR t.last_seen_at > disposition.observed_last_seen_at)
 ORDER BY t.url`, sourceID, revision, cutoff, plan.Bootstrap)
	if err != nil {
		return plan, err
	}
	defer rows.Close()
	for rows.Next() {
		var document PlannedDocument
		if err = rows.Scan(&document.URL, &document.PublicationDate, &document.UnavailableStatus, &document.NextAttemptAt); err != nil {
			return plan, err
		}
		resourceRows, queryErr := s.pool.Query(ctx, `SELECT resource_url,required FROM acquisition_dependencies WHERE source_id=$1 AND parent_url=$2 ORDER BY resource_url`, sourceID, document.URL)
		if queryErr != nil {
			return plan, queryErr
		}
		for resourceRows.Next() {
			var resource PlannedResource
			if queryErr = resourceRows.Scan(&resource.URL, &resource.Required); queryErr != nil {
				resourceRows.Close()
				return plan, queryErr
			}
			document.Resources = append(document.Resources, resource)
		}
		queryErr = resourceRows.Err()
		resourceRows.Close()
		if queryErr != nil {
			return plan, queryErr
		}
		plan.Documents = append(plan.Documents, document)
	}
	return plan, rows.Err()
}

func (s *TrackingStore) RecordVersion(ctx context.Context, sourceID, raw string, versionID int64, checkedAt time.Time) (bool, error) {
	if versionID < 1 || checkedAt.IsZero() {
		return false, ErrInvalidConfiguration
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var previous *int64
	err = tx.QueryRow(ctx, `SELECT last_version_id FROM acquisition_targets WHERE source_id=$1 AND url=$2 FOR UPDATE`, sourceID, raw).Scan(&previous)
	if err != nil {
		return false, err
	}
	if _, err = tx.Exec(ctx, `UPDATE acquisition_targets SET last_version_id=$3,last_checked_at=$4 WHERE source_id=$1 AND url=$2`, sourceID, raw, versionID, checkedAt.UTC()); err != nil {
		return false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return false, err
	}
	return previous != nil && *previous != versionID, nil
}

func (s *TrackingStore) CompleteBootstrap(ctx context.Context, sourceID string, revision int, completedAt time.Time) error {
	tag, err := s.pool.Exec(ctx, "UPDATE acquisition_bootstrap SET completed_at=$3 WHERE source_id=$1 AND configuration=$2 AND completed_at IS NULL", sourceID, revision, completedAt.UTC())
	if err == nil && tag.RowsAffected() == 0 {
		var exists bool
		if queryErr := s.pool.QueryRow(ctx, "SELECT completed_at IS NOT NULL FROM acquisition_bootstrap WHERE source_id=$1 AND configuration=$2", sourceID, revision).Scan(&exists); queryErr != nil || !exists {
			return ErrInvalidConfiguration
		}
	}
	return err
}

func dateValue(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format("2006-01-02")
}

func registryNotFound() error { return errors.New("tracked document not found") }

// RecordUnavailable spaces permanent-looking HTTP failures 1h, 2h, 4h, ...
// up to 24h. Targets stay in the plan, so deferral cannot make a check complete.
func (s *TrackingStore) RecordUnavailable(ctx context.Context, source string, revision int, raw string, status int, at time.Time) error {
	if (status != 404 && status != 410) || at.IsZero() {
		return ErrInvalidConfiguration
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO acquisition_unavailable_targets
 (source_id,configuration,url,http_status,consecutive_failures,first_failed_at,last_failed_at,next_attempt_at)
 VALUES($1,$2,$3,$4,1,$5,$5,$5::timestamptz+interval '1 hour')
 ON CONFLICT(source_id,configuration,url) DO UPDATE SET
 http_status=EXCLUDED.http_status,
 consecutive_failures=CASE WHEN acquisition_unavailable_targets.recovered_at IS NULL THEN LEAST(acquisition_unavailable_targets.consecutive_failures+1,32) ELSE 1 END,
 last_failed_at=EXCLUDED.last_failed_at,
 next_attempt_at=EXCLUDED.last_failed_at+interval '1 hour' * CASE WHEN acquisition_unavailable_targets.recovered_at IS NULL THEN LEAST(24,power(2,LEAST(acquisition_unavailable_targets.consecutive_failures,5))) ELSE 1 END,
 recovered_at=NULL`, source, revision, raw, status, at.UTC())
	return err
}

func (s *TrackingStore) ClearUnavailable(ctx context.Context, source string, revision int, raw string, at time.Time) error {
	if at.IsZero() {
		return ErrInvalidConfiguration
	}
	_, err := s.pool.Exec(ctx, `UPDATE acquisition_unavailable_targets SET recovered_at=$4,consecutive_failures=0
 WHERE source_id=$1 AND configuration=$2 AND url=$3 AND recovered_at IS NULL`, source, revision, raw, at.UTC())
	return err
}
