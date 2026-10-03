package linking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
)

type linkStore interface {
	Current(context.Context, int64, int) (MeasureContext, error)
	Candidates(context.Context, MeasureContext, int) ([]Candidate, error)
	SemanticCandidates(context.Context, MeasureContext, string, int) ([]Candidate, error)
	Put(context.Context, Result) error
	Get(context.Context, int64) (Result, bool, error)
}
type processingStore interface {
	StartRun(context.Context, processing.RunRequest) (processing.Run, error)
	StartAttempt(context.Context, processing.AttemptStart) (processing.Attempt, error)
	FinishAttempt(context.Context, processing.AttemptFinish) (processing.Attempt, error)
}

type Runner struct {
	Store                       linkStore
	Processing                  processingStore
	Adapter                     inference.Adapter
	Model, ConfigurationVersion string
	PriceVersion                *string
	MaxCandidates               int
	SemanticConfiguration       string
	SemanticConfigurationFor    func(context.Context, int64) (string, error)
	DisableThinking             bool
	Now                         func() time.Time
}
type Payload struct {
	ExtractionRunID int64  `json:"extraction_run_id"`
	MeasureOrdinal  int    `json:"measure_ordinal"`
	Workload        string `json:"workload"`
}

func Enqueue(ctx context.Context, queue *jobs.Store, policy inference.RetryPolicy, payload Payload, at time.Time) (jobs.Job, error) {
	if queue == nil || payload.ExtractionRunID < 1 || payload.MeasureOrdinal < 1 || at.IsZero() {
		return jobs.Job{}, ErrInvalid
	}
	if payload.Workload == "" {
		payload.Workload = "ordinary"
	}
	body, _ := json.Marshal(payload)
	request, err := policy.JobRequest(Kind, fmt.Sprintf("linking:%d:%d", payload.ExtractionRunID, payload.MeasureOrdinal), body, at)
	if err != nil {
		return jobs.Job{}, err
	}
	return queue.Enqueue(ctx, request)
}
func (r *Runner) Handler() jobs.Handler {
	return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		var p Payload
		if json.Unmarshal(job.Payload, &p) != nil || p.ExtractionRunID < 1 || p.MeasureOrdinal < 1 || !validWorkload(p.Workload) {
			return jobs.Result{}, failure("invalid_linking_job", false)
		}
		return r.run(ctx, job, p)
	}
}

func (r *Runner) run(ctx context.Context, job jobs.Job, p Payload) (jobs.Result, error) {
	if r.Store == nil || r.Processing == nil || r.Adapter == nil || r.Model == "" || r.ConfigurationVersion == "" || job.ID < 1 || job.Attempt < 1 {
		return jobs.Result{}, failure("linking_configuration_invalid", false)
	}
	limit := r.MaxCandidates
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return jobs.Result{}, failure("linking_configuration_invalid", false)
	}
	now := time.Now
	if r.Now != nil {
		now = r.Now
	}
	current, err := r.Store.Current(ctx, p.ExtractionRunID, p.MeasureOrdinal)
	if err != nil {
		return jobs.Result{}, failure("current_measure_unavailable", false)
	}
	candidates, err := r.Store.Candidates(ctx, current, limit)
	if err != nil {
		return jobs.Result{}, failure("linking_candidates_unavailable", true)
	}
	semanticConfiguration := r.SemanticConfiguration
	if r.SemanticConfigurationFor != nil {
		semanticConfiguration, err = r.SemanticConfigurationFor(ctx, current.RunID)
		if err != nil {
			return jobs.Result{}, failure("semantic_configuration_unavailable", true)
		}
	}
	if semanticConfiguration != "" {
		semantic, semanticErr := r.Store.SemanticCandidates(ctx, current, semanticConfiguration, limit)
		if semanticErr != nil {
			return jobs.Result{}, failure("semantic_candidates_unavailable", true)
		}
		candidates = mergeCandidates(candidates, semantic, limit)
	}
	subject, candidateHash := candidateSubject(candidates)
	run, err := r.Processing.StartRun(ctx, processing.RunRequest{IdempotencyKey: fmt.Sprintf("linking:%d:%d:%s:%s", current.RunID, current.Ordinal, candidateHash, r.ConfigurationVersion), Workload: p.Workload, Stage: "linking", ConfigurationVersionID: r.ConfigurationVersion, SourceID: &current.SourceID, DocumentVersionID: &current.DocumentVersionID, Subject: subject, CreatedAt: now().UTC()})
	if err != nil {
		return jobs.Result{}, failure("linking_run_unavailable", true)
	}
	if existing, ok, getErr := r.Store.Get(ctx, run.ID); getErr != nil {
		return jobs.Result{}, failure("linking_result_unavailable", true)
	} else if ok {
		return jobResult(existing)
	}
	jobID, attemptNumber := job.ID, job.Attempt
	attempt, err := r.Processing.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, QueueJobID: &jobID, QueueAttemptNumber: &attemptNumber, StartedAt: now().UTC()})
	if err != nil {
		return jobs.Result{}, failure("linking_attempt_unavailable", true)
	}
	ctx = inference.WithAttempt(ctx, attempt, r.ConfigurationVersion)
	base := Result{RunID: run.ID, CurrentRunID: current.RunID, CurrentOrdinal: current.Ordinal, Candidates: candidates, CreatedAt: now().UTC()}
	if len(candidates) == 0 {
		base.Status, base.ReasonCode = "no_relation", "no_candidates"
		return r.persist(ctx, attempt, base, "succeeded", processing.Usage{Status: "not_applicable"}, nil, "")
	}
	request, err := Request(r.Model, current, candidates)
	if err != nil {
		base.Status, base.ReasonCode = "unresolved", "output_invalid"
		return r.persist(ctx, attempt, base, "failed", processing.Usage{Status: "not_applicable"}, nil, "linking_request_invalid")
	}
	if r.DisableThinking {
		request.ChatTemplateKwargs = &inference.ChatTemplateOptions{EnableThinking: false}
	}
	if r.Adapter.Name() == "local-openai-chat" {
		request = inference.LocalChatRequest(request)
	}
	response, err := r.Adapter.Complete(ctx, request)
	if deferred, deferErr := inference.FinishDeferral(ctx, r.Processing, attempt, err); deferred {
		return jobs.Result{}, deferErr
	}
	if err != nil {
		code, temporary := "provider_error", true
		var call *inference.CallError
		if errors.As(err, &call) {
			code, temporary = call.Code, call.Temporary
		} else if errors.Is(err, inference.ErrInvalid) {
			temporary = false
		}
		usage, _ := linkingUsage(inference.Usage{}, nil)
		if !temporary || job.Attempt >= job.MaxAttempts {
			base.Status, base.ReasonCode = "unresolved", "provider_error"
			if temporary {
				base.ReasonCode = "attempts_exhausted"
			}
			return r.persist(ctx, attempt, base, "failed", usage, nil, code)
		}
		if err = r.finish(ctx, attempt, now(), "failed", usage, nil, code); err != nil {
			return jobs.Result{}, failure("linking_attempt_unavailable", true)
		}
		return jobs.Result{}, failure(code, true)
	}
	parsed, err := Parse(response.Content, current, candidates)
	if err != nil {
		base.Status, base.ReasonCode = "unresolved", "output_invalid"
		usage, _ := linkingUsage(response.Usage, nil)
		return r.persist(ctx, attempt, base, "failed", usage, nil, "linking_output_invalid")
	}
	parsed.RunID = run.ID
	parsed.Candidates = candidates
	parsed.CreatedAt = now().UTC()
	parsed.ProviderResponseID = response.ID
	if parsed.ProviderResponseID == "" {
		parsed.ProviderResponseID = "not_reported"
	}
	parsed.ReturnedModel = response.Model
	parsed.InputTokens, parsed.OutputTokens, parsed.CacheReadTokens = response.Usage.InputTokens, response.Usage.OutputTokens, response.Usage.CacheReadTokens
	usage, price := linkingUsage(response.Usage, r.PriceVersion)
	return r.persist(ctx, attempt, parsed, "succeeded", usage, price, "")
}
func mergeCandidates(baseline, semantic []Candidate, limit int) []Candidate {
	result := append([]Candidate(nil), baseline...)
	seen := map[string]bool{}
	for _, c := range result {
		seen[fmt.Sprintf("%d:%d", c.RunID, c.Ordinal)] = true
	}
	for _, c := range semantic {
		k := fmt.Sprintf("%d:%d", c.RunID, c.Ordinal)
		if !seen[k] && len(result) < limit {
			result = append(result, c)
			seen[k] = true
		}
	}
	return result
}
func (r *Runner) persist(ctx context.Context, a processing.Attempt, v Result, outcome string, u processing.Usage, price *string, code string) (jobs.Result, error) {
	if err := r.finish(ctx, a, v.CreatedAt, outcome, u, price, code); err != nil {
		return jobs.Result{}, failure("linking_attempt_unavailable", true)
	}
	if err := r.Store.Put(ctx, v); err != nil {
		return jobs.Result{}, failure("linking_result_unavailable", true)
	}
	return jobResult(v)
}
func (r *Runner) finish(ctx context.Context, a processing.Attempt, at time.Time, outcome string, u processing.Usage, price *string, code string) error {
	_, err := r.Processing.FinishAttempt(ctx, processing.AttemptFinish{RunID: a.RunID, Number: a.Number, FinishedAt: at.UTC(), Outcome: outcome, Usage: u, PriceVersionID: price, ErrorCode: code})
	return err
}
func linkingUsage(v inference.Usage, price *string) (processing.Usage, *string) {
	status := "reported"
	if v.InputTokens == nil || v.OutputTokens == nil {
		status, price = "partial", nil
	}
	if v.InputTokens == nil && v.OutputTokens == nil && v.CacheReadTokens == nil {
		return processing.Usage{Status: "unavailable"}, nil
	}
	if v.CacheReadTokens != nil {
		price = nil
	}
	return processing.Usage{Status: status, InputTokens: v.InputTokens, OutputTokens: v.OutputTokens, CacheReadTokens: v.CacheReadTokens}, price
}
func jobResult(v Result) (jobs.Result, error) {
	payload, _ := json.Marshal(struct {
		RunID  int64  `json:"run_id"`
		Status string `json:"status"`
		Reason string `json:"reason_code"`
	}{v.RunID, v.Status, v.ReasonCode})
	return jobs.JSONResult(json.RawMessage(payload), jobs.Effect{Key: fmt.Sprintf("linking-run:%d", v.RunID), Kind: "linking_ready", Payload: payload})
}
func failure(code string, temporary bool) error {
	return &jobs.HandlerError{Failure: jobs.Failure{Code: code, Detail: "measure update linking failed", Temporary: temporary}}
}
func validWorkload(v string) bool {
	return v == "ordinary" || v == "bootstrap" || v == "evaluation" || v == "reprocessing"
}
