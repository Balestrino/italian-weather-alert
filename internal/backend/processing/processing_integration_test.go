//go:build integration

package processing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func processingTestDB(t *testing.T, maxConns ...int32) *pgxpool.Pool {
	t.Helper()
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	name := "iwa-processing-test-" + hex.EncodeToString(random[:6])
	password := hex.EncodeToString(random)
	passwordFile := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		output, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("isolated PostgreSQL operation failed: %v: %s", err, strings.ReplaceAll(string(output), password, "[redacted]"))
		}
		return strings.TrimSpace(string(output))
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	binding := listener.Addr().String() + ":5432"
	if err = listener.Close(); err != nil {
		t.Fatal(err)
	}
	run("run", "--pull=never", "--detach", "--name", name, "--publish", binding, "--mount", "type=bind,src="+passwordFile+",dst=/run/secrets/password,readonly", "--env", "POSTGRES_PASSWORD_FILE=/run/secrets/password", "--env", "POSTGRES_USER=iwa", "--env", "POSTGRES_DB=iwa", "postgres:16-alpine")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", name).Run() })
	address := run("port", name, "5432/tcp")
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatal(err)
	}
	portNumber, _ := strconv.Atoi(port)
	configuration, err := pgxpool.ParseConfig("postgres://iwa@localhost/iwa?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	configuration.ConnConfig.Host = host
	configuration.ConnConfig.Port = uint16(portNumber)
	configuration.ConnConfig.Password = password
	if len(maxConns) > 0 {
		configuration.MaxConns = maxConns[0]
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), configuration)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	for pool.Ping(ctx) != nil {
		if ctx.Err() != nil {
			t.Fatal("test PostgreSQL did not become ready")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return pool
}

func TestVersionedProcessingAccounting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := processingTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	store := New(pool)
	base := time.Date(2026, 9, 16, 14, 0, 0, 0, time.UTC)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	qwen := ModelVersion{ID: "regolo-qwen-fixture-v1", Provider: "regolo-fixture", Model: "qwen-fixture", Revision: "catalog-2026-09-16", Capabilities: json.RawMessage(`{"structured":true}`), CreatedAt: base}
	must(store.RegisterModel(ctx, qwen))
	must(store.RegisterModel(ctx, qwen))
	changedModel := qwen
	changedModel.Capabilities = json.RawMessage(`{"structured":false}`)
	if err := store.RegisterModel(ctx, changedModel); !errors.Is(err, ErrConflict) {
		t.Fatalf("changed model version accepted: %v", err)
	}
	ocrModel := ModelVersion{ID: "regolo-ocr-fixture-v1", Provider: "regolo-fixture", Model: "ocr-fixture", Revision: "catalog-2026-09-16", Capabilities: json.RawMessage(`{"images":true}`), CreatedAt: base}
	embeddingModel := ModelVersion{ID: "embedding-fixture-v1", Provider: "fixture", Model: "embedding-fixture", Revision: "1", Capabilities: json.RawMessage(`{"dimensions":3}`), CreatedAt: base}
	must(store.RegisterModel(ctx, ocrModel))
	must(store.RegisterModel(ctx, embeddingModel))

	prompt := PromptVersion{ID: "classification-prompt-v1", Name: "classification", Stage: "classification", Revision: "1", Body: "Classify untrusted fixture content.", CreatedAt: base}
	must(store.RegisterPrompt(ctx, prompt))
	must(store.RegisterPrompt(ctx, prompt))
	ocrPrompt := PromptVersion{ID: "ocr-prompt-v1", Name: "ocr", Stage: "ocr", Revision: "1", Body: "Read the retained fixture page.", CreatedAt: base}
	must(store.RegisterPrompt(ctx, ocrPrompt))

	qwenPrice := PriceVersion{ID: "qwen-price-v1", ModelVersionID: qwen.ID, Currency: "EUR", ProvenanceURL: "https://provider.example/pricing", ObservedAt: base, CreatedAt: base, Details: json.RawMessage(`{"kind":"published fixture"}`), Rates: []Rate{
		{Metric: "output_tokens", UnitSize: 1000, PriceMicrounits: 200},
		{Metric: "cache_write_tokens", UnitSize: 1000, PriceMicrounits: 50},
		{Metric: "input_tokens", UnitSize: 1000, PriceMicrounits: 100},
		{Metric: "cache_read_tokens", UnitSize: 1000, PriceMicrounits: 20},
	}}
	must(store.RegisterPrice(ctx, qwenPrice))
	must(store.RegisterPrice(ctx, qwenPrice))
	changedPrice := qwenPrice
	changedPrice.Rates = append([]Rate(nil), qwenPrice.Rates...)
	changedPrice.Rates[0].PriceMicrounits++
	if err := store.RegisterPrice(ctx, changedPrice); !errors.Is(err, ErrConflict) {
		t.Fatalf("changed pricing version accepted: %v", err)
	}
	ocrPrice := PriceVersion{ID: "ocr-price-v1", ModelVersionID: ocrModel.ID, Currency: "EUR", ProvenanceURL: "https://provider.example/ocr-pricing", ObservedAt: base, CreatedAt: base, Details: json.RawMessage(`{"kind":"published fixture"}`), Rates: []Rate{{Metric: "pages", UnitSize: 1, PriceMicrounits: 500}}}
	embeddingPrice := PriceVersion{ID: "embedding-price-v1", ModelVersionID: embeddingModel.ID, Currency: "EUR", ProvenanceURL: "https://provider.example/embedding-pricing", ObservedAt: base, CreatedAt: base, Details: json.RawMessage(`{"kind":"published fixture"}`), Rates: []Rate{{Metric: "input_tokens", UnitSize: 1000, PriceMicrounits: 10}}}
	must(store.RegisterPrice(ctx, ocrPrice))
	must(store.RegisterPrice(ctx, embeddingPrice))

	classificationConfig := ConfigurationVersion{ID: "classification-config-v1", Name: "classification", Stage: "classification", Revision: "1", ModelVersionID: &qwen.ID, PromptVersionID: &prompt.ID, LogicVersion: "fixture-code-v1", Settings: json.RawMessage(`{"temperature":0}`), CreatedAt: base}
	ocrConfig := ConfigurationVersion{ID: "ocr-config-v1", Name: "ocr", Stage: "ocr", Revision: "1", ModelVersionID: &ocrModel.ID, PromptVersionID: &ocrPrompt.ID, LogicVersion: "fixture-code-v1", Settings: json.RawMessage(`{"format":"page"}`), CreatedAt: base}
	embeddingConfig := ConfigurationVersion{ID: "embedding-config-v1", Name: "embedding", Stage: "embedding", Revision: "1", ModelVersionID: &embeddingModel.ID, LogicVersion: "fixture-code-v1", Settings: json.RawMessage(`{"dimensions":3}`), CreatedAt: base}
	collectionConfig := ConfigurationVersion{ID: "collection-config-v1", Name: "collection", Stage: "collection", Revision: "1", LogicVersion: "fixture-code-v1", Settings: json.RawMessage(`{"method":"fixture"}`), CreatedAt: base}
	for _, configuration := range []ConfigurationVersion{classificationConfig, ocrConfig, embeddingConfig, collectionConfig} {
		must(store.RegisterConfiguration(ctx, configuration))
		must(store.RegisterConfiguration(ctx, configuration))
	}
	wrongStage := classificationConfig
	wrongStage.ID, wrongStage.Name, wrongStage.Revision, wrongStage.Stage = "wrong-stage", "wrong-stage", "1", "ocr"
	if err := store.RegisterConfiguration(ctx, wrongStage); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cross-stage prompt accepted: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE processing_prompt_versions SET body='changed' WHERE id=$1", prompt.ID); err == nil {
		t.Fatal("immutable prompt was updated")
	}

	queue := jobs.New(pool)
	queued, err := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "processing", Kind: "fixture", IdempotencyKey: "bootstrap-classification", Payload: json.RawMessage(`{}`), MaxAttempts: 3, RetryBase: time.Second, AvailableAt: base})
	must(err)
	claim, err := queue.Claim(ctx, "processing", "fixture-worker", base, 10*time.Second)
	must(err)
	if claim.ID != queued.ID {
		t.Fatal("wrong queue job claimed")
	}
	run, err := store.StartRun(ctx, RunRequest{IdempotencyKey: "run-bootstrap-classification", Workload: "bootstrap", Stage: "classification", ConfigurationVersionID: classificationConfig.ID, Subject: json.RawMessage(`{"document":"fixture"}`), CreatedAt: base})
	must(err)
	replay, err := store.StartRun(ctx, RunRequest{IdempotencyKey: "run-bootstrap-classification", Workload: "bootstrap", Stage: "classification", ConfigurationVersionID: classificationConfig.ID, Subject: json.RawMessage(` {"document":"fixture"} `), CreatedAt: base.Add(time.Hour)})
	if err != nil || replay.ID != run.ID {
		t.Fatalf("run replay failed: %#v %v", replay, err)
	}
	if _, err = store.StartRun(ctx, RunRequest{IdempotencyKey: "run-bootstrap-classification", Workload: "bootstrap", Stage: "classification", ConfigurationVersionID: classificationConfig.ID, Subject: json.RawMessage(`{"document":"different"}`), CreatedAt: base}); !errors.Is(err, ErrConflict) {
		t.Fatalf("changed run reused idempotency key: %v", err)
	}
	first, err := store.StartAttempt(ctx, AttemptStart{RunID: run.ID, QueueJobID: &claim.ID, QueueAttemptNumber: &claim.Attempt, StartedAt: base})
	must(err)
	firstFinished, err := store.FinishAttempt(ctx, AttemptFinish{RunID: run.ID, Number: first.Number, FinishedAt: base.Add(1500 * time.Millisecond), Outcome: "failed", Usage: Usage{Status: "partial", InputTokens: integer(100)}, PriceVersionID: &qwenPrice.ID, ErrorCode: "temporary_provider"})
	must(err)
	if firstFinished.CostStatus != "unknown" || firstFinished.EstimatedCostMicrounits != nil {
		t.Fatal("partial usage produced a misleading total")
	}
	second, err := store.StartAttempt(ctx, AttemptStart{RunID: run.ID, QueueJobID: &claim.ID, QueueAttemptNumber: &claim.Attempt, StartedAt: base.Add(2 * time.Second)})
	must(err)
	secondFinished, err := store.FinishAttempt(ctx, AttemptFinish{RunID: run.ID, Number: second.Number, FinishedAt: base.Add(4500 * time.Millisecond), Outcome: "succeeded", Usage: Usage{Status: "reported", InputTokens: integer(1500), OutputTokens: integer(500), CacheReadTokens: integer(1000), CacheWriteTokens: integer(0)}, PriceVersionID: &qwenPrice.ID})
	must(err)
	tokensRecorded := secondFinished.InputTokens != nil && *secondFinished.InputTokens == 1500 &&
		secondFinished.OutputTokens != nil && *secondFinished.OutputTokens == 500 &&
		secondFinished.CacheReadTokens != nil && *secondFinished.CacheReadTokens == 1000 &&
		secondFinished.CacheWriteTokens != nil && *secondFinished.CacheWriteTokens == 0
	if secondFinished.EstimatedCostMicrounits == nil || *secondFinished.EstimatedCostMicrounits != 270 || secondFinished.DurationMS == nil || *secondFinished.DurationMS != 2500 || !tokensRecorded {
		t.Fatalf("wrong cost/timing: %#v", secondFinished)
	}
	if _, err = store.FinishAttempt(ctx, AttemptFinish{RunID: run.ID, Number: second.Number, FinishedAt: base.Add(5 * time.Second), Outcome: "succeeded", Usage: Usage{Status: "unavailable"}}); !errors.Is(err, ErrFinished) {
		t.Fatalf("finished attempt changed: %v", err)
	}

	ocrRun, err := store.StartRun(ctx, RunRequest{IdempotencyKey: "run-ordinary-ocr", Workload: "ordinary", Stage: "ocr", ConfigurationVersionID: ocrConfig.ID, Subject: json.RawMessage(`{"page":1}`), CreatedAt: base})
	must(err)
	ocrUnknown, err := store.StartAttempt(ctx, AttemptStart{RunID: ocrRun.ID, StartedAt: base})
	must(err)
	unknown, err := store.FinishAttempt(ctx, AttemptFinish{RunID: ocrRun.ID, Number: ocrUnknown.Number, FinishedAt: base.Add(time.Second), Outcome: "failed", Usage: Usage{Status: "unavailable"}, PriceVersionID: &ocrPrice.ID, ErrorCode: "provider_usage_missing"})
	must(err)
	if unknown.InputTokens != nil || unknown.OutputTokens != nil || unknown.EstimatedCostMicrounits != nil || unknown.UsageStatus != "unavailable" {
		t.Fatal("unknown provider usage was recorded as zero")
	}
	ocrReported, err := store.StartAttempt(ctx, AttemptStart{RunID: ocrRun.ID, StartedAt: base.Add(2 * time.Second)})
	must(err)
	ocrDone, err := store.FinishAttempt(ctx, AttemptFinish{RunID: ocrRun.ID, Number: ocrReported.Number, FinishedAt: base.Add(3 * time.Second), Outcome: "succeeded", Usage: Usage{Status: "reported", OtherUnits: map[string]int64{"pages": 2}}, PriceVersionID: &ocrPrice.ID})
	must(err)
	if ocrDone.EstimatedCostMicrounits == nil || *ocrDone.EstimatedCostMicrounits != 1000 {
		t.Fatal("OCR page cost not calculated")
	}

	collectionRun, err := store.StartRun(ctx, RunRequest{IdempotencyKey: "run-ordinary-collection", Workload: "ordinary", Stage: "collection", ConfigurationVersionID: collectionConfig.ID, Subject: json.RawMessage(`{"check":"fixture"}`), CreatedAt: base})
	must(err)
	collectionAttempt, err := store.StartAttempt(ctx, AttemptStart{RunID: collectionRun.ID, StartedAt: base})
	must(err)
	collectionDone, err := store.FinishAttempt(ctx, AttemptFinish{RunID: collectionRun.ID, Number: collectionAttempt.Number, FinishedAt: base.Add(10 * time.Millisecond), Outcome: "succeeded", Usage: Usage{Status: "not_applicable"}})
	must(err)
	if collectionDone.CostStatus != "not_applicable" {
		t.Fatal("ordinary collection was assigned inference cost")
	}

	embeddingRun, err := store.StartRun(ctx, RunRequest{IdempotencyKey: "run-ordinary-embedding", Workload: "ordinary", Stage: "embedding", ConfigurationVersionID: embeddingConfig.ID, Subject: json.RawMessage(`{"document":"fixture"}`), CreatedAt: base})
	must(err)
	embeddingAttempt, err := store.StartAttempt(ctx, AttemptStart{RunID: embeddingRun.ID, StartedAt: base})
	must(err)
	embeddingDone, err := store.FinishAttempt(ctx, AttemptFinish{RunID: embeddingRun.ID, Number: embeddingAttempt.Number, FinishedAt: base.Add(20 * time.Millisecond), Outcome: "succeeded", Usage: Usage{Status: "reported", InputTokens: integer(2000)}, PriceVersionID: &embeddingPrice.ID})
	must(err)
	if embeddingDone.EstimatedCostMicrounits == nil || *embeddingDone.EstimatedCostMicrounits != 20 {
		t.Fatal("embedding usage not separated")
	}

	loaded, err := store.Run(ctx, run.ID)
	must(err)
	if len(loaded.Attempts) != 2 || loaded.Attempts[0].Outcome != "failed" || loaded.Attempts[1].Outcome != "succeeded" {
		t.Fatalf("per-run attempts unavailable: %#v", loaded.Attempts)
	}
	summaries, err := store.Summaries(ctx)
	must(err)
	seen := map[string]Summary{}
	for _, summary := range summaries {
		seen[summary.Workload+"/"+summary.Stage+"/"+summary.Currency] = summary
	}
	classification := seen["bootstrap/classification/EUR"]
	if classification.Runs != 1 || classification.Attempts != 2 || classification.UnknownUsageAttempts != 1 || classification.UnknownCosts != 1 || classification.EstimatedCostMicrounits == nil || *classification.EstimatedCostMicrounits != 270 {
		t.Fatalf("bootstrap summary lost retries or unknowns: %#v", classification)
	}
	ocr := seen["ordinary/ocr/EUR"]
	if ocr.Attempts != 2 || ocr.UnknownUsageAttempts != 1 || ocr.UnknownCosts != 1 || ocr.EstimatedCostMicrounits == nil || *ocr.EstimatedCostMicrounits != 1000 {
		t.Fatalf("OCR summary incorrect: %#v", ocr)
	}
	if seen["ordinary/collection/"].UnknownCosts != 0 || seen["ordinary/embedding/EUR"].EstimatedCostMicrounits == nil {
		t.Fatalf("ordinary collection/embedding separation missing: %#v", seen)
	}
	var provenance string
	var observed time.Time
	if err = pool.QueryRow(ctx, "SELECT provenance_url,observed_at FROM processing_price_versions WHERE id=$1", qwenPrice.ID).Scan(&provenance, &observed); err != nil || provenance != qwenPrice.ProvenanceURL || !observed.Equal(base) {
		t.Fatalf("pricing provenance missing: %s %s %v", provenance, observed, err)
	}
}
