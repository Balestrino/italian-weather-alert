// Package acquisition discovers and retains documents within configured source
// boundaries. It never discovers new channels or changes activation state.
package acquisition

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidConfiguration = errors.New("invalid acquisition configuration")
	ErrCrawlUnavailable     = errors.New("crawl service unavailable")
	ErrUnrecognizedContent  = errors.New("crawl result has no recognizable original")
	ErrRequiredAttachment   = errors.New("required attachment unavailable")
)

type Link struct {
	URL  string
	Text string
}

type Page struct {
	URL        string
	HTML       []byte
	MediaType  string
	StatusCode int
	Links      []Link
}

// CrawlFailure distinguishes source reachability from crawler/service failure.
// Code is stable and safe to persist; RetryAfter is the source instruction
// returned by Crawl4AI, not a retry hint from the internal service itself.
type CrawlFailure struct {
	Code       string
	Reachable  bool
	StatusCode int
	RetryAfter *time.Time
	cause      error
}

func (e *CrawlFailure) Error() string { return e.Code }
func (e *CrawlFailure) Unwrap() error { return e.cause }

type Crawler interface {
	Crawl(context.Context, string) (Page, error)
}

type Crawl4AI struct {
	BaseURL string
	Token   string
	Client  *http.Client
	Now     func() time.Time
}

type crawlResponse struct {
	Success bool `json:"success"`
	Results []struct {
		URL             string            `json:"url"`
		HTML            string            `json:"html"`
		Success         bool              `json:"success"`
		StatusCode      int               `json:"status_code"`
		ResponseHeaders map[string]string `json:"response_headers"`
		Links           struct {
			Internal []struct {
				Href string `json:"href"`
				Text string `json:"text"`
			} `json:"internal"`
		} `json:"links"`
	} `json:"results"`
}

func (c *Crawl4AI) Crawl(ctx context.Context, target string) (Page, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") || base.User != nil || strings.TrimSpace(c.Token) == "" {
		return Page{}, ErrInvalidConfiguration
	}
	u, err := url.Parse(target)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" {
		return Page{}, ErrInvalidConfiguration
	}
	body, _ := json.Marshal(map[string]any{
		"urls":           []string{u.String()},
		"browser_config": map[string]any{},
		"crawler_config": map[string]any{"cache_mode": "BYPASS"},
	})
	endpoint := *base
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/crawl"
	endpoint.RawQuery = ""
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return Page{}, ErrCrawlUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return Page{}, fmt.Errorf("%w: request failed", ErrCrawlUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Page{}, &CrawlFailure{Code: "crawl_service_http_error", StatusCode: resp.StatusCode, cause: ErrCrawlUnavailable}
	}
	// A single retained document is capped at 32 MiB. Allow envelope overhead,
	// but never buffer an unbounded crawler response.
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 36<<20))
	if err != nil || len(raw) == 36<<20 {
		return Page{}, fmt.Errorf("%w: response too large", ErrCrawlUnavailable)
	}
	var decoded crawlResponse
	if json.Unmarshal(raw, &decoded) != nil || len(decoded.Results) != 1 {
		return Page{}, &CrawlFailure{Code: "crawl_service_invalid_response", cause: ErrCrawlUnavailable}
	}
	result := decoded.Results[0]
	page := Page{URL: result.URL, HTML: []byte(result.HTML), StatusCode: result.StatusCode}
	for _, item := range result.Links.Internal {
		page.Links = append(page.Links, Link{URL: item.Href, Text: item.Text})
	}
	reachable := result.StatusCode > 0
	now := time.Now
	if c.Now != nil {
		now = c.Now
	}
	retryAfter := parseRetryAfter(result.ResponseHeaders, now().UTC())
	if result.StatusCode == http.StatusTooManyRequests {
		return page, &CrawlFailure{Code: "source_rate_limited", Reachable: true, StatusCode: result.StatusCode, RetryAfter: retryAfter, cause: ErrCrawlUnavailable}
	}
	if result.StatusCode < 200 || result.StatusCode >= 300 || !result.Success || !decoded.Success {
		return page, &CrawlFailure{Code: "source_http_error", Reachable: reachable, StatusCode: result.StatusCode, RetryAfter: retryAfter, cause: ErrCrawlUnavailable}
	}
	if strings.TrimSpace(result.HTML) == "" {
		return page, &CrawlFailure{Code: "unrecognized_content", Reachable: true, StatusCode: result.StatusCode, cause: ErrUnrecognizedContent}
	}
	return page, nil
}

func parseRetryAfter(headers map[string]string, now time.Time) *time.Time {
	var raw string
	for key, value := range headers {
		if strings.EqualFold(key, "Retry-After") {
			raw = strings.TrimSpace(value)
			break
		}
	}
	if raw == "" {
		return nil
	}
	if seconds, err := strconv.ParseUint(raw, 10, 31); err == nil {
		value := now.Add(time.Duration(seconds) * time.Second).UTC()
		return &value
	}
	if value, err := http.ParseTime(raw); err == nil && value.After(now) {
		value = value.UTC()
		return &value
	}
	return nil
}
