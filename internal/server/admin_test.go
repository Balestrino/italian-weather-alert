package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/Balestrino/italian-weather-alert/internal/health"
)

func TestAdminBrowserBoundary(t *testing.T) {
	for _, test := range []struct {
		name, host, origin, fetch string
		allowed                   bool
	}{
		{"local CLI", "127.0.0.1:8081", "", "", true},
		{"localhost", "localhost:18081", "http://localhost:18081", "same-origin", true},
		{"IPv6", "[::1]:8081", "http://[::1]:8081", "same-origin", true},
		{"IPv6 default port", "[::1]", "http://[::1]", "same-origin", true},
		{"navigation", "127.0.0.1:18081", "", "none", true},
		{"remote Host", "alerts.example", "", "", false},
		{"rebound Host", "evil.example:8081", "http://evil.example:8081", "same-origin", false},
		{"lookalike Host", "localhost.evil.example:8081", "", "", false},
		{"container Host", "admin:8081", "", "", false},
		{"cross origin", "127.0.0.1:8081", "https://evil.example", "", false},
		{"public local port", "127.0.0.1:8081", "http://127.0.0.1:8080", "same-site", false},
		{"cross site without origin", "127.0.0.1:8081", "", "cross-site", false},
		{"opaque origin", "127.0.0.1:8081", "null", "", false},
		{"origin credentials", "127.0.0.1:8081", "http://user@127.0.0.1:8081", "", false},
		{"origin path", "127.0.0.1:8081", "http://127.0.0.1:8081/", "", false},
		{"wrong scheme", "127.0.0.1:8081", "https://127.0.0.1:8081", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			// A side effect behind the same boundary used by the admin router proves
			// rejected requests cannot reach future mutation handlers.
			for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} {
				calls := 0
				h := adminBoundary(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(204) }))
				r := httptest.NewRequest(method, "http://127.0.0.1:8081/admin/sources", nil)
				r.Host = test.host
				if test.origin != "" {
					r.Header.Set("Origin", test.origin)
				}
				r.Header.Set("Sec-Fetch-Site", test.fetch)
				r.Header.Set("X-Forwarded-Host", "localhost:8081")
				r.Header.Set("X-Forwarded-For", "127.0.0.1")
				w := httptest.NewRecorder()
				h.ServeHTTP(w, r)
				if test.allowed && (calls != 1 || w.Code != 204) || !test.allowed && (calls != 0 || w.Code != 403) {
					t.Fatalf("%s: status=%d handler calls=%d", method, w.Code, calls)
				}
				if w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("X-Frame-Options") != "DENY" || w.Header().Get("Content-Security-Policy") == "" || w.Header().Get("Access-Control-Allow-Origin") != "" {
					t.Fatal("missing browser protection or exposed CORS")
				}
			}
		})
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1:8081/admin/sources", nil)
	r.Header.Add("Origin", "http://127.0.0.1:8081")
	r.Header.Add("Origin", "https://evil.example")
	if adminOriginMatches(r) {
		t.Fatal("multiple origins accepted")
	}
}

func TestAdministrationIsolatedFromPublicTransports(t *testing.T) {
	const secret = "admin-boundary-secret-sentinel"
	t.Setenv("IWA_POSTGRES_PASSWORD", secret)
	checks := map[string]health.Check{"postgres": func(context.Context) error { return errors.New(secret) }}
	fake := &publicQueriesFake{}
	public := httptest.NewServer(handlerWithLimits("public", checks, time.Now, PublicLimits{Allowance: 1000, Window: time.Minute, MaxPageSize: 100}, fake))
	defer public.Close()
	admin := httptest.NewServer(Handler("admin", checks))
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

func TestTailscaleAdminBoundary(t *testing.T) {
	const origin = "https://iwa.tail123.ts.net"
	for _, tc := range []struct {
		name, host, origin, fetch string
		allowed                   bool
	}{
		{"navigation", "iwa.tail123.ts.net", "", "none", true},
		{"form", "iwa.tail123.ts.net", origin, "same-origin", true},
		{"local", "127.0.0.1:8081", "http://127.0.0.1:8081", "same-origin", true},
		{"other tailnet host", "other.tail123.ts.net", "", "", false},
		{"suffix attack", "iwa.tail123.ts.net.evil.example", "", "", false},
		{"wrong origin", "iwa.tail123.ts.net", "https://evil.example", "", false},
		{"cleartext origin", "iwa.tail123.ts.net", "http://iwa.tail123.ts.net", "", false},
		{"opaque origin", "iwa.tail123.ts.net", "null", "", false},
		{"origin path", "iwa.tail123.ts.net", origin + "/", "", false},
		{"cross-site", "iwa.tail123.ts.net", "", "cross-site", false},
		{"local origin on proxy", "iwa.tail123.ts.net", "http://localhost", "", false},
		{"proxy origin on local", "localhost", origin, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
				calls := 0
				h := adminBoundaryWithOrigin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(204) }), origin)
				r := httptest.NewRequest(method, "http://127.0.0.1:8081/admin/sources", nil)
				r.Host = tc.host
				if tc.origin != "" {
					r.Header.Set("Origin", tc.origin)
				}
				r.Header.Set("Sec-Fetch-Site", tc.fetch)
				r.Header.Set("X-Forwarded-Host", "iwa.tail123.ts.net")
				r.Header.Set("X-Forwarded-Proto", "https")
				w := httptest.NewRecorder()
				h.ServeHTTP(w, r)
				if tc.allowed && (w.Code != 204 || calls != 1) || !tc.allowed && (w.Code != 403 || calls != 0) {
					t.Fatalf("%s: status=%d calls=%d", method, w.Code, calls)
				}
			}
		})
	}
	for _, configured := range []string{"", "https://*.ts.net", "http://iwa.tail123.ts.net"} {
		h := adminBoundaryWithOrigin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected access") }), configured)
		r := httptest.NewRequest("GET", origin+"/admin/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatal(w.Code)
		}
	}
	// Exercise both the main page and nested routes through the production router.
	h := HandlerWithAdministration(nil, AdminRuntime{TailscaleOrigin: origin})
	for _, path := range []string{"/admin/", "/admin/status", "/health/live", "/admin/sources"} {
		r := httptest.NewRequest("GET", "http://iwa.tail123.ts.net"+path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		expected := 200
		if path == "/admin/sources" {
			expected = 503
		} // No registry configured in this test.
		if w.Code != expected {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
	r := httptest.NewRequest("POST", origin+"/admin/sources", nil)
	r.Header.Add("Origin", origin)
	r.Header.Add("Origin", origin)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("duplicate origins accepted")
	}
}
