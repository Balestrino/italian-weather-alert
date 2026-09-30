package notifications

import (
	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestReportPreservesUnknownFactsAndPartialReopening(t *testing.T) {
	at := time.Date(2026, 9, 26, 19, 0, 0, 0, time.UTC)
	subject := "area di sgambamento cani"
	locator := "subject"
	tomorrow := "2026-09-27"
	s := ReportSnapshot{ObservedAt: at, Regions: []ReportRegion{{Name: "Toscana", Municipalities: []ReportMunicipality{{ISTAT: "050004", Name: "Calcinaia"}}, Warnings: []publicquery.RegionalWarning{{Level: "unknown", Zone: "A4", Validity: publicquery.Temporal{Precision: "date", Date: &tomorrow}, Quality: publicquery.Quality{Interpretation: publicquery.Dimension{Limitations: []string{"newer bulletin unsupported"}}}}}, Measures: []publicquery.Measure{{MunicipalityISTAT: "050004", Action: "closure", Evidence: []publicquery.Evidence{{Locator: &locator, Passage: &subject}}, Validity: publicquery.Temporal{Precision: "unknown"}, NewerUninterpretedDocumentIDs: []string{"new"}}, {MunicipalityISTAT: "050004", Action: "reopening", Place: &subject, Validity: publicquery.Temporal{Precision: "unknown"}}}}}}
	mail, err := RenderReport(s, "Europe/Rome")
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{"Domani", "non determinato", "chiusura", "riapertura", subject, "Associazione alle zone non disponibile", "Documenti successivi non interpretati", "newer bulletin unsupported", "Validità incerta"} {
		if !strings.Contains(mail.Body, wanted) {
			t.Fatal("missing", wanted)
		}
	}
	if strings.Contains(mail.Body, "livello verde") {
		t.Fatal("unknown became green")
	}
	empty, _ := RenderReport(ReportSnapshot{ObservedAt: at}, "Europe/Rome")
	if !strings.Contains(empty.Body, "Nessun territorio abilitato") {
		t.Fatal("empty scope")
	}
}
func TestReportTruncationIsBoundedAndDeclared(t *testing.T) {
	s := ReportSnapshot{ObservedAt: time.Now(), Regions: []ReportRegion{{Name: strings.Repeat("è", MaxReportBytes), Municipalities: []ReportMunicipality{{ISTAT: "050004", Name: "Calcinaia"}}}}}
	m, err := RenderReport(s, "Europe/Rome")
	if err != nil || len(m.Body) > MaxReportBytes || !utf8.ValidString(m.Body) || !strings.Contains(m.Body, "REPORT TRONCATO") || !strings.Contains(m.Body, "1 comuni in raccolta") {
		t.Fatal("invalid truncation")
	}
}

func TestReportConflictAndPhaseValidityRemainExplicit(t *testing.T) {
	at := time.Date(2026, 9, 26, 19, 0, 0, 0, time.UTC)
	past := "2026-09-25"
	s := ReportSnapshot{ObservedAt: at, Regions: []ReportRegion{{Name: "Toscana", Municipalities: []ReportMunicipality{{ISTAT: "050004", Name: "Calcinaia"}}, Measures: []publicquery.Measure{{MunicipalityISTAT: "050004", Action: "closure", Validity: publicquery.Temporal{Precision: "unknown", Original: "Date di inizio conflittuali: 20 agosto / 20 settembre"}}}, Phases: []publicquery.OperationalPhase{{MunicipalityISTAT: "050004", Phase: "expired phase", Validity: publicquery.Temporal{Precision: "date", Date: &past}}, {MunicipalityISTAT: "050004", Phase: "COC attivo", Validity: publicquery.Temporal{Precision: "condition", Original: "fino al termine dell’emergenza"}}}}}}
	mail, err := RenderReport(s, "Europe/Rome")
	if err != nil || !strings.Contains(mail.Body, "Date di inizio conflittuali") || !strings.Contains(mail.Body, "fino al termine dell’emergenza") || strings.Contains(mail.Body, "expired phase") {
		t.Fatal("temporal conflict or phase validity lost", err)
	}
}
