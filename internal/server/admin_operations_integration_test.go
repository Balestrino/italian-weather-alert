//go:build integration

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/domain"
	"github.com/Balestrino/italian-weather-alert/internal/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/linking"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/operations"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

func TestOperationalPagesAccountingAndSelectedActions(t *testing.T) {
	ctx := context.Background()
	pool := adminTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, jobs.Migrate, processing.Migrate, acquisition.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, interpretation.Migrate, domain.Migrate, domain.MigrateTemporal, domain.MigrateQuality} {
		must(m(ctx, pool))
	}
	empty, emptyErr := operations.New(pool).Overview(ctx, time.Now())
	must(emptyErr)
	if empty.SourceIssues != 0 || empty.FailedJobs != 0 || empty.PendingDocuments != 0 {
		t.Fatal("nonempty overview")
	}
	reg := registry.New(pool)
	now := time.Now().UTC().Add(-time.Hour)
	ev := registry.Evidence{URL: "https://source.example/report", Locator: "synthetic evaluation", ObservedAt: now}
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "Comune", OfficialURL: "https://source.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "fixture", URL: "https://source.example"}))
	cfg := registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/notices"}, AccessMethod: "fixture", Attribution: "Comune", Provenance: &ev, Policy: registry.Policy{Evidence: &ev, CollectionPermitted: true, RetentionPermitted: true, PublicationPermitted: true}}
	for _, id := range []string{"s", "other", "draft"} {
		must(reg.CreateSource(ctx, registry.Source{ID: id, AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "050004"}, cfg, "fixture"))
	}
	must(reg.RecordPreview(ctx, "s", 1, "fixture", ev))
	must(reg.EnableCollection(ctx, "s", 1, "fixture"))
	must(acquisition.NewScheduleStore(pool).SyncEnabled(ctx, now))
	_, err := pool.Exec(ctx, `UPDATE acquisition_source_status SET last_complete_at=$1,last_reachable_at=$1,last_content_at=$1,first_error_at=$2,last_error_at=$2,last_error_code='invalid_content',consecutive_failures=1 WHERE source_id='s'`, now, now.Add(time.Minute))
	must(err)
	retained := documents.New(pool, adminObjects{})
	retain := func(id, source string) documents.Version {
		u := "https://source.example/notices/" + id
		v, e := retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: source, Configuration: 1, URL: u, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: u, Role: "original", Required: true, SourceID: source, Configuration: 1, MediaType: "text/html", Bytes: []byte("Avviso comunale senza misure meteo.")}}})
		must(e)
		return v
	}
	version := retain("failed", "s")
	irrelevant := retain("irrelevant", "s")
	other := retain("other", "other")
	queue := jobs.New(pool)
	proc := processing.New(pool)
	model := processing.ModelVersion{ID: "fixture-model", Provider: "fixture", Model: "qwen-fixture", Revision: "1", Capabilities: json.RawMessage(`{}`), CreatedAt: now}
	must(proc.RegisterModel(ctx, model))
	config := processing.ConfigurationVersion{ID: "fixture-class", Name: "fixture", Stage: "classification", Revision: "1", ModelVersionID: &model.ID, LogicVersion: "fixture-1", Settings: json.RawMessage(`{"api_key":"must-not-appear"}`), CreatedAt: now}
	must(proc.RegisterConfiguration(ctx, config))
	price := processing.PriceVersion{ID: "price", ModelVersionID: model.ID, Currency: "EUR", ProvenanceURL: "https://provider.example/pricing", ObservedAt: now, CreatedAt: now, Details: json.RawMessage(`{}`), Rates: []processing.Rate{{Metric: "input_tokens", UnitSize: 1, PriceMicrounits: 2}, {Metric: "output_tokens", UnitSize: 1, PriceMicrounits: 3}}}
	must(proc.RegisterPrice(ctx, price))
	payload, _ := json.Marshal(map[string]any{"document_version_id": version.ID, "workload": "bootstrap", "private": "must-not-appear"})
	request := jobs.EnqueueRequest{Queue: "fixture", Kind: "fixture", IdempotencyKey: "failed-job", Payload: payload, MaxAttempts: 3, RetryBase: time.Second, AvailableAt: now}
	job, err := queue.Enqueue(ctx, request)
	must(err)
	source := "s"
	run, err := proc.StartRun(ctx, processing.RunRequest{IdempotencyKey: "failed-run", Workload: "bootstrap", Stage: "classification", ConfigurationVersionID: config.ID, SourceID: &source, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{}`), CreatedAt: now})
	must(err)
	number := func(n int64) *int64 { return &n }
	for i := 1; i <= 3; i++ {
		at := now.Add(time.Duration(i*10) * time.Second)
		claim, e := queue.Claim(ctx, "fixture", "test", at, time.Minute)
		must(e)
		attempt, e := proc.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, QueueJobID: &claim.ID, QueueAttemptNumber: &claim.Attempt, StartedAt: at})
		must(e)
		usage := processing.Usage{Status: "unavailable"}
		var priceID *string
		if i == 1 {
			usage = processing.Usage{Status: "reported", InputTokens: number(10), OutputTokens: number(2)}
			priceID = &price.ID
		}
		if i == 2 {
			usage = processing.Usage{Status: "partial", InputTokens: number(5), CacheReadTokens: number(3)}
		}
		_, e = proc.FinishAttempt(ctx, processing.AttemptFinish{RunID: run.ID, Number: attempt.Number, FinishedAt: at.Add(time.Second), Outcome: "failed", Usage: usage, PriceVersionID: priceID, ErrorCode: "provider_temporary"})
		must(e)
		must(queue.Fail(ctx, claim, jobs.Failure{Code: "provider_temporary", Detail: "must-not-appear", Temporary: true}, at.Add(time.Second)))
	}
	// Distinct steady-state OCR units and embedding costs must stay separate.
	for _, stage := range []string{"ocr", "embedding"} {
		c := config
		c.ID = stage
		c.Name = stage
		c.Stage = stage
		must(proc.RegisterConfiguration(ctx, c))
		r, e := proc.StartRun(ctx, processing.RunRequest{IdempotencyKey: stage, Workload: "ordinary", Stage: stage, ConfigurationVersionID: c.ID, SourceID: &source, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{}`), CreatedAt: now})
		must(e)
		a, e := proc.StartAttempt(ctx, processing.AttemptStart{RunID: r.ID, StartedAt: now})
		must(e)
		units := map[string]int64{"pages": 2}
		_, e = proc.FinishAttempt(ctx, processing.AttemptFinish{RunID: r.ID, Number: a.Number, FinishedAt: now.Add(2 * time.Second), Outcome: "succeeded", Usage: processing.Usage{Status: "partial", OtherUnits: units}})
		must(e)
	}
	// A completed irrelevant document is not an uninterpreted notice.
	catalog, err := classification.RegisterCatalog(ctx, proc, "fixture", "qwen3.8-27b", now)
	must(err)
	runner := &classification.Runner{Documents: retained, OCR: ocr.NewStore(pool), Processing: proc, Results: classification.NewStore(pool), Adapter: &lifecycleAdapter{}, Model: "qwen3.8-27b", ConfigurationVersion: catalog.ConfigurationVersionID}
	irrelevantJob, err := classification.Enqueue(ctx, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, classification.Payload{DocumentVersionID: irrelevant.ID, Workload: "ordinary"}, now)
	must(err)
	worker := jobs.Worker{Store: queue, Queue: irrelevantJob.Queue, ID: "classifier", Lease: time.Minute, PollInterval: 10 * time.Millisecond, Handlers: map[string]jobs.Handler{classification.Kind: runner.Handler()}}
	_, err = worker.RunOne(ctx)
	must(err)
	uninterpretedBeforeSuspension, err := operations.New(pool).Page(ctx, "documents", operations.Filter{SourceID: "s", Page: 1}, time.Now())
	must(err)
	if len(uninterpretedBeforeSuspension.Table.Rows) != 1 || uninterpretedBeforeSuspension.Table.Rows[0]["document_version_id"] != version.ID {
		t.Fatalf("irrelevant document requires extraction: %#v", uninterpretedBeforeSuspension)
	}
	// Preserve a quality finding with an evidence reference, independently of publication.
	must(reg.SuspendInterpretation(ctx, "s", 1, "operator", ev))
	scheduler := interpretation.New(pool, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, false)
	reports := operations.New(pool)
	initial, initialErr := reports.Overview(ctx, time.Now())
	must(initialErr)
	if initial.FailedJobs != 1 || initial.PendingDocuments != 3 {
		t.Fatalf("attempts duplicated overview: %#v", initial)
	}
	host := httptest.NewServer(HandlerWithAdministration(nil, AdminRuntime{Registry: reg, Interpretation: scheduler, Operations: reports, Jobs: queue}))
	defer host.Close()
	get := func(path string, html bool, want int) []byte {
		t.Helper()
		req, e := http.NewRequest("GET", host.URL+path, nil)
		must(e)
		if !html {
			req.Header.Set("Accept", "application/json")
		}
		res, e := http.DefaultClient.Do(req)
		must(e)
		defer res.Body.Close()
		b, e := io.ReadAll(res.Body)
		must(e)
		if res.StatusCode != want {
			t.Fatalf("%s: %d want %d: %s", path, res.StatusCode, want, b)
		}
		if bytes.Contains(b, []byte("must-not-appear")) {
			t.Fatalf("private data exposed in %s", path)
		}
		return b
	}
	load := func(section, query string) operations.Report {
		t.Helper()
		var r operations.Report
		must(json.Unmarshal(get("/admin/operations/"+section+query, false, 200), &r))
		return r
	}
	for _, section := range []string{"sources", "jobs", "documents", "findings", "usage"} {
		get("/admin/operations/"+section, true, 200)
		load(section, "")
	}
	sources := load("sources", "?source_id=s")
	if len(sources.Table.Rows) != 1 || sources.Table.Rows[0]["updating_state"] != "delayed" || sources.Table.Rows[0]["last_error_code"] != "invalid_content" {
		t.Fatalf("source state: %#v", sources)
	}
	draft := load("sources", "?source_id=draft")
	if draft.Table.Rows[0]["updating_state"] != "not_yet_verified" {
		t.Fatal("draft presented as verified")
	}
	if got := load("jobs", "?source_id=s&state=queued"); len(got.Table.Rows) != 0 {
		t.Fatal("state filter broadened")
	}
	get("/admin/operations/jobs?state=unknown", false, 400)
	get("/admin/operations/sources?state=failed", false, 400)
	failed := load("jobs", "?source_id=s&state=failed")
	if len(failed.Table.Rows) != 1 || failed.Table.Rows[0]["attempt_count"] != float64(3) || failed.Table.Rows[0]["measured_duration_ms"] != float64(3000) {
		t.Fatalf("failed jobs: %#v", failed)
	}
	findings := load("findings", "?source_id=s")
	if len(findings.Table.Rows) != 1 || findings.Table.Rows[0]["kind"] != "source_defect" {
		t.Fatalf("findings: %#v", findings)
	}
	// Suspended sources expose all affected versions; an unrelated source remains separate.
	docs := load("documents", "?source_id=s")
	if len(docs.Table.Rows) != 2 {
		t.Fatalf("suspension documents: %#v", docs)
	}
	docs = load("documents", fmt.Sprintf("?document_version_id=%d", other.ID))
	if len(docs.Table.Rows) != 1 || docs.Table.Rows[0]["classification_status"] != "not_processed" {
		t.Fatalf("unprocessed: %#v", docs)
	}
	usage := load("usage", fmt.Sprintf("?run_id=%d", run.ID))
	if len(usage.Table.Rows) != 3 || usage.Table.Rows[0]["pricing_provenance_url"] != price.ProvenanceURL || usage.Table.Rows[2]["input_tokens"] != nil {
		t.Fatalf("usage details: %#v", usage)
	}
	var count, retries, duration, knownCost, unknown float64
	for _, row := range usage.Totals.Rows {
		count += row["attempts"].(float64)
		retries += row["retries"].(float64)
		duration += row["measured_duration_ms"].(float64)
		unknown += row["unavailable_usage_attempts"].(float64)
		if v := row["known_cost_microunits"]; v != nil {
			knownCost += v.(float64)
		}
	}
	if count != 3 || retries != 2 || duration != 3000 || knownCost != 26 || unknown != 2 {
		t.Fatalf("totals %v %v %v %v %v", count, retries, duration, knownCost, unknown)
	}
	all := load("usage", "?source_id=s")
	if len(all.UnavailableCostCategories) != 3 || len(all.Units.Rows) != 2 {
		t.Fatalf("missing distinct infrastructure or OCR units: %#v", all)
	}
	stages := map[string]bool{}
	for _, row := range all.Totals.Rows {
		stages[row["workload"].(string)+":"+row["stage"].(string)] = true
	}
	for _, key := range []string{"bootstrap:classification", "ordinary:ocr", "ordinary:embedding"} {
		if !stages[key] {
			t.Fatalf("missing %s", key)
		}
	}
	for _, query := range []string{"?page=0", "?page=2&page=3", "?document_version_id=bad", "?unknown=1", "?run_id=0"} {
		get("/admin/operations/usage"+query, false, 400)
	}
	get("/admin/operations/sources?document_version_id=1", false, 400)
	post := func(path string, value any, want int) []byte {
		t.Helper()
		b, e := json.Marshal(value)
		must(e)
		res, e := http.Post(host.URL+path, "application/json", bytes.NewReader(b))
		must(e)
		defer res.Body.Close()
		b, e = io.ReadAll(res.Body)
		must(e)
		if res.StatusCode != want {
			t.Fatalf("POST %s: %d want %d %s", path, res.StatusCode, want, b)
		}
		return b
	}
	path := fmt.Sprintf("/admin/jobs/%d/relaunch", job.ID)
	post(path, map[string]any{"actor": "operator", "after_attempt": 2}, 409)
	formResponse, e := http.Post(host.URL+path, "application/x-www-form-urlencoded", strings.NewReader("actor=operator&after_attempt=2"))
	must(e)
	formResponse.Body.Close()
	if formResponse.StatusCode != 409 {
		t.Fatal("stale guided attempt accepted")
	}
	post(path, map[string]any{"actor": "", "after_attempt": 3}, 400)
	post(path, map[string]any{"actor": "operator", "after_attempt": 3, "payload": map[string]any{}}, 400)
	post(path, map[string]any{"actor": "operator", "after_attempt": 3}, 200)
	// Duplicate submissions and concurrent operators cannot add multiple budgets.
	var wg sync.WaitGroup
	failures := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); failures <- queue.Relaunch(ctx, job.ID, 3, "operator", time.Now()) }()
	}
	wg.Wait()
	close(failures)
	for e := range failures {
		must(e)
	}
	replay, err := queue.Enqueue(ctx, request)
	must(err)
	if replay.ID != job.ID || replay.MaxAttempts != 6 || replay.Attempt != 3 {
		t.Fatalf("relaunch changed identity or budget: %#v", replay)
	}
	claim, err := queue.Claim(ctx, "fixture", "relaunch-worker", time.Now().UTC(), time.Minute)
	must(err)
	if claim.Attempt != 4 {
		t.Fatalf("history renumbered: %d", claim.Attempt)
	}
	failedAt := time.Now().UTC().Truncate(time.Microsecond)
	must(queue.Fail(ctx, claim, jobs.Failure{Code: "provider_temporary", Detail: "temporary", Temporary: true}, failedAt))
	retry := load("jobs", "")
	if retry.Table.Rows[0]["state"] != "retry_wait" {
		t.Fatal("transient retry hidden")
	}
	var available time.Time
	must(pool.QueryRow(ctx, "SELECT available_at FROM processing_jobs WHERE id=$1", job.ID).Scan(&available))
	if !available.Equal(failedAt.Add(time.Second)) {
		t.Fatal("relaunch did not restart backoff")
	}
	claim, err = queue.Claim(ctx, "fixture", "relaunch-worker", available, time.Minute)
	must(err)
	must(queue.Complete(ctx, claim, jobs.Result{Payload: json.RawMessage(`{}`), Effects: []jobs.Effect{{Key: "one-effect", Kind: "fixture", Payload: json.RawMessage(`{}`)}}}, available.Add(time.Second)))
	post(path, map[string]any{"actor": "operator", "after_attempt": 5}, 409)
	attempts, err := queue.Attempts(ctx, job.ID)
	must(err)
	if len(attempts) != 5 {
		t.Fatal("attempt history lost")
	}
	var audit int
	must(pool.QueryRow(ctx, "SELECT count(*) FROM processing_job_relaunches WHERE job_id=$1", job.ID).Scan(&audit))
	completed := load("jobs", fmt.Sprintf("?job_id=%d", job.ID))
	if len(completed.Table.Rows) != 1 || completed.Table.Rows[0]["state"] != "succeeded" || len(completed.Table.Rows[0]["relaunches"].([]any)) != 1 {
		t.Fatalf("relaunch audit unavailable: %#v", completed)
	}
	if audit != 1 {
		t.Fatal("duplicate relaunch audit")
	}
	// Existing scoped reprocessing controls remain available from operational pages.
	result := post("/admin/sources/s/reprocess-interpretation", map[string]any{"actor": "operator", "revision": 1, "error_code": "provider_temporary"}, 200)
	var selection struct {
		ID    string `json:"reprocessing_id"`
		Count int    `json:"selected_versions"`
	}
	must(json.Unmarshal(result, &selection))
	if selection.Count != 1 {
		t.Fatalf("reprocessing broadened: %s", result)
	}
	var selected int64
	must(pool.QueryRow(ctx, "SELECT document_version_id FROM interpretation_reprocessing_selections WHERE request_id=$1", selection.ID).Scan(&selected))
	if selected != version.ID {
		t.Fatal("wrong selected version")
	}
	// The same snapshot totals cover all pages, not just the displayed 100 rows.
	for i := 0; i < 100; i++ {
		a, e := proc.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, StartedAt: now})
		must(e)
		_, e = proc.FinishAttempt(ctx, processing.AttemptFinish{RunID: run.ID, Number: a.Number, FinishedAt: now.Add(time.Second), Outcome: "failed", Usage: processing.Usage{Status: "unavailable"}, ErrorCode: "fixture"})
		must(e)
	}
	first := load("usage", fmt.Sprintf("?run_id=%d", run.ID))
	second := load("usage", fmt.Sprintf("?run_id=%d&page=2", run.ID))
	if len(first.Table.Rows) != 100 || !first.More || len(second.Table.Rows) != 3 || second.More {
		t.Fatal("pagination lost attempts")
	}
	a, _ := json.Marshal(first.Totals)
	b, _ := json.Marshal(second.Totals)
	if !bytes.Equal(a, b) {
		t.Fatal("totals depend on page")
	}
	// Overview counts entities across the full scope, not attempts or a page.
	for i := 0; i < 101; i++ {
		retain(fmt.Sprintf("overview-%d", i), "other")
	}
	overview, e := reports.Overview(ctx, time.Now())
	must(e)
	if overview.SourceIssues != 1 || overview.FailedJobs != 0 || overview.PendingDocuments != 104 {
		t.Fatalf("overview: %#v", overview)
	}
	issues := load("sources", "?issue=attention")
	if len(issues.Table.Rows) != 1 || issues.Table.Rows[0]["source_id"] != "s" {
		t.Fatal("issue drilldown mismatch")
	}
	get("/admin/operations/sources?issue=unknown", false, 400)
	// Archival clears operational failures while preserving linked accounting.
	archivedJob, e := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "archive-fixture", Kind: "fixture", IdempotencyKey: "archive-failed", Payload: payload, MaxAttempts: 3, RetryBase: time.Second, AvailableAt: now})
	must(e)
	archivedClaim, e := queue.Claim(ctx, "archive-fixture", "fixture", now, time.Minute)
	must(e)
	paid, e := proc.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, QueueJobID: &archivedClaim.ID, QueueAttemptNumber: &archivedClaim.Attempt, StartedAt: now})
	must(e)
	_, e = proc.FinishAttempt(ctx, processing.AttemptFinish{RunID: paid.RunID, Number: paid.Number, FinishedAt: now.Add(time.Second), Outcome: "failed", Usage: processing.Usage{Status: "reported", InputTokens: number(31), OutputTokens: number(6)}, ErrorCode: "fixture"})
	must(e)
	must(queue.Fail(ctx, archivedClaim, jobs.Failure{Code: "fixture", Temporary: false}, now.Add(time.Second)))
	usageBefore, _ := json.Marshal(load("usage", "").Totals)
	archiveAt := time.Now().UTC().Add(time.Minute)
	if _, e = scheduler.ArchivePending(ctx, archiveAt, "operator", archiveAt); e == nil {
		t.Fatal("pending reset allowed active queue work")
	}
	counts, e := queue.ArchiveBefore(ctx, archiveAt, "operator", archiveAt)
	must(e)
	if counts["failed"] != 1 {
		t.Fatalf("failed archival scope: %v", counts)
	}
	active := load("jobs", "")
	if len(active.Table.Rows) != 2 || active.Table.Rows[0]["state"] != "succeeded" || active.Table.Rows[1]["state"] != "succeeded" {
		t.Fatalf("archive visible or success hidden: %#v", active.Table.Rows)
	}
	for _, query := range []string{"?state=failed", "?state=queued", fmt.Sprintf("?job_id=%d", archivedJob.ID)} {
		if len(load("jobs", query).Table.Rows) != 0 {
			t.Fatalf("archive exposed in %s", query)
		}
	}
	afterArchive, e := reports.Overview(ctx, time.Now())
	must(e)
	if afterArchive.FailedJobs != 0 || afterArchive.SourceIssues != overview.SourceIssues || afterArchive.PendingDocuments != overview.PendingDocuments {
		t.Fatalf("archive hid genuine source/document issues: %#v", afterArchive)
	}
	usageAfter, _ := json.Marshal(load("usage", "").Totals)
	if !bytes.Equal(usageBefore, usageAfter) {
		t.Fatal("archival changed historical token accounting")
	}
	// A pending-version reset preserves valid interpretations and blocks old work.
	resetAt := time.Now().UTC()
	archivedVersions, e := scheduler.ArchivePending(ctx, resetAt, "operator", resetAt)
	must(e)
	if archivedVersions < 100 {
		t.Fatalf("pending backlog not archived: %d", archivedVersions)
	}
	again, e := scheduler.ArchivePending(ctx, resetAt, "operator", resetAt)
	must(e)
	if again != 0 {
		t.Fatal("pending archival not idempotent")
	}
	validArchived, e := scheduler.Archived(ctx, irrelevant.ID)
	must(e)
	if validArchived {
		t.Fatal("successful interpretation archived")
	}
	if len(load("documents", "?source_id=other").Table.Rows) != 0 {
		t.Fatal("archived pending documents still visible")
	}
	scheduled, e := scheduler.Automatic(ctx, interpretation.AcquisitionEvent{DocumentVersionID: other.ID, EvidenceHash: other.Hash, Workload: "ordinary", ContentChanged: true, At: resetAt})
	must(e)
	if scheduled {
		t.Fatal("archived version rescheduled")
	}
	scheduled, e = scheduler.AfterOCR(ctx, other.ID, "ordinary", resetAt)
	must(e)
	if scheduled {
		t.Fatal("archived OCR completion scheduled classification")
	}
	oldPayload, _ := json.Marshal(map[string]any{"document_version_id": other.ID})
	called := false
	guarded := scheduler.Guard(func(context.Context, jobs.Job) (jobs.Result, error) { called = true; return jobs.Result{}, nil })
	_, e = guarded(ctx, jobs.Job{Payload: oldPayload})
	must(e)
	if called {
		t.Fatal("archived version reached inference handler")
	}
	sameURL := "https://source.example/notices/other"
	updated, e := retained.Retain(ctx, documents.Acquisition{ID: "post-reset-version", SourceID: "other", Configuration: 1, URL: sameURL, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: sameURL, Role: "original", Required: true, SourceID: "other", Configuration: 1, MediaType: "text/html", Bytes: []byte("Nuovo aggiornamento dopo il reset")}}})
	must(e)
	if updated.ID == other.ID {
		t.Fatal("fixture did not create a new version")
	}
	scheduled, e = scheduler.Automatic(ctx, interpretation.AcquisitionEvent{DocumentVersionID: updated.ID, EvidenceHash: updated.Hash, Workload: "ordinary", ContentChanged: true, At: time.Now().UTC()})
	must(e)
	if !scheduled {
		t.Fatal("new version blocked by archive")
	}
	pendingNow := load("documents", "?source_id=other")
	if len(pendingNow.Table.Rows) != 1 || pendingNow.Table.Rows[0]["document_version_id"] != float64(updated.ID) {
		t.Fatalf("new version absent from pending view: %#v", pendingNow.Table.Rows)
	}
	usageAfter, _ = json.Marshal(load("usage", "").Totals)
	if !bytes.Equal(usageBefore, usageAfter) {
		t.Fatal("document archival changed accounting")
	}
	// Two raw versions remain inspectable but share one pending representative.
	copyVersion, e := retained.Retain(ctx, documents.Acquisition{ID: "equivalent-copy", SourceID: "other", Configuration: 1, URL: sameURL, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: sameURL, Role: "original", Required: true, SourceID: "other", Configuration: 1, MediaType: "text/html", Bytes: []byte("Nuovo aggiornamento dopo il reset<!-- fixture noise -->")}}})
	must(e)
	_, e = pool.Exec(ctx, `INSERT INTO interpretation_preflight(document_version_id,representative_version_id,configuration,fingerprint,body,complete) VALUES($1,$1,'fixture',repeat('a',64),'{}',true),($2,$1,'fixture',repeat('a',64),'{}',true)`, updated.ID, copyVersion.ID)
	must(e)
	grouped := load("documents", "?source_id=other")
	if len(grouped.Table.Rows) != 1 || grouped.Table.Rows[0]["equivalent_versions"] != float64(1) {
		t.Fatalf("grouping lost: %#v", grouped)
	}
	detail := load("documents", fmt.Sprintf("?document_version_id=%d", copyVersion.ID))
	if len(detail.Table.Rows) != 1 || detail.Table.Rows[0]["equivalent_to_version"] != float64(updated.ID) {
		t.Fatalf("raw copy hidden: %#v", detail)
	}
	// Verified discovery sections are evidence, not unscheduled LLM documents.
	listingCfg := cfg
	listingCfg.Discovery = registry.Discovery{PaginationParameter: "page", DocumentPathPrefixes: []string{"/notices/"}, MaxPagesPerSection: 10, MaxDocuments: 10, ListingContentMarkers: []string{"<main>"}}
	must(reg.CreateSource(ctx, registry.Source{ID: "calcinaia-municipal", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "050004"}, listingCfg, "fixture"))
	makeListing := func(id, u string) documents.Version {
		t.Helper()
		v, e := retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: "calcinaia-municipal", Configuration: 1, URL: u, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: u, Role: "original", Required: true, SourceID: "calcinaia-municipal", Configuration: 1, MediaType: "text/html", Bytes: []byte("<main>Elenco avvisi</main>")}}})
		must(e)
		return v
	}
	listing := makeListing("listing", "https://source.example/notices?page=1")
	baseListing := makeListing("listing-base", "https://source.example/notices")
	unknownListing := makeListing("listing-unknown", "https://source.example/notices?page=1&meaningful=2")
	listPending := load("documents", "?source_id=calcinaia-municipal")
	if len(listPending.Table.Rows) != 1 || listPending.Table.Rows[0]["document_version_id"] != float64(unknownListing.ID) {
		t.Fatalf("discovery count incorrect: %#v", listPending)
	}
	listingDetail := load("documents", fmt.Sprintf("?document_version_id=%d", listing.ID))
	if len(listingDetail.Table.Rows) != 1 || listingDetail.Table.Rows[0]["evidence_role"] != "discovery_listing" {
		t.Fatal("listing provenance hidden")
	}
	_, e = acquisition.NewTrackingStore(pool).Remember(ctx, "calcinaia-municipal", 1, []acquisition.DiscoveredDocument{{URL: "https://source.example/notices?page=1"}}, time.Now())
	must(e)
	_, e = scheduler.Automatic(ctx, interpretation.AcquisitionEvent{DocumentVersionID: baseListing.ID, EvidenceHash: baseListing.Hash, ContentChanged: true, At: time.Now()})
	must(e)
	if got := load("documents", "?source_id=calcinaia-municipal"); len(got.Table.Rows) != 3 {
		t.Fatalf("target/trigger exceptions hidden: %#v", got)
	}
	// Database failure cannot masquerade as an empty healthy dashboard.
	pool.Close()
	body := get("/admin/operations/usage", false, 503)
	if !strings.Contains(string(body), "administration_unavailable") {
		t.Fatalf("unexpected storage failure: %s", body)
	}
}
