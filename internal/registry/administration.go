package registry

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Intervals are independent of draft collection boundaries. Overrides survive
// later activations; configuration snapshots retain the originally supplied values.
type Intervals struct {
	Revision     int  `json:"revision"`
	CheckSeconds *int `json:"check_seconds"`
	DelaySeconds *int `json:"delay_seconds"`
}

func (s *Store) Intervals(ctx context.Context, id string) (Intervals, error) {
	var v Intervals
	err := s.pool.QueryRow(ctx, `SELECT interval_revision,check_seconds_override,delay_seconds_override FROM registry_sources WHERE id=$1`, id).Scan(&v.Revision, &v.CheckSeconds, &v.DelaySeconds)
	if err == pgx.ErrNoRows {
		err = ErrNotFound
	}
	return v, err
}

func (s *Store) SetIntervals(ctx context.Context, id string, expected, check, delay int, actor string) error {
	if strings.TrimSpace(actor) == "" || check < 1 || delay < 1 || check > 31536000 || delay > 31536000 {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	st, err := state(ctx, tx, id, true)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `UPDATE registry_sources SET check_seconds_override=$3,delay_seconds_override=$4,interval_revision=interval_revision+1 WHERE id=$1 AND interval_revision=$2`, id, expected, check, delay)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrConflict
	}
	rev := st.LatestRevision
	if st.ActiveRevision != nil {
		rev = *st.ActiveRevision
	}
	if err = record(ctx, tx, id, rev, "intervals_changed", actor, Intervals{expected + 1, &check, &delay}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) Sources(ctx context.Context) ([]State, error) {
	rows, err := s.pool.Query(ctx, `SELECT id FROM registry_sources ORDER BY id`)
	if err != nil {
		return nil, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	result := []State{}
	for _, id := range ids {
		st, err := s.State(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, st)
	}
	return result, nil
}

func (s *Store) Versions(ctx context.Context, id string) ([]Version, error) {
	st, err := s.State(ctx, id)
	if err != nil {
		return nil, err
	}
	result := make([]Version, 0, st.LatestRevision)
	for revision := 1; revision <= st.LatestRevision; revision++ {
		v, err := s.Version(ctx, id, revision)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}
