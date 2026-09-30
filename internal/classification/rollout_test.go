package classification

import (
	"context"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"strings"
	"testing"
)

func TestOutputFixCanaryPreservesLegacyOutsideSelectedSource(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		url := "https://fixture.example/notice"
		docs := &fakeDocuments{version: documents.Version{ID: 8, Complete: true, Resources: []documents.Reference{{URL: url, Role: "original", Required: true, SourceID: "source-a", Configuration: 1, MediaType: "text/html", Hash: strings.Repeat("a", 64)}}}, bodies: map[string][]byte{url: []byte("<main>Chiusura per l'emergenza meteo.</main>")}}
		results, process := &memoryResults{}, &fakeProcessing{}
		adapter := &fakeAdapter{response: inference.Response{Content: `{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"l’emergenza meteo"}`}}
		fixes := map[string]bool{"source-b": true}
		if enabled {
			fixes["source-a"] = true
		}
		cache := &retryCheckpoints{values: map[string]processing.SegmentCheckpoint{}}
		runner := &Runner{Checkpoints: cache, CheckpointSources: map[string]bool{"other": true}, Documents: docs, OCR: fakeOCR{}, Processing: process, Results: results, Adapter: adapter, Model: "fixture", ConfigurationVersion: "new", LegacyConfigurationVersion: "legacy", OutputFixSources: fixes}
		payload, _ := json.Marshal(Payload{DocumentVersionID: 8, Workload: "evaluation"})
		_, err := runner.Handler()(context.Background(), jobs.Job{ID: 1, Attempt: 1, MaxAttempts: 3, Payload: payload})
		if enabled {
			if err != nil || results.value == nil || process.lastConfiguration != "new" {
				t.Fatalf("selected source did not use evaluated parser: %v", err)
			}
		} else {
			if err == nil || results.value != nil || process.lastConfiguration != "legacy" {
				t.Fatal("unselected source behavior changed")
			}
		}
		if len(cache.values) != 0 {
			t.Fatal("checkpoint written outside canary source")
		}
		if runner.legacyOutput || runner.ConfigurationVersion != "new" {
			t.Fatal("shared runner mutated")
		}
	}
}
