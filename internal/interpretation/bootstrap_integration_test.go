//go:build integration

package interpretation

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

func TestAutomaticKeepsBootstrapIdentityAcrossOrdinaryChecks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	pool := interpretationTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, classification.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	reg := registry.New(pool)
	if err := reg.CreateAuthority(ctx, registry.Authority{ID: "bootstrap-authority", Name: "Fixture", OfficialURL: "https://fixture.example"}); err != nil {
		t.Fatal(err)
	}
	if err := reg.CreateChannel(ctx, registry.Channel{ID: "bootstrap-channel", PublisherID: "bootstrap-authority", Platform: "fixture", URL: "https://fixture.example"}); err != nil {
		t.Fatal(err)
	}
	evidence := registry.Evidence{URL: "https://fixture.example/policy", Locator: "fixture", ObservedAt: now}
	cfg := registry.Configuration{URL: "https://fixture.example", Sections: []string{"https://fixture.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}
	if err := reg.CreateSource(ctx, registry.Source{ID: "bootstrap-source", AuthorityID: "bootstrap-authority", ChannelID: "bootstrap-channel", ProductID: "municipal", Territory: "050004"}, cfg, "test"); err != nil {
		t.Fatal(err)
	}
	docs := documents.New(pool, &interpretationObjects{})
	retain := func(id, text string) documents.Version {
		t.Helper()
		v, err := docs.Retain(ctx, documents.Acquisition{ID: id, SourceID: "bootstrap-source", Configuration: 1, URL: "https://fixture.example/notice", Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{
			{URL: "https://fixture.example/notice", Role: "original", Required: true, SourceID: "bootstrap-source", Configuration: 1, MediaType: "text/html", Bytes: []byte(text)},
			{URL: "https://fixture.example/notice.pdf", Role: "attachment", Required: true, SourceID: "bootstrap-source", Configuration: 1, MediaType: "application/pdf", Bytes: []byte("fixture PDF")},
		}})
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	v := retain("bootstrap-v1", "first notice")
	queue := jobs.New(pool)
	s := New(pool, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, false)
	event := AcquisitionEvent{DocumentVersionID: v.ID, EvidenceHash: v.Hash, Workload: "bootstrap", ContentChanged: true, At: now}
	if _, err := s.Automatic(ctx, event); err != nil {
		t.Fatal(err)
	}
	event.Workload = "ordinary"
	if _, err := s.Automatic(ctx, event); err != nil {
		t.Fatalf("ordinary revisit must retain bootstrap job identity: %v", err)
	}
	claim, err := queue.Claim(ctx, inference.Queue, "fixture-worker", now.Add(time.Second), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err = queue.Fail(ctx, claim, jobs.Failure{Code: "provider_rejected", Detail: "fixture rejection", Temporary: false}, now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err = s.Automatic(ctx, event); err != nil {
			t.Fatalf("failed bootstrap revisit: %v", err)
		}
	}
	var state, workload string
	var count, attempts int
	if err = pool.QueryRow(ctx, "SELECT state,payload->>'workload',attempt_count FROM processing_jobs WHERE id=$1", claim.ID).Scan(&state, &workload, &attempts); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM processing_jobs").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if state != "failed" || workload != "bootstrap" || attempts != 1 || count != 1 {
		t.Fatalf("historical failed job was changed/replayed: %s %s %d %d", state, workload, attempts, count)
	}
	// A separately completed OCR dependency must retain the first trigger's workload,
	// even when an ordinary check is the event that observes completion.
	process := processing.New(pool)
	catalog, err := ocr.RegisterCatalog(ctx, process, "openai-chat", "fixture-ocr", now)
	if err != nil {
		t.Fatal(err)
	}
	source := "bootstrap-source"
	run, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "fixture-ocr", Workload: "bootstrap", Stage: "ocr", ConfigurationVersionID: catalog.ConfigurationVersionID, SourceID: &source, DocumentVersionID: &v.ID, Subject: json.RawMessage(`{}`), CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if err = ocr.NewStore(pool).PutResource(ctx, ocr.ResourceResult{RunID: run.ID, DocumentVersionID: v.ID, ResourceURL: "https://fixture.example/notice.pdf", Status: "complete", PageCount: 0, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.AfterOCR(ctx, v.ID, "ordinary", now); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "SELECT payload->>'workload' FROM processing_jobs WHERE kind=$1", classification.Kind).Scan(&workload); err != nil {
		t.Fatal(err)
	}
	if workload != "bootstrap" {
		t.Fatalf("classification lost original workload: %s", workload)
	}
	changed := retain("ordinary-v2", "meaningfully changed notice")
	if _, err = s.Automatic(ctx, AcquisitionEvent{DocumentVersionID: changed.ID, EvidenceHash: changed.Hash, Workload: "ordinary", ContentChanged: true, At: now}); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "SELECT payload->>'workload' FROM processing_jobs WHERE kind=$1 AND payload->>'document_version_id'=$2", ocr.Kind, strconv.FormatInt(changed.ID, 10)).Scan(&workload); err != nil {
		t.Fatal(err)
	}
	if workload != "ordinary" {
		t.Fatalf("new version inherited bootstrap workload: %s", workload)
	}
}
