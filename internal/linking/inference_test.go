package linking

import (
	"errors"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/extraction"
)

func fixtureContexts() (MeasureContext, []Candidate) {
	current := MeasureContext{RunID: 30, Ordinal: 1, Municipality: "050004", Kind: "reopening", Subject: "sottopasso di via Maremmana", Place: "via Maremmana", Evidence: []extraction.Evidence{{Field: "kind", ResourceURL: "https://example/new", Quote: "è stato riaperto"}, {Field: "place", ResourceURL: "https://example/new", Quote: "via Maremmana"}}}
	candidates := []Candidate{{MeasureContext: MeasureContext{RunID: 10, Ordinal: 1, Municipality: "050004", Kind: "closure", Subject: "sottopasso di via Maremmana", Place: "via Maremmana", Evidence: []extraction.Evidence{{Field: "kind", ResourceURL: "https://example/old", Quote: "resta chiuso"}, {Field: "place", ResourceURL: "https://example/old", Quote: "via Maremmana"}}}, ExactPlace: true, TextRank: 0.7}}
	return current, candidates
}

func TestProviderSchemaKeepsLocalEvidenceUniqueness(t *testing.T) {
	current, candidates := fixtureContexts()
	request, err := Request("qwen3.8-27b", current, candidates)
	if err != nil || strings.Contains(string(request.ResponseFormat), "uniqueItems") {
		t.Fatal("unsupported provider grammar")
	}
	if _, err := Parse(`{"status":"linked","reason_code":"partial_reopening","relation":"reopens","candidate_index":1,"current_evidence_indices":[1,1],"candidate_evidence_indices":[1]}`, current, candidates); !errors.Is(err, ErrInvalid) {
		t.Fatal("duplicate evidence accepted")
	}
}

func TestParseLinksOnlySupportedPartialReopening(t *testing.T) {
	current, candidates := fixtureContexts()
	result, err := Parse(`{"status":"linked","reason_code":"partial_reopening","relation":"reopens","candidate_index":1,"current_evidence_indices":[1,2],"candidate_evidence_indices":[1,2]}`, current, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if result.CandidateRunID == nil || *result.CandidateRunID != 10 || result.Relation == nil || *result.Relation != "reopens" || len(result.CurrentEvidence) != 2 || len(result.CandidateEvidence) != 2 {
		t.Fatalf("supported partial reopening link lost: %#v", result)
	}
}

func TestParseLeavesAmbiguousRelationUnresolved(t *testing.T) {
	current, candidates := fixtureContexts()
	result, err := Parse(`{"status":"unresolved","reason_code":"ambiguous_relation","relation":null,"candidate_index":null,"current_evidence_indices":[1],"candidate_evidence_indices":[]}`, current, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if result.CandidateRunID != nil || result.Relation != nil || result.Status != "unresolved" {
		t.Fatalf("ambiguous link was asserted: %#v", result)
	}
	if _, err = Parse(`{"status":"linked","reason_code":"cross_document_evidence","relation":"reopens","candidate_index":2,"current_evidence_indices":[1],"candidate_evidence_indices":[1]}`, current, candidates); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown candidate accepted: %v", err)
	}
}

func TestParseOverridesUnsupportedChoiceWhenPlaceIsAmbiguous(t *testing.T) {
	current, candidates := fixtureContexts()
	current.Place = ""
	current.Subject = "sottopasso"
	candidates[0].Subject = "sottopasso"
	second := candidates[0]
	second.RunID = 11
	second.Place = "via della Botte"
	candidates = append(candidates, second)
	result, err := Parse(`{"status":"linked","reason_code":"cross_document_evidence","relation":"reopens","candidate_index":1,"current_evidence_indices":[1],"candidate_evidence_indices":[1]}`, current, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "unresolved" || result.ReasonCode != "ambiguous_relation" || result.CandidateRunID != nil || result.Relation != nil {
		t.Fatalf("unsupported ambiguous choice survived local validation: %#v", result)
	}
}

func TestParseKeepsUnresolvedDecisionWhenOptionalEvidenceIndexIsInvalid(t *testing.T) {
	current, candidates := fixtureContexts()
	result, err := Parse(`{"status":"unresolved","reason_code":"ambiguous_relation","relation":null,"candidate_index":null,"current_evidence_indices":[99],"candidate_evidence_indices":[]}`, current, candidates)
	if err != nil || result.Status != "unresolved" || result.ReasonCode != "ambiguous_relation" || len(result.CurrentEvidence) != 0 {
		t.Fatalf("safe unresolved result was discarded for an optional bad locator: %#v %v", result, err)
	}
}
