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
	"github.com/Balestrino/italian-weather-alert/internal/observation"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

type adminEvaluationFake struct{ recorded int }

func (f *adminEvaluationFake) RecordComparison(_ context.Context, _ string, comparison evaluation.ProcessingComparison) (evaluation.ComparisonReport, error) {
	f.recorded++
	return evaluation.ComparisonReport{Comparison: comparison, Passed: true}, nil
}

type adminObservationFake struct{ starts, reviews, assessments int }

func (f *adminObservationFake) Start(_ context.Context, request observation.StartRequest) (observation.Campaign, error) {
	f.starts++
	return observation.Campaign{ID: request.ID, StartedAt: request.StartedAt}, nil
}
func (f *adminObservationFake) RecordReview(_ context.Context, campaign, actor string, review observation.Review) (observation.Review, error) {
	f.reviews++
	review.Actor = actor
	return review, nil
}
func (f *adminObservationFake) Assess(_ context.Context, campaign, actor string, through time.Time, evidence registry.Evidence) (observation.Assessment, error) {
	f.assessments++
	return observation.Assessment{CampaignID: campaign, Actor: actor, Through: through, Evidence: evidence, Status: "running"}, nil
}
func (f *adminObservationFake) Report(_ context.Context, campaign string, through time.Time) (observation.Report, error) {
	return observation.Report{CampaignID: campaign, Through: through, Status: "running"}, nil
}
func (f *adminObservationFake) Get(_ context.Context, id string) (observation.Campaign, error) {
	return observation.Campaign{ID: id}, nil
}
func (f *adminObservationFake) List(context.Context) ([]observation.Campaign, error) {
	return []observation.Campaign{{ID: "trial"}}, nil
}
func (f *adminEvaluationFake) List(context.Context) ([]evaluation.ComparisonReport, error) {
	return []evaluation.ComparisonReport{{Passed: true}}, nil
}

func TestAdminMutationPayloadValidation(t *testing.T) {
	for _, test := range []struct {
		body, media string
		valid       bool
	}{
		{`{"revision":1,"actor":"operator"}`, "application/json", true},
		{`{"revision":1,"actor":"operator","report":{"success":true}}`, "application/json", false},
		{`{"revision":1,"actor":"operator"} {}`, "application/json", false},
		{`{"revision":"1","actor":"operator"}`, "application/json", false},
		{`{"revision":1,"actor":"operator"}`, "text/plain", false},
		{`payload=%7B%22revision%22%3A1%2C%22actor%22%3A%22operator%22%7D`, "application/x-www-form-urlencoded", true},
		{`payload={}&payload={}`, "application/x-www-form-urlencoded", false},
		{strings.Repeat(" ", 1<<20) + `{}`, "application/json", false},
	} {
		var input struct {
			Revision int    `json:"revision"`
			Actor    string `json:"actor"`
		}
		r := httptest.NewRequest("POST", "http://127.0.0.1/admin/sources/s/preview", strings.NewReader(test.body))
		r.Header.Set("Content-Type", test.media)
		if got := adminDecode(httptest.NewRecorder(), r, &input); got != test.valid {
			t.Fatalf("payload validation %s: got %v", test.media, got)
		}
	}
}

func TestAdminSourceRoutesRemainPrivateAndEscapeHistory(t *testing.T) {
	for _, path := range []string{"/admin/sources", "/admin/sources/s", "/admin/authorities", "/admin/channels", "/admin/processing-evaluations", "/admin/observation-campaigns", "/admin/observation-campaigns/trial", "/admin/observation-campaigns/trial/report", "/admin/observation-campaigns/trial/reviews", "/admin/observation-campaigns/trial/assess", "/admin/sources/s/configuration", "/admin/sources/s/intervals", "/admin/sources/s/preview", "/admin/sources/s/enable-collection", "/admin/sources/s/suspend-collection", "/admin/sources/s/regressions", "/admin/sources/s/accept", "/admin/sources/s/enable-public", "/admin/sources/s/suspend-public", "/admin/sources/s/suspend-interpretation", "/admin/sources/s/resume-interpretation", "/admin/sources/s/reprocess-interpretation"} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			rec := httptest.NewRecorder()
			Handler("public", nil).ServeHTTP(rec, httptest.NewRequest(method, "http://127.0.0.1"+path, nil))
			if rec.Code != 404 {
				t.Fatalf("admin route exposed: %s %s: %d", method, path, rec.Code)
			}
		}
	}
	rec := httptest.NewRecorder()
	adminRender(rec, "result", json.RawMessage(`{"actor":"</pre><script>alert(1)</script>"}`))
	if strings.Contains(rec.Body.String(), "<script>") {
		t.Fatal("source content executed in administration")
	}
}

func TestAdminObservationRoutesUseClosedPayloads(t *testing.T) {
	store := &adminObservationFake{}
	handler := HandlerWithAdministration(nil, AdminRuntime{Observations: store})
	at := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	post := func(path string, payload any, want int) string {
		t.Helper()
		body, _ := json.Marshal(payload)
		request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1"+path, bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s: %d %s", path, response.Code, response.Body.String())
		}
		return response.Body.String()
	}
	post("/admin/observation-campaigns", observation.StartRequest{ID: "trial", Actor: "operator", StartedAt: at, SourceIDs: []string{"source"}}, http.StatusOK)
	review := observation.Review{ID: "original", SourceID: "source", Kind: "original_comparison", Status: "pass", Evidence: registry.Evidence{URL: "https://evidence.example/review", Locator: "fixture", ObservedAt: at}}
	post("/admin/observation-campaigns/trial/reviews", map[string]any{"actor": "reviewer", "review": review}, http.StatusOK)
	post("/admin/observation-campaigns/trial/assess", map[string]any{"actor": "reviewer", "through": at, "evidence": review.Evidence}, http.StatusOK)
	if store.starts != 1 || store.reviews != 1 || store.assessments != 1 {
		t.Fatalf("observation mutations not dispatched: %#v", store)
	}
	post("/admin/observation-campaigns", map[string]any{"id": "trial", "actor": "operator", "started_at": at, "source_ids": []string{"source"}, "complete": true}, http.StatusBadRequest)
	if store.starts != 1 {
		t.Fatal("caller-supplied completion field was accepted")
	}
}

func TestAdminProcessingEvaluationRoutesUseClosedPayloads(t *testing.T) {
	store := &adminEvaluationFake{}
	handler := HandlerWithAdministration(nil, AdminRuntime{Evaluations: store})
	at := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	comparison := evaluation.ProcessingComparison{ID: "candidate", CorpusSHA256: strings.Repeat("a", 64), Reviewer: "reviewer", ReviewedAt: "2026-09-17", StartedAt: at, FinishedAt: at}
	body, _ := json.Marshal(map[string]any{"actor": "operator", "comparison": comparison})
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/admin/processing-evaluations", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || store.recorded != 1 || !strings.Contains(response.Body.String(), `"passed":true`) {
		t.Fatalf("processing evaluation import: %d %s", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1/admin/processing-evaluations", nil)
	request.Header.Set("Accept", "application/json")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"passed":true`) {
		t.Fatalf("processing evaluation list: %d %s", response.Code, response.Body.String())
	}
	body, _ = json.Marshal(map[string]any{"actor": "operator", "comparison": comparison, "passed": true})
	request = httptest.NewRequest(http.MethodPost, "http://127.0.0.1/admin/processing-evaluations", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || store.recorded != 1 {
		t.Fatal("caller-supplied pass verdict was accepted")
	}
}

func TestGuidedSourceControlsDecode(t *testing.T) {
	for _, tc := range []struct {
		action, body string
		valid        bool
	}{
		{"preview", "revision=1&actor=operator", true}, {"enable-collection", "revision=1&actor=operator", true}, {"suspend-collection", "revision=1&actor=operator", true},
		{"intervals", "expected_revision=0&check_seconds=600&delay_seconds=1800&actor=operator", true},
		{"preview", "revision=1&actor=", false}, {"preview", "revision=1&revision=2&actor=operator", false},
		{"intervals", "expected_revision=0&check_seconds=bad&delay_seconds=1&actor=operator", false},
		{"preview", "revision=1&actor=operator&enabled=true", false},
	} {
		var v map[string]any
		r := httptest.NewRequest("POST", "http://127.0.0.1/admin/sources/source/"+tc.action, strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if got := adminDecode(httptest.NewRecorder(), r, &v); got != tc.valid {
			t.Fatalf("%s %s: %v", tc.action, tc.body, got)
		}
	}
}

func TestGuidedReprocessingRequiresSelection(t *testing.T) {
	for _, body := range []string{"revision=1&actor=operator", "revision=1&actor=operator&from=not-a-date", "revision=1&actor=operator&from=&through=&error_code="} {
		r := httptest.NewRequest("POST", "http://127.0.0.1/admin/sources/source/reprocess-interpretation", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		var v any
		if adminDecode(httptest.NewRecorder(), r, &v) {
			t.Fatal("unscoped or invalid guided reprocessing accepted")
		}
	}
}
