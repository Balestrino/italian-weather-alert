//go:build integration

package notifications

import (
	"context"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
	"sync"
	"testing"
	"time"
)

func TestDailyReportDurableQueueAndRetry(t *testing.T) {
	ctx := context.Background()
	pool, restart := notificationTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, Migrate, Migrate} {
		if err := m(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	store := New(pool)
	cfg := DailyReportConfig{true, "21:00", "Europe/Rome"}
	at := time.Date(2026, 9, 26, 19, 0, 0, 0, time.UTC)
	calls := 0
	var mu sync.Mutex
	generate := func(context.Context, time.Time) (ReportMail, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		return ReportMail{Subject: "Report", Body: "original snapshot", ObservedAt: at}, nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := store.QueueDaily(ctx, cfg, at, generate); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if calls != 1 {
		t.Fatalf("duplicate snapshots: %d", calls)
	}
	restart()
	if _, err := store.QueueDaily(ctx, cfg, at.Add(time.Hour), generate); err != nil || calls != 1 {
		t.Fatal("restart regenerated", err)
	}
	relay, smtpCfg := smtpServer(t)
	sender, _ := NewSMTP(smtpCfg)
	relay.mu.Lock()
	relay.reject = true
	relay.mu.Unlock()
	if done, err := store.DeliverReport(ctx, sender, at); err != nil || !done {
		t.Fatal(err)
	}
	status, err := store.DailyReportStatuses(ctx)
	if err != nil || status[0].Error == nil || status[0].SentAt != nil || status[0].DeliveryAttempts != 1 {
		t.Fatal("failure not durable", err)
	}
	if done, err := store.DeliverReport(ctx, sender, at.Add(30*time.Second)); done || err != nil {
		t.Fatal("unbounded retry")
	}
	relay.mu.Lock()
	relay.reject = false
	relay.mu.Unlock()
	if done, err := store.DeliverReport(ctx, sender, at.Add(time.Minute)); !done || err != nil {
		t.Fatal(err)
	}
	status, _ = store.DailyReportStatuses(ctx)
	if status[0].SentAt == nil || relay.count() != 1 {
		t.Fatal("SMTP acceptance not recorded")
	}
	if done, _ := store.DeliverReport(ctx, sender, at.Add(time.Hour)); done {
		t.Fatal("duplicate delivery")
	}
	tomorrow := at.AddDate(0, 0, 1)
	fail := func(context.Context, time.Time) (ReportMail, error) {
		return ReportMail{}, errors.New("private upstream details")
	}
	if _, err := store.QueueDaily(ctx, cfg, tomorrow, fail); err != nil {
		t.Fatal(err)
	}
	status, _ = store.DailyReportStatuses(ctx)
	if status[0].Generated || status[0].Error == nil || *status[0].Error != "report_generation_failed" {
		t.Fatal("generation failure lost")
	}
	if _, err := store.QueueDaily(ctx, cfg, tomorrow.Add(time.Minute), generate); err != nil {
		t.Fatal(err)
	}
	status, _ = store.DailyReportStatuses(ctx)
	if len(status) != 2 || !status[0].Generated {
		t.Fatal("backfill or failed retry")
	}
}

func TestDailyGenerationFailureDoesNotBlockIncidentDelivery(t *testing.T) {
	ctx := context.Background()
	pool, _ := notificationTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, Migrate} {
		if err := m(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	relay, cfg := smtpServer(t)
	sender, _ := NewSMTP(cfg)
	cfg.DailyReport = &DailyReportConfig{true, "21:00", "Europe/Rome"}
	at := time.Date(2026, 9, 26, 19, 0, 0, 0, time.UTC)
	store := New(pool)
	worker := DailyReportWorker{Store: store, Reader: ReportReader{Pool: pool}, Sender: sender, Config: cfg}
	if err := worker.Tick(ctx, at); err != nil {
		t.Fatal(err)
	} // no territorial schema: generation fails durably
	statuses, err := store.DailyReportStatuses(ctx)
	if err != nil || len(statuses) != 1 || statuses[0].Error == nil || statuses[0].Generated {
		t.Fatal("failure invisible", err)
	}
	_, err = pool.Exec(ctx, `WITH incident AS (INSERT INTO notification_incidents(category,scope,opened_at,last_observed_at) VALUES('processing_failed','fixture',$1,$1) RETURNING id) INSERT INTO notification_messages(message_id,incident_id,kind,created_at,next_attempt_at) SELECT 'independent@iwa.invalid',id,'opened',$1,$1 FROM incident`, at)
	if err != nil {
		t.Fatal(err)
	}
	if worked, err := store.DeliverOne(ctx, sender, at); err != nil || !worked || relay.count() != 1 {
		t.Fatal("incident delivery blocked", err)
	}
}
