package acquisition

import (
	"context"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

type unavailableCrawler struct {
	fixtureCrawler
	failed string
	status int
}

func (c *unavailableCrawler) Crawl(ctx context.Context, raw string) (Page, error) {
	if raw == c.failed {
		c.calls = append(c.calls, raw)
		code := "source_http_error"
		if c.status == 429 {
			code = "source_rate_limited"
		}
		return Page{URL: raw, StatusCode: c.status}, &CrawlFailure{Code: code, StatusCode: c.status, Reachable: true}
	}
	return c.fixtureCrawler.Crawl(ctx, raw)
}

func TestScheduledMissingDocumentDoesNotStarveLaterTargets(t *testing.T) {
	for _, mode := range []string{"404", "410", "deferred", "recovered", "429"} {
		t.Run(mode, func(t *testing.T) {
			now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
			section := calcinaia + "/tipi-di-notizia/notizie"
			missing := calcinaia + "/novita/a-missing"
			later := calcinaia + "/novita/z-later"
			crawler := &unavailableCrawler{fixtureCrawler: fixtureCrawler{pages: map[string]Page{
				section: {URL: section, StatusCode: 200, HTML: []byte(`<main id="configured"></main>`)},
				missing: {URL: missing, StatusCode: 200, HTML: []byte("restored")},
				later:   {URL: later, StatusCode: 200, HTML: []byte("later notice")},
			}}, failed: missing, status: 404}
			tracker := &fakeTracker{plan: RevisionPlan{Documents: []PlannedDocument{{URL: missing}, {URL: later}}}}
			switch mode {
			case "410":
				crawler.status = 410
			case "429":
				crawler.status = 429
			case "deferred":
				next := now.Add(time.Hour)
				tracker.plan.Documents[0].NextAttemptAt = &next
				tracker.plan.Documents[0].UnavailableStatus = 404
			case "recovered":
				crawler.failed = ""
				tracker.plan.Documents[0].UnavailableStatus = 404
			}
			reg := &fakeRegistry{version: registry.Version{Configuration: registry.Configuration{
				AccessMethod: "crawl4ai", Sections: []string{section}, Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true},
				Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, PaginationParameter: "page", MaxPagesPerSection: 2, MaxDocuments: 20, ListingContentMarkers: []string{"configured"}, ListingItemClass: "card-body", ListingDateClass: "data", ListingDateLayout: "02 Jan 2006", ListingDateLocale: "it", BootstrapDays: 30},
			}}}
			out := (&Engine{Registry: reg, Retained: &stableRetention{}, Crawler: crawler, Tracking: tracker, Now: func() time.Time { return now }}).Check(context.Background(), "calcinaia", 1, now)
			if mode == "429" {
				if out.Complete || out.ErrorCode != "source_rate_limited" || out.Documents != 0 || len(tracker.failures) != 0 {
					t.Fatalf("rate limit bypassed: %+v", out)
				}
				return
			}
			if tracker.versions[later] == 0 {
				t.Fatal("later document was starved")
			}
			if mode == "recovered" {
				if !out.Complete || out.Documents != 2 || !tracker.complete || len(tracker.cleared) != 1 {
					t.Fatalf("recovery failed: %+v", out)
				}
				return
			}
			if out.Complete || out.ErrorCode != "source_http_error" || out.Documents != 1 || tracker.complete {
				t.Fatalf("partial check hidden: %+v", out)
			}
			if mode == "deferred" {
				for _, u := range crawler.calls {
					if u == missing {
						t.Fatal("retried before due")
					}
				}
			} else if tracker.failures[missing] != crawler.status {
				t.Fatal("failure identity/status missing")
			}
		})
	}
}
