//go:build integration

package ocr

import (
	"context"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNativePDFPersistsWithoutProviderCalls(t *testing.T) {
	ctx := context.Background()
	pool := ocrTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	evidence := registry.Evidence{URL: "https://municipal.example/review", Locator: "Synthetic text-only document review", ObservedAt: now}
	reg := registry.New(pool)
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "authority", Name: "Fixture", OfficialURL: "https://municipal.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "channel", PublisherID: "authority", Platform: "fixture", URL: "https://municipal.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "source", AuthorityID: "authority", ChannelID: "channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://municipal.example", Sections: []string{"https://municipal.example/notices"}, AccessMethod: "fixture", Attribution: "Fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}, LocalProcessing: &registry.LocalProcessing{Version: registry.LocalProcessingVersion, Evidence: evidence, TextPDFPathPrefixes: []string{"/orders/"}}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	retained := documents.New(pool, &integrationObjects{})
	url := "https://municipal.example/orders/notice.pdf"
	v, err := retained.Retain(ctx, documents.Acquisition{ID: "native-fixture", SourceID: "source", Configuration: 1, URL: url, Metadata: []byte(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "source", Configuration: 1, MediaType: "application/pdf", Bytes: []byte("synthetic native document")}}})
	if err != nil {
		t.Fatal(err)
	}
	v, err = retained.Version(ctx, v.ID)
	if err != nil || v.Resources[0].LocalProcessing == nil {
		t.Fatal("reviewed policy not loaded", err)
	}
	process := processing.New(pool)
	base, err := RegisterCatalog(ctx, process, "openai-chat", "deepseek-ocr-2", now)
	if err != nil {
		t.Fatal(err)
	}
	local, err := process.RegisterLocalProcessing(ctx, base.ConfigurationVersionID, now)
	if err != nil {
		t.Fatal(err)
	}
	again, err := process.RegisterLocalProcessing(ctx, base.ConfigurationVersionID, now.Add(time.Hour))
	if err != nil || local != again {
		t.Fatal("local catalog not immutable/idempotent", err)
	}
	adapter := &fakeAdapter{}
	reader := &fakeTextExtractor{pages: []TextPage{{1, "Il Comune dispone la chiusura del ponte per rischio idraulico."}, {2, "La disposizione resta valida fino alla conclusione dell'emergenza."}}}
	runner := &Runner{Documents: retained, Processing: process, Results: NewStore(pool), Adapter: &inference.RecordedAdapter{Adapter: adapter, Ledger: process}, Renderer: fakeRenderer{}, TextExtractor: reader, Model: "deepseek-ocr-2", ConfigurationVersion: base.ConfigurationVersionID, LocalConfigurationVersion: local, PriceVersion: base.PriceVersionID}
	queue := jobs.New(pool)
	if _, err := Enqueue(ctx, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, Payload{DocumentVersionID: v.ID, ResourceURL: url, Workload: "evaluation"}, now); err != nil {
		t.Fatal(err)
	}
	claim, err := queue.Claim(ctx, inference.Queue, "native-test", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Handler()(ctx, claim.Job); err != nil {
		t.Fatal(err)
	}
	pages, err := NewStore(pool).PagesForVersion(ctx, v.ID)
	if err != nil || len(pages) != 2 {
		t.Fatal("durable pages", pages, err)
	}
	var calls, priced int
	var config, usage string
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM processing_provider_calls`).Scan(&calls); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM processing_run_attempts WHERE price_version_id IS NOT NULL`).Scan(&priced); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT r.configuration_version_id,a.usage_status FROM processing_runs r JOIN processing_run_attempts a ON a.run_id=r.id WHERE r.document_version_id=$1`, v.ID).Scan(&config, &usage); err != nil {
		t.Fatal(err)
	}
	if calls != 0 || priced != 0 || config != local || usage != "not_applicable" || len(adapter.requests) != 0 || pages[1].PageNumber != 2 || pages[1].ReturnedModel != PDFTextVersion {
		t.Fatal("local provenance/accounting", calls, priced, config, usage, pages)
	}
}
