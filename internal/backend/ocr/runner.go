package ocr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
)

type documentStore interface {
	Version(context.Context, int64) (documents.Version, error)
	Read(context.Context, int64, string) ([]byte, error)
}

type processingStore interface {
	StartRun(context.Context, processing.RunRequest) (processing.Run, error)
	StartAttempt(context.Context, processing.AttemptStart) (processing.Attempt, error)
	FinishAttempt(context.Context, processing.AttemptFinish) (processing.Attempt, error)
}

type resultStore interface {
	PutPage(context.Context, PageResult) error
	Page(context.Context, int64, int) (PageResult, bool, error)
	PutResource(context.Context, ResourceResult) error
	Resource(context.Context, int64) (ResourceResult, bool, error)
}

type Runner struct {
	ReuseSources                    map[string]bool
	Artifacts                       *Store
	ProviderScope, RendererIdentity string
	Documents                       documentStore
	Processing                      processingStore
	Results                         resultStore
	Adapter                         inference.Adapter
	Renderer                        Renderer
	Model                           string
	ConfigurationVersion            string
	PriceVersion                    *string
	Now                             func() time.Time
}

type Payload struct {
	DocumentVersionID int64  `json:"document_version_id"`
	ResourceURL       string `json:"resource_url"`
	Workload          string `json:"workload"`
}

func Enqueue(ctx context.Context, queue *jobs.Store, policy inference.RetryPolicy, payload Payload, availableAt time.Time) (jobs.Job, error) {
	if queue == nil || payload.DocumentVersionID < 1 || payload.ResourceURL == "" || availableAt.IsZero() {
		return jobs.Job{}, ErrInvalid
	}
	if payload.Workload == "" {
		payload.Workload = "ordinary"
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return jobs.Job{}, ErrInvalid
	}
	key := fmt.Sprintf("ocr:%d:%s", payload.DocumentVersionID, shortHash(payload.ResourceURL))
	request, err := policy.JobRequest(Kind, key, body, availableAt)
	if err != nil {
		return jobs.Job{}, err
	}
	return queue.Enqueue(ctx, request)
}

func (r *Runner) Handler() jobs.Handler {
	return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		var payload Payload
		if json.Unmarshal(job.Payload, &payload) != nil || payload.DocumentVersionID < 1 || payload.ResourceURL == "" || (payload.Workload != "ordinary" && payload.Workload != "bootstrap" && payload.Workload != "evaluation" && payload.Workload != "reprocessing") {
			return jobs.Result{}, failure("invalid_ocr_job", false)
		}
		return r.run(ctx, job, payload)
	}
}

func (r *Runner) run(ctx context.Context, job jobs.Job, payload Payload) (jobs.Result, error) {
	if r.Documents == nil || r.Processing == nil || r.Results == nil || r.Adapter == nil || r.Renderer == nil || r.Model == "" || r.ConfigurationVersion == "" || job.ID < 1 || job.Attempt < 1 {
		return jobs.Result{}, failure("ocr_configuration_invalid", false)
	}
	now := time.Now
	if r.Now != nil {
		now = r.Now
	}
	version, err := r.Documents.Version(ctx, payload.DocumentVersionID)
	if err != nil {
		return jobs.Result{}, failure("retained_version_unavailable", true)
	}
	configuredRunner := *r
	source := ""
	for _, ref := range version.Resources {
		if ref.Role == "original" {
			source = ref.SourceID
			break
		}
	}
	if r.ReuseSources != nil && !r.ReuseSources[source] {
		configuredRunner.Artifacts = nil
	}
	r = &configuredRunner
	var reference *documents.Reference
	for index := range version.Resources {
		if version.Resources[index].URL == payload.ResourceURL {
			reference = &version.Resources[index]
			break
		}
	}
	if reference == nil {
		return jobs.Result{}, failure("retained_resource_unknown", false)
	}
	policyKey := ""
	if reference.Inference != nil && !strings.HasSuffix(reference.Inference.Policy, ":default") {
		policyKey = ":" + shortHash(reference.Inference.Policy)
	}
	subjectValue := map[string]any{"resource_url": payload.ResourceURL, "resource_sha256": reference.Hash}
	if policyKey != "" {
		subjectValue["eligibility_policy"] = policyKey
	}
	subject, _ := json.Marshal(subjectValue)
	run, err := r.Processing.StartRun(ctx, processing.RunRequest{
		IdempotencyKey: fmt.Sprintf("ocr:%d:%s:%s", version.ID, shortHash(payload.ResourceURL), r.ConfigurationVersion) + policyKey,
		Workload:       payload.Workload, Stage: "ocr", ConfigurationVersionID: r.ConfigurationVersion,
		SourceID: &reference.SourceID, DocumentVersionID: &version.ID, Subject: subject, CreatedAt: now().UTC(),
	})
	if err != nil {
		return jobs.Result{}, failure("ocr_run_unavailable", true)
	}
	if existing, ok, lookupErr := r.Results.Resource(ctx, run.ID); lookupErr != nil {
		return jobs.Result{}, failure("ocr_result_unavailable", true)
	} else if ok {
		return result(existing)
	}
	jobID, attemptNumber := job.ID, job.Attempt
	attempt, err := r.Processing.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, QueueJobID: &jobID, QueueAttemptNumber: &attemptNumber, StartedAt: now().UTC()})
	if err != nil {
		return jobs.Result{}, failure("ocr_attempt_unavailable", true)
	}
	ctx = inference.WithAttempt(ctx, attempt, r.ConfigurationVersion)

	if !reference.InferenceEligible() {
		if err = r.finish(ctx, attempt, now(), "succeeded", processing.Usage{Status: "not_applicable"}, nil, ""); err != nil {
			return jobs.Result{}, failure("ocr_attempt_unavailable", true)
		}
		skipped := ResourceResult{RunID: run.ID, DocumentVersionID: version.ID, ResourceURL: reference.URL, Status: "skipped", CreatedAt: now()}
		if err = r.Results.PutResource(ctx, skipped); err != nil {
			return jobs.Result{}, failure("ocr_result_unavailable", true)
		}
		return result(skipped)
	}
	if reference.Missing != "" {
		if err = r.finish(ctx, attempt, now(), "succeeded", processing.Usage{Status: "not_applicable"}, nil, ""); err != nil {
			return jobs.Result{}, failure("ocr_attempt_unavailable", true)
		}
		missing := ResourceResult{RunID: run.ID, DocumentVersionID: version.ID, ResourceURL: reference.URL, Status: "missing", ErrorCode: "attachment_" + reference.Missing, CreatedAt: now().UTC()}
		if err = r.Results.PutResource(ctx, missing); err != nil {
			return jobs.Result{}, failure("ocr_result_unavailable", true)
		}
		return result(missing)
	}
	var identities []renderedIdentity
	var pages []PageImage
	var renderErr error
	warm := false
	if r.Artifacts != nil {
		if r.ProviderScope == "" || r.RendererIdentity == "" {
			return r.failAttempt(ctx, attempt, now(), "ocr_cache_configuration_invalid", false)
		}
		identities, warm, err = r.Artifacts.warmManifest(ctx, r.manifestKey(reference.Hash, reference.MediaType))
		if err != nil {
			return r.failAttempt(ctx, attempt, now(), "ocr_cache_unavailable", true)
		}
	}
	if warm {
		for _, identity := range identities {
			pages = append(pages, PageImage{Number: identity.Number, MediaType: identity.MediaType})
		}
	} else {
		body, err := r.Documents.Read(ctx, version.ID, reference.URL)
		if err != nil {
			return r.failAttempt(ctx, attempt, now(), "retained_resource_unavailable", true)
		}
		pages, renderErr = r.pages(ctx, reference.MediaType, body)
		if renderErr == nil && r.Artifacts != nil {
			identities = r.identities(pages)
			if err = r.Artifacts.putManifest(ctx, r.manifestKey(reference.Hash, reference.MediaType), identities, now()); err != nil {
				return r.failAttempt(ctx, attempt, now(), "ocr_manifest_conflict", false)
			}
		}
	}
	if renderErr != nil {
		if errors.Is(renderErr, ErrRendererUnavailable) || errors.Is(renderErr, context.Canceled) || errors.Is(renderErr, context.DeadlineExceeded) {
			return r.failAttempt(ctx, attempt, now(), "renderer_unavailable", true)
		}
		if err = r.finish(ctx, attempt, now(), "succeeded", processing.Usage{Status: "not_applicable"}, nil, ""); err != nil {
			return jobs.Result{}, failure("ocr_attempt_unavailable", true)
		}
		unreadable := ResourceResult{RunID: run.ID, DocumentVersionID: version.ID, ResourceURL: reference.URL, Status: "unreadable", ErrorCode: "rasterization_failed", CreatedAt: now().UTC()}
		if err = r.Results.PutResource(ctx, unreadable); err != nil {
			return jobs.Result{}, failure("ocr_result_unavailable", true)
		}
		return result(unreadable)
	}
	usage := usageTotal{}
	unreadablePages := 0
	for pageIndex, page := range pages {
		if existing, ok, lookupErr := r.Results.Page(ctx, run.ID, page.Number); lookupErr != nil {
			return r.failAttempt(ctx, attempt, now(), "ocr_result_unavailable", true)
		} else if ok {
			if existing.Status == "unreadable" {
				unreadablePages++
			}
			continue
		}
		var artifact Artifact
		if r.Artifacts != nil {
			artifact, err = r.Artifacts.ClaimArtifact(ctx, identities[pageIndex].Identity, now(), time.Minute)
			if err != nil {
				return r.failAttemptWithUsage(ctx, attempt, now(), "ocr_cache_unavailable", true, usage)
			}
			if artifact.Ready {
				association := PageResult{RunID: run.ID, DocumentVersionID: version.ID, PageNumber: page.Number, ResourceURL: reference.URL, CreatedAt: now()}
				if err = r.Artifacts.AssociateArtifact(ctx, artifact.Key, association, true); err != nil {
					return r.failAttemptWithUsage(ctx, attempt, now(), "ocr_association_failed", true, usage)
				}
				if artifact.Page.Status == "unreadable" {
					unreadablePages++
				}
				continue
			}
			if !artifact.Owned {
				_, deferErr := inference.FinishDeferral(ctx, r.Processing, attempt, &jobs.DeferredError{Until: artifact.LeaseExpires, Code: "ocr_artifact_busy"})
				return jobs.Result{}, deferErr
			}
		}
		makePage := func(response inference.Response) PageResult {
			text, status := response.Content, "complete"
			if strings.TrimSpace(text) == "" {
				text, status = "", "unreadable"
			}
			responseID := response.ID
			if responseID == "" {
				responseID = "not_reported"
			}
			return PageResult{RunID: run.ID, DocumentVersionID: version.ID, PageNumber: page.Number, ResourceURL: reference.URL, Status: status, MediaType: page.MediaType, InputSHA256: digest(page.Bytes), OutputSHA256: digest([]byte(text)), ExtractedText: text, ProviderResponseID: responseID, ReturnedModel: response.Model, InputTokens: response.Usage.InputTokens, OutputTokens: response.Usage.OutputTokens, CacheReadTokens: response.Usage.CacheReadTokens, CreatedAt: now()}
		}
		var response inference.Response
		var callErr error
		if r.Artifacts != nil {
			response, callErr = r.callOwned(ctx, artifact, pageRequest(r.Model, page), func(response inference.Response) error {
				return r.Artifacts.CompleteArtifact(ctx, artifact, makePage(response), now())
			})
		} else {
			response, callErr = r.Adapter.Complete(ctx, pageRequest(r.Model, page))
		}

		if deferred, deferErr := inference.FinishDeferral(ctx, r.Processing, attempt, callErr); deferred {
			return jobs.Result{}, deferErr
		}
		if callErr != nil {
			code, temporary := "ocr_provider_error", true
			var provider *inference.CallError
			if errors.As(callErr, &provider) {
				code, temporary = provider.Code, provider.Temporary
			} else if errors.Is(callErr, inference.ErrInvalid) {
				code, temporary = "ocr_request_invalid", false
			}
			return r.failAttemptWithUsage(ctx, attempt, now(), code, temporary, usage)
		}
		usage.addResponse(response)

		pageResult := makePage(response)
		if pageResult.Status == "unreadable" {
			unreadablePages++
		}
		if r.Artifacts != nil {
			err = r.Artifacts.AssociateArtifact(ctx, artifact.Key, pageResult, false)
		} else {
			err = r.Results.PutPage(ctx, pageResult)
		}
		if err != nil {
			return r.failAttemptWithUsage(ctx, attempt, now(), "ocr_result_unavailable", true, usage)
		}

	}
	finalUsage, price := usage.finish(r.PriceVersion)
	if err = r.finish(ctx, attempt, now(), "succeeded", finalUsage, price, ""); err != nil {
		return jobs.Result{}, failure("ocr_attempt_unavailable", true)
	}
	status := "complete"
	if unreadablePages == len(pages) {
		status = "unreadable"
	} else if unreadablePages > 0 {
		status = "partial_unreadable"
	}
	resource := ResourceResult{RunID: run.ID, DocumentVersionID: version.ID, ResourceURL: reference.URL, Status: status, PageCount: len(pages), CreatedAt: now().UTC()}
	if status == "unreadable" {
		resource.ErrorCode = "all_pages_unreadable"
	}
	if err = r.Results.PutResource(ctx, resource); err != nil {
		return jobs.Result{}, failure("ocr_result_unavailable", true)
	}
	return result(resource)
}

func (r *Runner) pages(ctx context.Context, mediaType string, body []byte) ([]PageImage, error) {
	mediaType = strings.ToLower(strings.TrimSpace(strings.Split(mediaType, ";")[0]))
	switch mediaType {
	case "application/pdf":
		return r.Renderer.Render(ctx, body)
	case "image/png", "image/jpeg":
		if len(body) == 0 || len(body) > maxPageImageBytes {
			return nil, ErrRasterization
		}
		return []PageImage{{Number: 1, MediaType: mediaType, Bytes: body}}, nil
	default:
		return nil, ErrRasterization
	}
}

func pageRequest(model string, page PageImage) inference.Request {
	content, _ := json.Marshal([]any{
		map[string]any{"type": "text", "text": PromptBody},
		map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:" + page.MediaType + ";base64," + base64.StdEncoding.EncodeToString(page.Bytes), "format": page.MediaType}},
	})
	keepSpecialTokens := false
	return inference.Request{Model: model, Messages: []inference.Message{{Role: "user", Content: content}}, MaxTokens: 4096, SkipSpecialTokens: &keepSpecialTokens}
}

func (r *Runner) failAttempt(ctx context.Context, attempt processing.Attempt, finished time.Time, code string, temporary bool) (jobs.Result, error) {
	return r.failAttemptWithUsage(ctx, attempt, finished, code, temporary, usageTotal{})
}

func (r *Runner) failAttemptWithUsage(ctx context.Context, attempt processing.Attempt, finished time.Time, code string, temporary bool, usage usageTotal) (jobs.Result, error) {
	providerUsage, _ := usage.finish(nil)
	if usage.requests == 0 {
		providerUsage = processing.Usage{Status: "unavailable"}
	}
	if err := r.finish(ctx, attempt, finished, "failed", providerUsage, nil, code); err != nil {
		return jobs.Result{}, failure("ocr_attempt_unavailable", true)
	}
	return jobs.Result{}, failure(code, temporary)
}

func (r *Runner) finish(ctx context.Context, attempt processing.Attempt, at time.Time, outcome string, usage processing.Usage, price *string, code string) error {
	_, err := r.Processing.FinishAttempt(ctx, processing.AttemptFinish{RunID: attempt.RunID, Number: attempt.Number, FinishedAt: at.UTC(), Outcome: outcome, Usage: usage, PriceVersionID: price, ErrorCode: code})
	return err
}

func result(value ResourceResult) (jobs.Result, error) {
	payload, _ := json.Marshal(struct {
		RunID             int64  `json:"run_id"`
		DocumentVersionID int64  `json:"document_version_id"`
		ResourceURL       string `json:"resource_url"`
		Status            string `json:"status"`
		PageCount         int    `json:"page_count"`
	}{value.RunID, value.DocumentVersionID, value.ResourceURL, value.Status, value.PageCount})
	return jobs.JSONResult(json.RawMessage(payload), jobs.Effect{Key: fmt.Sprintf("ocr-run:%d", value.RunID), Kind: "ocr_resource_ready", Payload: payload})
}

func failure(code string, temporary bool) error {
	return &jobs.HandlerError{Failure: jobs.Failure{Code: code, Detail: "OCR resource processing failed", Temporary: temporary}}
}

func shortHash(value string) string { return stableID("", value) }

type usageTotal struct {
	requests                    int64
	input, output, cache        int64
	inputReports, outputReports int64
	cacheSeen                   bool
}

func (u *usageTotal) addResponse(response inference.Response) {
	u.requests++
	if response.Usage.InputTokens != nil {
		u.input += *response.Usage.InputTokens
		u.inputReports++
	}
	if response.Usage.OutputTokens != nil {
		u.output += *response.Usage.OutputTokens
		u.outputReports++
	}
	if response.Usage.CacheReadTokens != nil {
		u.cache += *response.Usage.CacheReadTokens
		u.cacheSeen = true
	}
}

func (u usageTotal) finish(price *string) (processing.Usage, *string) {
	status := "reported"
	if u.requests == 0 {
		return processing.Usage{Status: "not_applicable"}, nil
	}
	usage := processing.Usage{OtherUnits: map[string]int64{"requests": u.requests}}
	if u.inputReports == u.requests {
		usage.InputTokens = &u.input
	} else {
		status = "partial"
	}
	if u.outputReports == u.requests {
		usage.OutputTokens = &u.output
	} else {
		status = "partial"
	}
	if u.cacheSeen {
		usage.CacheReadTokens = &u.cache
		price = nil
	}
	usage.Status = status
	if status != "reported" {
		price = nil
	}
	return usage, price
}
