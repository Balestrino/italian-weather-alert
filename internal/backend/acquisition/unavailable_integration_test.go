//go:build integration

package acquisition

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"testing"
	"time"
)

func TestUnavailableTargetBackoffPersistenceAndRecovery(t *testing.T) {
	ctx := context.Background()
	pool := acquisitionTestDB(t)
	if err := registry.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	reg := registry.New(pool)
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	evidence := registry.Evidence{URL: "https://source.example/reuse", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "track-a", Name: "A", OfficialURL: "https://source.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "track-c", PublisherID: "track-a", Platform: "fixture", URL: "https://source.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "track-s", AuthorityID: "track-a", ChannelID: "track-c", ProductID: "municipal", Territory: "x"}, registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/list"}, AccessMethod: "html", Attribution: "A", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}

	tracker := NewTrackingStore(pool)
	raw := "https://source.example/missing"
	if _, err := tracker.Remember(ctx, "track-s", 1, []DiscoveredDocument{{URL: raw}}, now); err != nil {
		t.Fatal(err)
	}
	for i, want := range []time.Duration{time.Hour, 2 * time.Hour, 4 * time.Hour, 8 * time.Hour, 16 * time.Hour, 24 * time.Hour, 24 * time.Hour} {
		at := now.Add(time.Duration(i) * 24 * time.Hour)
		if err := tracker.RecordUnavailable(ctx, "track-s", 1, raw, 404, at); err != nil {
			t.Fatal(err)
		}
		plan, err := tracker.Plan(ctx, "track-s", 1, at.Add(time.Minute), 30)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.Documents) != 1 || plan.Documents[0].UnavailableStatus != 404 || plan.Documents[0].NextAttemptAt == nil || !plan.Documents[0].NextAttemptAt.Equal(at.Add(want)) {
			t.Fatalf("failure not retained in plan: %+v", plan)
		}
	}
	recovered := now.Add(8 * 24 * time.Hour)
	if err := tracker.ClearUnavailable(ctx, "track-s", 1, raw, recovered); err != nil {
		t.Fatal(err)
	}
	plan, err := tracker.Plan(ctx, "track-s", 1, recovered, 30)
	if err != nil || len(plan.Documents) != 1 || plan.Documents[0].NextAttemptAt != nil || plan.Documents[0].UnavailableStatus != 0 {
		t.Fatalf("recovery: %+v %v", plan, err)
	}
	var previousStatus int
	if err := pool.QueryRow(ctx, "SELECT http_status FROM acquisition_unavailable_targets WHERE source_id='track-s' AND recovered_at IS NOT NULL").Scan(&previousStatus); err != nil || previousStatus != 404 {
		t.Fatalf("recovery erased evidence: %v", err)
	}
	if err := tracker.RecordUnavailable(ctx, "track-s", 1, raw, 410, recovered); err != nil {
		t.Fatal(err)
	}
	plan, err = tracker.Plan(ctx, "track-s", 1, recovered, 30)
	if err != nil || plan.Documents[0].UnavailableStatus != 410 || !plan.Documents[0].NextAttemptAt.Equal(recovered.Add(time.Hour)) {
		t.Fatalf("new failure backoff not reset: %+v %v", plan, err)
	}
	other := "https://source.example/available"
	if _, err := tracker.Remember(ctx, "track-s", 1, []DiscoveredDocument{{URL: other}}, recovered); err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO acquisition_target_dispositions
 (source_id,configuration,url,decision,actor,reason,evidence,observed_last_seen_at)
 SELECT t.source_id,t.configuration,t.url,'exclude_stale_unavailable','test-review',
 'disappeared from verified listings and remains unavailable','{"check_id": 1}'::jsonb,t.last_seen_at
 FROM acquisition_targets t WHERE t.source_id='track-s' AND t.url=$1`, raw)
	if err != nil {
		t.Fatal(err)
	}
	plan, err = tracker.Plan(ctx, "track-s", 1, recovered, 30)
	if err != nil || len(plan.Documents) != 1 || plan.Documents[0].URL != other {
		t.Fatalf("disposition did not exclude only its URL: %+v %v", plan, err)
	}
	var retainedStatus int
	if err := pool.QueryRow(ctx, "SELECT http_status FROM acquisition_unavailable_targets WHERE source_id='track-s' AND recovered_at IS NULL").Scan(&retainedStatus); err != nil || retainedStatus != 410 {
		t.Fatalf("disposition changed the original failure: status=%d err=%v", retainedStatus, err)
	}
	if _, err := tracker.Remember(ctx, "track-s", 1, []DiscoveredDocument{{URL: raw}}, recovered.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	plan, err = tracker.Plan(ctx, "track-s", 1, recovered.Add(time.Minute), 30)
	if err != nil || len(plan.Documents) != 2 || plan.Documents[1].URL != raw || plan.Documents[1].UnavailableStatus != 410 {
		t.Fatalf("rediscovered URL was not reactivated: %+v %v", plan, err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
}
