package jobs

import (
	"context"
	"time"
)

// ArchiveBefore retires a fixed backlog without deleting identities or accounting
// links. Running work is rejected atomically; callers should stop workers first.
func (s *Store) ArchiveBefore(ctx context.Context, cutoff time.Time, actor string, at time.Time) (map[string]int64, error) {
	if cutoff.IsZero() || at.IsZero() || cutoff.After(at) || !validName(actor) {
		return nil, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	// Serialize with enqueue, claims and completion while choosing the cutoff set.
	if _, err = tx.Exec(ctx, "LOCK TABLE processing_jobs IN SHARE ROW EXCLUSIVE MODE"); err != nil {
		return nil, err
	}
	var running bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM processing_jobs WHERE created_at<=$1 AND state='running' AND archived_at IS NULL)", cutoff.UTC()).Scan(&running); err != nil {
		return nil, err
	}
	if running {
		return nil, ErrConflict
	}
	rows, err := tx.Query(ctx, `WITH retired AS (
 UPDATE processing_jobs SET archived_at=$2,archived_by=$3
 WHERE created_at<=$1 AND archived_at IS NULL AND state IN ('queued','retry_wait','failed')
 RETURNING state) SELECT state,count(*) FROM retired GROUP BY state`, cutoff.UTC(), at.UTC(), actor)
	if err != nil {
		return nil, err
	}
	counts := map[string]int64{}
	for rows.Next() {
		var state string
		var count int64
		if err = rows.Scan(&state, &count); err != nil {
			rows.Close()
			return nil, err
		}
		counts[state] = count
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return counts, nil
}
