package classification

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

//go:embed segmentation_schema.sql
var segmentationSchema string

//go:embed manifest_schema.sql
var manifestSchema string

//go:embed reuse_schema.sql
var reuseSchema string

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
	for _, migration := range []struct{ name, sql string }{{"009_classification", schema}, {"033_classification_segments", segmentationSchema}, {"053_interpretation_manifests", manifestSchema}, {"054_interpretation_reuse", reuseSchema}} {
		hash := sha256.Sum256([]byte(migration.sql))
		checksum := hex.EncodeToString(hash[:])
		var previous string
		err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name=$1", migration.name).Scan(&previous)
		if err == nil {
			if previous != checksum {
				return errors.New("classification migration checksum mismatch")
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
