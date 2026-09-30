package extraction

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Put(ctx context.Context, result Result) error {
	if result.RunID < 1 || result.DocumentVersionID < 1 || result.ClassificationRunID < 1 || result.Status == "" || result.ReasonCode == "" || result.ContentSHA256 == "" || result.CreatedAt.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO extraction_results(run_id,document_version_id,classification_run_id,status,reason_code,content_sha256,content_complete,provider_response_id,returned_model,input_tokens,output_tokens,cache_read_tokens,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,NULLIF($8,''),NULLIF($9,''),$10,$11,$12,$13) ON CONFLICT DO NOTHING`, result.RunID, result.DocumentVersionID, result.ClassificationRunID, result.Status, result.ReasonCode, result.ContentSHA256, result.ContentComplete, result.ProviderResponseID, result.ReturnedModel, result.InputTokens, result.OutputTokens, result.CacheReadTokens, result.CreatedAt.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var hash, status string
		if err = tx.QueryRow(ctx, "SELECT content_sha256,status FROM extraction_results WHERE run_id=$1", result.RunID).Scan(&hash, &status); err != nil {
			return err
		}
		if hash != result.ContentSHA256 || status != result.Status {
			return ErrConflict
		}
		return tx.Commit(ctx)
	}
	for _, segment := range result.Segments {
		if segment.Ordinal < 1 || segment.Total < segment.Ordinal || segment.DocumentVersionID != result.DocumentVersionID || segment.CoreStartByte < segment.StartByte || segment.CoreEndByte <= segment.CoreStartByte || segment.CoreEndByte > segment.EndByte || !validWindowOffsetMap(segment.Page, segment.StartByte, segment.EndByte, segment.SourceStartByte, segment.SourceEndByte, segment.NormalizationMap) {
			return ErrInvalid
		}
		mappingValue := segment.NormalizationMap
		if mappingValue == nil {
			mappingValue = []WindowOffset{}
		}
		mapping, marshalErr := json.Marshal(mappingValue)
		if marshalErr != nil {
			return ErrInvalid
		}
		if _, err = tx.Exec(ctx, `INSERT INTO extraction_segments(run_id,ordinal,total,document_version_id,resource_url,role,page_number,start_byte,end_byte,core_start_byte,core_end_byte,source_start_byte,source_end_byte,normalization_map,content_sha256,response_sha256,provider_response_id,returned_model,measure_count,input_tokens,output_tokens,cache_read_tokens)
 VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,0),$8,$9,$10,$11,CASE WHEN $12=0 AND $13=0 THEN NULL ELSE $12 END,CASE WHEN $12=0 AND $13=0 THEN NULL ELSE $13 END,$14,$15,$16,$17,$18,$19,$20,$21,$22)`, result.RunID, segment.Ordinal, segment.Total, segment.DocumentVersionID, segment.ResourceURL, segment.Role, segment.Page, segment.StartByte, segment.EndByte, segment.CoreStartByte, segment.CoreEndByte, segment.SourceStartByte, segment.SourceEndByte, mapping, segment.ContentSHA256, segment.ResponseSHA256, segment.ProviderResponseID, segment.ReturnedModel, segment.MeasureCount, segment.InputTokens, segment.OutputTokens, segment.CacheReadTokens); err != nil {
			return err
		}
	}
	for _, rawMeasure := range result.Measures {
		measure := rawMeasure
		if len(measure.TemporalCandidates) > 0 {
			if err = validateTemporalProjection(measure); err != nil {
				return err
			}
		}
		fields, marshalErr := json.Marshal(measure.IndeterminateFields)
		if marshalErr != nil {
			return ErrInvalid
		}
		if _, err = tx.Exec(ctx, `INSERT INTO extracted_measures(run_id,ordinal,kind,subject,place,valid_from_expression,valid_until_expression,indeterminate_fields)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, result.RunID, measure.Ordinal, measure.Kind, measure.Subject, measure.Place, measure.ValidFrom, measure.ValidUntil, fields); err != nil {
			return err
		}
		evidenceOrdinals := make(map[string]int, len(measure.Evidence))
		for index, evidence := range measure.Evidence {
			if _, err = tx.Exec(ctx, `INSERT INTO extraction_evidence(run_id,measure_ordinal,ordinal,field_name,resource_url,page_number,quote,segment_ordinal)
 VALUES($1,$2,$3,$4,$5,$6,$7,NULLIF($8,0))`, result.RunID, measure.Ordinal, index+1, evidence.Field, evidence.ResourceURL, evidence.Page, evidence.Quote, evidence.SegmentOrdinal); err != nil {
				return err
			}
			evidenceOrdinals[evidenceKey(evidence)] = index + 1
		}
		for candidateIndex, candidate := range measure.TemporalCandidates {
			ordinal := candidateIndex + 1
			if candidate.Ordinal != ordinal {
				return ErrInvalid
			}
			if _, err = tx.Exec(ctx, `INSERT INTO extracted_temporal_candidates(run_id,measure_ordinal,ordinal,field_name,original_expression,conflict_identity)
 VALUES($1,$2,$3,$4,$5,$6)`, result.RunID, measure.Ordinal, ordinal, candidate.Field, candidate.OriginalExpression, candidate.ConflictIdentity); err != nil {
				return err
			}
			for evidenceIndex, evidence := range candidate.Evidence {
				evidenceOrdinal, ok := evidenceOrdinals[evidenceKey(evidence)]
				if !ok {
					return ErrEvidence
				}
				if _, err = tx.Exec(ctx, `INSERT INTO extraction_temporal_candidate_evidence(run_id,measure_ordinal,candidate_ordinal,ordinal,evidence_ordinal)
 VALUES($1,$2,$3,$4,$5)`, result.RunID, measure.Ordinal, ordinal, evidenceIndex+1, evidenceOrdinal); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) Get(ctx context.Context, runID int64) (Result, bool, error) {
	var result Result
	err := s.pool.QueryRow(ctx, `SELECT run_id,document_version_id,classification_run_id,status,reason_code,content_sha256,content_complete,COALESCE(provider_response_id,''),COALESCE(returned_model,''),input_tokens,output_tokens,cache_read_tokens,created_at
 FROM extraction_results WHERE run_id=$1`, runID).Scan(&result.RunID, &result.DocumentVersionID, &result.ClassificationRunID, &result.Status, &result.ReasonCode, &result.ContentSHA256, &result.ContentComplete, &result.ProviderResponseID, &result.ReturnedModel, &result.InputTokens, &result.OutputTokens, &result.CacheReadTokens, &result.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, false, nil
	}
	if err != nil {
		return Result{}, false, err
	}
	segmentRows, err := s.pool.Query(ctx, `SELECT ordinal,total,document_version_id,resource_url,role,COALESCE(page_number,0),start_byte,end_byte,COALESCE(core_start_byte,start_byte),COALESCE(core_end_byte,end_byte),COALESCE(source_start_byte,0),COALESCE(source_end_byte,0),normalization_map,content_sha256,response_sha256,provider_response_id,returned_model,measure_count,input_tokens,output_tokens,cache_read_tokens FROM extraction_segments WHERE run_id=$1 ORDER BY ordinal`, runID)
	if err != nil {
		return Result{}, false, err
	}
	for segmentRows.Next() {
		var segment SegmentResult
		var mapping []byte
		if err = segmentRows.Scan(&segment.Ordinal, &segment.Total, &segment.DocumentVersionID, &segment.ResourceURL, &segment.Role, &segment.Page, &segment.StartByte, &segment.EndByte, &segment.CoreStartByte, &segment.CoreEndByte, &segment.SourceStartByte, &segment.SourceEndByte, &mapping, &segment.ContentSHA256, &segment.ResponseSHA256, &segment.ProviderResponseID, &segment.ReturnedModel, &segment.MeasureCount, &segment.InputTokens, &segment.OutputTokens, &segment.CacheReadTokens); err != nil {
			segmentRows.Close()
			return Result{}, false, err
		}
		if json.Unmarshal(mapping, &segment.NormalizationMap) != nil {
			segmentRows.Close()
			return Result{}, false, ErrInvalid
		}
		result.Segments = append(result.Segments, segment)
	}
	if err = segmentRows.Err(); err != nil {
		segmentRows.Close()
		return Result{}, false, err
	}
	segmentRows.Close()
	rows, err := s.pool.Query(ctx, `SELECT ordinal,kind,subject,place,valid_from_expression,valid_until_expression,indeterminate_fields FROM extracted_measures WHERE run_id=$1 ORDER BY ordinal`, runID)
	if err != nil {
		return Result{}, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var measure Measure
		var fields []byte
		if err = rows.Scan(&measure.Ordinal, &measure.Kind, &measure.Subject, &measure.Place, &measure.ValidFrom, &measure.ValidUntil, &fields); err != nil {
			return Result{}, false, err
		}
		if json.Unmarshal(fields, &measure.IndeterminateFields) != nil {
			return Result{}, false, ErrInvalid
		}
		evidenceRows, queryErr := s.pool.Query(ctx, `SELECT field_name,resource_url,page_number,quote,COALESCE(segment_ordinal,0) FROM extraction_evidence WHERE run_id=$1 AND measure_ordinal=$2 ORDER BY ordinal`, runID, measure.Ordinal)
		if queryErr != nil {
			return Result{}, false, queryErr
		}
		for evidenceRows.Next() {
			var evidence Evidence
			if queryErr = evidenceRows.Scan(&evidence.Field, &evidence.ResourceURL, &evidence.Page, &evidence.Quote, &evidence.SegmentOrdinal); queryErr != nil {
				evidenceRows.Close()
				return Result{}, false, queryErr
			}
			measure.Evidence = append(measure.Evidence, evidence)
		}
		queryErr = evidenceRows.Err()
		evidenceRows.Close()
		if queryErr != nil {
			return Result{}, false, queryErr
		}
		candidateRows, queryErr := s.pool.Query(ctx, `SELECT c.ordinal,c.field_name,c.original_expression,c.conflict_identity,e.field_name,e.resource_url,e.page_number,e.quote,COALESCE(e.segment_ordinal,0)
 FROM extracted_temporal_candidates c
 JOIN extraction_temporal_candidate_evidence ce ON ce.run_id=c.run_id AND ce.measure_ordinal=c.measure_ordinal AND ce.candidate_ordinal=c.ordinal
 JOIN extraction_evidence e ON e.run_id=ce.run_id AND e.measure_ordinal=ce.measure_ordinal AND e.ordinal=ce.evidence_ordinal
 WHERE c.run_id=$1 AND c.measure_ordinal=$2 ORDER BY c.ordinal,ce.ordinal`, runID, measure.Ordinal)
		if queryErr != nil {
			return Result{}, false, queryErr
		}
		for candidateRows.Next() {
			var ordinal int
			var field, expression string
			var conflict *string
			var evidence Evidence
			if queryErr = candidateRows.Scan(&ordinal, &field, &expression, &conflict, &evidence.Field, &evidence.ResourceURL, &evidence.Page, &evidence.Quote, &evidence.SegmentOrdinal); queryErr != nil {
				candidateRows.Close()
				return Result{}, false, queryErr
			}
			if len(measure.TemporalCandidates) == 0 || measure.TemporalCandidates[len(measure.TemporalCandidates)-1].Ordinal != ordinal {
				measure.TemporalCandidates = append(measure.TemporalCandidates, TemporalCandidate{Ordinal: ordinal, Field: field, OriginalExpression: expression, ConflictIdentity: conflict})
			}
			measure.TemporalCandidates[len(measure.TemporalCandidates)-1].Evidence = append(measure.TemporalCandidates[len(measure.TemporalCandidates)-1].Evidence, evidence)
		}
		queryErr = candidateRows.Err()
		candidateRows.Close()
		if queryErr != nil {
			return Result{}, false, queryErr
		}
		if len(measure.TemporalCandidates) > 0 {
			if queryErr = validateTemporalProjection(measure); queryErr != nil {
				return Result{}, false, queryErr
			}
		}
		result.Measures = append(result.Measures, measure)
	}
	return result, true, rows.Err()
}

// LatestVisibility keeps acquisition identity and the newest interpretation
// state together. An uninterpreted revision therefore cannot disappear behind
// an older successful extraction.
func (s *Store) LatestVisibility(ctx context.Context, documentID int64) (Visibility, bool, error) {
	var value Visibility
	err := s.pool.QueryRow(ctx, `SELECT d.id,v.id,d.official_url,v.first_acquired_at,COALESCE(r.run_id,0),COALESCE(r.status,'uninterpreted'),COALESCE(r.reason_code,'not_processed'),r.created_at
	FROM retained_documents d JOIN retained_versions v ON v.document_id=d.id
	LEFT JOIN LATERAL (SELECT run_id,status,reason_code,created_at FROM extraction_results WHERE document_version_id=v.id ORDER BY created_at DESC,run_id DESC LIMIT 1) r ON true
	WHERE d.id=$1 ORDER BY v.first_acquired_at DESC,v.id DESC LIMIT 1`, documentID).Scan(&value.DocumentID, &value.DocumentVersionID, &value.OfficialURL, &value.FirstAcquiredAt, &value.RunID, &value.Status, &value.ReasonCode, &value.InterpretedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visibility{}, false, nil
	}
	return value, err == nil, err
}
