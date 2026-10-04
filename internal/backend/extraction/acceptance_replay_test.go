package extraction

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
)

func TestNullPlaceReferencesRemainBounded(t *testing.T) {
	text := "Le strade sono state liberate."
	w := Window{Ordinal: 1, Total: 1, DocumentVersionID: 1, ResourceURL: "https://example.org/notice", Role: "original", ContextEndByte: len(text), CoreEndByte: len(text), Text: text}
	for _, tc := range []struct {
		refs  string
		valid bool
	}{{"[0]", true}, {"[1]", false}, {"[-1]", false}, {"[0,0]", false}} {
		raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["Le strade sono state liberate."],"measures":[{"kind":"operational_update","subject":"strade","place":null,"evidence_refs":{"kind":[0],"subject":[0],"place":` + tc.refs + `},"temporal_candidates":[]}]}`
		got, err := ParseWindow(raw, w)
		if (err == nil) != tc.valid || (tc.valid && len(got) != 1) {
			t.Fatalf("refs %s: %+v %v", tc.refs, got, err)
		}
		w.legacyLiteral = true
		if _, err = ParseWindow(raw, w); err == nil {
			t.Fatal("legacy parser semantics changed")
		}
		w.legacyLiteral = false
	}
}

func partialReopeningFixture() (string, string) {
	text := "Il sottopasso di via del Bosco, chiuso nella serata di ieri per acqua, è stato riaperto alla circolazione. Restano al momento in vigore le disposizioni: chiusura al pubblico della biblioteca e dei giardini e divieto di attività nei parchi, nelle ciclopiste in riva d’Arno e negli orti. Il Centro Operativo Comunale resta attivo."
	quote := "Restano al momento in vigore le disposizioni: chiusura al pubblico della biblioteca e dei giardini e divieto di attività nei parchi, nelle ciclopiste in riva d'Arno e negli orti."
	var candidates []map[string]any
	for _, place := range []string{"parchi", "ciclopiste in riva d'Arno", "orti"} {
		candidates = append(candidates, map[string]any{"kind": "prohibition", "subject": "attività", "place": place, "evidence_refs": map[string]any{"kind": []int{1}, "subject": []int{1}, "place": []int{1}}, "temporal_candidates": []any{}})
	}
	candidates = append(candidates,
		map[string]any{"kind": "operational_update", "subject": "sottopasso di via del Bosco", "place": "via del Bosco", "evidence_refs": map[string]any{"kind": []int{0, 1, 2}, "subject": []int{0}, "place": []int{0}}, "temporal_candidates": []any{map[string]any{"field": "valid_from", "original_expression": "nella serata di ieri", "evidence_refs": []int{0}}}},
		map[string]any{"kind": "activation", "subject": "Centro Operativo Comunale", "place": nil, "evidence_refs": map[string]any{"kind": []int{2}, "subject": []int{2}, "place": []int{}}, "temporal_candidates": []any{}},
	)
	raw, _ := json.Marshal(map[string]any{"envelope_version": CompactEnvelopeVersionV2, "window_ordinal": 1, "evidence": []string{strings.Split(text, ". ")[0], quote, "Il Centro Operativo Comunale resta attivo."}, "measures": candidates})
	return text, string(raw)
}

func TestPartialReopeningPreservesTypographicPlacesAndOtherRestrictions(t *testing.T) {
	text, raw := partialReopeningFixture()
	w := Window{Ordinal: 1, Total: 1, DocumentVersionID: 1, ResourceURL: "https://example.test/notice", Role: "original", ContextEndByte: len(text), CoreEndByte: len(text), Text: text}
	got, err := ParseWindow(raw, w)
	if err != nil {
		t.Fatal(err)
	}
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: w.ResourceURL, Text: text}}}
	got, err = Merge([][]Measure{got}, content, 1)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, m := range got {
		counts[m.Kind]++
		if m.Kind == "prohibition" && (m.Place == nil || !strings.Contains(text, *m.Place)) {
			t.Fatalf("literal listed place lost: %+v", m)
		}
		if m.Kind == "reopening" && (m.Subject != "sottopasso di via del Bosco" || m.ValidFrom != nil || m.ValidUntil != nil || len(m.TemporalCandidates) != 0) {
			t.Fatalf("old closure time transferred to reopening: %+v", m)
		}
	}
	if counts["reopening"] != 1 || counts["closure"] != 2 || counts["prohibition"] != 3 || counts["activation"] != 1 || counts["operational_update"] != 0 {
		t.Fatalf("partial update lost its independent scopes: %+v", counts)
	}
}

func TestNamedReopeningRequiresCompletedCoreOwnedSentence(t *testing.T) {
	positive := "Il sottopasso di via del Bosco, chiuso nella serata di ieri per acqua, è stato riaperto alla circolazione."
	for _, tc := range []struct {
		name, text string
		core       int
		legacy     bool
		want       int
	}{
		{"completed", positive, 0, false, 1},
		{"without aside", "La strada del Parco è stata riaperta alla circolazione.", 0, false, 1},
		{"consecutive subjects", positive + " Il ponte del Parco è stato riaperto alla circolazione.", 0, false, 2},
		{"qualified circulation", strings.Replace(positive, "alla circolazione.", "alla circolazione solo se autorizzati.", 1), 0, false, 0},
		{"negated", strings.Replace(positive, "è stato", "non è stato", 1), 0, false, 0},
		{"negated without aside", "Il sottopasso di via del Bosco non è stato riaperto alla circolazione.", 0, false, 0},
		{"uncertain", "Il ponte del Parco forse è stato riaperto alla circolazione.", 0, false, 0},
		{"prospective", strings.Replace(positive, "è stato riaperto", "sarà riaperto", 1), 0, false, 0},
		{"conditional", "Se migliora il tempo, " + positive, 0, false, 0},
		{"quoted", "Il documento cita «" + positive + "»", 0, false, 0},
		{"quoted later sentence", "Il documento cita «Testo precedente. " + positive + "»", 0, false, 0},
		{"double quoted later sentence", "Il documento cita \"Testo precedente. " + positive + "\"", 0, false, 0},
		{"heading", "Oggetto: aggiornamento. " + positive + " VISTO il maltempo.", 0, false, 0},
		{"context only", positive, strings.Index(positive, "è stato"), false, 0},
		{"legacy", positive, 0, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := Window{Ordinal: 1, ResourceURL: "https://example.test/notice", Page: 2, Text: tc.text, CoreStartByte: tc.core, CoreEndByte: len(tc.text), legacyLiteral: tc.legacy}
			got, err := ParseWindow(`{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":[],"measures":[]}`, w)
			if err != nil || len(got) != tc.want {
				t.Fatalf("unsupported reopening: %+v %v", got, err)
			}
			if len(got) > 0 && (got[0].ValidFrom != nil || got[0].ValidUntil != nil || got[0].Place != nil || len(supplementNamedReopenings(got, w)) != len(got)) {
				t.Fatalf("invented time/place or duplicated reopening: %+v", got)
			}
			for _, m := range got {
				for _, e := range m.Evidence {
					if e.Page == nil || *e.Page != 2 || e.ResourceURL != w.ResourceURL || !strings.Contains(w.Text, e.Quote) {
						t.Fatalf("original page evidence lost: %+v", e)
					}
				}
			}
		})
	}
}
func TestTemporalSupplementDoesNotCrossOperativeTransition(t *testing.T) {
	text := "Il Sindaco dispone a partire dalle ore 0.00 di giovedì 10 agosto la chiusura dei cimiteri. E’ stato inoltre sancito il divieto nei parchi a partire dalle ore 1.00 di venerdì 11 settembre."
	w := Window{Ordinal: 1, ResourceURL: "https://example.org/notice", Role: "original", Text: text}
	m := Measure{Kind: "prohibition", Subject: "parchi"}
	supplementLocalStartCandidates(&m, w)
	if len(m.TemporalCandidates) != 1 || !strings.Contains(m.TemporalCandidates[0].OriginalExpression, "settembre") {
		t.Fatalf("unrelated period leaked: %+v", m)
	}
	// A flattened list without a full stop must have the same scope boundary.
	w.Text = strings.Replace(text, "cimiteri. E’", "cimiteri E’", 1)
	m = Measure{Kind: "prohibition", Subject: "parchi"}
	supplementLocalStartCandidates(&m, w)
	if len(m.TemporalCandidates) != 1 || !strings.Contains(m.TemporalCandidates[0].OriginalExpression, "settembre") {
		t.Fatalf("flattened clause leaked: %+v", m)
	}
}

// The follow-up provider run names Fornacette in one listed location, not in
// the general ban. The narrow provider variant must not scope that ban.
func TestListedPlaceStillAppliesToItsOwnSubject(t *testing.T) {
	text := "È disposto il divieto di qualsiasi attività all'aperto: nei parchi pubblici; negli orti sociali di Fornacette."
	w := Window{Ordinal: 1, Total: 1, DocumentVersionID: 1, ResourceURL: "https://example.org/notice", Role: "original", ContextEndByte: len(text), CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["È disposto il divieto di qualsiasi attività all'aperto:","nei parchi pubblici; negli orti sociali di Fornacette"],"measures":[{"kind":"prohibition","subject":"orti sociali","place":"Fornacette","evidence_refs":{"kind":[0],"subject":[1],"place":[1]},"temporal_candidates":[]}]}`
	measures, err := ParseWindow(raw, w)
	if err != nil {
		t.Fatal(err)
	}
	if len(measures) != 1 || measures[0].Place == nil || *measures[0].Place != "Fornacette" {
		t.Fatalf("listed subject lost its supported place: %+v", measures)
	}
}
func TestWholeListQuoteDoesNotScopeGeneralSubject(t *testing.T) {
	text := "È disposto il divieto di qualsiasi attività all'aperto: nei parchi pubblici; negli orti sociali di Fornacette."
	w := Window{Ordinal: 1, Total: 1, DocumentVersionID: 1, ResourceURL: "https://example.org/notice", Role: "original", ContextEndByte: len(text), CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["È disposto il divieto di qualsiasi attività all'aperto: nei parchi pubblici; negli orti sociali di Fornacette"],"measures":[{"kind":"prohibition","subject":"qualsiasi attività all'aperto","place":"Fornacette","evidence_refs":{"kind":[0],"subject":[0],"place":[0]},"temporal_candidates":[]}]}`
	measures, err := ParseWindow(raw, w)
	if err != nil {
		t.Fatal(err)
	}
	if len(measures) != 1 || measures[0].Place != nil || !strings.Contains(measures[0].Subject, "parchi pubblici") || !strings.Contains(measures[0].Subject, "orti sociali di Fornacette") {
		t.Fatalf("one listed place narrowed a general subject: %+v", measures)
	}
}
