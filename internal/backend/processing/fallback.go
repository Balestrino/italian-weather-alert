package processing

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
)

type FallbackBinding struct {
	Scope, Model string
	Sources      []string
}

const fallbackPrimaryBlockedSQL = `EXISTS(SELECT 1 FROM processing_provider_gates p
 WHERE p.scope=$2 AND p.model IN ('*',$3)
 AND (p.state='held' AND p.reason IN ('quota','authentication','rate_limit','availability','transport')
 OR p.state='open' AND GREATEST(p.available_at,COALESCE(p.probe_expires_at,p.available_at))>$4))`

const fallbackSourceSQL = `EXISTS(SELECT 1 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 WHERE v.id=` + recoveryJobVersionSQL + ` AND d.source_id=ANY($7::text[]))`

const fallbackLocalBlockedSQL = `EXISTS(SELECT 1 FROM processing_provider_gates l
 WHERE l.scope=$5 AND l.model IN ('*',$6)
 AND (l.state='held' OR l.state='open' AND GREATEST(l.available_at,COALESCE(l.probe_expires_at,l.available_at))>$4))`

type providerFreeReuseKey struct{}

// WithProviderFreeReuse allows compatible persisted results, never provider calls.
func WithProviderFreeReuse(ctx context.Context) context.Context {
	return context.WithValue(ctx, providerFreeReuseKey{}, true)
}

func ProviderFreeReuseOnly(ctx context.Context) bool {
	only, _ := ctx.Value(providerFreeReuseKey{}).(bool)
	return only
}

func (s *Store) fallbackSelection(ctx context.Context, binding GateBinding, job jobs.Job, now time.Time) (eligible, selected bool, err error) {
	if binding.Fallback == nil {
		return false, false, nil
	}
	err = s.pool.QueryRow(ctx, `SELECT `+fallbackSourceSQL+`, `+fallbackSourceSQL+` AND `+fallbackPrimaryBlockedSQL+`
 FROM processing_jobs j WHERE j.id=$1 AND $5::text<>'' AND $6::text<>''`, job.ID, binding.Scope, binding.Model, now.UTC(), binding.Fallback.Scope, binding.Fallback.Model, binding.Fallback.Sources).Scan(&eligible, &selected)
	return
}

func (s *Store) ShouldFallback(ctx context.Context, binding GateBinding, job jobs.Job, now time.Time) (bool, error) {
	_, selected, err := s.fallbackSelection(ctx, binding, job, now)
	return selected, err
}

// Select entire runners, not transports inside a run. Each alternative keeps
// its own immutable configuration, model, gates, receipts and cache identity.
func (s *Store) FallbackHandler(primary, local jobs.Handler, binding GateBinding) jobs.Handler {
	return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		eligible, selected, err := s.fallbackSelection(ctx, binding, job, time.Now())
		if err != nil {
			return jobs.Result{}, fallbackRoutingFailure()
		}
		if ProviderFreeReuseOnly(ctx) {
			// Prefer a compatible primary cache even while held. A cache miss may
			// try the alternative cache within the same queue attempt, with both
			// transports blocked by the provider-free context.
			result, primaryErr := primary(ctx, job)
			if !eligible {
				return result, primaryErr
			}
			var failure *jobs.HandlerError
			missing := errors.As(primaryErr, &failure) && failure.Code == "equivalent_reuse_unavailable"
			if primaryErr == nil && fallbackProviderFailure(result, nil) {
				lookupErr := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM processing_run_attempts
 WHERE queue_job_id=$1 AND queue_attempt_number=$2 AND outcome='failed' AND error_code='equivalent_reuse_unavailable')`, job.ID, job.Attempt).Scan(&missing)
				if lookupErr != nil {
					return jobs.Result{}, fallbackRoutingFailure()
				}
			}
			if missing {
				return local(ctx, job)
			}
			return result, primaryErr
		}
		if selected {
			return local(ctx, job)
		}
		result, err := primary(ctx, job)
		// A first rejection can establish a hold. Retry only within the original
		// budget and keep the failed primary run/receipt; never refund this call.
		if job.Attempt >= job.MaxAttempts || !fallbackProviderFailure(result, err) {
			return result, err
		}
		selected, lookupErr := s.ShouldFallback(ctx, binding, job, time.Now())
		if lookupErr != nil {
			return jobs.Result{}, fallbackRoutingFailure()
		}
		if !selected {
			return result, err
		}
		return jobs.Result{}, &jobs.HandlerError{Failure: jobs.Failure{Code: "local_fallback_pending", Detail: "provider hold established; alternative selected on next attempt", Temporary: true}}
	}
}

func fallbackRoutingFailure() error {
	return &jobs.HandlerError{Failure: jobs.Failure{Code: "fallback_routing_unavailable", Detail: "cannot select provider configuration", Temporary: true}}
}

func fallbackProviderFailure(result jobs.Result, err error) bool {
	var failure *jobs.HandlerError
	if err != nil {
		if !errors.As(err, &failure) {
			return false
		}
		switch failure.Code {
		case "provider_rejected", "provider_temporary", "transport_error", "request_timeout":
			return true
		}
		return false
	}
	var output struct{ Status, ReasonCode string }
	// Runner results use snake_case; do not infer failure from queue success.
	var fields map[string]json.RawMessage
	if json.Unmarshal(result.Payload, &fields) != nil {
		return false
	}
	_ = json.Unmarshal(fields["status"], &output.Status)
	_ = json.Unmarshal(fields["reason_code"], &output.ReasonCode)
	return (output.Status == "uninterpreted" || output.Status == "unresolved") && (output.ReasonCode == "provider_error" || output.ReasonCode == "attempts_exhausted")
}
