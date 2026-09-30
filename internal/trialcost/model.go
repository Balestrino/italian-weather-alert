// Package trialcost derives auditable provider and infrastructure costs from
// one persisted observational campaign. Missing measurements remain explicit
// and prevent an operating-budget proposal.
package trialcost

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

var (
	ErrInvalid    = errors.New("invalid trial cost input")
	ErrNotFound   = errors.New("trial cost campaign not found")
	ErrConflict   = errors.New("trial cost conflict")
	ErrIncomplete = errors.New("trial cost report is incomplete")
)

const MonthlyProjection = 30 * 24 * time.Hour

type Input struct {
	CampaignID       string            `json:"campaign_id"`
	Category         string            `json:"category"`
	Workload         string            `json:"workload"`
	Component        string            `json:"component"`
	Revision         int               `json:"revision"`
	Status           string            `json:"status"`
	Basis            string            `json:"basis,omitempty"`
	Quantity         *int64            `json:"quantity,omitempty"`
	Unit             string            `json:"unit,omitempty"`
	UnitSize         *int64            `json:"unit_size,omitempty"`
	PriceMicrounits  *int64            `json:"price_microunits,omitempty"`
	CostMicrounits   *int64            `json:"cost_microunits,omitempty"`
	Currency         string            `json:"currency,omitempty"`
	EffectiveFrom    *time.Time        `json:"effective_from,omitempty"`
	EffectiveThrough *time.Time        `json:"effective_through,omitempty"`
	Evidence         registry.Evidence `json:"evidence"`
	Notes            string            `json:"notes"`
	Actor            string            `json:"actor,omitempty"`
	RecordedAt       time.Time         `json:"recorded_at,omitempty"`
}

type Breakdown struct {
	Workload                 string           `json:"workload"`
	Stage                    string           `json:"stage"`
	Provider                 string           `json:"provider,omitempty"`
	Model                    string           `json:"model,omitempty"`
	Currency                 string           `json:"currency,omitempty"`
	Runs                     int64            `json:"runs"`
	Attempts                 int64            `json:"attempts"`
	Retries                  int64            `json:"retries"`
	UnknownUsageAttempts     int64            `json:"unknown_usage_attempts"`
	UnknownCostAttempts      int64            `json:"unknown_cost_attempts"`
	UnknownRetryCostAttempts int64            `json:"unknown_retry_cost_attempts"`
	DurationMS               *int64           `json:"duration_ms,omitempty"`
	InputTokens              *int64           `json:"input_tokens,omitempty"`
	OutputTokens             *int64           `json:"output_tokens,omitempty"`
	CacheReadTokens          *int64           `json:"cache_read_tokens,omitempty"`
	CacheWriteTokens         *int64           `json:"cache_write_tokens,omitempty"`
	KnownCostMicrounits      *int64           `json:"known_cost_microunits,omitempty"`
	KnownRetryCostMicrounits *int64           `json:"known_retry_cost_microunits,omitempty"`
	OtherUnits               map[string]int64 `json:"other_units,omitempty"`
}

type PriceRate struct {
	Metric          string `json:"metric"`
	UnitSize        int64  `json:"unit_size"`
	PriceMicrounits int64  `json:"price_microunits"`
}

type PriceAssumption struct {
	ID               string          `json:"id"`
	Provider         string          `json:"provider"`
	Model            string          `json:"model"`
	Currency         string          `json:"currency"`
	ProvenanceURL    string          `json:"provenance_url"`
	ObservedAt       time.Time       `json:"observed_at"`
	EffectiveFrom    *time.Time      `json:"effective_from,omitempty"`
	EffectiveThrough *time.Time      `json:"effective_through,omitempty"`
	Rates            []PriceRate     `json:"rates"`
	Details          json.RawMessage `json:"details"`
	VerifiedForTrial bool            `json:"verified_for_trial"`
}

type Report struct {
	CampaignID          string            `json:"campaign_id"`
	StartedAt           time.Time         `json:"started_at"`
	Through             time.Time         `json:"through"`
	DurationSeconds     int64             `json:"duration_seconds"`
	ObservationComplete bool              `json:"observation_complete"`
	Breakdown           []Breakdown       `json:"processing_breakdown"`
	Prices              []PriceAssumption `json:"pricing_assumptions"`
	Infrastructure      []Input           `json:"infrastructure_inputs"`
	MissingMetrics      []string          `json:"missing_metrics"`
	Currencies          []string          `json:"currencies"`
	BudgetReady         bool              `json:"budget_ready"`
	ObservedAt          time.Time         `json:"observed_at"`
}

type Proposal struct {
	ID                            int64             `json:"id"`
	CampaignID                    string            `json:"campaign_id"`
	Actor                         string            `json:"actor"`
	Through                       time.Time         `json:"through"`
	Currency                      string            `json:"currency"`
	TrialDurationSeconds          int64             `json:"trial_duration_seconds"`
	BootstrapCostMicrounits       int64             `json:"bootstrap_cost_microunits"`
	OrdinaryTrialCostMicrounits   int64             `json:"ordinary_trial_cost_microunits"`
	OrdinaryMonthlyCostMicrounits int64             `json:"ordinary_monthly_cost_microunits"`
	RetryTrialCostMicrounits      int64             `json:"retry_trial_cost_microunits"`
	EvaluationCostMicrounits      int64             `json:"evaluation_cost_microunits"`
	ReprocessingCostMicrounits    int64             `json:"reprocessing_cost_microunits"`
	MonthlyProjectionSeconds      int64             `json:"monthly_projection_seconds"`
	Evidence                      registry.Evidence `json:"evidence"`
	Report                        Report            `json:"report"`
	RecordedAt                    time.Time         `json:"recorded_at"`
}
