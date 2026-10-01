package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/config"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/jackc/pgx/v5/pgxpool"
)

func makePreflight(ctx context.Context, pool *pgxpool.Pool, retained *documents.Store, ic config.Inference, r config.EfficiencyRollout) (*interpretation.Preflight, error) {
	if len(r.PreflightSources) == 0 {
		return nil, nil
	}
	renderer := os.Getenv("IWA_OCR_RENDERER_ID")
	enabled, _ := strconv.ParseBool(os.Getenv("IWA_OCR_REUSE_ENABLED"))
	if !enabled || renderer == "" {
		return nil, fmt.Errorf("preflight requires OCR reuse and renderer identity")
	}
	for source := range r.PreflightSources {
		if !r.OCRSources[source] || !r.InterpretationSources[source] {
			return nil, fmt.Errorf("preflight requires source-scoped stage reuse")
		}
	}
	store := processing.New(pool)
	now := time.Now()
	oc, e := ocr.RegisterCatalog(ctx, store, ic.OCR.Adapter, ic.OCR.Model, now)
	if e != nil {
		return nil, e
	}
	cl, e := classification.RegisterCatalog(ctx, store, ic.Qwen.Adapter, ic.Qwen.Model, now)
	if e != nil {
		return nil, e
	}
	lc, e := classification.RegisterLegacyCatalog(ctx, store, ic.Qwen.Adapter, ic.Qwen.Model, now)
	if e != nil {
		return nil, e
	}
	ex, e := extraction.RegisterCatalog(ctx, store, ic.Qwen.Adapter, ic.Qwen.Model, now)
	if e != nil {
		return nil, e
	}
	le, e := extraction.RegisterLegacyCatalog(ctx, store, ic.Qwen.Adapter, ic.Qwen.Model, now)
	if e != nil {
		return nil, e
	}
	oscope, e := inference.GateScope(ic.OCR.URL, config.GateAlias("IWA_OCR"))
	if e != nil {
		return nil, e
	}
	qscope, e := inference.GateScope(ic.Qwen.URL, config.GateAlias("IWA_QWEN"))
	if e != nil {
		return nil, e
	}
	wire, _ := json.Marshal([]any{interpretation.PreflightVersion, oc.ConfigurationVersionID, cl.ConfigurationVersionID, lc.ConfigurationVersionID, ex.ConfigurationVersionID, le.ConfigurationVersionID, oscope, qscope, r.OutputFixSources})
	sum := sha256.Sum256(wire)
	return &interpretation.Preflight{Documents: retained, Renderer: ocr.PopplerRenderer{DPI: 144}, RendererIdentity: "png-144dpi-v1:" + renderer, Configuration: hex.EncodeToString(sum[:]), Sources: r.PreflightSources}, nil
}

// prepareInterpretations is bounded local work: no inference adapter is constructed.
func prepareInterpretations(ctx context.Context, pool *pgxpool.Pool, c config.Config) bool {
	after, e1 := strconv.ParseInt(os.Args[2], 10, 64)
	limit, e2 := strconv.Atoi(os.Args[3])
	if c.Role != "admin" || e1 != nil || e2 != nil || after < 0 || limit < 1 || limit > 100 {
		slog.Error("preflight requires admin role, cursor and limit 1..100")
		return false
	}
	sc, e := config.LoadStorage()
	if e != nil {
		return false
	}
	ic, e := config.LoadInference()
	if e != nil {
		return false
	}
	r, e := config.LoadEfficiencyRollout()
	if e != nil {
		return false
	}
	objects, e := documents.NewS3(c.RustFSURL, sc.Bucket, sc.AccessKey, sc.SecretKey)
	if e != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	p, e := makePreflight(ctx, pool, documents.New(pool, objects), ic, r)
	if e != nil || p == nil {
		slog.Error("preflight configuration unavailable")
		return false
	}
	rows, e := pool.Query(ctx, `SELECT v.id FROM retained_versions v WHERE v.id>$1 AND NOT EXISTS(SELECT 1 FROM interpretation_archives a WHERE a.document_version_id=v.id) ORDER BY v.id LIMIT $2`, after, limit)
	if e != nil {
		return false
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if rows.Scan(&id) != nil {
			rows.Close()
			return false
		}
		ids = append(ids, id)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return false
	}
	scheduler := interpretation.New(pool, jobs.New(pool), inference.RetryPolicy{}, false)
	decisions := []interpretation.PreflightDecision{}
	cursor := after
	for _, id := range ids {
		d, err := scheduler.Prepare(ctx, p, id)
		if err != nil {
			slog.Error("local preflight failed", "version_id", id)
			return false
		}
		decisions = append(decisions, d)
		cursor = id
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"after_version": cursor, "decisions": decisions, "provider_calls": 0}) == nil
}
