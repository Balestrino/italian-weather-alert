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

//go:embed territorial_schema.sql
var territorialSchema string

func migrateTerritorialCatalog(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730021)"); err != nil {
		return err
	}
	hash := sha256.Sum256([]byte(territorialSchema))
	checksum := hex.EncodeToString(hash[:])
	var previous string
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='050_territorial_catalog'").Scan(&previous)
	if err == nil {
		if previous != checksum {
			return errors.New("territorial migration checksum mismatch")
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err = tx.Exec(ctx, territorialSchema); err != nil {
		return err
	}
	if err = seedRegions(ctx, tx); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES('050_territorial_catalog',$1)", checksum); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func MigrateTerritories(ctx context.Context, pool *pgxpool.Pool) error {
	if err := migrateTerritorialCatalog(ctx, pool); err != nil {
		return err
	}
	if err := migrateTerritorialScope(ctx, pool); err != nil {
		return err
	}
	if err := migrateTerritorialGates(ctx, pool); err != nil {
		return err
	}
	return migrateTerritorialGateHardening(ctx, pool)
}
