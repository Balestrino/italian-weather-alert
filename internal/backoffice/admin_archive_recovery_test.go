package backoffice

import (
	"bytes"
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"net/http"
	"net/http/httptest"
	"testing"
)

type archiveRecoveryStub struct {
	AdminInterpretation
	calls int
}

func (s *archiveRecoveryStub) RecoverArchived(_ context.Context, r interpretation.ArchiveRecovery) (interpretation.ArchiveRecovery, error) {
	s.calls++
	return r, nil
}
func (s *archiveRecoveryStub) ArchiveRecoveries(context.Context, string) ([]interpretation.ArchiveRecovery, error) {
	return []interpretation.ArchiveRecovery{}, nil
}

func TestArchiveRecoveryPrivateBoundaryAndClosedSchema(t *testing.T) {
	store := &archiveRecoveryStub{}
	h := HandlerWithAdministration(nil, AdminRuntime{Interpretation: store})
	for _, tc := range []struct {
		method, target, body string
		want                 int
	}{
		{"POST", "http://127.0.0.1/admin/sources/source/archive-recoveries", `{"source_id":"source","id":"one"}`, 200},
		{"POST", "http://127.0.0.1/admin/sources/source/archive-recoveries", `{"source_id":"foreign","id":"one"}`, 400},
		{"POST", "http://127.0.0.1/admin/sources/source/archive-recoveries", `{"source_id":"source","accept":true}`, 400},
		{"GET", "http://127.0.0.1/admin/sources/source/archive-recoveries", ``, 200},
		{"POST", "http://public.example/admin/sources/source/archive-recoveries", `{"source_id":"source"}`, 403},
	} {
		r := httptest.NewRequest(tc.method, tc.target, bytes.NewBufferString(tc.body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("%s %s: %d %s", tc.method, tc.target, w.Code, w.Body.String())
		}
	}
	if store.calls != 1 {
		t.Fatalf("rejected request reached store: %d", store.calls)
	}
	// The ordinary public listener has no administrative recovery operation.
	public := Handler(nil)
	r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/admin/sources/source/archive-recoveries", bytes.NewBufferString(`{"source_id":"source"}`))
	w := httptest.NewRecorder()
	public.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("public recovery route exposed: %d", w.Code)
	}
}
