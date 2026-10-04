//go:build integration

package extraction

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func extractionTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		t.Fatal(err)
	}
	name, password := "iwa-extraction-test-"+hex.EncodeToString(token[:6]), hex.EncodeToString(token)
	passwordFile := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("isolated PostgreSQL operation failed: %v: %s", err, strings.ReplaceAll(string(out), password, "[redacted]"))
		}
		return strings.TrimSpace(string(out))
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	binding := listener.Addr().String() + ":5432"
	_ = listener.Close()
	run("run", "--pull=never", "--detach", "--name", name, "--publish", binding, "--mount", "type=bind,src="+passwordFile+",dst=/run/secrets/password,readonly", "--env", "POSTGRES_PASSWORD_FILE=/run/secrets/password", "--env", "POSTGRES_USER=iwa", "--env", "POSTGRES_DB=iwa", "postgres:16-alpine")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", name).Run() })
	address := run("port", name, "5432/tcp")
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatal(err)
	}
	portNumber, _ := strconv.Atoi(port)
	configuration, err := pgxpool.ParseConfig("postgres://iwa@localhost/iwa?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	configuration.ConnConfig.Host, configuration.ConnConfig.Port, configuration.ConnConfig.Password = host, uint16(portNumber), password
	pool, err := pgxpool.NewWithConfig(context.Background(), configuration)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	deadline := time.Now().Add(25 * time.Second)
	for pool.Ping(context.Background()) != nil {
		if time.Now().After(deadline) {
			t.Fatal("test PostgreSQL did not become ready")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return pool
}

type integrationObjects struct {
	mu    sync.Mutex
	items map[string][]byte
}

func (m *integrationObjects) Ensure(_ context.Context, hash string, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.items == nil {
		m.items = map[string][]byte{}
	}
	m.items[hash] = append([]byte(nil), body...)
	return nil
}
func (m *integrationObjects) Read(_ context.Context, hash string, _ int64) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]byte(nil), m.items[hash]...), nil
}

func TestExtractionPersistsEvidenceAndExposesNewerUninterpretedVersion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := extractionTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, classification.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("idempotent extraction migration: %v", err)
	}
	now := time.Date(2026, 9, 17, 17, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://comune.example/policy", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "extract-authority", Name: "Extraction municipality", OfficialURL: "https://comune.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "extract-channel", PublisherID: "extract-authority", Platform: "fixture", URL: "https://comune.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "calcinaia-municipal", AuthorityID: "extract-authority", ChannelID: "extract-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://comune.example", Sections: []string{"https://comune.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	retained := documents.New(pool, &integrationObjects{})
	process := processing.New(pool)
	classCatalog, err := classification.RegisterCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	extractCatalog, err := RegisterCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	classStore, extractStore := classification.NewStore(pool), NewStore(pool)
	queue := jobs.New(pool)
	url := "https://comune.example/notice"

	retain := func(id, text string) documents.Version {
		t.Helper()
		noise := strings.Repeat("a", 64)
		if id == "extract-equivalent" {
			noise = strings.Repeat("b", 64)
		}
		version, retainErr := retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: "calcinaia-municipal", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "calcinaia-municipal", Configuration: 1, MediaType: "text/html", Bytes: []byte("<!-- js-view-dom-id-" + noise + " --><main>" + text + "</main>")}}})
		if retainErr != nil {
			t.Fatal(retainErr)
		}
		return version
	}
	classify := func(version documents.Version, key string) int64 {
		t.Helper()
		source := "calcinaia-municipal"
		run, startErr := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "classification-integration-" + key, Workload: "evaluation", Stage: "classification", ConfigurationVersionID: classCatalog.ConfigurationVersionID, SourceID: &source, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: now})
		if startErr != nil {
			t.Fatal(startErr)
		}
		relevant := true
		if putErr := classStore.Put(ctx, classification.Result{RunID: run.ID, DocumentVersionID: version.ID, Status: "classified", Relevant: &relevant, ReasonCode: "local_weather_measure", EvidenceQuote: "Comune", ContentSHA256: strings.Repeat("c", 64), ContentComplete: true, ProviderResponseID: "fixture", ReturnedModel: "qwen3.8-27b", CreatedAt: now}); putErr != nil {
			t.Fatal(putErr)
		}
		return run.ID
	}
	runExtraction := func(version documents.Version, classRun int64, response string, key string) Result {
		t.Helper()
		adapter := &fakeAdapter{response: inference.Response{ID: "extract-" + key, Model: "qwen3.8-27b", Content: response, Usage: inference.Usage{InputTokens: int64Pointer(100), OutputTokens: int64Pointer(50)}}}
		runner := &Runner{Manifests: classStore, ReuseSources: map[string]bool{"calcinaia-municipal": true}, Documents: retained, OCR: ocr.NewStore(pool), Classifications: classStore, Processing: process, Results: extractStore, Adapter: adapter, Model: "qwen3.8-27b", ConfigurationVersion: extractCatalog.ConfigurationVersionID, PriceVersion: extractCatalog.PriceVersionID, Now: func() time.Time { return now }}
		if key == "partial-grouped" {
			runner.ReuseSources = nil
		}
		job, enqueueErr := Enqueue(ctx, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, Payload{DocumentVersionID: version.ID, ClassificationRunID: classRun, Workload: "evaluation"}, now)
		if enqueueErr != nil {
			t.Fatal(enqueueErr)
		}
		claim, claimErr := queue.Claim(ctx, inference.Queue, "extraction-worker", now, time.Minute)
		if claimErr != nil || claim.ID != job.ID {
			t.Fatalf("claim extraction job: %#v %v", claim, claimErr)
		}
		output, handleErr := runner.Handler()(ctx, claim.Job)
		if handleErr != nil {
			t.Fatal(handleErr)
		}
		if completeErr := queue.Complete(ctx, claim, output, now.Add(time.Millisecond)); completeErr != nil {
			t.Fatal(completeErr)
		}
		var runID int64
		if queryErr := pool.QueryRow(ctx, "SELECT run_id FROM extraction_results WHERE document_version_id=$1 ORDER BY run_id DESC LIMIT 1", version.ID).Scan(&runID); queryErr != nil {
			t.Fatal(queryErr)
		}
		stored, found, getErr := extractStore.Get(ctx, runID)
		if getErr != nil || !found {
			t.Fatalf("stored extraction unavailable: %v", getErr)
		}
		return stored
	}

	first := retain("extract-v1", "Il Comune dispone la chiusura del sottopasso di via Maremmana fino al perdurare dell'emergenza.")
	firstClass := classify(first, "v1")
	valid := `{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["dispone la chiusura","sottopasso di via Maremmana","via Maremmana","fino al perdurare dell'emergenza"],"measures":[{"kind":"closure","subject":"sottopasso di via Maremmana","place":"via Maremmana","valid_from":null,"valid_until":"fino al perdurare dell'emergenza","evidence_refs":{"kind":[0],"subject":[1],"place":[2],"valid_from":[],"valid_until":[3]}}]}`
	stored := runExtraction(first, firstClass, valid, "v1")
	if stored.Status != "extracted" || len(stored.Measures) != 1 || len(stored.Measures[0].Evidence) != 4 || len(stored.Measures[0].TemporalCandidates) != 1 || stored.Measures[0].TemporalCandidates[0].OriginalExpression != "fino al perdurare dell'emergenza" || len(stored.Measures[0].TemporalCandidates[0].Evidence) != 1 || len(stored.Segments) != 1 || stored.Measures[0].Evidence[0].SegmentOrdinal != 1 || stored.Segments[0].CoreStartByte < stored.Segments[0].StartByte || stored.Segments[0].CoreEndByte > stored.Segments[0].EndByte {
		t.Fatalf("persisted measure/evidence mismatch: %#v", stored)
	}

	equivalent := retain("extract-equivalent", "Il Comune dispone la chiusura del sottopasso di via Maremmana fino al perdurare dell'emergenza.")
	equivalentClass := classify(equivalent, "equivalent")
	// Invalid provider output would fail if reuse fell through to the adapter.
	mapped := runExtraction(equivalent, equivalentClass, "not-json", "equivalent")
	if mapped.Status != "extracted" || mapped.DocumentVersionID != equivalent.ID || mapped.InputTokens != nil || len(mapped.Measures) != 1 || mapped.Measures[0].ValidUntil == nil || *mapped.Measures[0].ValidUntil != *stored.Measures[0].ValidUntil {
		t.Fatalf("equivalent extraction lost evidence: %#v", mapped)
	}
	var origin int64
	if err = pool.QueryRow(ctx, "SELECT original_run_id FROM interpretation_reuse WHERE run_id=$1", mapped.RunID).Scan(&origin); err != nil || origin != stored.RunID {
		t.Fatalf("reuse origin missing: %d %v", origin, err)
	}
	conflictText := "ORDINA la chiusura del sottopasso dal 10 settembre. Il dispositivo ordina la chiusura del sottopasso dal 10 agosto. Restano chiusi i parchi comunali fino a revoca."
	conflictVersion := retain("extract-conflict", conflictText)
	conflictClass := classify(conflictVersion, "conflict")
	conflictContent := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: url, Role: "original", Text: conflictText}}}
	september, august, until := "10 settembre", "10 agosto", "fino a revoca"
	conflictMeasures, mergeErr := Merge([][]Measure{{{
		Kind: "closure", Subject: "sottopasso", ValidFrom: &september,
		Evidence: []Evidence{{Field: "kind", ResourceURL: url, Quote: "ORDINA la chiusura", SegmentOrdinal: 1}, {Field: "subject", ResourceURL: url, Quote: "sottopasso", SegmentOrdinal: 1}, {Field: "valid_from", ResourceURL: url, Quote: september, SegmentOrdinal: 1}},
	}}, {{
		Kind: "closure", Subject: "sottopasso", ValidFrom: &august,
		Evidence: []Evidence{{Field: "kind", ResourceURL: url, Quote: "ordina la chiusura", SegmentOrdinal: 2}, {Field: "subject", ResourceURL: url, Quote: "sottopasso", SegmentOrdinal: 2}, {Field: "valid_from", ResourceURL: url, Quote: august, SegmentOrdinal: 2}},
	}, {
		Kind: "restriction", Subject: "parchi comunali", ValidUntil: &until,
		Evidence: []Evidence{{Field: "kind", ResourceURL: url, Quote: "Restano chiusi", SegmentOrdinal: 2}, {Field: "subject", ResourceURL: url, Quote: "parchi comunali", SegmentOrdinal: 2}, {Field: "valid_until", ResourceURL: url, Quote: until, SegmentOrdinal: 2}},
	}}}, conflictContent, 2)
	if mergeErr != nil {
		t.Fatal(mergeErr)
	}
	source := "calcinaia-municipal"
	conflictRun, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "extraction-integration-conflict", Workload: "evaluation", Stage: "extraction", ConfigurationVersionID: extractCatalog.ConfigurationVersionID, SourceID: &source, DocumentVersionID: &conflictVersion.ID, Subject: json.RawMessage(`{"fixture":"cross-window-conflict"}`), CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	segments := []SegmentResult{
		{Ordinal: 1, Total: 2, DocumentVersionID: conflictVersion.ID, ResourceURL: url, Role: "original", StartByte: 0, EndByte: len(conflictText), CoreStartByte: 0, CoreEndByte: len(conflictText), ContentSHA256: strings.Repeat("1", 64), ResponseSHA256: strings.Repeat("2", 64), ProviderResponseID: "conflict-window-1", ReturnedModel: "qwen3.8-27b", MeasureCount: 1},
		{Ordinal: 2, Total: 2, DocumentVersionID: conflictVersion.ID, ResourceURL: url, Role: "original", StartByte: 0, EndByte: len(conflictText), CoreStartByte: 0, CoreEndByte: len(conflictText), ContentSHA256: strings.Repeat("3", 64), ResponseSHA256: strings.Repeat("4", 64), ProviderResponseID: "conflict-window-2", ReturnedModel: "qwen3.8-27b", MeasureCount: 2},
	}
	if err = extractStore.Put(ctx, Result{RunID: conflictRun.ID, DocumentVersionID: conflictVersion.ID, ClassificationRunID: conflictClass, Status: "extracted", ReasonCode: "measures_extracted", ContentSHA256: strings.Repeat("5", 64), ContentComplete: true, ProviderResponseID: "segmented:conflict", ReturnedModel: "qwen3.8-27b", Measures: conflictMeasures, Segments: segments, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	persistedConflict, found, err := extractStore.Get(ctx, conflictRun.ID)
	if err != nil || !found || len(persistedConflict.Measures) != 2 || persistedConflict.Measures[0].ValidFrom != nil || len(persistedConflict.Measures[0].TemporalCandidates) != 2 || persistedConflict.Measures[0].TemporalCandidates[0].ConflictIdentity == nil || persistedConflict.Measures[0].TemporalCandidates[1].ConflictIdentity == nil || *persistedConflict.Measures[0].TemporalCandidates[0].ConflictIdentity != *persistedConflict.Measures[0].TemporalCandidates[1].ConflictIdentity || persistedConflict.Measures[1].ValidUntil == nil || *persistedConflict.Measures[1].ValidUntil != until {
		t.Fatalf("append-only conflict candidates did not round trip: %#v %v", persistedConflict, err)
	}
	if _, err = pool.Exec(ctx, "UPDATE extracted_temporal_candidates SET original_expression='10 luglio' WHERE run_id=$1", conflictRun.ID); err == nil {
		t.Fatal("append-only temporal candidate accepted an update")
	}

	second := retain("extract-v2", "Il Comune pubblica un aggiornamento meteo, ma il provvedimento allegato non è leggibile.")
	secondClass := classify(second, "v2")
	unsupported := `{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["è stato riaperto","sottopasso di via Maremmana","via Maremmana"],"measures":[{"kind":"reopening","subject":"sottopasso di via Maremmana","place":"via Maremmana","valid_from":null,"valid_until":null,"evidence_refs":{"kind":[0],"subject":[1],"place":[2],"valid_from":[],"valid_until":[]}}]}`
	failed := runExtraction(second, secondClass, unsupported, "v2")
	if failed.Status != "uninterpreted" || failed.ReasonCode != "extraction_evidence_invalid" {
		t.Fatalf("newer failed extraction not explicit: %#v", failed)
	}
	visibility, found, err := extractStore.LatestVisibility(ctx, first.DocumentID)
	if err != nil || !found {
		t.Fatalf("latest visibility unavailable: %v", err)
	}
	if visibility.DocumentVersionID != second.ID || visibility.Status != "uninterpreted" || visibility.OfficialURL != url {
		t.Fatalf("older extraction hid newer uninterpreted document: %#v", visibility)
	}
	third := retain("extract-v3", "Il Comune ha acquisito un nuovo aggiornamento ancora in attesa di interpretazione.")
	visibility, found, err = extractStore.LatestVisibility(ctx, first.DocumentID)
	if err != nil || !found || visibility.DocumentVersionID != third.ID || visibility.RunID != 0 || visibility.Status != "uninterpreted" || visibility.ReasonCode != "not_processed" || visibility.InterpretedAt != nil {
		t.Fatalf("unprocessed acquired revision was hidden: %#v %v", visibility, err)
	}

	text, partialResponse := partialReopeningFixture()
	partial := retain("extract-partial-reopening", text)
	partialStored := runExtraction(partial, classify(partial, "partial"), partialResponse, "partial")
	if partialStored.Status != "extracted" || len(partialStored.Measures) != 7 {
		t.Fatalf("partial reopening did not survive durable extraction: %#v", partialStored)
	}
	for _, m := range partialStored.Measures {
		if m.Kind == "prohibition" && (m.Place == nil || !strings.Contains(text, *m.Place)) {
			t.Fatalf("stored literal place lost: %#v", m)
		}
		if m.Kind == "reopening" && (m.ValidFrom != nil || m.ValidUntil != nil || len(m.TemporalCandidates) != 0) {
			t.Fatalf("stored reopening inherited old closure time: %#v", m)
		}
		for _, e := range m.Evidence {
			if e.ResourceURL != url || e.SegmentOrdinal != 1 || !strings.Contains(text, e.Quote) {
				t.Fatalf("stored partial-update evidence lost: %#v", e)
			}
		}
	}
	var grouped map[string]any
	if err = json.Unmarshal([]byte(partialResponse), &grouped); err != nil {
		t.Fatal(err)
	}
	firstProhibition := grouped["measures"].([]any)[0].(map[string]any)
	firstProhibition["subject"], firstProhibition["place"] = "attività nei parchi", nil
	remaining := []any{firstProhibition}
	for _, candidate := range grouped["measures"].([]any)[1:] {
		if candidate.(map[string]any)["kind"] != "prohibition" {
			remaining = append(remaining, candidate)
		}
	}
	grouped["measures"] = remaining
	body, err := json.Marshal(grouped)
	if err != nil {
		t.Fatal(err)
	}
	groupedStored := runExtraction(partial, classify(partial, "partial-grouped"), string(body), "partial-grouped")
	if groupedStored.Status != "extracted" || len(groupedStored.Measures) != 7 {
		t.Fatal("grouped provider response lost independent persisted restrictions", groupedStored)
	}
	places := map[string]bool{}
	for _, m := range groupedStored.Measures {
		if m.Kind == "prohibition" {
			if m.Subject != "attività" || m.Place == nil || places[*m.Place] {
				t.Fatal("grouped scope survived or place duplicated", m)
			}
			places[*m.Place] = true
		}
	}
	if len(places) != 3 || !places["ciclopiste in riva d’Arno"] {
		t.Fatal("prohibition field evidence did not round trip", places)
	}
	oldPartial, foundPartial, err := extractStore.Get(ctx, partialStored.RunID)
	beforeJSON, _ := json.Marshal(partialStored)
	afterJSON, _ := json.Marshal(oldPartial)
	if err != nil || !foundPartial || string(beforeJSON) != string(afterJSON) {
		t.Fatal("selected reevaluation rewrote prior extraction", err)
	}
	old, found, err := extractStore.Get(ctx, stored.RunID)
	if err != nil || !found || old.Measures[0].Kind != "closure" || *old.Measures[0].ValidUntil != *stored.Measures[0].ValidUntil {
		t.Fatalf("new reopening rewrote earlier extraction: %#v %v", old, err)
	}
}
