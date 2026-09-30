package ocr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/processing"
)

const PromptBody = "Free OCR."

type Catalog struct {
	ConfigurationVersionID string
	PriceVersionID         *string
}

type catalogStore interface {
	RegisterModel(context.Context, processing.ModelVersion) error
	RegisterPrompt(context.Context, processing.PromptVersion) error
	RegisterConfiguration(context.Context, processing.ConfigurationVersion) error
	RegisterPrice(context.Context, processing.PriceVersion) error
}

// RegisterCatalog makes the exact OCR model, prompt, renderer settings and
// observed public price immutable before a worker can create a run.
func RegisterCatalog(ctx context.Context, store catalogStore, adapter, model string, createdAt time.Time) (Catalog, error) {
	if store == nil || adapter == "" || model == "" || createdAt.IsZero() {
		return Catalog{}, ErrInvalid
	}
	modelID := stableID("ocr-model", adapter, model)
	promptID := "ocr-prompt-free-ocr-v1"
	configurationID := stableID("ocr-config", modelID, promptID, "poppler-144dpi-v1")
	if err := store.RegisterModel(ctx, processing.ModelVersion{ID: modelID, Provider: adapter, Model: model, Revision: "runtime-contract-v1", Capabilities: json.RawMessage(`{"input":["image/png","image/jpeg"],"raw_pdf":false}`), CreatedAt: createdAt}); err != nil {
		return Catalog{}, err
	}
	if err := store.RegisterPrompt(ctx, processing.PromptVersion{ID: promptID, Name: "deepseek-free-ocr", Stage: "ocr", Revision: "v1", Body: PromptBody, CreatedAt: createdAt}); err != nil {
		return Catalog{}, err
	}
	if err := store.RegisterConfiguration(ctx, processing.ConfigurationVersion{ID: configurationID, Name: "page-ocr", Stage: "ocr", Revision: "v1", ModelVersionID: &modelID, PromptVersionID: &promptID, LogicVersion: "page-evidence-v1", Settings: json.RawMessage(`{"max_tokens":4096,"pdf_dpi":144,"raw_pdf":false,"skip_special_tokens":false}`), CreatedAt: createdAt}); err != nil {
		return Catalog{}, err
	}
	result := Catalog{ConfigurationVersionID: configurationID}
	if adapter != "openai-chat" || model != "deepseek-ocr-2" {
		return result, nil
	}
	priceID := stableID("ocr-price", modelID, "regolo-2026-09-17")
	observed := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)
	err := store.RegisterPrice(ctx, processing.PriceVersion{
		ID: priceID, ModelVersionID: modelID, Currency: "EUR", ProvenanceURL: "https://regolo.ai/pricing/",
		ObservedAt: observed, CreatedAt: createdAt, Details: json.RawMessage(`{"billing":"per_request","price_eur":0.02}`),
		Rates: []processing.Rate{{Metric: "input_tokens", UnitSize: 1, PriceMicrounits: 0}, {Metric: "output_tokens", UnitSize: 1, PriceMicrounits: 0}, {Metric: "requests", UnitSize: 1, PriceMicrounits: 20000}},
	})
	if err != nil {
		return Catalog{}, err
	}
	result.PriceVersionID = &priceID
	return result, nil
}

func stableID(prefix string, parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		hash.Write([]byte(part))
		hash.Write([]byte{0})
	}
	return prefix + "-" + hex.EncodeToString(hash.Sum(nil))[:16]
}
