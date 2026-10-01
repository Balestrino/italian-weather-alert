package classification

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Put(ctx context.Context, result Result) error {
	if result.RunID < 1 || result.DocumentVersionID < 1 || result.Status == "" || result.ReasonCode == "" || result.ContentSHA256 == "" || result.CreatedAt.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO classification_results(run_id,document_version_id,status,relevant,reason_code,evidence_quote,content_sha256,content_complete,provider_response_id,returned_model,input_tokens,output_tokens,cache_read_tokens,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,''),NULLIF($10,''),$11,$12,$13,$14) ON CONFLICT DO NOTHING`, result.RunID, result.DocumentVersionID, result.Status, result.Relevant, result.ReasonCode, result.EvidenceQuote, result.ContentSHA256, result.ContentComplete, result.ProviderResponseID, result.ReturnedModel, result.InputTokens, result.OutputTokens, result.CacheReadTokens, result.CreatedAt.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var hash, status string
		err = tx.QueryRow(ctx, "SELECT content_sha256,status FROM classification_results WHERE run_id=$1", result.RunID).Scan(&hash, &status)
		if err == nil && (hash != result.ContentSHA256 || status != result.Status) {
			return ErrConflict
		}
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	for _, segment := range result.Segments {
		if segment.Ordinal < 1 || segment.Total < segment.Ordinal || segment.DocumentVersionID != result.DocumentVersionID {
			return ErrInvalid
		}
		if _, err = tx.Exec(ctx, `INSERT INTO classification_segments(run_id,ordinal,total,document_version_id,resource_url,role,page_number,start_byte,end_byte,content_sha256,response_sha256,relevant,reason_code,evidence_quote,provider_response_id,returned_model,input_tokens,output_tokens,cache_read_tokens)
 VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,0),$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`, result.RunID, segment.Ordinal, segment.Total, segment.DocumentVersionID, segment.ResourceURL, segment.Role, segment.Page, segment.StartByte, segment.EndByte, segment.ContentSHA256, segment.ResponseSHA256, segment.Relevant, segment.ReasonCode, segment.EvidenceQuote, segment.ProviderResponseID, segment.ReturnedModel, segment.InputTokens, segment.OutputTokens, segment.CacheReadTokens); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) Get(ctx context.Context, runID int64) (Result, bool, error) {
	var result Result
	err := s.pool.QueryRow(ctx, `SELECT run_id,document_version_id,status,relevant,reason_code,evidence_quote,content_sha256,content_complete,COALESCE(provider_response_id,''),COALESCE(returned_model,''),input_tokens,output_tokens,cache_read_tokens,created_at
 FROM classification_results WHERE run_id=$1`, runID).Scan(&result.RunID, &result.DocumentVersionID, &result.Status, &result.Relevant, &result.ReasonCode, &result.EvidenceQuote, &result.ContentSHA256, &result.ContentComplete, &result.ProviderResponseID, &result.ReturnedModel, &result.InputTokens, &result.OutputTokens, &result.CacheReadTokens, &result.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, false, nil
	}
	if err != nil {
		return Result{}, false, err
	}
	rows, err := s.pool.Query(ctx, `SELECT ordinal,total,document_version_id,resource_url,role,COALESCE(page_number,0),start_byte,end_byte,content_sha256,response_sha256,relevant,reason_code,evidence_quote,provider_response_id,returned_model,input_tokens,output_tokens,cache_read_tokens FROM classification_segments WHERE run_id=$1 ORDER BY ordinal`, runID)
	if err != nil {
		return Result{}, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var segment SegmentResult
		if err = rows.Scan(&segment.Ordinal, &segment.Total, &segment.DocumentVersionID, &segment.ResourceURL, &segment.Role, &segment.Page, &segment.StartByte, &segment.EndByte, &segment.ContentSHA256, &segment.ResponseSHA256, &segment.Relevant, &segment.ReasonCode, &segment.EvidenceQuote, &segment.ProviderResponseID, &segment.ReturnedModel, &segment.InputTokens, &segment.OutputTokens, &segment.CacheReadTokens); err != nil {
			return Result{}, false, err
		}
		result.Segments = append(result.Segments, segment)
	}
	return result, true, rows.Err()
}
