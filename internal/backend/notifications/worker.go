package notifications

import (
	"context"
	"log/slog"
	"time"
)

type Worker struct {
	Store  *Store
	Sender Sender
	Config Config
}

func (w *Worker) Run(ctx context.Context) error {
	if w.Store == nil || w.Sender == nil || !w.Config.Enabled || w.Config.Validate() != nil {
		return ErrConfig
	}
	ticker := time.NewTicker(time.Duration(w.Config.PollSeconds) * time.Second)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return nil
		}
		cycleCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := w.Store.Reconcile(cycleCtx, time.Now(), w.Config.reminder())
		cancel()
		if err != nil && ctx.Err() == nil {
			slog.Error("notification incident reconciliation failed")
		}
		// Delivery remains independent from detector availability. Failed SMTP
		// attempts persist a bounded backoff and do not stop collection workers.
		for i := 0; i < 100 && ctx.Err() == nil; i++ {
			attemptCtx, stop := context.WithTimeout(ctx, time.Duration(w.Config.TimeoutSeconds+5)*time.Second)
			worked, deliveryErr := w.Store.DeliverOne(attemptCtx, w.Sender, time.Now())
			stop()
			if deliveryErr != nil {
				if ctx.Err() == nil {
					slog.Error("notification outbox unavailable")
				}
				break
			}
			if !worked {
				break
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
