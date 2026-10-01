package processing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// InvalidOutput contains request/response bodies only, never credentials, HTTP
// headers or provider error bodies. Bytea preserves even NUL in invalid model text.
type InvalidOutput struct {
	RunID                         int64
	AttemptNumber, SegmentOrdinal int
	Request, Response             []byte
	FinishReason, ErrorCode       string
	Cached                        bool
	CreatedAt                     time.Time
}

func (s *Store) RecordInvalidOutput(ctx context.Context, v InvalidOutput) error {
	if v.RunID < 1 || v.AttemptNumber < 1 || v.SegmentOrdinal < 1 || len(v.Request) == 0 || len(v.Request) > 16<<20 || !json.Valid(v.Request) || len(v.Response) > 8<<20 || !callIdentifier.MatchString(v.FinishReason) || !validName(v.ErrorCode) || v.CreatedAt.IsZero() {
		return ErrInvalid
	}
	ih, rh := sha256.Sum256(v.Request), sha256.Sum256(v.Response)
	inputHash, responseHash := hex.EncodeToString(ih[:]), hex.EncodeToString(rh[:])
	// Only structured text interpretation attempts may use this capture path.
	var stage string
	if err := s.pool.QueryRow(ctx, `SELECT stage FROM processing_runs WHERE id=$1`, v.RunID).Scan(&stage); err != nil {
		return err
	}
	if stage != "classification" && stage != "extraction" {
		return ErrInvalid
	}
	response := v.Response
	if response == nil {
		response = []byte{}
	}
	tag, err := s.pool.Exec(ctx, `INSERT INTO processing_invalid_outputs(run_id,attempt_number,segment_ordinal,input_sha256,response_sha256,request_bytes,response_bytes,finish_reason,error_code,cached,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING`, v.RunID, v.AttemptNumber, v.SegmentOrdinal, inputHash, responseHash, v.Request, response, v.FinishReason, v.ErrorCode, v.Cached, v.CreatedAt.UTC())
	if err != nil || tag.RowsAffected() == 1 {
		return err
	}
	var matches bool
	err = s.pool.QueryRow(ctx, `SELECT input_sha256=$4 AND response_sha256=$5 AND finish_reason=$6 AND error_code=$7 AND cached=$8 FROM processing_invalid_outputs WHERE run_id=$1 AND attempt_number=$2 AND segment_ordinal=$3`, v.RunID, v.AttemptNumber, v.SegmentOrdinal, inputHash, responseHash, v.FinishReason, v.ErrorCode, v.Cached).Scan(&matches)
	if err == nil && !matches {
		return ErrConflict
	}
	return err
}
