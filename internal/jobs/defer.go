package jobs

import (
	"context"
	"time"
)

type DeferredError struct {
	Until time.Time
	Code  string
}

func (e *DeferredError) Error() string { return "processing_deferred" }

// Defer preserves monotonically numbered history and refunds this scheduling
// attempt by extending its absolute ceiling. The configured retry budget is unchanged.
func (s *Store) Defer(ctx context.Context, c Claim, deferred DeferredError, now time.Time) error {
	if !validName(deferred.Code) || !deferred.Until.After(now) || now.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = lockClaim(ctx, tx, c, now); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE processing_attempts SET outcome='deferred',finished_at=$3,error_code=$4,error_detail='waiting for provider recovery' WHERE job_id=$1 AND number=$2`, c.ID, c.Attempt, now.UTC(), deferred.Code)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE processing_jobs SET state='retry_wait',max_attempts=max_attempts+1,available_at=$2,claimed_by=NULL,claim_token=NULL,lease_expires_at=NULL,last_error_code=$3,last_error_detail='waiting for provider recovery',updated_at=$4 WHERE id=$1`, c.ID, deferred.Until.UTC(), deferred.Code, now.UTC())
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
