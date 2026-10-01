package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

type runEvidence struct {
	measurement   Measurement
	configuration ConfigurationIdentity
	documentID    *int64
	createdAt     time.Time
	firstAttempt  time.Time
	lastAttempt   time.Time
}

func (s *Store) RecordComparison(ctx context.Context, actor string, comparison ProcessingComparison) (ComparisonReport, error) {
	if s == nil || s.pool == nil || strings.TrimSpace(actor) == "" || actor != strings.TrimSpace(actor) || comparison.Validate() != nil || comparison.FinishedAt.After(time.Now().UTC()) {
		return ComparisonReport{}, ErrComparisonInvalid
	}
	report := ComparisonReport{Comparison: comparison, RecordedBy: actor}
	baselineConfigs, candidateConfigs := map[string]ConfigurationIdentity{}, map[string]ConfigurationIdentity{}
	baselineDocuments, candidateDocuments := map[string]map[int64]bool{}, map[string]map[int64]bool{}
	for _, item := range comparison.Cases {
		if item.Baseline.Status == "pass" {
			report.Baseline.PassedCases++
		} else {
			report.Baseline.FailedCases++
		}
		if item.Candidate.Status == "pass" {
			report.Candidate.PassedCases++
		} else {
			report.Candidate.FailedCases++
		}
		for _, variant := range []struct {
			name      string
			value     Outcome
			report    *VariantReport
			configs   map[string]ConfigurationIdentity
			documents map[string]map[int64]bool
		}{{"baseline", item.Baseline, &report.Baseline, baselineConfigs, baselineDocuments}, {"candidate", item.Candidate, &report.Candidate, candidateConfigs, candidateDocuments}} {
			for _, runID := range variant.value.RunIDs {
				evidence, err := s.resolveRun(ctx, item.ID, variant.name, runID)
				if err != nil {
					return ComparisonReport{}, err
				}
				if evidence.createdAt.Before(comparison.StartedAt) || evidence.createdAt.After(comparison.FinishedAt) || evidence.firstAttempt.Before(comparison.StartedAt) || evidence.lastAttempt.After(comparison.FinishedAt) {
					return ComparisonReport{}, ErrComparisonInvalid
				}
				variant.report.Measurements = append(variant.report.Measurements, evidence.measurement)
				variant.configs[evidence.configuration.ID] = evidence.configuration
				if evidence.documentID != nil {
					if variant.documents[item.ID] == nil {
						variant.documents[item.ID] = map[int64]bool{}
					}
					variant.documents[item.ID][*evidence.documentID] = true
				}
			}
		}
	}
	for _, item := range comparison.Cases {
		if !sameIDs(baselineDocuments[item.ID], candidateDocuments[item.ID]) {
			return ComparisonReport{}, ErrComparisonInvalid
		}
	}
	report.Baseline.Configurations = configValues(baselineConfigs)
	report.Candidate.Configurations = configValues(candidateConfigs)
	if !hasStage(report.Baseline.Configurations, "linking") || hasStage(report.Baseline.Configurations, "embedding") || !hasStage(report.Candidate.Configurations, "linking") || !hasStage(report.Candidate.Configurations, "embedding") {
		return ComparisonReport{}, ErrComparisonInvalid
	}
	report.Baseline.ProcessVersion = processVersion(report.Baseline.Configurations)
	report.Candidate.ProcessVersion = processVersion(report.Candidate.Configurations)
	if report.Baseline.ProcessVersion == report.Candidate.ProcessVersion {
		return ComparisonReport{}, ErrComparisonInvalid
	}
	report.Passed = report.Candidate.FailedCases == 0 && report.Candidate.PassedCases == len(comparison.Cases)
	body, err := json.Marshal(report)
	if err != nil {
		return ComparisonReport{}, ErrComparisonInvalid
	}
	tag, err := s.pool.Exec(ctx, `INSERT INTO processing_change_evaluations(id,corpus_sha256,baseline_process_version,candidate_process_version,passed,actor,report) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`, comparison.ID, comparison.CorpusSHA256, report.Baseline.ProcessVersion, report.Candidate.ProcessVersion, report.Passed, actor, body)
	if err != nil {
		return ComparisonReport{}, err
	}
	if tag.RowsAffected() == 0 {
		var existing []byte
		if err = s.pool.QueryRow(ctx, "SELECT report FROM processing_change_evaluations WHERE id=$1", comparison.ID).Scan(&existing); err != nil {
			return ComparisonReport{}, err
		}
		var prior ComparisonReport
		if json.Unmarshal(existing, &prior) != nil || !reflect.DeepEqual(prior, report) {
			return ComparisonReport{}, ErrComparisonInvalid
		}
		return prior, nil
	}
	return s.Get(ctx, comparison.ID)
}

func (s *Store) resolveRun(ctx context.Context, caseID, variant string, runID int64) (runEvidence, error) {
	var result runEvidence
	var workload string
	err := s.pool.QueryRow(ctx, `SELECT r.workload,r.stage,r.configuration_version_id,r.document_version_id,c.revision,c.logic_version,c.content_hash,c.model_version_id,c.prompt_version_id,r.created_at
 FROM processing_runs r JOIN processing_configuration_versions c ON c.id=r.configuration_version_id WHERE r.id=$1`, runID).Scan(&workload, &result.measurement.Stage, &result.measurement.ConfigurationVersionID, &result.documentID, &result.configuration.Revision, &result.configuration.LogicVersion, &result.configuration.ContentSHA256, &result.configuration.ModelVersionID, &result.configuration.PromptVersionID, &result.createdAt)
	if errors.Is(err, pgx.ErrNoRows) || workload != "evaluation" {
		return runEvidence{}, ErrComparisonInvalid
	}
	if err != nil {
		return runEvidence{}, err
	}
	result.configuration.ID, result.configuration.Stage = result.measurement.ConfigurationVersionID, result.measurement.Stage
	result.measurement.CaseID, result.measurement.Variant, result.measurement.RunID = caseID, variant, runID
	var running int64
	var currencies []string
	err = s.pool.QueryRow(ctx, `SELECT count(*),COALESCE(sum(duration_ms),0),count(*) FILTER(WHERE finished_at IS NULL),
 count(*) FILTER(WHERE usage_status IN('partial','unavailable')),count(*) FILTER(WHERE cost_status='unknown'),
 sum(input_tokens),sum(output_tokens),sum(cache_read_tokens),sum(cache_write_tokens),sum(estimated_cost_microunits),
	 COALESCE(array_agg(DISTINCT p.currency) FILTER(WHERE p.currency IS NOT NULL),'{}'),min(a.started_at),max(a.finished_at)
 FROM processing_run_attempts a LEFT JOIN processing_price_versions p ON p.id=a.price_version_id WHERE a.run_id=$1`, runID).Scan(&result.measurement.Attempts, &result.measurement.DurationMS, &running, &result.measurement.UnknownUsageAttempts, &result.measurement.UnknownCostAttempts, &result.measurement.InputTokens, &result.measurement.OutputTokens, &result.measurement.CacheReadTokens, &result.measurement.CacheWriteTokens, &result.measurement.EstimatedCostMicrounits, &currencies, &result.firstAttempt, &result.lastAttempt)
	if err != nil {
		return runEvidence{}, err
	}
	if result.measurement.Attempts < 1 || running != 0 || len(currencies) > 1 {
		return runEvidence{}, ErrComparisonInvalid
	}
	if len(currencies) == 1 {
		result.measurement.Currency = currencies[0]
	}
	return result, nil
}

func (s *Store) Get(ctx context.Context, id string) (ComparisonReport, error) {
	var body []byte
	var recorded time.Time
	if err := s.pool.QueryRow(ctx, "SELECT report,recorded_at FROM processing_change_evaluations WHERE id=$1", id).Scan(&body, &recorded); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ComparisonReport{}, ErrComparisonInvalid
		}
		return ComparisonReport{}, err
	}
	var report ComparisonReport
	if json.Unmarshal(body, &report) != nil {
		return ComparisonReport{}, ErrComparisonInvalid
	}
	report.RecordedAt = recorded.UTC()
	return report, nil
}

func (s *Store) List(ctx context.Context) ([]ComparisonReport, error) {
	rows, err := s.pool.Query(ctx, "SELECT id FROM processing_change_evaluations ORDER BY recorded_at,id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ComparisonReport
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		value, getErr := s.Get(ctx, id)
		if getErr != nil {
			return nil, getErr
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (s *Store) Passed(ctx context.Context, id string) (string, error) {
	var version string
	var passed bool
	err := s.pool.QueryRow(ctx, "SELECT candidate_process_version,passed FROM processing_change_evaluations WHERE id=$1", id).Scan(&version, &passed)
	if err != nil || !passed {
		return "", ErrComparisonInvalid
	}
	return version, nil
}

func configValues(values map[string]ConfigurationIdentity) []ConfigurationIdentity {
	result := make([]ConfigurationIdentity, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func hasStage(values []ConfigurationIdentity, stage string) bool {
	for _, value := range values {
		if value.Stage == stage {
			return true
		}
	}
	return false
}

func sameIDs(a, b map[int64]bool) bool {
	if len(a) == 0 || len(a) != len(b) {
		return false
	}
	for id := range a {
		if !b[id] {
			return false
		}
	}
	return true
}
