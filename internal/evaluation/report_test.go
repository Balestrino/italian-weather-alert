package evaluation

import (
	"strings"
	"testing"
	"time"
)

func reviewedRun() Run {
	now := time.Now().UTC()
	r := Run{ID: "run", SourceID: "calcinaia", Revision: 1, Suite: "municipal", CorpusSHA256: strings.Repeat("a", 64), Reviewer: "reviewer", ReviewedAt: "2026-09-16", ProcessVersion: "revision", Mode: "service", StartedAt: now, FinishedAt: now}
	for _, stage := range []string{"discovery", "ocr", "classification", "extraction", "linking"} {
		r.Checks = append(r.Checks, Check{CaseID: "case", Kind: "simulation", Stage: stage, MeasureField: stage == "extraction", Field: "result", Expected: "supported result", Outcome: "pass", Actual: "supported result", Evidence: "retained/report#case"})
	}
	return r
}
func TestReportsSeparateFailuresAndExpectedUncertainty(t *testing.T) {
	r := reviewedRun()
	r.Checks[0].Outcome = "omission"
	r.Checks[1].Outcome = "unsupported_assertion"
	r.Checks[2].Outcome = "indeterminate"
	r.Checks[3].Outcome = "indeterminate"
	r.Checks[3].ExpectedUnknown = true
	r.Checks = append(r.Checks, Check{CaseID: "case", Kind: "simulation", Stage: "extraction", MeasureField: true, Field: "place", Expected: "via Maremmana", Outcome: "pass", Actual: "via Maremmana", Evidence: "retained/report#case"})
	got, err := r.Reports()
	if err != nil || got.Passed || len(got.Omissions) != 1 || len(got.UnsupportedAssertions) != 1 || len(got.IndeterminateFields) != 2 || got.UsefulFields != 1 || got.ExpectedUsefulFields != 1 {
		t.Fatalf("%+v %v", got, err)
	}
	r = reviewedRun()
	r.Checks[3].Outcome = "indeterminate"
	r.Checks[3].ExpectedUnknown = true
	r.Checks = append(r.Checks, Check{CaseID: "case", Kind: "simulation", Stage: "extraction", MeasureField: true, Field: "place", Expected: "via Maremmana", Outcome: "pass", Actual: "via Maremmana", Evidence: "retained/report#case"})
	got, err = r.Reports()
	if err != nil || !got.Passed {
		t.Fatalf("documented unknown should pass: %+v %v", got, err)
	}
}
func TestEmptyExtractorAndContractTestsCannotPass(t *testing.T) {
	controlOnly := reviewedRun()
	controlOnly.Checks[3].MeasureField = false
	controlReport, err := controlOnly.Reports()
	if err != nil || controlReport.Passed || controlReport.UsefulFields != 0 {
		t.Fatal("passing visibility controls substituted for useful extracted fields")
	}
	for _, outcome := range []string{"indeterminate", "not_run"} {
		r := reviewedRun()
		for i := range r.Checks {
			r.Checks[i].Outcome = outcome
		}
		got, err := r.Reports()
		if err != nil || got.Passed || got.UsefulFields != 0 {
			t.Fatalf("%+v %v", got, err)
		}
	}
	r := reviewedRun()
	r.Mode = "contract_test"
	got, err := r.Reports()
	if err != nil || got.Passed {
		t.Fatal("mock passed service gate")
	}
}
func TestReportRejectsMissingStagesDuplicatesAndInvalidReview(t *testing.T) {
	for _, change := range []func(*Run){func(r *Run) { r.Checks = r.Checks[:4] }, func(r *Run) { r.Checks = append(r.Checks, r.Checks[0]) }, func(r *Run) { r.Reviewer = "" }, func(r *Run) { r.CorpusSHA256 = "bad" }, func(r *Run) { r.ReviewedAt = "2099-01-01" }, func(r *Run) { r.Checks[0].Outcome = "" }, func(r *Run) { r.Checks[0].Evidence = "" }, func(r *Run) { r.Checks[0].ExpectedUnknown = true }} {
		r := reviewedRun()
		change(&r)
		if r.Validate() == nil {
			t.Fatal("invalid report accepted")
		}
	}
}
func TestRerunPinsExpectationsButAllowsReorderedChecks(t *testing.T) {
	r := reviewedRun()
	hash := r.ContractHash()
	r.Checks[0], r.Checks[1] = r.Checks[1], r.Checks[0]
	r.Checks[0].Outcome = "omission"
	r.Checks[0].Actual = "missing"
	r.ID = "rerun"
	if r.ContractHash() != hash {
		t.Fatal("result changed expected contract")
	}
	r.Checks[0].Expected = "weaker criterion"
	if r.ContractHash() == hash {
		t.Fatal("weakened expectation retained hash")
	}
}
