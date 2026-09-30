//go:build integration

package processing

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

func TestProviderCallReceiptsAndUnknownUsage(t *testing.T) {
	ctx := context.Background()
	pool := processingTestDB(t)
	for _, m := range []func(context.Context) error{
		func(c context.Context) error { return registry.Migrate(c, pool) }, func(c context.Context) error { return documents.Migrate(c, pool) },
		func(c context.Context) error { return jobs.Migrate(c, pool) }, func(c context.Context) error { return Migrate(c, pool) },
	} {
		if err := m(ctx); err != nil {
			t.Fatal(err)
		}
	}
	s := New(pool)
	now := time.Now().UTC()
	if err := s.RegisterConfiguration(ctx, ConfigurationVersion{ID: "calls-test", Name: "calls", Stage: "classification", Revision: "1", LogicVersion: "test", Settings: json.RawMessage(`{}`), CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	run, err := s.StartRun(ctx, RunRequest{IdempotencyKey: "calls-test", Workload: "evaluation", Stage: "classification", ConfigurationVersionID: "calls-test", Subject: json.RawMessage(`{}`), CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	attempt, err := s.StartAttempt(ctx, AttemptStart{RunID: run.ID, StartedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if _, found, err := s.CallUsage(ctx, run.ID, attempt.Number); err != nil || found {
		t.Fatal("legacy attempt treated as ledger")
	}
	start := CallStart{RunID: run.ID, AttemptNumber: attempt.Number, Ordinal: 1, InputSHA256: strings.Repeat("a", 64), Provider: "fixture", RequestedModel: "model", StartedAt: now}
	id, err := s.StartCall(ctx, start)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.StartCall(ctx, start); err == nil {
		t.Fatal("duplicate call ordinal accepted")
	}
	input, output := int64(17), int64(4)
	finish := CallFinish{ID: id, FinishedAt: now.Add(time.Second), State: "received", HTTPStatus: 200, Usage: Usage{Status: "reported", InputTokens: &input, OutputTokens: &output}}
	if err = s.FinishCall(ctx, finish); err != nil {
		t.Fatal(err)
	}
	if err = s.FinishCall(ctx, finish); err != ErrFinished {
		t.Fatalf("receipt mutable: %v", err)
	}
	start.Ordinal = 2
	id, err = s.StartCall(ctx, start)
	if err != nil {
		t.Fatal(err)
	}
	u, found, err := s.CallUsage(ctx, run.ID, attempt.Number)
	if err != nil || !found || u.Status != "partial" || u.InputTokens == nil || *u.InputTokens != 17 {
		t.Fatalf("unknown call treated as zero: %#v %v", u, err)
	}
	if err = s.ResolveInterruptedCall(ctx, id, now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err = Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	u, _, err = s.CallUsage(ctx, run.ID, attempt.Number)
	if err != nil || u.Status != "partial" || *u.OutputTokens != 4 {
		t.Fatalf("repeat migration changed accounting: %#v %v", u, err)
	}
	wrong := int64(999)
	finished, err := s.FinishAttempt(ctx, AttemptFinish{RunID: run.ID, Number: attempt.Number, FinishedAt: now.Add(3 * time.Second), Outcome: "failed", ErrorCode: "classification_output_invalid", Usage: Usage{Status: "reported", InputTokens: &wrong}})
	if err != nil || finished.InputTokens == nil || *finished.InputTokens != 17 || finished.UsageStatus != "partial" {
		t.Fatalf("receipts not authoritative: %#v %v", finished, err)
	}
	queue := jobs.New(pool)
	job, err := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "inference", Kind: "fixture", IdempotencyKey: "interrupted", Payload: json.RawMessage(`{}`), MaxAttempts: 3, RetryBase: time.Second, AvailableAt: now})
	if err != nil {
		t.Fatal(err)
	}
	claim, err := queue.Claim(ctx, "inference", "owner", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	no := claim.Attempt
	a2, err := s.StartAttempt(ctx, AttemptStart{RunID: run.ID, QueueJobID: &job.ID, QueueAttemptNumber: &no, StartedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	start.AttemptNumber = a2.Number
	start.Ordinal = 1
	id, err = s.StartCall(ctx, start)
	if err != nil {
		t.Fatal(err)
	}
	finish.ID = id
	if err = s.FinishCall(ctx, finish); err != nil {
		t.Fatal(err)
	}
	if err = s.RecoverInterruptedAttempts(ctx, run.ID, now.Add(4*time.Second)); err != nil {
		t.Fatal(err)
	}
	active, err := s.Attempt(ctx, run.ID, a2.Number)
	if err != nil || active.Outcome != "running" {
		t.Fatal("active owner prematurely recovered")
	}
	if err = queue.Fail(ctx, claim, jobs.Failure{Code: "test_shutdown", Detail: "fixture", Temporary: false}, now.Add(5*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err = s.RecoverInterruptedAttempts(ctx, run.ID, now.Add(6*time.Second)); err != nil {
		t.Fatal(err)
	}
	recovered, err := s.Attempt(ctx, run.ID, a2.Number)
	if err != nil || recovered.Outcome != "failed" || recovered.InputTokens == nil || *recovered.InputTokens != 17 || recovered.UsageStatus != "reported" {
		t.Fatalf("durable receipt lost on recovery: %#v %v", recovered, err)
	}
}
