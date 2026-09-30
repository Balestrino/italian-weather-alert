package notifications

import (
	"context"
	"log/slog"
	"time"
)

type DailyReportWorker struct {
	Store  *Store
	Reader ReportReader
	Sender ReportSender
	Config Config
}

func (w *DailyReportWorker) Generate(ctx context.Context, at time.Time) (ReportMail, error) {
	snapshot, err := w.Reader.Read(ctx, at)
	if err != nil {
		return ReportMail{}, err
	}
	return RenderReport(snapshot, w.Config.DailyReport.Timezone)
}
func (w *DailyReportWorker) Tick(ctx context.Context, at time.Time) error {
	generationCtx, cancel := context.WithTimeout(ctx, 40*time.Second)
	_, generationErr := w.Store.QueueDaily(generationCtx, *w.Config.DailyReport, at, w.Generate)
	cancel()
	// A failed generator cannot prevent delivery of previous snapshots.
	for i := 0; i < 10 && ctx.Err() == nil; i++ {
		deliveryCtx, stop := context.WithTimeout(ctx, time.Duration(w.Config.TimeoutSeconds+5)*time.Second)
		worked, err := w.Store.DeliverReport(deliveryCtx, w.Sender, at)
		stop()
		if err != nil {
			return err
		}
		if !worked {
			break
		}
	}
	return generationErr
}
func (w *DailyReportWorker) Run(ctx context.Context) error {
	if w.Store == nil || w.Reader.Pool == nil || w.Sender == nil || w.Config.Validate() != nil || w.Config.DailyReport == nil || !w.Config.DailyReport.Enabled {
		return ErrConfig
	}
	ticker := time.NewTicker(time.Duration(w.Config.PollSeconds) * time.Second)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return nil
		}
		if err := w.Tick(ctx, time.Now()); err != nil && ctx.Err() == nil {
			slog.Error("daily report cycle failed")
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
