//go:build integration

package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/embedding"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"testing"
	"time"
)

func TestOCRReusePreservesProvenanceUntilDependentsExpire(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := domainTestDB(t)
	migrations := []func(context.Context, *pgxpool.Pool) error{
		registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate,
		acquisition.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate,
		linking.Migrate, embedding.Migrate, interpretation.Migrate, Migrate,
		MigrateGeography, MigrateTemporal, MigrateQuality, MigrateRetention,
	}
	for _, migrate := range migrations {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if err := MigrateRetention(ctx, pool); err != nil {
		t.Fatalf("idempotent retention migration: %v", err)
	}

	now := time.Date(2026, 1, 31, 10, 30, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://calcinaia.example/terms", Locator: "retention fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "retention-authority", Name: "Comune", OfficialURL: "https://calcinaia.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "retention-channel", PublisherID: "retention-authority", Platform: "municipal-site", URL: "https://calcinaia.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "retention-source", AuthorityID: "retention-authority", ChannelID: "retention-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://calcinaia.example", Sections: []string{"https://calcinaia.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	objects := &retentionObjects{domainObjects: &domainObjects{items: map[string][]byte{}}}
	documentsStore := documents.New(pool, objects)
	retain := func(id, path, body string) documents.Version {
		t.Helper()
		url := "https://calcinaia.example/notices/" + path
		version, err := documentsStore.Retain(ctx, documents.Acquisition{ID: id, SourceID: "retention-source", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "retention-source", Configuration: 1, MediaType: "text/html", Bytes: []byte(body)}}})
		if err != nil {
			t.Fatal(err)
		}
		return version
	}

	original := retain("original", "original", "old original")
	current := retain("current", "current", "current original")
	if _, err := pool.Exec(ctx, "UPDATE retained_versions SET first_acquired_at=$1 WHERE id=$2", now, original.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "UPDATE retained_versions SET first_acquired_at=$1 WHERE id=$2", now.AddDate(0, 2, 0), current.ID); err != nil {
		t.Fatal(err)
	}
	process := processing.New(pool)
	catalog, err := ocr.RegisterCatalog(ctx, process, "fixture", "model", now)
	if err != nil {
		t.Fatal(err)
	}
	makeRun := func(v documents.Version, key string) processing.Run {
		t.Helper()
		run, e := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: key, Stage: "ocr", Workload: "evaluation", ConfigurationVersionID: catalog.ConfigurationVersionID, DocumentVersionID: &v.ID, Subject: json.RawMessage("{}"), CreatedAt: now})
		if e != nil {
			t.Fatal(e)
		}
		return run
	}
	originalRun := makeRun(original, "original-run")
	currentRun := makeRun(current, "current-run")
	store := ocr.NewStore(pool)
	identity := ocr.ArtifactIdentity{Scope: "fixture", Model: "model", Configuration: catalog.ConfigurationVersionID, Renderer: "fixture", InputSHA256: strings.Repeat("a", 64), RequestSHA256: strings.Repeat("b", 64)}
	a, err := store.ClaimArtifact(ctx, identity, now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	text := "retained original OCR"
	hash := sha256.Sum256([]byte(text))
	page := ocr.PageResult{RunID: originalRun.ID, DocumentVersionID: original.ID, PageNumber: 1, ResourceURL: "https://calcinaia.example/notices/original", Status: "complete", MediaType: "image/png", InputSHA256: identity.InputSHA256, OutputSHA256: hex.EncodeToString(hash[:]), ExtractedText: text, ReturnedModel: "model", ProviderResponseID: "receipt", CreatedAt: now}
	if err = store.CompleteArtifact(ctx, a, page, now); err != nil {
		t.Fatal(err)
	}
	if err = store.AssociateArtifact(ctx, a.Key, page, false); err != nil {
		t.Fatal(err)
	}
	page.RunID = currentRun.ID
	page.DocumentVersionID = current.ID
	page.ResourceURL = "https://calcinaia.example/notices/current"
	if err = store.AssociateArtifact(ctx, a.Key, page, true); err != nil {
		t.Fatal(err)
	}
	attempt, err := process.StartAttempt(ctx, processing.AttemptStart{RunID: originalRun.ID, StartedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	call, err := process.StartCall(ctx, processing.CallStart{RunID: originalRun.ID, AttemptNumber: attempt.Number, Ordinal: 1, InputSHA256: identity.RequestSHA256, Provider: "fixture", RequestedModel: "model", StartedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if err = process.FinishCall(ctx, processing.CallFinish{ID: call, FinishedAt: now, State: "received", Usage: processing.Usage{Status: "unavailable"}}); err != nil {
		t.Fatal(err)
	}
	retention := NewRetention(pool, objects)
	kept, err := retention.Cleanup(ctx, now.AddDate(0, 3, 1))
	if err != nil || kept.DeletedVersions != 0 {
		t.Fatalf("shared provenance removed: %#v %v", kept, err)
	}
	mapped, ok, err := store.Page(ctx, currentRun.ID, 1)
	if err != nil || !ok || mapped.ExtractedText != text {
		t.Fatal("dependent evidence lost")
	}
	removed, err := retention.Cleanup(ctx, now.AddDate(0, 6, 0))
	if err != nil || removed.DeletedVersions != 2 {
		t.Fatalf("expired graph not removed: %#v %v", removed, err)
	}
	var remaining int
	if err = pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM ocr_artifacts)+(SELECT count(*) FROM ocr_artifact_associations)+(SELECT count(*) FROM processing_provider_calls)").Scan(&remaining); err != nil || remaining != 0 {
		t.Fatalf("expired cache/receipt retained: %d %v", remaining, err)
	}
}
