package classification

import (
	"encoding/json"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/inference"
)

func InvalidOutputReason(raw, finish string, content Content) string {
	prefix := "classification_output_"
	if reason := inference.OutputEnvelopeReason(raw, finish); reason != "" {
		return prefix + reason
	}
	var object map[string]json.RawMessage
	_ = json.Unmarshal([]byte(raw), &object)
	if len(object) != 3 || object["relevant"] == nil || object["reason_code"] == nil || object["evidence_quote"] == nil {
		return prefix + "schema"
	}
	var decision Decision
	if json.Unmarshal([]byte(raw), &decision) != nil {
		return prefix + "schema"
	}
	if decision.ReasonCode != "regional_warning" && decision.ReasonCode != "local_weather_measure" && decision.ReasonCode != "weather_operational_update" && decision.ReasonCode != "not_relevant" || decision.Relevant != (decision.ReasonCode != "not_relevant") {
		return prefix + "reason"
	}
	quote := normalize(decision.EvidenceQuote)
	if quote == "" {
		return prefix + "quotation"
	}
	if !strings.Contains(strings.ToLower(content.Text), strings.ToLower(quote)) {
		if normalized := NormalizeOCR(quote).Text; normalized != "" && strings.Contains(strings.ToLower(content.Text), strings.ToLower(normalized)) {
			return prefix + "normalization"
		}
		return prefix + "quotation"
	}
	return prefix + "schema"
}
