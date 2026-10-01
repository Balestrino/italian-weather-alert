//go:build integration

package ocr

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
	"time"
)

type reuseAdapter struct {
	calls  int
	failAt int
}

func (a *reuseAdapter) Name() string { return "fixture" }
func (a *reuseAdapter) Complete(_ context.Context, _ inference.Request) (inference.Response, error) {
	a.calls++
	if a.calls == a.failAt {
		return inference.Response{}, &inference.CallError{Code: "temporary", Temporary: true}
	}
	return inference.Response{ID: fmt.Sprint(a.calls), Model: "deepseek-ocr-2", Content: "literal page text", Usage: inference.Usage{InputTokens: pointer(10), OutputTokens: pointer(5)}}, nil
}

type reuseRenderer struct {
	calls   int
	changed bool
}

func (r *reuseRenderer) Render(context.Context, []byte) ([]PageImage, error) {
	r.calls++
	second := "page-two"
	if r.changed {
		second = "changed-page-two"
	}
	return []PageImage{{Number: 1, MediaType: "image/png", Bytes: []byte("page-one")}, {Number: 2, MediaType: "image/png", Bytes: []byte(second)}}, nil
}
func TestRunnerReusesPagesAndRenderManifest(t *testing.T) {
	ctx := context.Background()
	pool := ocrTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, Migrate, Migrate} {
		if err := m(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://comune.example/policy", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "ocr-authority", Name: "OCR municipality", OfficialURL: "https://comune.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "ocr-channel", PublisherID: "ocr-authority", Platform: "fixture", URL: "https://comune.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "ocr-source", AuthorityID: "ocr-authority", ChannelID: "ocr-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://comune.example", Sections: []string{"https://comune.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	retained := documents.New(pool, &integrationObjects{})
	originalURL, missingURL := "https://comune.example/scan.png", "https://comune.example/missing.pdf"
	version, err := retained.Retain(ctx, documents.Acquisition{ID: "ocr-fixture", SourceID: "ocr-source", Configuration: 1, URL: originalURL, Metadata: json.RawMessage(`{"fixture":"CAL-OCR-SCAN"}`), Resources: []documents.Resource{
		{URL: originalURL, Role: "original", Required: true, SourceID: "ocr-source", Configuration: 1, MediaType: "image/png", Bytes: []byte("synthetic scan")},
		{URL: missingURL, Role: "attachment", Required: true, SourceID: "ocr-source", Configuration: 1, Missing: "unavailable"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	process := processing.New(pool)
	catalog, err := RegisterCatalog(ctx, process, "openai-chat", "deepseek-ocr-2", now)
	if err != nil {
		t.Fatal(err)
	}

	_ = version
	store := NewStore(pool)
	adapter := &reuseAdapter{failAt: 2}
	renderer := &reuseRenderer{}
	runner := Runner{Documents: retained, Processing: process, Results: store, Artifacts: store, ProviderScope: "fixture", RendererIdentity: "fixture-v1", Adapter: &inference.RecordedAdapter{Adapter: adapter, Ledger: process}, Renderer: renderer, Model: "deepseek-ocr-2", ConfigurationVersion: catalog.ConfigurationVersionID, Now: func() time.Time { return now }}
	queue := jobs.New(pool)
	retain := func(id, body, other string) documents.Version {
		t.Helper()
		v, e := retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: "ocr-source", Configuration: 1, URL: originalURL, Metadata: json.RawMessage("{}"), Resources: []documents.Resource{
			{URL: originalURL, Role: "original", Required: true, SourceID: "ocr-source", Configuration: 1, MediaType: "application/pdf", Bytes: []byte(body)},
			{URL: missingURL, Role: "attachment", Required: true, SourceID: "ocr-source", Configuration: 1, MediaType: "text/plain", Bytes: []byte(other)},
		}})
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	run := func(v documents.Version, wantFailure bool) {
		t.Helper()
		_, e := Enqueue(ctx, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, Payload{DocumentVersionID: v.ID, ResourceURL: originalURL, Workload: "evaluation"}, now)
		if e != nil {
			t.Fatal(e)
		}
		claim, e := queue.Claim(ctx, "inference", "reuse-test", now, time.Minute)
		if e != nil {
			t.Fatal(e)
		}
		output, e := runner.Handler()(ctx, claim.Job)
		if wantFailure {
			if e == nil {
				t.Fatal("expected later-page failure")
			}
			if e = queue.Fail(ctx, claim, jobs.Failure{Code: "temporary", Detail: "fixture", Temporary: true}, now); e != nil {
				t.Fatal(e)
			}
		} else {
			if e != nil {
				t.Fatal(e)
			}
			if e = queue.Complete(ctx, claim, output, now); e != nil {
				t.Fatal(e)
			}
		}
		now = now.Add(2 * time.Second)
	}
	first := retain("reuse-first", "pdf-one", "a")
	run(first, true)
	if adapter.calls != 2 {
		t.Fatal("cold calls missing")
	}
	run(first, false)
	if adapter.calls != 3 {
		t.Fatalf("completed first page was repeated: calls=%d", adapter.calls)
	}
	second := retain("reuse-second", "pdf-two-metadata-only", "a")
	run(second, false)
	if adapter.calls != 3 {
		t.Fatal("identical rendered pages called provider")
	}
	renderCalls := renderer.calls
	third := retain("reuse-third", "pdf-two-metadata-only", "different attachment")
	run(third, false)
	if adapter.calls != 3 || renderer.calls != renderCalls {
		t.Fatal("warm resource was rendered or called again")
	}
	renderer.changed = true
	fourth := retain("reuse-fourth", "pdf-changed", "a")
	run(fourth, false)
	if adapter.calls != 4 {
		t.Fatalf("changed page did not call exactly once: %d", adapter.calls)
	}
	var calls, tokens int64
	if err = pool.QueryRow(ctx, "SELECT count(*),COALESCE(sum(input_tokens+output_tokens),0) FROM processing_provider_calls").Scan(&calls, &tokens); err != nil {
		t.Fatal(err)
	}
	if calls != 4 || tokens != 45 {
		t.Fatalf("ledger changed paid consumption: %d %d", calls, tokens)
	}
	var copies int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM ocr_page_results WHERE document_version_id IN ($1,$2) AND (input_tokens IS NOT NULL OR output_tokens IS NOT NULL)", second.ID, third.ID).Scan(&copies); err != nil {
		t.Fatal(err)
	}
	if copies != 0 {
		t.Fatal("reuse copied paid tokens")
	}
	runner.ReuseSources = map[string]bool{"another-source": true}
	fifth := retain("outside-canary", "new-metadata-same-pages", "a")
	run(fifth, false)
	if adapter.calls != 6 {
		t.Fatal("unselected OCR source reused artifacts")
	}
}
