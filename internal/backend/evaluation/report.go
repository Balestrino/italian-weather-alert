// Package evaluation records comparisons with reviewed expectations. It does
// not use a model's self-assessment as a correctness verdict.
package evaluation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid regression report")

// Check is one comparison, not a count supplied by an acceptance form.
// ExpectedUnknown is part of the reviewed expectation, never inferred from an
// empty model output. Evidence identifies the retained actual result/comparison.
type Check struct {
	CaseID          string `json:"case_id"`
	Kind            string `json:"kind"`
	Stage           string `json:"stage"`
	Field           string `json:"field"`
	Expected        string `json:"expected"`
	ExpectedUnknown bool   `json:"expected_unknown"`
	MeasureField    bool   `json:"measure_field"`
	Outcome         string `json:"outcome"`
	Actual          string `json:"actual"`
	Evidence        string `json:"evidence"`
}

type Run struct {
	ID             string    `json:"id"`
	SourceID       string    `json:"source_id"`
	Revision       int       `json:"revision"`
	Suite          string    `json:"suite"`
	CorpusSHA256   string    `json:"corpus_sha256"`
	Reviewer       string    `json:"reviewer"`
	ReviewedAt     string    `json:"reviewed_at"`
	ProcessVersion string    `json:"process_version"`
	Mode           string    `json:"mode"` // service or contract_test
	StartedAt      time.Time `json:"started_at"`
	FinishedAt     time.Time `json:"finished_at"`
	Correction     string    `json:"correction"`
	Checks         []Check   `json:"checks"`
}

type Reports struct {
	Omissions             []Check `json:"omissions"`
	UnsupportedAssertions []Check `json:"unsupported_assertions"`
	IndeterminateFields   []Check `json:"indeterminate_fields"`
	NotRun                []Check `json:"not_run"`
	UsefulFields          int     `json:"useful_fields"`
	ExpectedUsefulFields  int     `json:"expected_useful_fields"`
	Passed                bool    `json:"passed"`
}

func (r Run) Validate() error {
	for _, value := range []string{r.ID, r.SourceID, r.Suite, r.Reviewer, r.ProcessVersion} {
		if strings.TrimSpace(value) == "" {
			return ErrInvalid
		}
	}
	digest, err := hex.DecodeString(r.CorpusSHA256)
	if err != nil || len(digest) != sha256.Size {
		return ErrInvalid
	}
	reviewed, err := time.Parse("2006-01-02", r.ReviewedAt)
	if err != nil || reviewed.After(r.StartedAt) || r.Revision < 1 || r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) || len(r.Checks) == 0 || len(r.Checks) > 10000 || (r.Mode != "service" && r.Mode != "contract_test") {
		return ErrInvalid
	}
	stages, keys := map[string]bool{}, map[string]bool{}
	for _, c := range r.Checks {
		if c.Stage != "discovery" && c.Stage != "ocr" && c.Stage != "classification" && c.Stage != "extraction" && c.Stage != "linking" {
			return ErrInvalid
		}
		if c.MeasureField && c.Stage != "extraction" {
			return ErrInvalid
		}
		if c.Kind != "observed" && c.Kind != "synthetic" && c.Kind != "simulation" && c.Kind != "normative_example" {
			return ErrInvalid
		}
		if c.CaseID == "" || c.Field == "" || strings.TrimSpace(c.Expected) == "" || strings.TrimSpace(c.Actual) == "" || strings.TrimSpace(c.Evidence) == "" {
			return ErrInvalid
		}
		if c.Outcome != "pass" && c.Outcome != "omission" && c.Outcome != "unsupported_assertion" && c.Outcome != "indeterminate" && c.Outcome != "not_run" {
			return ErrInvalid
		}
		key := checkKey(c)
		if keys[key] || (c.ExpectedUnknown && c.Outcome == "pass") {
			return ErrInvalid
		}
		keys[key], stages[c.Stage] = true, true
	}
	// Missing stages must be represented explicitly by not_run checks.
	if len(stages) != 5 {
		return ErrInvalid
	}
	return nil
}

func (r Run) Reports() (Reports, error) {
	if err := r.Validate(); err != nil {
		return Reports{}, err
	}
	result := Reports{Omissions: []Check{}, UnsupportedAssertions: []Check{}, IndeterminateFields: []Check{}, NotRun: []Check{}, Passed: r.Mode == "service"}
	for _, c := range r.Checks {
		if c.MeasureField && !c.ExpectedUnknown {
			result.ExpectedUsefulFields++
		}
		switch c.Outcome {
		case "pass":
			if c.MeasureField {
				result.UsefulFields++
			}
		case "omission":
			result.Omissions = append(result.Omissions, c)
			result.Passed = false
		case "unsupported_assertion":
			result.UnsupportedAssertions = append(result.UnsupportedAssertions, c)
			result.Passed = false
		case "not_run":
			result.NotRun = append(result.NotRun, c)
			result.Passed = false
		case "indeterminate":
			result.IndeterminateFields = append(result.IndeterminateFields, c)
			if !c.ExpectedUnknown {
				result.Passed = false
			}
		}
	}
	if result.UsefulFields == 0 {
		result.Passed = false
	}
	return result, nil
}

func checkKey(c Check) string {
	b, _ := json.Marshal([]string{c.CaseID, c.Stage, c.Field})
	return string(b)
}

// ContractHash pins both the reviewed corpus and every expected comparison.
// A rerun cannot clear a failure by dropping cases or weakening expectations.
func (r Run) ContractHash() string {
	checks := append([]Check(nil), r.Checks...)
	for i := range checks {
		checks[i].Outcome, checks[i].Actual, checks[i].Evidence = "", "", ""
	}
	sort.Slice(checks, func(i, j int) bool { return checkKey(checks[i]) < checkKey(checks[j]) })
	b, _ := json.Marshal(struct {
		Corpus string
		Checks []Check
	}{r.CorpusSHA256, checks})
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
