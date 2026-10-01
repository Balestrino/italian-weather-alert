package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/workerdiag"
)

type Handler func(context.Context, Job) (Result, error)

// HandlerError carries only an operator-safe stable code/detail. Unknown errors
// are persisted as internal_error without copying potentially secret text.
type HandlerError struct{ Failure }

func (e *HandlerError) Error() string { return e.Code }

type Worker struct {
	Store        *Store
	Queue        string
	ID           string
	Lease        time.Duration
	PollInterval time.Duration
	Handlers     map[string]Handler
	BeforeClaim  func(context.Context, time.Time) error
	now          func() time.Time
}

func (w *Worker) validate() error {
	if w.Store == nil || !validName(w.Queue) || !validName(w.ID) || w.Lease < 3*time.Second || w.Lease > time.Hour || w.PollInterval < 10*time.Millisecond || w.PollInterval > time.Minute || len(w.Handlers) == 0 {
		return ErrInvalid
	}
	if w.now == nil {
		w.now = time.Now
	}
	return nil
}

// Run processes jobs until cancellation. Multiple processes can run the same
// queue; PostgreSQL SKIP LOCKED ensures that each active lease is exclusive.
func (w *Worker) Run(ctx context.Context) error {
	if err := w.validate(); err != nil {
		return err
	}
	failures := 0
	for {
		if ctx.Err() != nil {
			return nil
		}
		worked, err := w.RunOne(ctx)
		if err != nil && !errors.Is(err, ErrNoJob) {
			if ctx.Err() != nil {
				return nil
			}
			failures++
			if failures > 3 || !(errors.Is(err, ErrStaleClaim) || workerdiag.Retryable(err)) {
				return err
			}
			op, category, state := workerdiag.Describe(err)
			if errors.Is(err, ErrStaleClaim) {
				category = "lease_lost"
			}
			slog.Warn("worker recovering", "component", "queue", "operation", op, "category", category, "sqlstate", state, "consecutive_failures", failures)
			if !workerdiag.Backoff(ctx, failures) {
				return nil
			}
			continue
		}
		failures = 0
		if worked {
			continue
		}
		timer := time.NewTimer(w.PollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}

func (w *Worker) RunOne(ctx context.Context) (bool, error) {
	if err := w.validate(); err != nil {
		return false, err
	}
	if w.BeforeClaim != nil {
		if err := w.BeforeClaim(ctx, w.now()); err != nil {
			return false, workerdiag.Wrap("before_claim", err)
		}
	}
	claim, err := w.Store.Claim(ctx, w.Queue, w.ID, w.now(), w.Lease)
	if errors.Is(err, ErrNoJob) {
		return false, nil
	}
	if err != nil {
		return false, workerdiag.Wrap("claim", err)
	}
	handler := w.Handlers[claim.Kind]
	if handler == nil {
		return true, w.Store.Fail(ctx, claim, Failure{Code: "unsupported_job_kind", Detail: "no handler is registered", Temporary: false}, w.now())
	}
	hbCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	var heartbeatErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(w.Lease / 3)
		defer ticker.Stop()
		for {
			select {
			case <-hbCtx.Done():
				return
			case <-ticker.C:
				beatCtx, beatCancel := context.WithTimeout(hbCtx, min(w.Lease/3, 10*time.Second))
				err := w.Store.Heartbeat(beatCtx, claim, w.now(), w.Lease)
				beatCancel()
				if err != nil {
					if hbCtx.Err() != nil {
						return
					}
					heartbeatErr = err
					cancel()
					return
				}
			}
		}
	}()
	result, handlerErr := safeHandle(hbCtx, handler, claim.Job)
	cancel()
	wg.Wait()
	if heartbeatErr != nil {
		return true, workerdiag.Wrap("heartbeat", heartbeatErr)
	}
	if ctx.Err() != nil {
		// Leave the lease running. Another process will persist the abandoned
		// attempt and retry after expiry.
		return true, nil
	}
	if handlerErr == nil {
		return true, workerdiag.Wrap("complete", w.Store.Complete(ctx, claim, result, w.now()))
	}
	var deferred *DeferredError
	if errors.As(handlerErr, &deferred) {
		return true, workerdiag.Wrap("defer", w.Store.Defer(ctx, claim, *deferred, w.now()))
	}
	failure := Failure{Code: "internal_error", Detail: "handler failed", Temporary: true}
	var typed *HandlerError
	if errors.As(handlerErr, &typed) {
		failure = typed.Failure
	}
	return true, workerdiag.Wrap("fail", w.Store.Fail(ctx, claim, failure, w.now()))
}

func safeHandle(ctx context.Context, handler Handler, job Job) (result Result, err error) {
	defer func() {
		if recover() != nil {
			err = &HandlerError{Failure{Code: "handler_panic", Detail: "handler panicked", Temporary: true}}
		}
	}()
	return handler(ctx, job)
}

// JSONResult builds the object-shaped result required by the durable store.
func JSONResult(value any, effects ...Effect) (Result, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return Result{}, fmt.Errorf("encode job result: %w", err)
	}
	return Result{Payload: b, Effects: effects}, nil
}
