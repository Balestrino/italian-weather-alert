//go:build integration

package acquisition

import (
	"context"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"testing"
	"time"
)

type ownershipCrawler struct {
	run func(context.Context) (Page, error)
}

func (c ownershipCrawler) Crawl(ctx context.Context, _ string) (Page, error) { return c.run(ctx) }

func TestAcquisitionLeaseLossCancelsCrawl(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool := acquisitionTestDB(t)
	if err := registry.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	reg := registry.New(pool)
	base := time.Now().UTC()
	evidence := registry.Evidence{URL: "https://source.example/reuse", Locator: "fixture", ObservedAt: base}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "A", OfficialURL: "https://source.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "fixture", URL: "https://source.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "s", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "x"}, registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/list"}, AccessMethod: "crawl4ai", Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, PaginationParameter: "page", MaxPagesPerSection: 2, MaxDocuments: 20, ListingContentMarkers: []string{"listing"}}, Attribution: "A", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
		reg.RecordPreview(ctx, "s", 1, "test", evidence), reg.EnableCollection(ctx, "s", 1, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}

	crawler := ownershipCrawler{run: func(owned context.Context) (Page, error) {
		if _, err := pool.Exec(ctx, "UPDATE acquisition_source_status SET lease_expires_at=now()-interval '1 second' WHERE source_id='s'"); err != nil {
			return Page{}, err
		}
		select {
		case <-owned.Done():
			return Page{}, owned.Err()
		case <-time.After(4 * time.Second):
			t.Error("crawl outlived lost lease")
			return Page{}, nil
		}
	}}
	worker := CheckWorker{Store: NewScheduleStore(pool), Engine: &Engine{Registry: reg, Retained: &fakeRetention{}, Crawler: crawler}, ID: "owner", Lease: 3 * time.Second, PollInterval: time.Second, Schedule: func(context.Context, RetainedPage, time.Time) error {
		t.Error("scheduled after ownership loss")
		return nil
	}}
	started := time.Now()
	if _, err := worker.RunOne(ctx); !errors.Is(err, ErrStaleCheck) {
		t.Fatalf("expected lost ownership: %v", err)
	}
	if time.Since(started) > 3*time.Second {
		t.Fatal("crawl cancellation was not prompt")
	}
	var checks int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM acquisition_checks").Scan(&checks); err != nil {
		t.Fatal(err)
	}
	if checks != 0 {
		t.Fatal("stale owner persisted completed check")
	}
}
