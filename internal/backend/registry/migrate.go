package registry

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

//go:embed controls_schema.sql
var controlsSchema string

//go:embed interpretation_schema.sql
var interpretationSchema string

//go:embed regression_schema.sql
var regressionSchema string

//go:embed release_schema.sql
var releaseSchema string

//go:embed public_scope_schema.sql
var publicScopeSchema string

// Migrate applies this module's schemas atomically, serializing concurrent
// invocations. Changed historical SQL is rejected instead of silently accepted.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730021)"); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS iwa_migrations (name text PRIMARY KEY, checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	for _, migration := range []struct{ name, sql string }{{"001_registry", schema}, {"021_source_controls", controlsSchema}, {"023_interpretation_controls", interpretationSchema}, {"027_regression", regressionSchema}, {"032_release_acceptance", releaseSchema}, {"053_public_territorial_scope", publicScopeSchema}} {
		hash := sha256.Sum256([]byte(migration.sql))
		checksum := hex.EncodeToString(hash[:])
		var previous string
		err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name=$1", migration.name).Scan(&previous)
		if err == nil {
			if previous != checksum {
				return errors.New("registry migration checksum mismatch")
			}
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if _, err = tx.Exec(ctx, migration.sql); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES ($1,$2)", migration.name, checksum); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
