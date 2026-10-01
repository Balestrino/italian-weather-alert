package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
)

type CallLedger interface {
	StartCall(context.Context, processing.CallStart) (int64, error)
	FinishCall(context.Context, processing.CallFinish) error
}

type attemptKey struct{}
type callAttempt struct {
	run           int64
	number        int
	ordinal       atomic.Int32
	configuration string
}

func WithAttempt(ctx context.Context, attempt processing.Attempt, configuration ...string) context.Context {
	c := ""
	if len(configuration) > 0 {
		c = configuration[0]
	}
	return context.WithValue(ctx, attemptKey{}, &callAttempt{run: attempt.RunID, number: attempt.Number, configuration: c})
}

type RecordedAdapter struct {
	Adapter Adapter
	Ledger  CallLedger
}

func (a *RecordedAdapter) Name() string { return a.Adapter.Name() }

func startRecorded(ctx context.Context, ledger CallLedger, provider, model string, input any) (int64, error) {
	attempt, ok := ctx.Value(attemptKey{}).(*callAttempt)
	if !ok || ledger == nil {
		return 0, &CallError{Code: "call_context_missing"}
	}
	wire, err := json.Marshal(input)
	if err != nil {
		return 0, ErrInvalid
	}
	hash := sha256.Sum256(wire)
	id, err := ledger.StartCall(ctx, processing.CallStart{RunID: attempt.run, AttemptNumber: attempt.number, Ordinal: int(attempt.ordinal.Add(1)), InputSHA256: hex.EncodeToString(hash[:]), Provider: provider, RequestedModel: model, StartedAt: time.Now().UTC()})
	if err != nil {
		var deferred *jobs.DeferredError
		if errors.As(err, &deferred) {
			return 0, err
		}
		return 0, &CallError{Code: "call_intent_unavailable", Temporary: true}
	}
	return id, nil
}

var receiptIdentifier = regexp.MustCompile(`^[A-Za-z0-9_.:/-]{0,200}$`)

func receiptID(value string) string {
	if receiptIdentifier.MatchString(value) {
		return value
	}
	return ""
}

func finishRecorded(ctx context.Context, ledger CallLedger, id int64, response Response, callErr error, embedding bool) error {
	status := "reported"
	if response.Usage.InputTokens == nil || (!embedding && response.Usage.OutputTokens == nil) {
		status = "partial"
	}
	if response.Usage.InputTokens == nil && response.Usage.OutputTokens == nil && response.Usage.CacheReadTokens == nil {
		status = "unavailable"
	}
	f := processing.CallFinish{ID: id, FinishedAt: time.Now().UTC(), State: "received", HTTPStatus: response.Diagnostic.HTTPStatus,
		Category: response.Diagnostic.Category, ProviderCode: response.Diagnostic.ProviderCode, RequestID: receiptID(response.Diagnostic.RequestID),
		ResponseID: receiptID(response.ID), ReturnedModel: receiptID(response.Model), FinishReason: receiptID(response.FinishReason),
		Usage: processing.Usage{Status: status, InputTokens: response.Usage.InputTokens, OutputTokens: response.Usage.OutputTokens, CacheReadTokens: response.Usage.CacheReadTokens}}
	if callErr != nil {
		f.State, f.ErrorCode = "failed", "provider_error"
		var call *CallError
		if errors.As(callErr, &call) {
			f.ErrorCode = receiptID(call.Code)
			f.RetryAfter = call.RetryAfter
			if f.HTTPStatus == 0 {
				f.HTTPStatus = call.StatusCode
			}
		}
		if f.HTTPStatus == 0 {
			f.State, f.Category = "uncertain", "transport"
		}
	}
	// A canceled handler must still get a bounded chance to persist received usage.
	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := ledger.FinishCall(finishCtx, f); err != nil {
		return &CallError{Code: "call_receipt_unavailable", Temporary: true}
	}
	return callErr
}

func (a *RecordedAdapter) Complete(ctx context.Context, request Request) (Response, error) {
	if err := ValidateRequest(request); err != nil {
		return Response{}, err
	}
	id, err := startRecorded(ctx, a.Ledger, a.Name(), request.Model, request)
	if err != nil {
		return Response{}, err
	}
	response, err := a.Adapter.Complete(ctx, request)
	err = finishRecorded(ctx, a.Ledger, id, response, err, false)
	if err == nil {
		response.CallID = id
	}
	return response, err
}

type RecordedEmbedder struct {
	Adapter Embedder
	Ledger  CallLedger
}

func (a *RecordedEmbedder) Name() string { return a.Adapter.Name() }
func (a *RecordedEmbedder) Embed(ctx context.Context, model string, input []string, dimensions int) (EmbeddingResponse, error) {
	if model == "" || len(input) == 0 || len(input) > 100 || dimensions < 1 || dimensions > 8192 {
		return EmbeddingResponse{}, ErrInvalid
	}
	id, err := startRecorded(ctx, a.Ledger, a.Name(), model, struct {
		Model      string
		Input      []string
		Dimensions int
	}{model, input, dimensions})
	if err != nil {
		return EmbeddingResponse{}, err
	}
	response, err := a.Adapter.Embed(ctx, model, input, dimensions)
	return response, finishRecorded(ctx, a.Ledger, id, Response{Model: response.Model, Usage: Usage{InputTokens: response.InputTokens}, Diagnostic: response.Diagnostic}, err, true)
}
