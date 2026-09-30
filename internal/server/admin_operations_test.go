package server

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/operations"
)

type operationFixture struct {
	calls int
	fail  bool
}

func (s *operationFixture) Page(_ context.Context, section string, f operations.Filter, at time.Time) (operations.Report, error) {
	s.calls++
	if s.fail || f.SourceID == "unavailable" {
		return operations.Report{}, errors.New("database password must not appear")
	}
	if f.SourceID == "empty" {
		return operations.Report{Section: section, Filter: f, ObservedAt: at, Table: operations.Table{Columns: []string{"job_id"}}}, nil
	}
	return operations.Report{Section: section, Filter: f, ObservedAt: at, More: true, Table: operations.Table{Columns: []string{"job_id", "state", "attempt_count", "last_error_code", "source_id"}, Rows: []map[string]any{{"job_id": int64(9), "state": "failed", "attempt_count": 3, "last_error_code": "<script>alert(1)</script>", "source_id": "source"}}}}, nil
}

type relaunchFixture struct{ calls int }

func (s *relaunchFixture) Relaunch(context.Context, int64, int, string, time.Time) error {
	s.calls++
	return nil
}

func TestOperationsIsolationFormsAndFailures(t *testing.T) {
	fixture := &operationFixture{}
	relaunch := &relaunchFixture{}
	handler := HandlerWithAdministration(nil, AdminRuntime{Operations: fixture, Jobs: relaunch})
	for _, path := range []string{"/admin/operations/sources", "/admin/operations/jobs", "/admin/operations/documents", "/admin/operations/findings", "/admin/operations/dpc-comparisons", "/admin/operations/usage", "/admin/jobs/9/relaunch"} {
		for _, method := range []string{"GET", "POST"} {
			rec := httptest.NewRecorder()
			Handler("public", nil).ServeHTTP(rec, httptest.NewRequest(method, "http://127.0.0.1"+path, nil))
			if rec.Code != 404 {
				t.Fatalf("public %s %s = %d", method, path, rec.Code)
			}
		}
	}
	for _, origin := range []string{"http://evil.example", "http://127.0.0.1:8080"} {
		r := httptest.NewRequest("POST", "http://127.0.0.1/admin/jobs/9/relaunch", strings.NewReader(`{"actor":"test","after_attempt":3}`))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		if rec.Code != 403 || relaunch.calls != 0 {
			t.Fatal("cross-origin relaunch reached queue")
		}
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1/admin/jobs/9/relaunch", strings.NewReader(`payload=%7B%22actor%22%3A%22test%22%2C%22after_attempt%22%3A3%7D`))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)
	if rec.Code != 200 || relaunch.calls != 1 {
		t.Fatalf("form relaunch: %d %s", rec.Code, rec.Body)
	}
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "http://127.0.0.1/admin/operations/jobs?source_id=a%26b&page=2&state=failed", nil))
	html := rec.Body.String()
	if rec.Code != 200 || strings.Contains(html, "<script>") || !strings.Contains(html, "&lt;script&gt;") || !strings.Contains(html, `action="/admin/jobs/9/relaunch"`) || !strings.Contains(html, "page=3&amp;source_id=a%26b&amp;state=failed") {
		t.Fatalf("unsafe or incomplete HTML: %s", html)
	}
	fixture.fail = true
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "http://127.0.0.1/admin/operations/usage", nil))
	if rec.Code != 503 || strings.Contains(rec.Body.String(), "password") {
		t.Fatal("storage failure exposed private details")
	}
}

func TestGuidedRelaunchValidation(t *testing.T) {
	fixture := &relaunchFixture{}
	h := HandlerWithAdministration(nil, AdminRuntime{Jobs: fixture})
	for _, tc := range []struct {
		body string
		code int
	}{
		{"actor=operator&after_attempt=3", 200},
		{"actor=&after_attempt=3", 400},
		{"actor=a&actor=b&after_attempt=3", 400},
		{"actor=a&after_attempt=0", 400},
		{"actor=a&after_attempt=3&extra=1", 400},
	} {
		before := fixture.calls
		r := httptest.NewRequest("POST", "http://127.0.0.1/admin/jobs/9/relaunch", strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("Accept", "text/html")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.code || !strings.Contains(w.Body.String(), "Torna alla vista aggiornata") {
			t.Fatalf("guided form: %d %s", w.Code, w.Body)
		}
		if tc.code != 200 && fixture.calls != before {
			t.Fatal("invalid input mutated job")
		}
	}
}

func TestUsagePresentationPreservesUnknownAndCurrency(t *testing.T) {
	report := operations.Report{Section: "usage", Filter: operations.Filter{Page: 1}, Table: operations.Table{Columns: []string{"run_id", "cost_status", "estimated_cost_microunits", "pricing_provenance_url"}, Rows: []map[string]any{{"run_id": 1, "cost_status": "unknown", "estimated_cost_microunits": nil, "pricing_provenance_url": "https://provider.example/prices"}}}, Totals: &operations.Table{Columns: []string{"currency", "known_cost_microunits", "cost_complete"}, Rows: []map[string]any{{"currency": "EUR", "known_cost_microunits": 12, "cost_complete": false}, {"currency": "USD", "known_cost_microunits": 3, "cost_complete": true}}}}
	w := httptest.NewRecorder()
	renderUI(w, operationTemplates, "page", report)
	body := w.Body.String()
	for _, want := range []string{"EUR", "USD", "non disponibile", "unknown", "https://provider.example/prices", "Sì", "No", "subtotali", "bootstrap", "embedding", "milionesimi"} {
		if !strings.Contains(body, want) {
			t.Fatalf("usage lost %s: %s", want, body)
		}
	}
}
