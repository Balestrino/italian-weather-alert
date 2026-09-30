//go:build integration

package ocr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestArtifactOwnershipAndCurrentVersionEvidence(t *testing.T) {
	ctx := context.Background()
	pool := ocrTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, Migrate, Migrate} {
		if err := m(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://comune.example/policy", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "ocr-authority", Name: "OCR municipality", OfficialURL: "https://comune.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "ocr-channel", PublisherID: "ocr-authority", Platform: "fixture", URL: "https://comune.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "ocr-source", AuthorityID: "ocr-authority", ChannelID: "ocr-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://comune.example", Sections: []string{"https://comune.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	retained := documents.New(pool, &integrationObjects{})
	originalURL, missingURL := "https://comune.example/scan.png", "https://comune.example/missing.pdf"
	version, err := retained.Retain(ctx, documents.Acquisition{ID: "ocr-fixture", SourceID: "ocr-source", Configuration: 1, URL: originalURL, Metadata: json.RawMessage(`{"fixture":"CAL-OCR-SCAN"}`), Resources: []documents.Resource{
		{URL: originalURL, Role: "original", Required: true, SourceID: "ocr-source", Configuration: 1, MediaType: "image/png", Bytes: []byte("synthetic scan")},
		{URL: missingURL, Role: "attachment", Required: true, SourceID: "ocr-source", Configuration: 1, Missing: "unavailable"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	process := processing.New(pool)
	catalog, err := RegisterCatalog(ctx, process, "openai-chat", "deepseek-ocr-2", now)
	if err != nil {
		t.Fatal(err)
	}

	store := NewStore(pool)
	identity := ArtifactIdentity{Scope: "provider", Model: "deepseek-ocr-2", Configuration: catalog.ConfigurationVersionID, Renderer: "poppler-v1", InputSHA256: strings.Repeat("a", 64), RequestSHA256: strings.Repeat("b", 64)}
	var wg sync.WaitGroup
	owned := make(chan Artifact, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, e := store.ClaimArtifact(ctx, identity, now, time.Minute)
			if e != nil {
				t.Error(e)
			} else if a.Owned {
				owned <- a
			}
		}()
	}
	wg.Wait()
	close(owned)
	var first Artifact
	count := 0
	for a := range owned {
		first = a
		count++
	}
	if count != 1 {
		t.Fatalf("owners=%d", count)
	}
	replacement, err := store.ClaimArtifact(ctx, identity, now.Add(time.Minute), time.Minute)
	if err != nil || !replacement.Owned || replacement.Generation != first.Generation+1 {
		t.Fatalf("replacement: %#v %v", replacement, err)
	}
	hash := sha256.Sum256([]byte("retained OCR text"))
	page := PageResult{Status: "complete", MediaType: "image/png", InputSHA256: identity.InputSHA256, OutputSHA256: hex.EncodeToString(hash[:]), ExtractedText: "retained OCR text", ProviderResponseID: "receipt", ReturnedModel: identity.Model, InputTokens: pointer(100), OutputTokens: pointer(20)}
	if err = store.CompleteArtifact(ctx, first, page, now.Add(time.Minute)); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale owner published: %v", err)
	}
	if err = store.RenewArtifact(ctx, replacement, now.Add(70*time.Second), time.Minute); err != nil {
		t.Fatal(err)
	}
	if err = store.CompleteArtifact(ctx, replacement, page, now.Add(80*time.Second)); err != nil {
		t.Fatal(err)
	}
	cached, err := store.ClaimArtifact(ctx, identity, now.Add(90*time.Second), time.Minute)
	if err != nil || !cached.Ready || cached.Owned || cached.Page.ExtractedText != page.ExtractedText {
		t.Fatalf("cache miss: %#v %v", cached, err)
	}
	for _, field := range []string{"model", "configuration", "renderer", "request", "image"} {
		changed := identity
		switch field {
		case "model":
			changed.Model = "different"
		case "configuration":
			changed.Configuration = "different"
		case "renderer":
			changed.Renderer = "different"
		case "request":
			changed.RequestSHA256 = strings.Repeat("c", 64)
		case "image":
			changed.InputSHA256 = strings.Repeat("d", 64)
		}
		a, e := store.ClaimArtifact(ctx, changed, now, time.Minute)
		if e != nil || !a.Owned || a.Key == cached.Key {
			t.Fatalf("%s did not invalidate: %v", field, e)
		}
	}
	run, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "artifact-map", Workload: "evaluation", Stage: "ocr", ConfigurationVersionID: catalog.ConfigurationVersionID, DocumentVersionID: &version.ID, Subject: json.RawMessage("{}"), CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	association := PageResult{RunID: run.ID, DocumentVersionID: version.ID, ResourceURL: originalURL, PageNumber: 2, CreatedAt: now.Add(90 * time.Second)}
	for i := 0; i < 2; i++ {
		if err = store.AssociateArtifact(ctx, cached.Key, association, true); err != nil {
			t.Fatal(err)
		}
	}
	mapped, found, err := store.Page(ctx, run.ID, 2)
	if err != nil || !found || mapped.DocumentVersionID != version.ID || mapped.ResourceURL != originalURL || mapped.PageNumber != 2 || mapped.InputTokens != nil || mapped.OutputTokens != nil || mapped.ExtractedText != page.ExtractedText {
		t.Fatalf("invalid provenance/usage: %#v %v", mapped, err)
	}
	association.DocumentVersionID++
	if err = store.AssociateArtifact(ctx, cached.Key, association, true); !errors.Is(err, ErrConflict) {
		t.Fatalf("mismatched version accepted: %v", err)
	}
}
