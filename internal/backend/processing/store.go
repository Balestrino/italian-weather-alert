package processing

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) RegisterModel(ctx context.Context, model ModelVersion) error {
	capabilities, _, err := canonicalObject(model.Capabilities)
	if !validName(model.ID) || !validName(model.Provider) || !validName(model.Model) || !validName(model.Revision) || model.CreatedAt.IsZero() || err != nil {
		return ErrInvalid
	}
	hash := digest([]byte(model.Provider), []byte(model.Model), []byte(model.Revision), capabilities)
	tag, err := s.pool.Exec(ctx, `INSERT INTO processing_model_versions(id,provider,model,revision,capabilities,content_hash,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`, model.ID, model.Provider, model.Model, model.Revision, capabilities, hash, model.CreatedAt.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	return verifyHash(ctx, s.pool, "processing_model_versions", model.ID, hash)
}

func (s *Store) RegisterPrompt(ctx context.Context, prompt PromptVersion) error {
	if !validName(prompt.ID) || !validName(prompt.Name) || !validStages[prompt.Stage] || !validName(prompt.Revision) || prompt.Body == "" || len(prompt.Body) > maxJSONBytes || prompt.CreatedAt.IsZero() {
		return ErrInvalid
	}
	hash := digest([]byte(prompt.Name), []byte(prompt.Stage), []byte(prompt.Revision), []byte(prompt.Body))
	tag, err := s.pool.Exec(ctx, `INSERT INTO processing_prompt_versions(id,name,stage,revision,body,content_hash,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`, prompt.ID, prompt.Name, prompt.Stage, prompt.Revision, prompt.Body, hash, prompt.CreatedAt.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	return verifyHash(ctx, s.pool, "processing_prompt_versions", prompt.ID, hash)
}

func (s *Store) RegisterPrice(ctx context.Context, price PriceVersion) error {
	details, _, err := canonicalObject(price.Details)
	if !validName(price.ID) || !validName(price.ModelVersionID) || !currencyPattern.MatchString(price.Currency) || !validHTTPURL(price.ProvenanceURL) || price.ObservedAt.IsZero() || price.CreatedAt.IsZero() || len(price.Rates) == 0 || err != nil {
		return ErrInvalid
	}
	if price.EffectiveFrom != nil && price.EffectiveFrom.IsZero() || price.EffectiveThrough != nil && price.EffectiveThrough.IsZero() || price.EffectiveFrom != nil && price.EffectiveThrough != nil && !price.EffectiveThrough.After(*price.EffectiveFrom) {
		return ErrInvalid
	}
	rates := append([]Rate(nil), price.Rates...)
	sort.Slice(rates, func(i, j int) bool { return rates[i].Metric < rates[j].Metric })
	for i, rate := range rates {
		if !validName(rate.Metric) || rate.UnitSize < 1 || rate.PriceMicrounits < 0 || i > 0 && rate.Metric == rates[i-1].Metric {
			return ErrInvalid
		}
	}
	hashBody, _ := json.Marshal(struct {
		Model, Currency, URL string
		Observed             time.Time
		From, Through        *time.Time
		Details              json.RawMessage
		Rates                []Rate
	}{price.ModelVersionID, price.Currency, price.ProvenanceURL, price.ObservedAt.UTC(), utcPointer(price.EffectiveFrom), utcPointer(price.EffectiveThrough), details, rates})
	hash := digest(hashBody)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO processing_price_versions(id,model_version_id,currency,provenance_url,observed_at,effective_from,effective_through,details,content_hash,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING`, price.ID, price.ModelVersionID, price.Currency, price.ProvenanceURL, price.ObservedAt.UTC(), utcPointer(price.EffectiveFrom), utcPointer(price.EffectiveThrough), details, hash, price.CreatedAt.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		if err = verifyHash(ctx, tx, "processing_price_versions", price.ID, hash); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	for _, rate := range rates {
		if _, err = tx.Exec(ctx, `INSERT INTO processing_price_rates(price_version_id,metric,unit_size,price_microunits) VALUES($1,$2,$3,$4)`, price.ID, rate.Metric, rate.UnitSize, rate.PriceMicrounits); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) RegisterConfiguration(ctx context.Context, configuration ConfigurationVersion) error {
	settings, _, err := canonicalObject(configuration.Settings)
	if !validName(configuration.ID) || !validName(configuration.Name) || !validStages[configuration.Stage] || !validName(configuration.Revision) || !validName(configuration.LogicVersion) || configuration.CreatedAt.IsZero() || err != nil || !validOptionalName(configuration.ModelVersionID) || !validOptionalName(configuration.PromptVersionID) {
		return ErrInvalid
	}
	if configuration.PromptVersionID != nil {
		var stage string
		if err = s.pool.QueryRow(ctx, "SELECT stage FROM processing_prompt_versions WHERE id=$1", *configuration.PromptVersionID).Scan(&stage); errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		} else if err != nil {
			return err
		}
		if stage != configuration.Stage {
			return ErrInvalid
		}
	}
	hashBody, _ := json.Marshal(struct {
		Name, Stage, Revision string
		Model, Prompt         *string
		Logic                 string
		Settings              json.RawMessage
	}{configuration.Name, configuration.Stage, configuration.Revision, configuration.ModelVersionID, configuration.PromptVersionID, configuration.LogicVersion, settings})
	hash := digest(hashBody)
	tag, err := s.pool.Exec(ctx, `INSERT INTO processing_configuration_versions(id,name,stage,revision,model_version_id,prompt_version_id,logic_version,settings,content_hash,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING`, configuration.ID, configuration.Name, configuration.Stage, configuration.Revision, configuration.ModelVersionID, configuration.PromptVersionID, configuration.LogicVersion, settings, hash, configuration.CreatedAt.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	return verifyHash(ctx, s.pool, "processing_configuration_versions", configuration.ID, hash)
}

func (s *Store) StartRun(ctx context.Context, request RunRequest) (Run, error) {
	subject, _, err := canonicalObject(request.Subject)
	if !validName(request.IdempotencyKey) || !validWorkloads[request.Workload] || !validStages[request.Stage] || !validName(request.ConfigurationVersionID) || request.CreatedAt.IsZero() || err != nil || !validOptionalName(request.SourceID) || request.DocumentVersionID != nil && *request.DocumentVersionID < 1 {
		return Run{}, ErrInvalid
	}
	var configurationStage string
	if err = s.pool.QueryRow(ctx, "SELECT stage FROM processing_configuration_versions WHERE id=$1", request.ConfigurationVersionID).Scan(&configurationStage); errors.Is(err, pgx.ErrNoRows) {
		return Run{}, ErrNotFound
	} else if err != nil {
		return Run{}, err
	}
	if configurationStage != request.Stage {
		return Run{}, ErrInvalid
	}
	hashBody, _ := json.Marshal(struct {
		Workload, Stage, Configuration string
		Source                         *string
		Document                       *int64
		Subject                        json.RawMessage
	}{request.Workload, request.Stage, request.ConfigurationVersionID, request.SourceID, request.DocumentVersionID, subject})
	hash := digest(hashBody)
	var run Run
	err = s.pool.QueryRow(ctx, `INSERT INTO processing_runs(idempotency_key,request_hash,workload,stage,configuration_version_id,source_id,document_version_id,subject,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(idempotency_key) DO NOTHING
 RETURNING id,idempotency_key,workload,stage,configuration_version_id,source_id,document_version_id,subject,created_at`, request.IdempotencyKey, hash, request.Workload, request.Stage, request.ConfigurationVersionID, request.SourceID, request.DocumentVersionID, subject, request.CreatedAt.UTC()).Scan(&run.ID, &run.IdempotencyKey, &run.Workload, &run.Stage, &run.ConfigurationVersionID, &run.SourceID, &run.DocumentVersionID, &run.Subject, &run.CreatedAt)
	if err == nil {
		return run, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Run{}, err
	}
	var oldHash string
	err = s.pool.QueryRow(ctx, `SELECT id,idempotency_key,workload,stage,configuration_version_id,source_id,document_version_id,subject,created_at,request_hash
 FROM processing_runs WHERE idempotency_key=$1`, request.IdempotencyKey).Scan(&run.ID, &run.IdempotencyKey, &run.Workload, &run.Stage, &run.ConfigurationVersionID, &run.SourceID, &run.DocumentVersionID, &run.Subject, &run.CreatedAt, &oldHash)
	if err != nil {
		return Run{}, err
	}
	if oldHash != hash {
		return Run{}, ErrConflict
	}
	return run, nil
}

func (s *Store) StartAttempt(ctx context.Context, start AttemptStart) (Attempt, error) {
	if start.RunID < 1 || start.StartedAt.IsZero() || (start.QueueJobID == nil) != (start.QueueAttemptNumber == nil) || start.QueueJobID != nil && (*start.QueueJobID < 1 || *start.QueueAttemptNumber < 1) {
		return Attempt{}, ErrInvalid
	}
	if err := s.RecoverInterruptedAttempts(ctx, start.RunID, start.StartedAt); err != nil {
		return Attempt{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Attempt{}, err
	}
	defer tx.Rollback(ctx)
	var exists int
	if err = tx.QueryRow(ctx, "SELECT 1 FROM processing_runs WHERE id=$1 FOR UPDATE", start.RunID).Scan(&exists); errors.Is(err, pgx.ErrNoRows) {
		return Attempt{}, ErrNotFound
	} else if err != nil {
		return Attempt{}, err
	}
	var number int
	if err = tx.QueryRow(ctx, "SELECT COALESCE(max(number),0)+1 FROM processing_run_attempts WHERE run_id=$1", start.RunID).Scan(&number); err != nil {
		return Attempt{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO processing_run_attempts(run_id,number,queue_job_id,queue_attempt_number,started_at,outcome)
 VALUES($1,$2,$3,$4,$5,'running')`, start.RunID, number, start.QueueJobID, start.QueueAttemptNumber, start.StartedAt.UTC())
	if err != nil {
		return Attempt{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Attempt{}, err
	}
	return Attempt{RunID: start.RunID, Number: number, QueueJobID: start.QueueJobID, QueueAttemptNumber: start.QueueAttemptNumber, StartedAt: start.StartedAt.UTC(), Outcome: "running"}, nil
}

func (s *Store) FinishAttempt(ctx context.Context, finish AttemptFinish) (Attempt, error) {
	if finish.RunID < 1 || finish.Number < 1 || finish.FinishedAt.IsZero() || (finish.Outcome != "succeeded" && finish.Outcome != "failed") || finish.Outcome == "succeeded" && finish.ErrorCode != "" || finish.Outcome == "failed" && !validName(finish.ErrorCode) || !validOptionalName(finish.PriceVersionID) {
		return Attempt{}, ErrInvalid
	}
	// Receipts are authoritative for instrumented attempts. Legacy attempts keep
	// their original runner-supplied usage; the two sources are never added.
	if usage, found, err := s.CallUsage(ctx, finish.RunID, finish.Number); err != nil {
		return Attempt{}, err
	} else if found {
		finish.Usage = usage
	} else if finish.ErrorCode == "provider_cached_rejection" {
		finish.Usage = Usage{Status: "not_applicable"}
	}
	quantities, other, err := validateUsage(finish.Usage)
	if err != nil {
		return Attempt{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Attempt{}, err
	}
	defer tx.Rollback(ctx)
	var started time.Time
	var modelVersionID *string
	err = tx.QueryRow(ctx, `SELECT a.started_at,c.model_version_id FROM processing_run_attempts a
 JOIN processing_runs r ON r.id=a.run_id JOIN processing_configuration_versions c ON c.id=r.configuration_version_id
 WHERE a.run_id=$1 AND a.number=$2 AND a.finished_at IS NULL FOR UPDATE OF a`, finish.RunID, finish.Number).Scan(&started, &modelVersionID)
	if errors.Is(err, pgx.ErrNoRows) {
		var finished bool
		if e := tx.QueryRow(ctx, "SELECT finished_at IS NOT NULL FROM processing_run_attempts WHERE run_id=$1 AND number=$2", finish.RunID, finish.Number).Scan(&finished); errors.Is(e, pgx.ErrNoRows) {
			return Attempt{}, ErrNotFound
		} else if e != nil {
			return Attempt{}, e
		}
		return Attempt{}, ErrFinished
	}
	if err != nil {
		return Attempt{}, err
	}
	if finish.FinishedAt.Before(started) {
		return Attempt{}, ErrInvalid
	}
	costStatus := "unknown"
	var estimatedCost *int64
	if finish.Usage.Status == "not_applicable" {
		if finish.PriceVersionID != nil {
			return Attempt{}, ErrInvalid
		}
		costStatus = "not_applicable"
	} else if finish.PriceVersionID != nil {
		var priceModel string
		if err = tx.QueryRow(ctx, "SELECT model_version_id FROM processing_price_versions WHERE id=$1", *finish.PriceVersionID).Scan(&priceModel); errors.Is(err, pgx.ErrNoRows) {
			return Attempt{}, ErrNotFound
		} else if err != nil {
			return Attempt{}, err
		}
		if modelVersionID == nil || *modelVersionID != priceModel {
			return Attempt{}, ErrInvalid
		}
		if finish.Usage.Status == "reported" {
			rates, e := loadRates(ctx, tx, *finish.PriceVersionID)
			if e != nil {
				return Attempt{}, e
			}
			if cost, ok := calculateCost(quantities, rates); ok {
				costStatus, estimatedCost = "calculated", &cost
			}
		}
	}
	duration := finish.FinishedAt.Sub(started).Milliseconds()
	_, err = tx.Exec(ctx, `UPDATE processing_run_attempts SET finished_at=$3,duration_ms=$4,outcome=$5,usage_status=$6,
 input_tokens=$7,output_tokens=$8,cache_read_tokens=$9,cache_write_tokens=$10,other_units=$11,
 price_version_id=$12,cost_status=$13,estimated_cost_microunits=$14,error_code=NULLIF($15,'')
 WHERE run_id=$1 AND number=$2`, finish.RunID, finish.Number, finish.FinishedAt.UTC(), duration, finish.Outcome, finish.Usage.Status, finish.Usage.InputTokens, finish.Usage.OutputTokens, finish.Usage.CacheReadTokens, finish.Usage.CacheWriteTokens, other, finish.PriceVersionID, costStatus, estimatedCost, finish.ErrorCode)
	if err != nil {
		return Attempt{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Attempt{}, err
	}
	return s.Attempt(ctx, finish.RunID, finish.Number)
}

func (s *Store) Attempt(ctx context.Context, runID int64, number int) (Attempt, error) {
	var attempt Attempt
	var other []byte
	err := s.pool.QueryRow(ctx, `SELECT run_id,number,queue_job_id,queue_attempt_number,started_at,finished_at,duration_ms,outcome,
 COALESCE(usage_status,''),input_tokens,output_tokens,cache_read_tokens,cache_write_tokens,COALESCE(other_units,'{}'::jsonb),price_version_id,
 COALESCE(cost_status,''),estimated_cost_microunits,COALESCE(error_code,'') FROM processing_run_attempts WHERE run_id=$1 AND number=$2`, runID, number).Scan(&attempt.RunID, &attempt.Number, &attempt.QueueJobID, &attempt.QueueAttemptNumber, &attempt.StartedAt, &attempt.FinishedAt, &attempt.DurationMS, &attempt.Outcome, &attempt.UsageStatus, &attempt.InputTokens, &attempt.OutputTokens, &attempt.CacheReadTokens, &attempt.CacheWriteTokens, &other, &attempt.PriceVersionID, &attempt.CostStatus, &attempt.EstimatedCostMicrounits, &attempt.ErrorCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return Attempt{}, ErrNotFound
	}
	if err != nil {
		return Attempt{}, err
	}
	if json.Unmarshal(other, &attempt.OtherUnits) != nil {
		return Attempt{}, ErrInvalid
	}
	return attempt, nil
}

func (s *Store) Run(ctx context.Context, id int64) (Run, error) {
	var run Run
	err := s.pool.QueryRow(ctx, `SELECT id,idempotency_key,workload,stage,configuration_version_id,source_id,document_version_id,subject,created_at
 FROM processing_runs WHERE id=$1`, id).Scan(&run.ID, &run.IdempotencyKey, &run.Workload, &run.Stage, &run.ConfigurationVersionID, &run.SourceID, &run.DocumentVersionID, &run.Subject, &run.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, ErrNotFound
	}
	if err != nil {
		return Run{}, err
	}
	rows, err := s.pool.Query(ctx, "SELECT number FROM processing_run_attempts WHERE run_id=$1 ORDER BY number", id)
	if err != nil {
		return Run{}, err
	}
	defer rows.Close()
	var numbers []int
	for rows.Next() {
		var number int
		if err = rows.Scan(&number); err != nil {
			return Run{}, err
		}
		numbers = append(numbers, number)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return Run{}, err
	}
	rows.Close()
	for _, number := range numbers {
		attempt, e := s.Attempt(ctx, id, number)
		if e != nil {
			return Run{}, e
		}
		run.Attempts = append(run.Attempts, attempt)
	}
	return run, nil
}

func (s *Store) Summaries(ctx context.Context) ([]Summary, error) {
	rows, err := s.pool.Query(ctx, `SELECT r.workload,r.stage,COALESCE(p.currency,''),count(DISTINCT r.id),count(*),
 count(*) FILTER (WHERE a.usage_status IN ('partial','unavailable')),
 count(*) FILTER (WHERE a.cost_status='unknown'),sum(a.input_tokens),sum(a.output_tokens),sum(a.cache_read_tokens),sum(a.cache_write_tokens),sum(a.estimated_cost_microunits)
 FROM processing_run_attempts a JOIN processing_runs r ON r.id=a.run_id
 LEFT JOIN processing_price_versions p ON p.id=a.price_version_id WHERE a.finished_at IS NOT NULL
 GROUP BY r.workload,r.stage,COALESCE(p.currency,'') ORDER BY r.workload,r.stage,COALESCE(p.currency,'')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var summaries []Summary
	for rows.Next() {
		var summary Summary
		if err = rows.Scan(&summary.Workload, &summary.Stage, &summary.Currency, &summary.Runs, &summary.Attempts, &summary.UnknownUsageAttempts, &summary.UnknownCosts, &summary.InputTokens, &summary.OutputTokens, &summary.CacheReadTokens, &summary.CacheWriteTokens, &summary.EstimatedCostMicrounits); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, rows.Err()
}

type hashQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func verifyHash(ctx context.Context, query hashQuerier, table, id, expected string) error {
	allowed := map[string]bool{"processing_model_versions": true, "processing_prompt_versions": true, "processing_price_versions": true, "processing_configuration_versions": true}
	if !allowed[table] {
		return ErrInvalid
	}
	var actual string
	err := query.QueryRow(ctx, "SELECT content_hash FROM "+table+" WHERE id=$1", id).Scan(&actual)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	if actual != expected {
		return ErrConflict
	}
	return nil
}

func validOptionalName(value *string) bool { return value == nil || validName(*value) }

func utcPointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func validateUsage(usage Usage) (map[string]int64, json.RawMessage, error) {
	if usage.Status != "reported" && usage.Status != "partial" && usage.Status != "unavailable" && usage.Status != "not_applicable" {
		return nil, nil, ErrInvalid
	}
	quantities := map[string]int64{}
	fields := []struct {
		name  string
		value *int64
	}{{"input_tokens", usage.InputTokens}, {"output_tokens", usage.OutputTokens}, {"cache_read_tokens", usage.CacheReadTokens}, {"cache_write_tokens", usage.CacheWriteTokens}}
	for _, field := range fields {
		if field.value != nil {
			if *field.value < 0 {
				return nil, nil, ErrInvalid
			}
			quantities[field.name] = *field.value
		}
	}
	other := map[string]int64{}
	for metric, quantity := range usage.OtherUnits {
		if !validName(metric) || quantity < 0 || metric == "input_tokens" || metric == "output_tokens" || metric == "cache_read_tokens" || metric == "cache_write_tokens" {
			return nil, nil, ErrInvalid
		}
		other[metric], quantities[metric] = quantity, quantity
	}
	if (usage.Status == "reported" || usage.Status == "partial") && len(quantities) == 0 || (usage.Status == "unavailable" || usage.Status == "not_applicable") && len(quantities) != 0 {
		return nil, nil, ErrInvalid
	}
	body, err := json.Marshal(other)
	if err != nil {
		return nil, nil, ErrInvalid
	}
	return quantities, body, nil
}

func loadRates(ctx context.Context, tx pgx.Tx, priceID string) (map[string]Rate, error) {
	rows, err := tx.Query(ctx, "SELECT metric,unit_size,price_microunits FROM processing_price_rates WHERE price_version_id=$1", priceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rates := map[string]Rate{}
	for rows.Next() {
		var rate Rate
		if err = rows.Scan(&rate.Metric, &rate.UnitSize, &rate.PriceMicrounits); err != nil {
			return nil, err
		}
		rates[rate.Metric] = rate
	}
	return rates, rows.Err()
}

func calculateCost(quantities map[string]int64, rates map[string]Rate) (int64, bool) {
	if len(quantities) == 0 || len(quantities) != len(rates) {
		return 0, false
	}
	total := new(big.Int)
	for metric, quantity := range quantities {
		rate, ok := rates[metric]
		if !ok {
			return 0, false
		}
		line := new(big.Int).Mul(big.NewInt(quantity), big.NewInt(rate.PriceMicrounits))
		line.Add(line, big.NewInt(rate.UnitSize-1))
		line.Quo(line, big.NewInt(rate.UnitSize))
		total.Add(total, line)
	}
	if !total.IsInt64() || total.Sign() < 0 || total.Cmp(big.NewInt(math.MaxInt64)) > 0 {
		return 0, false
	}
	return total.Int64(), true
}
