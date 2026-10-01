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

//go:embed regional_projection_schema.sql
var regionalProjectionSchema string

func MigrateRegionalProjection(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730021)"); err != nil {
		return err
	}
	hash := sha256.Sum256([]byte(regionalProjectionSchema))
	checksum := hex.EncodeToString(hash[:])
	var previous string
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='062_regional_projection'").Scan(&previous)
	if err == nil {
		if previous != checksum {
			return errors.New("regional projection migration checksum mismatch")
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err = tx.Exec(ctx, regionalProjectionSchema); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES('062_regional_projection',$1)", checksum); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
