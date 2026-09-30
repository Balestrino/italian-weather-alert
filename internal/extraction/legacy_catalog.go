package extraction

import (
	"context"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"time"
)

// RegisterLegacyCatalog preserves the immutable pre-canary processing contract.
func RegisterLegacyCatalog(ctx context.Context, store catalogStore, adapter, model string, createdAt time.Time) (Catalog, error) {
	base, err := classification.RegisterLegacyCatalog(ctx, store, adapter, model, createdAt)
	if err != nil {
		return Catalog{}, err
	}
	promptID := "extraction-prompt-compact-evidence-it-v9"
	configurationID := stableID("extraction-config", base.ModelVersionID, promptID, "explicit-temporal-candidates-quality-non-thinking-v17")
	if err = store.RegisterPrompt(ctx, processing.PromptVersion{ID: promptID, Name: "toscana-measure-extraction", Stage: "extraction", Revision: "v9", Body: PromptBody, CreatedAt: createdAt}); err != nil {
		return Catalog{}, err
	}
	if err = store.RegisterConfiguration(ctx, processing.ConfigurationVersion{ID: configurationID, Name: "evidence-backed-measure-extraction", Stage: "extraction", Revision: "v17", ModelVersionID: &base.ModelVersionID, PromptVersionID: &promptID, LogicVersion: "normalized-ocr-operational-windows-compact-evidence-v2-explicit-temporal-candidates-quality-v12", Settings: json.RawMessage(`{"envelope_version":"compact-evidence-v2","max_window_core_bytes":4096,"max_window_context_bytes_per_side":1024,"max_completion_tokens_per_window":4096,"max_response_bytes_per_window":16384,"provider_timeout_seconds":90,"input":"normalized_retained_text_plus_ocr_in_operational_windows","evidence_table":"one_per_window_zero_based_indices_unreferenced_literal_rows_ignored","temporal_candidates":"explicit_array_per_measure_with_original_expressions_and_evidence","temporal_conflicts":"shared_deterministic_identity_and_field_only_indeterminate_projection","local_temporal_completion":"same_sentence_subject_and_kind_predicate_only","activation_temporal_scope":"activation_predicate_required_in_temporal_evidence","local_kind_completion":"explicit_prohibition_suspension_activation_formulas_in_core_only","exact_duplicate_facts":"collapsed_with_distinct_evidence","unsupported_candidates":"rejected_individually_with_raw_response_retained","unsupported_evidence_refs":"ignored_per_candidate_when_other_refs_remain_literal","nearby_subject_context":"literal_within_512_bytes_of_operative_evidence","unsupported_optional_place":"projected_to_unknown_without_erasing_measure","kind_predicate":"validated_locally","indeterminate_fields":"derived_locally_from_missing_or_conflicting_candidates","operative_predicate_ownership":"core_start","cross_window_evidence":"rejected_locally","qwen_enable_thinking":false,"field_values":"literal_source_substrings_only","local_measure_times":"exclude_alert_and_page_metadata"}`), CreatedAt: createdAt}); err != nil {
		return Catalog{}, err
	}
	return Catalog{ConfigurationVersionID: configurationID, PriceVersionID: base.PriceVersionID}, nil
}
