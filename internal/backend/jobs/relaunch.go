package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Relaunch retries exactly the selected failed job. Its identity, payload,
// previous attempts and effect keys survive. A stale page cannot grant a second
// budget, and repeated delivery of the same command is idempotent.
func (s *Store) Relaunch(ctx context.Context, id int64, afterAttempt int, actor string, at time.Time) error {
	if id < 1 || afterAttempt < 1 || !validName(actor) || at.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var state string
	var attempts, budget int
	err = tx.QueryRow(ctx, `SELECT state,attempt_count,attempt_budget FROM processing_jobs WHERE id=$1 AND archived_at IS NULL FOR UPDATE`, id).Scan(&state, &attempts, &budget)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoJob
	}
	if err != nil {
		return err
	}
	var recorded bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM processing_job_relaunches WHERE job_id=$1 AND after_attempt=$2)`, id, afterAttempt).Scan(&recorded); err != nil {
		return err
	}
	if recorded {
		return tx.Commit(ctx)
	}
	if state != "failed" || attempts != afterAttempt || attempts > 2147483647-budget {
		return ErrConflict
	}
	if _, err = tx.Exec(ctx, `INSERT INTO processing_job_relaunches(job_id,after_attempt,actor,created_at) VALUES($1,$2,$3,$4)`, id, afterAttempt, actor, at.UTC()); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE processing_jobs SET state='queued',max_attempts=attempt_count+attempt_budget,available_at=$2,updated_at=$2,completed_at=NULL WHERE id=$1`, id, at.UTC()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
