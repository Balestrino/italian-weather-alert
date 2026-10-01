//go:build integration

package interpretation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/evaluation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestProcessingComparisonGatesOnlyExplicitSelectedReprocessing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := interpretationTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, evaluation.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 9, 17, 14, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://calcinaia.example/policy", Locator: "retained fixture", ObservedAt: now.Add(-time.Hour)}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "calcinaia", Name: "Calcinaia", OfficialURL: "https://calcinaia.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "calcinaia-web", PublisherID: "calcinaia", Platform: "fixture", URL: "https://calcinaia.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "calcinaia", AuthorityID: "calcinaia", ChannelID: "calcinaia-web", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://calcinaia.example", Sections: []string{"https://calcinaia.example/notices"}, AccessMethod: "fixture", Attribution: "Comune", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	retained := documents.New(pool, &interpretationObjects{})
	versions := make([]documents.Version, 0, 2)
	for index, id := range []string{"partial-reopen", "ambiguous-link"} {
		url := "https://calcinaia.example/notices/" + id
		version, err := retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: "calcinaia", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "calcinaia", Configuration: 1, MediaType: "text/html", Bytes: []byte(id)}}})
		if err != nil {
			t.Fatal(err)
		}
		acquired := now.Add(time.Duration(index-2) * time.Hour)
		if _, err = pool.Exec(ctx, "UPDATE retained_versions SET first_acquired_at=$2 WHERE id=$1", version.ID, acquired); err != nil {
			t.Fatal(err)
		}
		versions = append(versions, version)
	}
	process := processing.New(pool)
	qwen := processing.ModelVersion{ID: "qwen-v1", Provider: "fixture", Model: "qwen3.8-27b", Revision: "verified", Capabilities: json.RawMessage(`{"structured":true}`), CreatedAt: now.Add(-time.Hour)}
	embed := processing.ModelVersion{ID: "embedding-v1", Provider: "fixture", Model: "Qwen3-Embedding-8B", Revision: "verified", Capabilities: json.RawMessage(`{"dimensions":1024}`), CreatedAt: now.Add(-time.Hour)}
	prompt := processing.PromptVersion{ID: "link-prompt-v1", Name: "linking", Stage: "linking", Revision: "v1", Body: "reviewed linking fixture", CreatedAt: now.Add(-time.Hour)}
	for _, err := range []error{process.RegisterModel(ctx, qwen), process.RegisterModel(ctx, embed), process.RegisterPrompt(ctx, prompt)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	linkConfig := processing.ConfigurationVersion{ID: "link-config-v1", Name: "linking", Stage: "linking", Revision: "v1", ModelVersionID: &qwen.ID, PromptVersionID: &prompt.ID, LogicVersion: "evidence-linking-v1", Settings: json.RawMessage(`{"semantic":false}`), CreatedAt: now.Add(-time.Hour)}
	embedConfig := processing.ConfigurationVersion{ID: "embed-config-v1", Name: "semantic", Stage: "embedding", Revision: "v1", ModelVersionID: &embed.ID, LogicVersion: "cosine-v1", Settings: json.RawMessage(`{"dimensions":1024}`), CreatedAt: now.Add(-time.Hour)}
	for _, config := range []processing.ConfigurationVersion{linkConfig, embedConfig} {
		if err := process.RegisterConfiguration(ctx, config); err != nil {
			t.Fatal(err)
		}
	}
	finish := func(key, stage, config string, versionID int64, duration time.Duration, input, output int64) int64 {
		t.Helper()
		source := "calcinaia"
		run, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: key, Workload: "evaluation", Stage: stage, ConfigurationVersionID: config, SourceID: &source, DocumentVersionID: &versionID, Subject: json.RawMessage(`{"retained_case":true}`), CreatedAt: now.Add(-30 * time.Minute)})
		if err != nil {
			t.Fatal(err)
		}
		attempt, err := process.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, StartedAt: now.Add(-20 * time.Minute)})
		if err != nil {
			t.Fatal(err)
		}
		usage := processing.Usage{Status: "reported", InputTokens: &input}
		if output >= 0 {
			usage.OutputTokens = &output
		}
		if _, err = process.FinishAttempt(ctx, processing.AttemptFinish{RunID: run.ID, Number: attempt.Number, FinishedAt: now.Add(-20 * time.Minute).Add(duration), Outcome: "succeeded", Usage: usage}); err != nil {
			t.Fatal(err)
		}
		return run.ID
	}
	comparison := evaluation.ProcessingComparison{ID: "semantic-linking-v1", CorpusSHA256: strings.Repeat("a", 64), Reviewer: "human reviewer", ReviewedAt: "2026-09-17", StartedAt: now.Add(-30 * time.Minute), FinishedAt: now.Add(-10 * time.Minute)}
	for index, version := range versions {
		baseline := finish("baseline-"+comparison.ID+string(rune('a'+index)), "linking", linkConfig.ID, version.ID, 120*time.Millisecond, 80, 20)
		embedding := finish("embedding-"+comparison.ID+string(rune('a'+index)), "embedding", embedConfig.ID, version.ID, 25*time.Millisecond, 17, -1)
		candidate := finish("candidate-"+comparison.ID+string(rune('a'+index)), "linking", linkConfig.ID, version.ID, 125*time.Millisecond, 90, 22)
		comparison.Cases = append(comparison.Cases, evaluation.ComparisonCase{ID: []string{"CAL-PARTIAL-REOPEN", "CAL-AMBIGUOUS-LINK"}[index], Kind: "simulation", Expected: "human-reviewed linking outcome", Evidence: "docs/prerequisiti-mvp/casi-valutazione.json", Baseline: evaluation.Outcome{Status: []string{"pass", "omission"}[index], Actual: []string{"correct scoped reopening", "candidate missing"}[index], Evidence: "retained baseline result", RunIDs: []int64{baseline}}, Candidate: evaluation.Outcome{Status: "pass", Actual: []string{"correct scoped reopening", "ambiguous relation remains unresolved"}[index], Evidence: "retained semantic result", RunIDs: []int64{embedding, candidate}}})
	}
	store := evaluation.New(pool)
	queue := jobs.New(pool)
	scheduler := New(pool, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, true)
	scheduler.Evaluations = store
	failed := comparison
	failed.ID = "semantic-linking-failed"
	failed.Cases = append([]evaluation.ComparisonCase(nil), comparison.Cases...)
	failed.Cases[1].Candidate.Status = "omission"
	failedReport, err := store.RecordComparison(ctx, "operator", failed)
	if err != nil || failedReport.Passed {
		t.Fatalf("failed comparison: %#v %v", failedReport, err)
	}
	if _, _, err = scheduler.Reprocess(ctx, Selection{SourceID: "calcinaia", Actor: "operator", ProcessingEvaluationID: failed.ID, At: now}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("failed evaluation authorized reprocessing: %v", err)
	}
	var jobsBefore int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM processing_jobs").Scan(&jobsBefore); err != nil || jobsBefore != 0 {
		t.Fatalf("evaluation created jobs: %d %v", jobsBefore, err)
	}
	report, err := store.RecordComparison(ctx, "operator", comparison)
	if err != nil || !report.Passed || report.Baseline.FailedCases != 1 || report.Candidate.PassedCases != 2 || len(report.Candidate.Configurations) != 2 || report.Candidate.ProcessVersion == report.Baseline.ProcessVersion {
		t.Fatalf("comparison report: %#v %v", report, err)
	}
	if replay, replayErr := store.RecordComparison(ctx, "operator", comparison); replayErr != nil || replay.Candidate.ProcessVersion != report.Candidate.ProcessVersion {
		t.Fatalf("idempotent comparison replay: %#v %v", replay, replayErr)
	}
	changed := comparison
	changed.Cases = append([]evaluation.ComparisonCase(nil), comparison.Cases...)
	changed.Cases[0].Candidate.Actual = "changed verdict"
	if _, changedErr := store.RecordComparison(ctx, "operator", changed); !errors.Is(changedErr, evaluation.ErrComparisonInvalid) {
		t.Fatalf("comparison identity was mutable: %v", changedErr)
	}
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM processing_jobs").Scan(&jobsBefore); err != nil || jobsBefore != 0 {
		t.Fatalf("recording a passed evaluation replayed history: %d %v", jobsBefore, err)
	}
	from := now.Add(-90 * time.Minute)
	requestID, selected, err := scheduler.Reprocess(ctx, Selection{SourceID: "calcinaia", From: &from, Actor: "operator", ProcessingEvaluationID: comparison.ID, At: now.Add(time.Minute)})
	if err != nil || selected != 1 {
		t.Fatalf("selected reprocessing: id=%s selected=%d err=%v", requestID, selected, err)
	}
	var evaluationID, processVersion string
	if err = pool.QueryRow(ctx, "SELECT processing_evaluation_id,candidate_process_version FROM interpretation_reprocessing_requests WHERE id=$1", requestID).Scan(&evaluationID, &processVersion); err != nil {
		t.Fatal(err)
	}
	if evaluationID != comparison.ID || processVersion != report.Candidate.ProcessVersion {
		t.Fatal("reprocessing did not retain the passed candidate identity")
	}
	var selectedVersion, queued, retainedRuns int64
	if err = pool.QueryRow(ctx, "SELECT document_version_id FROM interpretation_reprocessing_selections WHERE request_id=$1", requestID).Scan(&selectedVersion); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM processing_jobs").Scan(&queued); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM processing_runs WHERE workload='evaluation'").Scan(&retainedRuns); err != nil {
		t.Fatal(err)
	}
	if selectedVersion != versions[1].ID || queued != 1 || retainedRuns != 6 {
		t.Fatalf("scope/history changed: selected=%d queued=%d runs=%d", selectedVersion, queued, retainedRuns)
	}
}
