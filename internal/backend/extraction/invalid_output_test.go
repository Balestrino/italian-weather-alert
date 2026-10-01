package extraction

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
)

func TestRejectedExtractionPreservesDiagnosticEvidence(t *testing.T) {
	for _, raw := range []string{`invalid` + "\x00" + `response`, `{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":[],"measures":[{"kind":"closure","subject":"invented","evidence_refs":{"kind":[9]}}]}`} {
		t.Run(raw, func(t *testing.T) {
			runner, results, adapter, process := fixtureRunner(t, true, raw)
			adapter.response.FinishReason = "stop"
			if _, err := runJob(t, runner); err != nil {
				t.Fatal(err)
			}
			if len(adapter.requests) != 1 || len(process.invalidOutputs) != 1 {
				t.Fatal("rejection retried or diagnostic lost")
			}
			capture := process.invalidOutputs[0]
			request, _ := json.Marshal(adapter.requests[0])
			if !bytes.Equal(capture.Request, request) || string(capture.Response) != raw || capture.RunID != 81 || capture.AttemptNumber != 1 || capture.SegmentOrdinal != 1 || capture.FinishReason != "stop" || capture.Cached || capture.CreatedAt != runner.Now() {
				t.Fatalf("incomplete capture: %#v", capture)
			}
			if results.value == nil || results.value.Status != "uninterpreted" || len(results.value.Measures) != 0 || process.finished[0].ErrorCode != capture.ErrorCode {
				t.Fatal("invalid output accepted or diagnostic differs")
			}
		})
	}
}

func TestExtractionCaptureFailureStopsWithoutPublishingOrRetry(t *testing.T) {
	runner, results, adapter, process := fixtureRunner(t, true, `invalid`)
	process.captureErr = errors.New("storage unavailable")
	_, err := runJob(t, runner)
	var failure *jobs.HandlerError
	if !errors.As(err, &failure) || failure.Code != "extraction_diagnostic_unavailable" || failure.Temporary {
		t.Fatalf("capture failed open: %v", err)
	}
	if len(adapter.requests) != 1 || results.value != nil || len(process.finished) != 1 || process.finished[0].ErrorCode != "extraction_diagnostic_unavailable" || process.finished[0].Usage.InputTokens == nil || *process.finished[0].Usage.InputTokens != 150 {
		t.Fatal("capture failure lost accounting or published a result")
	}
}
