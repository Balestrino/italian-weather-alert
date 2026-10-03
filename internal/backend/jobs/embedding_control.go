package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// EmbeddingControl is shared by every worker in one environment's database.
// Existing environment variables never override its disabled initial state.
type EmbeddingControl struct {
	SourceID  string    `json:"source_id,omitempty"`
	Enabled   bool      `json:"enabled"`
	Revision  int       `json:"revision"`
	Actor     string    `json:"actor"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Store) SourceEmbeddingState(ctx context.Context, source string) (EmbeddingControl, error) {
	if !validName(source) {
		return EmbeddingControl{}, ErrInvalid
	}
	state := EmbeddingControl{SourceID: source}
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(c.enabled,false),COALESCE(c.revision,0),COALESCE(c.actor,''),COALESCE(c.updated_at,'epoch'::timestamptz)
        FROM registry_sources r LEFT JOIN processing_source_embedding_controls c ON c.source_id=r.id WHERE r.id=$1`, source).Scan(&state.Enabled, &state.Revision, &state.Actor, &state.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrInvalid
	}
	return state, err
}

func (s *Store) EmbeddingAllowedForRun(ctx context.Context, run int64) (bool, error) {
	var allowed bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM processing_embedding_control WHERE singleton AND enabled)
        AND processing_embedding_source_allowed(jsonb_build_object('extraction_run_id',$1::bigint))`, run).Scan(&allowed)
	return allowed, err
}

func (s *Store) SetSourceEmbeddingEnabled(ctx context.Context, source string, expected int, enabled bool, actor string, at time.Time) (EmbeddingControl, error) {
	if !validName(source) || expected < 0 || !validName(actor) || at.IsZero() {
		return EmbeddingControl{}, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return EmbeddingControl{}, err
	}
	defer tx.Rollback(ctx)
	// Use the same lock as queue admission, including absent/new source flags.
	if _, err = tx.Exec(ctx, `SELECT singleton FROM processing_embedding_control WHERE singleton FOR UPDATE`); err != nil {
		return EmbeddingControl{}, err
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM registry_sources WHERE id=$1)`, source).Scan(&exists); err != nil {
		return EmbeddingControl{}, err
	}
	if !exists {
		return EmbeddingControl{}, ErrInvalid
	}
	var revision int
	if err = tx.QueryRow(ctx, `SELECT COALESCE((SELECT revision FROM processing_source_embedding_controls WHERE source_id=$1),0)`, source).Scan(&revision); err != nil {
		return EmbeddingControl{}, err
	}
	if revision != expected {
		return EmbeddingControl{}, ErrConflict
	}
	state := EmbeddingControl{SourceID: source, Enabled: enabled, Revision: revision + 1, Actor: actor, UpdatedAt: at.UTC()}
	if _, err = tx.Exec(ctx, `INSERT INTO processing_source_embedding_controls(source_id,enabled,revision,actor,updated_at) VALUES($1,$2,$3,$4,$5)
        ON CONFLICT(source_id) DO UPDATE SET enabled=EXCLUDED.enabled,revision=EXCLUDED.revision,actor=EXCLUDED.actor,updated_at=EXCLUDED.updated_at`, source, enabled, state.Revision, actor, state.UpdatedAt); err != nil {
		return state, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO processing_embedding_control_events(source_id,revision,enabled,actor,created_at) VALUES($1,$2,$3,$4,$5)`, source, state.Revision, enabled, actor, state.UpdatedAt); err != nil {
		return state, err
	}
	return state, tx.Commit(ctx)
}

func (s *Store) EmbeddingState(ctx context.Context) (EmbeddingControl, error) {
	var state EmbeddingControl
	err := s.pool.QueryRow(ctx, `SELECT enabled,revision,actor,updated_at FROM processing_embedding_control WHERE singleton`).Scan(&state.Enabled, &state.Revision, &state.Actor, &state.UpdatedAt)
	return state, err
}

func (s *Store) EmbeddingEnabled(ctx context.Context) (bool, error) {
	state, err := s.EmbeddingState(ctx)
	return state.Enabled, err
}

func (s *Store) EmbeddingActive(ctx context.Context) (bool, error) {
	var enabled bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM processing_embedding_control WHERE singleton AND enabled)
        AND EXISTS(SELECT 1 FROM processing_source_embedding_controls c JOIN registry_sources r ON r.id=c.source_id WHERE c.enabled)`).Scan(&enabled)
	return enabled, err
}

func (s *Store) SourceEmbeddingStates(ctx context.Context) ([]EmbeddingControl, error) {
	rows, err := s.pool.Query(ctx, `SELECT r.id,COALESCE(c.enabled,false),COALESCE(c.revision,0),COALESCE(c.actor,''),COALESCE(c.updated_at,'epoch'::timestamptz)
        FROM registry_sources r LEFT JOIN processing_source_embedding_controls c ON c.source_id=r.id ORDER BY r.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	states := []EmbeddingControl{}
	for rows.Next() {
		var state EmbeddingControl
		if err = rows.Scan(&state.SourceID, &state.Enabled, &state.Revision, &state.Actor, &state.UpdatedAt); err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	return states, rows.Err()
}

func (s *Store) SetEmbeddingEnabled(ctx context.Context, expected int, enabled bool, actor string, at time.Time) (EmbeddingControl, error) {
	if expected < 0 || !validName(actor) || at.IsZero() {
		return EmbeddingControl{}, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return EmbeddingControl{}, err
	}
	defer tx.Rollback(ctx)
	var state EmbeddingControl
	err = tx.QueryRow(ctx, `SELECT enabled,revision,actor,updated_at FROM processing_embedding_control WHERE singleton FOR UPDATE`).Scan(&state.Enabled, &state.Revision, &state.Actor, &state.UpdatedAt)
	if err != nil {
		return state, err
	}
	if state.Revision != expected {
		return state, ErrConflict
	}
	state = EmbeddingControl{Enabled: enabled, Revision: expected + 1, Actor: actor, UpdatedAt: at.UTC()}
	if _, err = tx.Exec(ctx, `UPDATE processing_embedding_control SET enabled=$1,revision=$2,actor=$3,updated_at=$4 WHERE singleton`, state.Enabled, state.Revision, state.Actor, state.UpdatedAt); err != nil {
		return state, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO processing_embedding_control_events(revision,enabled,actor,created_at) VALUES($1,$2,$3,$4)`, state.Revision, state.Enabled, state.Actor, state.UpdatedAt); err != nil {
		return state, err
	}
	return state, tx.Commit(ctx)
}
