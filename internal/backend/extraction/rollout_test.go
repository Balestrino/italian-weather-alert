package extraction

import "testing"

func TestOutputFixCanaryPreservesLegacyWindowValidation(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		response := `{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["dispone la chiusura","sottopasso di via Maremmana","fino al perdurare dell’emergenza"],"measures":[{"kind":"closure","subject":"sottopasso di via Maremmana","place":null,"valid_from":null,"valid_until":"fino al perdurare dell'emergenza","evidence_refs":{"kind":[0],"subject":[1],"place":[],"valid_from":[],"valid_until":[2]}}]}`
		runner, results, _, process := fixtureRunner(t, true, response)
		runner.LegacyConfigurationVersion = "legacy"
		runner.OutputFixSources = map[string]bool{"other": true}
		if enabled {
			runner.OutputFixSources["calcinaia"] = true
		}
		if _, err := runJob(t, runner); err != nil {
			t.Fatal(err)
		}
		if results.value == nil {
			t.Fatal("missing result")
		}
		if enabled {
			if results.value.Status != "extracted" || process.lastConfiguration != "extraction-config" {
				t.Fatalf("selected source: %#v", results.value)
			}
		} else {
			if results.value.Status != "uninterpreted" || process.lastConfiguration != "legacy" {
				t.Fatal("unselected source behavior changed")
			}
		}
		if runner.legacyOutput || runner.ConfigurationVersion != "extraction-config" {
			t.Fatal("shared runner mutated")
		}
	}
}
