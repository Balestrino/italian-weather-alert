//go:build integration

package processing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func TestRecoveryScopeConcurrentCapAndFailure(t *testing.T) {
	ctx := context.Background()
	pool := processingTestDB(t)
	for _, m := range []func() error{func() error { return registry.Migrate(ctx, pool) }, func() error { return documents.Migrate(ctx, pool) }, func() error { return jobs.Migrate(ctx, pool) }, func() error { return Migrate(ctx, pool) }} {
		if e := m(); e != nil {
			t.Fatal(e)
		}
	}
	_, err := pool.Exec(ctx, `INSERT INTO registry_authorities VALUES ('a','a','https://example.org'); INSERT INTO registry_channels VALUES ('c','a','web','https://example.org',false); INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES ('s','a','c','municipal','test'); INSERT INTO retained_documents(source_id,official_url) VALUES ('s','https://example.org'); INSERT INTO retained_versions(document_id,content_hash,first_acquired_at,complete,metadata) VALUES (1,repeat('a',64),now(),true,'{}'),(1,repeat('b',64),now(),true,'{}');`)
	if err != nil {
		t.Fatal(err)
	}
	s := New(pool)
	now := time.Now().UTC()
	if err = s.RegisterConfiguration(ctx, ConfigurationVersion{ID: "recovery", Name: "recovery", Stage: "classification", Revision: "1", LogicVersion: "test", Settings: json.RawMessage(`{}`), CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	makeRun := func(id int64) Run {
		r, e := s.StartRun(ctx, RunRequest{IdempotencyKey: string(rune('a' + id)), Workload: "ordinary", Stage: "classification", ConfigurationVersionID: "recovery", DocumentVersionID: &id, Subject: json.RawMessage(`{}`), CreatedAt: now})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = s.StartAttempt(ctx, AttemptStart{RunID: r.ID, StartedAt: now}); e != nil {
			t.Fatal(e)
		}
		return r
	}
	allowed := makeRun(1)
	outside := makeRun(2)
	_, err = pool.Exec(ctx, `INSERT INTO processing_recovery_limits(version_ids,max_calls,started_at,actor) VALUES(ARRAY[1]::bigint[],2,$1,'test')`, now)
	if err != nil {
		t.Fatal(err)
	}
	start := CallStart{RunID: outside.ID, AttemptNumber: 1, Ordinal: 1, InputSHA256: strings.Repeat("a", 64), Provider: "fixture", RequestedModel: "fixture", StartedAt: now}
	var deferred *jobs.DeferredError
	if _, err = s.StartCall(ctx, start); !errors.As(err, &deferred) || deferred.Code != "recovery_scope_wait" {
		t.Fatalf("scope escaped: %v", err)
	}
	start.RunID = allowed.ID
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for i := 1; i <= 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v := start
			v.Ordinal = i
			if _, e := s.StartCall(ctx, v); e == nil {
				accepted.Add(1)
			} else {
				var d *jobs.DeferredError
				if !errors.As(e, &d) || d.Code != "recovery_call_limit" {
					t.Errorf("unexpected: %v", e)
				}
			}
		}(i)
	}
	wg.Wait()
	if accepted.Load() != 2 {
		t.Fatalf("cap: %d", accepted.Load())
	}
	// New store instance proves the cap survives restart; rejected calls add no intent.
	start.Ordinal = 9
	if _, err = New(pool).StartCall(ctx, start); !errors.As(err, &deferred) {
		t.Fatal("restart bypassed cap")
	}
	var count int
	pool.QueryRow(ctx, `SELECT count(*) FROM processing_provider_calls`).Scan(&count)
	if count != 2 {
		t.Fatal("denied call billed")
	}
	// A document-local quotation failure does not block another selected
	// document, but it prevents re-admitting the failed document itself.
	capture := InvalidOutput{RunID: allowed.ID, AttemptNumber: 1, SegmentOrdinal: 1, Request: []byte(`{"model":"fixture"}`), Response: []byte("invalid\x00output"), FinishReason: "stop", ErrorCode: "classification_output_quotation", CreatedAt: now}
	if e := s.RecordInvalidOutput(ctx, capture); e != nil {
		t.Fatal(e)
	}
	if e := s.RecordInvalidOutput(ctx, capture); e != nil {
		t.Fatal("idempotent capture", e)
	}
	conflicting := capture
	conflicting.Response = []byte("different")
	if e := s.RecordInvalidOutput(ctx, conflicting); !errors.Is(e, ErrConflict) {
		t.Fatal("capture overwritten", e)
	}
	var saved []byte
	if e := pool.QueryRow(ctx, `SELECT response_bytes FROM processing_invalid_outputs WHERE run_id=$1`, allowed.ID).Scan(&saved); e != nil || string(saved) != string(capture.Response) {
		t.Fatal("raw evidence lost", e)
	}
	if _, e := s.FinishAttempt(ctx, AttemptFinish{RunID: allowed.ID, Number: 1, FinishedAt: now.Add(time.Second), Outcome: "failed", ErrorCode: "classification_output_quotation", Usage: Usage{Status: "unavailable"}}); e != nil {
		t.Fatal(e)
	}
	if _, e := pool.Exec(ctx, `UPDATE processing_recovery_limits SET version_ids=ARRAY[1,2]::bigint[],max_calls=20`); e != nil {
		t.Fatal(e)
	}
	if _, e := s.StartCall(ctx, start); !errors.As(e, &deferred) || deferred.Code != "recovery_document_quarantined" {
		t.Fatal("failed document was not quarantined", e)
	}
	healthy := start
	healthy.RunID = outside.ID
	healthy.Ordinal = 1
	if _, e := s.StartCall(ctx, healthy); e != nil {
		t.Fatal("unrelated document stopped", e)
	}
	queue := jobs.New(pool)
	for _, v := range []int{1, 2} {
		body, _ := json.Marshal(map[string]int{"document_version_id": v})
		j, e := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "inference", Kind: "classify_relevance", IdempotencyKey: fmt.Sprint("scope-", v), Payload: body, MaxAttempts: 3, RetryBase: time.Second, AvailableAt: now})
		if e != nil {
			t.Fatal(e)
		}
		if e = s.DeferRecoveryJobs(ctx, now); e != nil {
			t.Fatal(e)
		}
		var at time.Time
		var tries int
		if e = pool.QueryRow(ctx, `SELECT available_at,attempt_count FROM processing_jobs WHERE id=$1`, j.ID).Scan(&at, &tries); e != nil || tries != 0 || at.After(now) != (v == 1) {
			t.Fatal("pre-claim quarantine scope incorrect", v, e)
		}
	}
	if e := s.RegisterConfiguration(ctx, ConfigurationVersion{ID: "recovery-extraction", Name: "recovery-extraction", Stage: "extraction", Revision: "1", LogicVersion: "test", Settings: json.RawMessage(`{}`), CreatedAt: now}); e != nil {
		t.Fatal(e)
	}
	for _, version := range []int64{1, 2} {
		parent, e := s.StartRun(ctx, RunRequest{IdempotencyKey: fmt.Sprint("parent-", version), Workload: "ordinary", Stage: "extraction", ConfigurationVersionID: "recovery-extraction", DocumentVersionID: &version, Subject: json.RawMessage(`{}`), CreatedAt: now})
		if e != nil {
			t.Fatal(e)
		}
		for _, kind := range []string{"embed_measure", "link_measure_update"} {
			body, _ := json.Marshal(map[string]int64{"extraction_run_id": parent.ID, "measure_ordinal": 1})
			child, e := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "inference", Kind: kind, IdempotencyKey: fmt.Sprint(kind, version), Payload: body, MaxAttempts: 3, RetryBase: time.Second, AvailableAt: now})
			if e != nil {
				t.Fatal(e)
			}
			if e = s.DeferRecoveryJobs(ctx, now); e != nil {
				t.Fatal(e)
			}
			var at time.Time
			var tries int
			if e = pool.QueryRow(ctx, `SELECT available_at,attempt_count FROM processing_jobs WHERE id=$1`, child.ID).Scan(&at, &tries); e != nil || tries != 0 || at.After(now) != (version == 1) {
				t.Fatal("child document scope/quarantine incorrect", version, kind, e)
			}
		}
	}
	if _, e := s.FinishAttempt(ctx, AttemptFinish{RunID: outside.ID, Number: 1, FinishedAt: now.Add(time.Second), Outcome: "failed", ErrorCode: "classification_diagnostic_unavailable", Usage: Usage{Status: "unavailable"}}); e != nil {
		t.Fatal(e)
	}
	if _, e := s.StartCall(ctx, start); !errors.As(e, &deferred) || deferred.Code != "recovery_error_hold" {
		t.Fatal("diagnostic storage failure failed open", e)
	}
	var callID int64
	pool.QueryRow(ctx, `SELECT min(id) FROM processing_provider_calls`).Scan(&callID)
	if err = s.FinishCall(ctx, CallFinish{ID: callID, FinishedAt: now.Add(time.Second), State: "failed", ErrorCode: "provider_error", Usage: Usage{Status: "unavailable"}}); err != nil {
		t.Fatal(err)
	}
	pool.Exec(ctx, `UPDATE processing_recovery_limits SET max_calls=20`)
	if _, err = s.StartCall(ctx, start); !errors.As(err, &deferred) || deferred.Code != "recovery_error_hold" {
		t.Fatalf("failure did not stop: %v", err)
	}
	// Pre-claim deferral leaves retry counters unchanged.
	q := jobs.New(pool)
	j, e := q.Enqueue(ctx, jobs.EnqueueRequest{Queue: "inference", Kind: "classify_relevance", IdempotencyKey: "excluded", Payload: json.RawMessage(`{"document_version_id":2}`), MaxAttempts: 3, RetryBase: time.Second, AvailableAt: now})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.DeferRecoveryJobs(ctx, now); e != nil {
		t.Fatal(e)
	}
	var attempts, cap int
	var available time.Time
	if e = pool.QueryRow(ctx, `SELECT attempt_count,max_attempts,available_at FROM processing_jobs WHERE id=$1`, j.ID).Scan(&attempts, &cap, &available); e != nil || attempts != 0 || cap != 3 || !available.After(now) {
		t.Fatal("deferral spent retries", e)
	}
}
