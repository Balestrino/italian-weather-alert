package processing

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type SegmentCheckpoint struct {
	Stage, Configuration, InputSHA256 string
	CallID                            int64
	Response                          json.RawMessage
}

func (c SegmentCheckpoint) valid() bool {
	return (c.Stage == "classification" || c.Stage == "extraction") && c.Configuration != "" && callHash.MatchString(c.InputSHA256)
}

func (s *Store) Checkpoint(ctx context.Context, key SegmentCheckpoint) (SegmentCheckpoint, bool, error) {
	if !key.valid() {
		return key, false, ErrInvalid
	}
	err := s.pool.QueryRow(ctx, `SELECT call_id,response FROM processing_segment_checkpoints WHERE stage=$1 AND configuration=$2 AND input_sha256=$3`, key.Stage, key.Configuration, key.InputSHA256).Scan(&key.CallID, &key.Response)
	if errors.Is(err, pgx.ErrNoRows) {
		return key, false, nil
	}
	return key, err == nil, err
}
func (s *Store) PutCheckpoint(ctx context.Context, c SegmentCheckpoint, at time.Time) error {
	if !c.valid() || c.CallID < 1 || at.IsZero() || len(c.Response) > 65536 || !json.Valid(c.Response) {
		return ErrInvalid
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO processing_segment_checkpoints(stage,configuration,input_sha256,call_id,response,created_at)
 SELECT $1,$2,$3,c.id,$5,$6 FROM processing_provider_calls c JOIN processing_runs r ON r.id=c.run_id
 WHERE c.id=$4 AND c.state='received' AND c.input_sha256=$3 AND r.configuration_version_id=$2 AND r.stage=$1
 ON CONFLICT DO NOTHING`, c.Stage, c.Configuration, c.InputSHA256, c.CallID, c.Response, at)
	if err != nil {
		return err
	}
	stored, ok, err := s.Checkpoint(ctx, c)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalid
	}
	// Equal requests may race. Preserve the first validated checkpoint; another
	// paid receipt stays in the ledger rather than replacing historical output.
	if len(stored.Response) == 0 {
		return ErrInvalid
	}
	return nil
}
func (s *Store) UseCheckpoint(ctx context.Context, c SegmentCheckpoint, attempt Attempt, ordinal int, reused bool) error {
	if !c.valid() || attempt.RunID < 1 || attempt.Number < 1 || ordinal < 1 {
		return ErrInvalid
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO processing_segment_checkpoint_uses(run_id,attempt_number,segment_ordinal,stage,configuration,input_sha256,reused)
 SELECT $1,$2,$3,$4,$5,$6,$7 WHERE EXISTS(SELECT 1 FROM processing_runs WHERE id=$1 AND stage=$4 AND configuration_version_id=$5)
 ON CONFLICT DO NOTHING`, attempt.RunID, attempt.Number, ordinal, c.Stage, c.Configuration, c.InputSHA256, reused)
	if err != nil {
		return err
	}
	var matches bool
	err = s.pool.QueryRow(ctx, `SELECT stage=$4 AND configuration=$5 AND input_sha256=$6 AND reused=$7 FROM processing_segment_checkpoint_uses WHERE run_id=$1 AND attempt_number=$2 AND segment_ordinal=$3`, attempt.RunID, attempt.Number, ordinal, c.Stage, c.Configuration, c.InputSHA256, reused).Scan(&matches)
	if errors.Is(err, pgx.ErrNoRows) || err == nil && !matches {
		return ErrInvalid
	}
	return err
}
