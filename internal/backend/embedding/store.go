package embedding

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func NewStore(p *pgxpool.Pool) *Store { return &Store{p} }
func (s *Store) Put(ctx context.Context, r Result) error {
	if r.RunID < 1 || r.ExtractionRunID < 1 || r.MeasureOrdinal < 1 || r.Dimensions < 1 || len(r.Vector) != r.Dimensions || r.ConfigurationVersionID == "" || r.ReturnedModel == "" || r.CreatedAt.IsZero() {
		return ErrInvalid
	}
	_, e := s.pool.Exec(ctx, `INSERT INTO measure_embeddings(run_id,extraction_run_id,measure_ordinal,configuration_version_id,dimensions,vector,returned_model,input_tokens,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(extraction_run_id,measure_ordinal,configuration_version_id) DO NOTHING`, r.RunID, r.ExtractionRunID, r.MeasureOrdinal, r.ConfigurationVersionID, r.Dimensions, r.Vector, r.ReturnedModel, r.InputTokens, r.CreatedAt.UTC())
	return e
}
func (s *Store) Get(ctx context.Context, run int64, ordinal int, config string) (Result, bool, error) {
	var r Result
	e := s.pool.QueryRow(ctx, `SELECT run_id,extraction_run_id,measure_ordinal,configuration_version_id,dimensions,vector,returned_model,input_tokens,created_at FROM measure_embeddings WHERE extraction_run_id=$1 AND measure_ordinal=$2 AND configuration_version_id=$3`, run, ordinal, config).Scan(&r.RunID, &r.ExtractionRunID, &r.MeasureOrdinal, &r.ConfigurationVersionID, &r.Dimensions, &r.Vector, &r.ReturnedModel, &r.InputTokens, &r.CreatedAt)
	if errors.Is(e, pgx.ErrNoRows) {
		return Result{}, false, nil
	}
	return r, e == nil, e
}
