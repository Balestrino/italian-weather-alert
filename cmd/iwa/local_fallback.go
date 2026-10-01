package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/config"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
)

func addLocalFallback(ctx context.Context, store *processing.Store, policy processing.GatePolicy, cfg config.LocalFallback,
	bindings []processing.GateBinding, handlers map[string]jobs.Handler,
	classifier *classification.Runner, extractor *extraction.Runner, linker *linking.Runner, reader *ocr.Runner) ([]processing.GateBinding, error) {
	if !cfg.Enabled {
		return bindings, nil
	}
	for _, source := range cfg.Sources {
		if classifier.OutputFixSources != nil && !classifier.OutputFixSources[source] || extractor.OutputFixSources != nil && !extractor.OutputFixSources[source] {
			return nil, fmt.Errorf("local fallback source %q requires evaluated output contracts", source)
		}
	}
	e := cfg.Endpoint
	transport, err := inference.NewOpenAIChat(e.Adapter, e.URL, e.APIKey, &http.Client{Timeout: cfg.Timeout})
	if err != nil {
		return nil, err
	}
	scope, err := inference.GateScope(e.URL, "local-fallback")
	if err != nil {
		return nil, err
	}
	adapter := &inference.GatedAdapter{Gate: inference.Gate{Store: store, Scope: scope, Policy: policy}, Adapter: &inference.RecordedAdapter{Adapter: transport, Ledger: store}}
	now := time.Now()
	catalog, err := classification.RegisterCatalog(ctx, store, e.Adapter, e.Model, now)
	if err != nil {
		return nil, err
	}
	localClassifier := *classifier
	localClassifier.Model, localClassifier.Adapter, localClassifier.ConfigurationVersion, localClassifier.LegacyConfigurationVersion, localClassifier.PriceVersion, localClassifier.DisableThinking = e.Model, adapter, catalog.ConfigurationVersionID, catalog.ConfigurationVersionID, catalog.PriceVersionID, true
	extractionCatalog, err := extraction.RegisterCatalog(ctx, store, e.Adapter, e.Model, now)
	if err != nil {
		return nil, err
	}
	localExtractor := *extractor
	localExtractor.Model, localExtractor.Adapter, localExtractor.ConfigurationVersion, localExtractor.LegacyConfigurationVersion, localExtractor.PriceVersion, localExtractor.DisableThinking = e.Model, adapter, extractionCatalog.ConfigurationVersionID, extractionCatalog.ConfigurationVersionID, extractionCatalog.PriceVersionID, true
	linkCatalog, err := linking.RegisterCatalog(ctx, store, e.Adapter, e.Model, now)
	if err != nil {
		return nil, err
	}
	localLinker := *linker
	localLinker.Model, localLinker.Adapter, localLinker.ConfigurationVersion, localLinker.PriceVersion, localLinker.DisableThinking = e.Model, adapter, linkCatalog.ConfigurationVersionID, linkCatalog.PriceVersionID, true
	alternatives := map[string]jobs.Handler{classification.Kind: localClassifier.Handler(), extraction.Kind: localExtractor.Handler(), linking.Kind: localLinker.Handler()}
	if cfg.OCR {
		ocrCatalog, err := ocr.RegisterCatalog(ctx, store, e.Adapter, e.Model, now)
		if err != nil {
			return nil, err
		}
		localOCR := *reader
		localOCR.DisableThinking = true
		localOCR.Model, localOCR.Adapter, localOCR.ConfigurationVersion, localOCR.PriceVersion, localOCR.ProviderScope = e.Model, adapter, ocrCatalog.ConfigurationVersionID, ocrCatalog.PriceVersionID, scope
		alternatives[ocr.Kind] = localOCR.Handler()
	}
	for index := range bindings {
		binding := &bindings[index]
		if alternative, found := alternatives[binding.Kind]; found {
			binding.Fallback = &processing.FallbackBinding{Scope: scope, Model: e.Model, Sources: cfg.Sources}
			handlers[binding.Kind] = store.FallbackHandler(handlers[binding.Kind], alternative, *binding)
		}
	}
	return bindings, nil
}
