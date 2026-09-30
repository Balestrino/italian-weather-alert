package embedding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/linking"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"time"
)

type measureStore interface {
	Current(context.Context, int64, int) (linking.MeasureContext, error)
}
type resultStore interface {
	Put(context.Context, Result) error
	Get(context.Context, int64, int, string) (Result, bool, error)
}
type processStore interface {
	StartRun(context.Context, processing.RunRequest) (processing.Run, error)
	StartAttempt(context.Context, processing.AttemptStart) (processing.Attempt, error)
	FinishAttempt(context.Context, processing.AttemptFinish) (processing.Attempt, error)
}
type Runner struct {
	Measures                    measureStore
	Results                     resultStore
	Processing                  processStore
	Adapter                     inference.Embedder
	Model, ConfigurationVersion string
	Dimensions                  int
	Now                         func() time.Time
}
type Payload struct {
	ExtractionRunID int64  `json:"extraction_run_id"`
	MeasureOrdinal  int    `json:"measure_ordinal"`
	Workload        string `json:"workload"`
}

func Enqueue(ctx context.Context, q *jobs.Store, policy inference.RetryPolicy, p Payload, at time.Time) (jobs.Job, error) {
	if q == nil || p.ExtractionRunID < 1 || p.MeasureOrdinal < 1 || at.IsZero() {
		return jobs.Job{}, ErrInvalid
	}
	if p.Workload == "" {
		p.Workload = "ordinary"
	}
	b, _ := json.Marshal(p)
	request, e := policy.JobRequest(Kind, fmt.Sprintf("embedding:%d:%d:%s", p.ExtractionRunID, p.MeasureOrdinal, "v1"), b, at)
	if e != nil {
		return jobs.Job{}, e
	}
	return q.Enqueue(ctx, request)
}
func (r *Runner) Handler() jobs.Handler {
	return func(ctx context.Context, j jobs.Job) (jobs.Result, error) {
		var p Payload
		if json.Unmarshal(j.Payload, &p) != nil || p.ExtractionRunID < 1 || p.MeasureOrdinal < 1 {
			return jobs.Result{}, fail("invalid_embedding_job", false)
		}
		return r.run(ctx, j, p)
	}
}
func (r *Runner) run(ctx context.Context, j jobs.Job, p Payload) (jobs.Result, error) {
	if r.Measures == nil || r.Results == nil || r.Processing == nil || r.Adapter == nil || r.Model == "" || r.ConfigurationVersion == "" || r.Dimensions != 1024 {
		return jobs.Result{}, fail("embedding_configuration_invalid", false)
	}
	now := time.Now
	if r.Now != nil {
		now = r.Now
	}
	m, e := r.Measures.Current(ctx, p.ExtractionRunID, p.MeasureOrdinal)
	if e != nil {
		return jobs.Result{}, fail("measure_unavailable", false)
	}
	if existing, ok, e := r.Results.Get(ctx, m.RunID, m.Ordinal, r.ConfigurationVersion); e != nil {
		return jobs.Result{}, fail("embedding_result_unavailable", true)
	} else if ok {
		return output(existing)
	}
	subject, _ := json.Marshal(map[string]any{"measure_run_id": m.RunID, "measure_ordinal": m.Ordinal})
	run, e := r.Processing.StartRun(ctx, processing.RunRequest{IdempotencyKey: fmt.Sprintf("embedding:%d:%d:%s", m.RunID, m.Ordinal, r.ConfigurationVersion), Workload: p.Workload, Stage: "embedding", ConfigurationVersionID: r.ConfigurationVersion, SourceID: &m.SourceID, DocumentVersionID: &m.DocumentVersionID, Subject: subject, CreatedAt: now().UTC()})
	if e != nil {
		return jobs.Result{}, fail("embedding_run_unavailable", true)
	}
	jobID, attemptNo := j.ID, j.Attempt
	attempt, e := r.Processing.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, QueueJobID: &jobID, QueueAttemptNumber: &attemptNo, StartedAt: now().UTC()})
	if e != nil {
		return jobs.Result{}, fail("embedding_attempt_unavailable", true)
	}
	ctx = inference.WithAttempt(ctx, attempt, r.ConfigurationVersion)
	text := m.Kind + " | " + m.Subject + " | " + m.Place
	response, e := r.Adapter.Embed(ctx, r.Model, []string{text}, r.Dimensions)
	if deferred, deferErr := inference.FinishDeferral(ctx, r.Processing, attempt, e); deferred {
		return jobs.Result{}, deferErr
	}
	if e != nil {
		code, temporary := "embedding_provider_error", true
		var call *inference.CallError
		if errors.As(e, &call) {
			code, temporary = call.Code, call.Temporary
		}
		if _, finishErr := r.Processing.FinishAttempt(ctx, processing.AttemptFinish{RunID: attempt.RunID, Number: attempt.Number, FinishedAt: now().UTC(), Outcome: "failed", Usage: processing.Usage{Status: "unavailable"}, ErrorCode: code}); finishErr != nil {
			return jobs.Result{}, fail("embedding_attempt_unavailable", true)
		}
		return jobs.Result{}, fail(code, temporary)
	}
	usage := processing.Usage{Status: "reported", InputTokens: response.InputTokens}
	if response.InputTokens == nil {
		usage.Status = "unavailable"
	}
	if _, e = r.Processing.FinishAttempt(ctx, processing.AttemptFinish{RunID: attempt.RunID, Number: attempt.Number, FinishedAt: now().UTC(), Outcome: "succeeded", Usage: usage}); e != nil {
		return jobs.Result{}, fail("embedding_attempt_unavailable", true)
	}
	result := Result{RunID: run.ID, ExtractionRunID: m.RunID, MeasureOrdinal: m.Ordinal, ConfigurationVersionID: r.ConfigurationVersion, Dimensions: r.Dimensions, Vector: response.Vectors[0], ReturnedModel: response.Model, InputTokens: response.InputTokens, CreatedAt: now().UTC()}
	if e = r.Results.Put(ctx, result); e != nil {
		return jobs.Result{}, fail("embedding_result_unavailable", true)
	}
	return output(result)
}
func output(r Result) (jobs.Result, error) {
	p, _ := json.Marshal(map[string]any{"run_id": r.RunID, "extraction_run_id": r.ExtractionRunID, "measure_ordinal": r.MeasureOrdinal, "dimensions": r.Dimensions})
	return jobs.JSONResult(json.RawMessage(p), jobs.Effect{Key: fmt.Sprintf("embedding-run:%d", r.RunID), Kind: "embedding_ready", Payload: p})
}
func fail(code string, temp bool) error {
	return &jobs.HandlerError{Failure: jobs.Failure{Code: code, Detail: "measure embedding failed", Temporary: temp}}
}
