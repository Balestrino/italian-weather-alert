package linking

import "testing"

func TestSemanticRetrievalAddsCandidatesWithoutReplacingBaseline(t *testing.T) {
	baseline := []Candidate{{MeasureContext: MeasureContext{RunID: 1, Ordinal: 1}, RetrievalMethod: "baseline_text"}}
	score := float32(.91)
	semantic := []Candidate{{MeasureContext: MeasureContext{RunID: 1, Ordinal: 1}, RetrievalMethod: "semantic", SemanticScore: &score}, {MeasureContext: MeasureContext{RunID: 2, Ordinal: 1}, RetrievalMethod: "semantic", SemanticScore: &score}}
	disabled := mergeCandidates(baseline, nil, 20)
	enabled := mergeCandidates(baseline, semantic, 20)
	if len(disabled) != 1 || disabled[0].RetrievalMethod != "baseline_text" {
		t.Fatalf("baseline changed while semantic disabled: %#v", disabled)
	}
	if len(enabled) != 2 || enabled[0].RunID != 1 || enabled[1].RunID != 2 || enabled[1].RetrievalMethod != "semantic" {
		t.Fatalf("semantic candidate did not augment baseline: %#v", enabled)
	}
}
