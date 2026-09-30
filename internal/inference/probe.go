package inference

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ProbeResult struct {
	Adapter, RequestedModel, ReturnedModel string
	ObservedAt                             time.Time
	Relevant                               bool
	InputTokens, OutputTokens              *int64
	CacheReadTokens                        *int64
}

// ProbeStructuredOutput checks the authenticated chat contract with bounded
// output. Its result contains no prompt, response text, endpoint or credential.
func ProbeStructuredOutput(ctx context.Context, adapter Adapter, model string, observedAt time.Time) (ProbeResult, error) {
	if adapter == nil || strings.TrimSpace(model) == "" || observedAt.IsZero() {
		return ProbeResult{}, ErrInvalid
	}
	format := json.RawMessage(`{"type":"json_schema","json_schema":{"name":"relevance_probe","strict":true,"schema":{"type":"object","properties":{"relevant":{"type":"boolean"}},"required":["relevant"],"additionalProperties":false}}}`)
	content, _ := json.Marshal("This is a synthetic weather closure notice. Return the requested JSON only.")
	// Qwen3.8 emits reasoning before the constrained final content. Small output
	// limits can therefore finish with valid reasoning but no JSON answer.
	response, err := adapter.Complete(ctx, Request{Model: model, Messages: []Message{{Role: "user", Content: content}}, ResponseFormat: format, MaxCompletionTokens: 1024})
	if err != nil {
		return ProbeResult{}, err
	}
	var output struct {
		Relevant *bool `json:"relevant"`
	}
	if json.Unmarshal([]byte(response.Content), &output) != nil || output.Relevant == nil {
		return ProbeResult{}, &CallError{Code: "structured_output_invalid"}
	}
	return ProbeResult{
		Adapter: adapter.Name(), RequestedModel: model, ReturnedModel: response.Model,
		ObservedAt: observedAt.UTC(), Relevant: *output.Relevant,
		InputTokens: response.Usage.InputTokens, OutputTokens: response.Usage.OutputTokens,
		CacheReadTokens: response.Usage.CacheReadTokens,
	}, nil
}

type OCRCheck struct {
	Name, Text string
}

type OCRProbeResult struct {
	Adapter, RequestedModel, ReturnedModel string
	ObservedAt                             time.Time
	InputSHA256, OutputSHA256              string
	OutputBytes                            int
	FinishReason                           string
	MatchedChecks, MissingChecks           []string
	InputTokens, OutputTokens              *int64
	CacheReadTokens                        *int64
}

// ProbeOCR verifies the documented Regolo image-message contract without
// returning provider text, prompts, endpoints or credentials. The caller
// supplies non-sensitive expected fragments for a reviewed fixture.
func ProbeOCR(ctx context.Context, adapter Adapter, model, mediaType string, image []byte, checks []OCRCheck, observedAt time.Time) (OCRProbeResult, error) {
	if adapter == nil || strings.TrimSpace(model) == "" || mediaType != "image/png" || len(image) == 0 || len(image) > maximumContentBytes || len(checks) == 0 || observedAt.IsZero() {
		return OCRProbeResult{}, ErrInvalid
	}
	for _, check := range checks {
		if strings.TrimSpace(check.Name) == "" || strings.TrimSpace(check.Text) == "" {
			return OCRProbeResult{}, ErrInvalid
		}
	}
	content, err := json.Marshal([]any{
		map[string]any{"type": "text", "text": "Free OCR."},
		map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(image), "format": mediaType}},
	})
	if err != nil {
		return OCRProbeResult{}, ErrInvalid
	}
	keepSpecialTokens := false
	response, err := adapter.Complete(ctx, Request{
		Model: model, Messages: []Message{{Role: "user", Content: content}},
		MaxTokens: 4096, SkipSpecialTokens: &keepSpecialTokens,
	})
	if err != nil {
		return OCRProbeResult{}, err
	}
	if strings.TrimSpace(response.Content) == "" {
		return OCRProbeResult{}, &CallError{Code: "ocr_output_empty"}
	}
	inputHash := sha256.Sum256(image)
	outputHash := sha256.Sum256([]byte(response.Content))
	result := OCRProbeResult{
		Adapter: adapter.Name(), RequestedModel: model, ReturnedModel: response.Model,
		ObservedAt: observedAt.UTC(), InputSHA256: fmt.Sprintf("%x", inputHash),
		OutputSHA256: fmt.Sprintf("%x", outputHash), OutputBytes: len(response.Content), FinishReason: response.FinishReason,
		InputTokens: response.Usage.InputTokens, OutputTokens: response.Usage.OutputTokens,
		CacheReadTokens: response.Usage.CacheReadTokens,
	}
	for _, check := range checks {
		if strings.Contains(strings.ToLower(response.Content), strings.ToLower(check.Text)) {
			result.MatchedChecks = append(result.MatchedChecks, check.Name)
		} else {
			result.MissingChecks = append(result.MissingChecks, check.Name)
		}
	}
	return result, nil
}
