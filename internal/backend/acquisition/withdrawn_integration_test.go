//go:build integration

package acquisition

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Replays the renamed-notice pattern through tracking, the engine and the
// schedule: a URL that leaves the listings and returns 404 no longer fails the
// source, a listed 404 still does, and a later 200 recovers the target.
func TestWithdrawnTargetCompletesScheduledChecks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := acquisitionTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	base := time.Date(2026, 10, 21, 6, 0, 0, 0, time.UTC)
	section := calcinaia + "/tipi-di-notizia/notizie"
	renamed := calcinaia + "/novita/volontari-azione-il-mc-donalds"
	corrected := calcinaia + "/novita/volontari-azione-con-mc-donalds"
	listedMissing := calcinaia + "/novita/ordinanza-ancora-elencata"
	evidence := registry.Evidence{URL: calcinaia + "/note-legali", Locator: "fixture", ObservedAt: base}
	configuration := registry.Configuration{URL: calcinaia, Sections: []string{section}, AccessMethod: "crawl4ai", Attribution: "Comune", CheckSeconds: 3600, DelaySeconds: 10800, BackoffBaseSeconds: 60, BackoffMaxSeconds: 3600,
		Policy:    registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true},
		Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, PaginationParameter: "page", MaxPagesPerSection: 2, MaxDocuments: 20, ListingContentMarkers: []string{"configured"}, ListingItemClass: "card-body", ListingDateClass: "data", ListingDateLayout: "02 Jan 2006", ListingDateLocale: "it", BootstrapDays: 30}}
	reg := registry.New(pool)
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "withdrawn-a", Name: "Comune", OfficialURL: calcinaia}),
		reg.CreateChannel(ctx, registry.Channel{ID: "withdrawn-c", PublisherID: "withdrawn-a", Platform: "fixture", URL: calcinaia}),
		reg.CreateSource(ctx, registry.Source{ID: "withdrawn-s", AuthorityID: "withdrawn-a", ChannelID: "withdrawn-c", ProductID: "municipal", Territory: "050004"}, configuration, "test"),
		reg.RecordPreview(ctx, "withdrawn-s", 1, "test", evidence),
		reg.EnableCollection(ctx, "withdrawn-s", 1, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	listing := func(paths ...string) Page {
		html := `<main id="configured">`
		for _, path := range paths {
			html += `<div class="card-body"><span class="data">20 Ott 2026</span><a href="` + path + `">Notizia</a></div>`
		}
		return Page{URL: section, StatusCode: 200, HTML: []byte(html + `</main>`)}
	}
	crawler := &multiUnavailableCrawler{fixtureCrawler: fixtureCrawler{pages: map[string]Page{
		section:       listing("/novita/volontari-azione-il-mc-donalds", "/novita/ordinanza-ancora-elencata"),
		renamed:       {URL: renamed, StatusCode: 200, HTML: []byte("prima versione")},
		corrected:     {URL: corrected, StatusCode: 200, HTML: []byte("titolo corretto")},
		listedMissing: {URL: listedMissing, StatusCode: 200, HTML: []byte("ordinanza")},
	}}, failed: map[string]int{}}
	retained := &stableRetention{}
	tracker := NewTrackingStore(pool)
	schedule := NewScheduleStore(pool)
	if err := schedule.SyncEnabled(ctx, base); err != nil {
		t.Fatal(err)
	}
	check := func(at time.Time) StoredCheck {
		t.Helper()
		claim, err := schedule.ClaimDue(ctx, "worker", at, time.Minute)
		if errors.Is(err, ErrNoDueCheck) {
			if _, err = pool.Exec(ctx, "UPDATE acquisition_source_status SET next_check_at=$1 WHERE source_id='withdrawn-s'", at); err != nil {
				t.Fatal(err)
			}
			claim, err = schedule.ClaimDue(ctx, "worker", at, time.Minute)
		}
		if err != nil {
			t.Fatal(err)
		}
		engine := &Engine{Registry: reg, Retained: retained, Crawler: crawler, Tracking: tracker, Now: func() time.Time { return at }}
		if err := schedule.Finish(ctx, claim, engine.Check(ctx, "withdrawn-s", 1, claim.StartedAt)); err != nil {
			t.Fatal(err)
		}
		checks, err := schedule.Checks(ctx, "withdrawn-s")
		if err != nil {
			t.Fatal(err)
		}
		return checks[len(checks)-1]
	}

	if first := check(base); !first.Complete || first.WithdrawnUnavailable != 0 {
		t.Fatalf("bootstrap check: %+v", first)
	}

	// The Comune corrects the slug: the old URL leaves the listing and 404s.
	crawler.pages[section] = listing("/novita/volontari-azione-con-mc-donalds", "/novita/ordinanza-ancora-elencata")
	crawler.failed[renamed] = 404
	renamedAt := base.Add(time.Hour)
	second := check(renamedAt)
	if !second.Complete || second.ErrorCode != "" || second.WithdrawnUnavailable != 1 {
		t.Fatalf("withdrawn target failed the check: %+v", second)
	}
	status, err := schedule.Status(ctx, "withdrawn-s", renamedAt)
	if err != nil || status.ConsecutiveFailures != 0 || status.LastCompleteAt == nil || !status.LastCompleteAt.Equal(renamedAt) {
		t.Fatalf("status after withdrawn check: %+v %v", status, err)
	}
	var failures int
	if err := pool.QueryRow(ctx, "SELECT consecutive_failures FROM acquisition_unavailable_targets WHERE url=$1 AND recovered_at IS NULL", renamed).Scan(&failures); err != nil || failures != 1 {
		t.Fatalf("withdrawn failure history not retained: %d %v", failures, err)
	}

	// A still-listed 404 keeps blocking; the withdrawn target is still counted.
	crawler.failed[listedMissing] = 404
	mixed := check(renamedAt.Add(2 * time.Hour))
	if mixed.Complete || mixed.ErrorCode != "source_http_error" || mixed.WithdrawnUnavailable != 1 {
		t.Fatalf("mixed check: %+v", mixed)
	}

	// Both URLs come back: the withdrawn target recovers with a new version.
	delete(crawler.failed, listedMissing)
	delete(crawler.failed, renamed)
	crawler.pages[renamed] = Page{URL: renamed, StatusCode: 200, HTML: []byte("seconda versione")}
	recovered := check(renamedAt.Add(4 * time.Hour))
	if !recovered.Complete || recovered.WithdrawnUnavailable != 0 {
		t.Fatalf("recovery check: %+v", recovered)
	}
	var recoveredAt *time.Time
	if err := pool.QueryRow(ctx, "SELECT recovered_at FROM acquisition_unavailable_targets WHERE url=$1", renamed).Scan(&recoveredAt); err != nil || recoveredAt == nil {
		t.Fatalf("withdrawn target not recovered: %v", err)
	}
	var versions int
	for _, item := range retained.items {
		if item.URL == renamed {
			versions++
		}
	}
	if versions != 2 {
		t.Fatalf("recovered target versions: %d", versions)
	}
}
