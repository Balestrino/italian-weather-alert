package classification

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
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
		EvidenceOptions   []string       `json:"evidence_options"`
	}{segment.DocumentVersionID, segmentMessage{segment.Ordinal, segment.Total, segment.ResourceURL, segment.Role, segment.Page, segment.StartByte, segment.EndByte, segment.Text, segment.JSONScalars}, evidenceOptions(segment.Text)})
	if err != nil || len(segment.Text) > MaxSegmentTextBytes {
		return inference.Request{}, ErrInvalid
	}
	input, _ := json.Marshal(string(body))
	system, _ := json.Marshal(PromptBody)
	format := evidenceResponseFormat(evidenceOptions(segment.Text))
	return inference.Request{Model: model, Messages: []inference.Message{{Role: "system", Content: system}, {Role: "user", Content: input}}, ResponseFormat: format, MaxCompletionTokens: 4096}, nil
}

// Offer short, mechanically copied passages. This constrains generation without
// changing the independent literal-evidence validator or deciding relevance.
func evidenceOptions(text string) []string {
	var out []string
	for start := 0; start < len(text); {
		end := min(start+120, len(text))
		for end < len(text) && !utf8.RuneStart(text[end]) {
			end--
		}
		if end < len(text) {
			if boundary := strings.LastIndexAny(text[start:end], " \n"); boundary > 60 {
				end = start + boundary
			}
		}
		piece := strings.TrimSpace(text[start:end])
		if piece != "" {
			out = append(out, piece)
		}
		start = end
	}
	return out
}

func evidenceResponseFormat(options []string) json.RawMessage {
	var format map[string]any
	_ = json.Unmarshal(responseFormat, &format)
	properties := format["json_schema"].(map[string]any)["schema"].(map[string]any)["properties"].(map[string]any)
	properties["evidence_quote"].(map[string]any)["enum"] = options
	encoded, _ := json.Marshal(format)
	return encoded
}

type segmentMessage struct {
	Ordinal     int          `json:"ordinal"`
	Total       int          `json:"total"`
	ResourceURL string       `json:"resource_url"`
	Role        string       `json:"role"`
	Page        int          `json:"page,omitempty"`
	StartByte   int          `json:"start_byte"`
	EndByte     int          `json:"end_byte"`
	Text        string       `json:"text"`
	JSONScalars []JSONScalar `json:"json_scalars,omitempty"`
}

func LocalSegmentRequest(model string, segment Segment) (inference.Request, error) {
	request, err := SegmentRequest(model, segment)
	if err != nil {
		return inference.Request{}, err
	}
	request.Messages[0].Content, _ = json.Marshal(LocalPromptBody)
	return inference.LocalChatRequest(request), nil
}

func legacySegmentRequest(model string, segment Segment) (inference.Request, error) {
	request, err := SegmentRequest(model, segment)
	if err != nil {
		return inference.Request{}, err
	}
	body, err := json.Marshal(struct {
		DocumentVersionID int64          `json:"document_version_id"`
		Segment           segmentMessage `json:"segment"`
	}{segment.DocumentVersionID, segmentMessage{segment.Ordinal, segment.Total, segment.ResourceURL, segment.Role, segment.Page, segment.StartByte, segment.EndByte, segment.Text, nil}})
	if err != nil {
		return inference.Request{}, err
	}
	request.Messages[0].Content, _ = json.Marshal(LegacyPromptBody)
	request.Messages[1].Content, _ = json.Marshal(string(body))
	request.ResponseFormat = responseFormat
	return request, nil
}

func ParseDecision(raw string, content fullContent) (Decision, error) {
	return parseDecision(raw, content, CanonicalLiteral)
}

func ParseLocalDecision(raw string, content Content) (Decision, error) {
	return parseDecision(raw, content, CanonicalLocalLiteral)
}

func parseDecision(raw string, content Content, literalMatch func(string, string) (string, bool)) (Decision, error) {
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
	quote, literal := literalMatch(content.Text, decision.EvidenceQuote)
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
