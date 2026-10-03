//go:build integration

package interpretation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
)

func TestExtractionSchedulesEmbeddingOnlyForEnabledSources(t *testing.T) {
	pool := interpretationTestDB(t)
	ctx := context.Background()
	if err := jobs.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `CREATE TABLE registry_sources(id text PRIMARY KEY);
        INSERT INTO registry_sources VALUES('source-a'),('source-b');
        CREATE TABLE processing_runs(id bigint PRIMARY KEY,source_id text,stage text);
        INSERT INTO processing_runs VALUES(1,'source-a','extraction'),(2,'source-a','extraction'),(3,'source-b','extraction'),(4,'source-a','extraction');
        CREATE TABLE interpretation_reuse(run_id bigint,original_run_id bigint);`); err != nil {
		t.Fatal(err)
	}
	queue := jobs.New(pool)
	s := New(pool, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, false)
	s.EmbeddingEnabled = queue.EmbeddingAllowedForRun
	var projected []int64
	s.ProjectExtraction = func(_ context.Context, id int64, _ time.Time) (int, error) {
		projected = append(projected, id)
		return 1, nil
	}
	now := time.Now()
	extract := func(run int64, want string) {
		t.Helper()
		n, err := s.AfterExtraction(ctx, extraction.Result{RunID: run, Status: "extracted", Measures: []extraction.Measure{{Ordinal: 1}}}, "ordinary", now)
		if err != nil || n != 1 {
			t.Fatal(n, err)
		}
		var kind string
		if err = pool.QueryRow(ctx, `SELECT kind FROM processing_jobs WHERE (payload->>'extraction_run_id')::bigint=$1`, run).Scan(&kind); err != nil || kind != want {
			t.Fatal(kind, want, err)
		}
	}
	extract(1, "link_measure_update")
	if _, err := queue.SetEmbeddingEnabled(ctx, 0, true, "global", now); err != nil {
		t.Fatal(err)
	}
	if _, err := queue.SetSourceEmbeddingEnabled(ctx, "source-a", 0, true, "source", now); err != nil {
		t.Fatal(err)
	}
	extract(2, "embed_measure")
	extract(3, "link_measure_update")
	if _, err := queue.SetEmbeddingEnabled(ctx, 1, false, "global-off", now); err != nil {
		t.Fatal(err)
	}
	extract(4, "link_measure_update")
	if len(projected) != 4 {
		t.Fatal("ordinary projection callback omitted", projected)
	}
	if _, e := pool.Exec(ctx, `INSERT INTO processing_runs VALUES(5,'source-a','extraction'),(6,'source-a','extraction')`); e != nil {
		t.Fatal(e)
	}
	s.ProjectExtraction = func(context.Context, int64, time.Time) (int, error) {
		return 0, errors.New("synthetic projection error")
	}
	if _, e := s.AfterExtraction(ctx, extraction.Result{RunID: 5, Status: "extracted", Measures: []extraction.Measure{{Ordinal: 1}}}, "ordinary", now); e == nil {
		t.Fatal("projection failure swallowed")
	}
	var scheduled int
	if e := pool.QueryRow(ctx, `SELECT count(*) FROM processing_jobs WHERE payload->>'extraction_run_id'='5'`).Scan(&scheduled); e != nil || scheduled != 0 {
		t.Fatal("linking scheduled after projection failure", scheduled, e)
	}
	if _, e := s.AfterExtraction(ctx, extraction.Result{RunID: 6, Status: "extracted", Measures: []extraction.Measure{{Ordinal: 1}}}, "evaluation", now); e != nil {
		t.Fatal("evaluation called public projector", e)
	}
}
