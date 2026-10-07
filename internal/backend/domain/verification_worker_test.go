package domain

import (
	"context"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"testing"
	"time"
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

func TestAutomaticRiskSelectsLiteralValidityWithoutInventingZoneOrIssuance(t *testing.T) {
	d := verificationDocument{Version: documents.Version{ID: 1, Metadata: json.RawMessage(`{"kind":"risk"}`)}, Source: verificationSourceState{ID: "platform", Territory: "050004"}, Texts: []verificationLiteral{
		{Text: "arancione", Selection: EvidenceSelection{JSONPointer: "/stati_allerta/0/oggi/allerta"}},
		{Text: "Temporali", Selection: EvidenceSelection{JSONPointer: "/stati_allerta/0/tipologia"}},
		{Text: "07/10/2026", Selection: EvidenceSelection{JSONPointer: "/stati_allerta/0/oggi/data_bollettino"}},
		{Text: "Valido dalle ore 17.00 di Mercoledì, 07 Ottobre 2026 alle ore 17.00 di Giovedì, 08 Ottobre 2026", Selection: EvidenceSelection{JSONPointer: "/stati_allerta/0/oggi/validita_cfr"}},
	}}
	w := VerificationWorker{}
	requests, err := w.requests(context.Background(), d, nil, nil, nil, time.Now())
	if err != nil || len(requests) != 1 {
		t.Fatal(requests, err)
	}
	fields := requests[0].Candidate.Fields
	if len(fields) != 3 || fields["validity"].JSONPointer == "" || fields["zone"].JSONPointer != "" || fields["issuance"].JSONPointer != "" {
		t.Fatal("invented identity or lost validity", fields)
	}
	primary := verificationDocument{Source: verificationSourceState{Product: "criticality"}, Texts: []verificationLiteral{{Text: "BOLLETTINO DI VALUTAZIONE DELLE CRITICITÀ Emissione di Mercoledì, 07 Ottobre 2026 , ore 13.10", Selection: EvidenceSelection{ResourceURL: "https://regional.example"}}}}
	selected := regionalEditionSelectors(primary)
	if len(selected) != 2 || selected["issuance"].EndByte == 0 || selected["product"].EndByte == 0 {
		t.Fatal("primary literal edition lost", selected)
	}
}
