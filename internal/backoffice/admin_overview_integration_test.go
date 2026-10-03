//go:build integration

package backoffice

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDetailedDashboardSnapshotHistoryAndDrilldowns(t *testing.T) {
	ctx := context.Background()
	pool := adminTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, acquisition.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate, interpretation.Migrate} {
		must(m(ctx, pool))
	}
	at := time.Date(2026, 10, 2, 14, 30, 0, 0, time.UTC)
	s := operations.New(pool)
	empty, err := s.Dashboard(ctx, "24h", at)
	must(err)
	if len(empty.History) != 24 || len(empty.Groups) != 0 || empty.Jobs != (operations.JobCounts{}) {
		t.Fatalf("empty scope: %#v", empty)
	}
	reg := registry.New(pool)
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "Synthetic", OfficialURL: "https://fixture.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "fixture", URL: "https://fixture.example"}))
	ev := registry.Evidence{URL: "https://fixture.example/evidence", Locator: "synthetic", ObservedAt: at.Add(-time.Hour)}
	cfg := registry.Configuration{URL: "https://fixture.example", Sections: []string{"https://fixture.example/notices"}, AccessMethod: "fixture", Attribution: "Fixture", Provenance: &ev, Policy: registry.Policy{Evidence: &ev, CollectionPermitted: true, RetentionPermitted: true}}
	for _, source := range []string{"s", "suspended"} {
		must(reg.CreateSource(ctx, registry.Source{ID: source, AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "050004"}, cfg, "fixture"))
		must(reg.RecordPreview(ctx, source, 1, "fixture", ev))
		must(reg.EnableCollection(ctx, source, 1, "fixture"))
	}
	retained := documents.New(pool, adminObjects{})
	versions := map[string]documents.Version{}
	for _, category := range []string{"unscheduled", "preparing", "queued", "running", "retry_wait", "failed", "incomplete", "suspended"} {
		source := "s"
		if category == "suspended" {
			source = "suspended"
		}
		u := "https://fixture.example/" + category
		v, e := retained.Retain(ctx, documents.Acquisition{ID: category, SourceID: source, Configuration: 1, URL: u, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: u, Role: "original", Required: true, SourceID: source, Configuration: 1, MediaType: "text/html", Bytes: []byte("Avviso sintetico " + category)}}})
		must(e)
		versions[category] = v
	}
	_, err = pool.Exec(ctx, `INSERT INTO interpretation_triggers(document_version_id,evidence_hash,reason,workload,created_at) VALUES($1,$2,'content_changed','ordinary',$3)`, versions["preparing"].ID, versions["preparing"].Hash, at.Add(-time.Hour))
	must(err)
	must(reg.SuspendInterpretation(ctx, "suspended", 1, "fixture", ev))
	queue := jobs.New(pool)
	enqueue := func(key, kind, category string) jobs.Job {
		t.Helper()
		body, _ := json.Marshal(map[string]any{"document_version_id": versions[category].ID})
		j, e := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "inference", Kind: kind, IdempotencyKey: key, Payload: body, MaxAttempts: 5, RetryBase: time.Second, AvailableAt: at.Add(-time.Hour)})
		must(e)
		return j
	}
	for i := 0; i < 105; i++ {
		enqueue(fmt.Sprint("queued-", i), "ocr_resource", "queued")
	}
	running := enqueue("running", "ocr_resource", "running")
	_, err = pool.Exec(ctx, `UPDATE processing_jobs SET state='running',claimed_by='fixture',claim_token='fixture-running',lease_expires_at=$2,attempt_count=1 WHERE id=$1`, running.ID, at.Add(-time.Minute))
	must(err)
	retry := enqueue("retry", "ocr_resource", "retry_wait")
	_, err = pool.Exec(ctx, `UPDATE processing_jobs SET state='retry_wait',available_at=$2 WHERE id=$1`, retry.ID, at.Add(time.Hour))
	must(err)
	failed := enqueue("failed", "classify_relevance", "failed")
	_, err = pool.Exec(ctx, `UPDATE processing_jobs SET state='failed',completed_at=$2,last_error_code='synthetic_failure',attempt_count=3 WHERE id=$1`, failed.ID, at.Add(-time.Minute))
	must(err)
	archived := enqueue("archived", "classify_relevance", "unscheduled")
	_, err = pool.Exec(ctx, `UPDATE processing_jobs SET state='failed',completed_at=$2,archived_at=$3,archived_by='fixture',last_error_code='archived_error' WHERE id=$1`, archived.ID, at.Add(-time.Hour), at)
	must(err)
	// Repeated attempts cross bucket boundaries; archived work stays in history.
	for i, item := range []struct {
		job      int64
		outcome  string
		finished time.Time
	}{
		{failed.ID, "retry", at.Truncate(time.Hour).Add(-time.Hour)},
		{failed.ID, "failed", at.Add(-time.Minute)},
		{failed.ID, "succeeded", at.Add(-2 * time.Minute)},
		{archived.ID, "abandoned", at.Add(-time.Minute)},
		{archived.ID, "failed", at}, // cutoff is excluded
		{archived.ID, "retry", at.Add(-40 * 24 * time.Hour)},
	} {
		_, e := pool.Exec(ctx, `INSERT INTO processing_attempts(job_id,number,worker_id,claim_token,started_at,lease_expires_at,finished_at,outcome) VALUES($1,$2,'fixture',$3,$4,$5,$5,$6)`, item.job, i+1, fmt.Sprint("attempt-", i), item.finished.Add(-time.Second), item.finished, item.outcome)
		must(e)
	}
	proc := processing.New(pool)
	catalog, err := classification.RegisterCatalog(ctx, proc, "fixture", "fixture-model", at.Add(-time.Hour))
	must(err)
	id := versions["incomplete"].ID
	source := "s"
	run, err := proc.StartRun(ctx, processing.RunRequest{IdempotencyKey: "incomplete", Workload: "ordinary", Stage: "classification", ConfigurationVersionID: catalog.ConfigurationVersionID, SourceID: &source, DocumentVersionID: &id, Subject: json.RawMessage(`{}`), CreatedAt: at.Add(-time.Hour)})
	must(err)
	_, err = pool.Exec(ctx, `INSERT INTO classification_results(run_id,document_version_id,status,reason_code,evidence_quote,content_sha256,content_complete,created_at) VALUES($1,$2,'undetermined','empty_content','',repeat('a',64),false,$3)`, run.ID, id, at.Add(-time.Hour))
	must(err)
	d, err := s.Dashboard(ctx, "24h", at)
	must(err)
	if d.Jobs != (operations.JobCounts{Queued: 105, Running: 1, RetryWait: 1, Failed: 1, Archived: 1, Due: 105, ExpiredLeases: 1}) {
		t.Fatalf("job counts inflated or archived included: %#v", d.Jobs)
	}
	if d.PendingDocuments != 8 || d.FailedJobs != 1 {
		t.Fatalf("overview differs: %#v", d.Overview)
	}
	var counts operations.DocumentBreakdown
	for _, b := range d.Documents {
		counts.Total += b.Total
		counts.Unscheduled += b.Unscheduled
		counts.Preparing += b.Preparing
		counts.Queued += b.Queued
		counts.Running += b.Running
		counts.RetryWait += b.RetryWait
		counts.Failed += b.Failed
		counts.Incomplete += b.Incomplete
		counts.Suspended += b.Suspended
	}
	if counts != (operations.DocumentBreakdown{Total: 8, Unscheduled: 1, Preparing: 1, Queued: 1, Running: 1, RetryWait: 1, Failed: 1, Incomplete: 1, Suspended: 1}) {
		t.Fatalf("document categories: %#v", counts)
	}
	if len(d.Reasons) != 1 || d.Reasons[0].Code != "synthetic_failure" || d.Reasons[0].Count != 1 {
		t.Fatalf("reason counts: %#v", d.Reasons)
	}
	chart := makeHistoryChart(d.History, "24h")
	if len(d.History) != 24 || chart.Succeeded != 1 || chart.Retry != 1 || chart.Failed != 1 || chart.Abandoned != 1 || d.History[22].Retry != 1 || d.History[23].Failed != 1 {
		t.Fatalf("history boundary/counts: %#v", d.History)
	}
	for _, period := range []string{"7d", "30d"} {
		report, e := s.Dashboard(ctx, period, at)
		must(e)
		want := 7
		if period == "30d" {
			want = 30
		}
		if len(report.History) != want {
			t.Fatalf("%s buckets=%d", period, len(report.History))
		}
	}
	for _, f := range []operations.Filter{{Page: 1, Kind: "ocr_resource", State: "queued", Queue: "inference"}, {Page: 2, Kind: "ocr_resource", State: "queued", Queue: "inference"}, {Page: 1, State: "archived"}, {Page: 1, State: "failed", ErrorCode: "synthetic_failure"}, {Page: 1, State: "failed", ErrorCode: "absent"}} {
		page, e := s.Page(ctx, "jobs", f, at)
		must(e)
		want := 1
		if f.State == "queued" {
			want = 100
			if f.Page == 2 {
				want = 5
			}
		}
		if f.ErrorCode == "absent" {
			want = 0
		}
		if len(page.Table.Rows) != want {
			t.Fatalf("filter %#v count %d want %d", f, len(page.Table.Rows), want)
		}
	}
	// Reads must leave attempts and jobs untouched.
	var jobsCount, attemptsCount int
	must(pool.QueryRow(ctx, `SELECT count(*) FROM processing_jobs`).Scan(&jobsCount))
	must(pool.QueryRow(ctx, `SELECT count(*) FROM processing_attempts`).Scan(&attemptsCount))
	if jobsCount != 109 || attemptsCount != 6 {
		t.Fatal("dashboard reads mutated queue")
	}
}
