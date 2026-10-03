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

//go:embed verification_schema.sql
var verificationSchema string

func MigrateVerification(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730021)"); err != nil {
		return err
	}
	h := sha256.Sum256([]byte(verificationSchema))
	checksum := hex.EncodeToString(h[:])
	var previous string
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='068_multi_source_verification'").Scan(&previous)
	if err == nil {
		if previous != checksum {
			return errors.New("multi-source verification migration checksum mismatch")
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err = tx.Exec(ctx, verificationSchema); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES('068_multi_source_verification',$1)", checksum); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
