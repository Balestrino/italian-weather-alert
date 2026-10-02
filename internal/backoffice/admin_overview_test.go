package backoffice

import (
	"context"
	"errors"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
)

type detailedDashboardFixture struct {
	operationFixture
	period        string
	detailFailure bool
}

func (f *detailedDashboardFixture) Dashboard(_ context.Context, period string, at time.Time) (operations.Dashboard, error) {
	f.period = period
	if f.detailFailure {
		return operations.Dashboard{}, errors.New("private query details")
	}
	return operations.Dashboard{
		Overview: operations.Overview{ObservedAt: at, PendingDocuments: 5, FailedJobs: 2},
		Period:   period, HistoryFrom: at.Add(-time.Hour),
		Jobs:      operations.JobCounts{Queued: 110, Running: 1, Failed: 2, Archived: 7},
		Groups:    []operations.JobGroup{{Queue: "inference", Kind: "ocr_resource", JobCounts: operations.JobCounts{Queued: 110, Failed: 2}}},
		Reasons:   []operations.JobReason{{Queue: "inference", Kind: "ocr_resource", State: "failed", Code: "<script>bad</script>", Count: 2, LastObserved: at}},
		Documents: []operations.DocumentBreakdown{{SourceID: "s&x", Total: 5, Unscheduled: 2, Queued: 3}},
		History:   []operations.HistoryBucket{{At: at.Add(-time.Hour), Succeeded: 3, Retry: 2, Failed: 1, Abandoned: 1}, {At: at}},
	}, nil
}

func TestDetailedDashboardRenderingAndValidation(t *testing.T) {
	f := &detailedDashboardFixture{}
	h := HandlerWithAdministration(nil, AdminRuntime{Operations: f})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1/admin/operations/?period=7d", nil))
	body := w.Body.String()
	for _, want := range []string{"Stato attuale dei job", "Storico dei tentativi", "Perché i documenti", ">110<", "selected", `role="img"`, `class="chart-success"`, `error_code=%3Cscript%3Ebad%3C%2Fscript%3E`, `kind=ocr_resource&amp;queue=inference&amp;state=archived`, `source_id=s%26x`, "&lt;script&gt;bad&lt;/script&gt;", "Un job può contribuire più volte"} {
		if w.Code != 200 || !strings.Contains(body, want) {
			t.Fatalf("missing %q: %d %s", want, w.Code, body)
		}
	}
	if f.period != "7d" || strings.Contains(body, "<script>bad") || strings.Contains(body, "ZgotmplZ") {
		t.Fatal("unsafe rendering or lost period")
	}
	for _, query := range []string{"period=bad", "period=24h&period=7d", "unexpected=1"} {
		w = httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1/admin/operations/?"+query, nil))
		if w.Code != 400 {
			t.Fatalf("accepted %s", query)
		}
	}
	f.detailFailure = true
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1/admin/operations/", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), ">104<") || !strings.Contains(w.Body.String(), "Non disponibile") || strings.Contains(w.Body.String(), "private query") {
		t.Fatal("partial failure hid summaries or exposed details")
	}
}

func TestHistoryChartEmptyAndStackScale(t *testing.T) {
	chart := makeHistoryChart([]operations.HistoryBucket{{Succeeded: 3, Retry: 2, Failed: 1, Abandoned: 1}, {}}, "24h")
	if chart.Maximum != 7 || chart.Succeeded != 3 || chart.Retry != 2 || chart.Failed != 1 || chart.Abandoned != 1 {
		t.Fatalf("wrong chart totals: %#v", chart)
	}
	var height float64
	for _, r := range chart.Bars[0].Rects {
		height += r.Height
		if r.Y < 39.99 || r.Height < 0 {
			t.Fatal("bar outside chart")
		}
	}
	if height < 179.99 || height > 180.01 {
		t.Fatal("stack does not share a common scale")
	}
	zero := makeHistoryChart([]operations.HistoryBucket{{}}, "7d")
	for _, r := range zero.Bars[0].Rects {
		if r.Height != 0 || r.Y != 220 {
			t.Fatal("empty chart has nonzero bars")
		}
	}
}

func TestOperationsOverviewBrowser(t *testing.T) {
	node := os.Getenv("IWA_BROWSER_NODE")
	if node == "" {
		t.Skip("set IWA_BROWSER_NODE and IWA_PLAYWRIGHT_MODULE for browser acceptance")
	}
	h := httptest.NewServer(HandlerWithAdministration(nil, AdminRuntime{Operations: &detailedDashboardFixture{}}))
	defer h.Close()
	script, err := filepath.Abs("../../scripts/check-operations-dashboard.cjs")
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(node, script, h.URL).CombinedOutput()
	if err != nil {
		t.Fatalf("browser: %v\n%s", err, out)
	}
	t.Log(string(out))
}
