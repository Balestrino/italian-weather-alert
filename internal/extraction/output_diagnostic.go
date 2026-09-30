package extraction

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
)

func InvalidOutputReason(raw, finish string, window Window, parseErr error) string {
	prefix := "extraction_output_"
	if reason := inference.OutputEnvelopeReason(raw, finish); reason != "" {
		return prefix + reason
	}
	var header struct {
		Envelope string   `json:"envelope_version"`
		Ordinal  int      `json:"window_ordinal"`
		Evidence []string `json:"evidence"`
		Measures []struct {
			Refs map[string][]int `json:"evidence_refs"`
		} `json:"measures"`
	}
	if json.Unmarshal([]byte(raw), &header) != nil {
		return prefix + "schema"
	}
	if header.Envelope != CompactEnvelopeVersion && header.Envelope != CompactEnvelopeVersionV2 {
		return prefix + "envelope_version"
	}
	if header.Ordinal != window.Ordinal {
		return prefix + "window_reference"
	}
	for _, measure := range header.Measures {
		for _, indices := range measure.Refs {
			for _, index := range indices {
				if index < 0 || index >= len(header.Evidence) {
					return prefix + "evidence_reference"
				}
			}
		}
	}
	for _, quote := range header.Evidence {
		normalized := classification.Normalize(quote)
		if !strings.Contains(strings.ToLower(window.Text), strings.ToLower(normalized)) {
			if canonical := classification.NormalizeOCR(normalized).Text; canonical != "" && literalReference(window.Content(), window.ResourceURL, windowPage(window), canonical) {
				return prefix + "normalization"
			}
			return prefix + "quotation"
		}
	}
	if errors.Is(parseErr, ErrEvidence) {
		return prefix + "evidence_reference"
	}
	if strings.TrimSpace(raw) == "" {
		return prefix + "empty"
	}
	return prefix + "schema"
}
