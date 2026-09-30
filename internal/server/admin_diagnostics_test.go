package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/diagnostics"
)

type diagnosticFixture struct {
	snapshots, comparisons int
}

func (f *diagnosticFixture) Snapshot(_ context.Context, sourceID string, at time.Time) (diagnostics.Snapshot, error) {
	f.snapshots++
	if sourceID == "missing" {
		return diagnostics.Snapshot{}, diagnostics.ErrNotFound
	}
	return diagnostics.Snapshot{SourceID: sourceID, ObservedAt: at, Availability: diagnostics.Signal{State: "available"}, Timeliness: diagnostics.Signal{State: "current"}, Interpretation: diagnostics.Signal{State: "supported"}}, nil
}

func (f *diagnosticFixture) Compare(_ context.Context, value diagnostics.Comparison) (diagnostics.Comparison, error) {
	f.comparisons++
	if value.Scope == "invalid" {
		return diagnostics.Comparison{}, diagnostics.ErrInvalid
	}
	value.ID = "comparison-id"
	return value, nil
}

func TestDiagnosticRoutesAreLocalAndMapStableFailures(t *testing.T) {
	fixture := &diagnosticFixture{}
	admin := HandlerWithAdministration(nil, AdminRuntime{Diagnostics: fixture})
	public := Handler("public", nil)
	for _, request := range []*http.Request{
		httptest.NewRequest("GET", "http://127.0.0.1/admin/diagnostics/sources/regional", nil),
		httptest.NewRequest("POST", "http://127.0.0.1/admin/diagnostics/dpc-comparisons", strings.NewReader(`{}`)),
	} {
		recorder := httptest.NewRecorder()
		public.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("public diagnostic route returned %d", recorder.Code)
		}
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "http://127.0.0.1/admin/diagnostics/sources/regional", nil)
	request.Header.Set("Accept", "application/json")
	admin.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"source_id":"regional"`) || fixture.snapshots != 1 {
		t.Fatalf("snapshot response: %d %s", recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	request = httptest.NewRequest("POST", "http://127.0.0.1/admin/diagnostics/dpc-comparisons", strings.NewReader(`{"regional_version_id":1,"dpc_version_id":2,"scope":"fixture","actor":"reviewer","dimensions":[],"compared_at":"2026-09-17T12:00:00Z"}`))
	request.Header.Set("Content-Type", "application/json")
	admin.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"id":"comparison-id"`) || fixture.comparisons != 1 {
		t.Fatalf("comparison response: %d %s", recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	request = httptest.NewRequest("GET", "http://127.0.0.1/admin/diagnostics/sources/missing", nil)
	request.Header.Set("Accept", "application/json")
	admin.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound || !strings.Contains(recorder.Body.String(), "diagnostic_evidence_not_found") {
		t.Fatalf("not-found response: %d %s", recorder.Code, recorder.Body.String())
	}
}
