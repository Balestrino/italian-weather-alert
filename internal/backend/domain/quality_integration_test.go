//go:build integration

package domain

import (
	"context"
	"encoding/json"
	"strings"
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

func TestIndependentQualityDimensionsAndPrecedingStateWarning(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := domainTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, acquisition.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, Migrate, MigrateTemporal, MigrateQuality} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if err := MigrateQuality(ctx, pool); err != nil {
		t.Fatalf("idempotent quality migration: %v", err)
	}
	now := time.Date(2026, 9, 17, 15, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://www.calcinaia.example/note-legali", Locator: "official municipal terms", ObservedAt: now}
	configuration := registry.Configuration{URL: "https://www.calcinaia.example", Sections: []string{"https://www.calcinaia.example/notices"}, AccessMethod: "fixture", Attribution: "Comune di Calcinaia", Provenance: &evidence, Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "quality-authority", Name: "Comune di Calcinaia", OfficialURL: "https://www.calcinaia.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "quality-channel", PublisherID: "quality-authority", Platform: "municipal-site", URL: "https://www.calcinaia.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "quality-source", AuthorityID: "quality-authority", ChannelID: "quality-channel", ProductID: "municipal", Territory: "050004"}, configuration, "test"),
		reg.RecordPreview(ctx, "quality-source", 1, "test", registry.Evidence{URL: configuration.Sections[0], Locator: "successful fixture preview", ObservedAt: now}),
		reg.EnableCollection(ctx, "quality-source", 1, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}

	documentStore := documents.New(pool, &domainObjects{})
	documentURL := "https://www.calcinaia.example/notices/closure"
	retain := func(id, body string) documents.Version {
		t.Helper()
		version, err := documentStore.Retain(ctx, documents.Acquisition{ID: id, SourceID: "quality-source", Configuration: 1, URL: documentURL, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: documentURL, Role: "original", Required: true, SourceID: "quality-source", Configuration: 1, MediaType: "text/html", Bytes: []byte(body)}}})
		if err != nil {
			t.Fatal(err)
		}
		return version
	}
	preceding := retain("quality-preceding", "chiusura fino a nuova comunicazione")
	// Keep acquisition chronology observable even on very fast hosts.
	time.Sleep(time.Millisecond)
	newer := retain("quality-newer", "nuova comunicazione non interpretabile")

	store := New(pool)
	measure := LocalMeasure{ID: "quality-closure", DocumentVersionID: preceding.ID, SourceID: "quality-source", MunicipalityISTAT: "050004", Kind: "closure", Subject: "sottopasso", Place: stringPointer("via Maremmana")}
	if err := store.PutLocalMeasure(ctx, measure); err != nil {
		t.Fatal(err)
	}
	condition := "until a later revocation is acquired and interpreted"
	if _, err := store.PutTemporal(ctx, TemporalValue{EntityKind: "local_measure", EntityID: measure.ID, Meaning: "validity", Original: "fino a nuova comunicazione", Precision: "conditional", Condition: &condition, EvidenceDocumentVersionID: preceding.ID, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordProvenance(ctx, ProvenanceAssessment{SourceID: "quality-source", Configuration: 1, State: "verified", EvidenceURL: evidence.URL, EvidenceLocator: evidence.Locator, Limitations: []string{"verified source does not imply complete updates"}, AssessedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordInterpretation(ctx, InterpretationAssessment{MeasureID: measure.ID, State: "supported", EvidenceDocumentVersionID: preceding.ID, Reason: "reviewed retained passage", Limitations: []string{"cessation remains conditional"}, Actor: "evaluation", RecordedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordPrecedingStateWarning(ctx, measure.ID, newer.ID, "newer document may supersede the closure but has no interpretation", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	schedule := acquisition.NewScheduleStore(pool)
	if err := schedule.SyncEnabled(ctx, now); err != nil {
		t.Fatal(err)
	}
	first, err := schedule.ClaimDue(ctx, "quality-worker", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	firstFinished := now.Add(time.Second)
	if err = schedule.Finish(ctx, first, acquisition.CheckOutcome{SourceID: first.SourceID, Configuration: first.Configuration, StartedAt: first.StartedAt, FinishedAt: firstFinished, Reachable: true, ContentRecognized: true, Complete: true, Documents: 1}); err != nil {
		t.Fatal(err)
	}
	secondStarted := firstFinished.Add(10 * time.Minute)
	second, err := schedule.ClaimDue(ctx, "quality-worker", secondStarted, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	failedAt := secondStarted.Add(time.Second)
	if err = schedule.Finish(ctx, second, acquisition.CheckOutcome{SourceID: second.SourceID, Configuration: second.Configuration, StartedAt: second.StartedAt, FinishedAt: failedAt, ErrorCode: "source_unreachable"}); err != nil {
		t.Fatal(err)
	}

	view, err := store.MeasureView(ctx, measure.ID, failedAt.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if view.State.Status != "undetermined" || view.Quality.Provenance.State != "verified" || view.Quality.Interpretation.State != "supported" || view.Quality.Updating.State != "failed" {
		t.Fatalf("quality dimensions were collapsed after source failure: %#v", view)
	}
	if len(view.Warnings) != 1 || view.Warnings[0].DocumentVersionID != newer.ID || view.Warnings[0].OfficialURL != documentURL || view.Warnings[0].Status != "uninterpreted" || view.Warnings[0].Reason != "not_processed" {
		t.Fatalf("newer uninterpreted document was hidden: %#v", view.Warnings)
	}
	if !containsLimitation(view.Quality.Interpretation.Limitations, "newer potentially superseding document is not interpreted") {
		t.Fatalf("preceding state has no warning: %#v", view.Quality.Interpretation)
	}

	if err = store.MarkInterpretationsUnreliable(ctx, []string{measure.ID}, "confirmed place extraction defect", "operator", failedAt.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	unreliable, err := store.MeasureView(ctx, measure.ID, failedAt.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if unreliable.Quality.Interpretation.State != "unreliable" || unreliable.Quality.Provenance.State != "verified" || unreliable.Quality.Updating.State != "failed" || unreliable.State.Status != "undetermined" || len(unreliable.Warnings) != 1 {
		t.Fatalf("confirmed defect affected unrelated dimensions or hid evidence: %#v", unreliable)
	}
	if !containsLimitation(unreliable.Quality.Interpretation.Limitations, "confirmed interpretation defect") {
		t.Fatalf("unreliable result omitted defect reason: %#v", unreliable.Quality.Interpretation)
	}
}

func containsLimitation(values []string, fragment string) bool {
	for _, value := range values {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}
