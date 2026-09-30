package embedding

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/linking"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"testing"
	"time"
)

type measuresFake struct{ value linking.MeasureContext }

func (f measuresFake) Current(context.Context, int64, int) (linking.MeasureContext, error) {
	return f.value, nil
}

type resultsFake struct{ value *Result }

func (f *resultsFake) Put(_ context.Context, r Result) error { f.value = &r; return nil }
func (f *resultsFake) Get(context.Context, int64, int, string) (Result, bool, error) {
	if f.value == nil {
		return Result{}, false, nil
	}
	return *f.value, true, nil
}

type processFake struct{ finished processing.AttemptFinish }

func (f *processFake) StartRun(context.Context, processing.RunRequest) (processing.Run, error) {
	return processing.Run{ID: 9}, nil
}
func (f *processFake) StartAttempt(context.Context, processing.AttemptStart) (processing.Attempt, error) {
	return processing.Attempt{RunID: 9, Number: 1}, nil
}
func (f *processFake) FinishAttempt(_ context.Context, x processing.AttemptFinish) (processing.Attempt, error) {
	f.finished = x
	return processing.Attempt{}, nil
}

type embedFake struct{}

func (embedFake) Name() string { return "fixture" }
func (embedFake) Embed(context.Context, string, []string, int) (inference.EmbeddingResponse, error) {
	tokens := int64(17)
	v := make([]float32, 1024)
	v[0] = 1
	return inference.EmbeddingResponse{Model: "Qwen3-Embedding-8B", Vectors: [][]float32{v}, InputTokens: &tokens}, nil
}
func TestEmbeddingRunRecordsIncrementalUsage(t *testing.T) {
	at := time.Date(2026, 9, 17, 18, 0, 0, 0, time.UTC)
	results, process := &resultsFake{}, &processFake{}
	runner := &Runner{Measures: measuresFake{linking.MeasureContext{RunID: 3, DocumentVersionID: 4, Ordinal: 1, SourceID: "s", Kind: "closure", Subject: "sottopasso", Place: "Maremmana"}}, Results: results, Processing: process, Adapter: embedFake{}, Model: "Qwen3-Embedding-8B", ConfigurationVersion: "embedding-config", Dimensions: 1024, Now: func() time.Time { return at }}
	payload := []byte(`{"extraction_run_id":3,"measure_ordinal":1,"workload":"evaluation"}`)
	if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 8, Attempt: 1, MaxAttempts: 3, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if results.value == nil || len(results.value.Vector) != 1024 || process.finished.Usage.InputTokens == nil || *process.finished.Usage.InputTokens != 17 || process.finished.Usage.Status != "reported" {
		t.Fatalf("embedding usage/vector missing: %#v %#v", results.value, process.finished)
	}
}
