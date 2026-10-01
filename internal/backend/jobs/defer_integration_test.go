//go:build integration

package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestProviderDeferralPreservesRetryBudget(t *testing.T) {
	pool, _ := queueTestDB(t)
	ctx := context.Background()
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	s := New(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	job, err := s.Enqueue(ctx, EnqueueRequest{Queue: "inference", Kind: "fixture", IdempotencyKey: "defer", Payload: json.RawMessage(`{}`), MaxAttempts: 3, RetryBase: 2 * time.Second, AvailableAt: now})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		claim, err := s.Claim(ctx, "inference", "worker", now, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if claim.MaxAttempts-claim.Attempt != 2 {
			t.Fatal("deferral spent retry budget")
		}
		if err = s.Defer(ctx, claim, DeferredError{Until: now.Add(time.Minute), Code: "provider_scope_held"}, now); err != nil {
			t.Fatal(err)
		}
		if _, err = s.Claim(ctx, "inference", "worker", now, time.Minute); err != ErrNoJob {
			t.Fatal("deferral ignored")
		}
		now = now.Add(time.Minute)
	}
	claim, err := s.Claim(ctx, "inference", "worker", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Fail(ctx, claim, Failure{Code: "temporary", Temporary: true}, now); err != nil {
		t.Fatal(err)
	}
	var available time.Time
	var ceiling, count, budget int
	if err = pool.QueryRow(ctx, `SELECT available_at,max_attempts,attempt_count,attempt_budget FROM processing_jobs WHERE id=$1`, job.ID).Scan(&available, &ceiling, &count, &budget); err != nil {
		t.Fatal(err)
	}
	if available.Sub(now) != 2*time.Second || ceiling-count != 2 || budget != 3 {
		t.Fatal("waiting increased backoff or reduced retry budget")
	}
	if err = Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
}
