package ocr

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

type pageQuerier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (s *Store) PutPage(ctx context.Context, page PageResult) error {
	return putPage(ctx, s.pool, page)
}
func putPage(ctx context.Context, db pageQuerier, page PageResult) error {
	if page.RunID < 1 || page.DocumentVersionID < 1 || page.PageNumber < 1 || page.ResourceURL == "" || (page.Status != "complete" && page.Status != "unreadable") || page.CreatedAt.IsZero() {
		return ErrInvalid
	}
	tag, err := db.Exec(ctx, `INSERT INTO ocr_page_results(run_id,page_number,document_version_id,resource_url,status,media_type,input_sha256,output_sha256,extracted_text,provider_response_id,returned_model,input_tokens,output_tokens,cache_read_tokens,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) ON CONFLICT DO NOTHING`, page.RunID, page.PageNumber, page.DocumentVersionID, page.ResourceURL, page.Status, page.MediaType, page.InputSHA256, page.OutputSHA256, page.ExtractedText, page.ProviderResponseID, page.ReturnedModel, page.InputTokens, page.OutputTokens, page.CacheReadTokens, page.CreatedAt.UTC())
	if err != nil || tag.RowsAffected() == 1 {
		return err
	}
	var hash, input, status, resource, media string
	var version int64
	err = db.QueryRow(ctx, "SELECT output_sha256,input_sha256,status,document_version_id,resource_url,media_type FROM ocr_page_results WHERE run_id=$1 AND page_number=$2", page.RunID, page.PageNumber).Scan(&hash, &input, &status, &version, &resource, &media)
	if err == nil && (hash != page.OutputSHA256 || input != page.InputSHA256 || status != page.Status || version != page.DocumentVersionID || resource != page.ResourceURL || media != page.MediaType) {
		return ErrConflict
	}
	return err
}

func (s *Store) Page(ctx context.Context, runID int64, pageNumber int) (PageResult, bool, error) {
	var page PageResult
	err := s.pool.QueryRow(ctx, `SELECT run_id,page_number,document_version_id,resource_url,status,media_type,input_sha256,output_sha256,extracted_text,provider_response_id,returned_model,input_tokens,output_tokens,cache_read_tokens,created_at
 FROM ocr_page_results WHERE run_id=$1 AND page_number=$2`, runID, pageNumber).Scan(&page.RunID, &page.PageNumber, &page.DocumentVersionID, &page.ResourceURL, &page.Status, &page.MediaType, &page.InputSHA256, &page.OutputSHA256, &page.ExtractedText, &page.ProviderResponseID, &page.ReturnedModel, &page.InputTokens, &page.OutputTokens, &page.CacheReadTokens, &page.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PageResult{}, false, nil
	}
	return page, err == nil, err
}

func (s *Store) PutResource(ctx context.Context, result ResourceResult) error {
	if result.RunID < 1 || result.DocumentVersionID < 1 || result.ResourceURL == "" || result.PageCount < 0 || result.CreatedAt.IsZero() {
		return ErrInvalid
	}
	tag, err := s.pool.Exec(ctx, `INSERT INTO ocr_resource_results(run_id,document_version_id,resource_url,status,page_count,error_code,created_at)
 VALUES($1,$2,$3,$4,$5,NULLIF($6,''),$7) ON CONFLICT DO NOTHING`, result.RunID, result.DocumentVersionID, result.ResourceURL, result.Status, result.PageCount, result.ErrorCode, result.CreatedAt.UTC())
	if err != nil || tag.RowsAffected() == 1 {
		return err
	}
	var status string
	var pages int
	err = s.pool.QueryRow(ctx, "SELECT status,page_count FROM ocr_resource_results WHERE run_id=$1", result.RunID).Scan(&status, &pages)
	if err == nil && (status != result.Status || pages != result.PageCount) {
		return ErrConflict
	}
	return err
}

func (s *Store) Resource(ctx context.Context, runID int64) (ResourceResult, bool, error) {
	var result ResourceResult
	err := s.pool.QueryRow(ctx, `SELECT run_id,document_version_id,resource_url,status,page_count,COALESCE(error_code,''),created_at
 FROM ocr_resource_results WHERE run_id=$1`, runID).Scan(&result.RunID, &result.DocumentVersionID, &result.ResourceURL, &result.Status, &result.PageCount, &result.ErrorCode, &result.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ResourceResult{}, false, nil
	}
	return result, err == nil, err
}

// PagesForVersion returns the newest persisted OCR page for each retained
// resource/page pair. Reprocessing history remains stored; classification uses
// this view only when its scheduler explicitly selects the document version.
func (s *Store) PagesForVersion(ctx context.Context, versionID int64) ([]PageResult, error) {
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT ON (resource_url,page_number)
 run_id,page_number,document_version_id,resource_url,status,media_type,input_sha256,output_sha256,extracted_text,provider_response_id,returned_model,input_tokens,output_tokens,cache_read_tokens,created_at
 FROM ocr_page_results WHERE document_version_id=$1 ORDER BY resource_url,page_number,created_at DESC,run_id DESC`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pages []PageResult
	for rows.Next() {
		var page PageResult
		if err = rows.Scan(&page.RunID, &page.PageNumber, &page.DocumentVersionID, &page.ResourceURL, &page.Status, &page.MediaType, &page.InputSHA256, &page.OutputSHA256, &page.ExtractedText, &page.ProviderResponseID, &page.ReturnedModel, &page.InputTokens, &page.OutputTokens, &page.CacheReadTokens, &page.CreatedAt); err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	return pages, rows.Err()
}

func (s *Store) ResourcesForVersion(ctx context.Context, versionID int64) ([]ResourceResult, error) {
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT ON (resource_url)
 run_id,document_version_id,resource_url,status,page_count,COALESCE(error_code,''),created_at
 FROM ocr_resource_results WHERE document_version_id=$1 ORDER BY resource_url,created_at DESC,run_id DESC`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []ResourceResult
	for rows.Next() {
		var result ResourceResult
		if err = rows.Scan(&result.RunID, &result.DocumentVersionID, &result.ResourceURL, &result.Status, &result.PageCount, &result.ErrorCode, &result.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}
