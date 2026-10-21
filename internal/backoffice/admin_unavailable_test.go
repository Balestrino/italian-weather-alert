package backoffice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
)

type unavailableTargetsFake struct {
	targets  []acquisition.UnavailableTarget
	requests []acquisition.DispositionRequest
	err      error
}

func (f *unavailableTargetsFake) ListUnavailable(context.Context, string) ([]acquisition.UnavailableTarget, error) {
	return f.targets, nil
}
func (f *unavailableTargetsFake) RecordDisposition(_ context.Context, request acquisition.DispositionRequest) (acquisition.TargetDisposition, error) {
	f.requests = append(f.requests, request)
	if f.err != nil {
		return acquisition.TargetDisposition{}, f.err
	}
	return acquisition.TargetDisposition{ID: 7, Decision: request.Decision, Actor: request.Actor}, nil
}

func TestAdminUnavailableTargetsRenderAndRecordDecisions(t *testing.T) {
	seen := time.Date(2026, 10, 14, 6, 32, 15, 597181000, time.UTC)
	store := &unavailableTargetsFake{targets: []acquisition.UnavailableTarget{
		{URL: "https://example.test/novita/renamed", Configuration: 1, HTTPStatus: 404, LastSeenAt: seen, Classification: "withdrawn"},
		{URL: "https://example.test/novita/coc", Configuration: 1, HTTPStatus: 410, LastSeenAt: seen, MeasureState: "unresolved", SupportsMeasure: true, DerivedMeasure: true, Classification: "blocking"},
		{URL: "https://example.test/novita/aged", Configuration: 1, HTTPStatus: 404, LastSeenAt: seen, MeasureState: "aged_unresolved", DerivedMeasure: true, Classification: "withdrawn"},
		{URL: `https://example.test/novita/"><script>alert(1)</script>`, Configuration: 1, HTTPStatus: 404, LastSeenAt: seen, Classification: "excluded"},
	}}
	handler := HandlerWithAdministration(nil, AdminRuntime{Registry: dashboardRegistry{}, UnavailableTargets: store})
	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest("GET", "http://127.0.0.1/admin/sources/source", nil))
	body := page.Body.String()
	if page.Code != 200 || !strings.Contains(body, "URL non disponibili") || !strings.Contains(body, "Ritirato dagli elenchi") || !strings.Contains(body, "misura in corso o irrisolta") || !strings.Contains(body, "misura ricavata, stato aged_unresolved") || strings.Count(body, `name="acknowledge_measures"><option`) != 2 || !strings.Contains(body, "Riprendi il controllo dell'URL") {
		t.Fatalf("unavailable section: %d %s", page.Code, body)
	}
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Fatal("target URL was not escaped")
	}
	if !strings.Contains(body, `value="2026-10-14T06:32:15.597181Z"`) {
		t.Fatal("observed last-seen time not carried by the form")
	}

	post := func(values url.Values) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest("POST", "http://127.0.0.1/admin/sources/source/unavailable-targets/decisions", strings.NewReader(values.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}
	form := url.Values{"configuration": {"1"}, "url": {"https://example.test/novita/renamed"}, "observed_last_seen_at": {"2026-10-14T06:32:15.597181Z"}, "decision": {"exclude_stale_unavailable"}, "acknowledge_measures": {"false"}, "actor": {"operator"}, "reason": {"absent from complete listings"}, "evidence": {`{"listing_pages": 14}`}}
	if response := post(form); response.Code != 200 || len(store.requests) != 1 {
		t.Fatalf("decision not recorded: %d %s", response.Code, response.Body.String())
	}
	recorded := store.requests[0]
	if recorded.SourceID != "source" || recorded.Configuration != 1 || recorded.Decision != acquisition.DecisionExclude || !recorded.ObservedLastSeenAt.Equal(seen) || string(recorded.Evidence) != `{"listing_pages": 14}` || recorded.AcknowledgeMeasures {
		t.Fatalf("decision fields: %+v", recorded)
	}
	store.err = acquisition.ErrTargetChanged
	if response := post(form); response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "target_changed") {
		t.Fatalf("stale decision: %d %s", response.Code, response.Body.String())
	}
	form.Del("actor")
	if response := post(form); response.Code != http.StatusBadRequest || len(store.requests) != 2 {
		t.Fatalf("missing operator accepted: %d", response.Code)
	}
}

type followedTargetsFake struct {
	targets []acquisition.FollowedTarget
	err     error
}

func (f followedTargetsFake) FollowedTargets(context.Context, string, time.Time) ([]acquisition.FollowedTarget, error) {
	return f.targets, f.err
}

func TestAdminFollowedTargetsSection(t *testing.T) {
	derived := time.Date(2026, 10, 21, 20, 43, 0, 0, time.UTC)
	followed := followedTargetsFake{targets: []acquisition.FollowedTarget{
		{URL: "https://example.test/novita/coc", MeasureState: "unresolved", RelevanceState: "relevant", AgeDays: 40, LimitDays: 90, InterpretationStateAt: &derived, Recheck: "unresolved"},
		{URL: `https://example.test/novita/"><script>alert(1)</script>`, MeasureState: "aged_unresolved", RelevanceState: "relevant", ExplicitReference: true, AgeDays: 120, LimitDays: 90, Recheck: "not_rechecked"},
	}}
	render := func(runtime AdminRuntime) string {
		t.Helper()
		page := httptest.NewRecorder()
		HandlerWithAdministration(nil, runtime).ServeHTTP(page, httptest.NewRequest("GET", "http://127.0.0.1/admin/sources/source", nil))
		if page.Code != 200 {
			t.Fatalf("source page: %d", page.Code)
		}
		return page.Body.String()
	}
	body := render(AdminRuntime{Registry: dashboardRegistry{}, FollowedTargets: followed})
	for _, want := range []string{"Documenti seguiti", "misura con fine ignota, entro il limite", "non più ricontrollato", "40 giorni (limite 90)", "riferimento esplicito", "non ancora"} {
		if !strings.Contains(body, want) {
			t.Fatalf("followed section missing %q", want)
		}
	}
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Fatal("followed target URL was not escaped")
	}
	// A read failure keeps the rest of the source page available.
	body = render(AdminRuntime{Registry: dashboardRegistry{}, FollowedTargets: followedTargetsFake{err: context.DeadlineExceeded}})
	if !strings.Contains(body, "Documenti seguiti") || !strings.Contains(body, "Elenco non disponibile.") || !strings.Contains(body, "Anteprima e raccolta") {
		t.Fatal("followed list failure hid the source page")
	}
}
