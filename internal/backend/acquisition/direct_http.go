package acquisition

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
)

var errResourceRedirectOutsideOrigin = errors.New("resource redirect outside configured origin")

// DirectHTTP fetches retained binary dependencies without passing them through
// the HTML crawler. Redirects cannot leave the origin selected by the source
// contract and response bodies share the retention layer's size ceiling.
type DirectHTTP struct {
	Client *http.Client
	Now    func() time.Time
}

func (d *DirectHTTP) Crawl(ctx context.Context, target string) (Page, error) {
	return d.crawl(ctx, target, nil)
}

// CrawlBounded checks both the initial URL and every redirect before a request.
func (d *DirectHTTP) CrawlBounded(ctx context.Context, target string, allowed func(string) bool) (Page, error) {
	if allowed == nil || !allowed(target) {
		return Page{}, ErrInvalidConfiguration
	}
	return d.crawl(ctx, target, allowed)
}

func (d *DirectHTTP) crawl(ctx context.Context, target string, allowed func(string) bool) (Page, error) {
	original, err := url.Parse(target)
	if err != nil || original.Host == "" || (original.Scheme != "http" && original.Scheme != "https") || original.User != nil || original.Fragment != "" {
		return Page{}, ErrInvalidConfiguration
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, original.String(), nil)
	if err != nil {
		return Page{}, ErrInvalidConfiguration
	}

	baseClient := d.Client
	if baseClient == nil {
		baseClient = http.DefaultClient
	}
	client := *baseClient
	configuredRedirect := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if !sameHTTPOrigin(original, req.URL) || allowed != nil && !allowed(req.URL.String()) {
			return errResourceRedirectOutsideOrigin
		}
		if configuredRedirect != nil {
			return configuredRedirect(req, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, errResourceRedirectOutsideOrigin) {
			return Page{}, &CrawlFailure{Code: "resource_redirect_outside_origin", Reachable: true, cause: errResourceRedirectOutsideOrigin}
		}
		return Page{}, &CrawlFailure{Code: "source_request_failed", cause: fmt.Errorf("%w: resource request failed", ErrCrawlUnavailable)}
	}
	defer resp.Body.Close()

	page := Page{URL: resp.Request.URL.String(), StatusCode: resp.StatusCode, MediaType: responseMediaType(resp.Header.Get("Content-Type"))}
	now := time.Now
	if d.Now != nil {
		now = d.Now
	}
	retryAfter := parseRetryAfter(map[string]string{"Retry-After": resp.Header.Get("Retry-After")}, now().UTC())
	if resp.StatusCode == http.StatusTooManyRequests {
		return page, &CrawlFailure{Code: "source_rate_limited", Reachable: true, StatusCode: resp.StatusCode, RetryAfter: retryAfter, cause: ErrCrawlUnavailable}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return page, &CrawlFailure{Code: "source_http_error", Reachable: true, StatusCode: resp.StatusCode, RetryAfter: retryAfter, cause: ErrCrawlUnavailable}
	}
	if resp.ContentLength > documents.MaxAcquisitionBytes {
		return page, &CrawlFailure{Code: "resource_too_large", Reachable: true, StatusCode: resp.StatusCode, cause: ErrCrawlUnavailable}
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, documents.MaxAcquisitionBytes+1))
	if readErr != nil {
		return page, &CrawlFailure{Code: "source_read_failed", Reachable: true, StatusCode: resp.StatusCode, cause: ErrCrawlUnavailable}
	}
	if len(body) > documents.MaxAcquisitionBytes {
		return page, &CrawlFailure{Code: "resource_too_large", Reachable: true, StatusCode: resp.StatusCode, cause: ErrCrawlUnavailable}
	}
	page.HTML = body
	if len(body) == 0 {
		return page, &CrawlFailure{Code: "unrecognized_content", Reachable: true, StatusCode: resp.StatusCode, cause: ErrUnrecognizedContent}
	}
	return page, nil
}

func sameHTTPOrigin(a, b *url.URL) bool {
	return a != nil && b != nil && a.Scheme == b.Scheme && strings.EqualFold(a.Host, b.Host)
}

func responseMediaType(header string) string {
	value, _, err := mime.ParseMediaType(header)
	if err != nil {
		return ""
	}
	return value
}
