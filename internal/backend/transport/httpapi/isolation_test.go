package httpapi

import (
	"context"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/health"
	"github.com/Balestrino/italian-weather-alert/internal/backoffice"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAdministrationIsolatedFromPublicTransports(t *testing.T) {
	const secret = "admin-boundary-secret-sentinel"
	t.Setenv("IWA_POSTGRES_PASSWORD", secret)
	checks := map[string]health.Check{"postgres": func(context.Context) error { return errors.New(secret) }}
	fake := &publicQueriesFake{}
	public := httptest.NewServer(handlerWithLimits(checks, time.Now, PublicLimits{Allowance: 1000, Window: time.Minute, MaxPageSize: 100}, fake))
	defer public.Close()
	admin := httptest.NewServer(backoffice.Handler(checks))
	defer admin.Close()
	request := func(base, method, path string) (int, string) {
		t.Helper()
		r, err := http.NewRequest(method, base+path, strings.NewReader(`{"collection_enabled":true,"public_enabled":true}`))
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Forwarded-Host", "localhost:8081")
		res, err := http.DefaultClient.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(body), secret) {
			t.Fatal("private configuration/error leaked")
		}
		return res.StatusCode, string(body)
	}
	for _, path := range []string{"/admin/", "/admin/status", "/admin/sources", "/admin/sources/activate", "/admin/reprocess", "/admin/config", "/admin/notifications", "/admin/backups", "/admin/processing-evaluations", "/config", "/debug/vars", "/v1/admin/sources", "/v1/config"} {
		for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
			if status, _ := request(public.URL, method, path); status != 404 {
				t.Fatalf("public %s %s = %d", method, path, status)
			}
		}
	}
	for _, path := range []string{"/v1/municipalities", "/v1/search", "/v1/sources/coverage", "/v1/documents/7", "/v1/municipalities/050004/situation"} {
		for _, method := range []string{"POST", "PUT", "PATCH", "DELETE"} {
			if status, _ := request(public.URL, method, path); status < 400 {
				t.Fatalf("public mutation accepted: %s %s", method, path)
			}
		}
	}
	for _, base := range []string{public.URL, admin.URL} {
		if status, _ := request(base, "GET", "/health/ready"); status != 503 {
			t.Fatal("failure hidden")
		}
	}
	if status, body := request(admin.URL, "GET", "/admin/"); status != 200 || !strings.Contains(body, "Operazioni") {
		t.Fatal("admin page unavailable")
	}
	if status, body := request(admin.URL, "GET", "/admin/status"); status != 200 || !strings.Contains(body, `"access":"local"`) {
		t.Fatal("admin status unavailable")
	}
	for _, path := range []string{"/mcp", "/v1/sources/coverage", "/config", "/debug/vars"} {
		if status, _ := request(admin.URL, "GET", path); status != 404 {
			t.Fatalf("admin exposed %s", path)
		}
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "admin-isolation-test", Version: "1"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: public.URL + "/mcp", HTTPClient: public.Client(), DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	listed, err := session.ListTools(t.Context(), nil)
	if err != nil || len(listed.Tools) != 5 {
		t.Fatalf("unexpected public tools: %v", err)
	}
	for _, name := range []string{"activate_source", "update_source", "set_config", "reprocess", "get_admin_status", "get_secrets"} {
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: map[string]any{"source_id": "calcinaia", "enabled": true}})
		if err == nil && (result == nil || !result.IsError) {
			t.Fatalf("administrative tool accepted: %s", name)
		}
		if err != nil && strings.Contains(err.Error(), secret) {
			t.Fatal("MCP error leaked secret")
		}
	}
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "get_source_coverage", Arguments: map[string]any{"source_id": "calcinaia", "collection_enabled": true}})
	if err == nil && (result == nil || !result.IsError) {
		t.Fatal("mutation parameter accepted by public tool")
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.calls) != 0 {
		t.Fatalf("rejected operations reached application: %v", fake.calls)
	}
}
