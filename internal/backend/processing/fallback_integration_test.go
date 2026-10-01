//go:build integration

package processing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLocalFallbackPreservesHoldsScopeAndAttemptBudget(t *testing.T) {
	ctx := context.Background()
	pool := processingTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	_, err := pool.Exec(ctx, `INSERT INTO registry_authorities VALUES ('a','a','https://example.org');
 INSERT INTO registry_channels VALUES ('c','a','web','https://example.org',false);
 INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES ('selected','a','c','municipal','test'),('outside','a','c','municipal','test');
 INSERT INTO retained_documents(source_id,official_url) VALUES ('selected','https://example.org/one'),('outside','https://example.org/two');
 INSERT INTO retained_versions(document_id,content_hash,first_acquired_at,complete,metadata) VALUES (1,repeat('a',64),now(),true,'{}'),(2,repeat('b',64),now(),true,'{}');`)
	if err != nil {
		t.Fatal(err)
	}
	s := New(pool)
	queue := jobs.New(pool)
	now := time.Now().UTC()
	policy := DefaultGatePolicy()
	for _, scope := range []string{"remote", "local"} {
		if _, err = s.AcquireGate(ctx, scope, "chat", now, policy); err != nil {
			t.Fatal(err)
		}
	}
	binding := GateBinding{Kind: "classify_relevance", Scope: "remote", Model: "chat", Fallback: &FallbackBinding{Scope: "local", Model: "chat", Sources: []string{"selected"}}}
	makeJob := func(version int64, kind string) jobs.Job {
		payload, _ := json.Marshal(map[string]int64{"document_version_id": version})
		job, e := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "inference", Kind: kind, IdempotencyKey: fmt.Sprintf("%s-%d", kind, version), Payload: payload, MaxAttempts: 3, RetryBase: time.Millisecond, AvailableAt: now})
		if e != nil {
			t.Fatal(e)
		}
		job.Attempt = 1
		return job
	}
	selected, outside := makeJob(1, binding.Kind), makeJob(2, binding.Kind)
	primaryCalls, localCalls := 0, 0
	primary := func(context.Context, jobs.Job) (jobs.Result, error) {
		primaryCalls++
		return jobs.Result{Payload: json.RawMessage(`{"status":"classified"}`)}, nil
	}
	local := func(context.Context, jobs.Job) (jobs.Result, error) {
		localCalls++
		return jobs.Result{Payload: json.RawMessage(`{"status":"classified"}`)}, nil
	}
	handler := s.FallbackHandler(primary, local, binding)
	if _, err = handler(ctx, selected); err != nil || primaryCalls != 1 || localCalls != 0 {
		t.Fatal("healthy primary not preferred", err)
	}
	hold := func(scope, category string) {
		permit, e := s.AcquireGate(ctx, scope, "chat", now, policy)
		if e != nil {
			t.Fatal(e)
		}
		if e = s.RecordGate(ctx, permit, category, 0, now, policy); e != nil {
			t.Fatal(e)
		}
	}
	hold("remote", "quota")
	if _, err = handler(ctx, selected); err != nil || primaryCalls != 1 || localCalls != 1 {
		t.Fatal("held remote was called", err)
	}
	if got, e := s.ShouldFallback(ctx, binding, outside, now); e != nil || got {
		t.Fatal("source boundary escaped", e)
	}
	if err = s.DeferHeldJobs(ctx, []GateBinding{binding}, now); err != nil {
		t.Fatal(err)
	}
	check := func(job jobs.Job, blocked bool) {
		var at time.Time
		var attempts, max int
		if e := pool.QueryRow(ctx, `SELECT available_at,attempt_count,max_attempts FROM processing_jobs WHERE id=$1`, job.ID).Scan(&at, &attempts, &max); e != nil || at.After(now) != blocked || attempts != 0 || max != 3 {
			t.Fatal("pre-claim scope/attempt changed", blocked, e, at, attempts, max)
		}
	}
	check(selected, false)
	check(outside, true)
	hold("local", "authentication")
	if err = s.DeferHeldJobs(ctx, []GateBinding{binding}, now); err != nil {
		t.Fatal(err)
	}
	check(selected, true)
	if err = s.ResumeGate(ctx, "remote", "*", "test", now); err != nil {
		t.Fatal(err)
	}
	if got, e := s.ShouldFallback(ctx, binding, selected, now); e != nil || got {
		t.Fatal("recovery probe lost preference", e)
	}
	if _, err = s.AcquireGate(ctx, "remote", "chat", now, policy); err != nil {
		t.Fatal(err)
	}
	if got, e := s.ShouldFallback(ctx, binding, selected, now); e != nil || !got {
		t.Fatal("in-flight probe admits another remote call", e)
	}
	if err = s.ResumeGate(ctx, "local", "*", "test", now); err != nil {
		t.Fatal(err)
	}
	if err = s.ResumeGate(ctx, "remote", "*", "test", now); err != nil {
		t.Fatal(err)
	}
	// First account failure schedules another attempt; it never calls the
	// alternative in the primary run or extends the original ceiling.
	failedPrimary := func(context.Context, jobs.Job) (jobs.Result, error) {
		hold("remote", "quota")
		return jobs.Result{}, &jobs.HandlerError{Failure: jobs.Failure{Code: "provider_rejected"}}
	}
	handler = s.FallbackHandler(failedPrimary, local, binding)
	_, err = handler(ctx, selected)
	var retry *jobs.HandlerError
	if !errors.As(err, &retry) || retry.Code != "local_fallback_pending" || !retry.Temporary || localCalls != 1 {
		t.Fatal("first rejection did not preserve next-attempt boundary", err)
	}
	selected.Attempt = 2
	if _, err = handler(ctx, selected); err != nil || localCalls != 2 {
		t.Fatal("subsequent local attempt unavailable", err)
	}
	if err = s.ResumeGate(ctx, "remote", "*", "test", now); err != nil {
		t.Fatal(err)
	}
	selected.Attempt = selected.MaxAttempts
	_, err = handler(ctx, selected)
	if !errors.As(err, &retry) || retry.Code != "provider_rejected" || localCalls != 2 {
		t.Fatal("exhausted attempt budget extended", err)
	}
	// Dependency jobs resolve the selected source through their extraction run.
	if err = s.RegisterConfiguration(ctx, ConfigurationVersion{ID: "extraction", Name: "extraction", Stage: "extraction", Revision: "1", LogicVersion: "fixture", Settings: json.RawMessage(`{}`), CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	version := int64(1)
	parent, e := s.StartRun(ctx, RunRequest{IdempotencyKey: "parent", Workload: "evaluation", Stage: "extraction", ConfigurationVersionID: "extraction", DocumentVersionID: &version, Subject: json.RawMessage(`{}`), CreatedAt: now})
	if e != nil {
		t.Fatal(e)
	}
	body, _ := json.Marshal(map[string]int64{"extraction_run_id": parent.ID, "measure_ordinal": 1})
	child, e := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "inference", Kind: "link_measure_update", IdempotencyKey: "child", Payload: body, MaxAttempts: 3, RetryBase: time.Second, AvailableAt: now})
	if e != nil {
		t.Fatal(e)
	}
	binding.Kind = child.Kind
	if got, e := s.ShouldFallback(ctx, binding, child, now); e != nil || !got {
		t.Fatal("dependency source scope lost", e)
	}
	if err = s.DeferHeldJobs(ctx, []GateBinding{binding}, now); err != nil {
		t.Fatal(err)
	}
	check(child, false)
	// Equivalent copies retain a zero-call path through either compatible cache,
	// irrespective of which provider is currently held.
	for _, configuration := range []string{"remote-cache", "local-cache"} {
		if err = s.RegisterConfiguration(ctx, ConfigurationVersion{ID: configuration, Name: configuration, Stage: "classification", Revision: "1", LogicVersion: "fixture", Settings: json.RawMessage(`{}`), CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
	}
	binding.Kind = "classify_relevance"
	reuseContext := WithProviderFreeReuse(ctx)
	cacheCalls := 0
	cache := func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		if !ProviderFreeReuseOnly(ctx) {
			t.Fatal("alternative lost zero-call restriction")
		}
		cacheCalls++
		run, e := s.StartRun(ctx, RunRequest{IdempotencyKey: fmt.Sprintf("local-cache-%d", job.ID), Workload: "evaluation", Stage: "classification", ConfigurationVersionID: "local-cache", DocumentVersionID: &version, Subject: json.RawMessage(`{}`), CreatedAt: now})
		if e != nil {
			t.Fatal(e)
		}
		a, e := s.StartAttempt(ctx, AttemptStart{RunID: run.ID, QueueJobID: &job.ID, QueueAttemptNumber: &job.Attempt, StartedAt: now})
		if e != nil {
			t.Fatal(e)
		}
		_, e = s.FinishAttempt(ctx, AttemptFinish{RunID: run.ID, Number: a.Number, FinishedAt: now, Outcome: "succeeded", Usage: Usage{Status: "not_applicable"}})
		if e != nil {
			t.Fatal(e)
		}
		return jobs.Result{Payload: json.RawMessage(`{"status":"classified"}`)}, nil
	}
	for index, persisted := range []bool{false, true} {
		_, e := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "reuse", Kind: binding.Kind, IdempotencyKey: fmt.Sprintf("reuse-%d", index), Payload: json.RawMessage(`{"document_version_id":1}`), MaxAttempts: 1, RetryBase: time.Second, AvailableAt: now})
		if e != nil {
			t.Fatal(e)
		}
		claim, e := queue.Claim(ctx, "reuse", "test", now, time.Minute)
		if e != nil {
			t.Fatal(e)
		}
		miss := func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
			if !persisted {
				return jobs.Result{}, &jobs.HandlerError{Failure: jobs.Failure{Code: "equivalent_reuse_unavailable"}}
			}
			run, e := s.StartRun(ctx, RunRequest{IdempotencyKey: fmt.Sprintf("remote-cache-%d", job.ID), Workload: "evaluation", Stage: "classification", ConfigurationVersionID: "remote-cache", DocumentVersionID: &version, Subject: json.RawMessage(`{}`), CreatedAt: now})
			if e != nil {
				t.Fatal(e)
			}
			a, e := s.StartAttempt(ctx, AttemptStart{RunID: run.ID, QueueJobID: &job.ID, QueueAttemptNumber: &job.Attempt, StartedAt: now})
			if e != nil {
				t.Fatal(e)
			}
			_, e = s.FinishAttempt(ctx, AttemptFinish{RunID: run.ID, Number: a.Number, FinishedAt: now, Outcome: "failed", ErrorCode: "equivalent_reuse_unavailable", Usage: Usage{Status: "not_applicable"}})
			if e != nil {
				t.Fatal(e)
			}
			return jobs.Result{Payload: json.RawMessage(`{"status":"uninterpreted","reason_code":"provider_error"}`)}, nil
		}
		if _, e = s.FallbackHandler(miss, cache, binding)(reuseContext, claim.Job); e != nil || cacheCalls != index+1 {
			t.Fatal("alternative cache unavailable at final queue attempt", e)
		}
		if e = queue.Complete(ctx, claim, jobs.Result{Payload: json.RawMessage(`{}`)}, now); e != nil {
			t.Fatal(e)
		}
	}
	// A primary cache hit is preferred even under a hold; out-of-scope misses
	// and validation failures never select the alternative cache.
	if _, err = s.FallbackHandler(primary, cache, binding)(reuseContext, selected); err != nil || cacheCalls != 2 {
		t.Fatal("compatible primary cache ignored", err)
	}
	miss := func(context.Context, jobs.Job) (jobs.Result, error) {
		return jobs.Result{}, &jobs.HandlerError{Failure: jobs.Failure{Code: "equivalent_reuse_unavailable"}}
	}
	if _, err = s.FallbackHandler(miss, cache, binding)(reuseContext, outside); err == nil || cacheCalls != 2 {
		t.Fatal("cache source boundary escaped", err)
	}
	invalid := func(context.Context, jobs.Job) (jobs.Result, error) {
		return jobs.Result{}, &jobs.HandlerError{Failure: jobs.Failure{Code: "classification_output_schema"}}
	}
	if _, err = s.FallbackHandler(invalid, cache, binding)(reuseContext, selected); err == nil || cacheCalls != 2 {
		t.Fatal("invalid response disguised as cache miss", err)
	}

}
