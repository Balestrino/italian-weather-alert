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

//go:embed territorial_gate_hardening_schema.sql
var territorialGateHardeningSchema string

func migrateTerritorialGateHardening(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730021)"); err != nil {
		return err
	}
	hash := sha256.Sum256([]byte(territorialGateHardeningSchema))
	checksum := hex.EncodeToString(hash[:])
	var previous string
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='054_territorial_gate_hardening'").Scan(&previous)
	if err == nil {
		if previous != checksum {
			return errors.New("territorial gate migration checksum mismatch")
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err = tx.Exec(ctx, territorialGateHardeningSchema); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES('054_territorial_gate_hardening',$1)", checksum); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
