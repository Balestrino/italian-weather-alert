package interpretation

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

//go:embed evaluation_schema.sql
var evaluationSchema string

//go:embed archive_schema.sql
var archiveSchema string

//go:embed preflight_schema.sql
var preflightSchema string

//go:embed archive_recovery_schema.sql
var archiveRecoverySchema string

func Migrate(ctx context.Context, p *pgxpool.Pool) error {
	tx, e := p.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(730021)"); e != nil {
		return e
	}
	for _, migration := range []struct{ name, sql string }{{"013_interpretation_scheduling", schema}, {"029_reprocessing_evaluation", evaluationSchema}, {"058_pending_interpretation_archival", archiveSchema}, {"059_interpretation_preflight", preflightSchema}, {"096_selected_archive_recovery", archiveRecoverySchema}} {
		h := sha256.Sum256([]byte(migration.sql))
		sum := hex.EncodeToString(h[:])
		var old string
		e = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name=$1", migration.name).Scan(&old)
		if e == nil {
			if old != sum {
				return errors.New("interpretation migration checksum mismatch")
			}
			continue
		}
		if !errors.Is(e, pgx.ErrNoRows) {
			return e
		}
		if _, e = tx.Exec(ctx, migration.sql); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES($1,$2)", migration.name, sum); e != nil {
			return e
		}
	}
	return tx.Commit(ctx)
}
