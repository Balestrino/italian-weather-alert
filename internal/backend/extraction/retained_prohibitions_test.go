package extraction

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRetainedProhibitionListCompletesGroupedAndMissingPlaces(t *testing.T) {
	text := "Il ponte di via del Bosco è stato riaperto alla circolazione. Restano al momento in vigore, salvo successive modifiche, le disposizioni: chiusura al pubblico della biblioteca e dei giardini e divieto di attività nei parchi e giardini pubblici, nelle aree verdi, nelle ciclopiste in riva d’Arno e negli orti sociali. Il Centro Operativo Comunale resta attivo."
	quote := text[strings.Index(text, "Restano"):strings.Index(text, ". Il Centro")]
	for _, candidates := range []any{
		[]any{},
		[]any{map[string]any{"kind": "prohibition", "subject": "attività nei parchi e giardini pubblici", "place": nil, "evidence_refs": map[string]any{"kind": []int{0}, "subject": []int{0}, "place": []int{}}, "temporal_candidates": []any{}}},
	} {
		b, _ := json.Marshal(map[string]any{"envelope_version": "compact-evidence-v2", "window_ordinal": 1, "evidence": []string{quote}, "measures": candidates})
		w := Window{Ordinal: 1, Total: 1, Text: text, ResourceURL: "https://example.test/notice", CoreEndByte: len(text)}
		got, err := ParseWindow(string(b), w)
		if err != nil {
			t.Fatal(err)
		}
		places := map[string]bool{}
		kinds := map[string]int{}
		for _, m := range got {
			kinds[m.Kind]++
			if m.Kind != "prohibition" {
				continue
			}
			if m.Subject != "attività" || m.Place == nil || places[*m.Place] || m.ValidFrom == nil || *m.ValidFrom != "al momento" || m.ValidUntil == nil || *m.ValidUntil != "salvo successive modifiche" {
				t.Fatal("lost independent scope or relative validity", m)
			}
			places[*m.Place] = true
			for _, e := range m.Evidence {
				if !strings.Contains(text, e.Quote) || e.ResourceURL != w.ResourceURL || e.SegmentOrdinal != 1 {
					t.Fatal("literal field provenance lost", e)
				}
			}
		}
		if kinds["prohibition"] != 4 || kinds["closure"] != 2 || kinds["reopening"] != 1 || !places["parchi e giardini pubblici"] || !places["ciclopiste in riva d’Arno"] {
			t.Fatal("scope list or independent measures lost", got)
		}
		if again := supplementRetainedProhibitions(got, w); len(again) != len(got) {
			t.Fatal("duplicate list completion")
		}
		w.legacyLiteral = true
		old, err := ParseWindow(string(b), w)
		if err != nil || len(old) != len(candidates.([]any)) {
			t.Fatal("legacy response changed", old, err)
		}
	}
}

func TestRetainedProhibitionListRefusesUnsupportedScopes(t *testing.T) {
	positive := "Restano in vigore le disposizioni: divieto di attività nei parchi, nelle aree verdi e negli orti."
	for _, text := range []string{
		"Non " + positive,
		"Se piove, " + positive,
		"La relazione cita «" + positive + "»",
		"Oggetto: " + positive + " VISTO il bollettino.",
		strings.Replace(positive, "Restano", "Resteranno", 1),
		strings.Replace(positive, "negli orti", "negli orti, esclusi quelli privati", 1),
		strings.Replace(positive, "nelle aree verdi e negli orti", "piazze e orti", 1),
		strings.TrimSuffix(positive, "."),
	} {
		w := Window{Ordinal: 1, Text: text, ResourceURL: "https://example.test/notice", CoreEndByte: len(text)}
		if got := supplementRetainedProhibitions(nil, w); len(got) != 0 {
			t.Fatal("unsupported list completion", text, got)
		}
	}
	w := Window{Ordinal: 2, Text: positive, ResourceURL: "https://example.test/notice", CoreStartByte: len("Restano"), CoreEndByte: len(positive)}
	if got := supplementRetainedProhibitions(nil, w); len(got) != 0 {
		t.Fatal("context-only list completed", got)
	}
}

func TestRetainedProhibitionListPreservesExistingDistinctTimes(t *testing.T) {
	quote := "Restano in vigore fino al 10 ottobre le disposizioni: divieto di attività nei parchi, nelle aree verdi e negli orti"
	w := Window{Ordinal: 1, Text: quote + ".", ResourceURL: "https://example.test/notice", CoreEndByte: len(quote) + 1}
	prior := []Measure{{Ordinal: 1, Kind: "prohibition", Subject: "attività", Evidence: []Evidence{{Field: "kind", Quote: quote}}, TemporalCandidates: []TemporalCandidate{{Field: "valid_until", OriginalExpression: "10 ottobre"}}}}
	before, _ := json.Marshal(prior)
	got := supplementRetainedProhibitions(prior, w)
	after, _ := json.Marshal(got)
	if string(before) != string(after) {
		t.Fatal("existing temporal evidence overwritten", got)
	}
}

func TestCompletedReopeningDoesNotProjectRetrospectiveClosure(t *testing.T) {
	text := "Il sottopasso di via Sintetica, chiuso nella serata di ieri per la presenza di acqua, è stato riaperto alla circolazione."
	w := Window{Ordinal: 1, Text: text, ResourceURL: "https://example.test/notice", CoreEndByte: len(text)}
	prior := []Measure{{Ordinal: 1, Kind: "closure", Subject: "sottopasso di via Sintetica", Evidence: []Evidence{{Field: "kind", Quote: text}}, TemporalCandidates: []TemporalCandidate{{Field: "valid_until", OriginalExpression: "nella serata di ieri"}}}}
	got := supplementNamedReopenings(prior, w)
	if len(got) != 1 || got[0].Kind != "reopening" || got[0].ValidFrom != nil || got[0].ValidUntil != nil || len(got[0].TemporalCandidates) != 0 {
		t.Fatal("retrospective closure became restriction or dated reopening", got)
	}
	prior[0].Evidence = append(prior[0].Evidence, Evidence{Field: "kind", Quote: "Resta la chiusura al pubblico per il tratto pedonale"})
	if got = supplementNamedReopenings(prior, w); len(got) != 2 || got[0].Kind != "closure" {
		t.Fatal("independent closure erased", got)
	}
	w.CoreStartByte = len("Il sottopasso")
	if got = supplementNamedReopenings(prior[:1], w); len(got) != 1 || got[0].Kind != "closure" {
		t.Fatal("context-only reopening changed closure", got)
	}
}
