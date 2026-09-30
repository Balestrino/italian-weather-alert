// Package inference defines provider-neutral inference calls and a bounded
// durable retry policy. Provider adapters perform exactly one HTTP attempt;
// retry history and increasing waits belong to the PostgreSQL job queue.
package inference

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/jobs"
)

var ErrInvalid = errors.New("invalid inference request")

const (
	Queue               = "inference"
	DefaultMaxAttempts  = 3
	DefaultRetryBase    = 2 * time.Second
	maximumContentBytes = 8 << 20
)

type Message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type Request struct {
	Model               string               `json:"model"`
	Messages            []Message            `json:"messages"`
	ResponseFormat      json.RawMessage      `json:"response_format,omitempty"`
	MaxCompletionTokens int                  `json:"max_completion_tokens,omitempty"`
	MaxTokens           int                  `json:"max_tokens,omitempty"`
	SkipSpecialTokens   *bool                `json:"skip_special_tokens,omitempty"`
	ChatTemplateKwargs  *ChatTemplateOptions `json:"chat_template_kwargs,omitempty"`
}

// ChatTemplateOptions is explicit and optional: omission preserves the provider's
// default. A false value must survive encoding for Qwen's non-thinking mode.
type ChatTemplateOptions struct {
	EnableThinking bool `json:"enable_thinking"`
}

type Usage struct {
	InputTokens, OutputTokens, CacheReadTokens *int64
}

type Response struct {
	CallID                               int64
	ID, Model, Content, ReasoningContent string
	FinishReason                         string
	Usage                                Usage
	Diagnostic                           Diagnostic
}

type Adapter interface {
	Name() string
	Complete(context.Context, Request) (Response, error)
}

// CallError intentionally excludes response bodies, request contents and
// credentials. Code is stable and safe to persist in job attempt history.
type CallError struct {
	Code       string
	StatusCode int
	Temporary  bool
	RetryAfter time.Duration
	Diagnostic Diagnostic
	Usage      Usage
}

func (e *CallError) Error() string { return e.Code }

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
}

func (p RetryPolicy) Validate() error {
	if p.MaxAttempts < 1 || p.MaxAttempts > DefaultMaxAttempts || p.BaseDelay < time.Millisecond || p.BaseDelay > time.Hour {
		return ErrInvalid
	}
	return nil
}

// JobRequest centralizes the inference-only three-attempt ceiling. The jobs
// store applies BaseDelay, 2*BaseDelay, ... after temporary failures.
func (p RetryPolicy) JobRequest(kind, idempotencyKey string, payload json.RawMessage, availableAt time.Time) (jobs.EnqueueRequest, error) {
	if err := p.Validate(); err != nil || strings.TrimSpace(kind) == "" || strings.TrimSpace(idempotencyKey) == "" || availableAt.IsZero() {
		return jobs.EnqueueRequest{}, ErrInvalid
	}
	return jobs.EnqueueRequest{
		Queue:          Queue,
		Kind:           kind,
		IdempotencyKey: idempotencyKey,
		Payload:        payload,
		MaxAttempts:    p.MaxAttempts,
		RetryBase:      p.BaseDelay,
		AvailableAt:    availableAt.UTC(),
	}, nil
}

func ValidateRequest(req Request) error {
	if strings.TrimSpace(req.Model) == "" || len(req.Messages) == 0 || req.MaxCompletionTokens < 0 || req.MaxTokens < 0 {
		return ErrInvalid
	}
	for _, message := range req.Messages {
		if (message.Role != "system" && message.Role != "user" && message.Role != "assistant") || len(message.Content) == 0 || len(message.Content) > maximumContentBytes || !validMessageContent(message.Content) {
			return ErrInvalid
		}
	}
	if len(req.ResponseFormat) > 0 && (len(req.ResponseFormat) > maximumContentBytes || !json.Valid(req.ResponseFormat)) {
		return ErrInvalid
	}
	return nil
}

func validMessageContent(raw json.RawMessage) bool {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	switch content := value.(type) {
	case string:
		return true
	case []any:
		return len(content) > 0
	default:
		return false
	}
}
