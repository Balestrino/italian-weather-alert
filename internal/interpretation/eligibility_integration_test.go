//go:build integration

package interpretation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
	"testing"
	"time"
)

func TestEligibilitySchedulingAndCompletenessAgree(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := interpretationTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, classification.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 9, 17, 20, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://policy.example/rules", Locator: "fixture", ObservedAt: now}
	addSource := func(id string) {
		t.Helper()
		if err := reg.CreateAuthority(ctx, registry.Authority{ID: id + "-authority", Name: id, OfficialURL: "https://" + id + ".example"}); err != nil {
			t.Fatal(err)
		}
		if err := reg.CreateChannel(ctx, registry.Channel{ID: id + "-channel", PublisherID: id + "-authority", Platform: "fixture", URL: "https://" + id + ".example"}); err != nil {
			t.Fatal(err)
		}
		if err := reg.CreateSource(ctx, registry.Source{ID: id, AuthorityID: id + "-authority", ChannelID: id + "-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://" + id + ".example", Sections: []string{"https://" + id + ".example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"); err != nil {
			t.Fatal(err)
		}
	}
	addSource("source-a")
	addSource("source-b")

	version, err := reg.Version(ctx, "source-a", 1)
	if err != nil {
		t.Fatal(err)
	}
	cfg := version.Configuration
	iconURL := "https://source-a.example/brand.png"
	sum := sha256.Sum256([]byte("reviewed branding"))
	hash := hex.EncodeToString(sum[:])
	cfg.InferenceEligibility = &registry.InferenceEligibility{Version: registry.EligibilityVersion, Decorations: []registry.DecorationRule{{URL: iconURL, SHA256: hash, Reason: "reviewed_decoration", Evidence: evidence}}}
	revision, err := reg.AppendConfiguration(ctx, "source-a", 1, cfg, "test")
	if err != nil {
		t.Fatal(err)
	}
	if err = reg.EnableCollection(ctx, "source-a", revision, "test"); err == nil {
		t.Fatal("eligibility revision bypassed preview")
	}
	if err = reg.RecordPreview(ctx, "source-a", revision, "test", evidence); err != nil {
		t.Fatal(err)
	}
	if err = reg.EnableCollection(ctx, "source-a", revision, "test"); err != nil {
		t.Fatal(err)
	}
	retained := documents.New(pool, &interpretationObjects{})
	acquire := func(id, missing string) documents.Version {
		t.Helper()
		resources := []documents.Resource{
			{URL: "https://source-a.example/notice", Role: "original", Required: true, SourceID: "source-a", Configuration: revision, MediaType: "text/html", Bytes: []byte("Avviso comunale")},
			{URL: iconURL, Role: "resource", Required: true, SourceID: "source-a", Configuration: revision, MediaType: "image/png", Bytes: []byte("reviewed branding")},
		}
		if missing != "" {
			resources = append(resources, documents.Resource{URL: "https://source-a.example/required.pdf", Role: "attachment", Required: true, SourceID: "source-a", Configuration: revision, Missing: missing})
		}
		v, e := retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: "source-a", Configuration: revision, URL: "https://source-a.example/notice", Metadata: json.RawMessage("{}"), Resources: resources})
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	scheduler := New(pool, jobs.New(pool), inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, false)
	for _, missing := range []string{"", "unavailable"} {
		v := acquire("eligibility-"+missing, missing)
		urls, e := scheduler.ocrResources(ctx, v.ID)
		if e != nil || len(urls) != 0 {
			t.Fatalf("decoration scheduled: %v %v", urls, e)
		}
		scheduled, e := scheduler.Automatic(ctx, AcquisitionEvent{DocumentVersionID: v.ID, EvidenceHash: v.Hash, ContentChanged: true, At: now})
		if e != nil || !scheduled {
			t.Fatalf("classification stranded: %v", e)
		}
		content, e := classification.GatherContent(ctx, retained, ocr.NewStore(pool), v)
		if e != nil || content.Complete != (missing == "") || len(content.Sections) != 1 {
			t.Fatalf("eligibility/completeness differs: %#v %v", content, e)
		}
	}
	var ocrJobs, classJobs int
	if err = pool.QueryRow(ctx, "SELECT count(*) FILTER(WHERE kind=$1),count(*) FILTER(WHERE kind=$2) FROM processing_jobs", ocr.Kind, classification.Kind).Scan(&ocrJobs, &classJobs); err != nil {
		t.Fatal(err)
	}
	if ocrJobs != 0 || classJobs != 2 {
		t.Fatalf("unexpected scheduling: %d %d", ocrJobs, classJobs)
	}
	v := acquire("eligibility-canonical", "")
	process := processing.New(pool)
	catalog, e := classification.RegisterCatalog(ctx, process, "fixture", "model", now)
	if e != nil {
		t.Fatal(e)
	}
	manifestStore := classification.NewStore(pool)
	content, e := classification.GatherContent(ctx, retained, ocr.NewStore(pool), v)
	if e != nil {
		t.Fatal(e)
	}
	manifest, e := classification.BuildManifest(ctx, retained, ocr.NewStore(pool), v, content, catalog.ConfigurationVersionID)
	if e != nil {
		t.Fatal(e)
	}
	makeRun := func(key string) int64 {
		t.Helper()
		run, e := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: key, Stage: "classification", Workload: "evaluation", ConfigurationVersionID: catalog.ConfigurationVersionID, DocumentVersionID: &v.ID, Subject: json.RawMessage("{}"), CreatedAt: now})
		if e != nil {
			t.Fatal(e)
		}
		if e = manifestStore.PutManifest(ctx, run.ID, manifest); e != nil {
			t.Fatal(e)
		}
		return run.ID
	}
	original := makeRun("original-interpretation")
	copyRun := makeRun("copied-interpretation")
	if e = manifestStore.RecordReuse(ctx, copyRun, original, now); e != nil {
		t.Fatal(e)
	}
	for _, id := range []int64{original, copyRun, copyRun} {
		if _, e = scheduler.AfterExtraction(ctx, extraction.Result{RunID: id, Status: "extracted", Measures: []extraction.Measure{{Ordinal: 1}}}, "evaluation", now); e != nil {
			t.Fatal(e)
		}
	}
	var linkingJobs int
	if e = pool.QueryRow(ctx, "SELECT count(*) FROM processing_jobs WHERE kind='link_measure_update'").Scan(&linkingJobs); e != nil {
		t.Fatal(e)
	}
	if linkingJobs != 1 {
		t.Fatalf("equivalent evidence duplicated downstream effects: %d", linkingJobs)
	}

}
