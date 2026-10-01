//go:build integration

package notifications

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func notificationTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	name := "iwa-notifications-test-" + hex.EncodeToString(random[:6])
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

func TestDurableIncidentsSMTPAndRecovery(t *testing.T) {
	ctx := context.Background()
	pool, restart := notificationTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, acquisition.Migrate, Migrate, Migrate} {
		must(migrate(ctx, pool))
	}
	store := New(pool)
	relay, cfg := smtpServer(t)
	sender, err := NewSMTP(cfg)
	must(err)
	base := time.Now().UTC().Truncate(time.Microsecond)
	execSQL := func(sql string, args ...any) { t.Helper(); _, err := pool.Exec(ctx, sql, args...); must(err) }
	execSQL(`INSERT INTO registry_authorities VALUES('fixture','Fixture','https://example.test');
 INSERT INTO registry_channels VALUES('fixture','fixture','Fixture','https://example.test',false);
 INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES('source','fixture','fixture','municipal','fixture');
 INSERT INTO registry_configurations(source_id,revision,body,actor) VALUES('source',1,'{}','fixture');
 UPDATE registry_sources SET latest_revision=1,active_revision=1,collection_enabled=true WHERE id='source'`)
	schedule := acquisition.NewScheduleStore(pool)
	must(schedule.SyncEnabled(ctx, base))
	claim, err := schedule.ClaimDue(ctx, "test", base, time.Minute)
	must(err)
	must(schedule.Finish(ctx, claim, acquisition.CheckOutcome{SourceID: "source", Configuration: 1, StartedAt: base, FinishedAt: base, Reachable: true, ContentRecognized: true, Complete: true}))
	reconcile := func(at time.Time, reminder time.Duration) { t.Helper(); must(store.Reconcile(ctx, at, reminder)) }
	report := func() Report { t.Helper(); r, err := store.Report(ctx); must(err); return r }
	drain := func(at time.Time) {
		t.Helper()
		for i := 0; i < 50; i++ {
			worked, err := store.DeliverOne(ctx, sender, at)
			must(err)
			if !worked {
				return
			}
		}
		t.Fatal("outbox did not drain")
	}
	reconcile(base, time.Hour)
	// The first transient failure remains in acquisition status, without email.
	at := base.Add(10 * time.Minute)
	claim, err = schedule.ClaimDue(ctx, "test", at, time.Minute)
	must(err)
	must(schedule.Finish(ctx, claim, acquisition.CheckOutcome{SourceID: "source", Configuration: 1, StartedAt: at, FinishedAt: at, Reachable: true, ErrorCode: "fixture_temporary"}))
	reconcile(at, time.Hour)
	if report().TotalMessages != 0 {
		t.Fatal("transient error sent email")
	}
	at = base.Add(30 * time.Minute)
	reconcile(at, time.Hour)
	for i := 0; i < 3; i++ {
		reconcile(at, time.Hour)
	}
	if r := report(); r.OpenIncidents != 1 || r.TotalMessages != 1 {
		t.Fatalf("source grouping: %+v", r)
	}
	// Rejection persists a safe code and a one-minute retry, not another incident.
	relay.mu.Lock()
	relay.reject = true
	relay.mu.Unlock()
	drain(at)
	if r := report(); r.PendingMessages != 1 || r.Messages[0].Attempts != 1 || r.Messages[0].ErrorCode != "smtp_delivery_failed" {
		t.Fatal("delivery failure not retained")
	}
	if relay.count() != 0 {
		t.Fatal("rejected message counted as sent")
	}
	restart()
	store = New(pool)
	relay.mu.Lock()
	relay.reject = false
	relay.mu.Unlock()
	drain(at.Add(59 * time.Second))
	if relay.count() != 0 {
		t.Fatal("SMTP backoff ignored")
	}
	at = at.Add(time.Minute)
	drain(at)
	if relay.count() != 1 || report().PendingMessages != 0 {
		t.Fatal("restart lost notification")
	}
	reconcile(at.Add(time.Hour), 0)
	if report().TotalMessages != 1 {
		t.Fatal("disabled reminders sent")
	}
	reconcile(at.Add(time.Hour), time.Hour)
	drain(at.Add(time.Hour))
	if relay.count() != 2 {
		t.Fatal("configured reminder absent")
	}
	execSQL(`UPDATE registry_sources SET collection_enabled=false WHERE id='source'`)
	reconcile(at.Add(2*time.Hour), time.Hour)
	if report().OpenIncidents != 1 || report().TotalMessages != 2 {
		t.Fatal("suspension fabricated recovery")
	}
	execSQL(`UPDATE registry_sources SET collection_enabled=true WHERE id='source'`)
	at = at.Add(2 * time.Hour)
	claim, err = schedule.ClaimDue(ctx, "test", at, time.Minute)
	must(err)
	must(schedule.Finish(ctx, claim, acquisition.CheckOutcome{SourceID: "source", Configuration: 1, StartedAt: at, FinishedAt: at, Reachable: true, ContentRecognized: true, Complete: true}))
	reconcile(at, time.Hour)
	drain(at)
	if report().OpenIncidents != 0 || relay.count() != 3 {
		t.Fatal("complete check did not recover source")
	}
	// A later outage is a new incident, including sources with no successful baseline.
	execSQL(`UPDATE acquisition_source_status SET last_complete_at=NULL,last_started_at=$1`, at)
	reconcile(at.Add(30*time.Minute), time.Hour)
	if report().OpenIncidents != 1 {
		t.Fatal("no-baseline delay absent")
	}
	execSQL(`UPDATE acquisition_source_status SET last_complete_at=$1`, at.Add(31*time.Minute))
	reconcile(at.Add(31*time.Minute), time.Hour)
	drain(at.Add(31 * time.Minute))
	execSQL(`UPDATE registry_sources SET collection_enabled=false WHERE id='source'`)

	// Exhausted jobs in the same queue/stage produce one incident. Automatic
	// retries and explicit relaunches are not recoveries; all affected jobs must succeed.
	at = base.Add(5 * time.Hour)
	queue := jobs.New(pool)
	previous := report().TotalMessages
	var failed []jobs.Claim
	for n := 0; n < 2; n++ {
		_, err := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "fixture", Kind: "stage", IdempotencyKey: fmt.Sprint(n), Payload: json.RawMessage(`{}`), MaxAttempts: 3, RetryBase: time.Second, AvailableAt: at})
		must(err)
		for attempt := 0; attempt < 3; attempt++ {
			j, err := queue.Claim(ctx, "fixture", "test", at, time.Minute)
			must(err)
			must(queue.Fail(ctx, j, jobs.Failure{Code: "temporary", Detail: "private upstream text", Temporary: true}, at))
			reconcile(at, time.Hour)
			if n == 0 && attempt < 2 && report().TotalMessages != previous {
				t.Fatal("processing retry sent email")
			}
			if attempt == 2 {
				failed = append(failed, j)
			}
			at = at.Add(3 * time.Second)
		}
	}
	if report().TotalMessages != previous+1 {
		t.Fatal("exhausted jobs were not grouped")
	}
	for i, j := range failed {
		must(queue.Relaunch(ctx, j.ID, j.Attempt, "fixture", at))
		reconcile(at, time.Hour)
		if report().OpenIncidents != 1 {
			t.Fatal("relaunch fabricated recovery")
		}
		resumed, err := queue.Claim(ctx, "fixture", "test", at, time.Minute)
		must(err)
		must(queue.Complete(ctx, resumed, jobs.Result{Payload: json.RawMessage(`{}`)}, at))
		reconcile(at, time.Hour)
		if i == 0 && report().OpenIncidents != 1 {
			t.Fatal("partial group recovery")
		}
	}
	if report().OpenIncidents != 0 || report().TotalMessages != previous+2 {
		t.Fatal("successful relaunch missing recovery")
	}
	drain(at)

	// Backup results are durable hooks. Even failure and recovery between monitor
	// polls retain both notifications; repeated failed runs share the incident.
	previous = report().TotalMessages
	must(store.RecordBackup(ctx, "daily", "run-1", false, at))
	must(store.RecordBackup(ctx, "daily", "run-1", false, at))
	if err = store.RecordBackup(ctx, "daily", "run-1", true, at); !errors.Is(err, ErrInvalid) {
		t.Fatal("backup history rewritten")
	}
	must(store.RecordBackup(ctx, "daily", "run-2", false, at.Add(time.Second)))
	must(store.RecordBackup(ctx, "daily", "run-3", true, at.Add(2*time.Second)))
	// Older out-of-order results must not reopen a recovered backup incident.
	must(store.RecordBackup(ctx, "daily", "old-run", false, at.Add(-time.Hour)))
	reconcile(at.Add(3*time.Second), time.Hour)
	if report().OpenIncidents != 0 || report().TotalMessages != previous+2 {
		t.Fatal("backup grouping or recovery")
	}
	// Concurrent dispatchers cannot both send the same pending row.
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := store.DeliverOne(ctx, sender, at.Add(3*time.Second)); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		must(err)
	}
	drain(at.Add(3 * time.Second))
	if report().PendingMessages != 0 || relay.count() != report().TotalMessages {
		t.Fatal("missing or duplicate email")
	}
	relay.mu.Lock()
	all := strings.Join(relay.messages, "\n")
	relay.mu.Unlock()
	for _, secret := range []string{"private upstream text", "do-not-persist"} {
		if strings.Contains(all, secret) {
			t.Fatal("upstream details leaked into email")
		}
	}
	for _, category := range []string{"source_delay", "processing_failed", "backup_failed"} {
		if !strings.Contains(all, category) {
			t.Fatal("missing category")
		}
	}
	// Read failure remains an error, not a synthetic healthy/empty result.
	execSQL(`ALTER TABLE acquisition_source_status RENAME TO acquisition_source_status_unavailable`)
	if err = store.Reconcile(ctx, at.Add(time.Hour), time.Hour); err == nil {
		t.Fatal("storage failure hidden")
	}
	execSQL(`ALTER TABLE acquisition_source_status_unavailable RENAME TO acquisition_source_status`)
}

func TestActivatedUnscheduledSourceAndWorker(t *testing.T) {
	ctx := context.Background()
	pool, _ := notificationTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, acquisition.Migrate, Migrate} {
		must(migrate(ctx, pool))
	}
	_, err := pool.Exec(ctx, `INSERT INTO registry_authorities VALUES('fixture','Fixture','https://example.test');
 INSERT INTO registry_channels VALUES('fixture','fixture','Fixture','https://example.test',false);
 INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES('unscheduled','fixture','fixture','municipal','fixture');
 INSERT INTO registry_configurations(source_id,revision,body,actor) VALUES('unscheduled',1,'{"delay_seconds":60}','fixture');
 UPDATE registry_sources SET latest_revision=1,active_revision=1,collection_enabled=true WHERE id='unscheduled'`)
	must(err)
	store := New(pool)
	base := time.Now().UTC().Truncate(time.Microsecond).Add(-time.Minute)
	must(store.Reconcile(ctx, base, time.Hour))
	must(store.Reconcile(ctx, base.Add(59*time.Second), time.Hour))
	r, err := store.Report(ctx)
	must(err)
	if r.TotalMessages != 0 {
		t.Fatal("initial grace period absent")
	}
	must(store.Reconcile(ctx, base.Add(time.Minute), time.Hour))
	r, err = store.Report(ctx)
	must(err)
	if r.OpenIncidents != 1 || r.TotalMessages != 1 {
		t.Fatal("unscheduled enabled source was ignored")
	}
	relay, cfg := smtpServer(t)
	sender, err := NewSMTP(cfg)
	must(err)
	worker := Worker{Store: store, Sender: sender, Config: cfg}
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- worker.Run(workerCtx) }()
	defer func() {
		cancel()
		select {
		case err := <-done:
			must(err)
		case <-time.After(3 * time.Second):
			t.Error("notification worker did not stop")
		}
	}()
	deadline := time.Now().Add(3 * time.Second)
	for relay.count() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if relay.count() != 1 {
		t.Fatal("runtime worker did not deliver queued incident")
	}
}

func TestArchivedFailuresDoNotOpenProcessingIncidents(t *testing.T) {
	ctx := context.Background()
	pool, _ := notificationTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, acquisition.Migrate, Migrate} {
		if err := m(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	queue := jobs.New(pool)
	_, err := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "fixture", Kind: "fixture", IdempotencyKey: "old-failure", Payload: json.RawMessage(`{}`), MaxAttempts: 1, RetryBase: time.Second, AvailableAt: now})
	if err != nil {
		t.Fatal(err)
	}
	claim, err := queue.Claim(ctx, "fixture", "worker", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err = queue.Fail(ctx, claim, jobs.Failure{Code: "fixture", Temporary: false}, now); err != nil {
		t.Fatal(err)
	}
	if _, err = queue.ArchiveBefore(ctx, now, "operator", now); err != nil {
		t.Fatal(err)
	}
	if err = New(pool).Reconcile(ctx, now, time.Hour); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM notification_incidents WHERE category='processing_failed'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("archived error reopened: %d %v", count, err)
	}
}
