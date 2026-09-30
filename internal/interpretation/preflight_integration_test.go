//go:build integration

package interpretation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

func TestPreflightContiguousGroups(t *testing.T) {
	ctx := context.Background()
	pool := interpretationTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate, Migrate} {
		if e := m(ctx, pool); e != nil {
			t.Fatal(e)
		}
	}
	reg := registry.New(pool)
	source := "calcinaia-municipal"
	if e := reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "a", OfficialURL: "https://example.org"}); e != nil {
		t.Fatal(e)
	}
	if e := reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "fixture", URL: "https://example.org"}); e != nil {
		t.Fatal(e)
	}
	ev := registry.Evidence{URL: "https://example.org/policy", Locator: "fixture", ObservedAt: time.Now()}
	if e := reg.CreateSource(ctx, registry.Source{ID: source, AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://example.org", Sections: []string{"https://example.org/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &ev, CollectionPermitted: true, RetentionPermitted: true}}, "test"); e != nil {
		t.Fatal(e)
	}
	retained := documents.New(pool, &interpretationObjects{})
	p := &Preflight{Documents: retained, Renderer: preflightRenderer{}, Configuration: "cfg", RendererIdentity: "renderer", Sources: map[string]bool{source: true}}
	s := New(pool, jobs.New(pool), inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, false)
	retain := func(id, text string) documents.Version {
		t.Helper()
		b := []byte(`<div class="js-view-dom-id-` + strings.Repeat(id, 64) + `">` + text + `</div>`)
		v, e := retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: source, Configuration: 1, URL: "https://example.org/notice", Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: "https://example.org/notice", SourceID: source, Configuration: 1, Role: "original", Required: true, MediaType: "text/html", Bytes: b}}})
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	a := retain("a", "Warning A")
	b := retain("b", "Warning A")
	c := retain("c", "Warning B")
	d := retain("d", "Warning A")
	if a.ID == b.ID || b.ID == c.ID || c.ID == d.ID {
		t.Fatal("fixture must retain distinct versions for preflight grouping")
	}
	for _, tc := range []struct {
		v   documents.Version
		rep int64
	}{{a, a.ID}, {b, a.ID}, {c, c.ID}, {d, d.ID}} {
		dec, e := s.Prepare(ctx, p, tc.v.ID)
		if e != nil || dec.RepresentativeID != tc.rep {
			t.Fatalf("decision %+v %v", dec, e)
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dec, e := s.Prepare(ctx, p, b.ID)
			if e != nil || dec.RepresentativeID != a.ID {
				t.Errorf("concurrent %+v %v", dec, e)
			}
		}()
	}
	wg.Wait()
	if _, e := pool.Exec(ctx, `INSERT INTO interpretation_archives VALUES($1,now(),now(),'test')`, d.ID); e != nil {
		t.Fatal(e)
	}
	f := retain("e", "Warning A")
	dec, e := s.Prepare(ctx, p, f.ID)
	if e != nil || dec.RepresentativeID != f.ID {
		t.Fatalf("archive boundary %+v %v", dec, e)
	}
	// Equivalent pending copies spend no attempt budget; meaningful C remains eligible.
	s.Preflight = p
	now := time.Now()
	for _, v := range []documents.Version{a, b, c} {
		if _, err := s.Automatic(ctx, AcquisitionEvent{DocumentVersionID: v.ID, EvidenceHash: v.Hash, ContentChanged: true, At: now}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.DeferEquivalent(ctx, now); err != nil {
		t.Fatal(err)
	}
	var attempts int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM processing_jobs WHERE payload->>'document_version_id'=$1", fmt.Sprint(b.ID)).Scan(&attempts); err != nil || attempts != 0 {
		t.Fatalf("spent waiting budget %d %v", attempts, err)
	}
	claim, err := s.Queue.Claim(ctx, inference.Queue, "test", now.Add(time.Second), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	var payload classification.Payload
	_ = json.Unmarshal(claim.Payload, &payload)
	if payload.DocumentVersionID != a.ID {
		t.Fatal("representative not first")
	}
	catalog, err := classification.RegisterCatalog(ctx, processing.New(pool), "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	adapter := &preflightAdapter{}
	runner := &classification.Runner{Documents: retained, OCR: ocr.NewStore(pool), Processing: processing.New(pool), Results: classification.NewStore(pool), Manifests: classification.NewStore(pool), ReuseSources: map[string]bool{source: true}, Adapter: adapter, Model: "qwen3.8-27b", ConfigurationVersion: catalog.ConfigurationVersionID}
	if _, err = s.Guard(runner.Handler())(ctx, claim.Job); err != nil {
		t.Fatal(err)
	}
	if err = s.Queue.Complete(ctx, claim, jobs.Result{Payload: json.RawMessage(`{"status":"fixture"}`)}, now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	eq, waiting, err := s.Equivalent(ctx, b.ID)
	if err != nil || !eq || waiting {
		t.Fatalf("representative did not release copy: %v %v %v", eq, waiting, err)
	}
	// A fresh scheduler simulates a restart. Ready copies are released exactly
	// once and their validated result is still materialized without a provider.
	resumed := New(pool, s.Queue, s.Policy, false)
	resumed.Preflight = p
	for i := 0; i < 2; i++ {
		if err = resumed.ReleaseEquivalent(ctx, now.Add(time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	var copyJobs int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM processing_jobs WHERE payload->>'document_version_id'=$1", fmt.Sprint(b.ID)).Scan(&copyJobs); err != nil || copyJobs != 1 {
		t.Fatalf("copy release %d %v", copyJobs, err)
	}
	if _, err = pool.Exec(ctx, "UPDATE processing_jobs SET available_at=$2 WHERE payload->>'document_version_id'=$1", fmt.Sprint(c.ID), now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	copyClaim, err := s.Queue.Claim(ctx, inference.Queue, "test-copy", now.Add(2*time.Minute), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal(copyClaim.Payload, &payload)
	if payload.DocumentVersionID != b.ID {
		t.Fatalf("wrong copy %d", payload.DocumentVersionID)
	}
	// A nil downstream adapter would panic on any provider fallback. Reuse succeeds.
	runner.Adapter = &inference.GatedAdapter{}
	if _, err = s.Guard(runner.Handler())(ctx, copyClaim.Job); err != nil {
		t.Fatal(err)
	}
	if adapter.calls != 1 {
		t.Fatalf("duplicate provider call: %d", adapter.calls)
	}
	var reused int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM interpretation_reuse").Scan(&reused); err != nil || reused != 1 {
		t.Fatalf("reuse missing %d %v", reused, err)
	}
	rotated := *p
	rotated.Configuration = "cfg-v2"
	if dec, err := s.Prepare(ctx, &rotated, a.ID); err != nil || dec.RepresentativeID != a.ID {
		t.Fatalf("catalog rotation must preserve prior decision: %+v %v", dec, err)
	}
	var storedConfiguration string
	if err := pool.QueryRow(ctx, "SELECT configuration FROM interpretation_preflight WHERE document_version_id=$1", a.ID).Scan(&storedConfiguration); err != nil || storedConfiguration != "cfg" {
		t.Fatalf("prior preflight was rewritten: %q %v", storedConfiguration, err)
	}
	s.Preflight = &rotated
	if _, err := s.Automatic(ctx, AcquisitionEvent{DocumentVersionID: a.ID, EvidenceHash: a.Hash, ContentChanged: true, At: now.Add(4 * time.Minute)}); err != nil {
		t.Fatalf("scheduled unchanged version after catalog rotation: %v", err)
	}
	g := retain("g", "Warning A")
	if dec, err := s.Prepare(ctx, &rotated, g.ID); err != nil || dec.RepresentativeID != g.ID {
		t.Fatalf("new version must use new catalog boundary: %+v %v", dec, err)
	}
	// Remove queue/result dependents before testing the independent retention FK.
	// Use C's group instead: its retained version has no processing run yet.
	if _, err = pool.Exec(ctx, "DELETE FROM interpretation_triggers WHERE document_version_id=$1", c.ID); err != nil {
		t.Fatal(err)
	}
	// Retention removes dependent grouping, never leaves an unresolvable representative.
	if _, e = pool.Exec(ctx, "DELETE FROM retained_resources WHERE version_id=$1", f.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = pool.Exec(ctx, "DELETE FROM retained_acquisitions WHERE version_id=$1", f.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = pool.Exec(ctx, "DELETE FROM retained_versions WHERE id=$1", f.ID); e != nil {
		t.Fatal(e)
	}
	var count int
	if e = pool.QueryRow(ctx, "SELECT count(*) FROM interpretation_preflight WHERE document_version_id=$1", f.ID).Scan(&count); e != nil || count != 0 {
		t.Fatalf("dangling group %d %v", count, e)
	}
}

type preflightAdapter struct{ calls int }

func (a *preflightAdapter) Name() string { return "fixture" }
func (a *preflightAdapter) Complete(context.Context, inference.Request) (inference.Response, error) {
	a.calls++
	return inference.Response{Model: "qwen3.8-27b", Content: `{"relevant":false,"reason_code":"not_relevant","evidence_quote":"Warning A"}`}, nil
}

func TestEquivalentPDFWaitsBeforeOCRQueue(t *testing.T) {
	ctx := context.Background()
	pool := interpretationTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate, Migrate} {
		if err := m(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	reg := registry.New(pool)
	source := "cfr-criticality"
	if err := reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "CFR fixture", OfficialURL: "https://example.org"}); err != nil {
		t.Fatal(err)
	}
	if err := reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "fixture", URL: "https://example.org"}); err != nil {
		t.Fatal(err)
	}
	ev := registry.Evidence{URL: "https://example.org/policy", Locator: "fixture", ObservedAt: time.Now()}
	cfg := registry.Configuration{URL: "https://example.org", Sections: []string{"https://example.org"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &ev, CollectionPermitted: true, RetentionPermitted: true}}
	if err := reg.CreateSource(ctx, registry.Source{ID: source, AuthorityID: "a", ChannelID: "c", ProductID: "criticality", Territory: "Toscana"}, cfg, "test"); err != nil {
		t.Fatal(err)
	}
	retained := documents.New(pool, &interpretationObjects{})
	s := New(pool, jobs.New(pool), inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, false)
	s.Preflight = &Preflight{Documents: retained, Renderer: preflightRenderer{}, Configuration: "cfg", RendererIdentity: "renderer", Sources: map[string]bool{source: true}}
	var versions []documents.Version
	for i, pdf := range []string{"map-orange;metadata=one", "map-orange;metadata=two", "map-red;metadata=three"} {
		v, err := retained.Retain(ctx, documents.Acquisition{ID: fmt.Sprint(i), SourceID: source, Configuration: 1, URL: "https://example.org", Metadata: json.RawMessage(`{"issued":"2026-09-24"}`), Resources: []documents.Resource{
			{URL: "https://example.org", SourceID: source, Configuration: 1, Role: "original", Required: true, MediaType: "text/html", Bytes: []byte("Warning")},
			{URL: "https://example.org/bulletin.pdf", SourceID: source, Configuration: 1, Role: "attachment", Required: true, MediaType: "application/pdf", Bytes: []byte(pdf)},
		}})
		if err != nil {
			t.Fatal(err)
		}
		versions = append(versions, v)
		scheduled, err := s.Automatic(ctx, AcquisitionEvent{DocumentVersionID: v.ID, EvidenceHash: v.Hash, ContentChanged: true, At: time.Now()})
		if err != nil || scheduled != (i != 1) {
			t.Fatalf("PDF %d scheduled=%v err=%v", i, scheduled, err)
		}
	}
	// No provider adapter exists in this fixture: preparation is local rendering.
	if err := s.ReleaseEquivalent(ctx, time.Now()); err != nil {
		t.Fatal(err)
	}
	for i, v := range versions {
		var count int
		err := pool.QueryRow(ctx, `SELECT count(*) FROM processing_jobs WHERE payload->>'document_version_id'=$1`, fmt.Sprint(v.ID)).Scan(&count)
		want := 1
		if i == 1 {
			want = 0
		}
		if err != nil || count != want {
			t.Fatalf("PDF %d queued=%d want=%d err=%v", i, count, want, err)
		}
	}
	var triggers int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM interpretation_triggers`).Scan(&triggers); err != nil || triggers != 3 {
		t.Fatal("durable trigger lost", triggers, err)
	}
}
