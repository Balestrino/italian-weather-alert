package classification

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Store) EquivalentRun(ctx context.Context, runID int64, stage string) (int64, bool, error) {
	table, status := "classification_results", "classified"
	if stage == "extraction" {
		table, status = "extraction_results", "extracted"
	} else if stage != "classification" {
		return 0, false, ErrInvalid
	}
	var original int64
	err := s.pool.QueryRow(ctx, `SELECT prior.run_id FROM interpretation_input_manifests current
 JOIN interpretation_input_manifests prior ON prior.document_id=current.document_id AND prior.source_id=current.source_id AND prior.configuration=current.configuration AND prior.manifest_sha256=current.manifest_sha256
 JOIN `+table+` result ON result.run_id=prior.run_id AND result.status=$2
 WHERE current.run_id=$1 AND current.reusable AND prior.reusable AND prior.run_id<current.run_id ORDER BY prior.run_id LIMIT 1`, runID, status).Scan(&original)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return original, err == nil, err
}

func (s *Store) RecordReuse(ctx context.Context, runID, original int64, at time.Time) error {
	if runID <= original || original < 1 || at.IsZero() {
		return ErrInvalid
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO interpretation_reuse(run_id,original_run_id,created_at)
 SELECT c.run_id,p.run_id,$3 FROM interpretation_input_manifests c JOIN interpretation_input_manifests p ON c.document_id=p.document_id AND c.source_id=p.source_id AND c.configuration=p.configuration AND c.manifest_sha256=p.manifest_sha256
 WHERE c.run_id=$1 AND p.run_id=$2 AND c.reusable AND p.reusable ON CONFLICT DO NOTHING`, runID, original, at)
	if err != nil {
		return err
	}
	var stored int64
	if err = s.pool.QueryRow(ctx, `SELECT original_run_id FROM interpretation_reuse WHERE run_id=$1`, runID).Scan(&stored); err != nil {
		return err
	}
	if stored != original {
		return ErrConflict
	}
	return nil
}

// RemapResult revalidates every segment decision against its current literal
// interval. Any shape or quotation uncertainty falls back to ordinary inference.
func RemapResult(old Result, segments []Segment, content Content, runID, versionID int64, at time.Time) (Result, bool) {
	if old.ReturnedModel == StructuredClassifierVersion && old.ProviderResponseID == "local:"+StructuredClassifierVersion && old.Status == "classified" && old.Relevant != nil && *old.Relevant && old.ReasonCode == "regional_warning" && old.ContentComplete && content.Complete && len(old.Segments) == 0 && len(segments) > 0 {
		if quote, ok := CanonicalLiteral(content.Text, old.EvidenceQuote); ok && quote != "" {
			old.RunID, old.DocumentVersionID, old.ContentSHA256, old.CreatedAt = runID, versionID, content.Hash, at
			old.EvidenceQuote = quote
			old.InputTokens, old.OutputTokens, old.CacheReadTokens = nil, nil, nil
			return old, true
		}
		return Result{}, false
	}
	if old.Status != "classified" || old.Relevant == nil || !old.ContentComplete || !content.Complete || len(old.Segments) != len(segments) || len(segments) == 0 {
		return Result{}, false
	}
	result := old
	result.RunID = runID
	result.DocumentVersionID = versionID
	result.ContentSHA256 = content.Hash
	result.CreatedAt = at
	result.InputTokens = nil
	result.OutputTokens = nil
	result.CacheReadTokens = nil
	result.Segments = append([]SegmentResult(nil), old.Segments...)
	for index, segment := range segments {
		prior := &result.Segments[index]
		if prior.Ordinal != segment.Ordinal || prior.Total != segment.Total || prior.ResourceURL != segment.ResourceURL || prior.Role != segment.Role || prior.Page != segment.Page || prior.StartByte != segment.StartByte || prior.EndByte != segment.EndByte {
			return Result{}, false
		}
		body, _ := json.Marshal(Decision{Relevant: prior.Relevant, ReasonCode: prior.ReasonCode, EvidenceQuote: prior.EvidenceQuote})
		if _, err := ParseDecision(string(body), segment.Content()); err != nil {
			return Result{}, false
		}
		prior.DocumentVersionID = versionID
		prior.ContentSHA256 = segment.Hash
		prior.InputTokens = nil
		prior.OutputTokens = nil
		prior.CacheReadTokens = nil
	}
	return result, true
}
