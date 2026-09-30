package classification

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/inference"
)

var responseFormat = json.RawMessage(`{"type":"json_schema","json_schema":{"name":"relevance_classification","strict":true,"schema":{"type":"object","properties":{"relevant":{"type":"boolean"},"reason_code":{"type":"string","enum":["regional_warning","local_weather_measure","weather_operational_update","not_relevant"]},"evidence_quote":{"type":"string"}},"required":["relevant","reason_code","evidence_quote"],"additionalProperties":false}}}`)

type Decision struct {
	Relevant      bool   `json:"relevant"`
	ReasonCode    string `json:"reason_code"`
	EvidenceQuote string `json:"evidence_quote"`
}

func Request(model string, versionID int64, content fullContent) (inference.Request, error) {
	if model == "" || !content.Complete || len(content.Sections) == 0 {
		return inference.Request{}, ErrInvalid
	}
	inputObject, err := contentMessage(versionID, content)
	if err != nil {
		return inference.Request{}, err
	}
	// OpenAI-compatible chat content is a string or a multimodal parts array.
	// Keep the structured document envelope inside a string instead of sending
	// it as an unsupported object-valued messages[].content.
	input, _ := json.Marshal(string(inputObject))
	system, _ := json.Marshal(PromptBody)
	return inference.Request{Model: model, Messages: []inference.Message{{Role: "system", Content: system}, {Role: "user", Content: input}}, ResponseFormat: responseFormat, MaxCompletionTokens: 4096}, nil
}

func SegmentRequest(model string, segment Segment) (inference.Request, error) {
	if model == "" || segment.DocumentVersionID < 1 || segment.Ordinal < 1 || segment.Total < segment.Ordinal || segment.ResourceURL == "" || segment.StartByte < 0 || segment.EndByte <= segment.StartByte || segment.Text == "" {
		return inference.Request{}, ErrInvalid
	}
	body, err := json.Marshal(struct {
		DocumentVersionID int64          `json:"document_version_id"`
		Segment           segmentMessage `json:"segment"`
	}{segment.DocumentVersionID, segmentMessage{segment.Ordinal, segment.Total, segment.ResourceURL, segment.Role, segment.Page, segment.StartByte, segment.EndByte, segment.Text}})
	if err != nil || len(segment.Text) > MaxSegmentTextBytes {
		return inference.Request{}, ErrInvalid
	}
	input, _ := json.Marshal(string(body))
	system, _ := json.Marshal(PromptBody)
	return inference.Request{Model: model, Messages: []inference.Message{{Role: "system", Content: system}, {Role: "user", Content: input}}, ResponseFormat: responseFormat, MaxCompletionTokens: 4096}, nil
}

type segmentMessage struct {
	Ordinal     int    `json:"ordinal"`
	Total       int    `json:"total"`
	ResourceURL string `json:"resource_url"`
	Role        string `json:"role"`
	Page        int    `json:"page,omitempty"`
	StartByte   int    `json:"start_byte"`
	EndByte     int    `json:"end_byte"`
	Text        string `json:"text"`
}

func ParseDecision(raw string, content fullContent) (Decision, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &fields) != nil || len(fields) != 3 || fields["relevant"] == nil || string(fields["relevant"]) == "null" || fields["reason_code"] == nil || fields["evidence_quote"] == nil {
		return Decision{}, ErrInvalid
	}
	var decision Decision
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&decision) != nil || !errors.Is(decoder.Decode(new(any)), io.EOF) {
		return Decision{}, ErrInvalid
	}
	validReason := decision.ReasonCode == "regional_warning" || decision.ReasonCode == "local_weather_measure" || decision.ReasonCode == "weather_operational_update" || decision.ReasonCode == "not_relevant"
	quote, literal := CanonicalLiteral(content.Text, decision.EvidenceQuote)
	if !validReason || quote == "" || decision.Relevant != (decision.ReasonCode != "not_relevant") || !literal {
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

func explicitOperativeEvidence(text string) (string, bool) {
	lower := strings.ToLower(text)
	for _, marker := range []string{
		"ordina in via contingibile e urgente",
		"ordina in via contingenti e urgente",
		"dispone la chiusura",
		"resta chiuso",
		"è stato riaperto",
		"e' stato riaperto",
		"coc) resta attivo",
	} {
		if index := strings.Index(lower, marker); index >= 0 && index+len(marker) <= len(text) {
			return text[index : index+len(marker)], true
		}
	}
	return "", false
}
