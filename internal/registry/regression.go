package registry

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/evaluation"

	"github.com/jackc/pgx/v5"
)

// RecordRegression imports an operator's evidence-backed comparison. It never
// edits measures or enables publication. CAS and the source lock serialize it
// with acceptance/activation; an earlier in-flight run cannot clear a failure.
func (s *Store) RecordRegression(ctx context.Context, id string, rev int, actor, previousID string, run evaluation.Run) error {
	reports, err := run.Reports()
	if err != nil || run.SourceID != id || run.Revision != rev || run.FinishedAt.After(time.Now().UTC()) {
		return ErrInvalid
	}
	return s.change(ctx, id, rev, actor, func(tx pgx.Tx, st State, c Configuration) error {
		var lastID, contract string
		var passed bool
		var recorded time.Time
		err := tx.QueryRow(ctx, `SELECT id,contract_hash,passed,recorded_at FROM registry_regressions WHERE source_id=$1 AND revision<=$2 AND suite=$3 ORDER BY revision DESC,recorded_at DESC,id DESC LIMIT 1`, id, rev, run.Suite).Scan(&lastID, &contract, &passed, &recorded)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if previousID != lastID {
			return ErrConflict
		}
		if lastID != "" {
			if contract != run.ContractHash() || !run.StartedAt.After(recorded) {
				return ErrPrerequisite
			}
			if !passed && reports.Passed && strings.TrimSpace(run.Correction) == "" {
				return ErrPrerequisite
			}
		}
		body, err := json.Marshal(run)
		if err != nil {
			return ErrInvalid
		}
		_, err = tx.Exec(ctx, `INSERT INTO registry_regressions(id,source_id,revision,suite,contract_hash,previous_id,passed,actor,report) VALUES($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9)`, run.ID, id, rev, run.Suite, run.ContractHash(), previousID, reports.Passed, actor, body)
		return err
	})
}

// All suites must have a successful latest run. Failed evaluations invalidate
// earlier acceptance; passing a rerun still requires a new explicit acceptance.
const regressionReadySQL = `NOT EXISTS (
 SELECT 1 FROM (SELECT DISTINCT ON (suite) passed,revision FROM registry_regressions
 WHERE source_id=$1 AND revision<=$2 ORDER BY suite,revision DESC,recorded_at DESC,id DESC) latest WHERE NOT passed OR revision<>$2)`

const acceptedSQL = `EXISTS(SELECT 1 FROM registry_events e WHERE e.source_id=$1 AND e.revision=$2 AND e.kind='acceptance'
 AND NOT EXISTS(SELECT 1 FROM registry_regressions r WHERE r.source_id=e.source_id AND r.revision=e.revision AND NOT r.passed AND r.recorded_at>=e.created_at))`

func regressionReady(ctx context.Context, tx pgx.Tx, id string, rev int) (bool, error) {
	var ready bool
	err := tx.QueryRow(ctx, "SELECT "+regressionReadySQL, id, rev).Scan(&ready)
	return ready, err
}

type Regression struct {
	Run        evaluation.Run     `json:"run"`
	Reports    evaluation.Reports `json:"reports"`
	Actor      string             `json:"actor"`
	PreviousID *string            `json:"previous_id"`
	RecordedAt time.Time          `json:"recorded_at"`
}

func (s *Store) Regressions(ctx context.Context, id string) ([]Regression, error) {
	if _, err := s.State(ctx, id); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT report,actor,previous_id,recorded_at FROM registry_regressions WHERE source_id=$1 ORDER BY recorded_at,id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Regression{}
	for rows.Next() {
		var value Regression
		var raw []byte
		if err = rows.Scan(&raw, &value.Actor, &value.PreviousID, &value.RecordedAt); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(raw, &value.Run); err != nil {
			return nil, err
		}
		if value.Reports, err = value.Run.Reports(); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
