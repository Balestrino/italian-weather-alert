//go:build integration

package acquisition

import (
	"context"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

// Without interpretation tables the refresh must leave tracked states as they
// are; the 099 migration accepts aged_unresolved and keeps existing values.
func TestInterpretationStateMigrationAndMissingTables(t *testing.T) {
	ctx := context.Background()
	pool := acquisitionTestDB(t)
	if err := registry.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	reg := registry.New(pool)
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	evidence := registry.Evidence{URL: "https://source.example/reuse", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "state-a", Name: "A", OfficialURL: "https://source.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "state-c", PublisherID: "state-a", Platform: "fixture", URL: "https://source.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "state-s", AuthorityID: "state-a", ChannelID: "state-c", ProductID: "municipal", Territory: "x"}, registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/list"}, AccessMethod: "html", Attribution: "A", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	tracker := NewTrackingStore(pool)
	raw := "https://source.example/notices/coc"
	if _, err := tracker.Remember(ctx, "state-s", 1, []DiscoveredDocument{{URL: raw}}, now); err != nil {
		t.Fatal(err)
	}
	if err := tracker.SetReviewState(ctx, "state-s", raw, "relevant", "ongoing"); err != nil {
		t.Fatal(err)
	}
	if err := tracker.RefreshInterpretationState(ctx, "state-s", now, 90); err != nil {
		t.Fatal(err)
	}
	var measure, relevance string
	var derivedAt *time.Time
	if err := pool.QueryRow(ctx, "SELECT measure_state,relevance_state,interpretation_state_at FROM acquisition_targets WHERE url=$1", raw).Scan(&measure, &relevance, &derivedAt); err != nil {
		t.Fatal(err)
	}
	if measure != "ongoing" || relevance != "relevant" || derivedAt != nil {
		t.Fatalf("refresh changed states without interpretation tables: %s %s %v", measure, relevance, derivedAt)
	}
	if err := tracker.RefreshInterpretationState(ctx, "state-s", now, 0); err == nil {
		t.Fatal("invalid limit accepted")
	}
	if _, err := pool.Exec(ctx, "UPDATE acquisition_targets SET measure_state='aged_unresolved' WHERE url=$1", raw); err != nil {
		t.Fatalf("aged_unresolved rejected: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE acquisition_targets SET measure_state='expired' WHERE url=$1", raw); err == nil {
		t.Fatal("unknown measure state accepted")
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "SELECT measure_state FROM acquisition_targets WHERE url=$1", raw).Scan(&measure); err != nil || measure != "aged_unresolved" {
		t.Fatalf("repeated migration changed the stored state: %s %v", measure, err)
	}
}
