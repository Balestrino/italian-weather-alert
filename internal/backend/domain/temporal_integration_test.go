//go:build integration

package domain

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMeasureScopedCalcinaiaUpdatesAndTemporalPrecision(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := domainTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, Migrate, MigrateTemporal} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if err := MigrateTemporal(ctx, pool); err != nil {
		t.Fatalf("idempotent temporal migration: %v", err)
	}
	now := time.Date(2026, 9, 17, 14, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://www.calcinaia.example/note-legali", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "calcinaia-time-authority", Name: "Comune di Calcinaia", OfficialURL: "https://www.calcinaia.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "calcinaia-time-channel", PublisherID: "calcinaia-time-authority", Platform: "municipal-site", URL: "https://www.calcinaia.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "calcinaia-time-source", AuthorityID: "calcinaia-time-authority", ChannelID: "calcinaia-time-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://www.calcinaia.example", Sections: []string{"https://www.calcinaia.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}

	documentStore := documents.New(pool, &domainObjects{})
	retain := func(id, path, body string) documents.Version {
		t.Helper()
		url := "https://www.calcinaia.example/" + path
		version, err := documentStore.Retain(ctx, documents.Acquisition{ID: id, SourceID: "calcinaia-time-source", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "calcinaia-time-source", Configuration: 1, MediaType: "text/html", Bytes: []byte(body)}}})
		if err != nil {
			t.Fatal(err)
		}
		return version
	}
	closureVersion := retain("calcinaia-closure", "closure", "chiusura sottopasso e parchi fino al perdurare dell'emergenza")
	reopeningVersion := retain("calcinaia-reopening", "reopening", "il sottopasso di via Maremmana è stato riaperto; restano le altre restrizioni; pagina aggiornata 09:36")
	conflictVersion := retain("calcinaia-date-conflict", "date-conflict", "pubblicato 9 settembre; riepilogo 10 settembre; testo e piattaforma 10 agosto")

	process := processing.New(pool)
	classCatalog, err := classification.RegisterCatalog(ctx, process, "fixture-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	extractCatalog, err := extraction.RegisterCatalog(ctx, process, "fixture-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	linkCatalog, err := linking.RegisterCatalog(ctx, process, "fixture-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	classStore, extractStore := classification.NewStore(pool), extraction.NewStore(pool)
	sourceID := "calcinaia-time-source"
	extract := func(key string, version documents.Version, measures []extraction.Measure) int64 {
		t.Helper()
		classRun, runErr := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "task-5.3-class-" + key, Workload: "evaluation", Stage: "classification", ConfigurationVersionID: classCatalog.ConfigurationVersionID, SourceID: &sourceID, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: now})
		if runErr != nil {
			t.Fatal(runErr)
		}
		relevant := true
		if runErr = classStore.Put(ctx, classification.Result{RunID: classRun.ID, DocumentVersionID: version.ID, Status: "classified", Relevant: &relevant, ReasonCode: "local_weather_measure", EvidenceQuote: "fixture", ContentSHA256: strings.Repeat("c", 64), ContentComplete: true, ProviderResponseID: "fixture", ReturnedModel: "qwen3.8-27b", CreatedAt: now}); runErr != nil {
			t.Fatal(runErr)
		}
		extractionRun, runErr := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "task-5.3-extract-" + key, Workload: "evaluation", Stage: "extraction", ConfigurationVersionID: extractCatalog.ConfigurationVersionID, SourceID: &sourceID, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: now})
		if runErr != nil {
			t.Fatal(runErr)
		}
		if runErr = extractStore.Put(ctx, extraction.Result{RunID: extractionRun.ID, DocumentVersionID: version.ID, ClassificationRunID: classRun.ID, Status: "extracted", ReasonCode: "measures_extracted", ContentSHA256: strings.Repeat("d", 64), ContentComplete: true, ProviderResponseID: "fixture", ReturnedModel: "qwen3.8-27b", Measures: measures, CreatedAt: now}); runErr != nil {
			t.Fatal(runErr)
		}
		return extractionRun.ID
	}
	closureFrom := "dalle ore 18:00 del 20 agosto 2026"
	conditionalEnd := "fino al perdurare dell'emergenza"
	closureRun := extract("closure", closureVersion, []extraction.Measure{
		{Ordinal: 1, Kind: "closure", Subject: "sottopasso", Place: stringPointer("via Maremmana"), ValidFrom: &closureFrom, ValidUntil: &conditionalEnd, IndeterminateFields: []string{}, Evidence: []extraction.Evidence{{Field: "subject", ResourceURL: "https://www.calcinaia.example/closure", Quote: "chiusura sottopasso"}}},
		{Ordinal: 2, Kind: "restriction", Subject: "accesso ai parchi", Place: stringPointer("parchi comunali"), ValidFrom: &closureFrom, ValidUntil: &conditionalEnd, IndeterminateFields: []string{}, Evidence: []extraction.Evidence{{Field: "subject", ResourceURL: "https://www.calcinaia.example/closure", Quote: "parchi"}}},
	})
	reopeningRun := extract("reopening", reopeningVersion, []extraction.Measure{{Ordinal: 1, Kind: "reopening", Subject: "sottopasso", Place: stringPointer("via Maremmana"), IndeterminateFields: []string{"valid_from", "valid_until"}, Evidence: []extraction.Evidence{{Field: "subject", ResourceURL: "https://www.calcinaia.example/reopening", Quote: "è stato riaperto"}}}})

	store := New(pool)
	measures := []LocalMeasure{
		{ID: "maremmana-closed", DocumentVersionID: closureVersion.ID, SourceID: sourceID, MunicipalityISTAT: "050004", Kind: "closure", Subject: "sottopasso", Place: stringPointer("via Maremmana")},
		{ID: "parks-restricted", DocumentVersionID: closureVersion.ID, SourceID: sourceID, MunicipalityISTAT: "050004", Kind: "restriction", Subject: "accesso ai parchi", Place: stringPointer("parchi comunali")},
		{ID: "maremmana-reopened", DocumentVersionID: reopeningVersion.ID, SourceID: sourceID, MunicipalityISTAT: "050004", Kind: "reopening", Subject: "sottopasso", Place: stringPointer("via Maremmana")},
		{ID: "date-conflict-measure", DocumentVersionID: conflictVersion.ID, SourceID: sourceID, MunicipalityISTAT: "050004", Kind: "restriction", Subject: "misura con date discordanti"},
	}
	for _, measure := range measures {
		if err = store.PutLocalMeasure(ctx, measure); err != nil {
			t.Fatal(err)
		}
	}
	for _, binding := range []struct {
		id      string
		run     int64
		ordinal int
	}{{"maremmana-closed", closureRun, 1}, {"parks-restricted", closureRun, 2}, {"maremmana-reopened", reopeningRun, 1}} {
		if err = store.BindExtraction(ctx, binding.id, binding.run, binding.ordinal); err != nil {
			t.Fatal(err)
		}
	}

	linkRun, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "task-5.3-link-reopening", Workload: "evaluation", Stage: "linking", ConfigurationVersionID: linkCatalog.ConfigurationVersionID, SourceID: &sourceID, DocumentVersionID: &reopeningVersion.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	relation := "reopens"
	candidateRun, candidateOrdinal := closureRun, 1
	if err = linking.NewStore(pool).Put(ctx, linking.Result{RunID: linkRun.ID, CurrentRunID: reopeningRun, CurrentOrdinal: 1, Status: "linked", ReasonCode: "partial_reopening", Relation: &relation, CandidateRunID: &candidateRun, CandidateOrdinal: &candidateOrdinal, ProviderResponseID: "fixture", ReturnedModel: "qwen3.8-27b", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	update, err := store.ApplyLinkedUpdate(ctx, linkRun.ID, now)
	if err != nil || update.TargetMeasureID != "maremmana-closed" || update.UpdateMeasureID != "maremmana-reopened" || update.Relation != "reopens" {
		t.Fatalf("measure-scoped update failed: %#v %v", update, err)
	}
	closedState, err := store.MeasureState(ctx, "maremmana-closed", now.Add(time.Minute))
	parksState, parksErr := store.MeasureState(ctx, "parks-restricted", now.Add(time.Minute))
	if err != nil || parksErr != nil || closedState.Status != "superseded" || len(closedState.UpdatedBy) != 1 || closedState.UpdatedBy[0] != "maremmana-reopened" || parksState.Status == "superseded" || len(parksState.UpdatedBy) != 0 {
		t.Fatalf("partial reopening changed another measure: closure=%#v parks=%#v errors=%v/%v", closedState, parksState, err, parksErr)
	}

	start, err := EuropeRomeInstant(closureFrom, 2026, time.August, 20, 18, 0, "time stated by a Calcinaia municipal act")
	if err != nil {
		t.Fatal(err)
	}
	start.EntityKind, start.EntityID, start.Meaning, start.EvidenceDocumentVersionID, start.CreatedAt = "local_measure", "maremmana-closed", "validity", closureVersion.ID, now
	condition := "cessation of the emergency is established by later evidence"
	conditional := TemporalValue{EntityKind: "local_measure", EntityID: "parks-restricted", Meaning: "validity", Original: conditionalEnd, Precision: "conditional", Condition: &condition, EvidenceDocumentVersionID: closureVersion.ID, CreatedAt: now}
	for _, temporal := range []TemporalValue{start, conditional} {
		if _, err = store.PutTemporal(ctx, temporal); err != nil {
			t.Fatal(err)
		}
	}

	reopeningUnknown := TemporalValue{EntityKind: "local_measure", EntityID: "maremmana-reopened", Meaning: "event", Original: "il sottopasso è stato riaperto", Precision: "unknown", EvidenceDocumentVersionID: reopeningVersion.ID, CreatedAt: now}
	pageUpdate, err := EuropeRomeInstant("Ultimo aggiornamento: 21 agosto 2026 ore 09:36", 2026, time.August, 21, 9, 36, "timestamp displayed by the Calcinaia municipal page")
	if err != nil {
		t.Fatal(err)
	}
	pageUpdate.EntityKind, pageUpdate.EntityID, pageUpdate.Meaning, pageUpdate.EvidenceDocumentVersionID, pageUpdate.CreatedAt = "local_measure", "maremmana-reopened", "source_modification", reopeningVersion.ID, now
	for _, temporal := range []TemporalValue{reopeningUnknown, pageUpdate} {
		if _, err = store.PutTemporal(ctx, temporal); err != nil {
			t.Fatal(err)
		}
	}
	reopeningTimes, err := store.TemporalValues(ctx, "local_measure", "maremmana-reopened")
	if err != nil || len(reopeningTimes) != 2 || reopeningTimes[0].Meaning != "event" || reopeningTimes[0].Instant != nil || reopeningTimes[1].Meaning != "source_modification" || reopeningTimes[1].Instant == nil || reopeningTimes[1].Instant.Format(time.RFC3339) != "2026-08-21T07:36:00Z" {
		t.Fatalf("page time became an invented reopening time: %#v %v", reopeningTimes, err)
	}

	publicationDate := time.Date(2026, time.September, 9, 0, 0, 0, 0, time.UTC)
	textDate := time.Date(2026, time.September, 10, 0, 0, 0, 0, time.UTC)
	platformDate := time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC)
	conflictGroup := "measure-start-date"
	conflictValues := []TemporalValue{
		{EntityKind: "local_measure", EntityID: "date-conflict-measure", Meaning: "publication", Original: "9 settembre 2026", Precision: "date", Date: &publicationDate, EvidenceDocumentVersionID: conflictVersion.ID, CreatedAt: now},
		{EntityKind: "local_measure", EntityID: "date-conflict-measure", Meaning: "event", Original: "10 settembre", Precision: "date", Date: &textDate, ConflictGroup: &conflictGroup, EvidenceDocumentVersionID: conflictVersion.ID, CreatedAt: now},
		{EntityKind: "local_measure", EntityID: "date-conflict-measure", Meaning: "platform", Original: "10/08/2026", Precision: "date", Date: &platformDate, ConflictGroup: &conflictGroup, EvidenceDocumentVersionID: conflictVersion.ID, CreatedAt: now},
		{EntityKind: "local_measure", EntityID: "date-conflict-measure", Meaning: "validity", Original: "conflicting start dates", Precision: "unknown", ConflictGroup: &conflictGroup, EvidenceDocumentVersionID: conflictVersion.ID, CreatedAt: now},
	}
	for _, temporal := range conflictValues {
		if _, err = store.PutTemporal(ctx, temporal); err != nil {
			t.Fatal(err)
		}
	}
	storedConflict, err := store.TemporalValues(ctx, "local_measure", "date-conflict-measure")
	if err != nil || len(storedConflict) != 4 || storedConflict[0].Date == nil || storedConflict[0].Instant != nil || storedConflict[3].Precision != "unknown" || storedConflict[3].Instant != nil {
		t.Fatalf("date-only/source-platform conflict was collapsed: %#v %v", storedConflict, err)
	}
}
