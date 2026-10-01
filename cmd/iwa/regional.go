package main

import (
	"context"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/backend/config"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
	"strconv"
)

func projectRegionalVersion(ctx context.Context, pool *pgxpool.Pool, c config.Config) bool {
	if c.Role != "admin" {
		slog.Error("regional projection requires admin role")
		return false
	}
	id, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil || id < 1 {
		slog.Error("invalid version ID")
		return false
	}
	sc, err := config.LoadStorage()
	if err != nil {
		slog.Error("storage configuration invalid")
		return false
	}
	objects, err := documents.NewS3(c.RustFSURL, sc.Bucket, sc.AccessKey, sc.SecretKey)
	if err != nil {
		slog.Error("storage initialization failed")
		return false
	}
	report, err := domain.New(pool).ProjectCFR(ctx, documents.New(pool, objects), id)
	if err != nil {
		slog.Error("regional projection failed", "reason", err.Error())
		return false
	}
	return json.NewEncoder(os.Stdout).Encode(report) == nil
}
