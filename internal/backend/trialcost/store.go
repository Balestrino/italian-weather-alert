package trialcost

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) RecordInput(ctx context.Context, campaignID, actor string, input Input) (Input, error) {
	if s == nil || s.pool == nil || !validName(campaignID) || !validName(actor) || !validInput(input) {
		return Input{}, ErrInvalid
	}
	input.CampaignID, input.Actor = campaignID, actor
	if input.Status == "reported" {
		cost, ok := pricedCost(*input.Quantity, *input.UnitSize, *input.PriceMicrounits)
		if !ok {
			return Input{}, ErrInvalid
		}
		input.CostMicrounits = &cost
		from, through := input.EffectiveFrom.UTC(), input.EffectiveThrough.UTC()
		input.EffectiveFrom, input.EffectiveThrough = &from, &through
	}
	input.Evidence.ObservedAt = input.Evidence.ObservedAt.UTC()
	evidence, _ := json.Marshal(input.Evidence)
	lockKey, _ := json.Marshal([]string{campaignID, input.Category, input.Workload, input.Component})
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Input{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", string(lockKey)); err != nil {
		return Input{}, err
	}
	var exists bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM observation_campaigns WHERE id=$1)", campaignID).Scan(&exists); err != nil {
		return Input{}, err
	}
	if !exists {
		return Input{}, ErrNotFound
	}
	if err = tx.QueryRow(ctx, `SELECT COALESCE(max(revision),0)+1 FROM trial_cost_inputs
 WHERE campaign_id=$1 AND category=$2 AND workload=$3 AND component=$4`, campaignID, input.Category, input.Workload, input.Component).Scan(&input.Revision); err != nil {
		return Input{}, err
	}
	err = tx.QueryRow(ctx, `INSERT INTO trial_cost_inputs(campaign_id,category,workload,component,revision,status,basis,quantity,unit,unit_size,price_microunits,cost_microunits,currency,effective_from,effective_through,evidence,notes,actor)
 VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8,NULLIF($9,''),$10,$11,$12,NULLIF($13,''),$14,$15,$16,$17,$18) RETURNING recorded_at`,
		campaignID, input.Category, input.Workload, input.Component, input.Revision, input.Status, input.Basis, input.Quantity, input.Unit, input.UnitSize, input.PriceMicrounits, input.CostMicrounits, input.Currency, input.EffectiveFrom, input.EffectiveThrough, evidence, input.Notes, actor).Scan(&input.RecordedAt)
	if err != nil {
		return Input{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Input{}, err
	}
	input.RecordedAt = input.RecordedAt.UTC()
	return input, nil
}

func (s *Store) Report(ctx context.Context, campaignID string, through time.Time) (Report, error) {
	if s == nil || s.pool == nil || !validName(campaignID) || through.IsZero() || through.After(time.Now().UTC()) {
		return Report{}, ErrInvalid
	}
	through = through.UTC()
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Report{}, err
	}
	defer tx.Rollback(ctx)
	report := Report{CampaignID: campaignID, Through: through, ObservedAt: time.Now().UTC()}
	err = tx.QueryRow(ctx, `SELECT started_at,EXISTS(SELECT 1 FROM observation_assessments a
 WHERE a.campaign_id=c.id AND a.status='complete' AND a.through_at=$2)
 FROM observation_campaigns c WHERE c.id=$1`, campaignID, through).Scan(&report.StartedAt, &report.ObservationComplete)
	if errors.Is(err, pgx.ErrNoRows) {
		return Report{}, ErrNotFound
	}
	if err != nil {
		return Report{}, err
	}
	report.StartedAt = report.StartedAt.UTC()
	if !through.After(report.StartedAt) {
		return Report{}, ErrInvalid
	}
	report.DurationSeconds = int64(through.Sub(report.StartedAt).Seconds())
	if report.Breakdown, err = processingBreakdown(ctx, tx, campaignID, report.StartedAt, through); err != nil {
		return Report{}, err
	}
	if err = attachOtherUnits(ctx, tx, campaignID, report.StartedAt, through, report.Breakdown); err != nil {
		return Report{}, err
	}
	if report.Prices, err = pricingAssumptions(ctx, tx, campaignID, report.StartedAt, through); err != nil {
		return Report{}, err
	}
	if report.Infrastructure, err = latestInputs(ctx, tx, campaignID); err != nil {
		return Report{}, err
	}
	report.MissingMetrics, report.Currencies, err = completeness(ctx, tx, report)
	if err != nil {
		return Report{}, err
	}
	report.BudgetReady = len(report.MissingMetrics) == 0
	if err = tx.Commit(ctx); err != nil {
		return Report{}, err
	}
	return report, nil
}

func processingBreakdown(ctx context.Context, tx pgx.Tx, campaignID string, started, through time.Time) ([]Breakdown, error) {
	rows, err := tx.Query(ctx, `SELECT r.workload,r.stage,COALESCE(m.provider,''),COALESCE(m.model,''),COALESCE(p.currency,''),
 count(DISTINCT r.id),count(*),count(*) FILTER(WHERE a.number>1),
 count(*) FILTER(WHERE a.usage_status IN ('partial','unavailable')),
 count(*) FILTER(WHERE a.cost_status='unknown'),
 count(*) FILTER(WHERE a.number>1 AND a.cost_status='unknown'),
 sum(a.duration_ms),sum(a.input_tokens),sum(a.output_tokens),sum(a.cache_read_tokens),sum(a.cache_write_tokens),
 sum(a.estimated_cost_microunits),sum(a.estimated_cost_microunits) FILTER(WHERE a.number>1)
 FROM processing_runs r JOIN observation_campaign_sources cs ON cs.source_id=r.source_id AND cs.campaign_id=$1
 JOIN processing_run_attempts a ON a.run_id=r.id
 JOIN processing_configuration_versions c ON c.id=r.configuration_version_id
 LEFT JOIN processing_model_versions m ON m.id=c.model_version_id
 LEFT JOIN processing_price_versions p ON p.id=a.price_version_id
 WHERE a.started_at >= $2 AND a.finished_at <= $3
 AND (r.document_version_id IS NULL OR EXISTS(SELECT 1 FROM retained_resources resource
   WHERE resource.version_id=r.document_version_id AND resource.source_id=cs.source_id AND resource.configuration=cs.configuration))
 GROUP BY r.workload,r.stage,m.provider,m.model,p.currency
 ORDER BY r.workload,r.stage,m.provider,m.model,p.currency`, campaignID, started, through)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Breakdown
	for rows.Next() {
		var item Breakdown
		if err = rows.Scan(&item.Workload, &item.Stage, &item.Provider, &item.Model, &item.Currency, &item.Runs, &item.Attempts, &item.Retries,
			&item.UnknownUsageAttempts, &item.UnknownCostAttempts, &item.UnknownRetryCostAttempts, &item.DurationMS, &item.InputTokens, &item.OutputTokens,
			&item.CacheReadTokens, &item.CacheWriteTokens, &item.KnownCostMicrounits, &item.KnownRetryCostMicrounits); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func attachOtherUnits(ctx context.Context, tx pgx.Tx, campaignID string, started, through time.Time, breakdown []Breakdown) error {
	rows, err := tx.Query(ctx, `SELECT r.workload,r.stage,COALESCE(m.provider,''),COALESCE(m.model,''),COALESCE(p.currency,''),u.key,sum(u.value::bigint)::bigint
 FROM processing_runs r JOIN observation_campaign_sources cs ON cs.source_id=r.source_id AND cs.campaign_id=$1
 JOIN processing_run_attempts a ON a.run_id=r.id
 JOIN processing_configuration_versions c ON c.id=r.configuration_version_id
 LEFT JOIN processing_model_versions m ON m.id=c.model_version_id
 LEFT JOIN processing_price_versions p ON p.id=a.price_version_id
 CROSS JOIN LATERAL jsonb_each_text(a.other_units) u
 WHERE a.started_at >= $2 AND a.finished_at <= $3
 AND (r.document_version_id IS NULL OR EXISTS(SELECT 1 FROM retained_resources resource
   WHERE resource.version_id=r.document_version_id AND resource.source_id=cs.source_id AND resource.configuration=cs.configuration))
 GROUP BY r.workload,r.stage,m.provider,m.model,p.currency,u.key`, campaignID, started, through)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var workload, stage, provider, model, currency, metric string
		var quantity int64
		if err = rows.Scan(&workload, &stage, &provider, &model, &currency, &metric, &quantity); err != nil {
			return err
		}
		for index := range breakdown {
			item := &breakdown[index]
			if item.Workload == workload && item.Stage == stage && item.Provider == provider && item.Model == model && item.Currency == currency {
				if item.OtherUnits == nil {
					item.OtherUnits = map[string]int64{}
				}
				item.OtherUnits[metric] = quantity
				break
			}
		}
	}
	return rows.Err()
}

func pricingAssumptions(ctx context.Context, tx pgx.Tx, campaignID string, started, through time.Time) ([]PriceAssumption, error) {
	rows, err := tx.Query(ctx, `SELECT p.id,m.provider,m.model,p.currency,p.provenance_url,p.observed_at,p.effective_from,p.effective_through,p.details,
 min(a.started_at),max(a.started_at)
 FROM processing_price_versions p JOIN processing_model_versions m ON m.id=p.model_version_id
 JOIN processing_run_attempts a ON a.price_version_id=p.id JOIN processing_runs r ON r.id=a.run_id
 JOIN observation_campaign_sources cs ON cs.source_id=r.source_id AND cs.campaign_id=$1
 WHERE a.started_at >= $2 AND a.finished_at <= $3
 AND (r.document_version_id IS NULL OR EXISTS(SELECT 1 FROM retained_resources resource
   WHERE resource.version_id=r.document_version_id AND resource.source_id=cs.source_id AND resource.configuration=cs.configuration))
 GROUP BY p.id,m.provider,m.model,p.currency,p.provenance_url,p.observed_at,p.effective_from,p.effective_through,p.details
 ORDER BY p.id`, campaignID, started, through)
	if err != nil {
		return nil, err
	}
	var result []PriceAssumption
	for rows.Next() {
		var item PriceAssumption
		var first, last time.Time
		if err = rows.Scan(&item.ID, &item.Provider, &item.Model, &item.Currency, &item.ProvenanceURL, &item.ObservedAt, &item.EffectiveFrom, &item.EffectiveThrough, &item.Details, &first, &last); err != nil {
			return nil, err
		}
		item.ObservedAt = item.ObservedAt.UTC()
		if item.EffectiveFrom != nil {
			value := item.EffectiveFrom.UTC()
			item.EffectiveFrom = &value
		}
		if item.EffectiveThrough != nil {
			value := item.EffectiveThrough.UTC()
			item.EffectiveThrough = &value
		}
		item.VerifiedForTrial = item.EffectiveFrom != nil && item.EffectiveThrough != nil && !item.ObservedAt.After(first) && !first.Before(*item.EffectiveFrom) && last.Before(*item.EffectiveThrough)
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for index := range result {
		item := &result[index]
		rateRows, rateErr := tx.Query(ctx, "SELECT metric,unit_size,price_microunits FROM processing_price_rates WHERE price_version_id=$1 ORDER BY metric", item.ID)
		if rateErr != nil {
			return nil, rateErr
		}
		for rateRows.Next() {
			var rate PriceRate
			if rateErr = rateRows.Scan(&rate.Metric, &rate.UnitSize, &rate.PriceMicrounits); rateErr != nil {
				rateRows.Close()
				return nil, rateErr
			}
			item.Rates = append(item.Rates, rate)
		}
		rateErr = rateRows.Err()
		rateRows.Close()
		if rateErr != nil {
			return nil, rateErr
		}
	}
	return result, nil
}

func latestInputs(ctx context.Context, tx pgx.Tx, campaignID string) ([]Input, error) {
	rows, err := tx.Query(ctx, `SELECT DISTINCT ON(category,workload,component)
 category,workload,component,revision,status,COALESCE(basis,''),quantity,COALESCE(unit,''),unit_size,price_microunits,cost_microunits,
 COALESCE(currency,''),effective_from,effective_through,evidence,notes,actor,recorded_at
 FROM trial_cost_inputs WHERE campaign_id=$1 ORDER BY category,workload,component,revision DESC`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Input
	for rows.Next() {
		var item Input
		var evidence []byte
		item.CampaignID = campaignID
		if err = rows.Scan(&item.Category, &item.Workload, &item.Component, &item.Revision, &item.Status, &item.Basis, &item.Quantity, &item.Unit, &item.UnitSize,
			&item.PriceMicrounits, &item.CostMicrounits, &item.Currency, &item.EffectiveFrom, &item.EffectiveThrough, &evidence, &item.Notes, &item.Actor, &item.RecordedAt); err != nil {
			return nil, err
		}
		if json.Unmarshal(evidence, &item.Evidence) != nil {
			return nil, ErrInvalid
		}
		item.RecordedAt = item.RecordedAt.UTC()
		result = append(result, item)
	}
	return result, rows.Err()
}

func completeness(ctx context.Context, tx pgx.Tx, report Report) ([]string, []string, error) {
	missing := []string{}
	currencies := map[string]bool{}
	if !report.ObservationComplete {
		missing = append(missing, "observation_trial_not_complete")
	}
	var unfinished, unattempted int64
	err := tx.QueryRow(ctx, `SELECT
 count(*) FILTER(WHERE a.finished_at IS NULL),
 count(DISTINCT r.id) FILTER(WHERE a.run_id IS NULL)
 FROM processing_runs r JOIN observation_campaign_sources cs ON cs.source_id=r.source_id AND cs.campaign_id=$1
 LEFT JOIN processing_run_attempts a ON a.run_id=r.id AND a.started_at >= $2 AND a.started_at <= $3
 WHERE r.created_at >= $2 AND r.created_at <= $3
 AND (r.document_version_id IS NULL OR EXISTS(SELECT 1 FROM retained_resources resource
   WHERE resource.version_id=r.document_version_id AND resource.source_id=cs.source_id AND resource.configuration=cs.configuration))`, report.CampaignID, report.StartedAt, report.Through).Scan(&unfinished, &unattempted)
	if err != nil {
		return nil, nil, err
	}
	if unfinished > 0 {
		missing = append(missing, "unfinished_processing_attempts")
	}
	if unattempted > 0 {
		missing = append(missing, "processing_runs_without_attempts")
	}
	for _, item := range report.Breakdown {
		if item.UnknownUsageAttempts > 0 {
			missing = append(missing, "usage_unavailable:"+item.Workload+":"+item.Stage)
		}
		if item.UnknownCostAttempts > 0 {
			missing = append(missing, "provider_cost_unavailable:"+item.Workload+":"+item.Stage)
		}
		if item.KnownCostMicrounits != nil && item.Currency != "" {
			currencies[item.Currency] = true
		}
	}
	for _, price := range report.Prices {
		if !price.VerifiedForTrial {
			missing = append(missing, "provider_price_unverified:"+price.ID)
		}
	}
	required := map[string]bool{}
	for _, category := range []string{"hosting", "storage"} {
		for _, workload := range []string{"bootstrap", "ordinary"} {
			required[category+":"+workload] = false
		}
	}
	for _, input := range report.Infrastructure {
		key := input.Category + ":" + input.Workload
		required[key] = true
		if input.Status == "unavailable" {
			missing = append(missing, "infrastructure_cost_unavailable:"+key+":"+input.Component)
		}
		if input.Status == "reported" {
			if input.EffectiveFrom == nil || input.EffectiveThrough == nil || report.StartedAt.Before(*input.EffectiveFrom) || report.Through.After(*input.EffectiveThrough) {
				missing = append(missing, "infrastructure_price_unverified:"+key+":"+input.Component)
			}
			currencies[input.Currency] = true
		}
	}
	for key, present := range required {
		if !present {
			missing = append(missing, "infrastructure_metric_missing:"+key)
		}
	}
	if len(currencies) != 1 {
		missing = append(missing, "single_budget_currency_unavailable")
	}
	missing = uniqueSorted(missing)
	values := make([]string, 0, len(currencies))
	for currency := range currencies {
		values = append(values, currency)
	}
	slices.Sort(values)
	return missing, values, nil
}

func (s *Store) Propose(ctx context.Context, campaignID, actor string, through time.Time, evidence registry.Evidence) (Proposal, error) {
	if !validName(actor) || !validEvidence(evidence) || evidence.ObservedAt.Before(through) || evidence.ObservedAt.After(time.Now().UTC()) {
		return Proposal{}, ErrInvalid
	}
	report, err := s.Report(ctx, campaignID, through)
	if err != nil {
		return Proposal{}, err
	}
	if !report.BudgetReady || len(report.Currencies) != 1 {
		return Proposal{}, ErrIncomplete
	}
	proposal := Proposal{CampaignID: campaignID, Actor: actor, Through: through.UTC(), Currency: report.Currencies[0], TrialDurationSeconds: report.DurationSeconds,
		MonthlyProjectionSeconds: int64(MonthlyProjection.Seconds()), Evidence: evidence, Report: report}
	for _, item := range report.Breakdown {
		cost := pointerValue(item.KnownCostMicrounits)
		switch item.Workload {
		case "bootstrap":
			proposal.BootstrapCostMicrounits, err = safeAdd(proposal.BootstrapCostMicrounits, cost)
		case "ordinary":
			proposal.OrdinaryTrialCostMicrounits, err = safeAdd(proposal.OrdinaryTrialCostMicrounits, cost)
		case "evaluation":
			proposal.EvaluationCostMicrounits, err = safeAdd(proposal.EvaluationCostMicrounits, cost)
		case "reprocessing":
			proposal.ReprocessingCostMicrounits, err = safeAdd(proposal.ReprocessingCostMicrounits, cost)
		}
		if err != nil {
			return Proposal{}, ErrInvalid
		}
		proposal.RetryTrialCostMicrounits, err = safeAdd(proposal.RetryTrialCostMicrounits, pointerValue(item.KnownRetryCostMicrounits))
		if err != nil {
			return Proposal{}, ErrInvalid
		}
	}
	for _, input := range report.Infrastructure {
		if input.Status != "reported" {
			continue
		}
		if input.Workload == "bootstrap" {
			proposal.BootstrapCostMicrounits, err = safeAdd(proposal.BootstrapCostMicrounits, *input.CostMicrounits)
		} else {
			proposal.OrdinaryTrialCostMicrounits, err = safeAdd(proposal.OrdinaryTrialCostMicrounits, *input.CostMicrounits)
		}
		if err != nil {
			return Proposal{}, ErrInvalid
		}
	}
	proposal.OrdinaryMonthlyCostMicrounits, err = scaleCeil(proposal.OrdinaryTrialCostMicrounits, proposal.MonthlyProjectionSeconds, proposal.TrialDurationSeconds)
	if err != nil {
		return Proposal{}, ErrInvalid
	}
	evidenceBody, _ := json.Marshal(evidence)
	reportBody, _ := json.Marshal(report)
	err = s.pool.QueryRow(ctx, `INSERT INTO trial_budget_proposals(campaign_id,actor,through_at,currency,evidence,report,bootstrap_cost_microunits,
 ordinary_trial_cost_microunits,ordinary_monthly_cost_microunits,retry_trial_cost_microunits,evaluation_cost_microunits,reprocessing_cost_microunits)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id,recorded_at`, campaignID, actor, proposal.Through, proposal.Currency,
		evidenceBody, reportBody, proposal.BootstrapCostMicrounits, proposal.OrdinaryTrialCostMicrounits, proposal.OrdinaryMonthlyCostMicrounits,
		proposal.RetryTrialCostMicrounits, proposal.EvaluationCostMicrounits, proposal.ReprocessingCostMicrounits).Scan(&proposal.ID, &proposal.RecordedAt)
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return Proposal{}, ErrConflict
	}
	if err != nil {
		return Proposal{}, err
	}
	proposal.RecordedAt = proposal.RecordedAt.UTC()
	return proposal, nil
}

func (s *Store) Proposals(ctx context.Context, campaignID string) ([]Proposal, error) {
	if s == nil || s.pool == nil || !validName(campaignID) {
		return nil, ErrInvalid
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM observation_campaigns WHERE id=$1)", campaignID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	rows, err := s.pool.Query(ctx, `SELECT id,actor,through_at,currency,evidence,report,bootstrap_cost_microunits,ordinary_trial_cost_microunits,
 ordinary_monthly_cost_microunits,retry_trial_cost_microunits,evaluation_cost_microunits,reprocessing_cost_microunits,recorded_at
 FROM trial_budget_proposals WHERE campaign_id=$1 ORDER BY recorded_at,id`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Proposal
	for rows.Next() {
		var item Proposal
		var evidence, report []byte
		item.CampaignID, item.MonthlyProjectionSeconds = campaignID, int64(MonthlyProjection.Seconds())
		if err = rows.Scan(&item.ID, &item.Actor, &item.Through, &item.Currency, &evidence, &report, &item.BootstrapCostMicrounits,
			&item.OrdinaryTrialCostMicrounits, &item.OrdinaryMonthlyCostMicrounits, &item.RetryTrialCostMicrounits,
			&item.EvaluationCostMicrounits, &item.ReprocessingCostMicrounits, &item.RecordedAt); err != nil {
			return nil, err
		}
		if json.Unmarshal(evidence, &item.Evidence) != nil || json.Unmarshal(report, &item.Report) != nil {
			return nil, ErrInvalid
		}
		item.TrialDurationSeconds = item.Report.DurationSeconds
		item.Through, item.RecordedAt = item.Through.UTC(), item.RecordedAt.UTC()
		result = append(result, item)
	}
	return result, rows.Err()
}

func validInput(input Input) bool {
	if (input.Category != "hosting" && input.Category != "storage") || (input.Workload != "bootstrap" && input.Workload != "ordinary") || !validName(input.Component) || !validEvidence(input.Evidence) || len(input.Notes) > 4000 || input.CostMicrounits != nil {
		return false
	}
	if input.Status == "unavailable" || input.Status == "not_applicable" {
		return strings.TrimSpace(input.Notes) != "" && input.Basis == "" && input.Quantity == nil && input.Unit == "" && input.UnitSize == nil && input.PriceMicrounits == nil && input.Currency == "" && input.EffectiveFrom == nil && input.EffectiveThrough == nil
	}
	return input.Status == "reported" && (input.Basis == "measured" || input.Basis == "rate") && input.Quantity != nil && *input.Quantity >= 0 && validName(input.Unit) && input.UnitSize != nil && *input.UnitSize > 0 && input.PriceMicrounits != nil && *input.PriceMicrounits >= 0 && currencyPattern.MatchString(input.Currency) && input.EffectiveFrom != nil && input.EffectiveThrough != nil && input.EffectiveThrough.After(*input.EffectiveFrom)
}

func validEvidence(e registry.Evidence) bool {
	u, err := url.Parse(e.URL)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && u.User == nil && u.Fragment == "" && strings.TrimSpace(e.Locator) != "" && !e.ObservedAt.IsZero() && !e.ObservedAt.After(time.Now().UTC()) && (len(e.Report) == 0 || json.Valid(e.Report))
}

func validName(value string) bool {
	return value != "" && value == strings.TrimSpace(value) && len(value) <= 200 && !strings.ContainsAny(value, "\r\n")
}

func pricedCost(quantity, unitSize, price int64) (int64, bool) {
	if quantity < 0 || unitSize <= 0 || price < 0 {
		return 0, false
	}
	return bigCeil(quantity, price, unitSize)
}

func scaleCeil(value, numerator, denominator int64) (int64, error) {
	result, ok := bigCeil(value, numerator, denominator)
	if !ok {
		return 0, ErrInvalid
	}
	return result, nil
}

func bigCeil(left, right, denominator int64) (int64, bool) {
	if left < 0 || right < 0 || denominator <= 0 {
		return 0, false
	}
	value := new(big.Int).Mul(big.NewInt(left), big.NewInt(right))
	value.Add(value, big.NewInt(denominator-1))
	value.Quo(value, big.NewInt(denominator))
	return value.Int64(), value.IsInt64() && value.Sign() >= 0
}

func safeAdd(left, right int64) (int64, error) {
	if left < 0 || right < 0 || left > math.MaxInt64-right {
		return 0, ErrInvalid
	}
	return left + right, nil
}

func pointerValue(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func uniqueSorted(values []string) []string {
	slices.Sort(values)
	return slices.Compact(values)
}
