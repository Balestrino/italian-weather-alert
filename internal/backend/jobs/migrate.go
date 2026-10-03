package jobs

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

//go:embed relaunch_schema.sql
var relaunchSchema string

//go:embed defer_schema.sql
var deferSchema string

//go:embed archive_schema.sql
var archiveSchema string

//go:embed embedding_control_schema.sql
var embeddingControlSchema string

// Migrate applies the queue schema atomically and rejects edits to applied SQL.
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
	for _, migration := range []struct{ name, sql string }{{"003_jobs", schema}, {"024_job_relaunches", relaunchSchema}, {"048_job_deferrals", deferSchema}, {"057_job_archival", archiveSchema}, {"068_embedding_control", embeddingControlSchema}} {
		hash := sha256.Sum256([]byte(migration.sql))
		checksum := hex.EncodeToString(hash[:])
		var previous string
		err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name=$1", migration.name).Scan(&previous)
		if err == nil {
			if previous != checksum {
				return errors.New("jobs migration checksum mismatch")
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
