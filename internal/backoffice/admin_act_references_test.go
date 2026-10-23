package backoffice

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acts"
)

type actReferencesFake struct {
	references []acts.ReferenceView
	err        error
}

func (f actReferencesFake) SourceReferences(context.Context, string, int) ([]acts.ReferenceView, error) {
	return f.references, f.err
}

func TestAdminActReferencesSection(t *testing.T) {
	recorded := time.Date(2026, 10, 23, 8, 0, 0, 0, time.UTC)
	act := int64(21)
	references := actReferencesFake{references: []acts.ReferenceView{
		{DocumentID: 60, VersionID: 12, NoticeURL: "https://example.test/novita/criticita", Ordinal: 1, Kind: "ordinanza", Passage: `mediante apposita ordinanza <script>alert(1)</script>`, DetectedAt: recorded, Outcome: "unique_candidate", ResolvedAt: &recorded,
			Candidates: []acts.Candidate{{RecordID: "9000660", Category: "ORDINANZE", ActNumber: 38, ActYear: 2026, RegistryYear: 2026, RegistryNumber: 660}}, ActURL: "https://albo.example.test/display/9000660", ActVersionID: &act,
			Clarifications: []acts.ClarificationView{{MeasureID: "municipal-extraction-v1:1:1", Outcome: "clarified", RecordedAt: recorded}, {MeasureID: "municipal-extraction-v1:1:5", Outcome: "conflict", RecordedAt: recorded}}},
		{DocumentID: 213, VersionID: 13, NoticeURL: "https://example.test/novita/monitoraggio", Ordinal: 1, Kind: "ordinanza", Passage: "con ordinanza comunale", DetectedAt: recorded, Outcome: "ambiguous", Reason: "several_matching_albo_records", Clarifications: []acts.ClarificationView{}},
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
	body := render(AdminRuntime{Registry: dashboardRegistry{}, ActReferences: references})
	for _, want := range []string{"Atti citati", "unique_candidate", "registro 2026/660", "https://albo.example.test/display/9000660", "clarified municipal-extraction-v1:1:1", "conflict", "ambiguous (several_matching_albo_records)"} {
		if !strings.Contains(body, want) {
			t.Fatalf("act references section missing %q", want)
		}
	}
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Fatal("reference passage was not escaped")
	}
	body = render(AdminRuntime{Registry: dashboardRegistry{}, ActReferences: actReferencesFake{err: context.DeadlineExceeded}})
	if !strings.Contains(body, "Atti citati") || !strings.Contains(body, "Elenco non disponibile.") || !strings.Contains(body, "Anteprima e raccolta") {
		t.Fatal("act reference list failure hid the source page")
	}
	// The list is read-only: no route edits references or outcomes.
	for _, path := range []string{"/admin/sources/source/act-references", "/admin/sources/source/act-references/1"} {
		page := httptest.NewRecorder()
		HandlerWithAdministration(nil, AdminRuntime{Registry: dashboardRegistry{}, ActReferences: references}).ServeHTTP(page, httptest.NewRequest("POST", "http://127.0.0.1"+path, strings.NewReader("{}")))
		if page.Code < 400 {
			t.Fatalf("POST %s accepted: %d", path, page.Code)
		}
	}
}
