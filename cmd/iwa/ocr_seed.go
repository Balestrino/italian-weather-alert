package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/config"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
)

func seedOCR(ctx context.Context, pool *pgxpool.Pool, c config.Config) bool {
	after, e1 := strconv.ParseInt(os.Args[2], 10, 64)
	limit, e2 := strconv.Atoi(os.Args[3])
	rendererID := os.Getenv("IWA_OCR_RENDERER_ID")
	if c.Role != "admin" || e1 != nil || e2 != nil || after < 0 || limit < 1 || limit > 20 || rendererID == "" {
		slog.Error("OCR seed requires admin role, cursor, limit 1..20 and renderer identity")
		return false
	}
	sc, err := config.LoadStorage()
	if err != nil {
		slog.Error("OCR seed storage configuration invalid")
		return false
	}
	ic, err := config.LoadInference()
	if err != nil {
		slog.Error("OCR seed inference configuration invalid")
		return false
	}
	objects, err := documents.NewS3(c.RustFSURL, sc.Bucket, sc.AccessKey, sc.SecretKey)
	if err != nil {
		slog.Error("OCR seed storage unavailable")
		return false
	}
	scope, err := inference.GateScope(ic.OCR.URL, config.GateAlias("IWA_OCR"))
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	catalog, err := ocr.RegisterCatalog(ctx, processing.New(pool), ic.OCR.Adapter, ic.OCR.Model, time.Now())
	if err != nil {
		slog.Error("OCR seed catalog unavailable")
		return false
	}
	report, err := ocr.NewStore(pool).SeedHistorical(ctx, documents.New(pool, objects), ocr.PopplerRenderer{DPI: 144}, ocr.SeedOptions{AfterRun: after, Limit: limit, Scope: scope, Model: ic.OCR.Model, Configuration: catalog.ConfigurationVersionID, RendererIdentity: "png-144dpi-v1:" + rendererID})
	if json.NewEncoder(os.Stdout).Encode(report) != nil {
		return false
	}
	if err != nil {
		slog.Error("OCR seed interrupted; resume from reported cursor")
		return false
	}
	return true
}
