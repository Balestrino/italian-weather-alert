package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// backfillEquivalences reads retained manifests only: no storage, OCR or model calls.
func backfillEquivalences(ctx context.Context, pool *pgxpool.Pool, c config.Config) bool {
	after, e1 := strconv.ParseInt(os.Args[2], 10, 64)
	limit, e2 := strconv.Atoi(os.Args[3])
	if c.Role != "admin" || e1 != nil || e2 != nil || after < 0 || limit < 1 || limit > 1000 {
		slog.Error("equivalence backfill requires admin role, cursor and limit 1..1000")
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	report, err := classification.NewStore(pool).BackfillEquivalences(ctx, after, limit)
	if json.NewEncoder(os.Stdout).Encode(report) != nil {
		return false
	}
	if err != nil {
		slog.Error("equivalence backfill interrupted; resume from reported cursor")
		return false
	}
	return true
}
