package acquisition

import (
	"strings"
	"testing"
	"time"
)

func TestExplicitRegionalTableAndRejection(t *testing.T) {
	body := `<p>Emissione di Lunedì, 13 Aprile 2026, ore 12.48</p><table><tr><th>ZONE</th><th>RISCHIO</th><th>VALIDITÀ</th><th>CRITICITÀ</th></tr><tr><td>A4 A5</td><td>VENTO</td><td>dalle 12:00 del 13/04/2026 alle 23:59 del 13/04/2026</td><td>GIALLO</td></tr></table>`
	p, err := ProjectRegionalHTML("criticality", []byte(body))
	if err != nil || len(p.Facts) != 2 {
		t.Fatalf("rows: %+v %v", p, err)
	}
	if p.Facts[0].Level != "yellow" || p.Facts[0].Start.Format("15:04") != "10:00" || p.Facts[0].End.Format("15:04") != "21:59" {
		t.Fatal("level/time mapping")
	}
	for _, invalid := range []string{strings.Replace(body, "GIALLO", "blu", 1), strings.Replace(body, "VENTO", "UNKNOWN", 1), strings.Replace(body, "A4 A5", "Z9", 1), body + "Criticità previste: NESSUNA"} {
		if _, err := ProjectRegionalHTML("criticality", []byte(invalid)); err == nil {
			t.Fatal("unsupported/conflicting table accepted")
		}
	}
	if _, _, err := rowInterval("dalle 02:30 del 25/10/2026 alle 04:00 del 25/10/2026"); err == nil {
		t.Fatal("ambiguous Rome time accepted")
	}
	if _, err := parseRegionalDay("31 Febbraio 2026"); err == nil {
		t.Fatal("invalid date accepted")
	}
}
func TestMonitoringNoEventHasNoColor(t *testing.T) {
	p, err := ProjectRegionalHTML("monitoring", []byte(`<b>NESSUN AVVISO IN CORSO DI VALIDITÀ O EVENTO IN CORSO</b>`))
	if err != nil || p.Statement == "" || len(p.Facts) != 0 {
		t.Fatalf("%+v %v", p, err)
	}
}

// This is a reviewed transcription of the retained official non-green PDF,
// not a synthetic numeric-date table. Its expected zones/risks are REG-NONGREEN.
func TestItalianRegionalIntervals(t *testing.T) {
	for _, tc := range []struct{ raw, start, end string }{
		{"dalle ore 12.00 alle ore 23.59 di Lunedì, 13 Aprile 2026", "2026-04-13T10:00:00Z", "2026-04-13T21:59:00Z"},
		{"dalle ore 12.00 Lunedì, 13 Aprile 2026 alle ore 23.59 Martedì, 14 Aprile 2026", "2026-04-13T10:00:00Z", "2026-04-14T21:59:00Z"},
	} {
		start, end, err := rowInterval(tc.raw)
		if err != nil || start.Format(time.RFC3339) != tc.start || end.Format(time.RFC3339) != tc.end {
			t.Fatalf("%s: %v %v %v", tc.raw, start, end, err)
		}
	}
	for _, raw := range []string{
		"dalle ore 12.00 alle ore 23.59 di 31 Febbraio 2026",
		"dalle ore 23.59 alle ore 12.00 di 13 Aprile 2026",
		"dalle ore 02.30 alle ore 04.00 di 25 Ottobre 2026",
		"dalle ore 02.30 alle ore 04.00 di 29 Marzo 2026",
		"dalle ore 24.00 alle ore 25.00 di 13 Aprile 2026",
		"dalle ore 12.00 alle ore 23.59 di domani",
		"emissione 12 Aprile 2026 dalle ore 12.00 alle ore 23.59",
	} {
		if _, _, err := rowInterval(raw); err == nil {
			t.Fatalf("accepted ambiguous/invalid time: %s", raw)
		}
	}
}
