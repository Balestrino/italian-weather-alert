package main

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/config"
	"github.com/Balestrino/italian-weather-alert/internal/backend/embedding"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Provider setup is lazy: a disabled embedding feature needs no embedding
// credential, catalog or network call. Workers observe the persisted switch.
type embeddingRuntime struct {
	pool      *pgxpool.Pool
	queue     *jobs.Store
	gateStore *processing.Store
	policy    processing.GatePolicy
	measures  *linking.Store
	mu        sync.Mutex
	runner    *embedding.Runner
	binding   processing.GateBinding
}

func (r *embeddingRuntime) initialize(ctx context.Context) (*embedding.Runner, processing.GateBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.runner != nil {
		return r.runner, r.binding, nil
	}
	cfg, err := config.LoadEmbedding()
	if err != nil {
		return nil, processing.GateBinding{}, err
	}
	adapter, err := inference.NewOpenAIEmbedder(cfg.Adapter, cfg.URL, cfg.APIKey, &http.Client{Timeout: 90 * time.Second})
	if err != nil {
		return nil, processing.GateBinding{}, err
	}
	scope, err := inference.GateScope(cfg.URL, config.GateAlias("IWA_EMBEDDING"))
	if err != nil {
		return nil, processing.GateBinding{}, err
	}
	catalog, err := embedding.RegisterCatalog(ctx, r.gateStore, cfg.Adapter, cfg.Model, 1024, time.Now())
	if err != nil {
		return nil, processing.GateBinding{}, err
	}
	r.binding = processing.GateBinding{Kind: embedding.Kind, Scope: scope, Model: cfg.Model}
	r.runner = &embedding.Runner{Measures: r.measures, Results: embedding.NewStore(r.pool), Processing: r.gateStore, Adapter: &inference.GatedEmbedder{Gate: inference.Gate{Store: r.gateStore, Scope: scope, Policy: r.policy}, Adapter: &inference.RecordedEmbedder{Adapter: adapter, Ledger: r.gateStore}}, Model: cfg.Model, ConfigurationVersion: catalog.ConfigurationVersionID, Dimensions: 1024}
	return r.runner, r.binding, nil
}

func (r *embeddingRuntime) handle(ctx context.Context, job jobs.Job) (jobs.Result, error) {
	runner, _, err := r.initialize(ctx)
	if err != nil {
		return jobs.Result{}, &jobs.HandlerError{Failure: jobs.Failure{Code: "embedding_configuration_unavailable", Detail: "embedding configuration unavailable", Temporary: true}}
	}
	return runner.Handler()(ctx, job)
}

func (r *embeddingRuntime) semanticConfiguration(ctx context.Context, run int64) (string, error) {
	enabled, err := r.queue.EmbeddingAllowedForRun(ctx, run)
	if err != nil || !enabled {
		return "", err
	}
	runner, _, err := r.initialize(ctx)
	if err != nil {
		return "", err
	}
	return runner.ConfigurationVersion, nil
}
