//go:build integration

package interpretation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/evaluation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

type archiveRecoveryAdapter struct{ extract bool }

func (a archiveRecoveryAdapter) Name() string { return "fixture" }
func (a archiveRecoveryAdapter) Complete(context.Context, inference.Request) (inference.Response, error) {
	body := `{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"Il sottopasso di via del Bosco è stato riaperto alla circolazione."}`
	if a.extract {
		body = `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["Il sottopasso di via del Bosco è stato riaperto alla circolazione."],"measures":[]}`
	}
	return inference.Response{ID: "synthetic-recovery", Model: "qwen3.8-27b", Content: body}, nil
}

func TestSelectedArchiveRecoveryPreservesHistoryAndBoundsDescendants(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := interpretationTestDB(t)
	must := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate, Migrate} {
		must(m(ctx, pool))
	}
	reg := registry.New(pool)
	now := time.Now().UTC()
	proof := registry.Evidence{URL: "https://source.example/notice", Locator: "synthetic recovery comparison", ObservedAt: now}
	for _, id := range []string{"source", "foreign"} {
		must(reg.CreateAuthority(ctx, registry.Authority{ID: id, Name: id, OfficialURL: proof.URL}))
		must(reg.CreateChannel(ctx, registry.Channel{ID: id, PublisherID: id, Platform: "fixture", URL: proof.URL}))
		must(reg.CreateSource(ctx, registry.Source{ID: id, AuthorityID: id, ChannelID: id, ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: proof.URL, Sections: []string{proof.URL}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &proof, CollectionPermitted: true, RetentionPermitted: true}}, "operator"))
		must(reg.RecordPreview(ctx, id, 1, "operator", proof))
		must(reg.EnableCollection(ctx, id, 1, "operator"))
	}
	docs := documents.New(pool, &interpretationObjects{})
	retain := func(key, source string, complete bool) int64 {
		resources := []documents.Resource{{URL: proof.URL + "/" + key, Role: "original", Required: true, SourceID: source, Configuration: 1, MediaType: "text/html", Bytes: []byte("<main>Il sottopasso di via del Bosco è stato riaperto alla circolazione.</main>")}}
		if !complete {
			resources = append(resources, documents.Resource{URL: proof.URL + "/missing.pdf", Role: "attachment", Required: true, SourceID: source, Configuration: 1, Missing: "unavailable"})
		}
		v, e := docs.Retain(ctx, documents.Acquisition{ID: key, SourceID: source, Configuration: 1, URL: proof.URL + "/" + key, Metadata: json.RawMessage(`{}`), Resources: resources})
		must(e)
		_, e = pool.Exec(ctx, "INSERT INTO interpretation_archives VALUES($1,now(),now(),'historical-operator')", v.ID)
		must(e)
		return v.ID
	}
	a, b, foreign, incomplete := retain("selected", "source", true), retain("unselected", "source", true), retain("foreign", "foreign", true), retain("incomplete", "source", false)
	proof.ObservedAt = time.Now().UTC()
	run := evaluation.Run{ID: "review-pass", SourceID: "source", Revision: 1, Suite: "recovery", CorpusSHA256: strings.Repeat("a", 64), Reviewer: "synthetic reviewer", ReviewedAt: now.Format("2006-01-02"), ProcessVersion: "fixture-v25", Mode: "service", StartedAt: now, FinishedAt: time.Now().UTC()}
	for _, stage := range []string{"discovery", "ocr", "classification", "extraction", "linking"} {
		run.Checks = append(run.Checks, evaluation.Check{CaseID: "fixture", Kind: "synthetic", Stage: stage, Field: "result", Expected: "verified", Actual: "verified", Evidence: "synthetic input/output", Outcome: "pass", MeasureField: stage == "extraction"})
	}
	must(reg.RecordRegression(ctx, "source", 1, "operator", "", run))
	queue := jobs.New(pool)
	s := New(pool, queue, inference.RetryPolicy{MaxAttempts: 1, BaseDelay: time.Second}, false)
	request := ArchiveRecovery{ID: "recover-selected", SourceID: "source", Revision: 1, Actor: "operator", VersionIDs: []int64{a}, Regression: run.ID, Evidence: proof}
	for _, ids := range [][]int64{{a, foreign}, {a, incomplete}, {a, a}, {a, 999999}} {
		bad := request
		bad.ID = fmt.Sprint("invalid-", ids)
		bad.VersionIDs = ids
		if _, e := s.RecoverArchived(ctx, bad); e == nil {
			t.Fatalf("invalid scope admitted: %v", ids)
		}
	}
	var count int
	must(pool.QueryRow(ctx, "SELECT count(*) FROM interpretation_archive_recoveries").Scan(&count))
	if count != 0 {
		t.Fatal("invalid request left audit or selections")
	}
	bad := request
	bad.Revision = 2
	if _, e := s.RecoverArchived(ctx, bad); !errors.Is(e, registry.ErrConflict) {
		t.Fatalf("stale revision: %v", e)
	}
	bad = request
	bad.Regression = "missing"
	if _, e := s.RecoverArchived(ctx, bad); !errors.Is(e, registry.ErrPrerequisite) {
		t.Fatalf("missing regression: %v", e)
	}
	// Persisted authorization survives a queue outage; identical retry resumes it.
	unavailable, e := pgxpool.NewWithConfig(ctx, pool.Config())
	must(e)
	unavailable.Close()
	s.Queue = jobs.New(unavailable)
	if _, e = s.RecoverArchived(ctx, request); e == nil {
		t.Fatal("queue outage was hidden")
	}
	must(pool.QueryRow(ctx, "SELECT count(*) FROM interpretation_archive_recoveries").Scan(&count))
	if count != 1 {
		t.Fatal("queue outage lost durable authorization")
	}
	s.Queue = queue
	recovered, e := s.RecoverArchived(ctx, request)
	must(e)
	repeated, e := s.RecoverArchived(ctx, request)
	must(e)
	if !reflect.DeepEqual(recovered, repeated) {
		t.Fatal("retry changed audit")
	}
	bad = request
	bad.VersionIDs = []int64{b}
	if _, e := s.RecoverArchived(ctx, bad); !errors.Is(e, registry.ErrConflict) {
		t.Fatalf("changed identity: %v", e)
	}
	must(pool.QueryRow(ctx, "SELECT count(*) FROM processing_jobs").Scan(&count))
	if count != 1 {
		t.Fatalf("retry created %d jobs", count)
	}
	for _, q := range []string{"UPDATE interpretation_archive_recoveries SET revision=2", "UPDATE interpretation_reprocessing_requests SET actor='rewritten'", "DELETE FROM interpretation_archive_recovery_versions", "UPDATE interpretation_archives SET actor='rewritten'"} {
		if _, e = pool.Exec(ctx, q); e == nil {
			t.Fatal("immutable audit changed")
		}
	}
	// Automatic scheduling and unrelated requests still observe the archival marker.
	scheduled, e := s.Automatic(ctx, AcquisitionEvent{DocumentVersionID: a, EvidenceHash: strings.Repeat("a", 64), ContentChanged: true, At: time.Now().UTC()})
	must(e)
	if scheduled {
		t.Fatal("recovery enabled automatic scheduling")
	}
	hits := 0
	probe := s.Guard(func(context.Context, jobs.Job) (jobs.Result, error) { hits++; return jobs.Result{}, nil })
	payload := func(id int64, workload, selection string) json.RawMessage {
		raw, _ := json.Marshal(classification.Payload{DocumentVersionID: id, Workload: workload, SelectionID: selection})
		return raw
	}
	for _, job := range []jobs.Job{{Kind: classification.Kind, Payload: payload(a, "ordinary", request.ID)}, {Kind: classification.Kind, Payload: payload(a, "reprocessing", "other-request")}, {Kind: classification.Kind, Payload: payload(b, "reprocessing", request.ID)}} {
		_, e = probe(ctx, job)
		must(e)
	}
	if hits != 0 {
		t.Fatal("unrelated archived job reached handler")
	}
	// Execute classification through the actual runner and durable attempt binding.
	proc := processing.New(pool)
	catalog, e := classification.RegisterCatalog(ctx, proc, "openai-chat", "qwen3.8-27b", now)
	must(e)
	runner := classification.Runner{Documents: docs, OCR: ocr.NewStore(pool), Processing: proc, Results: classification.NewStore(pool), Adapter: archiveRecoveryAdapter{}, Model: "qwen3.8-27b", ConfigurationVersion: catalog.ConfigurationVersionID}
	job, e := queue.Claim(ctx, "inference", "recovery-worker", time.Now().UTC(), time.Minute)
	must(e)
	result, e := s.Guard(runner.Handler())(ctx, job.Job)
	must(e)
	must(queue.Complete(ctx, job, result, time.Now().UTC()))
	var output struct {
		RunID int64 `json:"run_id"`
	}
	must(json.Unmarshal(result.Payload, &output))
	if output.RunID < 1 {
		t.Fatalf("guard skipped selected classification: %s", result.Payload)
	}
	raw, _ := json.Marshal(extraction.Payload{DocumentVersionID: a, ClassificationRunID: output.RunID, Workload: "reprocessing"})
	_, e = probe(ctx, jobs.Job{Kind: extraction.Kind, Payload: raw})
	must(e)
	if hits != 1 {
		t.Fatal("selected extraction blocked")
	}
	// A classification job cannot borrow the descendant-only authorization.
	_, e = probe(ctx, jobs.Job{Kind: classification.Kind, Payload: raw})
	must(e)
	if hits != 1 {
		t.Fatal("classification borrowed descendant authorization")
	}
	ec, e := extraction.RegisterCatalog(ctx, proc, "openai-chat", "qwen3.8-27b", now)
	must(e)
	erunner := extraction.Runner{Documents: docs, OCR: ocr.NewStore(pool), Classifications: classification.NewStore(pool), Processing: proc, Results: extraction.NewStore(pool), Adapter: archiveRecoveryAdapter{extract: true}, Model: "qwen3.8-27b", ConfigurationVersion: ec.ConfigurationVersionID}
	_, e = extraction.Enqueue(ctx, queue, s.Policy, extraction.Payload{DocumentVersionID: a, ClassificationRunID: output.RunID, Workload: "reprocessing"}, time.Now().UTC())
	must(e)
	ejob, e := queue.Claim(ctx, "inference", "recovery-worker", time.Now().UTC(), time.Minute)
	must(e)
	eresult, e := s.Guard(erunner.Handler())(ctx, ejob.Job)
	must(e)
	must(queue.Complete(ctx, ejob, eresult, time.Now().UTC()))
	var eoutput struct {
		RunID int64 `json:"run_id"`
	}
	must(json.Unmarshal(eresult.Payload, &eoutput))
	er, found, e := extraction.NewStore(pool).Get(ctx, eoutput.RunID)
	must(e)
	if !found || er.Status != "extracted" || len(er.Measures) != 1 || er.Measures[0].Kind != "reopening" {
		t.Fatalf("recovery extraction missing: %+v", er)
	}
	raw, _ = json.Marshal(linking.Payload{ExtractionRunID: er.RunID, MeasureOrdinal: 1, Workload: "reprocessing"})
	_, e = probe(ctx, jobs.Job{Kind: linking.Kind, Payload: raw})
	must(e)
	if hits != 2 {
		t.Fatal("selected linking descendant blocked")
	}
	archived, e := s.Archived(ctx, a)
	must(e)
	if !archived {
		t.Fatal("archive marker removed")
	}
	history, e := s.ArchiveRecoveries(ctx, "source")
	must(e)
	if len(history) != 1 || history[0].Actor != "operator" {
		t.Fatal("audit readback missing")
	}
	// A later failed regression prevents granting another recovery.
	time.Sleep(time.Millisecond)
	failed := run
	failed.ID = "review-failed"
	failed.StartedAt = time.Now().UTC()
	failed.FinishedAt = failed.StartedAt
	failed.Checks = append([]evaluation.Check(nil), run.Checks...)
	failed.Checks[3].Outcome = "omission"
	must(reg.RecordRegression(ctx, "source", 1, "operator", run.ID, failed))
	bad = request
	bad.ID = "after-failure"
	bad.VersionIDs = []int64{b}
	if _, e := s.RecoverArchived(ctx, bad); !errors.Is(e, registry.ErrPrerequisite) {
		t.Fatalf("stale pass bypassed latest failure: %v", e)
	}
}
