//go:build integration

package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestEmbeddingControlsPreservePendingJobsAndSourceChoices(t *testing.T) {
	pool, _ := queueTestDB(t)
	ctx := context.Background()
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	// Synthetic ledger/registry contracts isolate controls from acquisition.
	if _, err := pool.Exec(ctx, `CREATE TABLE registry_sources(id text PRIMARY KEY);
        INSERT INTO registry_sources VALUES('source-a'),('source-b');
        CREATE TABLE processing_runs(id bigint PRIMARY KEY,source_id text,stage text);
        INSERT INTO processing_runs VALUES(1,'source-a','extraction'),(2,'source-b','extraction');`); err != nil {
		t.Fatal(err)
	}
	queue := New(pool)
	now := time.Now().UTC()
	state, err := queue.EmbeddingState(ctx)
	if err != nil || state.Enabled || state.Revision != 0 {
		t.Fatal(state, err)
	}
	for _, source := range []string{"source-a", "source-b"} {
		state, err = queue.SourceEmbeddingState(ctx, source)
		if err != nil || state.Enabled || state.Revision != 0 {
			t.Fatal(state, err)
		}
	}
	for _, entry := range []struct{ key, payload string }{{"a", `{"extraction_run_id":1}`}, {"b", `{"extraction_run_id":2}`}} {
		if _, err = queue.Enqueue(ctx, EnqueueRequest{Queue: "inference", Kind: "embed_measure", IdempotencyKey: entry.key, Payload: json.RawMessage(entry.payload), MaxAttempts: 3, RetryBase: time.Second, AvailableAt: now}); err != nil {
			t.Fatal(err)
		}
	}
	idle := func() {
		t.Helper()
		if _, err := queue.Claim(ctx, "inference", "worker", now, time.Minute); !errors.Is(err, ErrNoJob) {
			t.Fatal("disabled embedding claimed", err)
		}
	}
	idle()
	if _, err = queue.SetEmbeddingEnabled(ctx, 0, true, "global", now); err != nil {
		t.Fatal(err)
	}
	idle() // Global alone is insufficient.
	if _, err = queue.SetSourceEmbeddingEnabled(ctx, "source-a", 0, true, "source", now); err != nil {
		t.Fatal(err)
	}
	if _, err = queue.SetSourceEmbeddingEnabled(ctx, "source-a", 0, false, "stale", now); !errors.Is(err, ErrConflict) {
		t.Fatal("stale source change", err)
	}
	claim, err := queue.Claim(ctx, "inference", "worker-a", now, time.Minute)
	if err != nil || claim.IdempotencyKey != "a" {
		t.Fatal(claim, err)
	}
	if _, err = queue.SetEmbeddingEnabled(ctx, 1, false, "off", now); err != nil {
		t.Fatal(err)
	}
	if _, err = queue.SetSourceEmbeddingEnabled(ctx, "source-b", 0, true, "selected-while-off", now); err != nil {
		t.Fatal(err)
	}
	idle()
	// Already admitted work can finish while the global switch is off.
	if err = queue.Complete(ctx, claim, Result{Payload: json.RawMessage(`{"complete":true}`)}, now); err != nil {
		t.Fatal(err)
	}
	if err = Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	state, err = queue.EmbeddingState(ctx)
	if err != nil || state.Enabled || state.Revision != 2 {
		t.Fatal("migration reset control", state, err)
	}
	state, err = queue.SourceEmbeddingState(ctx, "source-a")
	if err != nil || !state.Enabled {
		t.Fatal("global toggle lost source choice", state, err)
	}
	var pending, budget, attempts int
	if err = pool.QueryRow(ctx, `SELECT count(*),sum(attempt_count),sum(max_attempts) FROM processing_jobs WHERE state='queued'`).Scan(&pending, &budget, &attempts); err != nil || pending != 1 || budget != 0 || attempts != 3 {
		t.Fatal("disabled queue changed", pending, budget, attempts, err)
	}
	if _, err = queue.SetEmbeddingEnabled(ctx, 2, true, "on", now); err != nil {
		t.Fatal(err)
	}
	claim, err = queue.Claim(ctx, "inference", "worker-b", now, time.Minute)
	if err != nil || claim.IdempotencyKey != "b" {
		t.Fatal("pending job did not resume", claim, err)
	}
	if _, err = queue.SetEmbeddingEnabled(ctx, 2, false, "stale", now); !errors.Is(err, ErrConflict) {
		t.Fatal("stale global change", err)
	}
	if _, err = queue.SetSourceEmbeddingEnabled(ctx, "unknown", 0, true, "actor", now); !errors.Is(err, ErrInvalid) {
		t.Fatal("unknown source", err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO registry_sources VALUES('new-source')`); err != nil {
		t.Fatal(err)
	}
	state, err = queue.SourceEmbeddingState(ctx, "new-source")
	if err != nil || state.Enabled {
		t.Fatal("new source enabled", state, err)
	}
	var events int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM processing_embedding_control_events`).Scan(&events); err != nil || events != 5 {
		t.Fatal(events, err)
	}
}
