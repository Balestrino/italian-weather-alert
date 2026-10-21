package acquisition

import (
	"context"
	"errors"
	"strings"
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
			tracker := &fakeTracker{plan: RevisionPlan{Documents: []PlannedDocument{{URL: missing, Listed: true}, {URL: later, Listed: true}}}}
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

func TestScheduledWithdrawnTargetsDoNotBlockTheCheck(t *testing.T) {
	cases := []struct {
		name                string
		missing             PlannedDocument
		blocking            bool
		complete, retried   bool
		withdrawn, failures int
	}{
		{name: "first 404 after leaving listings", missing: PlannedDocument{}, complete: true, retried: true, withdrawn: 1, failures: 1},
		{name: "410 after leaving listings", missing: PlannedDocument{UnavailableStatus: 404}, complete: true, retried: true, withdrawn: 1, failures: 1},
		{name: "deferred after leaving listings", missing: PlannedDocument{UnavailableStatus: 404, NextAttemptAt: new(time.Time)}, complete: true, withdrawn: 1},
		{name: "still listed", missing: PlannedDocument{Listed: true}, retried: true, failures: 1},
		{name: "protected by a measure", missing: PlannedDocument{Protected: true}, retried: true, failures: 1},
		{name: "deferred protected", missing: PlannedDocument{Protected: true, UnavailableStatus: 404, NextAttemptAt: new(time.Time)}},
		{name: "withdrawn beside a blocking target", missing: PlannedDocument{}, blocking: true, retried: true, withdrawn: 1, failures: 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
			section := calcinaia + "/tipi-di-notizia/notizie"
			missing := calcinaia + "/novita/a-renamed"
			listed := calcinaia + "/novita/b-listed-missing"
			later := calcinaia + "/novita/z-later"
			crawler := &multiUnavailableCrawler{fixtureCrawler: fixtureCrawler{pages: map[string]Page{
				section: {URL: section, StatusCode: 200, HTML: []byte(`<main id="configured"></main>`)},
				later:   {URL: later, StatusCode: 200, HTML: []byte("later notice")},
			}}, failed: map[string]int{missing: 404, listed: 404}}
			if tc.name == "410 after leaving listings" {
				crawler.failed[missing] = 410
			}
			target := tc.missing
			target.URL = missing
			if target.NextAttemptAt != nil {
				next := now.Add(time.Hour)
				target.NextAttemptAt = &next
			}
			documents := []PlannedDocument{target}
			if tc.blocking {
				documents = append(documents, PlannedDocument{URL: listed, Listed: true})
			}
			documents = append(documents, PlannedDocument{URL: later, Listed: true})
			tracker := &fakeTracker{plan: RevisionPlan{Documents: documents}, complete: true}
			reg := &fakeRegistry{version: registry.Version{Configuration: registry.Configuration{
				AccessMethod: "crawl4ai", Sections: []string{section}, Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true},
				Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, PaginationParameter: "page", MaxPagesPerSection: 2, MaxDocuments: 20, ListingContentMarkers: []string{"configured"}, ListingItemClass: "card-body", ListingDateClass: "data", ListingDateLayout: "02 Jan 2006", ListingDateLocale: "it", BootstrapDays: 30},
			}}}
			out := (&Engine{Registry: reg, Retained: &stableRetention{}, Crawler: crawler, Tracking: tracker, Now: func() time.Time { return now }}).Check(context.Background(), "calcinaia", 1, now)
			if tracker.versions[later] == 0 {
				t.Fatal("later document was starved")
			}
			if out.Complete != tc.complete || out.WithdrawnUnavailable != tc.withdrawn || len(tracker.failures) != tc.failures {
				t.Fatalf("outcome %+v failures %v", out, tracker.failures)
			}
			if tc.complete && (out.ErrorCode != "" || !out.ContentRecognized) {
				t.Fatalf("complete check kept an error state: %+v", out)
			}
			if !tc.complete && out.ErrorCode != "source_http_error" {
				t.Fatalf("blocking target hidden: %+v", out)
			}
			retried := false
			for _, u := range crawler.calls {
				retried = retried || u == missing
			}
			if retried != tc.retried {
				t.Fatalf("retried=%v, want %v", retried, tc.retried)
			}
		})
	}
}

type multiUnavailableCrawler struct {
	fixtureCrawler
	failed map[string]int
}

func (c *multiUnavailableCrawler) Crawl(ctx context.Context, raw string) (Page, error) {
	if status, ok := c.failed[raw]; ok {
		c.calls = append(c.calls, raw)
		return Page{URL: raw, StatusCode: status}, &CrawlFailure{Code: "source_http_error", StatusCode: status, Reachable: true}
	}
	return c.fixtureCrawler.Crawl(ctx, raw)
}

func TestScheduledCheckRefreshesInterpretationStateBeforePlanning(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	section := calcinaia + "/tipi-di-notizia/notizie"
	later := calcinaia + "/novita/z-later"
	for _, failing := range []bool{false, true} {
		crawler := &fixtureCrawler{pages: map[string]Page{
			section: {URL: section, StatusCode: 200, HTML: []byte(`<main id="configured"></main>`)},
			later:   {URL: later, StatusCode: 200, HTML: []byte("later notice")},
		}}
		tracker := &fakeTracker{plan: RevisionPlan{Documents: []PlannedDocument{{URL: later, Listed: true}}}, complete: true}
		if failing {
			tracker.refreshError = errors.New("database unavailable")
		}
		cfg := registry.Configuration{
			AccessMethod: "crawl4ai", Sections: []string{section}, Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true},
			Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, PaginationParameter: "page", MaxPagesPerSection: 2, MaxDocuments: 20, ListingContentMarkers: []string{"configured"}, ListingItemClass: "card-body", ListingDateClass: "data", ListingDateLayout: "02 Jan 2006", ListingDateLocale: "it", BootstrapDays: 30, UnresolvedMeasureDays: 45},
		}
		out := (&Engine{Registry: &fakeRegistry{version: registry.Version{Configuration: cfg}}, Retained: &stableRetention{}, Crawler: crawler, Tracking: tracker, Now: func() time.Time { return now }}).Check(context.Background(), "calcinaia", 1, now)
		if failing {
			if out.Complete || out.ErrorCode != "interpretation_state_failed" || strings.Join(tracker.calls, ",") != "remember,refresh" {
				t.Fatalf("refresh failure planned anyway: %+v calls=%v", out, tracker.calls)
			}
			continue
		}
		if !out.Complete || strings.Join(tracker.calls, ",") != "remember,refresh,plan" || tracker.refreshDays != 45 {
			t.Fatalf("refresh not ordered before planning: %+v calls=%v days=%d", out, tracker.calls, tracker.refreshDays)
		}
	}
}
