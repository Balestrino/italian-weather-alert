package linking

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed semantic_schema.sql
var semanticSchema string

func MigrateSemantic(ctx context.Context, pool *pgxpool.Pool) error {
	tx, e := pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730021)"); e != nil {
		return e
	}
	h := sha256.Sum256([]byte(semanticSchema))
	sum := hex.EncodeToString(h[:])
	var old string
	e = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='012_embedding'").Scan(&old)
	if e == nil {
		if old != sum {
			return errors.New("embedding migration checksum mismatch")
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return e
	}
	if _, e = tx.Exec(ctx, semanticSchema); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES('012_embedding',$1)", sum); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
