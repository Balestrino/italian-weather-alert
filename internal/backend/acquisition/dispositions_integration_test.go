//go:build integration

package acquisition

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func TestReviewedDispositionsGuardTargetState(t *testing.T) {
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
		reg.CreateAuthority(ctx, registry.Authority{ID: "dispo-a", Name: "A", OfficialURL: "https://source.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "dispo-c", PublisherID: "dispo-a", Platform: "fixture", URL: "https://source.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "dispo-s", AuthorityID: "dispo-a", ChannelID: "dispo-c", ProductID: "municipal", Territory: "x"}, registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/list"}, AccessMethod: "html", Attribution: "A", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true}}, "test"),
		reg.RecordPreview(ctx, "dispo-s", 1, "test", evidence),
		reg.EnableCollection(ctx, "dispo-s", 1, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	tracker := NewTrackingStore(pool)
	base := "https://source.example/notices/"
	urls := map[string]string{}
	for _, name := range []string{"renamed", "listed", "referenced", "measure", "recovered", "legacy"} {
		urls[name] = base + name
	}
	seenAt := now.Add(-2 * time.Hour)
	var all []DiscoveredDocument
	for _, url := range urls {
		all = append(all, DiscoveredDocument{URL: url})
	}
	if _, err := tracker.Remember(ctx, "dispo-s", 1, all, seenAt); err != nil {
		t.Fatal(err)
	}
	// The latest traversal still lists only "listed".
	if _, err := tracker.Remember(ctx, "dispo-s", 1, []DiscoveredDocument{{URL: urls["listed"]}}, now); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO acquisition_checks(source_id,configuration,worker_id,started_at,finished_at,reachable,content_recognized,complete,error_code,listing_count,document_count,check_state)
 VALUES('dispo-s',1,'test',$1,$1,true,false,false,'source_http_error',1,6,'collection_failure')`, now); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "UPDATE acquisition_targets SET explicit_reference=true WHERE url=$1", urls["referenced"]); err != nil {
		t.Fatal(err)
	}
	if err := tracker.SetReviewState(ctx, "dispo-s", urls["measure"], "relevant", "ongoing"); err != nil {
		t.Fatal(err)
	}
	for _, url := range urls {
		if err := tracker.RecordUnavailable(ctx, "dispo-s", 1, url, 404, now); err != nil {
			t.Fatal(err)
		}
	}
	if err := tracker.ClearUnavailable(ctx, "dispo-s", 1, urls["recovered"], now); err != nil {
		t.Fatal(err)
	}
	// A disposition written before this change keeps the acknowledgement default.
	if _, err := pool.Exec(ctx, `INSERT INTO acquisition_target_dispositions(source_id,configuration,url,decision,actor,reason,evidence,observed_last_seen_at)
 VALUES('dispo-s',1,$1,'exclude_stale_unavailable','legacy','reviewed manually','{}',$2)`, urls["legacy"], seenAt); err != nil {
		t.Fatal(err)
	}

	store := NewDispositionStore(pool)
	listed, err := store.ListUnavailable(ctx, "dispo-s")
	if err != nil {
		t.Fatal(err)
	}
	classes := map[string]string{}
	for _, item := range listed {
		classes[item.URL] = item.Classification
	}
	want := map[string]string{urls["renamed"]: "withdrawn", urls["listed"]: "blocking", urls["referenced"]: "blocking", urls["measure"]: "blocking", urls["legacy"]: "excluded"}
	if len(classes) != len(want) {
		t.Fatalf("classifications: %v", classes)
	}
	for url, class := range want {
		if classes[url] != class {
			t.Fatalf("%s classified %q, want %q", url, classes[url], class)
		}
	}
	for _, item := range listed {
		if item.URL == urls["legacy"] && (len(item.Dispositions) != 1 || item.Dispositions[0].MeasureAcknowledged) {
			t.Fatalf("legacy disposition: %+v", item.Dispositions)
		}
	}

	request := func(name, decision string, observed time.Time) DispositionRequest {
		return DispositionRequest{SourceID: "dispo-s", Configuration: 1, URL: urls[name], Decision: decision, ObservedLastSeenAt: observed, Actor: "operator", Reason: "absent from complete listings and unavailable", Evidence: json.RawMessage(`{"listing_pages": 1}`)}
	}
	reject := func(r DispositionRequest, want error) {
		t.Helper()
		if _, err := store.RecordDisposition(ctx, r); !errors.Is(err, want) {
			t.Fatalf("%s %s: got %v, want %v", r.URL, r.Decision, err, want)
		}
	}
	reject(request("renamed", DecisionExclude, seenAt.Add(time.Second)), ErrTargetChanged)
	reject(request("recovered", DecisionExclude, seenAt), ErrTargetRecovered)
	reject(request("referenced", DecisionExclude, seenAt), ErrTargetReferenced)
	reject(request("measure", DecisionExclude, seenAt), ErrMeasureAcknowledgementRequired)
	reject(request("renamed", DecisionResume, seenAt), ErrNothingToResume)
	reject(request("legacy", DecisionExclude, seenAt), ErrTargetAlreadyExcluded)
	inactive := request("renamed", DecisionExclude, seenAt)
	inactive.Configuration = 2
	reject(inactive, ErrConfigurationInactive)
	missing := request("renamed", DecisionExclude, seenAt)
	missing.URL = base + "never-seen"
	reject(missing, ErrTargetNotFound)
	invalid := request("renamed", DecisionExclude, seenAt)
	invalid.Evidence = json.RawMessage(`[]`)
	reject(invalid, ErrDispositionInvalid)

	acknowledged := request("measure", DecisionExclude, seenAt)
	acknowledged.AcknowledgeMeasures = true
	if item, err := store.RecordDisposition(ctx, acknowledged); err != nil || !item.MeasureAcknowledged {
		t.Fatalf("acknowledged exclusion: %+v %v", item, err)
	}

	// Concurrent submissions of the same reviewed exclusion record exactly one.
	var wg sync.WaitGroup
	results := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.RecordDisposition(ctx, request("renamed", DecisionExclude, seenAt))
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	succeeded := 0
	for err := range results {
		if err == nil {
			succeeded++
		} else if !errors.Is(err, ErrTargetAlreadyExcluded) {
			t.Fatalf("concurrent submission: %v", err)
		}
	}
	if succeeded != 1 {
		t.Fatalf("concurrent exclusions recorded %d decisions", succeeded)
	}
	plan, err := tracker.Plan(ctx, "dispo-s", 1, now, 30)
	if err != nil {
		t.Fatal(err)
	}
	for _, document := range plan.Documents {
		if document.URL == urls["renamed"] {
			t.Fatal("excluded target still planned")
		}
	}
	if _, err := store.RecordDisposition(ctx, request("renamed", DecisionResume, seenAt)); err != nil {
		t.Fatal(err)
	}
	plan, err = tracker.Plan(ctx, "dispo-s", 1, now, 30)
	if err != nil {
		t.Fatal(err)
	}
	resumed := false
	for _, document := range plan.Documents {
		resumed = resumed || document.URL == urls["renamed"]
	}
	if !resumed {
		t.Fatal("resumed target not planned")
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
}
