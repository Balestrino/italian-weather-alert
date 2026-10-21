//go:build integration

package domain

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// A recent open-ended municipal measure keeps an unlisted unavailable target
// blocking once its maintained state is refreshed.
func TestDerivedMeasureProtectsUnlistedTarget(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := domainTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, acquisition.Migrate, Migrate, MigrateTemporal} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://www.calcinaia.example/note-legali", Locator: "official municipal terms", ObservedAt: now}
	configuration := registry.Configuration{URL: "https://www.calcinaia.example", Sections: []string{"https://www.calcinaia.example/notices"}, AccessMethod: "fixture", Attribution: "Comune di Calcinaia", Provenance: &evidence, Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "withdrawn-authority", Name: "Comune di Calcinaia", OfficialURL: "https://www.calcinaia.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "withdrawn-channel", PublisherID: "withdrawn-authority", Platform: "municipal-site", URL: "https://www.calcinaia.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "withdrawn-source", AuthorityID: "withdrawn-authority", ChannelID: "withdrawn-channel", ProductID: "municipal", Territory: "050004"}, configuration, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	tracker := acquisition.NewTrackingStore(pool)
	documentStore := documents.New(pool, &domainObjects{})
	published := now.Add(-24 * time.Hour)
	withMeasure := "https://www.calcinaia.example/notices/coc"
	withoutMeasure := "https://www.calcinaia.example/notices/renamed"
	if _, err := tracker.Remember(ctx, "withdrawn-source", 1, []acquisition.DiscoveredDocument{{URL: withMeasure, PublicationDate: &published}, {URL: withoutMeasure, PublicationDate: &published}}, now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	retain := func(id, url string) documents.Version {
		t.Helper()
		version, err := documentStore.Retain(ctx, documents.Acquisition{ID: id, SourceID: "withdrawn-source", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "withdrawn-source", Configuration: 1, MediaType: "text/html", Bytes: []byte(id)}}})
		if err != nil {
			t.Fatal(err)
		}
		return version
	}
	measured := retain("withdrawn-coc", withMeasure)
	retain("withdrawn-renamed", withoutMeasure)
	if err := New(pool).PutLocalMeasure(ctx, LocalMeasure{ID: "withdrawn-coc-activation", DocumentVersionID: measured.ID, SourceID: "withdrawn-source", MunicipalityISTAT: "050004", Kind: "activation", Subject: "COC"}); err != nil {
		t.Fatal(err)
	}
	// The maintained state is refreshed before planning, as in a scheduled check.
	if err := tracker.RefreshInterpretationState(ctx, "withdrawn-source", now, 90); err != nil {
		t.Fatal(err)
	}
	plan, err := tracker.Plan(ctx, "withdrawn-source", 1, now, 30)
	if err != nil {
		t.Fatal(err)
	}
	protected := map[string]bool{}
	for _, document := range plan.Documents {
		if document.Listed {
			t.Fatalf("target unexpectedly listed: %+v", document)
		}
		protected[document.URL] = document.Protected
	}
	if len(protected) != 2 || !protected[withMeasure] || protected[withoutMeasure] {
		t.Fatalf("derived measure protection: %v", protected)
	}
}
