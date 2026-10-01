package config

import "testing"

func TestEfficiencyRolloutIsExplicitAndDefaultsOff(t *testing.T) {
	names := []string{"IWA_OCR_REUSE_SOURCES", "IWA_INTERPRETATION_REUSE_SOURCES", "IWA_SEGMENT_CHECKPOINT_SOURCES", "IWA_OUTPUT_FIX_SOURCES", "IWA_PREFLIGHT_SOURCES"}
	for _, name := range names {
		t.Setenv(name, "")
	}
	r, err := LoadEfficiencyRollout()
	if err != nil || r.OCRSources == nil || r.OutputFixSources == nil || len(r.PreflightSources)+len(r.OCRSources)+len(r.InterpretationSources)+len(r.CheckpointSources)+len(r.OutputFixSources) != 0 {
		t.Fatal("rollout not disabled")
	}
	t.Setenv(names[3], "source-a, source-b")
	r, err = LoadEfficiencyRollout()
	if err != nil || !r.OutputFixSources["source-a"] || r.OutputFixSources["source-c"] {
		t.Fatal("source scope lost")
	}
	for _, bad := range []string{"*", "source-a,source-a", "source-a,", "source/a"} {
		t.Setenv(names[3], bad)
		if _, err = LoadEfficiencyRollout(); err == nil {
			t.Fatalf("invalid list accepted %q", bad)
		}
	}
}
