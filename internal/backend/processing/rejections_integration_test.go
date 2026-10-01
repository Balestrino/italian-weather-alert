//go:build integration

package processing

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
)

func TestHeldNewJobsDoNotConsumeAttemptsAndRejectionsAreExplicit(t *testing.T) {
	ctx := context.Background()
	pool := processingTestDB(t)
	if err := jobs.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, gatesSchema+rejectionsSchema); err != nil {
		t.Fatal(err)
	}
	s := New(pool)
	now := time.Now().UTC()
	policy := DefaultGatePolicy()
	p, err := s.AcquireGate(ctx, "account", "ocr", now, policy)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.RecordGate(ctx, p, "quota", 0, now, policy); err != nil {
		t.Fatal(err)
	}
	q := jobs.New(pool)
	job, err := q.Enqueue(ctx, jobs.EnqueueRequest{Queue: "inference", Kind: "ocr", IdempotencyKey: "new-version", Payload: json.RawMessage(`{}`), MaxAttempts: 3, RetryBase: time.Second, AvailableAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.DeferHeldJobs(ctx, []GateBinding{{Kind: "ocr", Scope: "account", Model: "ocr"}}, now); err != nil {
		t.Fatal(err)
	}
	if _, err = q.Claim(ctx, "inference", "worker", now, time.Minute); err != jobs.ErrNoJob {
		t.Fatal("new version bypassed hold")
	}
	var count int
	if err = pool.QueryRow(ctx, "SELECT attempt_count FROM processing_jobs WHERE id=$1", job.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal("held job consumed attempt")
	}
	counts, archiveErr := q.ArchiveBefore(ctx, now, "operator", now)
	if archiveErr != nil || counts["queued"] != 1 {
		t.Fatalf("archive held job: %v %v", counts, archiveErr)
	}
	if err = s.DeferHeldJobs(ctx, []GateBinding{{Kind: "ocr", Scope: "account", Model: "ocr"}}, now.Add(time.Second)); err != nil {
		t.Fatalf("hold tried to mutate archived job: %v", err)
	}
	r := Rejection{Scope: "account", Model: "ocr", ConfigurationID: "v1", InputSHA256: strings.Repeat("a", 64), ErrorCode: "provider_rejected", HTTPStatus: 400, Category: "request", CreatedAt: now}
	if err = s.Reject(ctx, r); err != nil {
		t.Fatal(err)
	}
	if _, found, err := s.Rejected(ctx, r.Scope, r.Model, r.ConfigurationID, r.InputSHA256); err != nil || !found {
		t.Fatal("permanent rejection lost")
	}
	if _, found, err := s.Rejected(ctx, r.Scope, r.Model, "v2", r.InputSHA256); err != nil || found {
		t.Fatal("corrected configuration still blocked")
	}
	if err = s.ClearRejection(ctx, r.Scope, r.Model, r.ConfigurationID, r.InputSHA256, "operator", now); err != nil {
		t.Fatal(err)
	}
	if _, found, err := s.Rejected(ctx, r.Scope, r.Model, r.ConfigurationID, r.InputSHA256); err != nil || found {
		t.Fatal("explicit recovery did not clear input")
	}
}
