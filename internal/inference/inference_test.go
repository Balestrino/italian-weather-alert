package inference

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testRequest() Request {
	content, _ := json.Marshal("synthetic input")
	return Request{Model: "qwen-fixture", Messages: []Message{{Role: "user", Content: content}}, ResponseFormat: json.RawMessage(`{"type":"json_object"}`), MaxCompletionTokens: 32}
}

func TestOpenAIChatContractAndSecretHandling(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer private-adapter-fixture" {
			t.Error("missing authenticated POST")
		}
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "private-adapter-fixture") {
			t.Error("credential copied into body")
		}
		if json.Unmarshal(body, &received) != nil {
			t.Error("request is not JSON")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"call-1","model":"qwen-fixture","choices":[{"message":{"content":"{\"relevant\":true}","reasoning_content":"bounded"}}],"usage":{"prompt_tokens":11,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":3}}}`)
	}))
	defer server.Close()

	adapter, err := NewOpenAIChat("regolo-fixture", server.URL, "private-adapter-fixture", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Complete(context.Background(), testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if adapter.Name() != "regolo-fixture" || result.Model != "qwen-fixture" || result.Content != `{"relevant":true}` || result.Usage.InputTokens == nil || *result.Usage.InputTokens != 11 || result.Usage.CacheReadTokens == nil || *result.Usage.CacheReadTokens != 3 {
		t.Fatalf("unexpected response: %#v", result)
	}
	if received["stream"] != false || received["model"] != "qwen-fixture" {
		t.Fatalf("unexpected request: %#v", received)
	}
	if received["max_completion_tokens"] != float64(32) {
		t.Fatalf("completion limit omitted: %#v", received)
	}
}

func TestOpenAIChatExplicitThinkingModeAndFinishReason(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "provider_default", true: "disabled"}[disabled], func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				kwargs, present := body["chat_template_kwargs"]
				if present != disabled || disabled && string(kwargs) != `{"enable_thinking":false}` {
					t.Errorf("thinking option changed or omitted: %s", kwargs)
				}
				_, _ = io.WriteString(w, `{"model":"qwen-fixture","choices":[{"finish_reason":"length","message":{"content":"{}"}}]}`)
			}))
			defer server.Close()
			adapter, _ := NewOpenAIChat("fixture", server.URL, "private-token", server.Client())
			req := testRequest()
			if disabled {
				req.ChatTemplateKwargs = &ChatTemplateOptions{EnableThinking: false}
			}
			response, err := adapter.Complete(context.Background(), req)
			if err != nil || response.FinishReason != "length" {
				t.Fatalf("completion termination not retained: %#v, %v", response, err)
			}
		})
	}
}

func TestOpenAIChatClassifiesSafeTemporaryErrors(t *testing.T) {
	for _, tc := range []struct {
		status    int
		temporary bool
		code      string
	}{
		{http.StatusBadRequest, false, "provider_rejected"},
		{http.StatusUnauthorized, false, "provider_rejected"},
		{http.StatusTooManyRequests, true, "provider_temporary"},
		{http.StatusServiceUnavailable, true, "provider_temporary"},
	} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Retry-After", "9")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, `{"private":"response detail"}`)
			}))
			defer server.Close()
			adapter, _ := NewOpenAIChat("fixture", server.URL, "private-token", server.Client())
			_, err := adapter.Complete(context.Background(), testRequest())
			var call *CallError
			if !errors.As(err, &call) || call.Temporary != tc.temporary || call.Code != tc.code || strings.Contains(err.Error(), "private") {
				t.Fatalf("unsafe or incorrect error: %#v", err)
			}
			if tc.temporary && call.RetryAfter != 9*time.Second {
				t.Fatalf("Retry-After not retained: %s", call.RetryAfter)
			}
		})
	}
}

func TestOpenAIChatRejectsRedirectAndMalformedResponse(t *testing.T) {
	receivedRedirectAuth := ""
	target := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		receivedRedirectAuth = r.Header.Get("Authorization")
	}))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Redirect(w, &http.Request{}, target.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	adapter, _ := NewOpenAIChat("fixture", redirect.URL, "private-token", redirect.Client())
	_, err := adapter.Complete(context.Background(), testRequest())
	var call *CallError
	if !errors.As(err, &call) || receivedRedirectAuth != "" {
		t.Fatal("redirect followed or unsafe error returned")
	}

	malformed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"model":"fixture","choices":[]}`)
	}))
	defer malformed.Close()
	adapter, _ = NewOpenAIChat("fixture", malformed.URL, "private-token", malformed.Client())
	_, err = adapter.Complete(context.Background(), testRequest())
	if !errors.As(err, &call) || call.Code != "invalid_response" {
		t.Fatal("malformed provider response accepted")
	}
}

func TestOpenAIChatTreatsProviderDeadlineAsTemporary(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	adapter, _ := NewOpenAIChat("fixture", "https://provider.example/v1", "private-token", client)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, err := adapter.Complete(ctx, testRequest())
	var call *CallError
	if !errors.As(err, &call) || call.Code != "request_timeout" || !call.Temporary {
		t.Fatalf("deadline was not classified as temporary: %#v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestRetryPolicyCapsTotalAttemptsAndUsesDurableBackoff(t *testing.T) {
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	for _, attempts := range []int{1, 2, 3} {
		policy := RetryPolicy{MaxAttempts: attempts, BaseDelay: 2 * time.Second}
		req, err := policy.JobRequest("classification", "fixture-key", json.RawMessage(`{"run_id":1}`), now)
		if err != nil || req.MaxAttempts != attempts || req.RetryBase != 2*time.Second || req.Queue != Queue {
			t.Fatalf("valid policy rejected: %#v, %v", req, err)
		}
	}
	for _, attempts := range []int{0, 4, 100} {
		if _, err := (RetryPolicy{MaxAttempts: attempts, BaseDelay: time.Second}).JobRequest("classification", "fixture-key", json.RawMessage(`{"run_id":1}`), now); !errors.Is(err, ErrInvalid) {
			t.Fatalf("unbounded attempt count %d accepted", attempts)
		}
	}
}

func TestStructuredProbeIsProviderNeutralAndSanitized(t *testing.T) {
	in, out := int64(7), int64(2)
	observed := time.Date(2026, 9, 17, 12, 0, 0, 0, time.FixedZone("fixture", 3600))
	var request Request
	result, err := ProbeStructuredOutput(context.Background(), recordingAdapter{
		request:  &request,
		response: Response{Model: "local-qwen", Content: `{"relevant":false}`, Usage: Usage{InputTokens: &in, OutputTokens: &out}},
	}, "configured-qwen", observed)
	if err != nil {
		t.Fatal(err)
	}
	if result.Adapter != "recording-fixture" || result.RequestedModel != "configured-qwen" || result.ReturnedModel != "local-qwen" || result.Relevant || !result.ObservedAt.Equal(observed.UTC()) {
		t.Fatalf("unexpected probe result: %#v", result)
	}
	if request.MaxCompletionTokens != 1024 || len(request.ResponseFormat) == 0 {
		t.Fatalf("probe did not reserve bounded reasoning and structured-output space: %#v", request)
	}
}

func TestOCRProbeUsesImageContractAndReturnsOnlySanitizedEvidence(t *testing.T) {
	var request Request
	adapter := recordingAdapter{
		request:  &request,
		response: Response{Model: "deepseek-ocr-2", Content: "ORDINA dalle ore 18.00 del giorno 20 agosto 2026 fino al perdurare dell’emergenza", FinishReason: "stop", Usage: Usage{}},
	}
	result, err := ProbeOCR(context.Background(), adapter, "deepseek-ocr-2", "image/png", []byte("synthetic-png"), []OCRCheck{
		{Name: "heading", Text: "ORDINA"},
		{Name: "start", Text: "20 agosto 2026"},
		{Name: "missing", Text: "Fornacette"},
	}, time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if request.MaxTokens != 4096 || request.SkipSpecialTokens == nil || *request.SkipSpecialTokens || len(request.Messages) != 1 || !strings.Contains(string(request.Messages[0].Content), "data:image/png;base64") {
		t.Fatalf("unexpected OCR request: %#v", request)
	}
	if result.FinishReason != "stop" || result.ReturnedModel != "deepseek-ocr-2" || len(result.MatchedChecks) != 2 || len(result.MissingChecks) != 1 || result.OutputBytes == 0 || result.InputSHA256 == "" || result.OutputSHA256 == "" {
		t.Fatalf("unexpected sanitized result: %#v", result)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "perdurare") || strings.Contains(string(encoded), "synthetic-png") {
		t.Fatal("probe result leaked provider or input content")
	}
}

type recordingAdapter struct {
	request  *Request
	response Response
}

func (a recordingAdapter) Name() string { return "recording-fixture" }
func (a recordingAdapter) Complete(_ context.Context, request Request) (Response, error) {
	*a.request = request
	return a.response, nil
}

func TestRequestValidation(t *testing.T) {
	valid := testRequest()
	if err := ValidateRequest(valid); err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.Messages[0].Content = json.RawMessage(`not-json`)
	if !errors.Is(ValidateRequest(invalid), ErrInvalid) {
		t.Fatal("invalid content accepted")
	}
	invalid = valid
	invalid.Messages = []Message{{Role: "user", Content: json.RawMessage(`{"unsupported":"object content"}`)}}
	if !errors.Is(ValidateRequest(invalid), ErrInvalid) {
		t.Fatal("object-valued chat content accepted")
	}
}
