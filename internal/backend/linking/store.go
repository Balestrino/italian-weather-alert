package linking

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Current(ctx context.Context, runID int64, ordinal int) (MeasureContext, error) {
	var value MeasureContext
	var place *string
	err := s.pool.QueryRow(ctx, `SELECT m.run_id,e.document_version_id,m.ordinal,p.source_id,rs.territory,m.kind,m.subject,m.place,d.official_url,v.first_acquired_at
	FROM extracted_measures m JOIN extraction_results e ON e.run_id=m.run_id JOIN processing_runs p ON p.id=m.run_id JOIN retained_versions v ON v.id=e.document_version_id JOIN retained_documents d ON d.id=v.document_id JOIN registry_sources rs ON rs.id=p.source_id
	WHERE m.run_id=$1 AND m.ordinal=$2`, runID, ordinal).Scan(&value.RunID, &value.DocumentVersionID, &value.Ordinal, &value.SourceID, &value.Municipality, &value.Kind, &value.Subject, &place, &value.OfficialURL, &value.FirstAcquiredAt)
	if err != nil {
		return MeasureContext{}, err
	}
	if place != nil {
		value.Place = *place
	}
	value.Evidence, err = s.evidence(ctx, runID, ordinal)
	return value, err
}

func (s *Store) Candidates(ctx context.Context, current MeasureContext, limit int) ([]Candidate, error) {
	if current.RunID < 1 || current.Ordinal < 1 || current.Municipality == "" || limit < 1 || limit > 100 {
		return nil, ErrInvalid
	}
	query := strings.TrimSpace(current.Subject + " " + current.Place)
	rows, err := s.pool.Query(ctx, `SELECT m.run_id,e.document_version_id,m.ordinal,p.source_id,rs.territory,m.kind,m.subject,m.place,d.official_url,v.first_acquired_at,
	 ts_rank_cd(to_tsvector('simple',m.subject || ' ' || COALESCE(m.place,'')),plainto_tsquery('simple',$3)) AS rank,
	 ($4 <> '' AND lower(COALESCE(m.place,''))=lower($4)) AS exact_place
	FROM extracted_measures m JOIN extraction_results e ON e.run_id=m.run_id JOIN processing_runs p ON p.id=m.run_id JOIN registry_sources rs ON rs.id=p.source_id JOIN retained_versions v ON v.id=e.document_version_id JOIN retained_documents d ON d.id=v.document_id
	WHERE rs.territory=$1 AND NOT (m.run_id=$2 AND m.ordinal=$5) AND (v.first_acquired_at < $6 OR (v.first_acquired_at=$6 AND e.document_version_id < $7))
	 AND (($4 <> '' AND lower(COALESCE(m.place,''))=lower($4)) OR to_tsvector('simple',m.subject || ' ' || COALESCE(m.place,'')) @@ plainto_tsquery('simple',$3))
	ORDER BY exact_place DESC,rank DESC,v.first_acquired_at DESC,m.run_id DESC,m.ordinal LIMIT $8`, current.Municipality, current.RunID, query, current.Place, current.Ordinal, current.FirstAcquiredAt, current.DocumentVersionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var candidates []Candidate
	for rows.Next() {
		var candidate Candidate
		var place *string
		if err = rows.Scan(&candidate.RunID, &candidate.DocumentVersionID, &candidate.Ordinal, &candidate.SourceID, &candidate.Municipality, &candidate.Kind, &candidate.Subject, &place, &candidate.OfficialURL, &candidate.FirstAcquiredAt, &candidate.TextRank, &candidate.ExactPlace); err != nil {
			return nil, err
		}
		if place != nil {
			candidate.Place = *place
		}
		candidate.Evidence, err = s.evidence(ctx, candidate.RunID, candidate.Ordinal)
		if err != nil {
			return nil, err
		}
		candidate.RetrievalMethod = "baseline_text"
		candidates = append(candidates, candidate)
	}
	return candidates, rows.Err()
}

func (s *Store) SemanticCandidates(ctx context.Context, current MeasureContext, configuration string, limit int) ([]Candidate, error) {
	if configuration == "" || limit < 1 || limit > 100 {
		return nil, ErrInvalid
	}
	var currentVector []float32
	if err := s.pool.QueryRow(ctx, `SELECT vector FROM measure_embeddings WHERE extraction_run_id=$1 AND measure_ordinal=$2 AND configuration_version_id=$3`, current.RunID, current.Ordinal, configuration).Scan(&currentVector); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT m.extraction_run_id,m.measure_ordinal,m.vector FROM measure_embeddings m JOIN processing_runs p ON p.id=m.extraction_run_id JOIN registry_sources s ON s.id=p.source_id JOIN extraction_results e ON e.run_id=m.extraction_run_id JOIN retained_versions v ON v.id=e.document_version_id WHERE m.configuration_version_id=$1 AND s.territory=$2 AND NOT(m.extraction_run_id=$3 AND m.measure_ordinal=$4) AND (v.first_acquired_at<$5 OR (v.first_acquired_at=$5 AND e.document_version_id<$6))`, configuration, current.Municipality, current.RunID, current.Ordinal, current.FirstAcquiredAt, current.DocumentVersionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type scored struct {
		run     int64
		ordinal int
		score   float32
	}
	var values []scored
	for rows.Next() {
		var run int64
		var ordinal int
		var vector []float32
		if err = rows.Scan(&run, &ordinal, &vector); err != nil {
			return nil, err
		}
		if len(vector) != len(currentVector) {
			continue
		}
		var dot float64
		for i := range vector {
			dot += float64(vector[i] * currentVector[i])
		}
		values = append(values, scored{run, ordinal, float32(math.Max(-1, math.Min(1, dot)))})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(values, func(i, j int) bool { return values[i].score > values[j].score })
	if len(values) > limit {
		values = values[:limit]
	}
	result := make([]Candidate, 0, len(values))
	for _, item := range values {
		contextValue, e := s.Current(ctx, item.run, item.ordinal)
		if e != nil {
			return nil, e
		}
		score := item.score
		result = append(result, Candidate{MeasureContext: contextValue, RetrievalMethod: "semantic", SemanticScore: &score})
	}
	return result, nil
}

func (s *Store) evidence(ctx context.Context, runID int64, ordinal int) ([]extraction.Evidence, error) {
	rows, err := s.pool.Query(ctx, `SELECT field_name,resource_url,page_number,quote,COALESCE(segment_ordinal,0) FROM extraction_evidence WHERE run_id=$1 AND measure_ordinal=$2 ORDER BY ordinal`, runID, ordinal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []extraction.Evidence
	for rows.Next() {
		var value extraction.Evidence
		if err = rows.Scan(&value.Field, &value.ResourceURL, &value.Page, &value.Quote, &value.SegmentOrdinal); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func (s *Store) Put(ctx context.Context, result Result) error {
	if result.RunID < 1 || result.CurrentRunID < 1 || result.CurrentOrdinal < 1 || result.Status == "" || result.ReasonCode == "" || result.CreatedAt.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO linking_results(run_id,current_extraction_run_id,current_measure_ordinal,status,reason_code,relation,candidate_extraction_run_id,candidate_measure_ordinal,provider_response_id,returned_model,input_tokens,output_tokens,cache_read_tokens,created_at)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,''),NULLIF($10,''),$11,$12,$13,$14) ON CONFLICT DO NOTHING`, result.RunID, result.CurrentRunID, result.CurrentOrdinal, result.Status, result.ReasonCode, result.Relation, result.CandidateRunID, result.CandidateOrdinal, result.ProviderResponseID, result.ReturnedModel, result.InputTokens, result.OutputTokens, result.CacheReadTokens, result.CreatedAt.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var status string
		if err = tx.QueryRow(ctx, "SELECT status FROM linking_results WHERE run_id=$1", result.RunID).Scan(&status); err != nil {
			return err
		}
		if status != result.Status {
			return ErrConflict
		}
		return tx.Commit(ctx)
	}
	for i, c := range result.Candidates {
		method := c.RetrievalMethod
		if method == "" {
			method = "baseline_text"
		}
		if _, err = tx.Exec(ctx, `INSERT INTO linking_candidates(run_id,ordinal,candidate_extraction_run_id,candidate_measure_ordinal,retrieval_method,text_rank,exact_place,semantic_score) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, result.RunID, i+1, c.RunID, c.Ordinal, method, c.TextRank, c.ExactPlace, c.SemanticScore); err != nil {
			return err
		}
	}
	for side, values := range map[string][]extraction.Evidence{"current": result.CurrentEvidence, "candidate": result.CandidateEvidence} {
		for i, e := range values {
			if _, err = tx.Exec(ctx, `INSERT INTO linking_evidence(run_id,side,ordinal,field_name,resource_url,page_number,quote) VALUES($1,$2,$3,$4,$5,$6,$7)`, result.RunID, side, i+1, e.Field, e.ResourceURL, e.Page, e.Quote); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) Get(ctx context.Context, runID int64) (Result, bool, error) {
	var r Result
	err := s.pool.QueryRow(ctx, `SELECT run_id,current_extraction_run_id,current_measure_ordinal,status,reason_code,relation,candidate_extraction_run_id,candidate_measure_ordinal,COALESCE(provider_response_id,''),COALESCE(returned_model,''),input_tokens,output_tokens,cache_read_tokens,created_at FROM linking_results WHERE run_id=$1`, runID).Scan(&r.RunID, &r.CurrentRunID, &r.CurrentOrdinal, &r.Status, &r.ReasonCode, &r.Relation, &r.CandidateRunID, &r.CandidateOrdinal, &r.ProviderResponseID, &r.ReturnedModel, &r.InputTokens, &r.OutputTokens, &r.CacheReadTokens, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, false, nil
	}
	return r, err == nil, err
}

func candidateSubject(candidates []Candidate) (json.RawMessage, string) {
	type identity struct {
		RunID   int64 `json:"run_id"`
		Ordinal int   `json:"ordinal"`
	}
	ids := make([]identity, 0, len(candidates))
	for _, candidate := range candidates {
		ids = append(ids, identity{candidate.RunID, candidate.Ordinal})
	}
	body, _ := json.Marshal(map[string]any{"candidates": ids, "retrieval": "baseline_text"})
	hash := sha256.Sum256(body)
	return body, hex.EncodeToString(hash[:])[:16]
}
