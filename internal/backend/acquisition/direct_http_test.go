package acquisition

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDirectHTTPRejectsOutOfScopeRedirectBeforeRequest(t *testing.T) {
	outsideCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/allowed/start.pdf" {
			http.Redirect(w, r, "/private/other.pdf", http.StatusFound)
		} else {
			outsideCalls++
			_, _ = w.Write([]byte("outside"))
		}
	}))
	defer server.Close()
	allowed := func(raw string) bool { return strings.HasPrefix(raw, server.URL+"/allowed/") }
	client := &DirectHTTP{Client: server.Client()}
	_, err := client.CrawlBounded(context.Background(), server.URL+"/allowed/start.pdf", allowed)
	if err == nil || outsideCalls != 0 {
		t.Fatalf("redirect escaped scope: calls=%d err=%v", outsideCalls, err)
	}
	_, err = client.CrawlBounded(context.Background(), server.URL+"/private/other.pdf", allowed)
	if err == nil || outsideCalls != 0 {
		t.Fatal("initial URL escaped scope")
	}
}

func TestDirectHTTPRetainsExactBytesAndMediaType(t *testing.T) {
	want := []byte{0x25, 0x50, 0x44, 0x46, 0x00, 0xff, 0x10}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/pdf; version=1.7")
		_, _ = w.Write(want)
	}))
	defer server.Close()

	fetcher := &DirectHTTP{Client: server.Client()}
	page, err := fetcher.Crawl(context.Background(), server.URL+"/bulletin.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if string(page.HTML) != string(want) || page.MediaType != "application/pdf" || page.StatusCode != http.StatusOK {
		t.Fatalf("original response not retained: %#v", page)
	}
}

func TestDirectHTTPAllowsSameOriginAndRejectsCrossOriginRedirect(t *testing.T) {
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("foreign"))
	}))
	defer destination.Close()

	var origin *httptest.Server
	origin = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/same":
			http.Redirect(w, r, "/asset", http.StatusFound)
		case "/asset":
			_, _ = w.Write([]byte("same-origin"))
		case "/foreign":
			http.Redirect(w, r, destination.URL+"/asset", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer origin.Close()

	fetcher := &DirectHTTP{Client: origin.Client()}
	page, err := fetcher.Crawl(context.Background(), origin.URL+"/same")
	if err != nil || string(page.HTML) != "same-origin" {
		t.Fatalf("same-origin redirect failed: page=%#v err=%v", page, err)
	}
	_, err = fetcher.Crawl(context.Background(), origin.URL+"/foreign")
	var failure *CrawlFailure
	if !errors.As(err, &failure) || failure.Code != "resource_redirect_outside_origin" || !failure.Reachable {
		t.Fatalf("cross-origin redirect not rejected safely: %#v, %v", failure, err)
	}
}

func TestDirectHTTPExposesRateLimit(t *testing.T) {
	base := time.Date(2026, 9, 16, 20, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	fetcher := &DirectHTTP{Client: server.Client(), Now: func() time.Time { return base }}
	page, err := fetcher.Crawl(context.Background(), server.URL+"/map.png")
	var failure *CrawlFailure
	if !errors.As(err, &failure) || failure.Code != "source_rate_limited" || failure.RetryAfter == nil || !failure.RetryAfter.Equal(base.Add(time.Minute)) || page.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("unexpected rate-limit result: page=%#v failure=%#v err=%v", page, failure, err)
	}
}

func TestDirectHTTPRejectsDeclaredOversizedResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "33554433")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	fetcher := &DirectHTTP{Client: server.Client()}
	page, err := fetcher.Crawl(context.Background(), server.URL+"/oversized.pdf")
	var failure *CrawlFailure
	if !errors.As(err, &failure) || failure.Code != "resource_too_large" || page.StatusCode != http.StatusOK {
		t.Fatalf("oversized response not rejected: page=%#v failure=%#v err=%v", page, failure, err)
	}
}
