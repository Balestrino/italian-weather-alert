package httpapi

import (
	"context"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backoffice"
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/health"
)

func TestPublicIsolationAndReadiness(t *testing.T) {
	checks := map[string]health.Check{"postgres": func(context.Context) error { return errors.New("password=private-fixture-value") }}
	h := Handler(checks)
	for _, path := range []string{"/admin/status", "/admin/sources", "/config", "/debug/vars"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatalf("exposed %s", path)
		}
	}
	for _, role := range []string{"public", "admin"} {
		if role == "admin" {
			h = backoffice.Handler(checks)
		} else {
			h = Handler(checks)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1:8081/health/ready", nil))
		if w.Code != 503 || strings.Contains(w.Body.String(), "private-fixture-value") {
			t.Fatal("readiness failure hidden or credential exposed")
		}
		checks["postgres"] = func(context.Context) error { return nil }
		w = httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1:8081/health/ready", nil))
		if w.Code != 200 {
			t.Fatal("recovery hidden")
		}
		checks["postgres"] = func(context.Context) error { return errors.New("password=private-fixture-value") }
	}
}
func TestListenerShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- httpserver.Serve(ctx, ln, Handler(nil)) }()
	client := http.Client{Timeout: time.Second}
	resp, err := client.Get("http://" + ln.Addr().String() + "/health/live")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"status":"alive"`) {
		t.Fatal("unexpected admin status")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown did not finish")
	}
}
