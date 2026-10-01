package backoffice

import (
	"context"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/evaluation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/transport/httpapi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDashboardAssets(t *testing.T) {
	for _, base := range []string{"http://127.0.0.1", "https://iwa.tail123.ts.net"} {
		h := HandlerWithAdministration(nil, AdminRuntime{TailscaleOrigin: "https://iwa.tail123.ts.net", Operations: &operationFixture{}})
		for path, mime := range map[string]string{"/admin/assets/admin.css": "text/css", "/admin/assets/admin.js": "text/javascript", "/admin/": "text/html", "/admin/operations/jobs": "text/html"} {
			r := httptest.NewRequest("GET", base+path, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), mime) {
				t.Fatalf("%s: %d %s", path, w.Code, w.Body)
			}
			if !strings.Contains(w.Header().Get("Content-Security-Policy"), "style-src 'self'; script-src 'self'") {
				t.Fatal("asset policy missing")
			}
			r.Header.Set("Origin", "https://evil.example")
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 403 {
				t.Fatal("foreign origin accepted")
			}
			w = httptest.NewRecorder()
			httpapi.Handler(nil).ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1"+path, nil))
			if w.Code != 404 {
				t.Fatal("public dashboard exposed")
			}
		}
	}
}
func TestDashboardBrowser(t *testing.T) {
	node := os.Getenv("IWA_BROWSER_NODE")
	if node == "" {
		t.Skip("set IWA_BROWSER_NODE and IWA_PLAYWRIGHT_MODULE for browser acceptance")
	}
	h := httptest.NewServer(HandlerWithAdministration(nil, AdminRuntime{Alerts: alertsFixture{}, Registry: dashboardRegistry{}, Notifications: &notificationReportFake{}, Backups: &backupReportFake{}, Observations: &adminObservationFake{}, Evaluations: &adminEvaluationFake{}, TrialCosts: &trialCostFixture{}, Diagnostics: &diagnosticFixture{}, Operations: &operationFixture{}, Jobs: &relaunchFixture{}}))
	defer h.Close()
	script, err := filepath.Abs("../../scripts/check-dashboard.cjs")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(node, script, h.URL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("browser: %v\n%s", err, out)
	}
	t.Log(string(out))
}

func (s *operationFixture) Overview(_ context.Context, at time.Time) (operations.Overview, error) {
	if s.fail {
		return operations.Overview{}, errors.New("private dependency detail")
	}
	return operations.Overview{ObservedAt: at, SourceIssues: 2, FailedJobs: 12, PendingDocuments: 104}, nil
}
func TestDashboardOverviewPartialFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		h := HandlerWithAdministration(nil, AdminRuntime{Operations: &operationFixture{fail: fail}, Notifications: &notificationReportFake{fail: !fail}})
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1/admin/operations/", nil))
		body := w.Body.String()
		if w.Code != 200 || strings.Contains(body, "private dependency") || !strings.Contains(body, "Non disponibile") {
			t.Fatalf("untruthful overview: %s", body)
		}
		if !fail && (!strings.Contains(body, ">104<") || !strings.Contains(body, "?state=failed")) {
			t.Fatal("missing full count or drilldown")
		}
		if fail && strings.Contains(body, ">0<") {
			t.Fatal("failed summary rendered as zero")
		}
	}
}

type dashboardRegistry struct{ AdminRegistry }

func (dashboardRegistry) State(context.Context, string) (registry.State, error) {
	rev := 1
	return registry.State{Source: registry.Source{ID: "source", Territory: "Toscana"}, LatestRevision: 1, ActiveRevision: &rev}, nil
}
func (d dashboardRegistry) Sources(ctx context.Context) ([]registry.State, error) {
	s, _ := d.State(ctx, "source")
	return []registry.State{s}, nil
}
func (dashboardRegistry) Versions(context.Context, string) ([]registry.Version, error) {
	return []registry.Version{{Revision: 1, Actor: "fixture", Configuration: registry.Configuration{URL: "https://example.test/" + strings.Repeat("long-", 50), CheckSeconds: 600, DelaySeconds: 1800}}}, nil
}
func (dashboardRegistry) Events(context.Context, string) ([]registry.Event, error) {
	return []registry.Event{}, nil
}
func (dashboardRegistry) Intervals(context.Context, string) (registry.Intervals, error) {
	return registry.Intervals{}, nil
}
func (dashboardRegistry) InterpretationHistory(context.Context, string) ([]registry.InterpretationSuspension, error) {
	return nil, nil
}
func (dashboardRegistry) Disable(context.Context, string, int, string, bool) error { return nil }

func (dashboardRegistry) ReleaseReadiness(context.Context) (registry.ReleaseReadiness, error) {
	return registry.ReleaseReadiness{}, nil
}
func (dashboardRegistry) AcceptanceReviews(context.Context, string) ([]registry.AcceptanceReview, error) {
	return nil, nil
}
func (dashboardRegistry) Regressions(context.Context, string) ([]registry.Regression, error) {
	return nil, nil
}
func (dashboardRegistry) RecordRegression(context.Context, string, int, string, string, evaluation.Run) error {
	return nil
}

func TestSecondaryPageUnavailableHTML(t *testing.T) {
	h := HandlerWithAdministration(nil, AdminRuntime{})
	for _, path := range []string{"/admin/backups", "/admin/notifications", "/admin/observation-campaigns", "/admin/observation-campaigns/trial", "/admin/observation-campaigns/trial/report", "/admin/observation-campaigns/trial/cost-report?through=2026-09-18T00:00:00Z", "/admin/observation-campaigns/trial/budget-proposals", "/admin/processing-evaluations", "/admin/release-readiness", "/admin/sources/source/acceptance-reviews", "/admin/sources/source/regressions", "/admin/diagnostics/sources/source"} {
		r := httptest.NewRequest("GET", "http://127.0.0.1"+path, nil)
		r.Header.Set("Accept", "text/html")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 503 || !strings.Contains(w.Body.String(), "/admin/assets/admin.css") || !strings.Contains(w.Body.String(), "temporaneamente non disponibili") {
			t.Fatalf("unavailable %s: %d %s", path, w.Code, w.Body)
		}
	}
}
