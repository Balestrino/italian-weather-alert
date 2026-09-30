package extraction

import (
	"strings"
	"testing"
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
