package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backoffice"
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/backups"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/config"
	"github.com/Balestrino/italian-weather-alert/internal/backend/diagnostics"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/embedding"
	"github.com/Balestrino/italian-weather-alert/internal/backend/evaluation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/health"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/notifications"
	"github.com/Balestrino/italian-weather-alert/internal/backend/observation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publiccopy"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publicquery"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publicview"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/transport/httpapi"
	"github.com/Balestrino/italian-weather-alert/internal/backend/trialcost"
	"github.com/Balestrino/italian-weather-alert/internal/backend/workerdiag"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if !run() {
		os.Exit(1)
	}
}
func run() bool {
	reportCommand := len(os.Args) == 2 && (os.Args[1] == "report-preview" || os.Args[1] == "report-status")
	projectCommand := len(os.Args) == 3 && os.Args[1] == "project-regional-version"
	zonesCommand := len(os.Args) == 4 && os.Args[1] == "adopt-toscana-zones"
	preflightCommand := len(os.Args) == 4 && os.Args[1] == "interpretation-preflight"
	pendingArchiveCommand := len(os.Args) == 4 && os.Args[1] == "documents-archive-pending"
	archiveCommand := len(os.Args) == 4 && os.Args[1] == "jobs-archive"
	seedCommand := len(os.Args) == 4 && os.Args[1] == "ocr-seed"
	previewCommand := len(os.Args) == 4 && os.Args[1] == "preview"
	backupResultCommand := len(os.Args) == 6 && os.Args[1] == "backup-result"
	retentionPolicyCommand := len(os.Args) == 3 && os.Args[1] == "retention-policy"
	if (!reportCommand && !projectCommand && !zonesCommand && !preflightCommand && !pendingArchiveCommand && !archiveCommand && !seedCommand && !previewCommand && !retentionPolicyCommand && !backupResultCommand && len(os.Args) > 2) || (len(os.Args) == 2 && !reportCommand && os.Args[1] != "check" && os.Args[1] != "inference-catalog-check" && os.Args[1] != "migrate" && os.Args[1] != "storage-init" && os.Args[1] != "storage-recover" && os.Args[1] != "retention-cleanup" && os.Args[1] != "worker" && os.Args[1] != "backup-worker") {
		slog.Error("unknown command")
		return false
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	c, err := config.Load()
	if err != nil {
		slog.Error("configuration invalid", "reason", err.Error())
		return false
	}
	// Parse only a constant: never send a credential-bearing DSN to logs.
	pc, err := pgxpool.ParseConfig("postgres://iwa@localhost:5432/iwa?sslmode=disable")
	if err != nil {
		slog.Error("database configuration invalid")
		return false
	}
	host, port, err := net.SplitHostPort(c.PostgresHost)
	if err != nil {
		slog.Error("invalid database address")
		return false
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		slog.Error("invalid database port")
		return false
	}
	pc.ConnConfig.Port = uint16(portNumber)
	pc.ConnConfig.Host = host
	pc.ConnConfig.Password = c.PostgresPassword
	pc.MaxConns = 4
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		slog.Error("database initialization failed")
		return false
	}
	defer pool.Close()
	if reportCommand {
		if c.Role != "admin" {
			slog.Error("report commands require the admin role")
			return false
		}
		reportCtx, cancel := context.WithTimeout(ctx, 40*time.Second)
		defer cancel()
		if os.Args[1] == "report-status" {
			statuses, err := notifications.New(pool).DailyReportStatuses(reportCtx)
			if err != nil {
				slog.Error("report status unavailable")
				return false
			}
			return json.NewEncoder(os.Stdout).Encode(statuses) == nil
		}
		snapshot, err := (notifications.ReportReader{Pool: pool}).Read(reportCtx, time.Now())
		if err != nil {
			slog.Error("report preview unavailable")
			return false
		}
		mail, err := notifications.RenderReport(snapshot, "Europe/Rome")
		if err != nil {
			return false
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"mail": mail, "snapshot": snapshot}) == nil
	}

	if projectCommand {
		return projectRegionalVersion(ctx, pool, c)
	}
	if zonesCommand {
		if c.Role != "admin" {
			slog.Error("zone adoption requires the admin role")
			return false
		}
		body, err := os.ReadFile(os.Args[2])
		if err != nil {
			slog.Error("zone input unavailable")
			return false
		}
		id, err := domain.New(pool).AdoptToscanaZones(ctx, body, os.Args[3])
		if err != nil {
			slog.Error("zone adoption failed", "reason", err.Error())
			return false
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]string{"zone_dataset": id}) == nil
	}

	if len(os.Args) == 2 && os.Args[1] == "migrate" {
		if c.Role != "admin" {
			slog.Error("migration requires the admin role")
			return false
		}
		migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := registry.Migrate(migrationCtx, pool); err != nil {
			slog.Error("registry migration failed")
			return false
		}
		if err := documents.Migrate(migrationCtx, pool); err != nil {
			slog.Error("documents migration failed")
			return false
		}
		if err := jobs.Migrate(migrationCtx, pool); err != nil {
			slog.Error("jobs migration failed")
			return false
		}
		if err := processing.Migrate(migrationCtx, pool); err != nil {
			slog.Error("processing migration failed")
			return false
		}
		if err := evaluation.Migrate(migrationCtx, pool); err != nil {
			slog.Error("processing comparison migration failed")
			return false
		}
		if err := acquisition.Migrate(migrationCtx, pool); err != nil {
			slog.Error("acquisition migration failed")
			return false
		}
		if err := observation.Migrate(migrationCtx, pool); err != nil {
			slog.Error("observational trial migration failed")
			return false
		}
		if err := ocr.Migrate(migrationCtx, pool); err != nil {
			slog.Error("OCR migration failed")
			return false
		}
		if err := classification.Migrate(migrationCtx, pool); err != nil {
			slog.Error("classification migration failed")
			return false
		}
		if err := extraction.Migrate(migrationCtx, pool); err != nil {
			slog.Error("extraction migration failed")
			return false
		}
		if err := diagnostics.Migrate(migrationCtx, pool); err != nil {
			slog.Error("diagnostics migration failed")
			return false
		}
		if err := linking.Migrate(migrationCtx, pool); err != nil {
			slog.Error("linking migration failed")
			return false
		}
		if err := embedding.Migrate(migrationCtx, pool); err != nil {
			slog.Error("embedding migration failed")
			return false
		}
		if err := interpretation.Migrate(migrationCtx, pool); err != nil {
			slog.Error("interpretation scheduling migration failed")
			return false
		}
		if err := domain.Migrate(migrationCtx, pool); err != nil {
			slog.Error("domain behavior migration failed")
			return false
		}
		if err := domain.MigrateGeography(migrationCtx, pool); err != nil {
			slog.Error("geography migration failed")
			return false
		}
		if err := domain.MigrateTerritories(migrationCtx, pool); err != nil {
			slog.Error("territorial migration failed")
			return false
		}
		if err := domain.New(pool).BackfillToscana(migrationCtx); err != nil {
			slog.Error("territorial backfill failed")
			return false
		}
		if err := domain.MigrateTemporal(migrationCtx, pool); err != nil {
			slog.Error("measure time migration failed")
			return false
		}
		if err := domain.MigrateQuality(migrationCtx, pool); err != nil {
			slog.Error("quality dimensions migration failed")
			return false
		}
		if err := domain.MigrateRetention(migrationCtx, pool); err != nil {
			slog.Error("retention migration failed")
			return false
		}
		if err := publicquery.Migrate(migrationCtx, pool); err != nil {
			slog.Error("public query migration failed")
			return false
		}
		if err := publicview.Migrate(migrationCtx, pool); err != nil {
			slog.Error("public view migration failed")
			return false
		}
		if err := notifications.Migrate(migrationCtx, pool); err != nil {
			slog.Error("notification migration failed")
			return false
		}
		if err := backups.Migrate(migrationCtx, pool); err != nil {
			slog.Error("backup migration failed")
			return false
		}
		if err := trialcost.Migrate(migrationCtx, pool); err != nil {
			slog.Error("trial cost migration failed")
			return false
		}
		slog.Info("storage migrations complete")
		return true
	}
	if pendingArchiveCommand {
		if c.Role != "admin" {
			slog.Error("pending document archival requires the admin role")
			return false
		}
		cutoff, parseErr := time.Parse(time.RFC3339Nano, os.Args[2])
		if parseErr != nil {
			slog.Error("pending archival cutoff must be RFC3339")
			return false
		}
		archiveCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		count, archiveErr := interpretation.New(pool, jobs.New(pool), inference.RetryPolicy{}, false).ArchivePending(archiveCtx, cutoff, os.Args[3], time.Now().UTC())
		if archiveErr != nil {
			slog.Error("pending document archival failed; retire old queue work first")
			return false
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]int64{"archived_documents": count}) == nil
	}
	if archiveCommand {
		if c.Role != "admin" {
			slog.Error("job archival requires the admin role")
			return false
		}
		cutoff, parseErr := time.Parse(time.RFC3339Nano, os.Args[2])
		if parseErr != nil {
			slog.Error("job archival cutoff must be RFC3339")
			return false
		}
		archiveCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		counts, archiveErr := jobs.New(pool).ArchiveBefore(archiveCtx, cutoff, os.Args[3], time.Now().UTC())
		if archiveErr != nil {
			slog.Error("job archival failed; running work must finish before archival")
			return false
		}
		return json.NewEncoder(os.Stdout).Encode(counts) == nil
	}
	if len(os.Args) == 2 && os.Args[1] == "inference-catalog-check" {
		if c.Role != "admin" {
			slog.Error("inference catalog check requires the admin role")
			return false
		}
		checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := checkInferenceCatalogs(checkCtx, pool); err != nil {
			slog.Error("inference catalog compatibility check failed")
			return false
		}
		slog.Info("inference catalogs compatible; no provider calls or jobs created")
		return true
	}
	if len(os.Args) == 2 && os.Args[1] == "backup-worker" {
		if c.Role != "worker" {
			slog.Error("backup worker requires the worker role")
			return false
		}
		bc, err := backups.LoadConfig(os.Getenv("IWA_BACKUPS_CONFIG_FILE"))
		if err != nil {
			slog.Error("backup configuration invalid")
			return false
		}
		w := &backups.Worker{Store: backups.New(pool), Config: bc}
		if bc.Enabled {
			if bc.Endpoint == c.RustFSURL {
				slog.Error("backup destination must be separate from source storage")
				return false
			}
			sc, err := config.LoadStorage()
			if err != nil {
				slog.Error("backup source configuration invalid")
				return false
			}
			objects, err := documents.NewS3(c.RustFSURL, sc.Bucket, sc.AccessKey, sc.SecretKey)
			if err != nil {
				slog.Error("backup source initialization failed")
				return false
			}
			destination, err := backups.NewS3Destination(bc)
			if err != nil {
				slog.Error("backup destination invalid")
				return false
			}
			w.Objects = objects
			w.Dump = backups.PGDump{Pool: pool}
			w.Destination = destination
		}
		if w.Run(ctx) != nil {
			slog.Error("backup worker failed")
			return false
		}
		return true
	}
	if preflightCommand {
		return prepareInterpretations(ctx, pool, c)
	}
	if seedCommand {
		return seedOCR(ctx, pool, c)
	}
	if backupResultCommand {
		if c.Role != "admin" {
			slog.Error("backup result requires the admin role")
			return false
		}
		at, parseErr := time.Parse(time.RFC3339Nano, os.Args[5])
		if parseErr != nil || (os.Args[4] != "succeeded" && os.Args[4] != "failed") || at.After(time.Now()) {
			slog.Error("invalid backup result")
			return false
		}
		resultCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := notifications.New(pool).RecordBackup(resultCtx, os.Args[2], os.Args[3], os.Args[4] == "succeeded", at); err != nil {
			slog.Error("backup result recording failed")
			return false
		}
		slog.Info("backup result recorded")
		return true
	}
	if retentionPolicyCommand {
		if c.Role != "admin" {
			slog.Error("retention policy change requires the admin role")
			return false
		}
		months, parseErr := strconv.Atoi(os.Args[2])
		if parseErr != nil {
			slog.Error("invalid retention duration")
			return false
		}
		policy, policyErr := domain.NewRetention(pool, nil).SetPolicy(ctx, months, "cli-admin", time.Now())
		if policyErr != nil {
			slog.Error("retention policy change failed")
			return false
		}
		slog.Info("retention policy changed", "version", policy.ID, "months", policy.Months)
		return true
	}
	if len(os.Args) == 2 && os.Args[1] == "retention-cleanup" {
		if c.Role != "admin" {
			slog.Error("retention cleanup requires the admin role")
			return false
		}
		sc, loadErr := config.LoadStorage()
		if loadErr != nil {
			slog.Error("storage configuration invalid")
			return false
		}
		objects, storageErr := documents.NewS3(c.RustFSURL, sc.Bucket, sc.AccessKey, sc.SecretKey)
		if storageErr != nil {
			slog.Error("storage initialization failed")
			return false
		}
		cleanupCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
		defer cancel()
		result, cleanupErr := domain.NewRetention(pool, objects).Cleanup(cleanupCtx, time.Now())
		if cleanupErr != nil {
			slog.Error("retention cleanup failed")
			return false
		}
		slog.Info("retention cleanup complete", "run", result.RunID, "deleted_versions", result.DeletedVersions, "deleted_objects", result.DeletedObjects, "failed_objects", result.FailedObjects, "pending_objects", result.PendingObjects)
		return result.FailedObjects == 0
	}
	if len(os.Args) == 2 && (os.Args[1] == "storage-init" || os.Args[1] == "storage-recover") {
		if c.Role != "admin" {
			slog.Error("storage operation requires the admin role")
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
		opCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		if os.Args[1] == "storage-init" {
			if err = objects.Initialize(opCtx); err != nil {
				slog.Error("storage initialization failed")
				return false
			}
			slog.Info("private object bucket ready")
			return true
		}
		store := documents.New(pool, objects)
		ids, err := store.Pending(opCtx, 100)
		if err != nil {
			slog.Error("pending acquisitions unavailable")
			return false
		}
		failed := 0
		for _, id := range ids {
			if _, err = store.Recover(opCtx, id); err != nil {
				failed++
			}
		}
		slog.Info("storage recovery finished", "examined", len(ids), "failed", failed)
		return failed == 0
	}
	if previewCommand {
		if c.Role != "admin" {
			slog.Error("acquisition preview requires the admin role")
			return false
		}
		revision, parseErr := strconv.Atoi(os.Args[3])
		if parseErr != nil || revision < 1 {
			slog.Error("invalid configuration revision")
			return false
		}
		sc, loadErr := config.LoadStorage()
		if loadErr != nil {
			slog.Error("storage configuration invalid")
			return false
		}
		objects, storageErr := documents.NewS3(c.RustFSURL, sc.Bucket, sc.AccessKey, sc.SecretKey)
		if storageErr != nil {
			slog.Error("storage initialization failed")
			return false
		}
		engine := acquisition.Engine{
			Registry:  registry.New(pool),
			Retained:  documents.New(pool, objects),
			Crawler:   &acquisition.Crawl4AI{BaseURL: c.CrawlURL, Token: c.CrawlToken, Client: &http.Client{Timeout: 90 * time.Second}},
			Resources: &acquisition.DirectHTTP{Client: &http.Client{Timeout: 90 * time.Second}},
		}
		previewCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
		defer cancel()
		report, previewErr := engine.Preview(previewCtx, os.Args[2], revision, "cli-preview")
		if previewErr != nil {
			slog.Error("acquisition preview failed", "reason", previewErr.Error())
			return false
		}
		if json.NewEncoder(os.Stdout).Encode(report) != nil {
			slog.Error("acquisition preview output failed")
			return false
		}
		return true
	}
	if len(os.Args) == 2 && os.Args[1] == "worker" {
		if c.Role != "worker" {
			slog.Error("worker command requires the worker role")
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
		retained := documents.New(pool, objects)
		queue := jobs.New(pool)
		if _, err = jobs.EnqueuePendingDocuments(ctx, queue, retained, time.Now()); err != nil {
			slog.Error("pending acquisition reconciliation failed")
			return false
		}
		hostname, _ := os.Hostname()
		workerID := hostname + ":" + strconv.Itoa(os.Getpid())
		documentWorker := &jobs.Worker{Store: queue, Queue: jobs.DocumentQueue, ID: "documents-" + workerID, Lease: time.Minute, PollInterval: 500 * time.Millisecond, Handlers: map[string]jobs.Handler{jobs.DocumentKind: jobs.DocumentHandler(retained)}}
		inferenceConfig, err := config.LoadInference()
		if err != nil {
			slog.Error("inference configuration invalid")
			return false
		}
		gatePolicy, err := config.LoadProviderGatePolicy()
		if err != nil {
			slog.Error("provider gate configuration invalid")
			return false
		}
		rollout, err := config.LoadEfficiencyRollout()
		if err != nil {
			slog.Error("efficiency rollout configuration invalid")
			return false
		}
		gateStore := processing.New(pool)
		gateBindings := []processing.GateBinding{}
		ocrConfig := inferenceConfig.OCR
		ocrAdapter, err := inference.NewOpenAIChat(ocrConfig.Adapter, ocrConfig.URL, ocrConfig.APIKey, &http.Client{Timeout: 90 * time.Second})
		if err != nil {
			slog.Error("OCR adapter configuration invalid")
			return false
		}
		ocrAdapterGateScope, scopeErr := inference.GateScope(ocrConfig.URL, config.GateAlias("IWA_OCR"))
		if scopeErr != nil {
			slog.Error("provider gate scope invalid")
			return false
		}
		ocrAdapterGate := inference.Gate{Store: gateStore, Scope: ocrAdapterGateScope, Policy: gatePolicy}
		reuseOCR := false
		if value := os.Getenv("IWA_OCR_REUSE_ENABLED"); value != "" {
			reuseOCR, err = strconv.ParseBool(value)
			if err != nil {
				slog.Error("OCR reuse switch invalid")
				return false
			}
		}
		if reuseOCR && len(rollout.OCRSources) == 0 {
			slog.Error("OCR reuse requires an explicit source list")
			return false
		}
		// Keep account holds before claiming even when reuse is enabled: a cold
		// cache must not turn a hold into repeated local attempts.
		gateBindings = append(gateBindings, processing.GateBinding{Kind: ocr.Kind, Scope: ocrAdapterGateScope, Model: ocrConfig.Model})

		ocrCatalog, err := ocr.RegisterCatalog(ctx, processing.New(pool), ocrConfig.Adapter, ocrConfig.Model, time.Now())
		if err != nil {
			slog.Error("OCR catalog registration failed")
			return false
		}
		ocrRunner := &ocr.Runner{ReuseSources: rollout.OCRSources, Documents: retained, Processing: processing.New(pool), Results: ocr.NewStore(pool), Adapter: &inference.GatedAdapter{Gate: ocrAdapterGate, Adapter: &inference.RecordedAdapter{Adapter: ocrAdapter, Ledger: gateStore}}, Renderer: ocr.PopplerRenderer{DPI: 144}, Model: ocrConfig.Model, ConfigurationVersion: ocrCatalog.ConfigurationVersionID, PriceVersion: ocrCatalog.PriceVersionID}
		if reuseOCR {
			rendererID := os.Getenv("IWA_OCR_RENDERER_ID")
			if rendererID == "" {
				slog.Error("OCR reuse requires immutable renderer identity")
				return false
			}
			ocrRunner.Artifacts = ocr.NewStore(pool)
			ocrRunner.ProviderScope = ocrAdapterGateScope
			ocrRunner.RendererIdentity = "png-144dpi-v1:" + rendererID
		}

		qwenConfig := inferenceConfig.Qwen
		qwenAdapter, err := inference.NewOpenAIChat(qwenConfig.Adapter, qwenConfig.URL, qwenConfig.APIKey, &http.Client{Timeout: 90 * time.Second})
		if err != nil {
			slog.Error("Qwen adapter configuration invalid")
			return false
		}
		qwenAdapterGateScope, scopeErr := inference.GateScope(qwenConfig.URL, config.GateAlias("IWA_QWEN"))
		if scopeErr != nil {
			slog.Error("provider gate scope invalid")
			return false
		}
		qwenAdapterGate := inference.Gate{Store: gateStore, Scope: qwenAdapterGateScope, Policy: gatePolicy}
		gateBindings = append(gateBindings, processing.GateBinding{Kind: classification.Kind, Scope: qwenAdapterGateScope, Model: qwenConfig.Model})
		gateBindings = append(gateBindings, processing.GateBinding{Kind: extraction.Kind, Scope: qwenAdapterGateScope, Model: qwenConfig.Model})
		gateBindings = append(gateBindings, processing.GateBinding{Kind: linking.Kind, Scope: qwenAdapterGateScope, Model: qwenConfig.Model})

		classificationCatalog, err := classification.RegisterCatalog(ctx, processing.New(pool), qwenConfig.Adapter, qwenConfig.Model, time.Now())
		if err != nil {
			slog.Error("classification catalog registration failed")
			return false
		}
		legacyClassification, err := classification.RegisterLegacyCatalog(ctx, processing.New(pool), qwenConfig.Adapter, qwenConfig.Model, time.Now())
		if err != nil {
			slog.Error("legacy classification catalog registration failed")
			return false
		}
		reuseSources := rollout.InterpretationSources
		classificationRunner := &classification.Runner{OutputFixSources: rollout.OutputFixSources, CheckpointSources: rollout.CheckpointSources, LegacyConfigurationVersion: legacyClassification.ConfigurationVersionID, Checkpoints: processing.New(pool), Manifests: classification.NewStore(pool), ReuseSources: reuseSources, Documents: retained, OCR: ocr.NewStore(pool), Processing: processing.New(pool), Results: classification.NewStore(pool), Adapter: &inference.GatedAdapter{Gate: qwenAdapterGate, Adapter: &inference.RecordedAdapter{Adapter: qwenAdapter, Ledger: gateStore}}, Model: qwenConfig.Model, ConfigurationVersion: classificationCatalog.ConfigurationVersionID, PriceVersion: classificationCatalog.PriceVersionID, DisableThinking: qwenConfig.Adapter == "openai-chat" && qwenConfig.Model == "qwen3.8-27b"}
		extractionCatalog, err := extraction.RegisterCatalog(ctx, processing.New(pool), qwenConfig.Adapter, qwenConfig.Model, time.Now())
		if err != nil {
			slog.Error("extraction catalog registration failed")
			return false
		}
		legacyExtraction, err := extraction.RegisterLegacyCatalog(ctx, processing.New(pool), qwenConfig.Adapter, qwenConfig.Model, time.Now())
		if err != nil {
			slog.Error("legacy extraction catalog registration failed")
			return false
		}
		extractionRunner := &extraction.Runner{OutputFixSources: rollout.OutputFixSources, CheckpointSources: rollout.CheckpointSources, LegacyConfigurationVersion: legacyExtraction.ConfigurationVersionID, Checkpoints: processing.New(pool), Manifests: classification.NewStore(pool), ReuseSources: reuseSources, Documents: retained, OCR: ocr.NewStore(pool), Classifications: classification.NewStore(pool), Processing: processing.New(pool), Results: extraction.NewStore(pool), Adapter: &inference.GatedAdapter{Gate: qwenAdapterGate, Adapter: &inference.RecordedAdapter{Adapter: qwenAdapter, Ledger: gateStore}}, Model: qwenConfig.Model, ConfigurationVersion: extractionCatalog.ConfigurationVersionID, PriceVersion: extractionCatalog.PriceVersionID, DisableThinking: qwenConfig.Adapter == "openai-chat" && qwenConfig.Model == "qwen3.8-27b"}
		linkingCatalog, err := linking.RegisterCatalog(ctx, processing.New(pool), qwenConfig.Adapter, qwenConfig.Model, time.Now())
		if err != nil {
			slog.Error("linking catalog registration failed")
			return false
		}
		linkingStore := linking.NewStore(pool)
		linkingRunner := &linking.Runner{Store: linkingStore, Processing: processing.New(pool), Adapter: &inference.GatedAdapter{Gate: qwenAdapterGate, Adapter: &inference.RecordedAdapter{Adapter: qwenAdapter, Ledger: gateStore}}, Model: qwenConfig.Model, ConfigurationVersion: linkingCatalog.ConfigurationVersionID, PriceVersion: linkingCatalog.PriceVersionID, MaxCandidates: 20, DisableThinking: qwenConfig.Adapter == "openai-chat" && qwenConfig.Model == "qwen3.8-27b"}
		handlers := map[string]jobs.Handler{ocr.Kind: ocrRunner.Handler(), classification.Kind: classificationRunner.Handler(), extraction.Kind: extractionRunner.Handler(), linking.Kind: linkingRunner.Handler()}
		semanticEnabled, err := config.SemanticEnabled()
		if err != nil {
			slog.Error("semantic linking configuration invalid")
			return false
		}
		if semanticEnabled {
			embeddingConfig, loadErr := config.LoadEmbedding()
			if loadErr != nil {
				slog.Error("embedding configuration invalid")
				return false
			}
			embedder, adapterErr := inference.NewOpenAIEmbedder(embeddingConfig.Adapter, embeddingConfig.URL, embeddingConfig.APIKey, &http.Client{Timeout: 90 * time.Second})
			if adapterErr != nil {
				slog.Error("embedding adapter configuration invalid")
				return false
			}
			embedScope, scopeErr := inference.GateScope(embeddingConfig.URL, config.GateAlias("IWA_EMBEDDING"))
			if scopeErr != nil {
				slog.Error("embedding gate scope invalid")
				return false
			}
			embedGate := inference.Gate{Store: gateStore, Scope: embedScope, Policy: gatePolicy}
			gateBindings = append(gateBindings, processing.GateBinding{Kind: embedding.Kind, Scope: embedScope, Model: embeddingConfig.Model})

			embeddingCatalog, catalogErr := embedding.RegisterCatalog(ctx, processing.New(pool), embeddingConfig.Adapter, embeddingConfig.Model, 1024, time.Now())
			if catalogErr != nil {
				slog.Error("embedding catalog registration failed")
				return false
			}
			embeddingRunner := &embedding.Runner{Measures: linkingStore, Results: embedding.NewStore(pool), Processing: processing.New(pool), Adapter: &inference.GatedEmbedder{Gate: embedGate, Adapter: &inference.RecordedEmbedder{Adapter: embedder, Ledger: gateStore}}, Model: embeddingConfig.Model, ConfigurationVersion: embeddingCatalog.ConfigurationVersionID, Dimensions: 1024}
			handlers[embedding.Kind] = embeddingRunner.Handler()
			linkingRunner.SemanticConfiguration = embeddingCatalog.ConfigurationVersionID
		}
		localFallback, fallbackErr := config.LoadLocalFallback()
		if fallbackErr != nil {
			slog.Error("local fallback configuration invalid")
			return false
		}
		gateBindings, err = addLocalFallback(ctx, gateStore, gatePolicy, localFallback, gateBindings, handlers, classificationRunner, extractionRunner, linkingRunner, ocrRunner)
		if err != nil {
			slog.Error("local fallback initialization failed")
			return false
		}
		interpretationScheduler := interpretation.New(pool, queue, inference.RetryPolicy{MaxAttempts: inferenceConfig.MaxAttempts, BaseDelay: inferenceConfig.RetryBase}, semanticEnabled)
		interpretationScheduler.Preflight, err = makePreflight(ctx, pool, retained, inferenceConfig, rollout)
		if err != nil {
			slog.Error("preflight configuration invalid")
			return false
		}
		handlers[ocr.Kind] = interpretationScheduler.OCRHandler(handlers[ocr.Kind])
		handlers[classification.Kind] = interpretationScheduler.ClassificationHandler(handlers[classification.Kind])
		handlers[extraction.Kind] = interpretationScheduler.ExtractionHandler(handlers[extraction.Kind])
		if semanticEnabled {
			handlers[embedding.Kind] = interpretationScheduler.EmbeddingHandler(handlers[embedding.Kind])
		}
		for kind, handler := range handlers {
			handlers[kind] = interpretationScheduler.Guard(handler)
		}
		inferenceWorker := &jobs.Worker{Store: queue, Queue: inference.Queue, ID: "inference-" + workerID, Lease: 20 * time.Minute, PollInterval: 500 * time.Millisecond, Handlers: handlers}
		var nextEquivalentRelease time.Time
		inferenceWorker.BeforeClaim = func(c context.Context, at time.Time) error {
			if err := gateStore.DeferRecoveryJobs(c, at); err != nil {
				return err
			}
			if !at.Before(nextEquivalentRelease) {
				if err := interpretationScheduler.ReleaseEquivalent(c, at); err != nil {
					return err
				}
				nextEquivalentRelease = at.Add(30 * time.Second)
			}
			if err := interpretationScheduler.DeferEquivalent(c, at); err != nil {
				return err
			}
			return gateStore.DeferHeldJobs(c, gateBindings, at)
		}
		checkEngine := &acquisition.Engine{Registry: registry.New(pool), Retained: retained, Crawler: &acquisition.Crawl4AI{BaseURL: c.CrawlURL, Token: c.CrawlToken, Client: &http.Client{Timeout: 90 * time.Second}}, Resources: &acquisition.DirectHTTP{Client: &http.Client{Timeout: 90 * time.Second}}, Tracking: acquisition.NewTrackingStore(pool)}
		checkWorker := &acquisition.CheckWorker{Store: acquisition.NewScheduleStore(pool), Engine: checkEngine, ID: "acquisition-" + workerID, Lease: 20 * time.Minute, PollInterval: time.Second, Schedule: func(scheduleCtx context.Context, page acquisition.RetainedPage, at time.Time) error {
			if _, err := domain.New(pool).ProjectCFR(scheduleCtx, retained, page.VersionID); err != nil {
				return err
			}
			_, scheduleErr := interpretationScheduler.Automatic(scheduleCtx, interpretation.AcquisitionEvent{DocumentVersionID: page.VersionID, EvidenceHash: page.Hash, Workload: page.Workload, ContentChanged: !page.Changed, DependencyChanged: page.Changed, At: at})
			return scheduleErr
		}}
		notificationConfig, err := notifications.LoadConfig(os.Getenv("IWA_NOTIFICATIONS_CONFIG_FILE"))
		if err != nil {
			slog.Error("notification configuration invalid")
			return false
		}
		var notificationWorker *notifications.Worker
		var dailyReportWorker *notifications.DailyReportWorker
		if notificationConfig.Enabled {
			sender, senderErr := notifications.NewSMTP(notificationConfig)
			if senderErr != nil {
				slog.Error("notification SMTP configuration invalid")
				return false
			}
			notificationWorker = &notifications.Worker{Store: notifications.New(pool), Sender: sender, Config: notificationConfig}
			if notificationConfig.DailyReport != nil && notificationConfig.DailyReport.Enabled {
				dailyReportWorker = &notifications.DailyReportWorker{Store: notifications.New(pool), Reader: notifications.ReportReader{Pool: pool}, Sender: sender, Config: notificationConfig}
				slog.Info("daily report enabled", "local_time", notificationConfig.DailyReport.LocalTime, "timezone", notificationConfig.DailyReport.Timezone)
			}
		}
		if err := processing.New(pool).RecoverInterruptedAttempts(ctx, 0, time.Now()); err != nil {
			slog.Error("interrupted processing recovery unavailable")
			return false
		}
		workerCtx, workerCancel := context.WithCancel(ctx)
		workerCount := 3
		if dailyReportWorker != nil {
			workerCount++
		}
		if notificationWorker != nil {
			workerCount++
		}
		type workerOutcome struct {
			component string
			err       error
		}
		workerErrors := make(chan workerOutcome, workerCount)
		startWorker := func(name string, run func(context.Context) error) {
			go func() { workerErrors <- workerOutcome{name, run(workerCtx)} }()
		}
		report := func(out workerOutcome) {
			op, category, state := workerdiag.Describe(out.err)
			if errors.Is(out.err, jobs.ErrStaleClaim) || errors.Is(out.err, acquisition.ErrStaleCheck) {
				category = "lease_lost"
			}
			slog.Info("worker outcome", "component", out.component, "operation", op, "category", category, "sqlstate", state, "shutdown_requested", workerCtx.Err() != nil)
		}
		if notificationWorker != nil {
			startWorker("notifications", notificationWorker.Run)
		}
		if dailyReportWorker != nil {
			startWorker("daily-report", dailyReportWorker.Run)
		}
		startWorker("documents", documentWorker.Run)
		startWorker("acquisition", checkWorker.Run)
		startWorker("inference", inferenceWorker.Run)
		slog.Info("workers started", "queues", "documents,acquisition,inference")
		first := <-workerErrors
		report(first)
		err = first.err
		workerCancel()
		shutdownTimer := time.NewTimer(15 * time.Second)
		defer shutdownTimer.Stop()
		for i := 1; i < workerCount; i++ {
			var next workerOutcome
			select {
			case next = <-workerErrors:
			case <-shutdownTimer.C:
				slog.Error("worker shutdown timed out", "category", "shutdown_timeout")
				return false
			}
			report(next)
			if err == nil {
				err = next.err
			}
		}
		if err != nil {
			slog.Error("worker stopped unexpectedly")
			return false
		}
		slog.Info("workers stopped")
		return true
	}
	checks := map[string]health.Check{"postgres": pool.Ping, "rustfs": health.HTTP(c.RustFSURL+"/health/ready", ""), "crawl4ai": health.HTTP(c.CrawlURL+"/health", c.CrawlToken)}
	if len(os.Args) == 2 && os.Args[1] == "check" {
		host, port, err := net.SplitHostPort(c.Listen)
		if err != nil {
			return false
		}
		if host == "" || host == "0.0.0.0" || host == "::" {
			host = "127.0.0.1"
		}
		probeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
		return health.HTTP("http://"+net.JoinHostPort(host, port)+"/health/ready", "")(probeCtx) == nil
	}

	ln, err := net.Listen("tcp", c.Listen)
	if err != nil {
		slog.Error("listener unavailable")
		return false
	}
	slog.Info("service started", "role", c.Role)
	publicLimits := httpapi.PublicLimits{Allowance: c.PublicAllowance, Window: c.PublicWindow, MaxPageSize: c.PublicMaxPageSize, TrustedProxies: c.PublicTrustedProxies}
	var copyDocuments *documents.Store
	if c.Role == "public" && c.PublicCopyAccess {
		storage, loadErr := config.LoadStorage()
		if loadErr != nil {
			slog.Error("public copy storage configuration invalid")
			return false
		}
		objects, storageErr := documents.NewS3(c.RustFSURL, storage.Bucket, storage.AccessKey, storage.SecretKey)
		if storageErr != nil {
			slog.Error("public copy storage initialization failed")
			return false
		}
		copyDocuments = documents.New(pool, objects)
	}
	copyAccess, err := publiccopy.New(pool, copyDocuments, c.PublicCopyAccess, c.PublicBaseURL)
	if err != nil {
		slog.Error("public copy access initialization failed")
		return false
	}
	publicRuntime := httpapi.PublicRuntime{Limits: publicLimits, Views: publicview.New(pool), ViewLifetime: c.PublicViewLifetime, CursorKey: []byte(c.PublicCursorKey), Copies: copyAccess}
	var adminRuntime backoffice.AdminRuntime
	if c.Role == "admin" {
		storage, loadErr := config.LoadStorage()
		if loadErr != nil {
			slog.Error("admin storage configuration invalid")
			return false
		}
		objects, storageErr := documents.NewS3(c.RustFSURL, storage.Bucket, storage.AccessKey, storage.SecretKey)
		if storageErr != nil {
			slog.Error("admin storage initialization failed")
			return false
		}
		reg := registry.New(pool)
		engine := &acquisition.Engine{Registry: reg, Retained: documents.New(pool, objects),
			Crawler:   &acquisition.Crawl4AI{BaseURL: c.CrawlURL, Token: c.CrawlToken, Client: &http.Client{Timeout: 90 * time.Second}},
			Resources: &acquisition.DirectHTTP{Client: &http.Client{Timeout: 90 * time.Second}}}
		evaluations := evaluation.New(pool)
		adminScheduler := interpretation.New(pool, jobs.New(pool), inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, false)
		adminScheduler.Evaluations = evaluations
		adminRuntime = backoffice.AdminRuntime{Territories: operations.New(pool), TerritoryConfig: domain.New(pool), Alerts: operations.New(pool), Provider: processing.New(pool), TailscaleOrigin: c.AdminTailscaleOrigin, Backups: backups.New(pool), Notifications: notifications.New(pool), Operations: operations.New(pool), Diagnostics: diagnostics.New(pool), Jobs: jobs.New(pool), Registry: reg, Preview: engine, Interpretation: adminScheduler, Evaluations: evaluations, Observations: observation.New(pool), TrialCosts: trialcost.New(pool)}
	}
	handler := listenerHandler(c.Role, checks, publicquery.New(pool), publicRuntime, adminRuntime)
	if err := httpserver.Serve(ctx, ln, handler); err != nil {
		slog.Error("service stopped unexpectedly")
		return false
	}
	slog.Info("service stopped", "role", c.Role)
	return true
}
