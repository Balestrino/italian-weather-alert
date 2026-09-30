package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

type DiagnosticCase struct {
	Class   string
	Adapter Adapter
	Request Request
}
type DiagnosticResult struct {
	Class       string     `json:"class"`
	InputSHA256 string     `json:"input_sha256"`
	ObservedAt  time.Time  `json:"observed_at"`
	ElapsedMS   int64      `json:"elapsed_ms"`
	Diagnostic  Diagnostic `json:"diagnostic"`
	Usage       Usage      `json:"usage"`
	ErrorCode   string     `json:"error_code,omitempty"`
	NextAction  string     `json:"next_action"`
}

// Diagnose makes at most three one-shot calls, one per explicit input class.
// It returns only hashes and approved diagnostics, never prompts or output text.
func Diagnose(ctx context.Context, cases []DiagnosticCase) ([]DiagnosticResult, error) {
	if len(cases) < 1 || len(cases) > 3 {
		return nil, ErrInvalid
	}
	seen := map[string]bool{}
	for _, c := range cases {
		if c.Adapter == nil || !opaqueID.MatchString(c.Class) || seen[c.Class] || ValidateRequest(c.Request) != nil || c.Request.MaxTokens > 4096 || c.Request.MaxCompletionTokens > 4096 || c.Request.MaxTokens+c.Request.MaxCompletionTokens == 0 {
			return nil, ErrInvalid
		}
		seen[c.Class] = true
	}
	results := make([]DiagnosticResult, 0, len(cases))
	for _, c := range cases {
		if ctx.Err() != nil {
			return results, ctx.Err()
		}
		wire, _ := json.Marshal(c.Request)
		hash := sha256.Sum256(wire)
		start := time.Now().UTC()
		callCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
		response, err := c.Adapter.Complete(callCtx, c.Request)
		cancel()
		r := DiagnosticResult{Class: c.Class, InputSHA256: hex.EncodeToString(hash[:]), ObservedAt: start, ElapsedMS: time.Since(start).Milliseconds(), Diagnostic: response.Diagnostic, Usage: response.Usage, NextAction: "validate_output_against_retained_evidence"}
		if err != nil {
			r.ErrorCode = "provider_error"
			var call *CallError
			if errors.As(err, &call) {
				r.ErrorCode = receiptID(call.Code)
				r.Diagnostic = call.Diagnostic
				r.Usage = call.Usage
			}
			switch r.Diagnostic.Category {
			case "authentication":
				r.NextAction = "check_credential_configuration_and_hold_scope"
			case "quota":
				r.NextAction = "operator_review_account_quota_and_hold_scope"
			case "permission":
				r.NextAction = "check_model_access_before_retry"
			case "request":
				r.NextAction = "compare_request_contract_with_failing_input"
			case "rate_limit", "availability":
				r.NextAction = "respect_retry_after_and_bounded_cooldown"
			default:
				r.NextAction = "inspect_safe_diagnostics_without_automatic_retry"
			}
		}
		results = append(results, r)
		if r.Diagnostic.Category == "authentication" || r.Diagnostic.Category == "quota" {
			break
		}
	}
	return results, nil
}
