package domain

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed municipality_lifecycle_schema.sql
var municipalityLifecycleSchema string

// Only new identities are backfilled. Explicit operator choices always win.
const municipalityLifecycleBackfillSQL = `WITH identities AS(
 SELECT DISTINCT m.region_code,m.istat,EXISTS(SELECT 1 FROM territorial_current_sources a
 JOIN registry_sources s ON s.id=a.source_id WHERE a.region_code=m.region_code AND a.municipality_istat=m.istat AND s.product_id='municipal') enabled
 FROM territorial_municipalities m
), inserted AS(
 INSERT INTO territorial_municipality_state(region_code,istat,enabled,revision)
 SELECT region_code,istat,enabled,CASE WHEN enabled THEN 1 ELSE 0 END FROM identities
 ON CONFLICT DO NOTHING RETURNING *
)
INSERT INTO territorial_municipality_events(region_code,istat,revision,enabled,kind,actor)
SELECT region_code,istat,revision,true,'migration','municipality-lifecycle-migration' FROM inserted WHERE enabled`

func migrateMunicipalityLifecycle(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730021)"); err != nil {
		return err
	}
	hash := sha256.Sum256([]byte(municipalityLifecycleSchema + municipalityLifecycleBackfillSQL))
	checksum := hex.EncodeToString(hash[:])
	var previous string
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='065_municipality_lifecycle'").Scan(&previous)
	if err == nil {
		if previous != checksum {
			return errors.New("municipality lifecycle migration checksum mismatch")
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err = tx.Exec(ctx, municipalityLifecycleSchema); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, municipalityLifecycleBackfillSQL); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES('065_municipality_lifecycle',$1)", checksum); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
