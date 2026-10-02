//go:build integration

package processing

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueueMaintenanceSerializesWorkersAndReleasesBeforeInference(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	pool := processingTestDB(t)
	s := New(pool)
	var active atomic.Int32
	ready := make(chan error, 4)
	release := make(chan struct{})
	defer close(release)
	for i := 0; i < 4; i++ {
		go func() {
			err := s.MaintainQueue(ctx, func(ctx context.Context) error {
				if active.Add(1) != 1 {
					active.Add(-1)
					return errors.New("overlapping queue maintenance")
				}
				defer active.Add(-1)
				// The callback needs a pool connection independent of the lock.
				_, err := pool.Exec(ctx, "SELECT pg_sleep(0.02)")
				return err
			})
			ready <- err
			// Model work can overlap after all four maintenance calls return.
			select {
			case <-release:
			case <-ctx.Done():
			}
		}()
	}
	for i := 0; i < 4; i++ {
		select {
		case err := <-ready:
			if err != nil {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal("maintenance lock prevented parallel inference")
		}
	}

	want := errors.New("maintenance failed")
	if err := s.MaintainQueue(ctx, func(context.Context) error { return want }); !errors.Is(err, want) {
		t.Fatalf("callback failure lost: %v", err)
	}
	if err := s.MaintainQueue(ctx, func(context.Context) error { return nil }); err != nil {
		t.Fatalf("failed callback retained lock: %v", err)
	}

	// A cancelled lock waiter must not run its admission callback.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730072)"); err != nil {
		t.Fatal(err)
	}
	waitCtx, stop := context.WithTimeout(ctx, 100*time.Millisecond)
	defer stop()
	err = s.MaintainQueue(waitCtx, func(context.Context) error {
		t.Error("cancelled waiter ran maintenance")
		return nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock wait ignored cancellation: %v", err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if err = s.MaintainQueue(ctx, func(context.Context) error { return nil }); err != nil {
		t.Fatalf("cancelled waiter retained lock: %v", err)
	}
}
