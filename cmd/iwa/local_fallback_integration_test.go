//go:build integration

package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/config"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLocalFallbackCatalogsCoexistWithRemoteAndOtherLocalModels(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		t.Fatal(err)
	}
	name, password := "iwa-local-catalog-test-"+hex.EncodeToString(token[:6]), hex.EncodeToString(token)
	file := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(file, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	output, err := exec.CommandContext(ctx, "docker", "run", "--pull=never", "--detach", "--name", name, "--publish", address+":5432", "--mount", "type=bind,src="+file+",dst=/run/secrets/password,readonly", "--env", "POSTGRES_PASSWORD_FILE=/run/secrets/password", "--env", "POSTGRES_USER=iwa", "--env", "POSTGRES_DB=iwa", "postgres:16-alpine").CombinedOutput()
	if err != nil {
		t.Fatalf("isolated PostgreSQL failed: %v %s", err, strings.ReplaceAll(string(output), password, "[redacted]"))
	}
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", name).Run() })
	connection, err := pgxpool.ParseConfig("postgres://iwa@" + address + "/iwa?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	connection.ConnConfig.Password = password
	pool, err := pgxpool.NewWithConfig(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	for pool.Ping(ctx) != nil {
		if ctx.Err() != nil {
			t.Fatal("test database unavailable")
		}
		time.Sleep(100 * time.Millisecond)
	}
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	runtime := &embeddingRuntime{pool: pool, queue: jobs.New(pool)}
	configuration, runtimeErr := runtime.semanticConfiguration(ctx, 1)
	if runtimeErr != nil || configuration != "" || runtime.runner != nil {
		t.Fatal("disabled embedding initialized its provider", configuration, runtimeErr)
	}
	store := processing.New(pool)
	now := time.Now()
	if _, err := classification.RegisterCatalog(ctx, store, "openai-chat", "qwen3.8-27b", now); err != nil {
		t.Fatal(err)
	}
	if _, err := extraction.RegisterCatalog(ctx, store, "openai-chat", "qwen3.8-27b", now); err != nil {
		t.Fatal(err)
	}
	if _, err := linking.RegisterCatalog(ctx, store, "openai-chat", "qwen3.8-27b", now); err != nil {
		t.Fatal(err)
	}
	if _, err := ocr.RegisterCatalog(ctx, store, "openai-chat", "deepseek-ocr-2", now); err != nil {
		t.Fatal(err)
	}
	for _, model := range []string{"local-qwen-one", "local-qwen-two", "local-qwen-one"} {
		cfg := config.LocalFallback{Enabled: true, OCR: true, Endpoint: config.InferenceEndpoint{Adapter: "local-openai-chat", URL: "http://127.0.0.1:1/v1/chat/completions", Model: model, APIKey: "local-no-key"}, Sources: []string{"selected"}, Timeout: time.Second}
		bindings := []processing.GateBinding{}
		handlers := map[string]jobs.Handler{}
		for _, kind := range []string{classification.Kind, extraction.Kind, linking.Kind, ocr.Kind} {
			bindings = append(bindings, processing.GateBinding{Kind: kind, Scope: "remote", Model: "remote"})
			handlers[kind] = func(context.Context, jobs.Job) (jobs.Result, error) { return jobs.Result{}, nil }
		}
		result, err := addLocalFallback(ctx, store, processing.DefaultGatePolicy(), cfg, bindings, handlers, &classification.Runner{OutputFixSources: map[string]bool{"selected": true}}, &extraction.Runner{OutputFixSources: map[string]bool{"selected": true}}, &linking.Runner{}, &ocr.Runner{})
		if err != nil {
			t.Fatal("catalog alternatives cannot coexist", err)
		}
		for _, binding := range result {
			if binding.Fallback == nil || binding.Fallback.Model != model || binding.Fallback.Scope == binding.Scope {
				t.Fatal("lost independent fallback identity")
			}
		}
	}
	rows, err := pool.Query(ctx, `SELECT c.settings FROM processing_configuration_versions c JOIN processing_model_versions m ON m.id=c.model_version_id WHERE m.provider='local-openai-chat'`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count, derived := 0, 0
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var settings map[string]any
		if err := json.Unmarshal(raw, &settings); err != nil {
			t.Fatal(err)
		}
		if nested, ok := settings["fallback_settings"].(map[string]any); ok {
			if settings["local_processing"] == nil || settings["fallback_configuration"] == nil {
				t.Fatal("local processing provenance missing")
			}
			settings = nested
			derived++
		}
		if settings["temperature"] != float64(0) || settings["seed"] != float64(42) || settings["qwen_enable_thinking"] != false {
			t.Fatal("sampling provenance missing")
		}
		count++
	}
	if rows.Err() != nil || count-derived != 8 || derived != 6 {
		t.Fatal("local catalogs not independently registered", count, rows.Err())
	}
	var prices int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM processing_price_versions p JOIN processing_model_versions m ON m.id=p.model_version_id WHERE m.provider='local-openai-chat'`).Scan(&prices); err != nil || prices != 0 {
		t.Fatal("remote tariff assigned to local model", prices, err)
	}
}
