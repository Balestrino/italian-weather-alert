//go:build integration

package notifications

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHistoricalFailuresAndDailyBacklog(t *testing.T) {
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
	store, queue := New(pool), jobs.New(pool)
	relay, cfg := smtpServer(t)
	sender, err := NewSMTP(cfg)
	must(err)
	base := time.Now().UTC().Truncate(time.Microsecond)
	enqueue := func(key string, at time.Time) jobs.Job {
		t.Helper()
		j, err := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "inference", Kind: "stage", IdempotencyKey: key, Payload: json.RawMessage(`{}`), MaxAttempts: 1, RetryBase: time.Second, AvailableAt: at})
		must(err)
		return j
	}
	fail := func(key string, at time.Time) jobs.Claim {
		t.Helper()
		enqueue(key, at)
		j, err := queue.Claim(ctx, "inference", "fixture", at, time.Minute)
		must(err)
		must(queue.Fail(ctx, j, jobs.Failure{Code: "fixture_failure"}, at))
		return j
	}
	reconcile := func(at time.Time, interval time.Duration) { t.Helper(); must(store.Reconcile(ctx, at, interval)) }
	report := func() Report { t.Helper(); r, err := store.Report(ctx); must(err); return r }
	drain := func(at time.Time) {
		t.Helper()
		for n := 0; n < 20; n++ {
			worked, err := store.DeliverOne(ctx, sender, at)
			must(err)
			if !worked {
				return
			}
		}
		t.Fatal("outbox did not drain")
	}
	first := fail("first", base)
	fail("second", base)
	reconcile(base, 6*time.Hour)
	drain(base)
	// Unrelated success does not cover either failure; unchanged failures do
	// not send another stage reminder even though its interval elapsed.
	enqueue("unrelated-success", base.Add(time.Hour))
	j, err := queue.Claim(ctx, "inference", "fixture", base.Add(time.Hour), time.Minute)
	must(err)
	must(queue.Complete(ctx, j, jobs.Result{Payload: json.RawMessage(`{}`)}, base.Add(time.Hour)))
	reconcile(base.Add(6*time.Hour), 6*time.Hour)
	if r := report(); r.OpenIncidents != 1 || r.TotalMessages != 1 || r.Incidents[0].Workload.FailedJobs != 2 {
		t.Fatalf("historical failures lost or repeated: %+v", r)
	}
	// A new failure produces an update. Another failure during delayed SMTP
	// delivery must remain eligible, and the queued count must stay immutable.
	fail("third", base.Add(7*time.Hour))
	reconcile(base.Add(7*time.Hour), 6*time.Hour)
	fail("during-delivery", base.Add(7*time.Hour+time.Minute))
	reconcile(base.Add(7*time.Hour+time.Minute), 6*time.Hour)
	if r := report(); r.TotalMessages != 2 || r.Messages[0].Workload.FailedJobs != 3 || r.Incidents[0].Workload.FailedJobs != 4 {
		t.Fatalf("pending snapshot changed or duplicate queued: %+v", r)
	}
	drain(base.Add(8 * time.Hour))
	reconcile(base.Add(14*time.Hour), 6*time.Hour)
	if r := report(); r.TotalMessages != 3 || r.Messages[0].Workload.FailedJobs != 4 {
		t.Fatalf("delivery delay swallowed newer activity: %+v", r)
	}
	drain(base.Add(14 * time.Hour))
	reconcile(base.Add(20*time.Hour), 6*time.Hour)
	if report().TotalMessages != 3 {
		t.Fatal("unchanged failure repeated")
	}
	// Jobs waiting for a provider retain their attempts and count together.
	held := enqueue("held", base)
	_, err = pool.Exec(ctx, `UPDATE processing_jobs SET available_at=$2,last_error_code='provider_scope_held',updated_at=$2 WHERE id=$1`, held.ID, base.Add(24*time.Hour))
	must(err)
	reconcile(base.Add(24*time.Hour-time.Microsecond), 6*time.Hour)
	if report().TotalMessages != 3 {
		t.Fatal("backlog opened before age boundary")
	}
	reconcile(base.Add(24*time.Hour), 6*time.Hour)
	r := report()
	if r.TotalMessages != 4 || r.OpenIncidents != 2 || r.Messages[0].Category != "processing_backlog" || r.Messages[0].Workload != (Workload{Available: true, FailedJobs: 4, PendingJobs: 1, HeldJobs: 1}) {
		t.Fatalf("queue summary incorrect: %+v", r)
	}
	restart()
	store = New(pool)
	drain(base.Add(24 * time.Hour))
	reconcile(base.Add(30*time.Hour), 6*time.Hour)
	reconcile(base.Add(48*time.Hour-time.Microsecond), 6*time.Hour)
	if report().TotalMessages != 4 {
		t.Fatal("backlog reminder more frequent than daily")
	}
	reconcile(base.Add(48*time.Hour), 6*time.Hour)
	if report().TotalMessages != 5 {
		t.Fatal("daily backlog reminder absent")
	}
	drain(base.Add(48 * time.Hour))
	reconcile(base.Add(72*time.Hour), 0)
	reconcile(base.Add(72*time.Hour), 48*time.Hour)
	if report().TotalMessages != 5 {
		t.Fatal("disabled or longer reminder ignored")
	}
	reconcile(base.Add(96*time.Hour), 48*time.Hour)
	if report().TotalMessages != 6 {
		t.Fatal("longer backlog reminder absent")
	}
	drain(base.Add(96 * time.Hour))
	// Explicit relaunch is actionable, but is not a recovery. Stage reminders
	// must not be triggered merely by updated_at on held jobs.
	must(queue.Relaunch(ctx, first.ID, first.Attempt, "fixture", base.Add(97*time.Hour)))
	reconcile(base.Add(97*time.Hour), 6*time.Hour)
	if r := report(); r.TotalMessages != 7 || r.OpenIncidents != 2 || r.Messages[0].Category != "processing_failed" {
		t.Fatalf("explicit relaunch missing or fabricated recovery: %+v", r)
	}
	drain(base.Add(97 * time.Hour))
	_, err = pool.Exec(ctx, `UPDATE processing_jobs SET updated_at=$2 WHERE id=$1`, held.ID, base.Add(104*time.Hour))
	must(err)
	reconcile(base.Add(104*time.Hour), 6*time.Hour)
	if report().TotalMessages != 7 {
		t.Fatal("provider polling sent stage reminder")
	}
	var attempts int
	must(pool.QueryRow(ctx, `SELECT attempt_count FROM processing_jobs WHERE id=$1`, held.ID).Scan(&attempts))
	if attempts != 0 {
		t.Fatal("monitor consumed provider-held attempt")
	}
	// Retiring reviewed work clears an aged backlog without claiming success
	// of its failed stage or provider. All original failure attempts remain.
	_, err = queue.ArchiveBefore(ctx, base.Add(104*time.Hour), "reviewed-fixture", base.Add(105*time.Hour))
	must(err)
	reconcile(base.Add(105*time.Hour), 6*time.Hour)
	r = report()
	if r.TotalMessages != 8 || r.Messages[0].Category != "processing_backlog" || r.Messages[0].Kind != "recovered" || r.OpenIncidents != 1 {
		t.Fatalf("archival/recovery semantics: %+v", r)
	}
	drain(base.Add(105 * time.Hour))
	if !strings.Contains(r.Messages[0].Body(), "does not certify provider recovery") {
		t.Fatal("queue clearance claimed provider recovery")
	}
	must(pool.QueryRow(ctx, `SELECT count(*) FROM processing_attempts WHERE outcome='failed'`).Scan(&attempts))
	if attempts != 4 {
		t.Fatal("failure history changed")
	}
	relay.mu.Lock()
	mail := strings.Join(relay.messages, "\n")
	relay.mu.Unlock()
	if !strings.Contains(mail, "Provider-held pending jobs: 1") {
		t.Fatal("SMTP omitted held workload")
	}
}
