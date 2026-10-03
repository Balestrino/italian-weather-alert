package publicquery

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730021)"); err != nil {
		return err
	}
	hash := sha256.Sum256([]byte(schema))
	checksum := hex.EncodeToString(hash[:])
	var previous string
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='019_public_queries'").Scan(&previous)
	if err == nil {
		if previous != checksum {
			return errors.New("public query migration checksum mismatch")
		}
		if err = tx.Commit(ctx); err != nil {
			return err
		}
		return migrateProjections(ctx, pool)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err = tx.Exec(ctx, schema); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES('019_public_queries',$1)", checksum); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return migrateProjections(ctx, pool)
}

func migrateProjections(ctx context.Context, pool *pgxpool.Pool) error {
	if err := domain.MigrateRegionalProjection(ctx, pool); err != nil {
		return err
	}
	return domain.MigrateVerification(ctx, pool)
}
