package classification

import (
	"context"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"time"
)

// RegisterLegacyCatalog preserves the immutable pre-canary processing contract.
func RegisterLegacyCatalog(ctx context.Context, store catalogStore, adapter, model string, createdAt time.Time) (Catalog, error) {
	if store == nil || adapter == "" || model == "" || createdAt.IsZero() {
		return Catalog{}, ErrInvalid
	}
	modelID := stableID("classification-model", adapter, model)
	promptID := "classification-prompt-relevance-it-v3"
	configurationID := stableID("classification-config", modelID, promptID, "segmented-full-content-operative-guard-non-thinking-v6")
	if err := store.RegisterModel(ctx, processing.ModelVersion{ID: modelID, Provider: adapter, Model: model, Revision: "runtime-contract-v1", Capabilities: json.RawMessage(`{"structured_output":true,"reasoning_separate":true}`), CreatedAt: createdAt}); err != nil {
		return Catalog{}, err
	}
	if err := store.RegisterPrompt(ctx, processing.PromptVersion{ID: promptID, Name: "toscana-relevance", Stage: "classification", Revision: "v3", Body: PromptBody, CreatedAt: createdAt}); err != nil {
		return Catalog{}, err
	}
	if err := store.RegisterConfiguration(ctx, processing.ConfigurationVersion{ID: configurationID, Name: "full-document-relevance", Stage: "classification", Revision: "v6", ModelVersionID: &modelID, PromptVersionID: &promptID, LogicVersion: "bounded-segments-any-relevant-explicit-operative-guard-non-thinking-v3", Settings: json.RawMessage(`{"max_segment_text_bytes":12288,"max_completion_tokens_per_segment":4096,"input":"all_retained_text_plus_ocr_in_deterministic_segments","document_merge":"any_relevant","source_categories_decisive":false,"qwen_enable_thinking":false,"normalization":"typographic_apostrophes_to_ascii_with_source_offsets","literal_evidence":"single_contiguous_normalized_substring","negative_override":"high_precision_explicit_local_operative_formula_only"}`), CreatedAt: createdAt}); err != nil {
		return Catalog{}, err
	}
	result := Catalog{ModelVersionID: modelID, ConfigurationVersionID: configurationID}
	if adapter != "openai-chat" || model != "qwen3.8-27b" {
		return result, nil
	}
	priceID := stableID("classification-price", modelID, "regolo-2026-09-17")
	observed := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)
	if err := store.RegisterPrice(ctx, processing.PriceVersion{ID: priceID, ModelVersionID: modelID, Currency: "EUR", ProvenanceURL: "https://regolo.ai/pricing/", ObservedAt: observed, CreatedAt: createdAt, Details: json.RawMessage(`{"billing":"tokens","input_eur_per_token":0.0000005,"output_eur_per_token":0.0000021}`), Rates: []processing.Rate{{Metric: "input_tokens", UnitSize: 1000000, PriceMicrounits: 500000}, {Metric: "output_tokens", UnitSize: 1000000, PriceMicrounits: 2100000}}}); err != nil {
		return Catalog{}, err
	}
	result.PriceVersionID = &priceID
	return result, nil
}
