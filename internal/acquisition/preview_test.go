package acquisition

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

const calcinaia = "https://www.comune.calcinaia.pi.it"

type fakeRegistry struct {
	version  registry.Version
	recorded registry.Evidence
}

func (f *fakeRegistry) Version(context.Context, string, int) (registry.Version, error) {
	return f.version, nil
}
func (f *fakeRegistry) RecordPreview(_ context.Context, _ string, _ int, _ string, e registry.Evidence) error {
	f.recorded = e
	return nil
}

type fakeRetention struct {
	items []documents.Acquisition
}

func (f *fakeRetention) Retain(_ context.Context, a documents.Acquisition) (documents.Version, error) {
	f.items = append(f.items, a)
	return documents.Version{ID: int64(len(f.items))}, nil
}

type fixtureCrawler struct {
	pages map[string]Page
	calls []string
}

func (f *fixtureCrawler) Crawl(_ context.Context, raw string) (Page, error) {
	f.calls = append(f.calls, raw)
	p, ok := f.pages[raw]
	if !ok {
		return Page{}, fmt.Errorf("unexpected crawl %s", raw)
	}
	return p, nil
}

func TestCalcinaiaPreviewTraversesConfiguredListingsAndRetainsOriginals(t *testing.T) {
	section := calcinaia + "/tipi-di-notizia/notizie"
	page1 := section + "?page=1"
	page2 := section + "?page=2"
	doc1 := calcinaia + "/novita/allerta-meteo-mercoledi-9-e-giovedi-10-settembre"
	doc2 := calcinaia + "/novita/avviso-di-criticita-meteo-disposizioni-e-chiusure-sul-territorio-comunale"
	doc3 := calcinaia + "/novita/monitoraggio-situazione-del-territorio"
	crawler := &fixtureCrawler{pages: map[string]Page{
		section: {URL: section, HTML: []byte("listing zero"), StatusCode: 200, Links: []Link{
			{URL: "?page=1"}, {URL: doc1}, {URL: "https://unconfigured.example/alerts"}, {URL: "/amministrazione"},
		}},
		page1: {URL: page1, HTML: []byte("listing one with older dates"), StatusCode: 200, Links: []Link{
			{URL: "?page=0"}, {URL: "?page=2"}, {URL: doc2}, {URL: section + "?sort=unexpected"},
		}},
		page2: {URL: page2, HTML: []byte("listing two"), StatusCode: 200, Links: []Link{{URL: doc3}}},
		doc1:  {URL: doc1, HTML: []byte("original vigilance republication"), StatusCode: 200},
		doc2:  {URL: doc2, HTML: []byte("original closure notice"), StatusCode: 200},
		doc3:  {URL: doc3, HTML: []byte("original monitoring notice"), StatusCode: 200},
	}}
	policyEvidence := registry.Evidence{URL: calcinaia + "/note-legali", Locator: "CC BY 4.0", ObservedAt: time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)}
	reg := &fakeRegistry{version: registry.Version{SourceID: "calcinaia-municipal", Revision: 1, Configuration: registry.Configuration{
		URL: section, Sections: []string{section}, AccessMethod: "crawl4ai", Attribution: "Comune di Calcinaia",
		Policy:    registry.Policy{Evidence: &policyEvidence, CollectionPermitted: true, RetentionPermitted: true},
		Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, PaginationParameter: "page", MaxPagesPerSection: 150, MaxDocuments: 2000, ListingContentMarkers: []string{"listing"}},
	}}}
	retained := &fakeRetention{}
	observed := time.Date(2026, 9, 16, 18, 0, 0, 0, time.UTC)
	engine := Engine{Registry: reg, Retained: retained, Crawler: crawler, Now: func() time.Time { return observed }}

	report, err := engine.Preview(context.Background(), "calcinaia-municipal", 1, "operator")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Sections) != 1 || len(report.Sections[0].Listings) != 3 || len(report.Documents) != 3 {
		t.Fatalf("unexpected preview counts: %#v", report)
	}
	if len(retained.items) != 6 {
		t.Fatalf("expected every listing/document original retained, got %d", len(retained.items))
	}
	for _, item := range retained.items {
		if len(item.Resources) != 1 || item.Resources[0].Role != "original" || string(item.Resources[0].Bytes) == "" || item.Resources[0].URL != item.URL {
			t.Fatalf("original not retained correctly: %#v", item)
		}
	}
	joined := strings.Join(crawler.calls, "\n")
	if strings.Contains(joined, "unconfigured.example") || strings.Contains(joined, "/amministrazione") || strings.Contains(joined, "sort=unexpected") {
		t.Fatalf("preview escaped configured boundaries: %s", joined)
	}
	if reg.recorded.ObservedAt != observed || !strings.Contains(reg.recorded.Locator, "3 listings, 3 documents") {
		t.Fatalf("preview evidence not recorded: %#v", reg.recorded)
	}
}

func TestPreviewStopsAtConfiguredPageBound(t *testing.T) {
	section := calcinaia + "/tipi-di-notizia/notizie"
	reg := &fakeRegistry{version: registry.Version{Configuration: registry.Configuration{
		AccessMethod: "crawl4ai", Sections: []string{section}, Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true},
		Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, PaginationParameter: "page", MaxPagesPerSection: 1, MaxDocuments: 2000, ListingContentMarkers: []string{"first"}},
	}}}
	crawler := &fixtureCrawler{pages: map[string]Page{section: {HTML: []byte("first"), Links: []Link{{URL: "?page=1"}}}}}
	engine := Engine{Registry: reg, Retained: &fakeRetention{}, Crawler: crawler}
	if _, err := engine.Preview(context.Background(), "calcinaia", 1, "operator"); err == nil || !strings.Contains(err.Error(), "page limit") {
		t.Fatalf("expected page bound failure, got %v", err)
	}
}

func TestCheckKeepsHTTP200InvalidContentIncomplete(t *testing.T) {
	section := calcinaia + "/tipi-di-notizia/notizie"
	reg := &fakeRegistry{version: registry.Version{Configuration: registry.Configuration{
		AccessMethod: "crawl4ai", Sections: []string{section}, Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true},
		Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, PaginationParameter: "page", MaxPagesPerSection: 2, MaxDocuments: 20, ListingContentMarkers: []string{"configured-listing-marker"}},
	}}}
	crawler := &fixtureCrawler{pages: map[string]Page{section: {URL: section, HTML: []byte("generic HTTP 200 login page"), StatusCode: 200}}}
	retained := &fakeRetention{}
	finished := time.Date(2026, 9, 16, 18, 1, 0, 0, time.UTC)
	engine := Engine{Registry: reg, Retained: retained, Crawler: crawler, Now: func() time.Time { return finished }}
	outcome := engine.Check(context.Background(), "calcinaia", 1, finished.Add(-time.Minute))
	if !outcome.Reachable || outcome.ContentRecognized || outcome.Complete || outcome.ErrorCode != "unrecognized_content" {
		t.Fatalf("invalid content advanced check: %#v", outcome)
	}
	if len(retained.items) != 0 {
		t.Fatal("unrecognized content was retained as a successful listing")
	}
}

type fakeTracker struct {
	failures map[string]int
	cleared  []string
	plan     RevisionPlan
	versions map[string]int64
	complete bool
}

func (f *fakeTracker) Remember(context.Context, string, int, []DiscoveredDocument, time.Time) (DiscoveryResult, error) {
	return DiscoveryResult{}, nil
}
func (f *fakeTracker) Plan(context.Context, string, int, time.Time, int) (RevisionPlan, error) {
	result := f.plan
	result.Bootstrap = !f.complete
	return result, nil
}
func (f *fakeTracker) RecordVersion(_ context.Context, _ string, raw string, version int64, _ time.Time) (bool, error) {
	if f.versions == nil {
		f.versions = map[string]int64{}
	}
	previous, exists := f.versions[raw]
	f.versions[raw] = version
	return exists && previous != version, nil
}
func (f *fakeTracker) CompleteBootstrap(context.Context, string, int, time.Time) error {
	f.complete = true
	return nil
}

type stableRetention struct {
	ids   map[[32]byte]int64
	items []documents.Acquisition
}

func (s *stableRetention) Retain(_ context.Context, item documents.Acquisition) (documents.Version, error) {
	s.items = append(s.items, item)
	body, _ := json.Marshal(item)
	hash := sha256.Sum256(body)
	if s.ids == nil {
		s.ids = map[[32]byte]int64{}
	}
	id, exists := s.ids[hash]
	if !exists {
		id = int64(len(s.ids) + 1)
		s.ids[hash] = id
	}
	return documents.Version{ID: id}, nil
}

func TestChangedAttachmentRevisesUnchangedOldDocument(t *testing.T) {
	section := calcinaia + "/tipi-di-notizia/notizie"
	parent := calcinaia + "/novita/vecchia-chiusura-irrisolta"
	attachment := calcinaia + "/sites/default/files/ordinanza.pdf"
	published := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	listing := []byte(`<main id="configured"><div class="card-body"><span class="data">01 Giu 2026</span><a href="/novita/vecchia-chiusura-irrisolta">Chiusura</a></div></main>`)
	crawler := &fixtureCrawler{pages: map[string]Page{
		section:    {URL: section, HTML: listing, StatusCode: 200},
		parent:     {URL: parent, HTML: []byte("unchanged parent"), StatusCode: 200},
		attachment: {URL: attachment, HTML: []byte("PDF revision one"), MediaType: "application/pdf", StatusCode: 200},
	}}
	reg := &fakeRegistry{version: registry.Version{Configuration: registry.Configuration{
		AccessMethod: "crawl4ai", Sections: []string{section}, Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true},
		Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, PaginationParameter: "page", MaxPagesPerSection: 2, MaxDocuments: 20, ListingContentMarkers: []string{"configured"}, ListingItemClass: "card-body", ListingDateClass: "data", ListingDateLayout: "02 Jan 2006", ListingDateLocale: "it", BootstrapDays: 30},
	}}}
	tracker := &fakeTracker{plan: RevisionPlan{Documents: []PlannedDocument{{URL: parent, PublicationDate: &published, Resources: []PlannedResource{{URL: attachment, Required: true}}}}}}
	retained := &stableRetention{}
	firstTime := time.Date(2026, 9, 16, 18, 0, 0, 0, time.UTC)
	engine := Engine{Registry: reg, Retained: retained, Crawler: crawler, Tracking: tracker, Now: func() time.Time { return firstTime }}
	first := engine.Check(context.Background(), "calcinaia", 1, firstTime.Add(-time.Minute))
	if !first.Complete || first.Revisions != 0 || first.Documents != 1 {
		t.Fatalf("initial old ongoing acquisition failed: %#v", first)
	}
	crawler.pages[attachment] = Page{URL: attachment, HTML: []byte("PDF revision two"), MediaType: "application/pdf", StatusCode: 200}
	secondTime := firstTime.Add(10 * time.Minute)
	engine.Now = func() time.Time { return secondTime }
	second := engine.Check(context.Background(), "calcinaia", 1, secondTime.Add(-time.Minute))
	if !second.Complete || second.Revisions != 1 {
		t.Fatalf("changed attachment was not a parent evidence revision: %#v", second)
	}
	last := retained.items[len(retained.items)-1]
	if len(last.Resources) != 2 || last.Resources[1].Role != "attachment" || string(last.Resources[1].Bytes) != "PDF revision two" || string(last.Metadata) != `{"source_publication_date":"2026-06-01"}` {
		t.Fatalf("attachment or source chronology not retained: %#v", last)
	}
}

func TestMissingRequiredAttachmentRetainsDiscoverableReference(t *testing.T) {
	section := calcinaia + "/tipi-di-notizia/notizie"
	parent := calcinaia + "/novita/ordinanza-collegata"
	attachment := calcinaia + "/sites/default/files/ordinanza-mancante.pdf"
	published := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)
	listing := []byte(`<main id="configured"><div class="card-body"><span class="data">17 Set 2026</span><a href="/novita/ordinanza-collegata">Chiusura</a></div></main>`)
	crawler := &fixtureCrawler{pages: map[string]Page{
		section: {URL: section, HTML: listing, StatusCode: 200},
		parent:  {URL: parent, HTML: []byte("pagina che collega l'ordinanza"), StatusCode: 200},
	}}
	reg := &fakeRegistry{version: registry.Version{Configuration: registry.Configuration{
		AccessMethod: "crawl4ai", Sections: []string{section}, Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true},
		Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, PaginationParameter: "page", MaxPagesPerSection: 2, MaxDocuments: 20, ListingContentMarkers: []string{"configured"}, ListingItemClass: "card-body", ListingDateClass: "data", ListingDateLayout: "02 Jan 2006", ListingDateLocale: "it", BootstrapDays: 30},
	}}}
	tracker := &fakeTracker{plan: RevisionPlan{Documents: []PlannedDocument{{URL: parent, PublicationDate: &published, Resources: []PlannedResource{{URL: attachment, Required: true}}}}}}
	retained := &stableRetention{}
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	outcome := (&Engine{Registry: reg, Retained: retained, Crawler: crawler, Tracking: tracker, Now: func() time.Time { return now }}).Check(context.Background(), "calcinaia", 1, now.Add(-time.Minute))
	if outcome.Complete || outcome.ErrorCode != "required_attachment_unavailable" || outcome.Documents != 1 || len(retained.items) != 2 {
		t.Fatalf("missing required attachment was hidden or marked complete: %#v items=%d", outcome, len(retained.items))
	}
	last := retained.items[len(retained.items)-1]
	if len(last.Resources) != 2 || last.Resources[1].URL != attachment || !last.Resources[1].Required || last.Resources[1].Missing != "unavailable" || last.Resources[1].Bytes != nil {
		t.Fatalf("missing attachment reference not retained: %#v", last.Resources)
	}
}

func (f *fakeTracker) RecordUnavailable(_ context.Context, _ string, _ int, raw string, status int, _ time.Time) error {
	if f.failures == nil {
		f.failures = map[string]int{}
	}
	f.failures[raw] = status
	return nil
}
func (f *fakeTracker) ClearUnavailable(_ context.Context, _ string, _ int, raw string, _ time.Time) error {
	f.cleared = append(f.cleared, raw)
	return nil
}
