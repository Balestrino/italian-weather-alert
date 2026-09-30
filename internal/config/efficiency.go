package config

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

// EfficiencyRollout uses explicit source lists. Empty lists keep each optimization off.
type EfficiencyRollout struct {
	PreflightSources, OCRSources, InterpretationSources, CheckpointSources, OutputFixSources map[string]bool
}

var rolloutSource = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,199}$`)

func LoadEfficiencyRollout() (EfficiencyRollout, error) {
	result := EfficiencyRollout{}
	for name, target := range map[string]*map[string]bool{
		"IWA_PREFLIGHT_SOURCES":            &result.PreflightSources,
		"IWA_OCR_REUSE_SOURCES":            &result.OCRSources,
		"IWA_INTERPRETATION_REUSE_SOURCES": &result.InterpretationSources,
		"IWA_SEGMENT_CHECKPOINT_SOURCES":   &result.CheckpointSources,
		"IWA_OUTPUT_FIX_SOURCES":           &result.OutputFixSources,
	} {
		*target = map[string]bool{}
		value := strings.TrimSpace(os.Getenv(name))
		if value == "" {
			continue
		}
		for _, source := range strings.Split(value, ",") {
			source = strings.TrimSpace(source)
			if !rolloutSource.MatchString(source) || (*target)[source] {
				return EfficiencyRollout{}, errors.New("invalid efficiency source list")
			}
			(*target)[source] = true
		}
	}
	return result, nil
}
