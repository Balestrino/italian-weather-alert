//go:build integration

package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/workerdiag"
)

func TestWorkerDatabaseFailureAttribution(t *testing.T) {
	pool, _ := queueTestDB(t)
	ctx := context.Background()
	// Missing queue relation simulates an incompatible schema at the real claim
	// boundary. This must remain fatal and expose SQLSTATE, never SQL text.
	w := Worker{Store: New(pool), Queue: "test", ID: "diagnostic", Lease: 3 * time.Second, PollInterval: time.Second, Handlers: map[string]Handler{"test": func(context.Context, Job) (Result, error) { t.Fatal("unexpected handler"); return Result{}, nil }}}
	err := w.Run(ctx)
	op, category, state := workerdiag.Describe(err)
	if op != "claim" || category != "database" || state != "42P01" {
		t.Fatalf("incorrect failure attribution: %s %s %s", op, category, state)
	}
}
