//go:build integration

package processing

import (
	"context"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
	"strings"
	"testing"
	"time"
)

func TestCheckpointReceiptIdentityAndPaidUsage(t *testing.T) {
	ctx := context.Background()
	pool := processingTestDB(t)
	for _, m := range []func(context.Context) error{func(c context.Context) error { return registry.Migrate(c, pool) }, func(c context.Context) error { return documents.Migrate(c, pool) }, func(c context.Context) error { return jobs.Migrate(c, pool) }, func(c context.Context) error { return Migrate(c, pool) }} {
		if err := m(ctx); err != nil {
			t.Fatal(err)
		}
	}
	s := New(pool)
	now := time.Now().UTC()
	for _, stage := range []string{"classification", "extraction"} {
		config := stage + "-checkpoint"
		if err := s.RegisterConfiguration(ctx, ConfigurationVersion{ID: config, Name: config, Stage: stage, Revision: "1", LogicVersion: "test", Settings: json.RawMessage(`{}`), CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
		run, err := s.StartRun(ctx, RunRequest{IdempotencyKey: config, Workload: "evaluation", Stage: stage, ConfigurationVersionID: config, Subject: json.RawMessage(`{}`), CreatedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		a, err := s.StartAttempt(ctx, AttemptStart{RunID: run.ID, StartedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		hash := strings.Repeat("a", 64)
		id, err := s.StartCall(ctx, CallStart{RunID: run.ID, AttemptNumber: a.Number, Ordinal: 1, InputSHA256: hash, Provider: "fixture", RequestedModel: "model", StartedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		c := SegmentCheckpoint{Stage: stage, Configuration: config, InputSHA256: hash, CallID: id, Response: json.RawMessage(`{"Content":"validated"}`)}
		if err = s.PutCheckpoint(ctx, c, now); err != ErrInvalid {
			t.Fatalf("pending receipt accepted: %v", err)
		}
		n := int64(10)
		if err = s.FinishCall(ctx, CallFinish{ID: id, State: "received", FinishedAt: now, Usage: Usage{Status: "reported", InputTokens: &n, OutputTokens: &n}}); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			if err = s.PutCheckpoint(ctx, c, now); err != nil {
				t.Fatal(err)
			}
		}
		if err = s.UseCheckpoint(ctx, c, a, 1, false); err != nil {
			t.Fatal(err)
		}
		// A later failed call still contributes paid tokens.
		id, err = s.StartCall(ctx, CallStart{RunID: run.ID, AttemptNumber: a.Number, Ordinal: 2, InputSHA256: strings.Repeat("b", 64), Provider: "fixture", RequestedModel: "model", StartedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		if err = s.FinishCall(ctx, CallFinish{ID: id, State: "failed", ErrorCode: "provider_error", FinishedAt: now, Usage: Usage{Status: "reported", InputTokens: &n, OutputTokens: &n}}); err != nil {
			t.Fatal(err)
		}
		finished, err := s.FinishAttempt(ctx, AttemptFinish{RunID: run.ID, Number: a.Number, Outcome: "failed", ErrorCode: "provider_error", FinishedAt: now, Usage: Usage{Status: "unavailable"}})
		if err != nil || finished.InputTokens == nil || *finished.InputTokens != 20 {
			t.Fatalf("paid attempts lost: %#v %v", finished, err)
		}
		a, err = s.StartAttempt(ctx, AttemptStart{RunID: run.ID, StartedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		stored, ok, err := s.Checkpoint(ctx, c)
		if err != nil || !ok || stored.CallID != c.CallID {
			t.Fatal("checkpoint lost receipt")
		}
		if err = s.UseCheckpoint(ctx, c, a, 1, true); err != nil {
			t.Fatal(err)
		}
		if _, found, err := s.CallUsage(ctx, run.ID, a.Number); err != nil || found {
			t.Fatal("reuse counted as paid call")
		}
		for _, changed := range []SegmentCheckpoint{{Stage: stage, Configuration: config, InputSHA256: strings.Repeat("b", 64)}, {Stage: stage, Configuration: config + "-new", InputSHA256: hash}} {
			if _, ok, err := s.Checkpoint(ctx, changed); err != nil || ok {
				t.Fatal("incompatible checkpoint hit")
			}
		}
		if err = s.UseCheckpoint(ctx, c, a, 1, false); err != ErrInvalid {
			t.Fatal("conflicting provenance accepted")
		}
		if _, err = pool.Exec(ctx, `UPDATE processing_segment_checkpoints SET response='{}' WHERE stage=$1`, stage); err == nil {
			t.Fatal("checkpoint mutable")
		}
	}
}
