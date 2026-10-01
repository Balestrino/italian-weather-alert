package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSlidingBudgetAndRetry(t *testing.T) {
	now := time.Date(2026, 9, 17, 13, 0, 0, 0, time.UTC)
	budget := newSlidingBudget(PublicLimits{Allowance: 2, Window: 10 * time.Second, MaxPageSize: 25}, func() time.Time { return now })
	first := budget.take("192.0.2.1")
	second := budget.take("192.0.2.1")
	denied := budget.take("192.0.2.1")
	if !first.Allowed || first.Remaining != 1 || !second.Allowed || second.Remaining != 0 || denied.Allowed || denied.RetryAfterSeconds != 10 {
		t.Fatalf("unexpected initial states: %#v %#v %#v", first, second, denied)
	}
	now = now.Add(6 * time.Second)
	if state := budget.take("192.0.2.1"); state.Allowed || state.RetryAfterSeconds != 4 {
		t.Fatalf("sliding retry = %#v", state)
	}
	now = now.Add(4 * time.Second)
	if state := budget.take("192.0.2.1"); !state.Allowed || state.Remaining != 1 {
		t.Fatalf("expired event was not released: %#v", state)
	}
}

func TestSharedAPIAndMCPBudgetCountsInvalidRequests(t *testing.T) {
	now := time.Date(2026, 9, 17, 13, 0, 0, 0, time.UTC)
	limits := PublicLimits{Allowance: 1, Window: 30 * time.Second, MaxPageSize: 20}
	handler := handlerWithLimits(nil, func() time.Time { return now }, limits, &publicQueriesFake{})

	invalid := httptest.NewRequest(http.MethodGet, "/v1/search?kind=invalid", nil)
	invalid.RemoteAddr = "192.0.2.10:1000"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, invalid)
	if response.Code != http.StatusBadRequest || response.Header().Get("RateLimit") != `"iwa-public";r=0;t=30` || response.Header().Get("X-IWA-Max-Page-Size") != "20" {
		t.Fatalf("invalid request accounting: %d %#v %s", response.Code, response.Header(), response.Body.String())
	}

	mcpRequest := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28"}}}`))
	mcpRequest.RemoteAddr = "192.0.2.10:2000"
	mcpRequest.Header.Set("Content-Type", "application/json")
	mcpRequest.Header.Set("Mcp-Protocol-Version", publicMCPProtocolVersion)
	mcpRequest.Header.Set("Mcp-Method", "server/discover")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, mcpRequest)
	if response.Code != http.StatusTooManyRequests || response.Header().Get("Retry-After") != "30" {
		t.Fatalf("MCP did not share API budget: %d %#v %s", response.Code, response.Header(), response.Body.String())
	}
	var envelope publicResponse
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil || envelope.Error == nil || envelope.Error.Code != "rate_limited" || envelope.Error.RetryAfterSeconds == nil || *envelope.Error.RetryAfterSeconds != 30 {
		t.Fatalf("rate error is not actionable: %#v %v", envelope, err)
	}
}

func TestUnknownPublicAPIRequestConsumesBudget(t *testing.T) {
	now := time.Date(2026, 9, 17, 13, 0, 0, 0, time.UTC)
	limits := PublicLimits{Allowance: 1, Window: 30 * time.Second, MaxPageSize: 20}
	handler := handlerWithLimits(nil, func() time.Time { return now }, limits, &publicQueriesFake{})
	unknown := httptest.NewRequest(http.MethodPost, "/v1/not-an-operation", nil)
	unknown.RemoteAddr = "192.0.2.44:1000"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, unknown)
	if response.Code != http.StatusNotFound {
		t.Fatalf("unknown request status = %d", response.Code)
	}
	valid := httptest.NewRequest(http.MethodGet, "/v1/municipalities?istat=050004&evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z", nil)
	valid.RemoteAddr = "192.0.2.44:2000"
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, valid)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("unknown public request did not consume budget: %d", response.Code)
	}
}

func TestTrustedProxyClientAddress(t *testing.T) {
	now := time.Date(2026, 9, 17, 13, 0, 0, 0, time.UTC)
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("2001:db8:ffff::/48")}
	budget := newSlidingBudget(PublicLimits{Allowance: 1, Window: time.Minute, MaxPageSize: 10, TrustedProxies: trusted}, func() time.Time { return now })
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := budget.middleware(next)

	call := func(remote, forwarded string) int {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, "/v1/search", nil)
		request.RemoteAddr = remote
		if forwarded != "" {
			request.Header.Set("X-Forwarded-For", forwarded)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response.Code
	}
	if got := call("10.1.1.1:1000", "203.0.113.10"); got != http.StatusNoContent {
		t.Fatalf("first trusted-forwarded client = %d", got)
	}
	if got := call("10.1.1.1:1001", "203.0.113.11"); got != http.StatusNoContent {
		t.Fatalf("distinct forwarded client shared budget: %d", got)
	}
	if got := call("10.2.2.2:1002", "203.0.113.10, 10.1.1.1"); got != http.StatusTooManyRequests {
		t.Fatalf("right-to-left trusted chain bypassed budget: %d", got)
	}
	if got := call("192.0.2.20:1000", "198.51.100.1"); got != http.StatusNoContent {
		t.Fatalf("first untrusted peer = %d", got)
	}
	if got := call("192.0.2.20:1001", "198.51.100.2"); got != http.StatusTooManyRequests {
		t.Fatalf("untrusted spoofed header changed identity: %d", got)
	}
	if got := call("10.3.3.3:1000", "malformed"); got != http.StatusNoContent {
		t.Fatalf("malformed header did not safely fall back to peer: %d", got)
	}
	if got := call("10.3.3.3:1001", "still-malformed"); got != http.StatusTooManyRequests {
		t.Fatalf("malformed-header fallback did not consume peer budget: %d", got)
	}
}

func TestConfiguredMaximumPageSizeIsPublishedAndEnforced(t *testing.T) {
	limits := PublicLimits{Allowance: 10, Window: time.Minute, MaxPageSize: 7}
	handler := handlerWithLimits(nil, time.Now, limits, &publicQueriesFake{coverageCount: 8})
	request := httptest.NewRequest(http.MethodGet, "/v1/municipalities?page_size=8", nil)
	request.RemoteAddr = "192.0.2.1:1234"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || response.Header().Get("X-IWA-Max-Page-Size") != strconv.Itoa(limits.MaxPageSize) || !strings.Contains(response.Body.String(), "invalid_parameters") {
		t.Fatalf("maximum page size response: %d %#v %s", response.Code, response.Header(), response.Body.String())
	}
	request = httptest.NewRequest(http.MethodGet, "/v1/sources/coverage?municipality_istat=050004&evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z", nil)
	request.RemoteAddr = "192.0.2.2:1234"
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "exceeds the configured page size") {
		t.Fatalf("oversized result escaped the cap: %d %s", response.Code, response.Body.String())
	}
}
