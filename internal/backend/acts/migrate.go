// Package acts records acts explicitly referenced by municipal notices, their
// resolution and retention, and the measure clarifications derived from them.
package acts

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

// Migrate requires the registry, documents, processing and domain schemas.
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
	hash := sha256.Sum256([]byte(schema))
	checksum := hex.EncodeToString(hash[:])
	var previous string
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='101_referenced_acts'").Scan(&previous)
	if err == nil {
		if previous != checksum {
			return errors.New("referenced acts migration checksum mismatch")
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err = tx.Exec(ctx, schema); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES ('101_referenced_acts',$1)", checksum); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
