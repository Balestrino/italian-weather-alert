//go:build integration

package interpretation

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

type noLocalClassificationCalls struct{ calls int }

func (*noLocalClassificationCalls) Name() string { return "fixture" }
func (a *noLocalClassificationCalls) Complete(context.Context, inference.Request) (inference.Response, error) {
	a.calls++
	return inference.Response{}, inference.ErrInvalid
}

func TestStructuredClassificationPersistsReusesAndSchedulesExtraction(t *testing.T) {
	ctx := context.Background()
	pool := interpretationTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	url := "https://regional.example/monitoring"
	evidence := registry.Evidence{URL: "https://regional.example/review", Locator: "Synthetic explicit monitoring no-event format review", ObservedAt: now}
	policy := &registry.LocalProcessing{Version: registry.LocalProcessingVersion, Evidence: evidence, StructuredFormat: "cfr-monitoring-v1"}
	reg := registry.New(pool)
	cfg := registry.Configuration{URL: url, Sections: []string{url}, AccessMethod: "crawl4ai-html-pdf", Attribution: "Regional fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}, RegionalProduct: &registry.RegionalProductContract{Kind: "monitoring", ContentMarkers: []string{"NESSUN AVVISO"}}, LocalProcessing: policy}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "authority", Name: "Fixture", OfficialURL: "https://regional.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "channel", PublisherID: "authority", Platform: "fixture", URL: "https://regional.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "cfr-monitoring", AuthorityID: "authority", ChannelID: "channel", ProductID: "monitoring", Territory: "09"}, cfg, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	retained := documents.New(pool, &interpretationObjects{})
	v, err := retained.Retain(ctx, documents.Acquisition{ID: "monitoring-fixture", SourceID: "cfr-monitoring", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "cfr-monitoring", Configuration: 1, MediaType: "text/html", Bytes: []byte(`<b>NESSUN AVVISO IN CORSO DI VALIDITÀ O EVENTO IN CORSO</b>`)}}})
	if err != nil {
		t.Fatal(err)
	}
	process := processing.New(pool)
	catalog, err := classification.RegisterCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	local, err := process.RegisterLocalProcessing(ctx, catalog.ConfigurationVersionID, now)
	if err != nil {
		t.Fatal(err)
	}
	adapter := &noLocalClassificationCalls{}
	results := classification.NewStore(pool)
	runner := &classification.Runner{Documents: retained, OCR: ocr.NewStore(pool), Processing: process, Results: results, Manifests: results, ReuseSources: map[string]bool{"cfr-monitoring": true}, Adapter: &inference.RecordedAdapter{Adapter: adapter, Ledger: process}, Model: "qwen3.8-27b", ConfigurationVersion: catalog.ConfigurationVersionID, LocalConfigurationVersion: local, PriceVersion: catalog.PriceVersionID}
	queue := jobs.New(pool)
	retry := inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}
	var last classification.Result
	for _, selection := range []string{"first", "second"} {
		job, err := classification.Enqueue(ctx, queue, retry, classification.Payload{DocumentVersionID: v.ID, Workload: "reprocessing", SelectionID: selection}, now)
		if err != nil {
			t.Fatal(err)
		}
		claim, err := queue.Claim(ctx, inference.Queue, "local-classification", now, time.Minute)
		if err != nil || claim.ID != job.ID {
			t.Fatal("classification claim", err)
		}
		output, err := runner.Handler()(ctx, claim.Job)
		if err != nil {
			t.Fatal(err)
		}
		if err = queue.Complete(ctx, claim, output, now.Add(time.Second)); err != nil {
			t.Fatal(err)
		}
		var runID int64
		if err = pool.QueryRow(ctx, `SELECT run_id FROM classification_results ORDER BY run_id DESC LIMIT 1`).Scan(&runID); err != nil {
			t.Fatal(err)
		}
		last, _, err = results.Get(ctx, runID)
		if err != nil || last.Relevant == nil || !*last.Relevant || last.ReturnedModel != classification.StructuredClassifierVersion || last.InputTokens != nil {
			t.Fatal("durable local result", last, err)
		}
	}
	var reused, calls, prices int
	for _, q := range []struct {
		sql string
		out *int
	}{{`SELECT count(*) FROM interpretation_reuse`, &reused}, {`SELECT count(*) FROM processing_provider_calls`, &calls}, {`SELECT count(*) FROM processing_run_attempts WHERE price_version_id IS NOT NULL`, &prices}} {
		if err = pool.QueryRow(ctx, q.sql).Scan(q.out); err != nil {
			t.Fatal(err)
		}
	}
	if reused != 1 || calls != 0 || prices != 0 || adapter.calls != 0 {
		t.Fatal("local reuse/accounting", reused, calls, prices, adapter.calls)
	}
	scheduler := New(pool, queue, retry, false)
	if scheduled, err := scheduler.AfterClassification(ctx, last, "reprocessing", now); err != nil || !scheduled {
		t.Fatal("downstream extraction not admitted", err)
	}
	var count int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM processing_jobs WHERE kind=$1 AND (payload->>'classification_run_id')::bigint=$2`, extraction.Kind, last.RunID).Scan(&count); err != nil || count != 1 {
		t.Fatal("extraction provenance", count, err)
	}
	// Removing the shortcut changes both preflight configuration and manifest even
	// though the same full document remains available and existing results survive.
	version, err := retained.Version(ctx, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	p := &Preflight{Documents: retained, Configuration: "test", RendererIdentity: "renderer"}
	before, err := p.Build(ctx, version)
	if err != nil {
		t.Fatal(err)
	}
	content, err := classification.GatherContent(ctx, retained, ocr.NewStore(pool), version)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := classification.BuildManifest(ctx, retained, ocr.NewStore(pool), version, content, local)
	if err != nil {
		t.Fatal(err)
	}
	version.Resources[0].LocalProcessing = nil
	after, err := p.Build(ctx, version)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := classification.BuildManifest(ctx, retained, ocr.NewStore(pool), version, content, local)
	if err != nil {
		t.Fatal(err)
	}
	if before.Configuration == after.Configuration || manifest.Hash == changed.Hash {
		t.Fatal("policy-removal cache boundary")
	}
}
