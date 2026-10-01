//go:build integration

package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestLeaseLossCancelsHandlerAndFencesEffects(t *testing.T) {
	pool, _ := queueTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	store := New(pool)
	job, err := store.Enqueue(ctx, EnqueueRequest{Queue: "test", Kind: "test", IdempotencyKey: "ownership", Payload: json.RawMessage(`{}`), MaxAttempts: 3, RetryBase: time.Second, AvailableAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	w := Worker{Store: store, Queue: "test", ID: "owner", Lease: 3 * time.Second, PollInterval: time.Second, Handlers: map[string]Handler{"test": func(owned context.Context, _ Job) (Result, error) {
		if _, err := pool.Exec(ctx, "UPDATE processing_jobs SET lease_expires_at=now()-interval '1 second' WHERE id=$1", job.ID); err != nil {
			return Result{}, err
		}
		select {
		case <-owned.Done():
			return Result{}, owned.Err()
		case <-time.After(4 * time.Second):
			t.Error("handler outlived lost lease")
			return Result{}, nil
		}
	}}}
	started := time.Now()
	if _, err = w.RunOne(ctx); !errors.Is(err, ErrStaleClaim) {
		t.Fatalf("expected ownership failure: %v", err)
	}
	if time.Since(started) > 3*time.Second {
		t.Fatal("cancellation was not prompt")
	}
	var token string
	if err = pool.QueryRow(ctx, "SELECT claim_token FROM processing_jobs WHERE id=$1", job.ID).Scan(&token); err != nil {
		t.Fatal(err)
	}
	claim := Claim{Job: job, Token: token}
	claim.Attempt = 1
	if err = store.Complete(ctx, claim, Result{Payload: json.RawMessage(`{}`)}, time.Now()); !errors.Is(err, ErrStaleClaim) {
		t.Fatalf("expired owner completed: %v", err)
	}
	if err = store.Fail(ctx, claim, Failure{Code: "failed", Detail: "fixture", Temporary: true}, time.Now()); !errors.Is(err, ErrStaleClaim) {
		t.Fatalf("expired owner changed budget: %v", err)
	}
	var state string
	var attempts, maxAttempts int
	if err = pool.QueryRow(ctx, "SELECT state,attempt_count,max_attempts FROM processing_jobs WHERE id=$1", job.ID).Scan(&state, &attempts, &maxAttempts); err != nil {
		t.Fatal(err)
	}
	if state != "running" || attempts != 1 || maxAttempts != 3 {
		t.Fatal("lost owner mutated queue state")
	}
}

func TestTransientRecoveryIsBoundedAndDoesNotClaim(t *testing.T) {
	pool, _ := queueTestDB(t)
	calls := 0
	w := Worker{Store: New(pool), Queue: "test", ID: "recovery", Lease: 3 * time.Second, PollInterval: time.Second, Handlers: map[string]Handler{"test": nil}, BeforeClaim: func(context.Context, time.Time) error {
		calls++
		return &pgconn.PgError{Code: "40001", Message: "private"}
	}}
	if err := w.Run(context.Background()); err == nil || calls != 4 {
		t.Fatalf("recovery not bounded: calls=%d err=%v", calls, err)
	}
}
