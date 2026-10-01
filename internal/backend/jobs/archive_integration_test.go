//go:build integration

package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestArchivePreservesHistoryAndPreventsReplay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	pool, _ := queueTestDB(t)
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	s := New(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	request := func(key string, at time.Time) EnqueueRequest {
		return EnqueueRequest{Queue: key, Kind: "fixture", IdempotencyKey: key, Payload: json.RawMessage(`{}`), MaxAttempts: 3, RetryBase: time.Second, AvailableAt: at}
	}
	ids := map[string]int64{}
	for _, state := range []string{"failed", "retry_wait", "succeeded", "queued", "running"} {
		job, err := s.Enqueue(ctx, request(state, now))
		if err != nil {
			t.Fatal(err)
		}
		ids[state] = job.ID
		if state == "queued" {
			continue
		}
		claim, err := s.Claim(ctx, state, "fixture-worker", now, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if state == "running" {
			continue
		}
		if state == "succeeded" {
			err = s.Complete(ctx, claim, Result{Payload: json.RawMessage(`{"ok":true}`)}, now)
		} else {
			err = s.Fail(ctx, claim, Failure{Code: "fixture_error", Temporary: state == "retry_wait"}, now)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	fingerprint := func() string {
		t.Helper()
		var v string
		if err := pool.QueryRow(ctx, `SELECT md5((SELECT string_agg(row_to_json(a)::text,',' ORDER BY job_id,number) FROM processing_attempts a)||(SELECT row_to_json(j)::text FROM processing_jobs j WHERE id=$1))`, ids["succeeded"]).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	before := fingerprint()
	if _, err := s.ArchiveBefore(ctx, now, "operator", now); !errors.Is(err, ErrConflict) {
		t.Fatalf("running selection was not rejected: %v", err)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM processing_jobs WHERE archived_at IS NOT NULL").Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial archive: %d %v", count, err)
	}
	// Complete the in-flight job through its fenced claim, not by discarding it.
	var c Claim
	c.ID = ids["running"]
	c.Attempt = 1
	if err := pool.QueryRow(ctx, "SELECT claim_token FROM processing_jobs WHERE id=$1", c.ID).Scan(&c.Token); err != nil {
		t.Fatal(err)
	}
	if err := s.Complete(ctx, c, Result{Payload: json.RawMessage(`{}`)}, now); err != nil {
		t.Fatal(err)
	}
	before = fingerprint()
	fresh, err := s.Enqueue(ctx, request("new-evidence", now.Add(time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	counts, err := s.ArchiveBefore(ctx, now, "operator", now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if counts["failed"] != 1 || counts["retry_wait"] != 1 || counts["queued"] != 1 || len(counts) != 3 {
		t.Fatalf("incorrect scope: %v", counts)
	}
	if fingerprint() != before {
		t.Fatal("successful job or attempt history changed")
	}
	counts, err = s.ArchiveBefore(ctx, now, "operator", now.Add(time.Second))
	if err != nil || len(counts) != 0 {
		t.Fatalf("archive not idempotent: %v %v", counts, err)
	}
	for _, state := range []string{"failed", "retry_wait", "queued"} {
		j, err := s.Enqueue(ctx, request(state, now))
		if err != nil || j.ID != ids[state] {
			t.Fatalf("deduplication lost: %v %v", j, err)
		}
		if _, err = s.Claim(ctx, state, "fixture-worker", now.Add(time.Hour), time.Minute); !errors.Is(err, ErrNoJob) {
			t.Fatalf("archived job claimable: %s %v", state, err)
		}
	}
	if err = s.Relaunch(ctx, ids["failed"], 1, "operator", now); !errors.Is(err, ErrNoJob) {
		t.Fatalf("archived job relaunched: %v", err)
	}
	if _, err = pool.Exec(ctx, "UPDATE processing_jobs SET archived_at=NULL,archived_by=NULL WHERE id=$1", ids["failed"]); err == nil {
		t.Fatal("database allowed archive revival")
	}
	claim, err := s.Claim(ctx, "new-evidence", "fixture-worker", now.Add(time.Second), time.Minute)
	if err != nil || claim.ID != fresh.ID {
		t.Fatalf("new evidence blocked: %v %v", claim, err)
	}
}
