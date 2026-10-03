//go:build integration

package processing

import (
	"context"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
	"time"
)

func TestLocalProcessingPreservesCatalogsAcrossFallbackRevisions(t *testing.T) {
	ctx := context.Background()
	pool := processingTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	store := New(pool)
	now := time.Now().UTC()
	first := ConfigurationVersion{ID: "fallback-one", Name: "municipal-extraction", Stage: "extraction", Revision: "v1", LogicVersion: "fixture-v1", Settings: json.RawMessage(`{"fixture":1}`), CreatedAt: now}
	second := first
	second.ID = "fallback-two"
	second.Revision = "v2"
	second.Settings = json.RawMessage(`{"fixture":2}`)
	for _, c := range []ConfigurationVersion{first, second} {
		if err := store.RegisterConfiguration(ctx, c); err != nil {
			t.Fatal(err)
		}
	}
	old, err := store.RegisterLocalProcessing(ctx, first.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	var originalHash string
	if err = pool.QueryRow(ctx, `SELECT content_hash FROM processing_configuration_versions WHERE id=$1`, old).Scan(&originalHash); err != nil {
		t.Fatal(err)
	}
	updated, err := store.RegisterLocalProcessing(ctx, second.ID, now)
	if err != nil || old == updated {
		t.Fatal("fallback revision failed to obtain independent local configuration", updated, err)
	}
	again, err := store.RegisterLocalProcessing(ctx, second.ID, now.Add(time.Hour))
	if err != nil || again != updated {
		t.Fatal("updated registration is not idempotent", again, err)
	}
	legacy, err := store.RegisterLocalProcessing(ctx, first.ID, now.Add(time.Hour))
	if err != nil || legacy != old {
		t.Fatal("legacy registration changed", legacy, err)
	}
	var retainedHash, fallback string
	if err = pool.QueryRow(ctx, `SELECT content_hash FROM processing_configuration_versions WHERE id=$1`, old).Scan(&retainedHash); err != nil || retainedHash != originalHash {
		t.Fatal("prior configuration rewritten", err)
	}
	if err = pool.QueryRow(ctx, `SELECT settings->>'fallback_configuration' FROM processing_configuration_versions WHERE id=$1`, updated).Scan(&fallback); err != nil || fallback != second.ID {
		t.Fatal("wrong fallback association", fallback, err)
	}
}
