package acquisition

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/workerdiag"
)

type CheckWorker struct {
	Store        *ScheduleStore
	Engine       *Engine
	ID           string
	Lease        time.Duration
	PollInterval time.Duration
	Schedule     func(context.Context, RetainedPage, time.Time) error
	now          func() time.Time
}

func (w *CheckWorker) validate() error {
	if w.Store == nil || w.Engine == nil || !validWorker(w.ID) || w.Lease < 3*time.Second || w.Lease > time.Hour || w.PollInterval < 10*time.Millisecond || w.PollInterval > time.Minute {
		return ErrInvalidConfiguration
	}
	if w.now == nil {
		w.now = time.Now
	}
	return nil
}

func (w *CheckWorker) Run(ctx context.Context) error {
	if err := w.validate(); err != nil {
		return err
	}
	failures := 0
	for {
		if ctx.Err() != nil {
			return nil
		}
		worked, err := w.RunOne(ctx)
		if err != nil && !errors.Is(err, ErrNoDueCheck) {
			if ctx.Err() != nil {
				return nil
			}
			failures++
			if failures > 3 || !(errors.Is(err, ErrStaleCheck) || workerdiag.Retryable(err)) {
				return err
			}
			op, category, state := workerdiag.Describe(err)
			if errors.Is(err, ErrStaleCheck) {
				category = "lease_lost"
			}
			slog.Warn("worker recovering", "component", "acquisition", "operation", op, "category", category, "sqlstate", state, "consecutive_failures", failures)
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

func (w *CheckWorker) RunOne(ctx context.Context) (bool, error) {
	if err := w.validate(); err != nil {
		return false, err
	}
	now := w.now().UTC()
	if err := w.Store.SyncEnabled(ctx, now); err != nil {
		return false, workerdiag.Wrap("sync_enabled", err)
	}
	claim, err := w.Store.ClaimDue(ctx, w.ID, now, w.Lease)
	if errors.Is(err, ErrNoDueCheck) {
		return false, nil
	}
	if err != nil {
		return false, workerdiag.Wrap("claim_due", err)
	}
	hbCtx, cancel := context.WithCancel(ctx)
	heartbeat := make(chan error, 1)
	go func() {
		ticker := time.NewTicker(w.Lease / 3)
		defer ticker.Stop()
		for {
			select {
			case <-hbCtx.Done():
				heartbeat <- nil
				return
			case <-ticker.C:
				beatCtx, beatCancel := context.WithTimeout(hbCtx, min(w.Lease/3, 10*time.Second))
				err := w.Store.Heartbeat(beatCtx, claim, w.now(), w.Lease)
				beatCancel()
				if err != nil {
					if hbCtx.Err() != nil {
						heartbeat <- nil
						return
					}
					cancel()
					heartbeat <- err
					return
				}
			}
		}
	}()
	outcome := w.Engine.Check(hbCtx, claim.SourceID, claim.Configuration, claim.StartedAt)
	if hbCtx.Err() == nil && w.Schedule != nil {
		for _, retained := range outcome.Retained {
			if err = w.Schedule(hbCtx, retained, outcome.FinishedAt); err != nil {
				break
			}
		}
	}
	cancel()
	heartbeatErr := <-heartbeat
	if heartbeatErr != nil {
		return true, workerdiag.Wrap("heartbeat", heartbeatErr)
	}
	if ctx.Err() != nil {
		return true, nil
	}
	if err != nil {
		return true, workerdiag.Wrap("schedule", err)
	}
	return true, workerdiag.Wrap("finish", w.Store.Finish(ctx, claim, outcome))
}
