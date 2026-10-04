package domain

import (
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"testing"
)

func TestAutomaticMeasureEvidenceDoesNotBorrowPageDates(t *testing.T) {
	text := "Pubblicato il 04/10/2026. Ordinanza 42/2026 del 03/10/2026: chiusura del Sottopasso sintetico."
	d := verificationDocument{Texts: []verificationLiteral{{Text: text, Selection: EvidenceSelection{ResourceURL: "https://fixture.example/notice", JSONPointer: "/contenuto"}}}}
	m := extraction.Measure{Kind: "closure", Subject: "Sottopasso sintetico", Evidence: []extraction.Evidence{{Field: "kind", ResourceURL: d.Texts[0].Selection.ResourceURL, Quote: "chiusura del Sottopasso sintetico"}, {Field: "subject", ResourceURL: d.Texts[0].Selection.ResourceURL, Quote: "chiusura del Sottopasso sintetico"}}}
	fields := measureSelectors(d, m)
	if fields["edition"].StartByte == 0 || text[fields["edition"].StartByte:fields["edition"].EndByte] != "03/10/2026" {
		t.Fatal("page date borrowed", fields)
	}
	if normalizedField("kind", text[fields["kind"].StartByte:fields["kind"].EndByte]) != "closure" {
		t.Fatal("kind evidence lost")
	}
	d.Texts[0].Text = "Pubblicato il 04/10/2026. Ordinanza 42/2026: chiusura del Sottopasso sintetico."
	if _, ok := measureSelectors(d, m)["edition"]; ok {
		t.Fatal("edition inferred from publication date")
	}
	d.Texts[0].Text = "Ordinanza 42/2026 del 03/10/2026 e Ordinanza 43/2026 del 04/10/2026: chiusura del Sottopasso sintetico."
	if _, ok := measureSelectors(d, m)["edition"]; ok {
		t.Fatal("ambiguous act silently selected")
	}
}
