package observation

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

//go:embed schema_scope.sql
var scopeSchema string

//go:embed schema_corrections.sql
var correctionsSchema string

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730021)"); err != nil {
		return err
	}
	for _, migration := range []struct{ name, body string }{
		{"030_observational_trials", schema},
		{"095_municipal_observation_scope", scopeSchema},
		{"097_observation_review_corrections", correctionsSchema},
	} {
		hash := sha256.Sum256([]byte(migration.body))
		checksum := hex.EncodeToString(hash[:])
		var previous string
		err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name=$1", migration.name).Scan(&previous)
		if err == nil {
			if previous != checksum {
				return errors.New("observational trial migration checksum mismatch")
			}
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if _, err = tx.Exec(ctx, migration.body); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES($1,$2)", migration.name, checksum); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
