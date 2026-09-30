package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
)

type GateStore interface {
	AcquireGate(context.Context, string, string, time.Time, processing.GatePolicy) (processing.GatePermit, error)
	RecordGate(context.Context, processing.GatePermit, string, time.Duration, time.Time, processing.GatePolicy) error
	Rejected(context.Context, string, string, string, string) (processing.Rejection, bool, error)
	Reject(context.Context, processing.Rejection) error
}
type Gate struct {
	Store  GateStore
	Scope  string
	Policy processing.GatePolicy
}
type GatedAdapter struct {
	Adapter Adapter
	Gate    Gate
}

func (a *GatedAdapter) Name() string { return a.Adapter.Name() }

type GatedEmbedder struct {
	Adapter Embedder
	Gate    Gate
}

func (a *GatedEmbedder) Name() string { return a.Adapter.Name() }

// GateScope is an opaque identifier for an endpoint origin and an operator-named
// credential scope. No API key or key fingerprint is stored.
func GateScope(endpoint, alias string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || alias == "" || !opaqueID.MatchString(alias) {
		return "", ErrInvalid
	}
	h := sha256.Sum256([]byte(u.Scheme + "://" + u.Host + "\x00" + alias))
	return "provider-" + hex.EncodeToString(h[:16]), nil
}

type localReuseKey struct{}

// LocalReuseOnly prevents paid fallback when materializing an equivalent version.
func LocalReuseOnly(ctx context.Context) context.Context {
	return context.WithValue(ctx, localReuseKey{}, true)
}

func (g Gate) before(ctx context.Context, model string, input any) (processing.GatePermit, processing.Rejection, error) {
	if only, _ := ctx.Value(localReuseKey{}).(bool); only {
		return processing.GatePermit{}, processing.Rejection{}, &CallError{Code: "equivalent_reuse_unavailable"}
	}
	a, ok := ctx.Value(attemptKey{}).(*callAttempt)
	if !ok || a.configuration == "" || g.Store == nil {
		return processing.GatePermit{}, processing.Rejection{}, &CallError{Code: "call_context_missing"}
	}
	wire, err := json.Marshal(input)
	if err != nil {
		return processing.GatePermit{}, processing.Rejection{}, ErrInvalid
	}
	hash := sha256.Sum256(wire)
	key := processing.Rejection{Scope: g.Scope, Model: model, ConfigurationID: a.configuration, InputSHA256: hex.EncodeToString(hash[:])}
	old, found, err := g.Store.Rejected(ctx, key.Scope, key.Model, key.ConfigurationID, key.InputSHA256)
	if err != nil {
		return processing.GatePermit{}, key, &CallError{Code: "provider_gate_unavailable", Temporary: true}
	}
	if found {
		return processing.GatePermit{}, key, &CallError{Code: "provider_cached_rejection", StatusCode: old.HTTPStatus, Diagnostic: Diagnostic{HTTPStatus: old.HTTPStatus, Category: old.Category}}
	}
	p, err := g.Store.AcquireGate(ctx, g.Scope, model, time.Now(), g.Policy)
	var blocked *processing.GateBlocked
	if errors.As(err, &blocked) {
		until := blocked.State.AvailableAt
		if blocked.State.ProbeExpiresAt != nil && blocked.State.ProbeExpiresAt.After(until) {
			until = *blocked.State.ProbeExpiresAt
		}
		if until.Before(time.Now().Add(time.Second)) {
			until = time.Now().Add(time.Minute)
		}
		return p, key, &jobs.DeferredError{Until: until, Code: "provider_scope_held"}
	}
	if err != nil {
		return p, key, &CallError{Code: "provider_gate_unavailable", Temporary: true}
	}
	return p, key, nil
}
func (g Gate) after(ctx context.Context, p processing.GatePermit, key processing.Rejection, err error) error {
	var deferred *jobs.DeferredError
	if errors.As(err, &deferred) {
		return err
	}
	category := ""
	retry := time.Duration(0)
	var call *CallError
	if err != nil {
		category = "invalid_response"
		if errors.As(err, &call) {
			category = call.Diagnostic.Category
			retry = call.RetryAfter
			if category == "" {
				if call.Temporary {
					category = "transport"
				} else {
					category = "invalid_response"
				}
			}
		}
	}
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if e := g.Store.RecordGate(writeCtx, p, category, retry, time.Now(), g.Policy); e != nil {
		return &CallError{Code: "provider_gate_unavailable", Temporary: true}
	}
	if call != nil && !call.Temporary && (category == "request" || category == "permission" || category == "rejected") {
		key.ErrorCode, key.HTTPStatus, key.Category, key.CreatedAt = call.Code, call.StatusCode, category, time.Now()
		if e := g.Store.Reject(writeCtx, key); e != nil {
			return &CallError{Code: "provider_gate_unavailable", Temporary: true}
		}
	}
	return err
}
func (a *GatedAdapter) Complete(ctx context.Context, r Request) (Response, error) {
	if err := ValidateRequest(r); err != nil {
		return Response{}, err
	}
	p, key, err := a.Gate.before(ctx, r.Model, r)
	if err != nil {
		return Response{}, err
	}
	response, err := a.Adapter.Complete(ctx, r)
	return response, a.Gate.after(ctx, p, key, err)
}
func (a *GatedEmbedder) Embed(ctx context.Context, model string, input []string, dimensions int) (EmbeddingResponse, error) {
	p, key, err := a.Gate.before(ctx, model, struct {
		Model      string
		Input      []string
		Dimensions int
	}{model, input, dimensions})
	if err != nil {
		return EmbeddingResponse{}, err
	}
	response, err := a.Adapter.Embed(ctx, model, input, dimensions)
	return response, a.Gate.after(ctx, p, key, err)
}

// FinishDeferral gives runners a common non-failure retry path while preserving
// any receipts from earlier successful pages in this attempt.
func FinishDeferral(ctx context.Context, store interface {
	FinishAttempt(context.Context, processing.AttemptFinish) (processing.Attempt, error)
}, attempt processing.Attempt, err error) (bool, error) {
	var deferred *jobs.DeferredError
	if !errors.As(err, &deferred) {
		return false, nil
	}
	_, finishErr := store.FinishAttempt(ctx, processing.AttemptFinish{RunID: attempt.RunID, Number: attempt.Number, FinishedAt: time.Now().UTC(), Outcome: "failed", Usage: processing.Usage{Status: "not_applicable"}, ErrorCode: "provider_deferred"})
	if finishErr != nil {
		return true, &jobs.HandlerError{Failure: jobs.Failure{Code: "processing_attempt_unavailable", Detail: "cannot finish deferred attempt", Temporary: true}}
	}
	return true, deferred
}
