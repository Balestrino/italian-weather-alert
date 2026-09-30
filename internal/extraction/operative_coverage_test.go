package extraction

import (
	"strings"
	"testing"
)

func TestRetainedClosureCompletionRequiresPositiveCoreOwnedCoordination(t *testing.T) {
	positive := "Restano al momento in vigore le disposizioni adottate con ordinanza comunale: chiusura al pubblico dei cimiteri e dell’area di sgambamento cani e divieto di attività nei parchi."
	for _, test := range []struct {
		name, text string
		core       int
		expected   int
	}{
		{"positive", positive, 0, 2},
		{"negated", "Non " + positive, 0, 0},
		{"conditional", "Se piove, " + positive, 0, 0},
		{"context only", positive, len("Restano"), 0},
		{"single subject", "Restano in vigore le disposizioni: chiusura al pubblico dei cimiteri.", 0, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := Window{Ordinal: 1, Text: test.text, ResourceURL: "https://example.test/notice", CoreStartByte: test.core, CoreEndByte: len(test.text)}
			got := supplementRetainedClosures(nil, w)
			if len(got) != test.expected {
				t.Fatalf("unsupported completion: %+v", got)
			}
			if len(got) > 0 {
				if again := supplementRetainedClosures(got, w); len(again) != len(got) {
					t.Fatal("duplicate supplementation")
				}
				for _, m := range got {
					if m.ValidFrom != nil || m.ValidUntil != nil || strings.Contains(m.Subject, "divieto") {
						t.Fatalf("scope leakage: %+v", m)
					}
				}
			}
		})
	}
}
func TestOrdinanceHeadingRejectionPreservesLegacyAndOperativeBody(t *testing.T) {
	text := "Oggetto : CHIUSURA dei sottopassi. VISTO il maltempo. ORDINA la chiusura dei cimiteri."
	w := Window{Ordinal: 1, Text: text, ResourceURL: "https://example.test/ordinance", CoreEndByte: len(text)}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["CHIUSURA dei sottopassi","ORDINA la chiusura dei cimiteri"],"measures":[{"kind":"closure","subject":"sottopassi","place":null,"temporal_candidates":[],"evidence_refs":{"kind":[0],"subject":[0],"place":[]}},{"kind":"closure","subject":"cimiteri","place":null,"temporal_candidates":[],"evidence_refs":{"kind":[1],"subject":[1],"place":[]}}]}`
	got, err := ParseWindow(raw, w)
	if err != nil || len(got) != 1 || got[0].Subject != "cimiteri" {
		t.Fatalf("operative body lost: %+v %v", got, err)
	}
	contextOnly := w
	contextOnly.CoreEndByte = strings.Index(text, "VISTO")
	if _, err := ParseWindow(raw, contextOnly); err != ErrEvidence {
		t.Fatalf("heading acquired context-owned operative fact: %v", err)
	}
	w.legacyLiteral = true
	legacy, err := ParseWindow(raw, w)
	if err != nil || len(legacy) != 2 {
		t.Fatalf("legacy changed: %+v %v", legacy, err)
	}
}
