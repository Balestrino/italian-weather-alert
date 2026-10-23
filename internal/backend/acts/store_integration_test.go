//go:build integration

package acts

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func TestReferenceAndResolutionHistory(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := actsTestDB(t)
	migrateAll(t, ctx, pool)
	now := time.Date(2026, 10, 23, 8, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://comune.example.it/terms", Locator: "acts fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "acts-authority", Name: "Comune", OfficialURL: "https://comune.example.it"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "acts-channel", PublisherID: "acts-authority", Platform: "municipal-site", URL: "https://comune.example.it"}),
		reg.CreateSource(ctx, registry.Source{ID: "acts-source", AuthorityID: "acts-authority", ChannelID: "acts-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://comune.example.it", Sections: []string{"https://comune.example.it/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	url := "https://comune.example.it/notices/avviso"
	version, err := documents.New(pool, &actsObjects{}).Retain(ctx, documents.Acquisition{ID: "avviso", SourceID: "acts-source", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "acts-source", Configuration: 1, MediaType: "text/html", Bytes: []byte("<p>Emessa ordinanza di chiusura.</p>")}}})
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore(pool)
	refs := Detect(notice("Emessa ordinanza di chiusura dei cimiteri. Con decreto n. 9/2026 è attivato il centro."))
	stored, err := store.RecordReferences(ctx, version.ID, refs, now)
	if err != nil || len(stored) != 2 || stored[0].Kind != "decreto" || stored[0].Number != 9 || stored[1].Kind != "ordinanza" || stored[1].Number != 0 {
		t.Fatalf("stored references: %+v %v", stored, err)
	}
	again, err := store.RecordReferences(ctx, version.ID, refs, now.Add(time.Hour))
	if err != nil || len(again) != 2 || again[0].ID != stored[0].ID {
		t.Fatalf("detection repeated: %+v %v", again, err)
	}
	reference := stored[1].ID
	if _, err = store.RecordResolution(ctx, reference, "", Resolution{Outcome: "unavailable", Reason: "albo_unavailable: HTTP 503"}, now); err != nil {
		t.Fatal(err)
	}
	later := now.Add(10 * time.Minute)
	if _, err = store.RecordResolution(ctx, reference, "acts-source", Resolution{Outcome: "unique_candidate", AlboRecordID: "660", Candidates: []Candidate{{RecordID: "660", Category: "ORDINANZE", ActNumber: 38, ActYear: 2026, RegistryYear: 2026, RegistryNumber: 660, Object: "EMERGENZA"}}}, later); err != nil {
		t.Fatal(err)
	}
	if _, err = store.RecordResolution(ctx, stored[0].ID, "", Resolution{Outcome: "ambiguous", Reason: "several_matching_albo_records", Candidates: []Candidate{{RecordID: "1"}, {RecordID: "2"}}}, now); err != nil {
		t.Fatal(err)
	}
	if _, err = store.RecordResolution(ctx, stored[0].ID, "", Resolution{Outcome: "numbered"}, now); err == nil {
		t.Fatal("numbered outcome without albo record accepted")
	}
	before, found, err := store.LatestResolution(ctx, reference, later.Add(-time.Second))
	if err != nil || !found || before.Outcome != "unavailable" {
		t.Fatalf("boundary before the retry: %+v %v", before, err)
	}
	current, found, err := store.LatestResolution(ctx, reference, later)
	if err != nil || !found || current.Outcome != "unique_candidate" || current.AlboSourceID != "acts-source" || current.AlboRecordID != "660" || len(current.Candidates) != 1 || current.Candidates[0].ActNumber != 38 {
		t.Fatalf("retry did not supersede: %+v %v", current, err)
	}
	ambiguous, _, err := store.LatestResolution(ctx, stored[0].ID, later)
	if err != nil || ambiguous.Outcome != "ambiguous" || len(ambiguous.Candidates) != 2 {
		t.Fatalf("ambiguous candidates: %+v %v", ambiguous, err)
	}
	if _, found, err = store.LatestResolution(ctx, reference, now.Add(-time.Second)); err != nil || found {
		t.Fatalf("resolution before any outcome: %v %v", found, err)
	}
}
