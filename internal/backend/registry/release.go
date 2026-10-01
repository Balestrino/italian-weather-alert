package registry

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var ToscanaRisks = []string{
	"minor_network_hydro",
	"main_network_hydraulic",
	"thunderstorms",
	"wind",
	"coastal_waves",
	"snow",
	"ice",
}

const releaseAcceptedSQL = `EXISTS(
 SELECT 1 FROM registry_events e JOIN registry_acceptance_reviews rr ON rr.acceptance_event_id=e.id
 WHERE e.source_id=$1 AND e.revision=$2 AND e.kind='acceptance' AND rr.status='accepted'
   AND e.id=(SELECT max(latest.id) FROM registry_events latest WHERE latest.source_id=e.source_id AND latest.revision=e.revision AND latest.kind='acceptance')
   AND NOT EXISTS(SELECT 1 FROM registry_regressions r WHERE r.source_id=e.source_id AND r.revision=e.revision AND NOT r.passed AND r.recorded_at>=e.created_at))`

type AcceptanceReviewReport struct {
	SourceID                 string   `json:"source_id"`
	Revision                 int      `json:"revision"`
	Product                  string   `json:"product"`
	Territory                string   `json:"territory"`
	Interfaces               string   `json:"api_mcp"`
	SevenRisks               string   `json:"seven_risks"`
	ScannedAttachments       string   `json:"scanned_attachments"`
	History                  string   `json:"history"`
	OperationalFailureStates string   `json:"operational_failure_states"`
	CoverageStatus           string   `json:"coverage_status"`
	CoverageLimitations      []string `json:"coverage_limitations"`
	Issues                   []string `json:"issues"`
}

type AcceptanceReview struct {
	ID                int64                  `json:"id"`
	AcceptanceEventID int64                  `json:"acceptance_event_id"`
	Status            string                 `json:"status"`
	Actor             string                 `json:"actor"`
	Evidence          Evidence               `json:"evidence"`
	Report            AcceptanceReviewReport `json:"report"`
	RecordedAt        time.Time              `json:"recorded_at"`
}

type ReleaseSource struct {
	SourceID            string   `json:"source_id"`
	Revision            *int     `json:"revision"`
	PublicEnabled       bool     `json:"public_enabled"`
	ReviewStatus        string   `json:"review_status"`
	CoverageStatus      string   `json:"coverage_status"`
	CoverageLimitations []string `json:"coverage_limitations"`
}

type ReleaseScope struct {
	Product     string          `json:"product"`
	Territory   string          `json:"territory"`
	Status      string          `json:"status"`
	Limitations []string        `json:"limitations"`
	Sources     []ReleaseSource `json:"sources"`
}

type ReleaseReadiness struct {
	Status string         `json:"status"`
	Scopes []ReleaseScope `json:"scopes"`
}

func regionalProduct(product string) bool {
	return product == "vigilance" || product == "criticality" || product == "monitoring"
}

func sameStrings(left, right []string) bool {
	a, b := slices.Clone(left), slices.Clone(right)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

func validCoverageStatus(status string, limitations []string) bool {
	if status != "accepted_declared_scope" && status != "accepted_with_limitations" {
		return false
	}
	if status == "accepted_with_limitations" && len(limitations) == 0 {
		return false
	}
	if status == "accepted_declared_scope" && len(limitations) != 0 {
		return false
	}
	for _, limitation := range limitations {
		if strings.TrimSpace(limitation) == "" {
			return false
		}
	}
	return true
}

func acceptanceReviewReport(st State, configuration Configuration, acceptance Acceptance) AcceptanceReviewReport {
	report := AcceptanceReviewReport{
		SourceID: st.Source.ID, Revision: *st.ActiveRevision, Product: st.Source.ProductID, Territory: st.Source.Territory,
		Interfaces: "verified", ScannedAttachments: "verified", History: "verified", OperationalFailureStates: "verified",
		SevenRisks: "not_applicable", CoverageStatus: acceptance.CoverageStatus,
		CoverageLimitations: append([]string(nil), acceptance.CoverageLimitations...), Issues: []string{},
	}
	if regionalProduct(st.Source.ProductID) {
		report.SevenRisks = "verified"
		if !sameStrings(acceptance.RiskCoverage, ToscanaRisks) {
			report.SevenRisks = "pending"
			report.Issues = append(report.Issues, "seven_risk_review_incomplete")
		}
	}
	checks := []struct {
		ok    bool
		field *string
		issue string
	}{
		{acceptance.InterfacesEquivalent, &report.Interfaces, "api_mcp_equivalence_unverified"},
		{acceptance.ScannedAttachmentsVerified, &report.ScannedAttachments, "scanned_attachment_behavior_unverified"},
		{acceptance.HistoryVerified, &report.History, "history_behavior_unverified"},
		{acceptance.FailureBehaviorVerified, &report.OperationalFailureStates, "operational_failure_behavior_unverified"},
	}
	for _, check := range checks {
		if !check.ok {
			*check.field = "pending"
			report.Issues = append(report.Issues, check.issue)
		}
	}
	for _, limitation := range configuration.Limitations {
		if !slices.Contains(report.CoverageLimitations, limitation) {
			report.Issues = append(report.Issues, "configuration_limitation_omitted")
		}
	}
	if !validCoverageStatus(report.CoverageStatus, report.CoverageLimitations) {
		report.Issues = append(report.Issues, "coverage_status_invalid")
	}
	slices.Sort(report.CoverageLimitations)
	return report
}

// ReviewAcceptance records the operator's final, source-scoped release review.
// It never enables publication; EnablePublic independently requires its accepted
// result for the latest acceptance event.
func (s *Store) ReviewAcceptance(ctx context.Context, id string, rev int, actor string, evidence Evidence) (AcceptanceReview, error) {
	if s == nil || s.pool == nil || rev < 1 || strings.TrimSpace(actor) == "" || !validEvidence(evidence) || evidence.ObservedAt.After(time.Now().UTC()) {
		return AcceptanceReview{}, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return AcceptanceReview{}, err
	}
	defer tx.Rollback(ctx)
	st, err := state(ctx, tx, id, true)
	if err != nil {
		return AcceptanceReview{}, err
	}
	configuration, err := version(ctx, tx, id, rev)
	if err != nil {
		return AcceptanceReview{}, err
	}
	if st.ActiveRevision == nil || *st.ActiveRevision != rev || !st.CollectionEnabled || (!regionalProduct(st.Source.ProductID) && st.Source.ProductID != "municipal") {
		return AcceptanceReview{}, ErrPrerequisite
	}
	ready, err := regressionReady(ctx, tx, id, rev)
	if err != nil || !ready {
		if err != nil {
			return AcceptanceReview{}, err
		}
		return AcceptanceReview{}, ErrPrerequisite
	}
	var eventID int64
	var raw []byte
	err = tx.QueryRow(ctx, `SELECT id,evidence FROM registry_events WHERE source_id=$1 AND revision=$2 AND kind='acceptance' ORDER BY id DESC LIMIT 1`, id, rev).Scan(&eventID, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return AcceptanceReview{}, ErrPrerequisite
	}
	if err != nil {
		return AcceptanceReview{}, err
	}
	var acceptance Acceptance
	if json.Unmarshal(raw, &acceptance) != nil {
		return AcceptanceReview{}, ErrInvalid
	}
	if evidence.ObservedAt.Before(acceptance.Report.ObservedAt) {
		return AcceptanceReview{}, ErrInvalid
	}
	report := acceptanceReviewReport(st, configuration.Configuration, acceptance)
	status := "pending"
	if len(report.Issues) == 0 {
		status = "accepted"
	}
	evidenceRaw, _ := json.Marshal(evidence)
	reportRaw, _ := json.Marshal(report)
	var result AcceptanceReview
	result.AcceptanceEventID, result.Status, result.Actor, result.Evidence, result.Report = eventID, status, actor, evidence, report
	err = tx.QueryRow(ctx, `INSERT INTO registry_acceptance_reviews(source_id,revision,acceptance_event_id,product_id,territory,status,actor,evidence,report)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id,recorded_at`, id, rev, eventID, st.Source.ProductID, st.Source.Territory, status, actor, evidenceRaw, reportRaw).Scan(&result.ID, &result.RecordedAt)
	if err != nil {
		return AcceptanceReview{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return AcceptanceReview{}, err
	}
	result.RecordedAt = result.RecordedAt.UTC()
	return result, nil
}

func (s *Store) AcceptanceReviews(ctx context.Context, id string) ([]AcceptanceReview, error) {
	if _, err := s.State(ctx, id); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT id,acceptance_event_id,status,actor,evidence,report,recorded_at FROM registry_acceptance_reviews WHERE source_id=$1 ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []AcceptanceReview{}
	for rows.Next() {
		var value AcceptanceReview
		var evidenceRaw, reportRaw []byte
		if err = rows.Scan(&value.ID, &value.AcceptanceEventID, &value.Status, &value.Actor, &evidenceRaw, &reportRaw, &value.RecordedAt); err != nil {
			return nil, err
		}
		if json.Unmarshal(evidenceRaw, &value.Evidence) != nil || json.Unmarshal(reportRaw, &value.Report) != nil {
			return nil, ErrInvalid
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (s *Store) ReleaseReadiness(ctx context.Context) (ReleaseReadiness, error) {
	required := []ReleaseScope{{Product: "vigilance", Territory: "Toscana"}, {Product: "criticality", Territory: "Toscana"}, {Product: "monitoring", Territory: "Toscana"}, {Product: "municipal", Territory: "050004"}}
	for i := range required {
		scope := &required[i]
		scope.Status, scope.Limitations, scope.Sources = "pending", []string{"accepted source review is unavailable"}, []ReleaseSource{}
		rows, err := s.pool.Query(ctx, `SELECT s.id,s.active_revision,s.public_enabled,
 COALESCE((SELECT rr.status FROM registry_acceptance_reviews rr JOIN registry_events e ON e.id=rr.acceptance_event_id
   WHERE rr.source_id=s.id AND rr.revision=s.active_revision AND e.id=(SELECT max(latest.id) FROM registry_events latest WHERE latest.source_id=s.id AND latest.revision=s.active_revision AND latest.kind='acceptance')
   AND NOT EXISTS(SELECT 1 FROM registry_regressions r WHERE r.source_id=s.id AND r.revision=s.active_revision AND NOT r.passed AND r.recorded_at>=e.created_at)
   ORDER BY rr.id DESC LIMIT 1),'pending'),
 COALESCE((SELECT rr.report FROM registry_acceptance_reviews rr JOIN registry_events e ON e.id=rr.acceptance_event_id
   WHERE rr.source_id=s.id AND rr.revision=s.active_revision AND e.id=(SELECT max(latest.id) FROM registry_events latest WHERE latest.source_id=s.id AND latest.revision=s.active_revision AND latest.kind='acceptance')
   AND NOT EXISTS(SELECT 1 FROM registry_regressions r WHERE r.source_id=s.id AND r.revision=s.active_revision AND NOT r.passed AND r.recorded_at>=e.created_at)
   ORDER BY rr.id DESC LIMIT 1),'{}'::jsonb)
 FROM registry_sources s WHERE s.product_id=$1 AND s.territory=$2 ORDER BY s.id`, scope.Product, scope.Territory)
		if err != nil {
			return ReleaseReadiness{}, err
		}
		accepted, limited := 0, false
		for rows.Next() {
			var source ReleaseSource
			var raw []byte
			if err = rows.Scan(&source.SourceID, &source.Revision, &source.PublicEnabled, &source.ReviewStatus, &raw); err != nil {
				rows.Close()
				return ReleaseReadiness{}, err
			}
			var review AcceptanceReviewReport
			_ = json.Unmarshal(raw, &review)
			source.CoverageStatus, source.CoverageLimitations = review.CoverageStatus, review.CoverageLimitations
			if source.CoverageLimitations == nil {
				source.CoverageLimitations = []string{}
			}
			if source.ReviewStatus == "accepted" {
				accepted++
			}
			if source.ReviewStatus != "accepted" || source.CoverageStatus == "accepted_with_limitations" || len(source.CoverageLimitations) > 0 {
				limited = true
			}
			scope.Sources = append(scope.Sources, source)
		}
		if err = rows.Err(); err != nil {
			rows.Close()
			return ReleaseReadiness{}, err
		}
		rows.Close()
		if accepted > 0 {
			scope.Status, scope.Limitations = "accepted_declared_scope", []string{}
			if limited || accepted != len(scope.Sources) {
				scope.Status = "accepted_with_limitations"
				scope.Limitations = []string{"one or more declared source scopes are pending or limited"}
			}
		}
	}
	result := ReleaseReadiness{Status: "ready", Scopes: required}
	acceptedScopes := 0
	for _, scope := range required {
		if scope.Status != "pending" {
			acceptedScopes++
		}
		if scope.Status != "accepted_declared_scope" {
			result.Status = "partial"
		}
	}
	if acceptedScopes == 0 {
		result.Status = "pending"
	}
	return result, nil
}
