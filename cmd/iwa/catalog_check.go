package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/config"
	"github.com/Balestrino/italian-weather-alert/internal/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"time"
)

// checkInferenceCatalogs registers both rollout contracts without dispatching work.
func checkInferenceCatalogs(ctx context.Context, pool *pgxpool.Pool) error {
	settings, err := config.LoadInference()
	if err != nil {
		return err
	}
	store := processing.New(pool)
	now := time.Now().UTC()
	if _, err = ocr.RegisterCatalog(ctx, store, settings.OCR.Adapter, settings.OCR.Model, now); err != nil {
		return err
	}
	if _, err = classification.RegisterLegacyCatalog(ctx, store, settings.Qwen.Adapter, settings.Qwen.Model, now); err != nil {
		return err
	}
	if _, err = classification.RegisterCatalog(ctx, store, settings.Qwen.Adapter, settings.Qwen.Model, now); err != nil {
		return err
	}
	if _, err = extraction.RegisterLegacyCatalog(ctx, store, settings.Qwen.Adapter, settings.Qwen.Model, now); err != nil {
		return err
	}
	_, err = extraction.RegisterCatalog(ctx, store, settings.Qwen.Adapter, settings.Qwen.Model, now)
	return err
}
