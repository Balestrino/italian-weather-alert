package acquisition

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

//go:embed revision_schema.sql
var revisionSchema string

//go:embed expected_schema.sql
var expectedSchema string

//go:embed controls_schema.sql
var controlsSchema string

//go:embed unavailable_schema.sql
var unavailableSchema string

//go:embed target_dispositions_schema.sql
var targetDispositionsSchema string

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
	hash := sha256.Sum256([]byte(schema))
	checksum := hex.EncodeToString(hash[:])
	var previous string
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='005_acquisition'").Scan(&previous)
	if err == nil {
		if previous != checksum {
			return errors.New("acquisition migration checksum mismatch")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	} else {
		if _, err = tx.Exec(ctx, schema); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES ('005_acquisition',$1)", checksum); err != nil {
			return err
		}
	}
	revisionHash := sha256.Sum256([]byte(revisionSchema))
	revisionChecksum := hex.EncodeToString(revisionHash[:])
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='006_revision_tracking'").Scan(&previous)
	if err == nil {
		if previous != revisionChecksum {
			return errors.New("revision tracking migration checksum mismatch")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	} else {
		if _, err = tx.Exec(ctx, revisionSchema); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES ('006_revision_tracking',$1)", revisionChecksum); err != nil {
			return err
		}
	}
	expectedHash := sha256.Sum256([]byte(expectedSchema))
	expectedChecksum := hex.EncodeToString(expectedHash[:])
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='007_expected_publications'").Scan(&previous)
	if err == nil {
		if previous != expectedChecksum {
			return errors.New("expected publications migration checksum mismatch")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	} else {
		if _, err = tx.Exec(ctx, expectedSchema); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES ('007_expected_publications',$1)", expectedChecksum); err != nil {
			return err
		}
	}
	controlsHash := sha256.Sum256([]byte(controlsSchema))
	controlsChecksum := hex.EncodeToString(controlsHash[:])
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='022_acquisition_controls'").Scan(&previous)
	if err == nil {
		if previous != controlsChecksum {
			return errors.New("acquisition controls migration checksum mismatch")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	} else {
		if _, err = tx.Exec(ctx, controlsSchema); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES ('022_acquisition_controls',$1)", controlsChecksum); err != nil {
			return err
		}
	}
	missingHash := sha256.Sum256([]byte(unavailableSchema))
	missingChecksum := hex.EncodeToString(missingHash[:])
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='063_unavailable_targets'").Scan(&previous)
	if err == nil {
		if previous != missingChecksum {
			return errors.New("unavailable targets migration checksum mismatch")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	} else {
		if _, err = tx.Exec(ctx, unavailableSchema); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES ('063_unavailable_targets',$1)", missingChecksum); err != nil {
			return err
		}
	}
	dispositionsHash := sha256.Sum256([]byte(targetDispositionsSchema))
	dispositionsChecksum := hex.EncodeToString(dispositionsHash[:])
	err = tx.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='064_reviewed_target_dispositions'").Scan(&previous)
	if err == nil {
		if previous != dispositionsChecksum {
			return errors.New("reviewed target dispositions migration checksum mismatch")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	} else {
		if _, err = tx.Exec(ctx, targetDispositionsSchema); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES ('064_reviewed_target_dispositions',$1)", dispositionsChecksum); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
