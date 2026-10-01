//go:build integration

package linking

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
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func linkingTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		t.Fatal(err)
	}
	name, password := "iwa-linking-test-"+hex.EncodeToString(token[:6]), hex.EncodeToString(token)
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
	cfg, err := pgxpool.ParseConfig("postgres://iwa@localhost/iwa?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.Host, cfg.ConnConfig.Port, cfg.ConnConfig.Password = host, uint16(portNumber), password
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
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

type integrationAdapter struct {
	response inference.Response
	requests int
}

func (a *integrationAdapter) Name() string { return "fixture" }
func (a *integrationAdapter) Complete(context.Context, inference.Request) (inference.Response, error) {
	a.requests++
	return a.response, nil
}
func i64(v int64) *int64 { return &v }

func TestBaselineSelectsOldUnresolvedMeasureAndPersistsSupportedLink(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := linkingTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if err := MigrateSemantic(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 17, 18, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	policy := registry.Evidence{URL: "https://comune.example/policy", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{reg.CreateAuthority(ctx, registry.Authority{ID: "link-authority", Name: "Link municipality", OfficialURL: "https://comune.example"}), reg.CreateChannel(ctx, registry.Channel{ID: "link-channel", PublisherID: "link-authority", Platform: "fixture", URL: "https://comune.example"}), reg.CreateSource(ctx, registry.Source{ID: "link-source", AuthorityID: "link-authority", ChannelID: "link-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://comune.example", Sections: []string{"https://comune.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &policy, CollectionPermitted: true, RetentionPermitted: true}}, "test")} {
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
	extractCatalog, err := extraction.RegisterCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	linkCatalog, err := RegisterCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	classStore, extractStore := classification.NewStore(pool), extraction.NewStore(pool)
	type added struct {
		version documents.Version
		run     int64
	}
	add := func(key, url, text, kind, subject, place string, at time.Time) added {
		t.Helper()
		version, e := retained.Retain(ctx, documents.Acquisition{ID: key, SourceID: "link-source", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "link-source", Configuration: 1, MediaType: "text/html", Bytes: []byte("<main>" + text + "</main>")}}})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = pool.Exec(ctx, "UPDATE retained_versions SET first_acquired_at=$2 WHERE id=$1", version.ID, at); e != nil {
			t.Fatal(e)
		}
		source := "link-source"
		classRun, e := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "class-" + key, Workload: "evaluation", Stage: "classification", ConfigurationVersionID: classCatalog.ConfigurationVersionID, SourceID: &source, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: at})
		if e != nil {
			t.Fatal(e)
		}
		relevant := true
		if e = classStore.Put(ctx, classification.Result{RunID: classRun.ID, DocumentVersionID: version.ID, Status: "classified", Relevant: &relevant, ReasonCode: "local_weather_measure", EvidenceQuote: "Comune", ContentSHA256: strings.Repeat("c", 64), ContentComplete: true, ProviderResponseID: "fixture", ReturnedModel: "qwen3.8-27b", CreatedAt: at}); e != nil {
			t.Fatal(e)
		}
		extractRun, e := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "extract-" + key, Workload: "evaluation", Stage: "extraction", ConfigurationVersionID: extractCatalog.ConfigurationVersionID, SourceID: &source, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: at})
		if e != nil {
			t.Fatal(e)
		}
		p := place
		measure := extraction.Measure{Ordinal: 1, Kind: kind, Subject: subject, Place: &p, IndeterminateFields: []string{"valid_from", "valid_until"}, Evidence: []extraction.Evidence{{Field: "kind", ResourceURL: url, Quote: text}, {Field: "subject", ResourceURL: url, Quote: subject}, {Field: "place", ResourceURL: url, Quote: place}}}
		if e = extractStore.Put(ctx, extraction.Result{RunID: extractRun.ID, DocumentVersionID: version.ID, ClassificationRunID: classRun.ID, Status: "extracted", ReasonCode: "measures_extracted", ContentSHA256: strings.Repeat("d", 64), ContentComplete: true, ProviderResponseID: "fixture", ReturnedModel: "qwen3.8-27b", Measures: []extraction.Measure{measure}, CreatedAt: at}); e != nil {
			t.Fatal(e)
		}
		return added{version, extractRun.ID}
	}
	old := add("old", "https://comune.example/old", "Il Comune dispone la chiusura del sottopasso di via Maremmana.", "closure", "sottopasso di via Maremmana", "via Maremmana", now.AddDate(0, 0, -60))
	textOnly := add("old-text", "https://comune.example/old-text", "Il Comune limita il transito presso il sottopasso di via Maremmana.", "restriction", "sottopasso di via Maremmana", "Maremmana", now.AddDate(0, 0, -45))
	current := add("new", "https://comune.example/new", "Il sottopasso di via Maremmana è stato riaperto alla circolazione.", "reopening", "sottopasso di via Maremmana", "via Maremmana", now)
	store := NewStore(pool)
	currentContext, err := store.Current(ctx, current.run, 1)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := store.Candidates(ctx, currentContext, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 || candidates[0].RunID != old.run || !candidates[0].ExactPlace || candidates[1].RunID != textOnly.run || candidates[1].ExactPlace || candidates[1].TextRank <= 0 || now.Sub(candidates[0].FirstAcquiredAt) < 30*24*time.Hour {
		t.Fatalf("old unresolved baseline candidate missing: %#v", candidates)
	}
	adapter := &integrationAdapter{response: inference.Response{ID: "link-response", Model: "qwen3.8-27b", Content: `{"status":"linked","reason_code":"cross_document_evidence","relation":"reopens","candidate_index":1,"current_evidence_indices":[1,2,3],"candidate_evidence_indices":[1,2,3]}`, Usage: inference.Usage{InputTokens: i64(90), OutputTokens: i64(30)}}}
	runner := &Runner{Store: store, Processing: process, Adapter: adapter, Model: "qwen3.8-27b", ConfigurationVersion: linkCatalog.ConfigurationVersionID, PriceVersion: linkCatalog.PriceVersionID, MaxCandidates: 20, Now: func() time.Time { return now }}
	queue := jobs.New(pool)
	job, err := Enqueue(ctx, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, Payload{ExtractionRunID: current.run, MeasureOrdinal: 1, Workload: "evaluation"}, now)
	if err != nil {
		t.Fatal(err)
	}
	claim, err := queue.Claim(ctx, inference.Queue, "link-worker", now, time.Minute)
	if err != nil || claim.ID != job.ID {
		t.Fatalf("claim: %#v %v", claim, err)
	}
	output, err := runner.Handler()(ctx, claim.Job)
	if err != nil {
		t.Fatal(err)
	}
	if err = queue.Complete(ctx, claim, output, now.Add(time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	var linkedRun, candidateRun int64
	var relation string
	if err = pool.QueryRow(ctx, "SELECT run_id,candidate_extraction_run_id,relation FROM linking_results WHERE current_extraction_run_id=$1", current.run).Scan(&linkedRun, &candidateRun, &relation); err != nil {
		t.Fatal(err)
	}
	var snapshot, evidenceCount int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM linking_candidates WHERE run_id=$1", linkedRun).Scan(&snapshot); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM linking_evidence WHERE run_id=$1", linkedRun).Scan(&evidenceCount); err != nil {
		t.Fatal(err)
	}
	if candidateRun != old.run || relation != "reopens" || snapshot != 2 || evidenceCount != 6 || adapter.requests != 1 {
		t.Fatalf("supported link not persisted: candidate=%d relation=%s snapshot=%d evidence=%d calls=%d", candidateRun, relation, snapshot, evidenceCount, adapter.requests)
	}
	ambiguous := add("ambiguous", "https://comune.example/ambiguous", "Il Comune comunica modifiche alla viabilità in via Maremmana senza indicare l'atto precedente.", "restriction", "viabilità in via Maremmana", "via Maremmana", now.Add(time.Minute))
	adapter.response = inference.Response{ID: "ambiguous-response", Model: "qwen3.8-27b", Content: `{"status":"unresolved","reason_code":"ambiguous_relation","relation":null,"candidate_index":null,"current_evidence_indices":[1],"candidate_evidence_indices":[]}`, Usage: inference.Usage{InputTokens: i64(70), OutputTokens: i64(20)}}
	ambiguousJob, err := Enqueue(ctx, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, Payload{ExtractionRunID: ambiguous.run, MeasureOrdinal: 1, Workload: "evaluation"}, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	ambiguousClaim, err := queue.Claim(ctx, inference.Queue, "link-worker", now.Add(time.Minute), time.Minute)
	if err != nil || ambiguousClaim.ID != ambiguousJob.ID {
		t.Fatalf("ambiguous claim: %#v %v", ambiguousClaim, err)
	}
	ambiguousOutput, err := runner.Handler()(ctx, ambiguousClaim.Job)
	if err != nil {
		t.Fatal(err)
	}
	if err = queue.Complete(ctx, ambiguousClaim, ambiguousOutput, now.Add(time.Minute+time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	var ambiguousStatus, jobState string
	var attempts int
	if err = pool.QueryRow(ctx, "SELECT status FROM linking_results WHERE current_extraction_run_id=$1", ambiguous.run).Scan(&ambiguousStatus); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "SELECT state,attempt_count FROM processing_jobs WHERE id=$1", ambiguousJob.ID).Scan(&jobState, &attempts); err != nil {
		t.Fatal(err)
	}
	if ambiguousStatus != "unresolved" || jobState != "succeeded" || attempts != 1 || adapter.requests != 2 {
		t.Fatalf("ambiguous readable content retried: status=%s job=%s attempts=%d calls=%d", ambiguousStatus, jobState, attempts, adapter.requests)
	}
}
