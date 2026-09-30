package registry

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
)

// SuspendInterpretation records a source-wide confirmed defect independently
// of collection and public document visibility.
func (s *Store) SuspendInterpretation(ctx context.Context, id string, rev int, actor string, defect Evidence) error {
	if !validEvidence(defect) {
		return ErrInvalid
	}
	return s.change(ctx, id, rev, actor, func(tx pgx.Tx, st State, c Configuration) error {
		if st.ActiveRevision == nil || *st.ActiveRevision != rev || st.InterpretationSuspendedAt != nil {
			return ErrConflict
		}
		raw, err := json.Marshal(defect)
		if err != nil {
			return ErrInvalid
		}
		_, err = tx.Exec(ctx, `WITH suspended AS (
   INSERT INTO registry_interpretation_suspensions(source_id,revision,actor,defect) VALUES($1,$2,$3,$4) RETURNING suspended_at)
   UPDATE registry_sources SET interpretation_suspended_at=suspended.suspended_at FROM suspended WHERE id=$1`, id, rev, actor, raw)
		return err
	})
}

type InterpretationSuspension struct {
	ID          int64
	Revision    int
	Actor       string
	Defect      Evidence
	SuspendedAt time.Time
	ResumedAt   *time.Time
	ResumedBy   *string
	Recovery    json.RawMessage
}

// InterpretationHistory remains private; public results expose a fixed warning,
// never operator notes or diagnostic report contents.
func (s *Store) InterpretationHistory(ctx context.Context, id string) ([]InterpretationSuspension, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,revision,actor,defect,suspended_at,resumed_at,resumed_by,recovery FROM registry_interpretation_suspensions WHERE source_id=$1 ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []InterpretationSuspension{}
	for rows.Next() {
		var v InterpretationSuspension
		if err = rows.Scan(&v.ID, &v.Revision, &v.Actor, &v.Defect, &v.SuspendedAt, &v.ResumedAt, &v.ResumedBy, &v.Recovery); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
