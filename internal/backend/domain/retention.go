package domain

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/backend/evidenceguard"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ObjectDeleter interface {
	Delete(context.Context, string) error
}

type RetentionPolicy struct {
	ID        int64
	Months    int
	Actor     string
	CreatedAt time.Time
}

type CleanupResult struct {
	RunID, PolicyVersionID          int64
	DeletedVersions, DeletedObjects int
	FailedObjects, PendingObjects   int
	EvaluatedAt                     time.Time
}

type Retention struct {
	pool    *pgxpool.Pool
	objects ObjectDeleter
}

func NewRetention(pool *pgxpool.Pool, objects ObjectDeleter) *Retention {
	return &Retention{pool: pool, objects: objects}
}

func CalendarExpiry(firstAcquired time.Time, months int) (time.Time, error) {
	if firstAcquired.IsZero() || months < 1 || months > 120 {
		return time.Time{}, ErrInvalid
	}
	u := firstAcquired.UTC()
	targetFirst := time.Date(u.Year(), u.Month()+time.Month(months), 1, u.Hour(), u.Minute(), u.Second(), u.Nanosecond(), time.UTC)
	lastDay := time.Date(targetFirst.Year(), targetFirst.Month()+1, 0, u.Hour(), u.Minute(), u.Second(), u.Nanosecond(), time.UTC).Day()
	day := u.Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(targetFirst.Year(), targetFirst.Month(), day, u.Hour(), u.Minute(), u.Second(), u.Nanosecond(), time.UTC), nil
}

func (r *Retention) SetPolicy(ctx context.Context, months int, actor string, at time.Time) (RetentionPolicy, error) {
	if r == nil || r.pool == nil || months < 1 || months > 120 || strings.TrimSpace(actor) == "" || at.IsZero() {
		return RetentionPolicy{}, ErrInvalid
	}
	var policy RetentionPolicy
	err := r.pool.QueryRow(ctx, `INSERT INTO retention_policy_versions(months,actor,created_at) VALUES($1,$2,$3) RETURNING id,months,actor,created_at`, months, actor, at.UTC()).Scan(&policy.ID, &policy.Months, &policy.Actor, &policy.CreatedAt)
	return policy, err
}

func (r *Retention) CurrentPolicy(ctx context.Context) (RetentionPolicy, error) {
	if r == nil || r.pool == nil {
		return RetentionPolicy{}, ErrInvalid
	}
	var policy RetentionPolicy
	err := r.pool.QueryRow(ctx, `SELECT id,months,actor,created_at FROM retention_policy_versions ORDER BY id DESC LIMIT 1`).Scan(&policy.ID, &policy.Months, &policy.Actor, &policy.CreatedAt)
	return policy, err
}

func (r *Retention) Protect(ctx context.Context, versionID int64, scope, reference string, through *time.Time, at time.Time) error {
	if r == nil || r.pool == nil || versionID < 1 || strings.TrimSpace(scope) == "" || strings.TrimSpace(reference) == "" || at.IsZero() {
		return ErrInvalid
	}
	var until any
	if through != nil {
		until = through.UTC()
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO retention_version_protections(version_id,scope,reference,protect_through,created_at) VALUES($1,$2,$3,$4,$5)`, versionID, scope, reference, until, at.UTC())
	return err
}

func (r *Retention) Cleanup(ctx context.Context, evaluatedAt time.Time) (CleanupResult, error) {
	if r == nil || r.pool == nil || r.objects == nil || evaluatedAt.IsZero() {
		return CleanupResult{}, ErrInvalid
	}
	unlock, err := evidenceguard.Lock(ctx, r.pool, false)
	if err != nil {
		return CleanupResult{}, err
	}
	defer unlock()
	result, err := r.cleanupDatabase(ctx, evaluatedAt.UTC())
	if err != nil {
		return result, err
	}
	rows, err := r.pool.Query(ctx, `SELECT hash FROM retention_pending_objects ORDER BY queued_at,hash`)
	if err != nil {
		return result, err
	}
	var hashes []string
	for rows.Next() {
		var hash string
		if err = rows.Scan(&hash); err != nil {
			rows.Close()
			return result, err
		}
		hashes = append(hashes, hash)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return result, err
	}
	rows.Close()
	for _, hash := range hashes {
		deleteErr := r.objects.Delete(ctx, hash)
		outcome, code := "deleted", any(nil)
		if deleteErr != nil {
			outcome, code = "failed", "object_delete_failed"
			result.FailedObjects++
		} else {
			result.DeletedObjects++
		}
		tx, beginErr := r.pool.Begin(ctx)
		if beginErr != nil {
			return result, beginErr
		}
		_, recordErr := tx.Exec(ctx, `INSERT INTO retention_object_deletion_attempts(cleanup_run_id,hash,attempted_at,outcome,error_code) VALUES($1,$2,clock_timestamp(),$3,$4)`, result.RunID, hash, outcome, code)
		if recordErr == nil && deleteErr == nil {
			_, recordErr = tx.Exec(ctx, `DELETE FROM retention_pending_objects WHERE hash=$1`, hash)
		}
		if recordErr == nil {
			recordErr = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(ctx)
		}
		if recordErr != nil {
			return result, recordErr
		}
	}
	if err = r.pool.QueryRow(ctx, `SELECT count(*) FROM retention_pending_objects`).Scan(&result.PendingObjects); err != nil {
		return result, err
	}
	return result, nil
}

func (r *Retention) cleanupDatabase(ctx context.Context, evaluatedAt time.Time) (CleanupResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return CleanupResult{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(730022); SET LOCAL iwa.retention_cleanup='on'`); err != nil {
		return CleanupResult{}, err
	}
	var result CleanupResult
	result.EvaluatedAt = evaluatedAt
	var months int
	if err = tx.QueryRow(ctx, `SELECT id,months FROM retention_policy_versions ORDER BY id DESC LIMIT 1`).Scan(&result.PolicyVersionID, &months); err != nil {
		return result, err
	}
	statements := []string{
		`CREATE TEMP TABLE retention_candidates(id bigint PRIMARY KEY) ON COMMIT DROP`,
		`INSERT INTO retention_candidates
 SELECT v.id FROM retained_versions v
 WHERE $1 >= ((v.first_acquired_at AT TIME ZONE 'UTC') + make_interval(months => $2)) AT TIME ZONE 'UTC'
 AND NOT EXISTS (SELECT 1 FROM retention_version_protections p WHERE p.version_id=v.id AND (p.protect_through IS NULL OR p.protect_through >= $1))
 AND NOT EXISTS (
   SELECT 1 FROM domain_local_measures m
   WHERE m.document_version_id=v.id AND m.kind IN ('closure','restriction','prohibition','suspension','activation')
   AND NOT EXISTS (SELECT 1 FROM domain_measure_updates u WHERE u.target_measure_id=m.id AND u.relation IN ('reopens','cancels') AND u.applied_at <= $1)
   AND NOT EXISTS (
     SELECT 1 FROM domain_temporal_values t WHERE t.entity_kind='local_measure' AND t.entity_id=m.id AND t.meaning='validity'
     AND t.id=(SELECT max(t2.id) FROM domain_temporal_values t2 WHERE t2.entity_kind='local_measure' AND t2.entity_id=m.id AND t2.meaning='validity')
     AND t.precision='interval' AND t.end_instant <= $1))
 AND NOT EXISTS (
   SELECT 1 FROM domain_operational_phases p
   WHERE p.document_version_id=v.id
   AND NOT EXISTS (
     SELECT 1 FROM domain_temporal_values t WHERE t.entity_kind='operational_phase' AND t.entity_id=p.id AND t.meaning='validity'
     AND t.id=(SELECT max(t2.id) FROM domain_temporal_values t2 WHERE t2.entity_kind='operational_phase' AND t2.entity_id=p.id AND t2.meaning='validity')
     AND t.precision='interval' AND t.end_instant <= $1))
 AND NOT EXISTS (
   SELECT 1 FROM domain_regional_records rr
   WHERE rr.document_version_id=v.id AND (
	 EXISTS(SELECT 1 FROM domain_local_measures m WHERE m.regional_record_id=rr.id AND m.document_version_id<>v.id) OR
	 EXISTS(SELECT 1 FROM domain_operational_phases p WHERE p.regional_record_id=rr.id AND p.document_version_id<>v.id)))`,
		// Preserve original OCR provenance while any surviving version uses it.
		`WITH RECURSIVE dependencies(parent,child) AS (
 SELECT (a.result->>'DocumentVersionID')::bigint,p.document_version_id
 FROM ocr_artifacts a JOIN ocr_artifact_associations x ON x.artifact_key=a.artifact_key
 JOIN ocr_page_results p USING(run_id,page_number) WHERE a.state='ready'
 UNION
 SELECT origin.document_version_id,consumer.document_version_id
 FROM processing_segment_checkpoint_uses u
 JOIN processing_segment_checkpoints c USING(stage,configuration,input_sha256)
 JOIN processing_provider_calls call ON call.id=c.call_id
 JOIN processing_runs origin ON origin.id=call.run_id JOIN processing_runs consumer ON consumer.id=u.run_id
 UNION
 SELECT original.document_version_id,current.document_version_id FROM interpretation_reuse u
 JOIN processing_runs original ON original.id=u.original_run_id JOIN processing_runs current ON current.id=u.run_id
 ), needed(id) AS (
 SELECT id FROM retained_versions WHERE id NOT IN(SELECT id FROM retention_candidates)
 UNION SELECT d.parent FROM needed n JOIN dependencies d ON d.child=n.id)
 DELETE FROM retention_candidates WHERE id IN(SELECT id FROM needed)`,
		`CREATE TEMP TABLE retention_local_measures ON COMMIT DROP AS SELECT id FROM domain_local_measures WHERE document_version_id IN(SELECT id FROM retention_candidates)`,
		`CREATE TEMP TABLE retention_phases ON COMMIT DROP AS SELECT id FROM domain_operational_phases WHERE document_version_id IN(SELECT id FROM retention_candidates)`,
		`CREATE TEMP TABLE retention_regional ON COMMIT DROP AS SELECT id FROM domain_regional_records WHERE document_version_id IN(SELECT id FROM retention_candidates)`,
		`CREATE TEMP TABLE retention_extractions ON COMMIT DROP AS SELECT run_id FROM extraction_results WHERE document_version_id IN(SELECT id FROM retention_candidates)`,
		`CREATE TEMP TABLE retention_linking_runs ON COMMIT DROP AS
 SELECT run_id FROM linking_results WHERE run_id IN(SELECT id FROM processing_runs WHERE document_version_id IN(SELECT id FROM retention_candidates))
	UNION SELECT run_id FROM linking_results WHERE current_extraction_run_id IN(SELECT run_id FROM retention_extractions) OR candidate_extraction_run_id IN(SELECT run_id FROM retention_extractions)
	UNION SELECT linking_run_id FROM domain_measure_updates WHERE update_measure_id IN(SELECT id FROM retention_local_measures) OR target_measure_id IN(SELECT id FROM retention_local_measures)`,
		`CREATE TEMP TABLE retention_processing_runs ON COMMIT DROP AS
 SELECT id FROM processing_runs WHERE document_version_id IN(SELECT id FROM retention_candidates)
	UNION SELECT run_id FROM retention_extractions
	UNION SELECT run_id FROM retention_linking_runs
	UNION SELECT run_id FROM measure_embeddings WHERE extraction_run_id IN(SELECT run_id FROM retention_extractions)`,
		`DELETE FROM domain_preceding_state_warnings WHERE local_measure_id IN(SELECT id FROM retention_local_measures) OR newer_document_version_id IN(SELECT id FROM retention_candidates)`,
		`DELETE FROM domain_interpretation_events WHERE local_measure_id IN(SELECT id FROM retention_local_measures) OR evidence_document_version_id IN(SELECT id FROM retention_candidates)`,
		`DELETE FROM domain_temporal_values WHERE evidence_document_version_id IN(SELECT id FROM retention_candidates) OR (entity_kind='local_measure' AND entity_id IN(SELECT id FROM retention_local_measures)) OR (entity_kind='operational_phase' AND entity_id IN(SELECT id FROM retention_phases)) OR (entity_kind='regional_record' AND entity_id IN(SELECT id FROM retention_regional))`,
		`DELETE FROM domain_measure_updates WHERE linking_run_id IN(SELECT run_id FROM retention_linking_runs) OR update_measure_id IN(SELECT id FROM retention_local_measures) OR target_measure_id IN(SELECT id FROM retention_local_measures)`,
		`DELETE FROM domain_measure_bindings WHERE local_measure_id IN(SELECT id FROM retention_local_measures) OR extraction_run_id IN(SELECT run_id FROM retention_extractions)`,
		`DELETE FROM domain_local_measures WHERE id IN(SELECT id FROM retention_local_measures)`,
		`DELETE FROM domain_operational_phases WHERE id IN(SELECT id FROM retention_phases)`,
		`DELETE FROM domain_regional_facts WHERE regional_record_id IN(SELECT id FROM retention_regional)`,
		`DELETE FROM domain_regional_records WHERE id IN(SELECT id FROM retention_regional)`,
		`DELETE FROM measure_embeddings WHERE run_id IN(SELECT id FROM retention_processing_runs) OR extraction_run_id IN(SELECT run_id FROM retention_extractions)`,
		`DELETE FROM linking_evidence WHERE run_id IN(SELECT run_id FROM retention_linking_runs)`,
		`DELETE FROM linking_candidates WHERE run_id IN(SELECT run_id FROM retention_linking_runs) OR candidate_extraction_run_id IN(SELECT run_id FROM retention_extractions)`,
		`DELETE FROM linking_results WHERE run_id IN(SELECT run_id FROM retention_linking_runs)`,
		`DELETE FROM extraction_temporal_candidate_evidence WHERE run_id IN(SELECT run_id FROM retention_extractions)`,
		`DELETE FROM extracted_temporal_candidates WHERE run_id IN(SELECT run_id FROM retention_extractions)`,
		`DELETE FROM extraction_evidence WHERE run_id IN(SELECT run_id FROM retention_extractions)`,
		`DELETE FROM extracted_measures WHERE run_id IN(SELECT run_id FROM retention_extractions)`,
		`DELETE FROM extraction_segments WHERE run_id IN(SELECT run_id FROM retention_extractions)`,
		`DELETE FROM extraction_results WHERE run_id IN(SELECT run_id FROM retention_extractions)`,
		`DELETE FROM classification_segments WHERE run_id IN(SELECT id FROM retention_processing_runs) OR document_version_id IN(SELECT id FROM retention_candidates)`,
		`DELETE FROM classification_results WHERE run_id IN(SELECT id FROM retention_processing_runs) OR document_version_id IN(SELECT id FROM retention_candidates)`,
		`DELETE FROM ocr_artifact_associations WHERE run_id IN(SELECT id FROM retention_processing_runs)`,
		`DELETE FROM ocr_artifacts a WHERE (a.result->>'DocumentVersionID')::bigint IN(SELECT id FROM retention_candidates) AND NOT EXISTS(SELECT 1 FROM ocr_artifact_associations x WHERE x.artifact_key=a.artifact_key)`,
		`DELETE FROM ocr_page_results WHERE run_id IN(SELECT id FROM retention_processing_runs) OR document_version_id IN(SELECT id FROM retention_candidates)`,
		`DELETE FROM ocr_resource_results WHERE run_id IN(SELECT id FROM retention_processing_runs) OR document_version_id IN(SELECT id FROM retention_candidates)`,
		`DELETE FROM interpretation_reuse WHERE run_id IN(SELECT id FROM retention_processing_runs)`,
		`DELETE FROM interpretation_input_manifests WHERE run_id IN(SELECT id FROM retention_processing_runs)`,
		`DELETE FROM processing_segment_checkpoint_uses WHERE run_id IN(SELECT id FROM retention_processing_runs)`,
		`DELETE FROM processing_segment_checkpoints WHERE call_id IN(SELECT id FROM processing_provider_calls WHERE run_id IN(SELECT id FROM retention_processing_runs))`,
		`DELETE FROM processing_provider_calls WHERE run_id IN(SELECT id FROM retention_processing_runs)`,
		`DELETE FROM processing_run_attempts WHERE run_id IN(SELECT id FROM retention_processing_runs)`,
		`DELETE FROM processing_runs WHERE id IN(SELECT id FROM retention_processing_runs)`,
		`DELETE FROM interpretation_triggers WHERE document_version_id IN(SELECT id FROM retention_candidates)`,
		`DELETE FROM interpretation_reprocessing_selections WHERE document_version_id IN(SELECT id FROM retention_candidates)`,
		`CREATE TEMP TABLE retention_orphan_objects ON COMMIT DROP AS
 SELECT o.hash,o.object_key FROM retained_objects o
 WHERE EXISTS(SELECT 1 FROM retained_resources r WHERE r.object_hash=o.hash AND r.version_id IN(SELECT id FROM retention_candidates))
 AND NOT EXISTS(SELECT 1 FROM retained_resources r WHERE r.object_hash=o.hash AND r.version_id NOT IN(SELECT id FROM retention_candidates))`,
		`DELETE FROM retained_acquisitions WHERE version_id IN(SELECT id FROM retention_candidates)`,
		`DELETE FROM retained_resources WHERE version_id IN(SELECT id FROM retention_candidates)`,
		`DELETE FROM retained_versions WHERE id IN(SELECT id FROM retention_candidates)`,
		`DELETE FROM retained_documents d WHERE NOT EXISTS(SELECT 1 FROM retained_versions v WHERE v.document_id=d.id) AND NOT EXISTS(SELECT 1 FROM retained_acquisitions a WHERE a.document_id=d.id)`,
	}
	for index, statement := range statements {
		if index == 1 {
			_, err = tx.Exec(ctx, statement, evaluatedAt, months)
		} else {
			_, err = tx.Exec(ctx, statement)
		}
		if err != nil {
			return result, err
		}
	}
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM retention_candidates`).Scan(&result.DeletedVersions); err != nil {
		return result, err
	}
	if err = tx.QueryRow(ctx, `INSERT INTO retention_cleanup_runs(policy_version_id,evaluated_at,deleted_versions,queued_objects,committed_at)
 VALUES($1,$2,$3,(SELECT count(*) FROM retention_orphan_objects),clock_timestamp()) RETURNING id,queued_objects`, result.PolicyVersionID, evaluatedAt, result.DeletedVersions).Scan(&result.RunID, &result.PendingObjects); err != nil {
		return result, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO retention_pending_objects(hash,object_key,first_cleanup_run_id,queued_at)
 SELECT hash,object_key,$1,clock_timestamp() FROM retention_orphan_objects ON CONFLICT(hash) DO NOTHING`, result.RunID); err != nil {
		return result, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM retained_objects WHERE hash IN(SELECT hash FROM retention_orphan_objects)`); err != nil {
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return result, err
	}
	return result, nil
}
