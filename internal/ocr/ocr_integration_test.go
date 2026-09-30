//go:build integration

package ocr

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
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

func ocrTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		t.Fatal(err)
	}
	name, password := "iwa-ocr-test-"+hex.EncodeToString(token[:6]), hex.EncodeToString(token)
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

func TestOCRJobPersistsPageAndMissingEvidence(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := ocrTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("idempotent OCR migration: %v", err)
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
	adapter := &fakeAdapter{responses: []inference.Response{{ID: "provider-1", Model: "deepseek-ocr-2", Content: "ORDINA: area cani, cimiteri; dal 20 agosto 2026 ore 18.00 fino al perdurare dell'emergenza", Usage: inference.Usage{InputTokens: pointer(100), OutputTokens: pointer(50)}}}}
	results := NewStore(pool)
	runner := &Runner{Documents: retained, Processing: process, Results: results, Adapter: adapter, Renderer: fakeRenderer{}, Model: "deepseek-ocr-2", ConfigurationVersion: catalog.ConfigurationVersionID, PriceVersion: catalog.PriceVersionID, Now: func() time.Time { return now }}
	queue := jobs.New(pool)
	policy := inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}
	for index, resourceURL := range []string{originalURL, missingURL} {
		job, enqueueErr := Enqueue(ctx, queue, policy, Payload{DocumentVersionID: version.ID, ResourceURL: resourceURL, Workload: "evaluation"}, now.Add(time.Duration(index)*time.Second))
		if enqueueErr != nil {
			t.Fatal(enqueueErr)
		}
		claim, claimErr := queue.Claim(ctx, inference.Queue, "ocr-worker", now.Add(time.Duration(index)*time.Second), time.Minute)
		if claimErr != nil || claim.ID != job.ID {
			t.Fatalf("claim OCR job: %#v %v", claim, claimErr)
		}
		output, handleErr := runner.Handler()(ctx, claim.Job)
		if handleErr != nil {
			t.Fatal(handleErr)
		}
		if err = queue.Complete(ctx, claim, output, now.Add(time.Duration(index)*time.Second+time.Millisecond)); err != nil {
			t.Fatal(err)
		}
	}
	var complete, missing, pages int
	if err = pool.QueryRow(ctx, "SELECT count(*) FILTER (WHERE status='complete'),count(*) FILTER (WHERE status='missing') FROM ocr_resource_results").Scan(&complete, &missing); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM ocr_page_results WHERE document_version_id=$1 AND resource_url=$2 AND page_number=1 AND extracted_text LIKE '%20 agosto%'", version.ID, originalURL).Scan(&pages); err != nil {
		t.Fatal(err)
	}
	if complete != 1 || missing != 1 || pages != 1 || len(adapter.requests) != 1 {
		t.Fatalf("unexpected durable OCR evidence: complete=%d missing=%d pages=%d calls=%d", complete, missing, pages, len(adapter.requests))
	}
	// Seed from retained original bytes without a provider adapter. The existing
	// historical page remains the provenance owner and is never re-billed.
	options := SeedOptions{Limit: 1, Scope: "fixture", Model: "deepseek-ocr-2", Configuration: catalog.ConfigurationVersionID, RendererIdentity: "fixture-v1"}
	report, err := results.SeedHistorical(ctx, retained, fakeRenderer{}, options)
	if err != nil || report.Seeded != 1 || report.Examined != 1 {
		t.Fatalf("seed failed: %#v %v", report, err)
	}
	again, err := results.SeedHistorical(ctx, retained, fakeRenderer{}, options)
	if err != nil || again.Seeded != 0 || again.Existing != 1 || len(adapter.requests) != 1 {
		t.Fatalf("seed not idempotent/offline: %#v %v", again, err)
	}
	options.AfterRun = report.NextRun
	tail, err := results.SeedHistorical(ctx, retained, fakeRenderer{}, options)
	if err != nil || tail.Examined != 0 {
		t.Fatalf("cursor did not advance: %#v %v", tail, err)
	}
	// Retain an incompatible historical answer for the exact image/configuration.
	original, _, err := results.Page(ctx, report.NextRun, 1)
	if err != nil {
		t.Fatal(err)
	}
	conflictingRun, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "seed-conflict", Workload: "evaluation", Stage: "ocr", ConfigurationVersionID: catalog.ConfigurationVersionID, DocumentVersionID: &version.ID, Subject: json.RawMessage("{}"), CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	original.RunID = conflictingRun.ID
	original.ExtractedText = "conflicting historical output"
	original.OutputSHA256 = digest([]byte(original.ExtractedText))
	if err = results.PutPage(ctx, original); err != nil {
		t.Fatal(err)
	}
	if err = results.PutResource(ctx, ResourceResult{RunID: conflictingRun.ID, DocumentVersionID: version.ID, ResourceURL: originalURL, Status: "complete", PageCount: 1, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	options.AfterRun = 0
	options.RendererIdentity = "unseeded-environment"
	conflict, err := results.SeedHistorical(ctx, retained, fakeRenderer{}, options)
	if err != nil || conflict.Seeded != 0 || conflict.Reasons["conflicting_historical_outputs"] != 1 {
		t.Fatalf("ambiguous history seeded: %#v %v", conflict, err)
	}
	// Full pg_dump/psql restoration includes the new cache tables and provenance.
	port := strconv.Itoa(int(pool.Config().ConnConfig.Port))
	containerBytes, err := exec.CommandContext(ctx, "docker", "ps", "--filter", "publish="+port, "--format", "{{.ID}}").Output()
	if err != nil {
		t.Fatal(err)
	}
	container := strings.TrimSpace(string(containerBytes))
	if container == "" || strings.Contains(container, "\n") {
		t.Fatal("isolated database container ambiguous")
	}
	dump, err := exec.CommandContext(ctx, "docker", "exec", container, "pg_dump", "-U", "iwa", "-d", "iwa", "--no-owner", "--no-acl").Output()
	if err != nil {
		t.Fatal("fixture dump failed")
	}
	if err = exec.CommandContext(ctx, "docker", "exec", container, "createdb", "-U", "iwa", "restored").Run(); err != nil {
		t.Fatal(err)
	}
	restore := exec.CommandContext(ctx, "docker", "exec", "-i", container, "psql", "-U", "iwa", "-d", "restored", "-v", "ON_ERROR_STOP=1", "-q")
	restore.Stdin = strings.NewReader(string(dump))
	if err = restore.Run(); err != nil {
		t.Fatal("fixture restore failed")
	}
	restored, err := exec.CommandContext(ctx, "docker", "exec", container, "psql", "-U", "iwa", "-d", "restored", "-Atc", "SELECT count(*) FROM ocr_artifacts a JOIN retained_versions v ON v.id=(a.result->>'DocumentVersionID')::bigint WHERE a.state='ready'").Output()
	if err != nil || strings.TrimSpace(string(restored)) != "1" {
		t.Fatal("restored artifact provenance unavailable")
	}

}
