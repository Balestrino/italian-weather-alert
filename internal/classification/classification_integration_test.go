//go:build integration

package classification

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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

func classificationTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		t.Fatal(err)
	}
	name, password := "iwa-classification-test-"+hex.EncodeToString(token[:6]), hex.EncodeToString(token)
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

func TestClassificationJobPersistsVersionedDecision(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := classificationTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("idempotent classification migration: %v", err)
	}
	now := time.Date(2026, 9, 17, 15, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://comune.example/policy", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "class-authority", Name: "Classification municipality", OfficialURL: "https://comune.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "class-channel", PublisherID: "class-authority", Platform: "fixture", URL: "https://comune.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "calcinaia-municipal", AuthorityID: "class-authority", ChannelID: "class-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://comune.example", Sections: []string{"https://comune.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	retained := documents.New(pool, &integrationObjects{})
	url := "https://comune.example/notice"
	version, err := retained.Retain(ctx, documents.Acquisition{ID: "classification-fixture", SourceID: "calcinaia-municipal", Configuration: 1, URL: url, Metadata: json.RawMessage(`{"source_category":""}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "calcinaia-municipal", Configuration: 1, MediaType: "text/html", Bytes: []byte(`<!-- js-view-dom-id-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --><main><h1>Allerta meteo</h1><p>Per rischio idraulico il Comune dispone la chiusura del sottopasso dalle ore 18.</p></main>`)}}})
	if err != nil {
		t.Fatal(err)
	}
	process := processing.New(pool)
	catalog, err := RegisterCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	adapter := &fakeAdapter{response: inference.Response{ID: "qwen-integration", Model: "qwen3.8-27b", Content: `{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"dispone la chiusura del sottopasso"}`, Usage: inference.Usage{InputTokens: int64Pointer(120), OutputTokens: int64Pointer(25)}}}
	runner := &Runner{Manifests: NewStore(pool), ReuseSources: map[string]bool{"calcinaia-municipal": true}, Documents: retained, OCR: ocr.NewStore(pool), Processing: process, Results: NewStore(pool), Adapter: adapter, Model: "qwen3.8-27b", ConfigurationVersion: catalog.ConfigurationVersionID, PriceVersion: catalog.PriceVersionID, Now: func() time.Time { return now }}
	queue := jobs.New(pool)
	job, err := Enqueue(ctx, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, Payload{DocumentVersionID: version.ID, Workload: "evaluation"}, now)
	if err != nil {
		t.Fatal(err)
	}
	claim, err := queue.Claim(ctx, inference.Queue, "classification-worker", now, time.Minute)
	if err != nil || claim.ID != job.ID {
		t.Fatalf("claim classification job: %#v %v", claim, err)
	}
	output, err := runner.Handler()(ctx, claim.Job)
	if err != nil {
		t.Fatal(err)
	}
	if err = queue.Complete(ctx, claim, output, now.Add(time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	var relevant, complete bool
	var reason, quote string
	if err = pool.QueryRow(ctx, "SELECT relevant,content_complete,reason_code,evidence_quote FROM classification_results WHERE document_version_id=$1", version.ID).Scan(&relevant, &complete, &reason, &quote); err != nil {
		t.Fatal(err)
	}
	if !relevant || !complete || reason != "local_weather_measure" || quote != "dispone la chiusura del sottopasso" || len(adapter.requests) != 1 {
		t.Fatalf("unexpected persisted classification: relevant=%v complete=%v reason=%s quote=%s calls=%d", relevant, complete, reason, quote, len(adapter.requests))
	}
	var segments int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM classification_segments WHERE document_version_id=$1", version.ID).Scan(&segments); err != nil || segments != 1 {
		t.Fatalf("classification segment provenance not retained: count=%d err=%v", segments, err)
	}
	content, err := GatherContent(ctx, retained, ocr.NewStore(pool), version)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := BuildManifest(ctx, retained, ocr.NewStore(pool), version, content, catalog.ConfigurationVersionID)
	if err != nil {
		t.Fatal(err)
	}
	var runID int64
	if err = pool.QueryRow(ctx, "SELECT run_id FROM classification_results WHERE document_version_id=$1", version.ID).Scan(&runID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err = NewStore(pool).PutManifest(ctx, runID, manifest); err != nil {
			t.Fatal(err)
		}
	}
	var raw, meaningful string
	if err = pool.QueryRow(ctx, "SELECT v.content_hash,m.manifest_sha256 FROM interpretation_input_manifests m JOIN processing_runs r ON r.id=m.run_id JOIN retained_versions v ON v.id=r.document_version_id WHERE m.run_id=$1", runID).Scan(&raw, &meaningful); err != nil {
		t.Fatal(err)
	}
	if raw != version.Hash || meaningful != manifest.Hash {
		t.Fatal("raw provenance changed or manifest missing")
	}

	originalBody, err := retained.Read(ctx, version.ID, url)
	if err != nil {
		t.Fatal(err)
	}
	secondBody := strings.Replace(string(originalBody), strings.Repeat("a", 64), strings.Repeat("b", 64), 1)
	second, err := retained.Retain(ctx, documents.Acquisition{ID: "equivalent-classification", SourceID: "calcinaia-municipal", Configuration: 1, URL: url, Metadata: json.RawMessage(`{"source_category":""}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "calcinaia-municipal", Configuration: 1, MediaType: "text/html", Bytes: []byte(secondBody)}}})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID == version.ID {
		t.Fatal("raw versions collapsed")
	}
	_, err = Enqueue(ctx, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, Payload{DocumentVersionID: second.ID, Workload: "evaluation"}, now)
	if err != nil {
		t.Fatal(err)
	}
	next, err := queue.Claim(ctx, inference.Queue, "classification-worker", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	reused, err := runner.Handler()(ctx, next.Job)
	if err != nil {
		t.Fatal(err)
	}
	if err = queue.Complete(ctx, next, reused, now); err != nil {
		t.Fatal(err)
	}
	if len(adapter.requests) != 1 {
		t.Fatal("equivalent classification called provider")
	}
	var linked int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM interpretation_reuse u JOIN classification_results r ON r.run_id=u.run_id WHERE r.document_version_id=$1 AND r.input_tokens IS NULL", second.ID).Scan(&linked); err != nil || linked != 1 {
		t.Fatalf("current-version mapping missing: %d %v", linked, err)
	}

}
