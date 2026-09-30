//go:build integration

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/evaluation"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

func TestAdminRegressionEvidenceAndFailureGate(t *testing.T) {
	ctx := context.Background()
	pool := adminTestDB(t)
	reg := registry.New(pool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(registry.Migrate(ctx, pool))
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "Comune", OfficialURL: "https://comune.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "official", URL: "https://comune.example"}))
	e := registry.Evidence{URL: "https://comune.example/report", Locator: "test fixture", ObservedAt: time.Now().UTC()}
	c := registry.Configuration{URL: "https://comune.example/notizie", Sections: []string{"https://comune.example/notizie"}, AccessMethod: "html", Attribution: "Comune", Provenance: &e, Policy: registry.Policy{Evidence: &e, CollectionPermitted: true, RetentionPermitted: true, PublicationPermitted: true}}
	must(reg.CreateSource(ctx, registry.Source{ID: "s", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "050004"}, c, "test"))
	must(reg.RecordPreview(ctx, "s", 1, "test", e))
	must(reg.EnableCollection(ctx, "s", 1, "test"))
	h := HandlerWithAdministration(nil, AdminRuntime{Registry: reg})
	post := func(action string, payload any, want int) {
		t.Helper()
		b, err := json.Marshal(payload)
		must(err)
		r := httptest.NewRequest("POST", "http://127.0.0.1/admin/sources/s/"+action, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s: %d %s", action, w.Code, w.Body.String())
		}
	}
	now := time.Now().UTC()
	run := evaluation.Run{ID: "failed", SourceID: "s", Revision: 1, Suite: "municipal", CorpusSHA256: strings.Repeat("a", 64), Reviewer: "fixture reviewer", ReviewedAt: "2026-09-16", ProcessVersion: "test", Mode: "service", StartedAt: now, FinishedAt: now}
	for _, stage := range []string{"discovery", "ocr", "classification", "extraction", "linking"} {
		run.Checks = append(run.Checks, evaluation.Check{CaseID: "fixture", Kind: "simulation", Stage: stage, MeasureField: stage == "extraction", Field: "measure", Expected: "supported", Actual: "absent", Outcome: "omission", Evidence: "fixture/report"})
	}
	post("regressions", map[string]any{"revision": 1, "actor": "test", "previous_id": "", "run": run}, 200)
	post("regressions", map[string]any{"revision": 1, "actor": "test", "previous_id": "", "run": run, "passed": true}, 400)
	acceptance := registry.Acceptance{Report: e, PeriodStart: now.Add(-8 * 24 * time.Hour), PeriodEnd: now.Add(-time.Hour), Sections: c.Sections, ExtractionVerified: true, UpdatesVerified: true, AttachmentsVerified: true, ScannedAttachmentsVerified: true, HistoryVerified: true, FailureBehaviorVerified: true, InterfacesEquivalent: true, CoverageStatus: "accepted_declared_scope", CoverageLimitations: []string{}}
	post("accept", map[string]any{"revision": 1, "actor": "test", "acceptance": acceptance}, 409)
	post("enable-public", map[string]any{"revision": 1, "actor": "test"}, 409)
	r := httptest.NewRequest("GET", "http://127.0.0.1/admin/sources/s/regressions", nil)
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"omissions"`) || !strings.Contains(w.Body.String(), `"unsupported_assertions"`) || !strings.Contains(w.Body.String(), `"indeterminate_fields"`) {
		t.Fatalf("reports missing: %s", w.Body.String())
	}
}
