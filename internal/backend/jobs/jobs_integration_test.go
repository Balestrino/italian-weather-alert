//go:build integration

package jobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func queueTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	name := "iwa-jobs-test-" + hex.EncodeToString(random[:6])
	password := hex.EncodeToString(random)
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
	if err = listener.Close(); err != nil {
		t.Fatal(err)
	}
	run("run", "--pull=never", "--detach", "--name", name, "--publish", binding, "--mount", "type=bind,src="+passwordFile+",dst=/run/secrets/password,readonly", "--env", "POSTGRES_PASSWORD_FILE=/run/secrets/password", "--env", "POSTGRES_USER=iwa", "--env", "POSTGRES_DB=iwa", "postgres:16-alpine")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", name).Run() })
	address := run("port", name, "5432/tcp")
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatal(err)
	}
	portNumber, _ := strconv.Atoi(port)
	config, err := pgxpool.ParseConfig("postgres://iwa@localhost/iwa?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.Host = host
	config.ConnConfig.Port = uint16(portNumber)
	config.ConnConfig.Password = password
	config.MaxConns = 12
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	wait := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		for {
			if pool.Ping(ctx) == nil {
				return
			}
			if ctx.Err() != nil {
				t.Fatal("test PostgreSQL did not become ready")
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	wait()
	return pool, func() { run("restart", "--time", "5", name); wait() }
}

type memoryObjects struct {
	mu      sync.Mutex
	objects map[string][]byte
}

func (m *memoryObjects) Ensure(_ context.Context, hash string, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.objects == nil {
		m.objects = map[string][]byte{}
	}
	if existing, ok := m.objects[hash]; ok && string(existing) != string(body) {
		return documents.ErrStorage
	}
	m.objects[hash] = append([]byte(nil), body...)
	return nil
}

func (m *memoryObjects) Read(_ context.Context, hash string, _ int64) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	body, ok := m.objects[hash]
	if !ok {
		return nil, documents.ErrStorage
	}
	return append([]byte(nil), body...), nil
}

func queueFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (*documents.Store, documents.Acquisition) {
	t.Helper()
	reg := registry.New(pool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "municipality", Name: "Fixture municipality", OfficialURL: "https://comune.example"}))
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "publisher", Name: "Fixture publisher", OfficialURL: "https://publisher.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "channel", PublisherID: "publisher", Platform: "fixture", URL: "https://comune.example"}))
	evidence := registry.Evidence{URL: "https://comune.example/policy", Locator: "synthetic permission", ObservedAt: time.Now().UTC()}
	configuration := registry.Configuration{URL: "https://comune.example", Sections: []string{"https://comune.example/notices"}, AccessMethod: "html", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}
	must(reg.CreateSource(ctx, registry.Source{ID: "municipal", AuthorityID: "municipality", ChannelID: "channel", ProductID: "municipal", Territory: "050004"}, configuration, "test"))
	retained := documents.New(pool, &memoryObjects{})
	acquisition := documents.Acquisition{ID: "acquisition-1", SourceID: "municipal", Configuration: 1, URL: "https://comune.example/notice", Metadata: json.RawMessage(`{"published":"2026-09-01"}`), Resources: []documents.Resource{{URL: "https://comune.example/notice", Role: "original", Required: true, SourceID: "municipal", Configuration: 1, MediaType: "text/html", Bytes: []byte("fixture notice")}}}
	return retained, acquisition
}

func TestQueueRecoveryAttemptsAndIdempotentEffects(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool, restart := queueTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	retained, acquisition := queueFixture(t, ctx, pool)
	if err := retained.Stage(ctx, acquisition); err != nil {
		t.Fatal(err)
	}
	queue := New(pool)
	base := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	if count, err := EnqueuePendingDocuments(ctx, queue, retained, base); err != nil || count != 1 {
		t.Fatalf("pending reconciliation: count=%d err=%v", count, err)
	}
	if count, err := EnqueuePendingDocuments(ctx, queue, retained, base.Add(time.Second)); err != nil || count != 1 {
		t.Fatalf("idempotent reconciliation: count=%d err=%v", count, err)
	}
	first, err := queue.Claim(ctx, DocumentQueue, "worker-before-restart", base, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	firstResult, err := DocumentHandler(retained)(ctx, first.Job)
	if err != nil {
		t.Fatal(err)
	}
	// The document effect happened, but the process lost its acknowledgement.
	restart()
	second, err := queue.Claim(ctx, DocumentQueue, "worker-after-restart", base.Add(6*time.Second), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	secondResult, err := DocumentHandler(retained)(ctx, second.Job)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstResult.Payload) != string(secondResult.Payload) {
		t.Fatal("recovered handler produced a different document version")
	}
	if err = queue.Complete(ctx, first, firstResult, base.Add(7*time.Second)); !errors.Is(err, ErrStaleClaim) {
		t.Fatalf("stale worker completion accepted: %v", err)
	}
	// A transaction failure after outbox insertion must not leave a partial
	// logical/public effect.
	_, err = pool.Exec(ctx, `CREATE FUNCTION fail_job_completion() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.state='succeeded' THEN RAISE EXCEPTION 'injected completion failure'; END IF; RETURN NEW; END; $$`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `CREATE TRIGGER injected_completion_failure BEFORE UPDATE ON processing_jobs FOR EACH ROW EXECUTE FUNCTION fail_job_completion()`)
	if err != nil {
		t.Fatal(err)
	}
	if err = queue.Complete(ctx, second, secondResult, base.Add(7*time.Second)); err == nil {
		t.Fatal("completion failure was not injected")
	}
	if _, err = pool.Exec(ctx, `DROP TRIGGER injected_completion_failure ON processing_jobs`); err != nil {
		t.Fatal(err)
	}
	if err = queue.Complete(ctx, second, secondResult, base.Add(8*time.Second)); err != nil {
		t.Fatal(err)
	}
	duplicate, err := queue.Enqueue(ctx, EnqueueRequest{Queue: "effects", Kind: "fixture", IdempotencyKey: "same-logical-effect", Payload: json.RawMessage(`{}`), MaxAttempts: 1, RetryBase: time.Second, AvailableAt: base})
	if err != nil {
		t.Fatal(err)
	}
	duplicateClaim, err := queue.Claim(ctx, "effects", "effect-replay", base, 5*time.Second)
	if err != nil || duplicateClaim.ID != duplicate.ID {
		t.Fatalf("duplicate effect job claim: %v", err)
	}
	if err = queue.Complete(ctx, duplicateClaim, secondResult, base.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	var versions, effects int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM retained_versions`).Scan(&versions); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM processing_effects`).Scan(&effects); err != nil {
		t.Fatal(err)
	}
	if versions != 1 || effects != 1 {
		t.Fatalf("restart duplicated effects: versions=%d effects=%d", versions, effects)
	}
	attempts, err := queue.Attempts(ctx, first.ID)
	if err != nil || len(attempts) != 2 || attempts[0].Outcome != "abandoned" || attempts[0].ErrorCode != "lease_expired" || attempts[1].Outcome != "succeeded" {
		t.Fatalf("attempt history not persisted: %#v err=%v", attempts, err)
	}
}

func TestQueueRetryExhaustionConcurrencyAndWorker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool, _ := queueTestDB(t)
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	queue := New(pool)
	base := time.Date(2026, 9, 16, 13, 0, 0, 0, time.UTC)
	request := EnqueueRequest{Queue: "test", Kind: "retry", IdempotencyKey: "same", Payload: json.RawMessage(`{"value":1}`), MaxAttempts: 3, RetryBase: time.Second, AvailableAt: base}
	job, err := queue.Enqueue(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	again, err := queue.Enqueue(ctx, request)
	if err != nil || again.ID != job.ID {
		t.Fatal("equivalent enqueue was not idempotent")
	}
	request.Payload = json.RawMessage(`{"value":2}`)
	if _, err = queue.Enqueue(ctx, request); !errors.Is(err, ErrConflict) {
		t.Fatal("changed payload reused an idempotency key")
	}
	claim, err := queue.Claim(ctx, "test", "retry-worker", base, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err = queue.Fail(ctx, claim, Failure{Code: "temporary_upstream", Detail: "fixture failure", Temporary: true}, base); err != nil {
		t.Fatal(err)
	}
	if _, err = queue.Claim(ctx, "test", "early-worker", base.Add(999*time.Millisecond), 5*time.Second); !errors.Is(err, ErrNoJob) {
		t.Fatal("retry backoff was not enforced")
	}
	claim, err = queue.Claim(ctx, "test", "retry-worker", base.Add(time.Second), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err = queue.Fail(ctx, claim, Failure{Code: "temporary_upstream", Detail: "second fixture failure", Temporary: true}, base.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err = queue.Claim(ctx, "test", "early-worker", base.Add(2999*time.Millisecond), 5*time.Second); !errors.Is(err, ErrNoJob) {
		t.Fatal("second retry did not use an increasing wait")
	}
	claim, err = queue.Claim(ctx, "test", "retry-worker", base.Add(3*time.Second), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err = queue.Fail(ctx, claim, Failure{Code: "temporary_upstream", Detail: "third fixture failure", Temporary: true}, base.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	attempts, err := queue.Attempts(ctx, job.ID)
	if err != nil || len(attempts) != 3 || attempts[0].Outcome != "retry" || attempts[1].Outcome != "retry" || attempts[2].Outcome != "failed" {
		t.Fatalf("three-attempt exhaustion not persisted: %#v err=%v", attempts, err)
	}

	exhausted, err := queue.Enqueue(ctx, EnqueueRequest{Queue: "exhaust", Kind: "fixture", IdempotencyKey: "lease", Payload: json.RawMessage(`{}`), MaxAttempts: 1, RetryBase: time.Second, AvailableAt: base})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = queue.Claim(ctx, "exhaust", "crashed", base, 3*time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err = queue.Claim(ctx, "exhaust", "reaper", base.Add(4*time.Second), 3*time.Second); !errors.Is(err, ErrNoJob) {
		t.Fatal("exhausted lease was reclaimed")
	}
	attempts, err = queue.Attempts(ctx, exhausted.ID)
	if err != nil || len(attempts) != 1 || attempts[0].Outcome != "failed" || attempts[0].ErrorCode != "lease_expired" {
		t.Fatalf("lease exhaustion not persisted: %#v err=%v", attempts, err)
	}

	for i := 0; i < 6; i++ {
		_, err = queue.Enqueue(ctx, EnqueueRequest{Queue: "parallel", Kind: "fixture", IdempotencyKey: fmt.Sprint(i), Payload: json.RawMessage(`{}`), MaxAttempts: 1, RetryBase: time.Second, AvailableAt: base})
		if err != nil {
			t.Fatal(err)
		}
	}
	ids := make(chan int64, 6)
	errs := make(chan error, 6)
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			claimed, claimErr := queue.Claim(ctx, "parallel", fmt.Sprintf("worker-%d", i), base, 5*time.Second)
			if claimErr == nil {
				ids <- claimed.ID
			}
			errs <- claimErr
		}(i)
	}
	wg.Wait()
	close(ids)
	close(errs)
	seen := map[int64]bool{}
	for err = range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	for id := range ids {
		if seen[id] {
			t.Fatal("two workers claimed the same active job")
		}
		seen[id] = true
	}
	if len(seen) != 6 {
		t.Fatalf("claimed %d parallel jobs", len(seen))
	}

	_, err = queue.Enqueue(ctx, EnqueueRequest{Queue: "runner", Kind: "handled", IdempotencyKey: "one", Payload: json.RawMessage(`{}`), MaxAttempts: 1, RetryBase: time.Second, AvailableAt: base})
	if err != nil {
		t.Fatal(err)
	}
	worker := &Worker{Store: queue, Queue: "runner", ID: "go-worker", Lease: 6 * time.Second, PollInterval: 10 * time.Millisecond, now: func() time.Time { return base }, Handlers: map[string]Handler{"handled": func(context.Context, Job) (Result, error) {
		return JSONResult(struct {
			Handled bool `json:"handled"`
		}{true}, Effect{Key: "public-fixture:one", Kind: "public_projection", Payload: json.RawMessage(`{"visible":true}`)})
	}}}
	worked, err := worker.RunOne(ctx)
	if err != nil || !worked {
		t.Fatalf("Go worker did not process job: worked=%v err=%v", worked, err)
	}
	var count int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM processing_effects WHERE effect_key='public-fixture:one'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("worker public effect count=%d err=%v", count, err)
	}
}
