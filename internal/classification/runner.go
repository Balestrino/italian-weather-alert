package classification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
)

type processingStore interface {
	RecordInvalidOutput(context.Context, processing.InvalidOutput) error
	StartRun(context.Context, processing.RunRequest) (processing.Run, error)
	StartAttempt(context.Context, processing.AttemptStart) (processing.Attempt, error)
	FinishAttempt(context.Context, processing.AttemptFinish) (processing.Attempt, error)
}

type resultStore interface {
	Put(context.Context, Result) error
	Get(context.Context, int64) (Result, bool, error)
}

type Runner struct {
	OutputFixSources           map[string]bool
	CheckpointSources          map[string]bool
	LegacyConfigurationVersion string
	legacyOutput               bool
	Checkpoints                inference.CheckpointStore
	Manifests                  *Store
	ReuseSources               map[string]bool
	Documents                  documentStore
	OCR                        ocrStore
	Processing                 processingStore
	Results                    resultStore
	Adapter                    inference.Adapter
	Model                      string
	ConfigurationVersion       string
	PriceVersion               *string
	DisableThinking            bool
	Now                        func() time.Time
}

type Payload struct {
	DocumentVersionID int64  `json:"document_version_id"`
	Workload          string `json:"workload"`
	SelectionID       string `json:"selection_id,omitempty"`
}

func Enqueue(ctx context.Context, queue *jobs.Store, policy inference.RetryPolicy, payload Payload, availableAt time.Time) (jobs.Job, error) {
	if queue == nil || payload.DocumentVersionID < 1 || availableAt.IsZero() {
		return jobs.Job{}, ErrInvalid
	}
	if payload.Workload == "" {
		payload.Workload = "ordinary"
	}
	body, _ := json.Marshal(payload)
	if payload.Workload == "reprocessing" && payload.SelectionID == "" {
		return jobs.Job{}, ErrInvalid
	}
	request, err := policy.JobRequest(Kind, fmt.Sprintf("classification:%d:%s", payload.DocumentVersionID, payload.SelectionID), body, availableAt)
	if err != nil {
		return jobs.Job{}, err
	}
	return queue.Enqueue(ctx, request)
}

func (r *Runner) Handler() jobs.Handler {
	return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		var payload Payload
		if json.Unmarshal(job.Payload, &payload) != nil || payload.DocumentVersionID < 1 || (payload.Workload != "ordinary" && payload.Workload != "bootstrap" && payload.Workload != "evaluation" && payload.Workload != "reprocessing") || (payload.Workload == "reprocessing" && payload.SelectionID == "") {
			return jobs.Result{}, classFailure("invalid_classification_job", false)
		}
		return r.run(ctx, job, payload)
	}
}

func (r *Runner) run(ctx context.Context, job jobs.Job, payload Payload) (jobs.Result, error) {
	if r.Documents == nil || r.OCR == nil || r.Processing == nil || r.Results == nil || r.Adapter == nil || r.Model == "" || r.ConfigurationVersion == "" || job.ID < 1 || job.Attempt < 1 {
		return jobs.Result{}, classFailure("classification_configuration_invalid", false)
	}
	now := time.Now
	if r.Now != nil {
		now = r.Now
	}
	version, err := r.Documents.Version(ctx, payload.DocumentVersionID)
	if err != nil {
		return jobs.Result{}, classFailure("retained_version_unavailable", true)
	}
	configuredRunner := *r
	source := ""
	for _, ref := range version.Resources {
		if ref.Role == "original" {
			source = ref.SourceID
			break
		}
	}
	if r.OutputFixSources != nil && !r.OutputFixSources[source] {
		configuredRunner.legacyOutput = true
		configuredRunner.ConfigurationVersion = r.LegacyConfigurationVersion
	}
	if r.CheckpointSources != nil && !r.CheckpointSources[source] {
		configuredRunner.Checkpoints = nil
	}
	r = &configuredRunner
	content, err := gatherContent(ctx, r.Documents, r.OCR, version)
	if err != nil {
		if errors.Is(err, ErrInvalid) {
			return jobs.Result{}, classFailure("classification_content_too_large", false)
		}
		return jobs.Result{}, classFailure("classification_content_unavailable", true)
	}
	var segments []Segment
	if content.Complete && len(content.Sections) > 0 {
		segments, err = SegmentContent(version.ID, content)
		if err != nil {
			return jobs.Result{}, classFailure("classification_content_too_large", false)
		}
	}
	var manifest InputManifest
	manifestSuffix := ""
	if r.Manifests != nil {
		manifest, err = BuildManifest(ctx, r.Documents, r.OCR, version, content, r.ConfigurationVersion)
		if err != nil {
			return jobs.Result{}, classFailure("classification_manifest_unavailable", true)
		}
		manifestSuffix = ":" + manifest.Hash
	}
	subject, _ := json.Marshal(map[string]any{"content_sha256": content.Hash, "content_complete": content.Complete, "segment_count": len(segments), "segmentation": "normalized-byte-v1"})
	var sourceID *string
	if len(version.Resources) > 0 {
		sourceID = &version.Resources[0].SourceID
	}
	run, err := r.Processing.StartRun(ctx, processing.RunRequest{IdempotencyKey: fmt.Sprintf("classification:%d:%s:%s:%s", version.ID, content.Hash, r.ConfigurationVersion, payload.SelectionID) + manifestSuffix, Workload: payload.Workload, Stage: "classification", ConfigurationVersionID: r.ConfigurationVersion, SourceID: sourceID, DocumentVersionID: &version.ID, Subject: subject, CreatedAt: now().UTC()})
	if err != nil {
		return jobs.Result{}, classFailure("classification_run_unavailable", true)
	}
	if r.Manifests != nil {
		if err = r.Manifests.PutManifest(ctx, run.ID, manifest); err != nil {
			return jobs.Result{}, classFailure("classification_manifest_unavailable", true)
		}
	}
	if existing, ok, lookupErr := r.Results.Get(ctx, run.ID); lookupErr != nil {
		return jobs.Result{}, classFailure("classification_result_unavailable", true)
	} else if ok {
		return classResult(existing)
	}
	jobID, queueAttempt := job.ID, job.Attempt
	attempt, err := r.Processing.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, QueueJobID: &jobID, QueueAttemptNumber: &queueAttempt, StartedAt: now().UTC()})
	if err != nil {
		return jobs.Result{}, classFailure("classification_attempt_unavailable", true)
	}
	ctx = inference.WithAttempt(ctx, attempt, r.ConfigurationVersion)
	if !content.Complete || len(content.Sections) == 0 {
		reason := "incomplete_required_content"
		if content.Complete {
			reason = "empty_content"
		}
		undetermined := Result{RunID: run.ID, DocumentVersionID: version.ID, Status: "undetermined", ReasonCode: reason, ContentSHA256: content.Hash, ContentComplete: content.Complete, CreatedAt: now().UTC()}
		if err = r.finish(ctx, attempt, now(), "succeeded", processing.Usage{Status: "not_applicable"}, nil, ""); err != nil {
			return jobs.Result{}, classFailure("classification_attempt_unavailable", true)
		}
		if err = r.Results.Put(ctx, undetermined); err != nil {
			return jobs.Result{}, classFailure("classification_result_unavailable", true)
		}
		return classResult(undetermined)
	}
	if r.Manifests != nil && r.ReuseSources[manifest.SourceID] && manifest.Reusable {
		original, found, lookupErr := r.Manifests.EquivalentRun(ctx, run.ID, "classification")
		if lookupErr != nil {
			return r.fail(ctx, attempt, now(), "classification_reuse_unavailable", true, inference.Usage{})
		}
		if found {
			prior, ok, e := r.Results.Get(ctx, original)
			if e != nil {
				return r.fail(ctx, attempt, now(), "classification_reuse_unavailable", true, inference.Usage{})
			}
			if ok {
				if mapped, valid := RemapResult(prior, segments, content, run.ID, version.ID, now()); valid {
					if e = r.Manifests.RecordReuse(ctx, run.ID, original, now()); e != nil {
						return r.fail(ctx, attempt, now(), "classification_reuse_unavailable", true, inference.Usage{})
					}
					if e = r.finish(ctx, attempt, now(), "succeeded", processing.Usage{Status: "not_applicable"}, nil, ""); e != nil {
						return jobs.Result{}, classFailure("classification_attempt_unavailable", true)
					}
					if e = r.Results.Put(ctx, mapped); e != nil {
						return jobs.Result{}, classFailure("classification_result_unavailable", true)
					}
					return classResult(mapped)
				}
			}
		}
	}
	var selected Decision
	var selectedSet bool
	usageTotal := newUsageTotal()
	providerHash := sha256.New()
	returnedModel := ""
	segmentResults := make([]SegmentResult, 0, len(segments))
	for _, segment := range segments {
		request, requestErr := SegmentRequest(r.Model, segment)
		if requestErr != nil {
			return r.fail(ctx, attempt, now(), "classification_request_invalid", false, usageTotal.finish())
		}
		if r.DisableThinking {
			request.ChatTemplateKwargs = &inference.ChatTemplateOptions{EnableThinking: false}
		}
		response, cached, responseErr := inference.LoadCheckpoint(ctx, r.Checkpoints, "classification", r.ConfigurationVersion, segment.Ordinal, request)
		if responseErr == nil && !cached {
			response, responseErr = r.Adapter.Complete(ctx, request)
		}
		if deferred, deferErr := inference.FinishDeferral(ctx, r.Processing, attempt, responseErr); deferred {
			return jobs.Result{}, deferErr
		}
		if responseErr != nil {
			code, temporary := "classification_provider_error", true
			var call *inference.CallError
			if errors.As(responseErr, &call) {
				code, temporary = call.Code, call.Temporary
			} else if errors.Is(responseErr, inference.ErrInvalid) {
				code, temporary = "classification_request_invalid", false
			}
			return r.fail(ctx, attempt, now(), code, temporary, usageTotal.finish())
		}
		if !cached {
			usageTotal.add(response.Usage)
		}
		decision, parseErr := ParseDecision(response.Content, segment.Content())
		if r.legacyOutput {
			decision, parseErr = ParseLegacyDecision(response.Content, segment.Content())
		}
		if !r.legacyOutput && (response.FinishReason == "length" || response.FinishReason == "content_filter") {
			parseErr = ErrInvalid
		}
		if parseErr != nil {
			code := InvalidOutputReason(response.Content, response.FinishReason, segment.Content())
			requestBytes, marshalErr := json.Marshal(request)
			captureCtx, cancelCapture := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			captureErr := marshalErr
			if captureErr == nil {
				captureErr = r.Processing.RecordInvalidOutput(captureCtx, processing.InvalidOutput{RunID: attempt.RunID, AttemptNumber: attempt.Number, SegmentOrdinal: segment.Ordinal, Request: requestBytes, Response: []byte(response.Content), FinishReason: response.FinishReason, ErrorCode: code, Cached: cached, CreatedAt: now().UTC()})
			}
			cancelCapture()
			if captureErr != nil {
				return r.fail(ctx, attempt, now(), "classification_diagnostic_unavailable", false, usageTotal.finish())
			}
			return r.fail(ctx, attempt, now(), code, false, usageTotal.finish())
		}
		providerID := response.ID
		if providerID == "" {
			providerID = "not_reported"
		}
		model := response.Model
		if model == "" {
			model = r.Model
		}
		if returnedModel != "" && model != returnedModel {
			return r.fail(ctx, attempt, now(), "classification_output_model_changed", false, usageTotal.finish())
		}
		if !cached {
			if checkpointErr := inference.SaveCheckpoint(ctx, r.Checkpoints, "classification", r.ConfigurationVersion, segment.Ordinal, request, response); checkpointErr != nil {
				return r.fail(ctx, attempt, now(), "classification_checkpoint_unavailable", true, usageTotal.finish())
			}
		}
		returnedModel = model
		responseDigest := sha256.Sum256([]byte(response.Content))
		providerHash.Write([]byte(providerID))
		providerHash.Write(responseDigest[:])
		segmentResults = append(segmentResults, SegmentResult{
			Ordinal: segment.Ordinal, Total: segment.Total, DocumentVersionID: version.ID,
			ResourceURL: segment.ResourceURL, Role: segment.Role, Page: segment.Page, StartByte: segment.StartByte, EndByte: segment.EndByte,
			ContentSHA256: segment.Hash, ResponseSHA256: hex.EncodeToString(responseDigest[:]), Relevant: decision.Relevant,
			ReasonCode: decision.ReasonCode, EvidenceQuote: decision.EvidenceQuote, ProviderResponseID: providerID, ReturnedModel: model,
			InputTokens: response.Usage.InputTokens, OutputTokens: response.Usage.OutputTokens, CacheReadTokens: response.Usage.CacheReadTokens,
		})
		if !selectedSet || decision.Relevant && !selected.Relevant {
			selected, selectedSet = decision, true
		}
	}
	reported := usageTotal.finish()
	relevant := selected.Relevant
	classified := Result{RunID: run.ID, DocumentVersionID: version.ID, Status: "classified", Relevant: &relevant, ReasonCode: selected.ReasonCode, EvidenceQuote: selected.EvidenceQuote, ContentSHA256: content.Hash, ContentComplete: true, ProviderResponseID: "segmented:" + hex.EncodeToString(providerHash.Sum(nil)), ReturnedModel: returnedModel, InputTokens: reported.InputTokens, OutputTokens: reported.OutputTokens, CacheReadTokens: reported.CacheReadTokens, CreatedAt: now().UTC(), Segments: segmentResults}
	usage, price := classificationUsage(reported, r.PriceVersion)
	if err = r.finish(ctx, attempt, now(), "succeeded", usage, price, ""); err != nil {
		return jobs.Result{}, classFailure("classification_attempt_unavailable", true)
	}
	if err = r.Results.Put(ctx, classified); err != nil {
		return jobs.Result{}, classFailure("classification_result_unavailable", true)
	}
	return classResult(classified)
}

type usageTotal struct {
	calls                                          int
	input, output, cache                           int64
	inputKnown, outputKnown, cacheKnown, cacheSeen bool
}

func newUsageTotal() usageTotal {
	return usageTotal{inputKnown: true, outputKnown: true, cacheKnown: true}
}

func (u *usageTotal) add(value inference.Usage) {
	u.calls++
	if value.InputTokens == nil {
		u.inputKnown = false
	} else {
		u.input += *value.InputTokens
	}
	if value.OutputTokens == nil {
		u.outputKnown = false
	} else {
		u.output += *value.OutputTokens
	}
	if value.CacheReadTokens == nil {
		u.cacheKnown = false
	} else {
		u.cacheSeen = true
		u.cache += *value.CacheReadTokens
	}
}

func (u usageTotal) finish() inference.Usage {
	var result inference.Usage
	if u.calls > 0 && u.inputKnown {
		result.InputTokens = int64Value(u.input)
	}
	if u.calls > 0 && u.outputKnown {
		result.OutputTokens = int64Value(u.output)
	}
	if u.calls > 0 && u.cacheKnown && u.cacheSeen {
		result.CacheReadTokens = int64Value(u.cache)
	}
	return result
}

func int64Value(value int64) *int64 { return &value }

func (r *Runner) fail(ctx context.Context, attempt processing.Attempt, at time.Time, code string, temporary bool, reported inference.Usage) (jobs.Result, error) {
	usage, _ := classificationUsage(reported, nil)
	if reported.InputTokens == nil && reported.OutputTokens == nil && reported.CacheReadTokens == nil {
		usage = processing.Usage{Status: "unavailable"}
	}
	if err := r.finish(ctx, attempt, at, "failed", usage, nil, code); err != nil {
		return jobs.Result{}, classFailure("classification_attempt_unavailable", true)
	}
	return jobs.Result{}, classFailure(code, temporary)
}

func (r *Runner) finish(ctx context.Context, attempt processing.Attempt, at time.Time, outcome string, usage processing.Usage, price *string, code string) error {
	_, err := r.Processing.FinishAttempt(ctx, processing.AttemptFinish{RunID: attempt.RunID, Number: attempt.Number, FinishedAt: at.UTC(), Outcome: outcome, Usage: usage, PriceVersionID: price, ErrorCode: code})
	return err
}

func classificationUsage(reported inference.Usage, price *string) (processing.Usage, *string) {
	status := "reported"
	if reported.InputTokens == nil || reported.OutputTokens == nil {
		status, price = "partial", nil
	}
	if reported.InputTokens == nil && reported.OutputTokens == nil && reported.CacheReadTokens == nil {
		return processing.Usage{Status: "unavailable"}, nil
	}
	if reported.CacheReadTokens != nil {
		price = nil
	}
	return processing.Usage{Status: status, InputTokens: reported.InputTokens, OutputTokens: reported.OutputTokens, CacheReadTokens: reported.CacheReadTokens}, price
}

func classResult(value Result) (jobs.Result, error) {
	payload, _ := json.Marshal(struct {
		RunID             int64  `json:"run_id"`
		DocumentVersionID int64  `json:"document_version_id"`
		Status            string `json:"status"`
		Relevant          *bool  `json:"relevant"`
		ReasonCode        string `json:"reason_code"`
	}{value.RunID, value.DocumentVersionID, value.Status, value.Relevant, value.ReasonCode})
	return jobs.JSONResult(json.RawMessage(payload), jobs.Effect{Key: fmt.Sprintf("classification-run:%d", value.RunID), Kind: "classification_ready", Payload: payload})
}

func classFailure(code string, temporary bool) error {
	return &jobs.HandlerError{Failure: jobs.Failure{Code: code, Detail: "document relevance classification failed", Temporary: temporary}}
}
