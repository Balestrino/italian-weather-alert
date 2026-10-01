package processing

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
)

func TestFallbackOnlyFollowsProviderFailures(t *testing.T) {
	for _, test := range []struct {
		name, payload, code string
		expected            bool
	}{
		{"quota transport", "", "provider_rejected", true},
		{"temporary transport", "", "provider_temporary", true},
		{"failed extraction", `{"status":"uninterpreted","reason_code":"provider_error"}`, "", true},
		{"successful extraction", `{"status":"extracted"}`, "", false},
		{"invalid evidence", `{"status":"uninterpreted","reason_code":"extraction_evidence_invalid"}`, "", false},
		{"classification validation", "", "classification_output_quotation", false},
		{"storage failure", "", "extraction_diagnostic_unavailable", false},
		{"gate storage failure", "", "provider_gate_unavailable", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var err error
			if test.code != "" {
				err = &jobs.HandlerError{Failure: jobs.Failure{Code: test.code}}
			}
			if got := fallbackProviderFailure(jobs.Result{Payload: json.RawMessage(test.payload)}, err); got != test.expected {
				t.Fatal("failure class changed", got)
			}
		})
	}
	if fallbackProviderFailure(jobs.Result{}, errors.New("unexpected")) {
		t.Fatal("unknown failure rerouted")
	}
}
