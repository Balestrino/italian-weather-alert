//go:build integration

package interpretation

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

func interpretationTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		t.Fatal(err)
	}
	name, password := "iwa-interpretation-test-"+hex.EncodeToString(token[:6]), hex.EncodeToString(token)
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
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", name).Run() })
	// Docker owns published-port allocation; an OS-selected free port can
	// already belong to another container's forwarding rule.
	run("run", "--pull=never", "--detach", "--name", name, "--publish", "127.0.0.1::5432", "--mount", "type=bind,src="+passwordFile+",dst=/run/secrets/password,readonly", "--env", "POSTGRES_PASSWORD_FILE=/run/secrets/password", "--env", "POSTGRES_USER=iwa", "--env", "POSTGRES_DB=iwa", "postgres:16-alpine")
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

type interpretationObjects struct {
	mu    sync.Mutex
	items map[string][]byte
}

func (m *interpretationObjects) Ensure(_ context.Context, hash string, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.items == nil {
		m.items = map[string][]byte{}
	}
	m.items[hash] = append([]byte(nil), body...)
	return nil
}

func (m *interpretationObjects) Read(_ context.Context, hash string, _ int64) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]byte(nil), m.items[hash]...), nil
}

type interpretationAdapter struct{ calls int }

func (a *interpretationAdapter) Name() string { return "fixture" }
func (a *interpretationAdapter) Complete(context.Context, inference.Request) (inference.Response, error) {
	a.calls++
	return inference.Response{ID: "fixture-response", Model: "qwen3.8-27b", Content: `{"relevant":false,"reason_code":"not_relevant","evidence_quote":"Avviso comunale"}`}, nil
}

func TestSchedulingTracksDependenciesAndSelectedReprocessing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := interpretationTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, classification.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 9, 17, 20, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://policy.example/rules", Locator: "fixture", ObservedAt: now}
	addSource := func(id string) {
		t.Helper()
		if err := reg.CreateAuthority(ctx, registry.Authority{ID: id + "-authority", Name: id, OfficialURL: "https://" + id + ".example"}); err != nil {
			t.Fatal(err)
		}
		if err := reg.CreateChannel(ctx, registry.Channel{ID: id + "-channel", PublisherID: id + "-authority", Platform: "fixture", URL: "https://" + id + ".example"}); err != nil {
			t.Fatal(err)
		}
		if err := reg.CreateSource(ctx, registry.Source{ID: id, AuthorityID: id + "-authority", ChannelID: id + "-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://" + id + ".example", Sections: []string{"https://" + id + ".example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"); err != nil {
			t.Fatal(err)
		}
	}
	addSource("source-a")
	addSource("source-b")
	objects := &interpretationObjects{}
	retained := documents.New(pool, objects)
	retain := func(id, source, raw string, pdf []byte) documents.Version {
		t.Helper()
		resources := []documents.Resource{{URL: raw, Role: "original", Required: true, SourceID: source, Configuration: 1, MediaType: "text/html", Bytes: []byte("<main>Avviso comunale senza misure meteo.</main>")}}
		if pdf != nil {
			resources = append(resources, documents.Resource{URL: raw + "/ordinanza.pdf", Role: "attachment", Required: true, SourceID: source, Configuration: 1, MediaType: "application/pdf", Bytes: pdf})
		}
		version, err := retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: source, Configuration: 1, URL: raw, Metadata: json.RawMessage(`{}`), Resources: resources})
		if err != nil {
			t.Fatal(err)
		}
		return version
	}
	v1 := retain("a-v1", "source-a", "https://source-a.example/notice", []byte("first attachment"))
	v2 := retain("a-v2", "source-a", "https://source-a.example/notice", []byte("modified attachment"))
	vb := retain("b-v1", "source-b", "https://source-b.example/notice", nil)
	if _, err := pool.Exec(ctx, "UPDATE retained_versions SET first_acquired_at=CASE id WHEN $1 THEN $4::timestamptz WHEN $2 THEN $5::timestamptz ELSE $6::timestamptz END WHERE id=ANY($3)", v1.ID, v2.ID, []int64{v1.ID, v2.ID, vb.ID}, now.Add(-48*time.Hour), now.Add(-2*time.Hour), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	queue := jobs.New(pool)
	policy := inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}
	scheduler := New(pool, queue, policy, false)
	if scheduled, err := scheduler.Automatic(ctx, AcquisitionEvent{DocumentVersionID: v1.ID, EvidenceHash: v1.Hash, Workload: "ordinary", ContentChanged: true, At: now}); err != nil || !scheduled {
		t.Fatalf("initial schedule: %v %v", scheduled, err)
	}
	if _, err := scheduler.Automatic(ctx, AcquisitionEvent{DocumentVersionID: v1.ID, EvidenceHash: v1.Hash, Workload: "ordinary", ContentChanged: true, At: now.Add(time.Minute)}); err != nil {
		t.Fatal(err)
	}
	var ocrJobs, classificationJobs int
	if err := pool.QueryRow(ctx, "SELECT count(*) FILTER(WHERE kind=$1),count(*) FILTER(WHERE kind=$2) FROM processing_jobs", ocr.Kind, classification.Kind).Scan(&ocrJobs, &classificationJobs); err != nil {
		t.Fatal(err)
	}
	if ocrJobs != 1 || classificationJobs != 0 {
		t.Fatalf("unchanged evidence created calls or skipped OCR dependency: ocr=%d classification=%d", ocrJobs, classificationJobs)
	}
	process := processing.New(pool)
	ocrCatalog, err := ocr.RegisterCatalog(ctx, process, "openai-chat", "deepseek-ocr-2", now)
	if err != nil {
		t.Fatal(err)
	}
	putOCR := func(version documents.Version, key string) {
		t.Helper()
		source := "source-a"
		run, runErr := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: key, Workload: "ordinary", Stage: "ocr", ConfigurationVersionID: ocrCatalog.ConfigurationVersionID, SourceID: &source, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: now})
		if runErr != nil {
			t.Fatal(runErr)
		}
		if runErr = ocr.NewStore(pool).PutResource(ctx, ocr.ResourceResult{RunID: run.ID, DocumentVersionID: version.ID, ResourceURL: "https://source-a.example/notice/ordinanza.pdf", Status: "complete", PageCount: 0, CreatedAt: now}); runErr != nil {
			t.Fatal(runErr)
		}
		if ready, readyErr := scheduler.AfterOCR(ctx, version.ID, "ordinary", now); readyErr != nil || !ready {
			t.Fatalf("classification was not released after OCR: %v %v", ready, readyErr)
		}
	}
	putOCR(v1, "fixture-ocr-v1")
	if scheduled, err := scheduler.Automatic(ctx, AcquisitionEvent{DocumentVersionID: v2.ID, EvidenceHash: v2.Hash, Workload: "ordinary", DependencyChanged: true, At: now.Add(2 * time.Minute)}); err != nil || !scheduled {
		t.Fatalf("modified attachment schedule: %v %v", scheduled, err)
	}
	putOCR(v2, "fixture-ocr-v2")
	if err := pool.QueryRow(ctx, "SELECT count(*) FILTER(WHERE kind=$1),count(*) FILTER(WHERE kind=$2) FROM processing_jobs", ocr.Kind, classification.Kind).Scan(&ocrJobs, &classificationJobs); err != nil {
		t.Fatal(err)
	}
	if ocrJobs != 2 || classificationJobs != 2 {
		t.Fatalf("modified attachment did not schedule only affected evidence: ocr=%d classification=%d", ocrJobs, classificationJobs)
	}
	classCatalog, err := classification.RegisterCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	source := "source-a"
	oldRun, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "earlier-failed-classification", Workload: "ordinary", Stage: "classification", ConfigurationVersionID: classCatalog.ConfigurationVersionID, SourceID: &source, DocumentVersionID: &v1.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: now.Add(-24 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	attempt, err := process.StartAttempt(ctx, processing.AttemptStart{RunID: oldRun.ID, StartedAt: now.Add(-24 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = process.FinishAttempt(ctx, processing.AttemptFinish{RunID: oldRun.ID, Number: attempt.Number, FinishedAt: now.Add(-24*time.Hour + time.Second), Outcome: "failed", Usage: processing.Usage{Status: "unavailable"}, ErrorCode: "provider_timeout"}); err != nil {
		t.Fatal(err)
	}
	sourceScope := Selection{SourceID: "source-a", Actor: "operator", At: now.Add(3 * time.Minute)}
	sourceSelection, count, err := scheduler.Reprocess(ctx, sourceScope)
	if err != nil || count != 2 {
		t.Fatalf("source selection: id=%s count=%d err=%v", sourceSelection, count, err)
	}
	_, count, err = scheduler.Reprocess(ctx, Selection{From: timePointer(now.Add(-3 * time.Hour)), Actor: "operator", At: now.Add(4 * time.Minute)})
	if err != nil || count != 2 {
		t.Fatalf("period selection: count=%d err=%v", count, err)
	}
	_, count, err = scheduler.Reprocess(ctx, Selection{ErrorCode: "provider_timeout", Actor: "operator", At: now.Add(5 * time.Minute)})
	if err != nil || count != 1 {
		t.Fatalf("error selection: count=%d err=%v", count, err)
	}
	_ = retain("a-late", "source-a", "https://source-a.example/late", nil)
	if retriedID, retriedCount, retryErr := scheduler.Reprocess(ctx, sourceScope); retryErr != nil || retriedID != sourceSelection || retriedCount != 2 {
		t.Fatalf("persisted source selection expanded on retry: id=%s count=%d err=%v", retriedID, retriedCount, retryErr)
	}
	var selectedJobID int64
	var selectedPayload json.RawMessage
	if err = pool.QueryRow(ctx, "SELECT classification_job_id,j.payload FROM interpretation_reprocessing_selections s JOIN processing_jobs j ON j.id=s.classification_job_id WHERE request_id=$1 AND document_version_id=$2", sourceSelection, v1.ID).Scan(&selectedJobID, &selectedPayload); err != nil {
		t.Fatal(err)
	}
	claimAt := now.Add(10 * time.Minute)
	if _, err = pool.Exec(ctx, "UPDATE processing_jobs SET available_at=CASE WHEN id=$1 THEN $2::timestamptz ELSE $3::timestamptz END WHERE state='queued'", selectedJobID, claimAt.Add(-time.Second), claimAt.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	claim, err := queue.Claim(ctx, inference.Queue, "interpretation-test-worker", claimAt, time.Minute)
	if err != nil || claim.ID != selectedJobID {
		t.Fatalf("selected reprocessing claim: %#v %v", claim, err)
	}
	adapter := &interpretationAdapter{}
	runner := &classification.Runner{Documents: retained, OCR: ocr.NewStore(pool), Processing: process, Results: classification.NewStore(pool), Adapter: adapter, Model: "qwen3.8-27b", ConfigurationVersion: classCatalog.ConfigurationVersionID, Now: func() time.Time { return now.Add(6 * time.Minute) }}
	claim.Payload = selectedPayload
	if _, err = runner.Handler()(ctx, claim.Job); err != nil {
		t.Fatal(err)
	}
	var runs, earlier int
	if err = pool.QueryRow(ctx, "SELECT count(*),count(*) FILTER(WHERE id=$2) FROM processing_runs WHERE stage='classification' AND document_version_id=$1", v1.ID, oldRun.ID).Scan(&runs, &earlier); err != nil {
		t.Fatal(err)
	}
	if runs != 2 || earlier != 1 || adapter.calls != 1 {
		t.Fatalf("selected reprocessing replaced prior run: runs=%d earlier=%d calls=%d", runs, earlier, adapter.calls)
	}
}

func timePointer(value time.Time) *time.Time { return &value }
