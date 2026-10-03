package processing

import (
	"context"
	"time"
)

// MaintainQueue serializes pre-claim bulk updates across worker processes.
// The lock ends before claiming jobs or making provider calls, so inference
// remains parallel. The callback must use the store's pool, not this transaction.
func (s *Store) MaintainQueue(ctx context.Context, maintain func(context.Context) error) error {
	if maintain == nil {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730072)"); err != nil {
		return err
	}
	if err = maintain(ctx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
