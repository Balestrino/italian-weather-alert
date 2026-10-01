package processing

import (
	"context"
	"errors"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/jackc/pgx/v5"
)

// Only reviewed, document-local validation failures are quarantined. Unknown
// failures and capture/storage failures still stop the entire selected batch.
const localClassificationFailureSQL = `(r.stage='classification' AND COALESCE(a.error_code,'') IN ('classification_output_quotation','classification_output_normalization','classification_output_reason','classification_output_schema'))`

// Recovery guards all recorded providers together. A row lock serializes admission
// with the durable call insert, so concurrent workers cannot overspend the cap.
func admitRecoveryCall(ctx context.Context, tx pgx.Tx, input CallStart) error {
	var versions []int64
	var max int
	var since time.Time
	err := tx.QueryRow(ctx, `SELECT version_ids,max_calls,started_at FROM processing_recovery_limits WHERE singleton FOR UPDATE`).Scan(&versions, &max, &since)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var allowed bool
	if err = tx.QueryRow(ctx, `SELECT COALESCE(document_version_id=ANY($2),false) FROM processing_runs WHERE id=$1`, input.RunID, versions).Scan(&allowed); err != nil {
		return err
	}
	code := "recovery_scope_wait"
	if allowed {
		var calls int
		var failed bool
		if err = tx.QueryRow(ctx, `SELECT count(*),COALESCE(bool_or(state IN ('failed','uncertain')),false) FROM processing_provider_calls WHERE started_at >= $1`, since).Scan(&calls, &failed); err != nil {
			return err
		}
		var invalid bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM processing_run_attempts a JOIN processing_runs r ON r.id=a.run_id WHERE a.started_at >= $1 AND r.document_version_id=ANY($2) AND a.outcome='failed' AND COALESCE(a.error_code,'') <> 'provider_deferred' AND NOT `+localClassificationFailureSQL+`)`, since, versions).Scan(&invalid); err != nil {
			return err
		}
		if !failed && !invalid && calls < max {
			var quarantined bool
			if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM processing_run_attempts a JOIN processing_runs r ON r.id=a.run_id JOIN processing_runs current ON current.document_version_id=r.document_version_id WHERE current.id=$2 AND a.started_at >= $1 AND a.outcome='failed' AND `+localClassificationFailureSQL+`)`, since, input.RunID).Scan(&quarantined); err != nil {
				return err
			}
			if !quarantined {
				return nil
			}
			return &jobs.DeferredError{Until: input.StartedAt.Add(time.Minute), Code: "recovery_document_quarantined"}
		}
		code = "recovery_call_limit"
		if failed || invalid {
			code = "recovery_error_hold"
		}
	}
	return &jobs.DeferredError{Until: input.StartedAt.Add(time.Minute), Code: code}
}

// Deferral before claim preserves retry budgets for excluded work. Admission also
// enforces the limit immediately before transport, including already claimed jobs.
func (s *Store) DeferRecoveryJobs(ctx context.Context, now time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE processing_jobs j SET available_at=$1,updated_at=$2
 FROM processing_recovery_limits l
 WHERE l.singleton AND j.queue='inference' AND j.archived_at IS NULL
 AND j.state IN ('queued','retry_wait') AND j.available_at<$1
 AND (NOT COALESCE((j.payload->>'document_version_id')::bigint=ANY(l.version_ids),false)
 OR (SELECT count(*) FROM processing_provider_calls c WHERE c.started_at>=l.started_at)>=l.max_calls
 OR EXISTS(SELECT 1 FROM processing_provider_calls c WHERE c.started_at>=l.started_at AND c.state IN ('failed','uncertain'))
 OR EXISTS(SELECT 1 FROM processing_run_attempts a JOIN processing_runs r ON r.id=a.run_id WHERE a.started_at>=l.started_at AND r.document_version_id=ANY(l.version_ids) AND a.outcome='failed' AND COALESCE(a.error_code,'')<>'provider_deferred' AND (NOT `+localClassificationFailureSQL+` OR r.document_version_id=(j.payload->>'document_version_id')::bigint)))`, now.Add(time.Minute), now)
	return err
}
