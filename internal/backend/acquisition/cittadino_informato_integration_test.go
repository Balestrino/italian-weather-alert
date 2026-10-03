//go:build integration

package acquisition

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func TestCittadinoInformatoPersistenceAndScope(t *testing.T) {
	ctx := context.Background()
	pool := acquisitionTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(registry.Migrate(ctx, pool))
	must(documents.Migrate(ctx, pool))
	must(Migrate(ctx, pool))
	cfg, crawler := cittadinoFixture()
	reg := registry.New(pool)
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "municipality", Name: "Synthetic municipality", OfficialURL: "https://municipal.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "platform", PublisherID: "municipality", Platform: "cittadino-informato", URL: cfg.URL, External: true}))
	source := registry.Source{ID: "platform-source", AuthorityID: "municipality", ChannelID: "platform", ProductID: "municipal", Territory: "050004"}
	must(reg.CreateSource(ctx, source, cfg, "fixture"))
	wrong := source
	wrong.ID = "unselected"
	wrong.Territory = "050008"
	if err := reg.CreateSource(ctx, wrong, cfg, "fixture"); err != registry.ErrInvalid {
		t.Fatalf("municipality scope escaped: %v", err)
	}
	objects := scopedObjects{}
	retained := documents.New(pool, objects)
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	engine := Engine{Registry: reg, Retained: retained, Crawler: crawler, Resources: crawler, Tracking: NewTrackingStore(pool), Now: func() time.Time { return now }}
	first := engine.Check(ctx, source.ID, 1, now)
	if !first.Complete || first.NewDocuments != 4 || first.Documents != 4 {
		t.Fatalf("first check: %#v", first)
	}
	var versions int
	must(pool.QueryRow(ctx, "SELECT count(*) FROM retained_versions").Scan(&versions))
	if versions != 6 {
		t.Fatalf("missing listing/notice/risk versions: %d", versions)
	}
	raw := cittadinoDetailURL(cfg.CittadinoInformato, 1)
	var originalVersion int64
	must(pool.QueryRow(ctx, "SELECT last_version_id FROM acquisition_targets WHERE source_id=$1 AND url=$2", source.ID, raw).Scan(&originalVersion))
	body, err := retained.Read(ctx, originalVersion, raw)
	must(err)
	if string(body) != string(crawler.pages[raw].HTML) {
		t.Fatal("original JSON changed")
	}
	now = now.Add(time.Minute)
	second := engine.Check(ctx, source.ID, 1, now)
	if !second.Complete || second.NewDocuments != 0 || second.Revisions != 0 {
		t.Fatalf("repeat check: %#v", second)
	}
	must(pool.QueryRow(ctx, "SELECT count(*) FROM retained_versions").Scan(&versions))
	if versions != 6 {
		t.Fatal("unchanged versions duplicated")
	}
	page := crawler.pages[raw]
	page.HTML = []byte(strings.ReplaceAll(string(page.HTML), "Provvedimento sintetico.", "Provvedimento sintetico aggiornato."))
	crawler.pages[raw] = page
	now = now.Add(time.Minute)
	third := engine.Check(ctx, source.ID, 1, now)
	if !third.Complete || third.Revisions != 1 {
		t.Fatalf("revised check: %#v", third)
	}
	must(pool.QueryRow(ctx, "SELECT count(*) FROM retained_versions").Scan(&versions))
	if versions != 7 {
		t.Fatal("revised version missing")
	}
	old, err := retained.Read(ctx, originalVersion, raw)
	must(err)
	if string(old) != string(body) {
		t.Fatal("history overwritten")
	}
	v, err := retained.Version(ctx, originalVersion)
	must(err)
	var meta map[string]any
	must(json.Unmarshal(v.Metadata, &meta))
	if meta["official_link"] == raw || meta["api_url"] != raw {
		t.Fatal("public link and API evidence conflated")
	}
	for _, resource := range []documents.Resource{
		{URL: cfg.CittadinoInformato.APIBase() + "aggiornamenti/1?comune=cascina", Role: "original", Required: true, SourceID: source.ID, Configuration: 1, MediaType: "application/json", Bytes: []byte(`{}`)},
		{URL: "https://cittadinoinformato.it/cascina/other.pdf", Role: "attachment", Required: true, SourceID: source.ID, Configuration: 1, MediaType: "application/pdf", Bytes: []byte("synthetic")},
		{URL: "https://cittadinoinformato.it/cascina/", Role: "resource", SourceID: source.ID, Configuration: 1, MediaType: "text/html", Bytes: []byte("synthetic")},
	} {
		resources := []documents.Resource{{URL: raw, Role: "original", Required: true, SourceID: source.ID, Configuration: 1, MediaType: "application/json", Bytes: body}, resource}
		if resource.Role == "original" {
			resources = []documents.Resource{resource}
		}
		_, err := retained.Retain(ctx, documents.Acquisition{ID: "outside-" + resource.Role, SourceID: source.ID, Configuration: 1, URL: resources[0].URL, Resources: resources})
		if err != documents.ErrPolicy {
			t.Fatalf("retention accepted %s escape: %v", resource.Role, err)
		}
	}
	state, err := reg.State(ctx, source.ID)
	must(err)
	if state.CollectionEnabled || state.PublicEnabled || state.Accepted {
		t.Fatal("check changed activation/acceptance")
	}
	withoutContract := cfg
	withoutContract.CittadinoInformato = nil
	withoutContract.AccessMethod = "crawl4ai"
	if _, err := reg.AppendConfiguration(ctx, source.ID, 1, withoutContract, "fixture"); err != registry.ErrInvalid {
		t.Fatalf("platform contract removed: %v", err)
	}

	// An explicitly selected dependency remains missing until its bytes validate.
	cfg.CittadinoInformato.AttachmentPaths = []string{"/calcinaia/assets/"}
	revision, err := reg.AppendConfiguration(ctx, source.ID, 1, cfg, "fixture")
	must(err)
	const pdf = "https://cittadinoinformato.it/calcinaia/assets/synthetic.pdf"
	page = crawler.pages[raw]
	var notice cittadinoNotice
	must(json.Unmarshal(page.HTML, &notice))
	notice.Content += `<a href="` + pdf + `">Provvedimento</a>`
	page.HTML, err = json.Marshal(notice)
	must(err)
	crawler.pages[raw] = page
	crawler.pages[pdf] = Page{URL: pdf, StatusCode: 200, MediaType: "application/pdf", HTML: []byte("<html>upstream error</html>")}
	engine.validatePDF = func(_ context.Context, p Page) error {
		if !strings.HasPrefix(string(p.HTML), "%PDF-") {
			return errInvalidPDF
		}
		return nil
	}
	now = now.Add(time.Minute)
	failed := engine.Check(ctx, source.ID, revision, now)
	if failed.Complete || failed.ErrorCode != "invalid_attachment_pdf" {
		t.Fatalf("invalid dependency hidden: %#v", failed)
	}
	var incompleteID int64
	must(pool.QueryRow(ctx, "SELECT last_version_id FROM acquisition_targets WHERE source_id=$1 AND url=$2", source.ID, raw).Scan(&incompleteID))
	incomplete, err := retained.Version(ctx, incompleteID)
	must(err)
	if incomplete.Complete {
		t.Fatal("missing dependency marked complete")
	}
	for _, ref := range incomplete.Resources {
		if ref.URL == pdf && (ref.Missing != "unavailable" || ref.Hash != "") {
			t.Fatal("invalid dependency bytes retained")
		}
	}
	crawler.pages[pdf] = Page{URL: pdf, StatusCode: 200, MediaType: "application/pdf", HTML: syntheticPDF()}
	now = now.Add(time.Minute)
	recovered := engine.Check(ctx, source.ID, revision, now)
	if !recovered.Complete {
		t.Fatalf("dependency not recovered: %#v", recovered)
	}

	// A disappeared API detail stays diagnostic and respects its tracked retry.
	crawler.failures[raw] = &CrawlFailure{Code: "source_http_error", StatusCode: 404, Reachable: true}
	now = now.Add(time.Minute)
	unavailable := engine.Check(ctx, source.ID, revision, now)
	if unavailable.Complete || unavailable.ErrorCode != "source_http_error" {
		t.Fatalf("missing detail hidden: %#v", unavailable)
	}
	delete(crawler.failures, raw)
	now = now.Add(time.Minute)
	before := len(crawler.calls)
	deferred := engine.Check(ctx, source.ID, revision, now)
	if deferred.Complete {
		t.Fatal("deferred check reported complete")
	}
	for _, called := range crawler.calls[before:] {
		if called == raw {
			t.Fatal("retried disappeared detail before tracked deadline")
		}
	}
	now = now.Add(time.Hour)
	resumed := engine.Check(ctx, source.ID, revision, now)
	if !resumed.Complete {
		t.Fatalf("detail not recovered: %#v", resumed)
	}
	var unrecovered int
	must(pool.QueryRow(ctx, "SELECT count(*) FROM acquisition_unavailable_targets WHERE recovered_at IS NULL").Scan(&unrecovered))
	if unrecovered != 0 {
		t.Fatal("unavailable tracking stayed stale")
	}
}
