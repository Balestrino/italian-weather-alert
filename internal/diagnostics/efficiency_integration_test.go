//go:build integration

package diagnostics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEfficiencyReportFiltersReceiptsAndEstimates(t *testing.T) {
	ctx := context.Background()
	pool := diagnosticsTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, classification.Migrate} {
		if err := m(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Truncate(time.Second)
	s := processing.New(pool)
	reg := registry.New(pool)
	for _, source := range []string{"source-a", "source-b"} {
		evidence := registry.Evidence{URL: "https://fixture.example/policy", Locator: "fixture", ObservedAt: now}
		for _, err := range []error{reg.CreateAuthority(ctx, registry.Authority{ID: source, Name: source, OfficialURL: "https://fixture.example"}), reg.CreateChannel(ctx, registry.Channel{ID: source, PublisherID: source, Platform: "fixture", URL: "https://fixture.example"}), reg.CreateSource(ctx, registry.Source{ID: source, AuthorityID: source, ChannelID: source, ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://fixture.example", Sections: []string{"https://fixture.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test")} {
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	for i, scope := range []struct{ source, model, stage string }{{"source-a", "model-a", "classification"}, {"source-b", "model-a", "classification"}, {"source-a", "model-b", "classification"}, {"source-a", "model-a", "extraction"}} {
		model := scope.model
		config := fmt.Sprint("config-", i)
		if err := s.RegisterModel(ctx, processing.ModelVersion{ID: model, Provider: "fixture", Model: model, Revision: "1", Capabilities: json.RawMessage(`{}`), CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
		if err := s.RegisterConfiguration(ctx, processing.ConfigurationVersion{ID: config, Name: config, Stage: scope.stage, Revision: "1", LogicVersion: "1", ModelVersionID: &model, Settings: json.RawMessage(`{}`), CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
		run, err := s.StartRun(ctx, processing.RunRequest{IdempotencyKey: config, Workload: "evaluation", Stage: scope.stage, SourceID: &scope.source, ConfigurationVersionID: config, Subject: json.RawMessage(`{}`), CreatedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		a, err := s.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, StartedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		hash := strings.Repeat("a", 64)
		input, output := int64(100), int64(20)
		id, err := s.StartCall(ctx, processing.CallStart{RunID: run.ID, AttemptNumber: a.Number, Ordinal: 1, InputSHA256: hash, Provider: "fixture", RequestedModel: model, StartedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		if err = s.FinishCall(ctx, processing.CallFinish{ID: id, State: "received", FinishedAt: now, Usage: processing.Usage{Status: "reported", InputTokens: &input, OutputTokens: &output}}); err != nil {
			t.Fatal(err)
		}
		c := processing.SegmentCheckpoint{Stage: scope.stage, Configuration: config, InputSHA256: hash, CallID: id, Response: json.RawMessage(`{"Usage":{"InputTokens":100,"OutputTokens":20},"Content":"validated"}`)}
		if err = s.PutCheckpoint(ctx, c, now); err != nil {
			t.Fatal(err)
		}
		_, err = s.StartCall(ctx, processing.CallStart{RunID: run.ID, AttemptNumber: a.Number, Ordinal: 2, InputSHA256: hash, Provider: "fixture", RequestedModel: model, StartedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.FinishAttempt(ctx, processing.AttemptFinish{RunID: run.ID, Number: a.Number, FinishedAt: now, Outcome: "failed", ErrorCode: "classification_output_quotation", Usage: processing.Usage{Status: "unavailable"}}); err != nil {
			t.Fatal(err)
		}
		a, err = s.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, StartedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		if err = s.UseCheckpoint(ctx, c, a, 1, true); err != nil {
			t.Fatal(err)
		}
		if _, err = s.FinishAttempt(ctx, processing.AttemptFinish{RunID: run.ID, Number: a.Number, FinishedAt: now, Outcome: "succeeded", Usage: processing.Usage{Status: "not_applicable"}}); err != nil {
			t.Fatal(err)
		}
		// A legacy failed attempt carries usage but no verified provider-call count.
		a, err = s.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, StartedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		input, output = 11, 7
		if _, err = s.FinishAttempt(ctx, processing.AttemptFinish{RunID: run.ID, Number: a.Number, FinishedAt: now, Outcome: "failed", ErrorCode: "legacy_failure", Usage: processing.Usage{Status: "reported", InputTokens: &input, OutputTokens: &output}}); err != nil {
			t.Fatal(err)
		}
	}
	sql, err := os.ReadFile("../../scripts/inference-efficiency.sql")
	if err != nil {
		t.Fatal(err)
	}
	check := func(source, model, stage string, since, cutoff time.Time, want int) {
		t.Helper()
		tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		var raw []byte
		if err = tx.QueryRow(ctx, string(sql), since, cutoff, source, model, stage).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var report map[string][]map[string]any
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(raw, &fields)
		report = map[string][]map[string]any{}
		for _, key := range []string{"actual_calls", "reuse_estimates", "legacy_attempt_usage", "invalid_outputs"} {
			var rows []map[string]any
			if err = json.Unmarshal(fields[key], &rows); err != nil {
				t.Fatal(err)
			}
			report[key] = rows
		}
		for _, key := range []string{"actual_calls", "reuse_estimates", "legacy_attempt_usage", "invalid_outputs"} {
			if len(report[key]) != want {
				t.Fatalf("%s groups=%d want=%d: %s", key, len(report[key]), want, raw)
			}
		}
		for _, row := range report["actual_calls"] {
			if row["calls"] != float64(2) || row["unknown_or_partial_calls"] != float64(1) || row["known_input_tokens"] != float64(100) {
				t.Fatalf("actual usage: %s", raw)
			}
		}
		for _, row := range report["legacy_attempt_usage"] {
			if row["attempts"] != float64(1) || row["known_input_tokens"] != float64(11) {
				t.Fatalf("double counted legacy usage: %s", raw)
			}
		}
		for _, row := range report["reuse_estimates"] {
			if row["reused_units"] != float64(1) || row["estimated_avoided_input_tokens"] != float64(100) {
				t.Fatalf("reuse estimates: %s", raw)
			}
		}
	}
	check("", "", "", now.Add(-time.Second), now.Add(time.Hour), 4)
	check("source-a", "model-a", "classification", now.Add(-time.Second), now.Add(time.Hour), 1)
	check("source-b", "", "", now.Add(-time.Second), now.Add(time.Hour), 1)
	check("", "model-b", "", now.Add(-time.Second), now.Add(time.Hour), 1)
	check("", "", "extraction", now.Add(-time.Second), now.Add(time.Hour), 1)
	check("", "", "", now.Add(time.Hour), now.Add(2*time.Hour), 0)
	check("", "", "", now.Add(-time.Second), now, 0)

	// Additional independent OCR reuse and eligibility counters.
	url := "https://fixture.example/image.png"
	version := retainFixture(t, ctx, pool, "source-a", url, strings.Repeat("f", 64), now)
	catalog, err := ocr.RegisterCatalog(ctx, s, "fixture", "ocr-model", now)
	if err != nil {
		t.Fatal(err)
	}
	source := "source-a"
	run, err := s.StartRun(ctx, processing.RunRequest{IdempotencyKey: "ocr-reuse", Workload: "evaluation", Stage: "ocr", SourceID: &source, DocumentVersionID: &version, ConfigurationVersionID: catalog.ConfigurationVersionID, Subject: json.RawMessage(`{}`), CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	store := ocr.NewStore(pool)
	hash := strings.Repeat("f", 64)
	artifact, err := store.ClaimArtifact(ctx, ocr.ArtifactIdentity{Scope: "fixture", Model: "ocr-model", Configuration: catalog.ConfigurationVersionID, Renderer: "fixture", InputSHA256: hash, RequestSHA256: hash}, now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	input, output := int64(300), int64(40)
	digest := sha256.Sum256([]byte("hello"))
	page := ocr.PageResult{RunID: run.ID, DocumentVersionID: version, PageNumber: 1, ResourceURL: url, Status: "complete", MediaType: "image/png", InputSHA256: hash, OutputSHA256: hex.EncodeToString(digest[:]), ExtractedText: "hello", ProviderResponseID: "original", ReturnedModel: "ocr-model", InputTokens: &input, OutputTokens: &output, CreatedAt: now}
	if err = store.CompleteArtifact(ctx, artifact, page, now); err != nil {
		t.Fatal(err)
	}
	if err = store.AssociateArtifact(ctx, artifact.Key, page, true); err != nil {
		t.Fatal(err)
	}
	skipped, err := s.StartRun(ctx, processing.RunRequest{IdempotencyKey: "ocr-skip", Workload: "evaluation", Stage: "ocr", SourceID: &source, DocumentVersionID: &version, ConfigurationVersionID: catalog.ConfigurationVersionID, Subject: json.RawMessage(`{}`), CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if err = store.PutResource(ctx, ocr.ResourceResult{RunID: skipped.ID, DocumentVersionID: version, ResourceURL: url, Status: "skipped", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO processing_provider_gates(scope,model,state,reason,available_at,updated_at) VALUES('fixture','*','held','quota',$1,$1)`, now); err != nil {
		t.Fatal(err)
	}
	var raw []byte
	if err = pool.QueryRow(ctx, string(sql), now.Add(-time.Second), now.Add(time.Hour), "source-a", "ocr-model", "ocr").Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var report struct {
		Reuse []struct {
			Count int `json:"reused_units"`
			Input int `json:"estimated_avoided_input_tokens"`
		} `json:"reuse_estimates"`
		Skips []struct {
			Count int `json:"skipped_resources"`
		} `json:"skipped_resources"`
		Holds []any `json:"provider_holds_current_global"`
	}
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Reuse) != 1 || report.Reuse[0].Count != 1 || report.Reuse[0].Input != 300 || len(report.Skips) != 1 || report.Skips[0].Count != 1 || len(report.Holds) != 1 {
		t.Fatalf("OCR and global gate counts: %s", raw)
	}
}
