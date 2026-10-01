//go:build integration

package backoffice

import (
	"context"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publicquery"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
	"time"
)

func TestAlertsDatabase(t *testing.T) {
	ctx := context.Background()
	pool := adminTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, acquisition.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, interpretation.Migrate, domain.Migrate, domain.MigrateGeography, domain.MigrateTemporal, domain.MigrateQuality, publicquery.Migrate} {
		must(m(ctx, pool))
	}
	reg := registry.New(pool)
	now := time.Now()
	ev := registry.Evidence{URL: "https://source.example/report", Locator: "fixture", ObservedAt: now}
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "CFR", OfficialURL: "https://source.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "fixture", URL: "https://source.example"}))
	cfg := registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/notices"}, AccessMethod: "fixture", Attribution: "CFR", Provenance: &ev, Policy: registry.Policy{Evidence: &ev, CollectionPermitted: true, RetentionPermitted: true, PublicationPermitted: true}}
	must(reg.CreateSource(ctx, registry.Source{ID: "s", AuthorityID: "a", ChannelID: "c", ProductID: "criticality", Territory: "Toscana"}, cfg, "fixture"))
	must(reg.RecordPreview(ctx, "s", 1, "fixture", ev))
	must(reg.EnableCollection(ctx, "s", 1, "fixture"))
	must(acquisition.NewScheduleStore(pool).SyncEnabled(ctx, now))
	retained := documents.New(pool, adminObjects{})
	retain := func(id, metadata, body string) documents.Version {
		t.Helper()
		v, e := retained.Retain(ctx, documents.Acquisition{ID: id + body, SourceID: "s", Configuration: 1, URL: "https://source.example/" + id, Metadata: json.RawMessage(metadata), Resources: []documents.Resource{{URL: "https://source.example/" + id, Role: "original", Required: true, SourceID: "s", Configuration: 1, MediaType: "text/html", Bytes: []byte(body)}}})
		must(e)
		return v
	}
	retain("bulletin", `{"validity_expressions":["22 settembre 2026"]}`, "old")
	latest := retain("bulletin", `{"validity_expressions":["23 settembre 2026","24 settembre 2026"]}`, "new")
	retain("undated", `{}`, "unknown")
	at := time.Date(2026, 9, 23, 18, 0, 0, 0, time.UTC)
	// Acquisitions use the wall clock: place these synthetic versions before the frozen view.
	_, err := pool.Exec(ctx, "UPDATE retained_versions SET first_acquired_at=$1", at.Add(-time.Hour))
	must(err)
	result, err := operations.New(pool).Alerts(ctx, at)
	must(err)
	if result.FactsUnavailable {
		_, e := publicquery.New(pool).Search(ctx, publicquery.SearchQuery{Kind: "regional", QueryTime: publicquery.QueryTime{EvaluationTime: at, KnownAt: at}})
		t.Fatalf("consolidated query failed: %v", e)
	}
	for _, d := range result.Days {
		if len(d.Bulletins) != 1 || d.Bulletins[0].Version != latest.ID {
			t.Fatalf("latest bulletin: %+v", d)
		}
		if len(d.Regional) != 0 || len(d.Measures) != 0 {
			t.Fatal("invented facts")
		}
	}
	if len(result.UndatedBulletins) != 1 {
		t.Fatal(result)
	}
}
