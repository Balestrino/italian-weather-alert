package registry

import "testing"

func TestAcceptanceReviewReportRequiresExplicitScopeEvidence(t *testing.T) {
	revision := 2
	state := State{Source: Source{ID: "criticality", ProductID: "criticality", Territory: "Toscana"}, ActiveRevision: &revision}
	configuration := Configuration{Limitations: []string{"graphical extraction limited to retained resources"}}
	acceptance := Acceptance{
		RiskCoverage:               append([]string(nil), ToscanaRisks...),
		InterfacesEquivalent:       true,
		ScannedAttachmentsVerified: true,
		HistoryVerified:            true,
		FailureBehaviorVerified:    true,
		CoverageStatus:             "accepted_with_limitations",
		CoverageLimitations:        []string{"graphical extraction limited to retained resources"},
	}
	report := acceptanceReviewReport(state, configuration, acceptance)
	if len(report.Issues) != 0 || report.SevenRisks != "verified" || report.CoverageStatus != "accepted_with_limitations" {
		t.Fatalf("valid review rejected: %#v", report)
	}

	acceptance.RiskCoverage = acceptance.RiskCoverage[:6]
	acceptance.CoverageLimitations = []string{"different limitation"}
	report = acceptanceReviewReport(state, configuration, acceptance)
	if len(report.Issues) != 2 || report.SevenRisks != "pending" {
		t.Fatalf("missing risk/scope evidence was not retained: %#v", report)
	}
}

func TestCoverageStatusNeverUsesUnboundedCompleteClaim(t *testing.T) {
	for _, status := range []string{"complete", "full", "municipality_complete", ""} {
		if validCoverageStatus(status, nil) {
			t.Fatalf("accepted unbounded status %q", status)
		}
	}
	if !validCoverageStatus("accepted_declared_scope", nil) || !validCoverageStatus("accepted_with_limitations", []string{"one configured channel is pending"}) {
		t.Fatal("bounded statuses rejected")
	}
	if validCoverageStatus("accepted_declared_scope", []string{"undeclared gap"}) {
		t.Fatal("declared scope accepted with contradictory limitations")
	}
}
