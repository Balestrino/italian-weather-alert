package extraction

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
)

type documentStore interface {
	Version(context.Context, int64) (documents.Version, error)
	Read(context.Context, int64, string) ([]byte, error)
}

type ocrStore interface {
	PagesForVersion(context.Context, int64) ([]ocr.PageResult, error)
	ResourcesForVersion(context.Context, int64) ([]ocr.ResourceResult, error)
}

type classificationStore interface {
	Get(context.Context, int64) (classification.Result, bool, error)
}

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
	OutputFixSources            map[string]bool
	CheckpointSources           map[string]bool
	LegacyConfigurationVersion  string
	legacyOutput                bool
	Checkpoints                 inference.CheckpointStore
	Manifests                   *classification.Store
	ReuseSources                map[string]bool
	Documents                   documentStore
	OCR                         ocrStore
	Classifications             classificationStore
	Processing                  processingStore
	Results                     resultStore
	Adapter                     inference.Adapter
	Model, ConfigurationVersion string
	PriceVersion                *string
	DisableThinking             bool
	Now                         func() time.Time
}

type Payload struct {
	DocumentVersionID   int64  `json:"document_version_id"`
	ClassificationRunID int64  `json:"classification_run_id"`
	Workload            string `json:"workload"`
}

func Enqueue(ctx context.Context, queue *jobs.Store, policy inference.RetryPolicy, payload Payload, availableAt time.Time) (jobs.Job, error) {
	if queue == nil || payload.DocumentVersionID < 1 || payload.ClassificationRunID < 1 || availableAt.IsZero() {
		return jobs.Job{}, ErrInvalid
	}
	if payload.Workload == "" {
		payload.Workload = "ordinary"
	}
	body, _ := json.Marshal(payload)
	request, err := policy.JobRequest(Kind, fmt.Sprintf("extraction:%d:%d", payload.DocumentVersionID, payload.ClassificationRunID), body, availableAt)
	if err != nil {
		return jobs.Job{}, err
	}
	return queue.Enqueue(ctx, request)
}

func (r *Runner) Handler() jobs.Handler {
	return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		var payload Payload
		if json.Unmarshal(job.Payload, &payload) != nil || payload.DocumentVersionID < 1 || payload.ClassificationRunID < 1 || !validWorkload(payload.Workload) {
			return jobs.Result{}, failure("invalid_extraction_job", false)
		}
		return r.run(ctx, job, payload)
	}
}

func (r *Runner) run(ctx context.Context, job jobs.Job, payload Payload) (jobs.Result, error) {
	if r.Documents == nil || r.OCR == nil || r.Classifications == nil || r.Processing == nil || r.Results == nil || r.Adapter == nil || r.Model == "" || r.ConfigurationVersion == "" || job.ID < 1 || job.Attempt < 1 {
		return jobs.Result{}, failure("extraction_configuration_invalid", false)
	}
	now := time.Now
	if r.Now != nil {
		now = r.Now
	}
	version, err := r.Documents.Version(ctx, payload.DocumentVersionID)
	if err != nil {
		return jobs.Result{}, failure("retained_version_unavailable", true)
	}
	classified, found, err := r.Classifications.Get(ctx, payload.ClassificationRunID)
	if err != nil {
		return jobs.Result{}, failure("classification_result_unavailable", true)
	}
	if !found || classified.DocumentVersionID != version.ID {
		return jobs.Result{}, failure("classification_result_invalid", false)
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
	content, err := classification.GatherContent(ctx, r.Documents, r.OCR, version)
	if err != nil {
		return jobs.Result{}, failure("extraction_content_unavailable", true)
	}
	var windows []Window
	if content.Complete && len(content.Sections) > 0 {
		windows, err = BuildOperationalWindows(version.ID, content)
		if err != nil {
			return jobs.Result{}, failure("extraction_content_invalid", false)
		}
	}
	var manifest classification.InputManifest
	manifestSuffix := ""
	if r.Manifests != nil {
		manifest, err = classification.BuildManifest(ctx, r.Documents, r.OCR, version, content, r.ConfigurationVersion)
		if err != nil {
			return jobs.Result{}, failure("extraction_manifest_unavailable", true)
		}
		manifestSuffix = ":" + manifest.Hash
	}
	subject, _ := json.Marshal(map[string]any{"classification_run_id": classified.RunID, "content_sha256": content.Hash, "content_complete": content.Complete, "window_count": len(windows), "segmentation": "operational-core-context-v1", "normalization": "ocr-offset-map-v1"})
	var sourceID *string
	if len(version.Resources) > 0 {
		sourceID = &version.Resources[0].SourceID
	}
	run, err := r.Processing.StartRun(ctx, processing.RunRequest{IdempotencyKey: fmt.Sprintf("extraction:%d:%d:%s:%s", version.ID, classified.RunID, content.Hash, r.ConfigurationVersion) + manifestSuffix, Workload: payload.Workload, Stage: "extraction", ConfigurationVersionID: r.ConfigurationVersion, SourceID: sourceID, DocumentVersionID: &version.ID, Subject: subject, CreatedAt: now().UTC()})
	if err != nil {
		return jobs.Result{}, failure("extraction_run_unavailable", true)
	}
	if r.Manifests != nil {
		if err = r.Manifests.PutManifest(ctx, run.ID, manifest); err != nil {
			return jobs.Result{}, failure("extraction_manifest_unavailable", true)
		}
	}
	if existing, ok, lookupErr := r.Results.Get(ctx, run.ID); lookupErr != nil {
		return jobs.Result{}, failure("extraction_result_unavailable", true)
	} else if ok {
		return result(existing)
	}
	jobID, queueAttempt := job.ID, job.Attempt
	attempt, err := r.Processing.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, QueueJobID: &jobID, QueueAttemptNumber: &queueAttempt, StartedAt: now().UTC()})
	if err != nil {
		return jobs.Result{}, failure("extraction_attempt_unavailable", true)
	}
	ctx = inference.WithAttempt(ctx, attempt, r.ConfigurationVersion)
	base := Result{RunID: run.ID, DocumentVersionID: version.ID, ClassificationRunID: classified.RunID, ContentSHA256: content.Hash, ContentComplete: content.Complete, CreatedAt: now().UTC()}
	if classified.Status != "classified" || classified.Relevant == nil {
		return r.uninterpreted(ctx, attempt, base, "classification_undetermined", processing.Usage{Status: "not_applicable"})
	}
	if !*classified.Relevant {
		base.Status, base.ReasonCode = "not_applicable", "not_relevant"
		return r.persist(ctx, attempt, base, "succeeded", processing.Usage{Status: "not_applicable"}, nil, "")
	}
	if !content.Complete || len(content.Sections) == 0 {
		reason := "incomplete_required_content"
		if content.Complete {
			reason = "empty_content"
		}
		return r.uninterpreted(ctx, attempt, base, reason, processing.Usage{Status: "not_applicable"})
	}
	if r.Manifests != nil && r.ReuseSources[manifest.SourceID] && manifest.Reusable {
		original, found, lookupErr := r.Manifests.EquivalentRun(ctx, run.ID, "extraction")
		if lookupErr != nil {
			return jobs.Result{}, failure("extraction_reuse_unavailable", true)
		}
		if found {
			prior, ok, e := r.Results.Get(ctx, original)
			if e != nil {
				return jobs.Result{}, failure("extraction_reuse_unavailable", true)
			}
			if ok {
				if mapped, valid := RemapResult(prior, windows, content, base); valid {
					if e = r.Manifests.RecordReuse(ctx, run.ID, original, now()); e != nil {
						return jobs.Result{}, failure("extraction_reuse_unavailable", true)
					}
					return r.persist(ctx, attempt, mapped, "succeeded", processing.Usage{Status: "not_applicable"}, nil, "")
				}
			}
		}
	}
	usageTotal := newExtractionUsageTotal()
	providerHash := sha256.New()
	returnedModel := ""
	groups := make([][]Measure, 0, len(windows))
	segmentResults := make([]SegmentResult, 0, len(windows))
	content.LegacyLiteral = r.legacyOutput
	for _, window := range windows {
		window.legacyLiteral = r.legacyOutput
		request, requestErr := WindowRequest(r.Model, window)
		if requestErr != nil {
			base.Segments = segmentResults
			usage, _ := extractionUsage(usageTotal.finish(), nil)
			return r.persist(ctx, attempt, withStatus(base, "uninterpreted", "extraction_output_schema_invalid"), "failed", usage, nil, "extraction_output_schema_invalid")
		}
		if r.DisableThinking {
			request.ChatTemplateKwargs = &inference.ChatTemplateOptions{EnableThinking: false}
		}
		response, cached, responseErr := inference.LoadCheckpoint(ctx, r.Checkpoints, "extraction", r.ConfigurationVersion, window.Ordinal, request)
		if responseErr == nil && !cached {
			response, responseErr = r.Adapter.Complete(ctx, request)
		}
		if deferred, deferErr := inference.FinishDeferral(ctx, r.Processing, attempt, responseErr); deferred {
			return jobs.Result{}, deferErr
		}
		if responseErr != nil {
			code, temporary := "provider_error", true
			var call *inference.CallError
			if errors.As(responseErr, &call) {
				code, temporary = call.Code, call.Temporary
			} else if errors.Is(responseErr, inference.ErrInvalid) {
				temporary = false
			}
			base.Segments = segmentResults
			usage, _ := extractionUsage(usageTotal.finish(), nil)
			if !temporary || job.Attempt >= job.MaxAttempts {
				reason := "provider_error"
				if temporary {
					reason = "attempts_exhausted"
				}
				return r.persist(ctx, attempt, withStatus(base, "uninterpreted", reason), "failed", usage, nil, code)
			}
			if finishErr := r.finish(ctx, attempt, now(), "failed", usage, nil, code); finishErr != nil {
				return jobs.Result{}, failure("extraction_attempt_unavailable", true)
			}
			return jobs.Result{}, failure(code, true)
		}
		if !cached {
			usageTotal.add(response.Usage)
		}
		measures, parseErr := ParseWindow(response.Content, window)
		if !r.legacyOutput && (response.FinishReason == "length" || response.FinishReason == "content_filter") {
			parseErr = ErrInvalid
		}
		providerID := response.ID
		if providerID == "" {
			providerID = "not_reported"
		}
		model := response.Model
		if model == "" {
			model = r.Model
		}
		responseDigest := sha256.Sum256([]byte(response.Content))
		segmentResults = append(segmentResults, SegmentResult{
			Ordinal: window.Ordinal, Total: window.Total, DocumentVersionID: version.ID,
			ResourceURL: window.ResourceURL, Role: window.Role, Page: window.Page, StartByte: window.ContextStartByte, EndByte: window.ContextEndByte,
			CoreStartByte: window.CoreStartByte, CoreEndByte: window.CoreEndByte, SourceStartByte: window.SourceStartByte, SourceEndByte: window.SourceEndByte,
			NormalizationMap: window.OffsetMap, ContentSHA256: window.Hash, ResponseSHA256: hex.EncodeToString(responseDigest[:]), ProviderResponseID: providerID, ReturnedModel: model,
			MeasureCount: len(measures), InputTokens: response.Usage.InputTokens, OutputTokens: response.Usage.OutputTokens, CacheReadTokens: response.Usage.CacheReadTokens,
		})
		if parseErr != nil || returnedModel != "" && returnedModel != model {
			reason := "extraction_output_schema_invalid"
			if errors.Is(parseErr, ErrEvidence) {
				reason = "extraction_evidence_invalid"
			}
			base.Segments = segmentResults
			usage, _ := extractionUsage(usageTotal.finish(), nil)
			code := InvalidOutputReason(response.Content, response.FinishReason, window, parseErr)
			if parseErr == nil {
				code = "extraction_output_model_changed"
			}
			requestBytes, marshalErr := json.Marshal(request)
			captureCtx, cancelCapture := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			captureErr := marshalErr
			if captureErr == nil {
				captureErr = r.Processing.RecordInvalidOutput(captureCtx, processing.InvalidOutput{RunID: attempt.RunID, AttemptNumber: attempt.Number, SegmentOrdinal: window.Ordinal, Request: requestBytes, Response: []byte(response.Content), FinishReason: response.FinishReason, ErrorCode: code, Cached: cached, CreatedAt: now().UTC()})
			}
			cancelCapture()
			if captureErr != nil {
				if err := r.finish(ctx, attempt, now(), "failed", usage, nil, "extraction_diagnostic_unavailable"); err != nil {
					return jobs.Result{}, failure("extraction_attempt_unavailable", true)
				}
				return jobs.Result{}, failure("extraction_diagnostic_unavailable", false)
			}
			return r.persist(ctx, attempt, withStatus(base, "uninterpreted", reason), "failed", usage, nil, code)
		}
		if !cached {
			if checkpointErr := inference.SaveCheckpoint(ctx, r.Checkpoints, "extraction", r.ConfigurationVersion, window.Ordinal, request, response); checkpointErr != nil {
				usage, _ := extractionUsage(usageTotal.finish(), nil)
				if err := r.finish(ctx, attempt, now(), "failed", usage, nil, "extraction_checkpoint_unavailable"); err != nil {
					return jobs.Result{}, failure("extraction_attempt_unavailable", true)
				}
				return jobs.Result{}, failure("extraction_checkpoint_unavailable", true)
			}
		}
		returnedModel = model
		providerHash.Write([]byte(providerID))
		providerHash.Write(responseDigest[:])
		groups = append(groups, measures)
	}
	measures, mergeErr := Merge(groups, content, len(windows))
	base.Segments = segmentResults
	if mergeErr != nil {
		reason := "extraction_evidence_invalid"
		if errors.Is(mergeErr, ErrMergeDuplicate) {
			reason = "extraction_merge_duplicate"
		} else if errors.Is(mergeErr, ErrMergeConflict) {
			reason = "extraction_merge_conflict"
		}
		usage, _ := extractionUsage(usageTotal.finish(), nil)
		return r.persist(ctx, attempt, withStatus(base, "uninterpreted", reason), "failed", usage, nil, reason)
	}
	reported := usageTotal.finish()
	base.Status, base.ReasonCode, base.Measures = "extracted", "measures_extracted", measures
	base.ProviderResponseID, base.ReturnedModel = "segmented:"+hex.EncodeToString(providerHash.Sum(nil)), returnedModel
	base.InputTokens, base.OutputTokens, base.CacheReadTokens = reported.InputTokens, reported.OutputTokens, reported.CacheReadTokens
	usage, price := extractionUsage(reported, r.PriceVersion)
	return r.persist(ctx, attempt, base, "succeeded", usage, price, "")
}

type extractionUsageTotal struct {
	calls                                          int
	input, output, cache                           int64
	inputKnown, outputKnown, cacheKnown, cacheSeen bool
}

func newExtractionUsageTotal() extractionUsageTotal {
	return extractionUsageTotal{inputKnown: true, outputKnown: true, cacheKnown: true}
}

func (u *extractionUsageTotal) add(value inference.Usage) {
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

func (u extractionUsageTotal) finish() inference.Usage {
	var result inference.Usage
	if u.calls > 0 && u.inputKnown {
		result.InputTokens = extractionInt64(u.input)
	}
	if u.calls > 0 && u.outputKnown {
		result.OutputTokens = extractionInt64(u.output)
	}
	if u.calls > 0 && u.cacheKnown && u.cacheSeen {
		result.CacheReadTokens = extractionInt64(u.cache)
	}
	return result
}

func extractionInt64(value int64) *int64 { return &value }

func (r *Runner) uninterpreted(ctx context.Context, attempt processing.Attempt, base Result, reason string, usage processing.Usage) (jobs.Result, error) {
	return r.persist(ctx, attempt, withStatus(base, "uninterpreted", reason), "succeeded", usage, nil, "")
}

func withStatus(value Result, status, reason string) Result {
	value.Status, value.ReasonCode = status, reason
	return value
}

func (r *Runner) persist(ctx context.Context, attempt processing.Attempt, value Result, outcome string, usage processing.Usage, price *string, code string) (jobs.Result, error) {
	if err := r.finish(ctx, attempt, value.CreatedAt, outcome, usage, price, code); err != nil {
		return jobs.Result{}, failure("extraction_attempt_unavailable", true)
	}
	if err := r.Results.Put(ctx, value); err != nil {
		return jobs.Result{}, failure("extraction_result_unavailable", true)
	}
	return result(value)
}

func (r *Runner) finish(ctx context.Context, attempt processing.Attempt, at time.Time, outcome string, usage processing.Usage, price *string, code string) error {
	_, err := r.Processing.FinishAttempt(ctx, processing.AttemptFinish{RunID: attempt.RunID, Number: attempt.Number, FinishedAt: at.UTC(), Outcome: outcome, Usage: usage, PriceVersionID: price, ErrorCode: code})
	return err
}

func extractionUsage(reported inference.Usage, price *string) (processing.Usage, *string) {
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

func result(value Result) (jobs.Result, error) {
	payload, _ := json.Marshal(struct {
		RunID             int64  `json:"run_id"`
		DocumentVersionID int64  `json:"document_version_id"`
		Status            string `json:"status"`
		ReasonCode        string `json:"reason_code"`
		MeasureCount      int    `json:"measure_count"`
	}{value.RunID, value.DocumentVersionID, value.Status, value.ReasonCode, len(value.Measures)})
	kind := "extraction_ready"
	if value.Status == "uninterpreted" {
		kind = "extraction_uninterpreted"
	}
	return jobs.JSONResult(json.RawMessage(payload), jobs.Effect{Key: fmt.Sprintf("extraction-run:%d", value.RunID), Kind: kind, Payload: payload})
}

func failure(code string, temporary bool) error {
	return &jobs.HandlerError{Failure: jobs.Failure{Code: code, Detail: "document measure extraction failed", Temporary: temporary}}
}

func validWorkload(value string) bool {
	return value == "ordinary" || value == "bootstrap" || value == "evaluation" || value == "reprocessing"
}
