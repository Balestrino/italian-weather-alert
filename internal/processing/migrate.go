package processing

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

//go:embed calls_schema.sql
var callsSchema string

//go:embed gates_schema.sql
var gatesSchema string

//go:embed rejections_schema.sql
var rejectionsSchema string

//go:embed checkpoints_schema.sql
var checkpointsSchema string

//go:embed checkpoint_time_schema.sql
var checkpointTimeSchema string

//go:embed recovery_schema.sql
var recoverySchema string

//go:embed invalid_output_schema.sql
var invalidOutputSchema string

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
	for _, migration := range []struct{ name, sql string }{{"004_processing", schema}, {"046_provider_calls", callsSchema}, {"047_provider_gates", gatesSchema}, {"049_provider_rejections", rejectionsSchema}, {"055_segment_checkpoints", checkpointsSchema}, {"056_checkpoint_use_time", checkpointTimeSchema}, {"060_bounded_recovery", recoverySchema}, {"061_invalid_classification_output", invalidOutputSchema}} {
		hash := sha256.Sum256([]byte(migration.sql))
		checksum := hex.EncodeToString(hash[:])
		var previous string
		err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name=$1", migration.name).Scan(&previous)
		if err == nil {
			if previous != checksum {
				return errors.New("processing migration checksum mismatch")
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
