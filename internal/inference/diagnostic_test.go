package inference

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRejectionDiagnosticsDoNotRetainProviderMessages(t *testing.T) {
	for _, fixture := range []struct {
		status         int
		code, category string
	}{
		{401, "invalid_api_key", "authentication"}, {429, "insufficient_quota", "quota"},
		{429, "rate_limit_exceeded", "rate_limit"}, {400, "secret-key", "request"},
		{503, "server_error", "availability"},
	} {
		t.Run(fixture.category, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Request-ID", "secret-key")
				w.Header().Set("Retry-After", "120")
				w.WriteHeader(fixture.status)
				fmt.Fprintf(w, `{"error":{"code":%q,"message":"secret-key confidential prompt"},"usage":{"prompt_tokens":7}}`, fixture.code)
			}))
			defer server.Close()
			adapter, _ := NewOpenAIChat("fixture", server.URL, "secret-key", server.Client())
			response, err := adapter.Complete(context.Background(), Request{Model: "fixture", Messages: []Message{{Role: "user", Content: json.RawMessage(`"hello"`)}}})
			var call *CallError
			if !errors.As(err, &call) || call.Diagnostic.Category != fixture.category || call.StatusCode != fixture.status || call.RetryAfter.Seconds() != 120 {
				t.Fatalf("incorrect diagnostic: %#v %v", call, err)
			}
			if response.Usage.InputTokens == nil || *response.Usage.InputTokens != 7 {
				t.Fatal("error usage lost")
			}
			if strings.Contains(fmt.Sprintf("%+v %+v", call, call.Diagnostic), "secret-key") || call.Diagnostic.RequestID != "" {
				t.Fatal("secret escaped")
			}
		})
	}
}

func TestMalformedChatEnvelopeKeepsReportedUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "request-123")
		fmt.Fprint(w, `{"model":"fixture","choices":[],"usage":{"prompt_tokens":11,"completion_tokens":3}}`)
	}))
	defer server.Close()
	adapter, _ := NewOpenAIChat("fixture", server.URL, "private", server.Client())
	response, err := adapter.Complete(context.Background(), Request{Model: "fixture", Messages: []Message{{Role: "user", Content: json.RawMessage(`"hello"`)}}})
	if err == nil || response.Usage.InputTokens == nil || *response.Usage.InputTokens != 11 || response.Diagnostic.RequestID != "request-123" {
		t.Fatalf("receipt lost: %#v %v", response, err)
	}
}

func TestMalformedEmbeddingKeepsReportedUsage(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"model":"fixture","data":[],"usage":{"prompt_tokens":19}}`)
	}))
	defer server.Close()
	adapter, _ := NewOpenAIEmbedder("fixture", server.URL, "private", server.Client())
	response, err := adapter.Embed(context.Background(), "fixture", []string{"text"}, 2)
	if err == nil || response.InputTokens == nil || *response.InputTokens != 19 || response.Diagnostic.HTTPStatus != 200 {
		t.Fatalf("embedding usage lost: %#v %v", response, err)
	}
}

func TestRegoloExpiredTrialDiagnosticKeepsOnlyStableType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		fmt.Fprint(w, `{"error":{"code":"402","type":"trial_expired","message":"private account details secret-key"}}`)
	}))
	defer server.Close()
	adapter, _ := NewOpenAIChat("fixture", server.URL, "secret-key", server.Client())
	_, err := adapter.Complete(context.Background(), Request{Model: "fixture", Messages: []Message{{Role: "user", Content: json.RawMessage(`"hello"`)}}})
	var call *CallError
	if !errors.As(err, &call) || call.Diagnostic.ProviderCode != "trial_expired" || call.Diagnostic.Category != "quota" || call.Temporary {
		t.Fatalf("expired trial diagnostic lost or made retryable: %+v", call)
	}
	if call.Usage.InputTokens != nil || call.Usage.OutputTokens != nil || strings.Contains(fmt.Sprintf("%+v", call), "secret-key") {
		t.Fatal("unreported usage invented or provider details exposed")
	}
}
