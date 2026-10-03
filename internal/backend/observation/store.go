package observation

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Start(ctx context.Context, request StartRequest) (Campaign, error) {
	if s == nil || s.pool == nil || !validName(request.ID) || !validName(request.Actor) || request.StartedAt.IsZero() || request.StartedAt.After(time.Now().UTC()) || len(request.SourceIDs) == 0 || len(request.SourceIDs) > 32 {
		return Campaign{}, ErrInvalid
	}
	request.StartedAt = request.StartedAt.UTC()
	seen := map[string]bool{}
	for _, id := range request.SourceIDs {
		if !validName(id) || seen[id] {
			return Campaign{}, ErrInvalid
		}
		seen[id] = true
	}
	slices.Sort(request.SourceIDs)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Campaign{}, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, "INSERT INTO observation_campaigns(id,actor,started_at) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", request.ID, request.Actor, request.StartedAt)
	if err != nil {
		return Campaign{}, err
	}
	if tag.RowsAffected() == 0 {
		return Campaign{}, ErrConflict
	}
	for _, id := range request.SourceIDs {
		var active *int
		var product, territory string
		var enabled bool
		var body []byte
		var checkSeconds, delaySeconds int
		err = tx.QueryRow(ctx, `SELECT s.active_revision,s.product_id,s.territory,s.collection_enabled,c.body,
 COALESCE(s.check_seconds_override,(c.body->>'check_seconds')::integer,600),
 COALESCE(s.delay_seconds_override,(c.body->>'delay_seconds')::integer,1800)
 FROM registry_sources s LEFT JOIN registry_configurations c ON c.source_id=s.id AND c.revision=s.active_revision WHERE s.id=$1`, id).
			Scan(&active, &product, &territory, &enabled, &body, &checkSeconds, &delaySeconds)
		if errors.Is(err, pgx.ErrNoRows) {
			return Campaign{}, ErrInvalid
		}
		if err != nil {
			return Campaign{}, err
		}
		var configuration registry.Configuration
		if active == nil || !enabled || json.Unmarshal(body, &configuration) != nil || len(configuration.Sections) == 0 || len(configuration.Unresolved) != 0 || configuration.Provenance == nil || !configuration.Policy.CollectionPermitted || !configuration.Policy.RetentionPermitted || checkSeconds < 1 || delaySeconds < 1 {
			return Campaign{}, ErrInvalid
		}
		if product == "municipal" {
			if territory != "050004" {
				return Campaign{}, ErrInvalid
			}
		} else if (product != "vigilance" && product != "criticality" && product != "monitoring") || territory != "Toscana" {
			return Campaign{}, ErrInvalid
		}
		sections, _ := json.Marshal(configuration.Sections)
		if _, err = tx.Exec(ctx, `INSERT INTO observation_campaign_sources(campaign_id,source_id,configuration,product_id,territory,sections,check_seconds,delay_seconds)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, request.ID, id, *active, product, territory, sections, checkSeconds, delaySeconds); err != nil {
			return Campaign{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Campaign{}, err
	}
	return s.Get(ctx, request.ID)
}

func (s *Store) RecordReview(ctx context.Context, campaignID, actor string, review Review) (Review, error) {
	if s == nil || s.pool == nil || !validName(campaignID) || !validName(actor) || !validReview(review) {
		return Review{}, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Review{}, err
	}
	defer tx.Rollback(ctx)
	var started time.Time
	var configuration int
	err = tx.QueryRow(ctx, `SELECT c.started_at,s.configuration FROM observation_campaigns c
 JOIN observation_campaign_sources s ON s.campaign_id=c.id WHERE c.id=$1 AND s.source_id=$2`, campaignID, review.SourceID).Scan(&started, &configuration)
	if errors.Is(err, pgx.ErrNoRows) {
		return Review{}, ErrNotFound
	}
	if err != nil {
		return Review{}, err
	}
	var completed bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM observation_assessments WHERE campaign_id=$1 AND status='complete')", campaignID).Scan(&completed); err != nil {
		return Review{}, err
	}
	if completed || review.Evidence.ObservedAt.Before(started) || review.Evidence.ObservedAt.After(time.Now().UTC()) {
		return Review{}, ErrConflict
	}
	switch review.Kind {
	case "original_comparison", "attachment_comparison":
		var matches bool
		roleCondition := " AND r.role='original'"
		if review.Kind == "attachment_comparison" {
			roleCondition = " AND r.role IN ('attachment','resource')"
		}
		if review.Status == "pass" {
			roleCondition += " AND r.object_hash IS NOT NULL"
		}
		// Unchanged originals retain their first resource configuration. A later
		// finalized acquisition proves reuse under the campaign configuration
		// without rewriting that original or creating a new content version.
		query := `SELECT EXISTS(SELECT 1 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
	 JOIN retained_resources r ON r.version_id=v.id WHERE v.id=$1 AND d.source_id=$2 AND r.source_id=$2
	 AND (r.configuration=$3 OR EXISTS(SELECT 1 FROM retained_acquisitions a
	 WHERE a.version_id=v.id AND a.source_id=$2 AND a.configuration=$3 AND a.content_hash=v.content_hash AND a.acquired_at<=$4))` + roleCondition + `)`
		if err = tx.QueryRow(ctx, query, *review.VersionID, review.SourceID, configuration, review.Evidence.ObservedAt.UTC()).Scan(&matches); err != nil {
			return Review{}, err
		}
		if !matches {
			return Review{}, ErrInvalid
		}
	case "failure_actual":
		var matches bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM acquisition_checks WHERE id=$1 AND source_id=$2 AND configuration=$3
 AND NOT complete AND finished_at >= $4 AND finished_at <= $5)`, *review.CheckID, review.SourceID, configuration, started, review.Evidence.ObservedAt.UTC()).Scan(&matches); err != nil {
			return Review{}, err
		}
		if !matches {
			return Review{}, ErrInvalid
		}
	}
	evidence, _ := json.Marshal(review.Evidence)
	tag, err := tx.Exec(ctx, `INSERT INTO observation_reviews(campaign_id,id,source_id,kind,status,check_id,version_id,case_id,evidence,notes,actor)
 VALUES($1,$2,$3,$4,$5,$6,$7,NULLIF($8,''),$9,$10,$11) ON CONFLICT DO NOTHING`, campaignID, review.ID, review.SourceID, review.Kind, review.Status, review.CheckID, review.VersionID, review.CaseID, evidence, review.Notes, actor)
	if err != nil {
		return Review{}, err
	}
	if tag.RowsAffected() == 0 {
		return Review{}, ErrConflict
	}
	if err = tx.QueryRow(ctx, "SELECT recorded_at FROM observation_reviews WHERE campaign_id=$1 AND id=$2", campaignID, review.ID).Scan(&review.RecordedAt); err != nil {
		return Review{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Review{}, err
	}
	review.Actor, review.RecordedAt = actor, review.RecordedAt.UTC()
	return review, nil
}

func (s *Store) Assess(ctx context.Context, campaignID, actor string, through time.Time, evidence registry.Evidence) (Assessment, error) {
	now := time.Now().UTC()
	if s == nil || s.pool == nil || !validName(campaignID) || !validName(actor) || through.IsZero() || through.After(now) || !validEvidence(evidence) || evidence.ObservedAt.Before(through) || evidence.ObservedAt.After(now) {
		return Assessment{}, ErrInvalid
	}
	var completed bool
	if err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM observation_assessments WHERE campaign_id=$1 AND status='complete')", campaignID).Scan(&completed); err != nil {
		return Assessment{}, err
	}
	if completed {
		return Assessment{}, ErrConflict
	}
	report, err := s.Report(ctx, campaignID, through.UTC())
	if err != nil {
		return Assessment{}, err
	}
	evidenceBody, _ := json.Marshal(evidence)
	reportBody, _ := json.Marshal(report)
	var result Assessment
	result.CampaignID, result.Actor, result.Through, result.Status, result.Evidence, result.Report = campaignID, actor, through.UTC(), report.Status, evidence, report
	err = s.pool.QueryRow(ctx, `INSERT INTO observation_assessments(campaign_id,actor,through_at,status,evidence,report)
 VALUES($1,$2,$3,$4,$5,$6) RETURNING id,recorded_at`, campaignID, actor, result.Through, result.Status, evidenceBody, reportBody).Scan(&result.ID, &result.RecordedAt)
	if err != nil {
		return Assessment{}, err
	}
	result.RecordedAt = result.RecordedAt.UTC()
	return result, nil
}

func (s *Store) Report(ctx context.Context, campaignID string, through time.Time) (Report, error) {
	if s == nil || s.pool == nil || !validName(campaignID) || through.IsZero() || through.After(time.Now().UTC()) {
		return Report{}, ErrInvalid
	}
	campaign, err := s.campaign(ctx, campaignID, false)
	if err != nil {
		return Report{}, err
	}
	through = through.UTC()
	if through.Before(campaign.StartedAt) {
		return Report{}, ErrInvalid
	}
	report := Report{CampaignID: campaign.ID, StartedAt: campaign.StartedAt, Through: through, MinimumEndAt: campaign.MinimumEndAt}
	products := map[string]bool{}
	for _, source := range campaign.Sources {
		products[source.Product] = true
		item, sourceErr := s.sourceReport(ctx, campaign, source, through)
		if sourceErr != nil {
			return Report{}, sourceErr
		}
		report.Sources = append(report.Sources, item)
		for _, issue := range item.Issues {
			report.Issues = append(report.Issues, source.SourceID+":"+issue)
		}
	}
	for _, product := range []string{"vigilance", "criticality", "monitoring", "municipal"} {
		if !products[product] {
			report.MissingProducts = append(report.MissingProducts, product)
			report.Issues = append(report.Issues, "missing_product:"+product)
		}
	}
	if through.Before(campaign.MinimumEndAt) {
		report.Status = "running"
		report.Issues = append(report.Issues, "minimum_duration_not_reached")
	} else if len(report.Issues) != 0 {
		report.Status = "extended"
	} else {
		report.Status = "complete"
	}
	return report, nil
}

func (s *Store) sourceReport(ctx context.Context, campaign Campaign, source Source, through time.Time) (SourceReport, error) {
	result := SourceReport{Source: source}
	rows, err := s.pool.Query(ctx, `SELECT finished_at,complete FROM acquisition_checks
 WHERE source_id=$1 AND configuration=$2 AND finished_at >= $3 AND finished_at <= $4 ORDER BY finished_at,id`, source.SourceID, source.Configuration, campaign.StartedAt, through)
	if err != nil {
		return result, err
	}
	var checks []time.Time
	days := map[string]bool{}
	rome, _ := time.LoadLocation("Europe/Rome")
	for rows.Next() {
		var finished time.Time
		var complete bool
		if err = rows.Scan(&finished, &complete); err != nil {
			rows.Close()
			return result, err
		}
		finished = finished.UTC()
		checks = append(checks, finished)
		result.Checks++
		if complete {
			result.CompleteChecks++
		} else {
			result.FailedChecks++
		}
		days[finished.In(rome).Format("2006-01-02")] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	result.ObservedLocalDays = len(days)
	if len(checks) == 0 {
		result.Issues = append(result.Issues, "no_persisted_checks")
	} else {
		first, last := checks[0], checks[len(checks)-1]
		result.FirstCheckAt, result.LastCheckAt = &first, &last
		previous := campaign.StartedAt
		for _, checked := range checks {
			result.MaximumGapSeconds = max(result.MaximumGapSeconds, int64(checked.Sub(previous).Seconds()))
			previous = checked
		}
		result.MaximumGapSeconds = max(result.MaximumGapSeconds, int64(through.Sub(previous).Seconds()))
		if result.MaximumGapSeconds > int64(source.DelaySeconds) {
			result.Issues = append(result.Issues, "updating_gap_exceeded")
		}
	}
	if result.CompleteChecks == 0 {
		result.Issues = append(result.Issues, "no_complete_check")
	}
	if !through.Before(campaign.MinimumEndAt) && result.ObservedLocalDays < 7 {
		result.Issues = append(result.Issues, "fewer_than_seven_observed_days")
	}
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 WHERE d.source_id=$1 AND v.first_acquired_at >= $2 AND v.first_acquired_at <= $3`, source.SourceID, campaign.StartedAt, through).Scan(&result.VersionsAcquired); err != nil {
		return result, err
	}
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FILTER(WHERE r.required),count(*) FILTER(WHERE r.required AND r.object_hash IS NULL)
 FROM retained_resources r JOIN retained_versions v ON v.id=r.version_id JOIN retained_documents d ON d.id=v.document_id
 WHERE d.source_id=$1 AND r.role IN ('attachment','resource') AND
 (v.first_acquired_at BETWEEN $2 AND $3 OR EXISTS(SELECT 1 FROM observation_reviews review
   WHERE review.campaign_id=$4 AND review.source_id=$1 AND review.version_id=v.id AND review.kind='original_comparison'))`, source.SourceID, campaign.StartedAt, through, campaign.ID).
		Scan(&result.RequiredResources, &result.MissingRequiredResources); err != nil {
		return result, err
	}
	reviews, err := s.reviews(ctx, campaign.ID, source.SourceID)
	if err != nil {
		return result, err
	}
	result.Reviews = reviews
	attachmentNA := 0
	failureActual, failureRetained := 0, 0
	for _, review := range reviews {
		if review.Status != "pass" {
			result.Issues = append(result.Issues, "review_"+review.Status+":"+review.ID)
			continue
		}
		switch review.Kind {
		case "original_comparison":
			result.OriginalComparisons++
		case "attachment_comparison":
			result.AttachmentComparisons++
		case "attachment_not_applicable":
			attachmentNA++
		case "failure_actual":
			failureActual++
			result.FailureExercises++
		case "failure_retained":
			failureRetained++
			result.FailureExercises++
		case "absent_event_retained":
			result.RetainedEventCases++
		}
	}
	if result.OriginalComparisons == 0 {
		result.Issues = append(result.Issues, "original_comparison_missing")
	}
	if result.RequiredResources > 0 && result.AttachmentComparisons == 0 {
		result.Issues = append(result.Issues, "attachment_comparison_missing")
	}
	if result.RequiredResources == 0 && attachmentNA == 0 {
		result.Issues = append(result.Issues, "attachment_scope_unreviewed")
	}
	if result.MissingRequiredResources > 0 {
		result.Issues = append(result.Issues, "required_resource_missing")
	}
	if result.FailedChecks > 0 && failureActual == 0 {
		result.Issues = append(result.Issues, "observed_failure_unreviewed")
	}
	if result.FailedChecks == 0 && failureRetained == 0 {
		result.Issues = append(result.Issues, "failure_behavior_unexercised")
	}
	if result.VersionsAcquired == 0 && result.RetainedEventCases == 0 {
		result.Issues = append(result.Issues, "absent_event_case_missing")
	}
	return result, nil
}

func (s *Store) Get(ctx context.Context, id string) (Campaign, error) {
	return s.campaign(ctx, id, true)
}

func (s *Store) List(ctx context.Context) ([]Campaign, error) {
	rows, err := s.pool.Query(ctx, "SELECT id FROM observation_campaigns ORDER BY started_at,id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	result := make([]Campaign, 0, len(ids))
	for _, id := range ids {
		campaign, getErr := s.Get(ctx, id)
		if getErr != nil {
			return nil, getErr
		}
		result = append(result, campaign)
	}
	return result, nil
}

func (s *Store) campaign(ctx context.Context, id string, latest bool) (Campaign, error) {
	var campaign Campaign
	err := s.pool.QueryRow(ctx, "SELECT id,actor,started_at,created_at FROM observation_campaigns WHERE id=$1", id).Scan(&campaign.ID, &campaign.Actor, &campaign.StartedAt, &campaign.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, ErrNotFound
	}
	if err != nil {
		return Campaign{}, err
	}
	campaign.StartedAt, campaign.CreatedAt = campaign.StartedAt.UTC(), campaign.CreatedAt.UTC()
	campaign.MinimumEndAt = campaign.StartedAt.Add(MinimumDuration)
	rows, err := s.pool.Query(ctx, `SELECT source_id,configuration,product_id,territory,sections,check_seconds,delay_seconds
 FROM observation_campaign_sources WHERE campaign_id=$1 ORDER BY source_id`, id)
	if err != nil {
		return Campaign{}, err
	}
	for rows.Next() {
		var source Source
		var sections []byte
		if err = rows.Scan(&source.SourceID, &source.Configuration, &source.Product, &source.Territory, &sections, &source.CheckSeconds, &source.DelaySeconds); err != nil {
			rows.Close()
			return Campaign{}, err
		}
		if json.Unmarshal(sections, &source.Sections) != nil {
			rows.Close()
			return Campaign{}, ErrInvalid
		}
		campaign.Sources = append(campaign.Sources, source)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return Campaign{}, err
	}
	if latest {
		assessment, found, latestErr := s.latestAssessment(ctx, id)
		if latestErr != nil {
			return Campaign{}, latestErr
		}
		if found {
			campaign.Latest = &assessment
		}
	}
	return campaign, nil
}

func (s *Store) latestAssessment(ctx context.Context, campaignID string) (Assessment, bool, error) {
	var result Assessment
	var evidence, report []byte
	err := s.pool.QueryRow(ctx, `SELECT id,campaign_id,actor,through_at,status,evidence,report,recorded_at
 FROM observation_assessments WHERE campaign_id=$1 ORDER BY id DESC LIMIT 1`, campaignID).
		Scan(&result.ID, &result.CampaignID, &result.Actor, &result.Through, &result.Status, &evidence, &report, &result.RecordedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Assessment{}, false, nil
	}
	if err != nil {
		return Assessment{}, false, err
	}
	if json.Unmarshal(evidence, &result.Evidence) != nil || json.Unmarshal(report, &result.Report) != nil {
		return Assessment{}, false, ErrInvalid
	}
	result.Through, result.RecordedAt = result.Through.UTC(), result.RecordedAt.UTC()
	return result, true, nil
}

func (s *Store) reviews(ctx context.Context, campaignID, sourceID string) ([]Review, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,source_id,kind,status,check_id,version_id,COALESCE(case_id,''),evidence,notes,actor,recorded_at
 FROM observation_reviews WHERE campaign_id=$1 AND source_id=$2 ORDER BY recorded_at,id`, campaignID, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Review
	for rows.Next() {
		var review Review
		var evidence []byte
		if err = rows.Scan(&review.ID, &review.SourceID, &review.Kind, &review.Status, &review.CheckID, &review.VersionID, &review.CaseID, &evidence, &review.Notes, &review.Actor, &review.RecordedAt); err != nil {
			return nil, err
		}
		if json.Unmarshal(evidence, &review.Evidence) != nil {
			return nil, ErrInvalid
		}
		review.RecordedAt = review.RecordedAt.UTC()
		result = append(result, review)
	}
	return result, rows.Err()
}

func validReview(review Review) bool {
	if !validName(review.ID) || !validName(review.SourceID) || (review.Status != "pass" && review.Status != "fail" && review.Status != "unresolved") || !validEvidence(review.Evidence) || len(review.Notes) > 4000 {
		return false
	}
	switch review.Kind {
	case "original_comparison", "attachment_comparison":
		return review.VersionID != nil && *review.VersionID > 0 && review.CheckID == nil && review.CaseID == ""
	case "failure_actual":
		return review.CheckID != nil && *review.CheckID > 0 && review.VersionID == nil && review.CaseID == ""
	case "attachment_not_applicable":
		return review.CheckID == nil && review.VersionID == nil && review.CaseID == ""
	case "failure_retained", "absent_event_retained":
		return review.CheckID == nil && review.VersionID == nil && validName(review.CaseID)
	default:
		return false
	}
}

func validEvidence(e registry.Evidence) bool {
	u, err := url.Parse(e.URL)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && u.User == nil && u.Fragment == "" && strings.TrimSpace(e.Locator) != "" && !e.ObservedAt.IsZero()
}

func validName(value string) bool {
	return value != "" && value == strings.TrimSpace(value) && len(value) <= 200 && !strings.ContainsAny(value, "\r\n")
}
