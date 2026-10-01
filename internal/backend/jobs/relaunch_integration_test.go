//go:build integration

package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

func TestRelaunchMigratesExistingQueueAndSurvivesRestart(t *testing.T) {
	ctx := context.Background()
	pool, restart := queueTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	// Install the unchanged original migration and an already exhausted job.
	_, err := pool.Exec(ctx, `CREATE TABLE iwa_migrations(name text PRIMARY KEY,checksum text NOT NULL);`+schema)
	must(err)
	hash := sha256.Sum256([]byte(schema))
	_, err = pool.Exec(ctx, `INSERT INTO iwa_migrations VALUES('003_jobs',$1)`, hex.EncodeToString(hash[:]))
	must(err)
	at := time.Now().UTC().Truncate(time.Microsecond)
	body, digest, err := canonicalObject([]byte(`{}`))
	must(err)
	var id int64
	must(pool.QueryRow(ctx, `INSERT INTO processing_jobs(queue,kind,idempotency_key,payload,payload_hash,state,max_attempts,retry_base_ms,attempt_count,available_at,created_at,updated_at,completed_at) VALUES('fixture','fixture','old',$1,$2,'failed',100,1000,100,$3,$3,$3,$3) RETURNING id`, body, digest, at).Scan(&id))
	_, err = pool.Exec(ctx, `INSERT INTO processing_attempts(job_id,number,worker_id,claim_token,started_at,lease_expires_at,finished_at,outcome,error_code) SELECT $1,n,'old-worker','token-'||n,$2::timestamptz,$2::timestamptz+interval '1 minute',$2::timestamptz,'failed','fixture' FROM generate_series(1,100) n`, id, at)
	must(err)
	must(Migrate(ctx, pool))
	must(Migrate(ctx, pool))
	queue := New(pool)
	must(queue.Relaunch(ctx, id, 100, "operator", at))
	restart()
	job, err := queue.Enqueue(ctx, EnqueueRequest{Queue: "fixture", Kind: "fixture", IdempotencyKey: "old", Payload: body, MaxAttempts: 100, RetryBase: time.Second, AvailableAt: at})
	must(err)
	if job.MaxAttempts != 200 || job.Attempt != 100 || job.State != "queued" {
		t.Fatalf("migrated budget: %#v", job)
	}
	claim, err := queue.Claim(ctx, "fixture", "new-worker", time.Now(), time.Minute)
	must(err)
	if claim.Attempt != 101 {
		t.Fatal("renumbered existing history")
	}
	must(queue.Fail(ctx, claim, Failure{Code: "permanent_fixture", Temporary: false}, time.Now()))
	// Replaying the earlier request cannot start another cycle after a new failure.
	must(queue.Relaunch(ctx, id, 100, "operator", time.Now()))
	if _, err = queue.Claim(ctx, "fixture", "new-worker", time.Now(), time.Minute); !errors.Is(err, ErrNoJob) {
		t.Fatalf("stale relaunch restarted work: %v", err)
	}
	must(queue.Relaunch(ctx, id, 101, "operator", time.Now()))
	claim, err = queue.Claim(ctx, "fixture", "new-worker", time.Now(), time.Minute)
	must(err)
	if claim.Attempt != 102 || claim.MaxAttempts != 201 {
		t.Fatal("second explicit budget incorrect")
	}
	history, err := queue.Attempts(ctx, id)
	must(err)
	if len(history) != 102 || history[0].ErrorCode != "fixture" {
		t.Fatal("migrated attempt history lost")
	}
}
