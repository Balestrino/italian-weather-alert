package evaluation

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"
)

var ErrComparisonInvalid = errors.New("invalid processing comparison")

// Outcome is the human-reviewed result for one variant on one retained case.
// RunIDs identify the processing evidence from which timing, usage and catalog
// identities are resolved; callers cannot supply those measurements directly.
type Outcome struct {
	Status   string  `json:"status"`
	Actual   string  `json:"actual"`
	Evidence string  `json:"evidence"`
	RunIDs   []int64 `json:"run_ids"`
}

type ComparisonCase struct {
	ID        string  `json:"id"`
	Kind      string  `json:"kind"`
	Expected  string  `json:"expected"`
	Evidence  string  `json:"evidence"`
	Baseline  Outcome `json:"baseline"`
	Candidate Outcome `json:"candidate"`
}

type ProcessingComparison struct {
	ID           string           `json:"id"`
	CorpusSHA256 string           `json:"corpus_sha256"`
	Reviewer     string           `json:"reviewer"`
	ReviewedAt   string           `json:"reviewed_at"`
	StartedAt    time.Time        `json:"started_at"`
	FinishedAt   time.Time        `json:"finished_at"`
	Cases        []ComparisonCase `json:"cases"`
}

type ConfigurationIdentity struct {
	ID              string  `json:"id"`
	Stage           string  `json:"stage"`
	Revision        string  `json:"revision"`
	LogicVersion    string  `json:"logic_version"`
	ContentSHA256   string  `json:"content_sha256"`
	ModelVersionID  *string `json:"model_version_id"`
	PromptVersionID *string `json:"prompt_version_id"`
}

type Measurement struct {
	CaseID                  string `json:"case_id"`
	Variant                 string `json:"variant"`
	Stage                   string `json:"stage"`
	ConfigurationVersionID  string `json:"configuration_version_id"`
	RunID                   int64  `json:"run_id"`
	Attempts                int64  `json:"attempts"`
	DurationMS              int64  `json:"duration_ms"`
	UnknownUsageAttempts    int64  `json:"unknown_usage_attempts"`
	UnknownCostAttempts     int64  `json:"unknown_cost_attempts"`
	InputTokens             *int64 `json:"input_tokens"`
	OutputTokens            *int64 `json:"output_tokens"`
	CacheReadTokens         *int64 `json:"cache_read_tokens"`
	CacheWriteTokens        *int64 `json:"cache_write_tokens"`
	EstimatedCostMicrounits *int64 `json:"estimated_cost_microunits"`
	Currency                string `json:"currency"`
}

type VariantReport struct {
	ProcessVersion string                  `json:"process_version"`
	Configurations []ConfigurationIdentity `json:"configurations"`
	Measurements   []Measurement           `json:"measurements"`
	PassedCases    int                     `json:"passed_cases"`
	FailedCases    int                     `json:"failed_cases"`
}

type ComparisonReport struct {
	Comparison ProcessingComparison `json:"comparison"`
	Baseline   VariantReport        `json:"baseline"`
	Candidate  VariantReport        `json:"candidate"`
	Passed     bool                 `json:"passed"`
	RecordedBy string               `json:"recorded_by,omitempty"`
	RecordedAt time.Time            `json:"recorded_at,omitempty"`
}

func (c ProcessingComparison) Validate() error {
	for _, value := range []string{c.ID, c.Reviewer, c.ReviewedAt} {
		if strings.TrimSpace(value) == "" || value != strings.TrimSpace(value) {
			return ErrComparisonInvalid
		}
	}
	digest, err := hex.DecodeString(c.CorpusSHA256)
	if err != nil || len(digest) != sha256.Size || c.StartedAt.IsZero() || c.FinishedAt.Before(c.StartedAt) || len(c.Cases) < 2 || len(c.Cases) > 1000 {
		return ErrComparisonInvalid
	}
	reviewed, err := time.Parse("2006-01-02", c.ReviewedAt)
	if err != nil || reviewed.After(c.FinishedAt) {
		return ErrComparisonInvalid
	}
	seen := map[string]bool{}
	for _, item := range c.Cases {
		if item.ID == "" || item.Kind == "" || item.Expected == "" || item.Evidence == "" || seen[item.ID] {
			return ErrComparisonInvalid
		}
		seen[item.ID] = true
		if (item.Kind != "observed" && item.Kind != "synthetic" && item.Kind != "simulation") || validateOutcome(item.Baseline) != nil || validateOutcome(item.Candidate) != nil {
			return ErrComparisonInvalid
		}
	}
	return nil
}

func validateOutcome(value Outcome) error {
	if value.Status != "pass" && value.Status != "omission" && value.Status != "unsupported_assertion" && value.Status != "indeterminate" && value.Status != "not_run" {
		return ErrComparisonInvalid
	}
	if strings.TrimSpace(value.Actual) == "" || strings.TrimSpace(value.Evidence) == "" || len(value.RunIDs) == 0 || len(value.RunIDs) > 100 {
		return ErrComparisonInvalid
	}
	seen := map[int64]bool{}
	for _, id := range value.RunIDs {
		if id < 1 || seen[id] {
			return ErrComparisonInvalid
		}
		seen[id] = true
	}
	return nil
}

func processVersion(configurations []ConfigurationIdentity) string {
	values := append([]ConfigurationIdentity(nil), configurations...)
	sort.Slice(values, func(i, j int) bool { return values[i].ID < values[j].ID })
	h := sha256.New()
	for _, value := range values {
		for _, part := range []string{value.ID, value.Stage, value.Revision, value.LogicVersion, value.ContentSHA256, pointerValue(value.ModelVersionID), pointerValue(value.PromptVersionID)} {
			h.Write([]byte(part))
			h.Write([]byte{0})
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func pointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
