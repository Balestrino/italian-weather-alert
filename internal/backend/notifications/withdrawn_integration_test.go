//go:build integration

package notifications

import (
	"context"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Withdrawn unavailable targets complete checks, so they neither open a
// source-delay incident nor keep an existing one open.
func TestWithdrawnTargetsDoNotHoldSourceDelay(t *testing.T) {
	ctx := context.Background()
	pool, _ := notificationTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, acquisition.Migrate, Migrate} {
		must(migrate(ctx, pool))
	}
	_, err := pool.Exec(ctx, `INSERT INTO registry_authorities VALUES('fixture','Fixture','https://example.test');
 INSERT INTO registry_channels VALUES('fixture','fixture','Fixture','https://example.test',false);
 INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES('source','fixture','fixture','municipal','fixture');
 INSERT INTO registry_configurations(source_id,revision,body,actor) VALUES('source',1,'{}','fixture');
 UPDATE registry_sources SET latest_revision=1,active_revision=1,collection_enabled=true WHERE id='source'`)
	must(err)
	store := New(pool)
	schedule := acquisition.NewScheduleStore(pool)
	base := time.Now().UTC().Truncate(time.Microsecond).Add(-6 * time.Hour)
	must(schedule.SyncEnabled(ctx, base))
	finish := func(at time.Time, outcome acquisition.CheckOutcome) {
		t.Helper()
		if _, err := pool.Exec(ctx, "UPDATE acquisition_source_status SET next_check_at=$1 WHERE source_id='source'", at); err != nil {
			t.Fatal(err)
		}
		claim, err := schedule.ClaimDue(ctx, "test", at, time.Minute)
		must(err)
		outcome.SourceID, outcome.Configuration, outcome.StartedAt, outcome.FinishedAt = "source", 1, at, at
		must(schedule.Finish(ctx, claim, outcome))
	}
	withdrawnOnly := acquisition.CheckOutcome{Reachable: true, ContentRecognized: true, Complete: true, WithdrawnUnavailable: 1}
	open := func() int {
		t.Helper()
		r, err := store.Report(ctx)
		must(err)
		return r.OpenIncidents
	}

	at := base
	for i := 0; i < 12; i++ {
		finish(at, withdrawnOnly)
		must(store.Reconcile(ctx, at, time.Hour))
		at = at.Add(10 * time.Minute)
	}
	if open() != 0 {
		t.Fatal("withdrawn targets opened a source-delay incident")
	}

	// A blocking target delays the source; a later check with only withdrawn
	// targets left recovers it.
	for i := 0; i < 4; i++ {
		finish(at, acquisition.CheckOutcome{Reachable: true, ErrorCode: "source_http_error"})
		must(store.Reconcile(ctx, at, time.Hour))
		at = at.Add(10 * time.Minute)
	}
	if open() != 1 {
		t.Fatal("blocking target did not open the source-delay incident")
	}
	finish(at, withdrawnOnly)
	must(store.Reconcile(ctx, at, time.Hour))
	if open() != 0 {
		t.Fatal("withdrawn-only check did not recover the incident")
	}
}
