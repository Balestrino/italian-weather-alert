//go:build integration

package registry

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/evaluation"
)

func TestRegressionFailureBlocksOldAcceptanceUntilCorrectedRerun(t *testing.T) {
	ctx := context.Background()
	pool, restart := testDB(t)
	s := New(pool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	want := func(err, target error) {
		t.Helper()
		if !errors.Is(err, target) {
			t.Fatalf("got %v want %v", err, target)
		}
	}
	must(Migrate(ctx, pool))
	must(Migrate(ctx, pool))
	must(s.CreateAuthority(ctx, Authority{"municipality", "Comune", "https://comune.example"}))
	must(s.CreateChannel(ctx, Channel{"direct", "municipality", "official", "https://comune.example", false}))
	c := fixtureConfig()
	c.URL = "https://comune.example/notizie"
	c.Sections = []string{c.URL}
	for _, id := range []string{"calcinaia", "other"} {
		must(s.CreateSource(ctx, Source{id, "municipality", "direct", "municipal", "050004"}, c, "operator"))
		must(s.RecordPreview(ctx, id, 1, "operator", *c.Provenance))
		must(s.EnableCollection(ctx, id, 1, "operator"))
		must(s.Accept(ctx, id, 1, "reviewer", acceptance(c)))
		reviewAcceptance(t, s, ctx, id, 1)
	}
	run := func(id, outcome string) evaluation.Run {
		now := time.Now().UTC()
		r := evaluation.Run{ID: id, SourceID: "calcinaia", Revision: 1, Suite: "municipal", CorpusSHA256: strings.Repeat("a", 64), Reviewer: "test reviewer", ReviewedAt: "2026-09-16", ProcessVersion: "v1", Mode: "service", StartedAt: now, FinishedAt: now}
		for _, stage := range []string{"discovery", "ocr", "classification", "extraction", "linking"} {
			r.Checks = append(r.Checks, evaluation.Check{CaseID: "case", Kind: "simulation", Stage: stage, MeasureField: stage == "extraction", Field: "fact", Expected: "supported fact", Outcome: "pass", Actual: "supported fact", Evidence: "retained/test#case"})
		}
		r.Checks[0].Outcome = outcome
		return r
	}
	failed := run("failed", "omission")
	must(s.RecordRegression(ctx, "calcinaia", 1, "reviewer", "", failed))
	want(s.Accept(ctx, "calcinaia", 1, "reviewer", acceptance(c)), ErrPrerequisite)
	want(s.EnablePublic(ctx, "calcinaia", 1, "operator"), ErrPrerequisite)
	st, err := s.State(ctx, "calcinaia")
	must(err)
	if st.Accepted || st.PublicEnabled || !st.CollectionEnabled {
		t.Fatal("failure lost gate or collection")
	}
	must(s.EnablePublic(ctx, "other", 1, "operator"))
	restart()
	s = New(pool)
	must(Migrate(ctx, pool))
	want(s.EnablePublic(ctx, "calcinaia", 1, "operator"), ErrPrerequisite)
	corrected := run("corrected", "pass")
	corrected.Correction = "parser correction v2, rerun retained case"
	corrected.ProcessVersion = "v2"
	want(s.RecordRegression(ctx, "calcinaia", 1, "reviewer", "", corrected), ErrConflict)
	weakened := corrected
	weakened.Checks = append([]evaluation.Check(nil), corrected.Checks...)
	weakened.Checks[0].Expected = "less coverage"
	want(s.RecordRegression(ctx, "calcinaia", 1, "reviewer", "failed", weakened), ErrPrerequisite)
	old := corrected
	old.StartedAt = failed.StartedAt
	want(s.RecordRegression(ctx, "calcinaia", 1, "reviewer", "failed", old), ErrPrerequisite)
	noCorrection := corrected
	noCorrection.Correction = ""
	want(s.RecordRegression(ctx, "calcinaia", 1, "reviewer", "failed", noCorrection), ErrPrerequisite)
	must(s.RecordRegression(ctx, "calcinaia", 1, "reviewer", "failed", corrected))
	want(s.EnablePublic(ctx, "calcinaia", 1, "operator"), ErrPrerequisite)
	must(s.Accept(ctx, "calcinaia", 1, "reviewer", acceptance(c)))
	reviewAcceptance(t, s, ctx, "calcinaia", 1)
	must(s.EnablePublic(ctx, "calcinaia", 1, "operator"))
	// A separate suite cannot clear an outstanding failure in another suite.
	failedAgain := run("unsupported", "unsupported_assertion")
	must(s.RecordRegression(ctx, "calcinaia", 1, "reviewer", "corrected", failedAgain))
	otherSuite := run("unrelated-pass", "pass")
	otherSuite.Suite = "other-suite"
	must(s.RecordRegression(ctx, "calcinaia", 1, "reviewer", "", otherSuite))
	want(s.Accept(ctx, "calcinaia", 1, "reviewer", acceptance(c)), ErrPrerequisite)
	want(s.EnablePublic(ctx, "calcinaia", 1, "operator"), ErrPrerequisite)
	// Existing publication is not erased by a regression; suspension is separate.
	st, err = s.State(ctx, "calcinaia")
	must(err)
	if !st.PublicEnabled || st.Accepted {
		t.Fatal("existing publication erased or acceptance stale")
	}
	if _, err = pool.Exec(ctx, "UPDATE registry_regressions SET passed=true"); err == nil {
		t.Fatal("regressions mutable")
	}
	if _, err = pool.Exec(ctx, "DELETE FROM registry_regressions"); err == nil {
		t.Fatal("regressions deletable")
	}
	reports, err := s.Regressions(ctx, "calcinaia")
	must(err)
	if len(reports) != 4 || len(reports[0].Reports.Omissions) != 1 || len(reports[2].Reports.UnsupportedAssertions) != 1 {
		t.Fatalf("lost history: %+v", reports)
	}
	// A draft regression cannot invalidate the active revision's history.
	rev, err := s.AppendConfiguration(ctx, "other", 1, c, "operator")
	must(err)
	draft := run("draft-failure", "omission")
	draft.SourceID = "other"
	draft.Revision = rev
	must(s.RecordRegression(ctx, "other", rev, "reviewer", "", draft))
	must(s.EnablePublic(ctx, "other", 1, "operator"))
	// Renumbering a failed configuration cannot discard its evaluation suite.
	next, err := s.AppendConfiguration(ctx, "calcinaia", 1, c, "operator")
	must(err)
	must(s.RecordPreview(ctx, "calcinaia", next, "operator", *c.Provenance))
	must(s.EnableCollection(ctx, "calcinaia", next, "operator"))
	want(s.Accept(ctx, "calcinaia", next, "reviewer", acceptance(c)), ErrPrerequisite)
	want(s.EnablePublic(ctx, "calcinaia", next, "operator"), ErrPrerequisite)
	t.Log("failed/stale/incomplete regression gates, corrected rerun, reacceptance, source/revision/suite isolation and immutable history verified")
}
