package evaluation

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validComparison() ProcessingComparison {
	at := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	result := ProcessingComparison{ID: "semantic-v1", CorpusSHA256: strings.Repeat("a", 64), Reviewer: "reviewer", ReviewedAt: "2026-09-17", StartedAt: at, FinishedAt: at.Add(time.Minute)}
	for index, id := range []string{"partial-reopen", "ambiguous-link"} {
		result.Cases = append(result.Cases, ComparisonCase{ID: id, Kind: "simulation", Expected: "reviewed relation", Evidence: "retained/case", Baseline: Outcome{Status: "pass", Actual: "baseline result", Evidence: "run/baseline", RunIDs: []int64{int64(index*2 + 1)}}, Candidate: Outcome{Status: "pass", Actual: "candidate result", Evidence: "run/candidate", RunIDs: []int64{int64(index*2 + 2)}}})
	}
	return result
}

func TestProcessingComparisonRequiresPairedReviewedCases(t *testing.T) {
	value := validComparison()
	if err := value.Validate(); err != nil {
		t.Fatal(err)
	}
	value.Cases[1].Candidate.RunIDs = value.Cases[0].Candidate.RunIDs
	// A run may intentionally cover more than one reviewed assertion; pairing is
	// by case, while duplicate identities inside one outcome are rejected.
	if err := value.Validate(); err != nil {
		t.Fatal(err)
	}
	value.Cases[0].Candidate.RunIDs = []int64{2, 2}
	if !errors.Is(value.Validate(), ErrComparisonInvalid) {
		t.Fatal("duplicate run evidence accepted")
	}
}

func TestProcessVersionPinsCatalogIdentities(t *testing.T) {
	model, prompt := "model-v1", "prompt-v1"
	base := []ConfigurationIdentity{{ID: "link", Stage: "linking", Revision: "v1", LogicVersion: "logic", ContentSHA256: strings.Repeat("b", 64), ModelVersionID: &model, PromptVersionID: &prompt}}
	a := processVersion(base)
	base[0].PromptVersionID = nil
	if a == processVersion(base) {
		t.Fatal("prompt identity did not affect process version")
	}
}
