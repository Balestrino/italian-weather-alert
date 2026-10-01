package extraction

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
)

func stringPointer(value string) *string { return &value }
func intPointer(value int) *int          { return &value }

func TestParseRequiresFieldEvidenceAndOriginalPage(t *testing.T) {
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{
		{ResourceURL: "https://comune.example/notizia", Role: "original", Text: "Comune di Calcinaia. Avviso di allerta meteo con ordinanza allegata."},
		{ResourceURL: "https://comune.example/ordinanza.pdf", Role: "attachment", Page: 2, Text: "Ordina la chiusura dei cimiteri comunali dalle ore 18 del 20 agosto fino al perdurare dell'emergenza."},
	}}
	raw := `{"measures":[{"kind":"closure","subject":"cimiteri comunali","place":"Calcinaia","valid_from":"dalle ore 18 del 20 agosto","valid_until":null,"indeterminate_fields":["valid_until"],"evidence":[{"field":"kind","resource_url":"https://comune.example/ordinanza.pdf","page":2,"quote":"Ordina la chiusura"},{"field":"subject","resource_url":"https://comune.example/ordinanza.pdf","page":2,"quote":"cimiteri comunali"},{"field":"place","resource_url":"https://comune.example/notizia","page":null,"quote":"Comune di Calcinaia"},{"field":"valid_from","resource_url":"https://comune.example/ordinanza.pdf","page":2,"quote":"dalle ore 18 del 20 agosto"}]}]}`
	measures, err := Parse(raw, content)
	if err != nil {
		t.Fatal(err)
	}
	if len(measures) != 1 || measures[0].Ordinal != 1 || measures[0].Kind != "closure" || measures[0].ValidUntil != nil || len(measures[0].IndeterminateFields) != 1 || measures[0].Evidence[0].Page == nil || *measures[0].Evidence[0].Page != 2 {
		t.Fatalf("structured measure lost evidence or indeterminate state: %#v", measures)
	}
}

func TestProviderSchemaAndLocalDuplicateValidation(t *testing.T) {
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: "https://comune.example/a", Text: "Chiusura cimiteri."}}}
	request, err := Request("qwen3.8-27b", 1, content)
	if err != nil || strings.Contains(string(request.ResponseFormat), "uniqueItems") || request.MaxCompletionTokens < 4097 {
		t.Fatal("unsupported provider schema or insufficient reasoning/output budget")
	}
	raw := `{"measures":[{"kind":"closure","subject":"cimiteri","place":null,"valid_from":null,"valid_until":null,"indeterminate_fields":["place","valid_from","valid_until","place"],"evidence":[{"field":"kind","resource_url":"https://comune.example/a","page":null,"quote":"Chiusura"},{"field":"subject","resource_url":"https://comune.example/a","page":null,"quote":"cimiteri"}]}]}`
	if _, err := Parse(raw, content); !errors.Is(err, ErrInvalid) {
		t.Fatal("removing provider keyword weakened local duplicate validation")
	}
}

func TestParseRejectsUnsupportedFieldsAndWrongPages(t *testing.T) {
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: "https://comune.example/ordinanza.pdf", Page: 2, Text: "Ordina la chiusura dei cimiteri."}}}
	cases := []string{
		`{"measures":[{"kind":"closure","subject":"cimiteri","place":null,"valid_from":null,"valid_until":null,"indeterminate_fields":["place","valid_from","valid_until"],"evidence":[{"field":"kind","resource_url":"https://comune.example/ordinanza.pdf","page":1,"quote":"Ordina la chiusura"},{"field":"subject","resource_url":"https://comune.example/ordinanza.pdf","page":2,"quote":"cimiteri"}]}]}`,
		`{"measures":[{"kind":"closure","subject":"scuole","place":null,"valid_from":null,"valid_until":null,"indeterminate_fields":["place","valid_from","valid_until"],"evidence":[{"field":"kind","resource_url":"https://comune.example/ordinanza.pdf","page":2,"quote":"Ordina la chiusura"},{"field":"subject","resource_url":"https://comune.example/ordinanza.pdf","page":2,"quote":"scuole"}]}]}`,
	}
	for _, raw := range cases {
		if _, err := Parse(raw, content); !errors.Is(err, ErrEvidence) {
			t.Fatalf("unsupported assertion accepted: %s: %v", raw, err)
		}
	}
	missingIndeterminate := `{"measures":[{"kind":"closure","subject":"cimiteri","place":null,"valid_from":null,"valid_until":null,"indeterminate_fields":["place","valid_from"],"evidence":[{"field":"kind","resource_url":"https://comune.example/ordinanza.pdf","page":2,"quote":"Ordina la chiusura"},{"field":"subject","resource_url":"https://comune.example/ordinanza.pdf","page":2,"quote":"cimiteri"}]}]}`
	if _, err := Parse(missingIndeterminate, content); !errors.Is(err, ErrInvalid) {
		t.Fatalf("implicit unknown field accepted: %v", err)
	}
}

func TestMergeCollapsesDuplicateAndRejectsPlaceConflictAndUnsupportedEvidence(t *testing.T) {
	url := "https://comune.example/ordinanza"
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: url, Text: "Dispone la chiusura del ponte fino alle 18 e dalle ore 10."}}}
	base := Measure{Kind: "closure", Subject: "ponte", IndeterminateFields: []string{"place", "valid_from", "valid_until"}, Evidence: []Evidence{{Field: "kind", ResourceURL: url, Quote: "Dispone la chiusura", SegmentOrdinal: 1}, {Field: "subject", ResourceURL: url, Quote: "ponte", SegmentOrdinal: 1}}}
	duplicate := base
	duplicate.Evidence = []Evidence{{Field: "kind", ResourceURL: url, Quote: "Dispone la chiusura", SegmentOrdinal: 2}, {Field: "subject", ResourceURL: url, Quote: "ponte", SegmentOrdinal: 2}}
	merged, err := Merge([][]Measure{{base}, {duplicate}}, content, 2)
	if err != nil || len(merged) != 1 || len(merged[0].Evidence) != 4 {
		t.Fatalf("duplicate supported fact was not collapsed with its evidence: %#v %v", merged, err)
	}
	place := "ponte"
	located := base
	located.Place = &place
	located.Evidence = append(located.Evidence, Evidence{Field: "place", ResourceURL: url, Quote: place, SegmentOrdinal: 2})
	if _, err := Merge([][]Measure{{base}, {located}}, content, 2); !errors.Is(err, ErrMergeConflict) {
		t.Fatalf("conflicting measure identity accepted: %v", err)
	}
	unsupported := base
	unsupported.Evidence = append([]Evidence(nil), base.Evidence...)
	unsupported.Evidence[0].SegmentOrdinal = 3
	if _, err := Merge([][]Measure{{unsupported}}, content, 2); !errors.Is(err, ErrEvidence) {
		t.Fatalf("unsupported segment evidence accepted: %v", err)
	}
}

func TestMergePreservesCrossWindowCalcinaiaTemporalConflictAndUnrelatedMeasures(t *testing.T) {
	url := "https://comune.calcinaia.example/ordinanza"
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: url, Text: "ORDINA la chiusura del sottopasso dal 10 settembre. Il dispositivo ordina la chiusura del sottopasso dal 10 agosto. Restano chiusi i parchi comunali fino a revoca."}}}
	september, august, until := "10 settembre", "10 agosto", "fino a revoca"
	baseEvidence := func(segment int, predicate string) []Evidence {
		return []Evidence{{Field: "kind", ResourceURL: url, Quote: predicate, SegmentOrdinal: segment}, {Field: "subject", ResourceURL: url, Quote: "sottopasso", SegmentOrdinal: segment}}
	}
	septemberMeasure := Measure{Kind: "closure", Subject: "sottopasso", ValidFrom: &september, Evidence: append(baseEvidence(1, "ORDINA la chiusura"), Evidence{Field: "valid_from", ResourceURL: url, Quote: september, SegmentOrdinal: 1})}
	augustMeasure := Measure{Kind: "closure", Subject: "sottopasso", ValidFrom: &august, Evidence: append(baseEvidence(2, "ordina la chiusura"), Evidence{Field: "valid_from", ResourceURL: url, Quote: august, SegmentOrdinal: 2})}
	parks := Measure{Kind: "restriction", Subject: "parchi comunali", ValidUntil: &until, Evidence: []Evidence{{Field: "kind", ResourceURL: url, Quote: "Restano chiusi", SegmentOrdinal: 2}, {Field: "subject", ResourceURL: url, Quote: "parchi comunali", SegmentOrdinal: 2}, {Field: "valid_until", ResourceURL: url, Quote: until, SegmentOrdinal: 2}}}

	merged, err := Merge([][]Measure{{septemberMeasure}, {augustMeasure, parks}}, content, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged) != 2 || merged[0].ValidFrom != nil || !slices.Contains(merged[0].IndeterminateFields, "valid_from") || len(merged[0].TemporalCandidates) != 2 {
		t.Fatalf("cross-window temporal conflict was not projected narrowly: %#v", merged)
	}
	first, second := merged[0].TemporalCandidates[0], merged[0].TemporalCandidates[1]
	if first.OriginalExpression != september || second.OriginalExpression != august || first.ConflictIdentity == nil || second.ConflictIdentity == nil || *first.ConflictIdentity != *second.ConflictIdentity || len(first.Evidence) != 1 || len(second.Evidence) != 1 {
		t.Fatalf("candidate expressions, evidence, or shared conflict identity were lost: %#v", merged[0].TemporalCandidates)
	}
	if merged[1].Subject != "parchi comunali" || merged[1].ValidUntil == nil || *merged[1].ValidUntil != until || merged[1].TemporalCandidates[0].ConflictIdentity != nil {
		t.Fatalf("unrelated supported measure was changed by the conflict: %#v", merged[1])
	}
	reversed, err := Merge([][]Measure{{augustMeasure}, {septemberMeasure, parks}}, content, 2)
	if err != nil || reversed[0].TemporalCandidates[0].ConflictIdentity == nil || *reversed[0].TemporalCandidates[0].ConflictIdentity != *first.ConflictIdentity {
		t.Fatalf("conflict identity is not deterministic across window order: %#v %v", reversed, err)
	}
}

func TestMergeCombinesCompatibleFieldsAndKeepsUnknownsExplicit(t *testing.T) {
	url := "https://comune.example/ordinanza"
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: url, Text: "Dispone la chiusura del ponte dalle ore 10 fino alle 18."}}}
	from, until := "dalle ore 10", "fino alle 18"
	left := Measure{Kind: "closure", Subject: "ponte", ValidFrom: &from, IndeterminateFields: []string{"place", "valid_until"}, Evidence: []Evidence{{Field: "kind", ResourceURL: url, Quote: "Dispone la chiusura", SegmentOrdinal: 1}, {Field: "subject", ResourceURL: url, Quote: "ponte", SegmentOrdinal: 1}, {Field: "valid_from", ResourceURL: url, Quote: from, SegmentOrdinal: 1}}}
	right := Measure{Kind: "closure", Subject: "ponte", ValidUntil: &until, IndeterminateFields: []string{"place", "valid_from"}, Evidence: []Evidence{{Field: "kind", ResourceURL: url, Quote: "Dispone la chiusura", SegmentOrdinal: 2}, {Field: "subject", ResourceURL: url, Quote: "ponte", SegmentOrdinal: 2}, {Field: "valid_until", ResourceURL: url, Quote: until, SegmentOrdinal: 2}}}
	merged, err := Merge([][]Measure{{left}, {right}}, content, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged) != 1 || merged[0].ValidFrom == nil || *merged[0].ValidFrom != from || merged[0].ValidUntil == nil || *merged[0].ValidUntil != until || !slices.Equal(merged[0].IndeterminateFields, []string{"place"}) {
		t.Fatalf("compatible evidence merge lost facts or unknowns: %#v", merged)
	}
}
