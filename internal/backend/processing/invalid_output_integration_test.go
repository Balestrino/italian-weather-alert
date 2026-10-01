//go:build integration

package processing

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestInvalidOutputCaptureExtractionAndStageIsolation(t *testing.T) {
	ctx := context.Background()
	pool := processingTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	store := New(pool)
	now := time.Now().UTC()
	for _, stage := range []string{"extraction", "ocr"} {
		if err := store.RegisterConfiguration(ctx, ConfigurationVersion{ID: stage, Name: stage, Stage: stage, Revision: "1", LogicVersion: "fixture", Settings: json.RawMessage(`{}`), CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
		run, err := store.StartRun(ctx, RunRequest{IdempotencyKey: stage, Workload: "evaluation", Stage: stage, ConfigurationVersionID: stage, Subject: json.RawMessage(`{}`), CreatedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		attempt, err := store.StartAttempt(ctx, AttemptStart{RunID: run.ID, StartedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		capture := InvalidOutput{RunID: run.ID, AttemptNumber: attempt.Number, SegmentOrdinal: 1, Request: []byte(`{"model":"fixture"}`), Response: []byte("invalid\x00output"), FinishReason: "stop", ErrorCode: "extraction_output_schema", CreatedAt: now}
		err = store.RecordInvalidOutput(ctx, capture)
		if stage == "ocr" {
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("unrelated stage accepted: %v", err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if err = store.RecordInvalidOutput(ctx, capture); err != nil {
			t.Fatal("idempotent capture", err)
		}
		altered := capture
		altered.Response = []byte("changed")
		if err = store.RecordInvalidOutput(ctx, altered); !errors.Is(err, ErrConflict) {
			t.Fatal("capture overwritten", err)
		}
		var response []byte
		if err = pool.QueryRow(ctx, `SELECT response_bytes FROM processing_invalid_outputs WHERE run_id=$1`, run.ID).Scan(&response); err != nil || string(response) != string(capture.Response) {
			t.Fatal("verbatim response lost", err)
		}
	}
}
