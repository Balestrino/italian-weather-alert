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

func TestRoadClearanceRetainsExceptionAndRejectsUnsupportedCompletion(t *testing.T) {
	positive := "AGGIORNAMENTO: Si informa la cittadinanza che tutte le strade interessate dai detriti sono state liberate grazie alle squadre. Il tunnel del Parco resta chiuso."
	for _, test := range []struct {
		name, text string
		core       int
		legacy     bool
		want       int
	}{
		{"completed", positive, 0, false, 2},
		{"negated", strings.Replace(positive, "sono state liberate", "non sono state liberate", 1), 0, false, 1},
		{"conditional", "Se migliora il tempo, " + positive, 0, false, 1},
		{"prospective", strings.Replace(positive, "sono state liberate", "saranno liberate", 1), 0, false, 1},
		{"context only", positive, strings.Index(positive, "Il tunnel"), false, 1},
		{"legacy", positive, 0, true, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := Window{Ordinal: 1, Text: test.text, ResourceURL: "https://example.test/notice", CoreStartByte: test.core, CoreEndByte: len(test.text)}
			w.legacyLiteral = test.legacy
			raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["Il tunnel del Parco resta chiuso"],"measures":[{"kind":"closure","subject":"tunnel del Parco","place":null,"temporal_candidates":[],"evidence_refs":{"kind":[0],"subject":[0],"place":[]}}]}`
			got, err := ParseWindow(raw, w)
			if err != nil || len(got) != test.want || got[0].Kind != "closure" {
				t.Fatalf("exception or ownership lost: %+v %v", got, err)
			}
			if len(got) == 2 {
				update := got[1]
				if update.Kind != "operational_update" || update.ValidFrom != nil || update.ValidUntil != nil || update.Place != nil || len(update.Evidence) != 2 {
					t.Fatalf("unsupported reopening, time or place: %+v", update)
				}
				if again := supplementRoadClearance(got, w); len(again) != len(got) {
					t.Fatal("duplicate clearance update")
				}
			}
		})
	}
}

func TestDispositiveClosureListPreservesLiteralSubjectsAndTemporalScope(t *testing.T) {
	body := "ORDINA in via contingibile e urgente, a partire dalle ore 17.00 del giorno 12 ottobre 2026, fino al perdurare dell’emergenza: - la chiusura al pubblico: - della biblioteca comunale; - dei giardini comunali; - il divieto di attività nei parchi; DISPONE l'attivazione delle funzioni di supporto."
	for _, test := range []struct {
		name, text string
		core       int
		legacy     bool
		want       int
	}{
		{"dispositive", body, 0, false, 2},
		{"negated", "Non " + body, 0, false, 0},
		{"conditional", "Se peggiora, " + body, 0, false, 0},
		{"quoted", "Il documento cita «" + body + "»", 0, false, 0},
		{"heading", "Oggetto: chiusura al pubblico della biblioteca e dei giardini.", 0, false, 0},
		{"preamble", "Si considera: - la chiusura al pubblico: - della biblioteca comunale; - dei giardini comunali;", 0, false, 0},
		{"context only", body, strings.Index(body, "- la chiusura"), false, 0},
		{"legacy", body, 0, true, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := Window{Ordinal: 1, Text: test.text, ResourceURL: "https://example.test/ordinance", Page: 2, CoreStartByte: test.core, CoreEndByte: len(test.text)}
			w.legacyLiteral = test.legacy
			got, err := ParseWindow(`{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":[],"measures":[]}`, w)
			closures := []Measure{}
			for _, m := range got {
				if m.Kind == "closure" {
					closures = append(closures, m)
				}
			}
			if err != nil || len(closures) != test.want {
				t.Fatalf("unsupported completion: %+v %v", got, err)
			}
			if len(closures) == 2 {
				if closures[0].Subject != "biblioteca comunale" || closures[1].Subject != "giardini comunali" {
					t.Fatalf("list scopes lost: %+v", closures)
				}
				for _, m := range closures {
					if m.ValidFrom == nil || *m.ValidFrom != "a partire dalle ore 17.00 del giorno 12 ottobre 2026" || m.ValidUntil == nil || *m.ValidUntil != "fino al perdurare dell’emergenza" || m.Place != nil {
						t.Fatalf("temporal leakage or invented place: %+v", m)
					}
					for _, e := range m.Evidence {
						if e.Page == nil || *e.Page != 2 || !strings.Contains(w.Text, e.Quote) {
							t.Fatal("evidence lost")
						}
					}
				}
				if again := supplementDirectiveClosures(got, w); len(again) != len(got) {
					t.Fatal("duplicate directive closures")
				}
			}
		})
	}
}

func TestTemporalCandidateRequiresItsOwnMeasurePredicate(t *testing.T) {
	text := "Chiusura dei giardini dal 12 ottobre 2026. Divieto di attività all'aperto nei giardini."
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["Chiusura dei giardini dal 12 ottobre 2026","Divieto di attività all'aperto nei giardini"],"measures":[{"kind":"prohibition","subject":"attività all'aperto","place":null,"temporal_candidates":[{"field":"valid_from","original_expression":"12 ottobre 2026","evidence_refs":[0,1]}],"evidence_refs":{"kind":[1],"subject":[1],"place":[]}}]}`
	w := Window{Ordinal: 1, Text: text, ResourceURL: "https://example.test/notice", CoreEndByte: len(text)}
	got, err := ParseWindow(raw, w)
	if err != nil || len(got) != 1 || got[0].ValidFrom != nil || len(got[0].TemporalCandidates) != 0 {
		t.Fatalf("closure date transferred to prohibition: %+v %v", got, err)
	}
	w.legacyLiteral = true
	old, err := ParseWindow(raw, w)
	if err != nil || len(old) != 1 || old[0].ValidFrom == nil {
		t.Fatalf("legacy result rewritten: %+v %v", old, err)
	}
}
