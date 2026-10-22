package publicquery

import (
	"slices"
	"testing"
	"time"
)

func TestCarriedForwardKeepsProvenanceAndOwnOutcome(t *testing.T) {
	at := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	earlierRun, run, configuration := "35", int64(35), "extraction-config-a"
	earlier := documentAssessment{dimension: Dimension{State: "partial", Limitations: []string{"validated extraction has municipal facts awaiting domain projection"}, Evidence: []Evidence{}}, interpretedAt: &at, runID: &earlierRun}
	failed := documentAssessment{dimension: Dimension{State: "failed", Limitations: []string{"extraction_evidence_invalid"}, Evidence: []Evidence{}}}
	got := carriedForward(failed, presentedRun{reason: "extraction_evidence_invalid"}, earlier, presentedRun{run: &run, configuration: &configuration}, 60, 4291, "https://comune.example/avviso")
	want := []string{
		"carried_forward_from_equivalent_version:4291:run:35:configuration:extraction-config-a",
		"own_interpretation:failed:extraction_evidence_invalid",
		"validated extraction has municipal facts awaiting domain projection",
	}
	if got.dimension.State != "partial" || !slices.Equal(got.dimension.Limitations, want) || got.runID != &earlierRun || got.interpretedAt != &at {
		t.Fatalf("unexpected carried-forward assessment: %#v", got)
	}
	if len(got.dimension.Evidence) != 1 || got.dimension.Evidence[0].DocumentID != "60" || got.dimension.Evidence[0].VersionID != "4291" || got.dimension.Evidence[0].SourceURL != "https://comune.example/avviso" || *got.dimension.Evidence[0].Locator != carriedForwardLocator {
		t.Fatalf("unexpected carried-forward evidence: %#v", got.dimension.Evidence)
	}
	if len(earlier.dimension.Limitations) != 1 || len(earlier.dimension.Evidence) != 0 {
		t.Fatal("earlier assessment was mutated")
	}
	pending := documentAssessment{dimension: Dimension{State: "not_processed", Limitations: []string{"no interpretation was available at the knowledge boundary"}}}
	got = carriedForward(pending, presentedRun{}, earlier, presentedRun{}, 60, 4291, "https://comune.example/avviso")
	if got.dimension.Limitations[0] != "carried_forward_from_equivalent_version:4291:run:none:configuration:unknown" || got.dimension.Limitations[1] != "own_interpretation:not_available" {
		t.Fatalf("unexpected pending outcome: %#v", got.dimension.Limitations)
	}
	got = carriedForward(failed, presentedRun{}, earlier, presentedRun{run: &run, configuration: &configuration}, 60, 4291, "https://comune.example/avviso")
	if got.dimension.Limitations[1] != "own_interpretation:failed:unknown" {
		t.Fatal("missing failure code not reported as unknown")
	}
}
