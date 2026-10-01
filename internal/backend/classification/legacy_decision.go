package classification

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// ParseLegacyDecision is retained solely for controlled rollout and rollback.
func ParseLegacyDecision(raw string, content fullContent) (Decision, error) {
	var decision Decision
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&decision) != nil || !errors.Is(decoder.Decode(new(any)), io.EOF) {
		return Decision{}, ErrInvalid
	}
	validReason := decision.ReasonCode == "regional_warning" || decision.ReasonCode == "local_weather_measure" || decision.ReasonCode == "weather_operational_update" || decision.ReasonCode == "not_relevant"
	quote := normalize(decision.EvidenceQuote)
	if !validReason || quote == "" || decision.Relevant != (decision.ReasonCode != "not_relevant") || !strings.Contains(strings.ToLower(content.Text), strings.ToLower(quote)) {
		return Decision{}, ErrInvalid
	}
	decision.EvidenceQuote = quote
	if !decision.Relevant {
		if operative, ok := explicitOperativeEvidence(content.Text); ok {
			decision = Decision{Relevant: true, ReasonCode: "local_weather_measure", EvidenceQuote: operative}
		}
	}
	return decision, nil
}
