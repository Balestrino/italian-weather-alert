package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"strings"
	"testing"
	"time"
)

type retryCheckpoints struct {
	values map[string]processing.SegmentCheckpoint
	reused int
}

func (c *retryCheckpoints) Checkpoint(_ context.Context, k processing.SegmentCheckpoint) (processing.SegmentCheckpoint, bool, error) {
	v, ok := c.values[k.Configuration+k.InputSHA256]
	return v, ok, nil
}
func (c *retryCheckpoints) PutCheckpoint(_ context.Context, k processing.SegmentCheckpoint, _ time.Time) error {
	c.values[k.Configuration+k.InputSHA256] = k
	return nil
}
func (c *retryCheckpoints) UseCheckpoint(_ context.Context, _ processing.SegmentCheckpoint, _ processing.Attempt, _ int, reused bool) error {
	if reused {
		c.reused++
	}
	return nil
}

type retryAdapter struct {
	calls   int
	respond func(inference.Request) string
}

func (a *retryAdapter) Name() string { return "fixture" }
func (a *retryAdapter) Complete(_ context.Context, r inference.Request) (inference.Response, error) {
	a.calls++
	if a.calls == 2 {
		return inference.Response{}, &inference.CallError{Code: "overloaded", Temporary: true}
	}
	n := int64(10)
	return inference.Response{CallID: int64(a.calls), Model: "qwen3.8-27b", ID: fmt.Sprint(a.calls), Content: a.respond(r), Usage: inference.Usage{InputTokens: &n, OutputTokens: &n}}, nil
}
func TestRetryUsesValidatedSegmentCheckpoints(t *testing.T) {
	runner, results, _, _ := fixtureRunner(t, true, `{}`)
	docs := runner.Documents.(*fakeDocuments)
	docs.bodies["https://comune.example/notizia"] = []byte("<main>" + strings.Repeat("Avviso comunale senza misure meteo. ", 2000) + "</main>")
	cache := &retryCheckpoints{values: map[string]processing.SegmentCheckpoint{}}
	a := &retryAdapter{respond: func(r inference.Request) string {
		var encoded string
		_ = json.Unmarshal(r.Messages[1].Content, &encoded)
		var e struct {
			Window struct {
				Ordinal int `json:"ordinal"`
			} `json:"window"`
		}
		_ = json.Unmarshal([]byte(encoded), &e)
		return fmt.Sprintf(`{"envelope_version":"compact-evidence-v1","window_ordinal":%d,"evidence":[],"measures":[]}`, e.Window.Ordinal)
	}}
	runner.Checkpoints = cache
	runner.CheckpointSources = map[string]bool{"calcinaia": true}
	runner.Adapter = a
	payload, _ := json.Marshal(Payload{DocumentVersionID: docs.version.ID, Workload: "evaluation", ClassificationRunID: 61})
	job := jobs.Job{ID: 70, Attempt: 1, MaxAttempts: 3, Payload: payload}
	if _, err := runner.Handler()(context.Background(), job); err == nil {
		t.Fatal("expected second-segment failure")
	}
	if a.calls != 2 || len(cache.values) != 1 {
		t.Fatalf("first segment not saved: %d %d", a.calls, len(cache.values))
	}
	job.Attempt = 2
	if _, err := runner.Handler()(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if results.value == nil || len(results.value.Segments) < 2 || a.calls != len(results.value.Segments)+1 || cache.reused != 1 {
		t.Fatalf("repeated or missing work: calls=%d reused=%d result=%#v", a.calls, cache.reused, results.value)
	}
	// Replaying all completed segments is free, but changed configuration misses.
	results.value = nil
	job.Attempt = 3
	before := a.calls
	if _, err := runner.Handler()(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if a.calls != before {
		t.Fatal("warm replay called provider")
	}
	results.value = nil
	runner.ConfigurationVersion += "-changed"
	job.Attempt = 4
	if _, err := runner.Handler()(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if a.calls <= before {
		t.Fatal("changed configuration reused old output")
	}
	results.value = nil
	before = a.calls
	for url, body := range docs.bodies {
		docs.bodies[url] = []byte(strings.ReplaceAll(string(body), "Avviso comunale", "Avviso comunale aggiornato"))
	}
	job.Attempt++
	if _, err := runner.Handler()(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if a.calls <= before {
		t.Fatal("changed input reused old output")
	}
}
