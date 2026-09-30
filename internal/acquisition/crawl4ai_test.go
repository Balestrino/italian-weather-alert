package acquisition

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCrawl4AIContract(t *testing.T) {
	var received string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/crawl" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer private-token" {
			t.Fatal("missing bearer token")
		}
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"success":true,"results":[{"url":"https://www.comune.calcinaia.pi.it/tipi-di-notizia/notizie","html":"<html>original</html>","success":true,"status_code":200,"links":{"internal":[{"href":"/novita/allerta-meteo","text":"Allerta meteo"}]}}]}`)
	}))
	defer server.Close()

	client := Crawl4AI{BaseURL: server.URL, Token: "private-token", Client: server.Client()}
	page, err := client.Crawl(context.Background(), "https://www.comune.calcinaia.pi.it/tipi-di-notizia/notizie")
	if err != nil {
		t.Fatal(err)
	}
	if string(page.HTML) != "<html>original</html>" || page.StatusCode != 200 || len(page.Links) != 1 || page.Links[0].URL != "/novita/allerta-meteo" {
		t.Fatalf("unexpected page: %#v", page)
	}
	for _, forbidden := range []string{"deep_crawl_strategy", "js_code", "headers", "private-token"} {
		if strings.Contains(received, forbidden) {
			t.Fatalf("unsafe field or secret in request body: %s", forbidden)
		}
	}
	if !strings.Contains(received, `"cache_mode":"BYPASS"`) {
		t.Fatalf("cache bypass missing from request: %s", received)
	}
}

func TestCrawl4AIExposesSourceRetryAfter(t *testing.T) {
	base := time.Date(2026, 9, 16, 18, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"success":false,"results":[{"url":"https://source.example/list","html":"rate limited","success":false,"status_code":429,"response_headers":{"Retry-After":"1200"},"links":{"internal":[]}}]}`)
	}))
	defer server.Close()
	client := Crawl4AI{BaseURL: server.URL, Token: "token", Client: server.Client(), Now: func() time.Time { return base }}
	page, err := client.Crawl(context.Background(), "https://source.example/list")
	var failure *CrawlFailure
	if !errors.As(err, &failure) || failure.Code != "source_rate_limited" || !failure.Reachable || page.StatusCode != 429 || failure.RetryAfter == nil || !failure.RetryAfter.Equal(base.Add(20*time.Minute)) {
		t.Fatalf("unexpected rate-limit result: page=%#v failure=%#v err=%v", page, failure, err)
	}
}

func TestCrawl4AIRejectsInvalidAndUnrecognizedResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"success":true,"results":[{"url":"https://example.test","html":"","success":true,"status_code":200}]}`)
	}))
	defer server.Close()
	client := Crawl4AI{BaseURL: server.URL, Token: "token", Client: server.Client()}
	if _, err := client.Crawl(context.Background(), "https://example.test"); !errors.Is(err, ErrUnrecognizedContent) {
		t.Fatalf("expected unrecognized content, got %v", err)
	}
	client.Token = ""
	if _, err := client.Crawl(context.Background(), "https://example.test"); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatalf("expected invalid configuration, got %v", err)
	}
}
