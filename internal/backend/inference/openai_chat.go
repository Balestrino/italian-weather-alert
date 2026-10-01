package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseBytes = 8 << 20

// OpenAIChat is the currently documented Regolo chat transport. It remains an
// Adapter implementation rather than a domain dependency, so a local provider
// can replace it without changing classification/OCR behavior.
type OpenAIChat struct {
	name, endpoint, apiKey string
	client                 *http.Client
}

func NewOpenAIChat(name, endpoint, apiKey string, client *http.Client) (*OpenAIChat, error) {
	u, err := url.Parse(endpoint)
	if strings.TrimSpace(name) == "" || strings.TrimSpace(apiKey) == "" || err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, ErrInvalid
	}
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	} else {
		clone := *client
		client = &clone
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &OpenAIChat{name: name, endpoint: endpoint, apiKey: apiKey, client: client}, nil
}

func (a *OpenAIChat) Name() string { return a.name }

func (a *OpenAIChat) Complete(ctx context.Context, request Request) (Response, error) {
	if err := ValidateRequest(request); err != nil {
		return Response{}, err
	}
	body, err := json.Marshal(struct {
		Model               string               `json:"model"`
		Messages            []Message            `json:"messages"`
		ResponseFormat      json.RawMessage      `json:"response_format,omitempty"`
		MaxCompletionTokens int                  `json:"max_completion_tokens,omitempty"`
		MaxTokens           int                  `json:"max_tokens,omitempty"`
		SkipSpecialTokens   *bool                `json:"skip_special_tokens,omitempty"`
		ChatTemplateKwargs  *ChatTemplateOptions `json:"chat_template_kwargs,omitempty"`
		Temperature         *float64             `json:"temperature,omitempty"`
		Seed                *int64               `json:"seed,omitempty"`
		Stream              bool                 `json:"stream"`
	}{request.Model, request.Messages, request.ResponseFormat, request.MaxCompletionTokens, request.MaxTokens, request.SkipSpecialTokens, request.ChatTemplateKwargs, request.Temperature, request.Seed, false})
	if err != nil {
		return Response{}, ErrInvalid
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, &CallError{Code: "request_build_failed"}
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	response, err := a.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return Response{}, &CallError{Code: "request_cancelled", Temporary: false}
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return Response{}, &CallError{Code: "request_timeout", Temporary: true}
		}
		return Response{}, &CallError{Code: "transport_error", Temporary: true}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		failure := rejection(response, a.apiKey)
		return Response{Usage: failure.Usage, Diagnostic: failure.Diagnostic}, failure
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes+1))
	var wire struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content          string `json:"content"`
				ReasoningContent string `json:"reasoning_content"`
			} `json:"message"`
		} `json:"choices"`
		Usage *wireUsage `json:"usage"`
	}
	err = decoder.Decode(&wire)
	result := Response{Usage: wire.Usage.usage(), Diagnostic: responseDiagnostic(response, a.apiKey)}
	invalid := func() (Response, error) {
		return result, &CallError{Code: "invalid_response", StatusCode: response.StatusCode, Diagnostic: result.Diagnostic, Usage: result.Usage}
	}
	if err != nil || len(wire.Choices) != 1 || strings.TrimSpace(wire.Model) == "" {
		return invalid()
	}
	if err = decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return invalid()
	}
	result.ID, result.Model = safeIdentifier(wire.ID, a.apiKey), wire.Model
	result.Content, result.ReasoningContent = wire.Choices[0].Message.Content, wire.Choices[0].Message.ReasoningContent
	result.FinishReason = wire.Choices[0].FinishReason
	return result, nil
}

func statusError(response *http.Response) *CallError {
	temporary := response.StatusCode == http.StatusRequestTimeout || response.StatusCode == http.StatusTooEarly || response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
	code := "provider_rejected"
	if temporary {
		code = "provider_temporary"
	}
	return &CallError{Code: code, StatusCode: response.StatusCode, Temporary: temporary, RetryAfter: retryAfter(response.Header.Get("Retry-After"), time.Now())}
}

func retryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(value); err == nil && when.After(now) {
		return when.Sub(now)
	}
	return 0
}
