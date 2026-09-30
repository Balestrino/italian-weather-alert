package interpretation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/territory"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/embedding"
	"github.com/Balestrino/italian-weather-alert/internal/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/linking"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
)

var ErrInvalid = errors.New("invalid interpretation schedule")

type Scheduler struct {
	pool        *pgxpool.Pool
	Queue       *jobs.Store
	Policy      inference.RetryPolicy
	Semantic    bool
	Preflight   *Preflight
	Evaluations interface {
		Passed(context.Context, string) (string, error)
	}
}

func New(pool *pgxpool.Pool, queue *jobs.Store, policy inference.RetryPolicy, semantic bool) *Scheduler {
	return &Scheduler{pool: pool, Queue: queue, Policy: policy, Semantic: semantic}
}

type AcquisitionEvent struct {
	DocumentVersionID                 int64
	EvidenceHash, Workload            string
	ContentChanged, DependencyChanged bool
	At                                time.Time
}

func (s *Scheduler) Automatic(ctx context.Context, e AcquisitionEvent) (bool, error) {
	if e.Workload == "" {
		e.Workload = "ordinary"
	}
	if e.DocumentVersionID < 1 || len(e.EvidenceHash) != 64 || e.At.IsZero() {
		return false, ErrInvalid
	}
	if !e.ContentChanged && !e.DependencyChanged {
		return false, nil
	}
	if s == nil || s.pool == nil || s.Queue == nil {
		return false, ErrInvalid
	}
	if err := territory.CheckVersion(ctx, s.pool, e.DocumentVersionID); err != nil {
		if errors.Is(err, territory.ErrDisabled) {
			return false, nil
		}
		return false, err
	}
	if archived, err := s.Archived(ctx, e.DocumentVersionID); err != nil || archived {
		return false, err
	}
	if blocked, err := s.suspended(ctx, e.DocumentVersionID, e.Workload); err != nil || blocked {
		return false, err
	}
	if s.Preflight != nil && e.Workload != "evaluation" && e.Workload != "reprocessing" {
		if _, err := s.Prepare(ctx, s.Preflight, e.DocumentVersionID); err != nil {
			return false, err
		}
	}
	reason := "content_changed"
	if e.DependencyChanged {
		reason = "dependency_changed"
	}
	var jobID *int64
	// A later ordinary check can revisit evidence first seen during bootstrap.
	// Its dependencies belong to the original trigger; changing their workload
	// would conflict with the immutable payload behind the same job identity.
	err := s.pool.QueryRow(ctx, `INSERT INTO interpretation_triggers(document_version_id,evidence_hash,reason,workload,created_at) VALUES($1,$2,$3,$4,$5) ON CONFLICT(document_version_id) DO UPDATE SET document_version_id=EXCLUDED.document_version_id RETURNING classification_job_id,workload`, e.DocumentVersionID, e.EvidenceHash, reason, e.Workload, e.At.UTC()).Scan(&jobID, &e.Workload)
	if err != nil {
		return false, err
	}
	if jobID != nil {
		return false, nil
	}
	// Keep the durable trigger, but do not manufacture one OCR job per PDF
	// copy while its representative still lacks a validated interpretation.
	// ReleaseEquivalent revisits these triggers after the representative finishes.
	if e.Workload != "evaluation" && e.Workload != "reprocessing" {
		if _, waiting, err := s.Equivalent(ctx, e.DocumentVersionID); err != nil || waiting {
			return false, err
		}
	}
	resources, err := s.ocrResources(ctx, e.DocumentVersionID)
	if err != nil {
		return false, err
	}
	if len(resources) == 0 {
		return s.scheduleClassification(ctx, e.DocumentVersionID, e.Workload, e.At)
	}
	for _, resource := range resources {
		if _, err = ocr.Enqueue(ctx, s.Queue, s.Policy, ocr.Payload{DocumentVersionID: e.DocumentVersionID, ResourceURL: resource, Workload: e.Workload}, e.At); err != nil {
			return false, err
		}
	}
	if ready, readyErr := s.AfterOCR(ctx, e.DocumentVersionID, e.Workload, e.At); readyErr != nil {
		return false, readyErr
	} else if ready {
		return true, nil
	}
	return true, nil
}

func (s *Scheduler) ocrResources(ctx context.Context, versionID int64) ([]string, error) {

	version, err := documents.New(s.pool, nil).Version(ctx, versionID)
	if err != nil {
		return nil, err
	}
	var resources []string
	for _, ref := range version.Resources {
		media := strings.ToLower(strings.TrimSpace(strings.Split(ref.MediaType, ";")[0]))
		if ref.InferenceEligible() && ref.Missing == "" && (media == "application/pdf" || media == "image/png" || media == "image/jpeg") {
			resources = append(resources, ref.URL)
		}
	}
	return resources, nil
}

func (s *Scheduler) scheduleClassification(ctx context.Context, versionID int64, workload string, at time.Time) (bool, error) {
	// OCR completion can be observed by a later collection cycle. Preserve the
	// first trigger's accounting identity for the downstream job as well.
	err := s.pool.QueryRow(ctx, "SELECT workload FROM interpretation_triggers WHERE document_version_id=$1", versionID).Scan(&workload)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}
	if blocked, err := s.suspended(ctx, versionID, workload); err != nil || blocked {
		return false, err
	}
	job, err := classification.Enqueue(ctx, s.Queue, s.Policy, classification.Payload{DocumentVersionID: versionID, Workload: workload}, at)
	if err != nil {
		return false, err
	}
	tag, err := s.pool.Exec(ctx, "UPDATE interpretation_triggers SET classification_job_id=$2 WHERE document_version_id=$1 AND classification_job_id IS NULL", versionID, job.ID)
	return err == nil && tag.RowsAffected() == 1, err
}

func (s *Scheduler) AfterOCR(ctx context.Context, versionID int64, workload string, at time.Time) (bool, error) {
	if archived, err := s.Archived(ctx, versionID); err != nil || archived {
		return false, err
	}
	resources, err := s.ocrResources(ctx, versionID)
	if err != nil {
		return false, err
	}
	if len(resources) == 0 {
		return s.scheduleClassification(ctx, versionID, workload, at)
	}
	var completed int
	err = s.pool.QueryRow(ctx, `SELECT count(DISTINCT resource_url) FROM ocr_resource_results WHERE document_version_id=$1 AND resource_url=ANY($2) AND status<>'skipped'`, versionID, resources).Scan(&completed)
	if err != nil || completed != len(resources) {
		return false, err
	}
	return s.scheduleClassification(ctx, versionID, workload, at)
}

type Selection struct {
	SourceID, ErrorCode, Actor, ProcessingEvaluationID string
	From, Through                                      *time.Time
	At                                                 time.Time
}

func (s *Scheduler) Reprocess(ctx context.Context, selection Selection) (string, int, error) {
	if s != nil && s.pool != nil && selection.SourceID != "" {
		if err := territory.CheckExecution(ctx, s.pool, selection.SourceID); err != nil {
			return "", 0, err
		}
	}
	if s == nil || s.pool == nil || s.Queue == nil || strings.TrimSpace(selection.Actor) == "" || selection.At.IsZero() || (selection.SourceID == "" && selection.ErrorCode == "" && selection.From == nil && selection.Through == nil) {
		return "", 0, ErrInvalid
	}
	var processVersion string
	if selection.ProcessingEvaluationID != "" {
		if s.Evaluations == nil {
			return "", 0, ErrInvalid
		}
		var evaluationErr error
		processVersion, evaluationErr = s.Evaluations.Passed(ctx, selection.ProcessingEvaluationID)
		if evaluationErr != nil {
			return "", 0, ErrInvalid
		}
	}
	id := selectionID(selection)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", 0, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO interpretation_reprocessing_requests(id,source_id,acquired_from,acquired_through,error_code,actor,created_at,processing_evaluation_id,candidate_process_version) VALUES($1,NULLIF($2,''),$3,$4,NULLIF($5,''),$6,$7,NULLIF($8,''),NULLIF($9,'')) ON CONFLICT DO NOTHING`, id, selection.SourceID, selection.From, selection.Through, selection.ErrorCode, selection.Actor, selection.At.UTC(), selection.ProcessingEvaluationID, processVersion)
	if err != nil {
		return "", 0, err
	}
	var rows pgx.Rows
	if tag.RowsAffected() == 1 {
		rows, err = tx.Query(ctx, `SELECT DISTINCT v.id FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id WHERE ($1='' OR d.source_id=$1) AND ($2::timestamptz IS NULL OR v.first_acquired_at >= $2) AND ($3::timestamptz IS NULL OR v.first_acquired_at <= $3) AND ($4='' OR EXISTS(SELECT 1 FROM processing_runs r JOIN processing_run_attempts a ON a.run_id=r.id WHERE r.document_version_id=v.id AND a.error_code=$4)) ORDER BY v.id`, selection.SourceID, selection.From, selection.Through, selection.ErrorCode)
	} else {
		rows, err = tx.Query(ctx, `SELECT document_version_id FROM interpretation_reprocessing_selections WHERE request_id=$1 ORDER BY document_version_id`, id)
	}
	if err != nil {
		return "", 0, err
	}
	var versions []int64
	for rows.Next() {
		var versionID int64
		if err = rows.Scan(&versionID); err != nil {
			rows.Close()
			return "", 0, err
		}
		versions = append(versions, versionID)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", 0, err
	}
	for _, versionID := range versions {
		if err := territory.CheckVersion(ctx, tx, versionID); err != nil {
			return "", 0, err
		}
	}
	if tag.RowsAffected() == 1 {
		for _, version := range versions {
			if _, err = tx.Exec(ctx, `INSERT INTO interpretation_reprocessing_selections(request_id,document_version_id) VALUES($1,$2)`, id, version); err != nil {
				return id, 0, err
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return id, 0, err
	}
	count := 0
	for _, version := range versions {
		job, e := classification.Enqueue(ctx, s.Queue, s.Policy, classification.Payload{DocumentVersionID: version, Workload: "reprocessing", SelectionID: id}, selection.At)
		if e != nil {
			return id, count, e
		}
		_, e = s.pool.Exec(ctx, `UPDATE interpretation_reprocessing_selections SET classification_job_id=$3 WHERE request_id=$1 AND document_version_id=$2 AND classification_job_id IS NULL`, id, version, job.ID)
		if e != nil {
			return id, count, e
		}
		count++
	}
	return id, count, nil
}
func (s *Scheduler) AfterClassification(ctx context.Context, r classification.Result, workload string, at time.Time) (bool, error) {
	if r.Status != "classified" || r.Relevant == nil || !*r.Relevant {
		return false, nil
	}
	_, e := extraction.Enqueue(ctx, s.Queue, s.Policy, extraction.Payload{DocumentVersionID: r.DocumentVersionID, ClassificationRunID: r.RunID, Workload: workload}, at)
	return e == nil, e
}
func (s *Scheduler) AfterExtraction(ctx context.Context, r extraction.Result, workload string, at time.Time) (int, error) {
	if r.Status != "extracted" {
		return 0, nil
	}
	// Equivalent revisions retain their own evidence, but downstream public
	// effects and linking work keep the original immutable interpretation ID.
	var canonical int64
	if err := s.pool.QueryRow(ctx, `WITH RECURSIVE lineage(id) AS (
     SELECT $1::bigint UNION ALL SELECT u.original_run_id FROM interpretation_reuse u JOIN lineage l ON l.id=u.run_id
    ) SELECT min(id) FROM lineage`, r.RunID).Scan(&canonical); err != nil {
		return 0, err
	}
	r.RunID = canonical
	count := 0
	for _, m := range r.Measures {
		if s.Semantic {
			if _, e := embedding.Enqueue(ctx, s.Queue, s.Policy, embedding.Payload{ExtractionRunID: r.RunID, MeasureOrdinal: m.Ordinal, Workload: workload}, at); e != nil {
				return count, e
			}
			count++
		} else {
			if _, e := linking.Enqueue(ctx, s.Queue, s.Policy, linking.Payload{ExtractionRunID: r.RunID, MeasureOrdinal: m.Ordinal, Workload: workload}, at); e != nil {
				return count, e
			}
			count++
		}
	}
	return count, nil
}

func (s *Scheduler) AfterEmbedding(ctx context.Context, extractionRunID int64, measureOrdinal int, workload string, at time.Time) (bool, error) {
	_, err := linking.Enqueue(ctx, s.Queue, s.Policy, linking.Payload{ExtractionRunID: extractionRunID, MeasureOrdinal: measureOrdinal, Workload: workload}, at)
	return err == nil, err
}

func (s *Scheduler) OCRHandler(next jobs.Handler) jobs.Handler {
	return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		var input ocr.Payload
		if json.Unmarshal(job.Payload, &input) != nil {
			return next(ctx, job)
		}
		result, err := next(ctx, job)
		if err != nil {
			return result, err
		}
		var output struct {
			DocumentVersionID int64 `json:"document_version_id"`
		}
		if json.Unmarshal(result.Payload, &output) != nil || output.DocumentVersionID != input.DocumentVersionID {
			return jobs.Result{}, &jobs.HandlerError{Failure: jobs.Failure{Code: "ocr_schedule_output_invalid", Detail: "OCR dependency output is invalid", Temporary: false}}
		}
		if _, err = s.AfterOCR(ctx, output.DocumentVersionID, input.Workload, time.Now().UTC()); err != nil {
			return jobs.Result{}, &jobs.HandlerError{Failure: jobs.Failure{Code: "ocr_dependency_schedule_failed", Detail: "OCR dependency scheduling failed", Temporary: true}}
		}
		return result, nil
	}
}

func (s *Scheduler) ClassificationHandler(next jobs.Handler) jobs.Handler {
	return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		var input classification.Payload
		if json.Unmarshal(job.Payload, &input) != nil {
			return next(ctx, job)
		}
		result, err := next(ctx, job)
		if err != nil {
			return result, err
		}
		var output struct {
			RunID             int64  `json:"run_id"`
			DocumentVersionID int64  `json:"document_version_id"`
			Status            string `json:"status"`
			Relevant          *bool  `json:"relevant"`
		}
		if json.Unmarshal(result.Payload, &output) != nil || output.RunID < 1 || output.DocumentVersionID != input.DocumentVersionID {
			return jobs.Result{}, scheduleFailure("classification_schedule_output_invalid", false)
		}
		if _, err = s.AfterClassification(ctx, classification.Result{RunID: output.RunID, DocumentVersionID: output.DocumentVersionID, Status: output.Status, Relevant: output.Relevant}, input.Workload, time.Now().UTC()); err != nil {
			return jobs.Result{}, scheduleFailure("extraction_dependency_schedule_failed", true)
		}
		return result, nil
	}
}

func (s *Scheduler) ExtractionHandler(next jobs.Handler) jobs.Handler {
	return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		var input extraction.Payload
		if json.Unmarshal(job.Payload, &input) != nil {
			return next(ctx, job)
		}
		result, err := next(ctx, job)
		if err != nil {
			return result, err
		}
		var output struct {
			RunID        int64  `json:"run_id"`
			Status       string `json:"status"`
			MeasureCount int    `json:"measure_count"`
		}
		if json.Unmarshal(result.Payload, &output) != nil || output.RunID < 1 || output.MeasureCount < 0 {
			return jobs.Result{}, scheduleFailure("extraction_schedule_output_invalid", false)
		}
		interpreted := extraction.Result{RunID: output.RunID, Status: output.Status}
		for ordinal := 1; ordinal <= output.MeasureCount; ordinal++ {
			interpreted.Measures = append(interpreted.Measures, extraction.Measure{Ordinal: ordinal})
		}
		if _, err = s.AfterExtraction(ctx, interpreted, input.Workload, time.Now().UTC()); err != nil {
			return jobs.Result{}, scheduleFailure("linking_dependency_schedule_failed", true)
		}
		return result, nil
	}
}

func (s *Scheduler) EmbeddingHandler(next jobs.Handler) jobs.Handler {
	return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		var input embedding.Payload
		if json.Unmarshal(job.Payload, &input) != nil {
			return next(ctx, job)
		}
		result, err := next(ctx, job)
		if err != nil {
			return result, err
		}
		var output struct {
			ExtractionRunID int64 `json:"extraction_run_id"`
			MeasureOrdinal  int   `json:"measure_ordinal"`
		}
		if json.Unmarshal(result.Payload, &output) != nil || output.ExtractionRunID != input.ExtractionRunID || output.MeasureOrdinal != input.MeasureOrdinal {
			return jobs.Result{}, scheduleFailure("embedding_schedule_output_invalid", false)
		}
		if _, err = s.AfterEmbedding(ctx, output.ExtractionRunID, output.MeasureOrdinal, input.Workload, time.Now().UTC()); err != nil {
			return jobs.Result{}, scheduleFailure("semantic_linking_dependency_schedule_failed", true)
		}
		return result, nil
	}
}

func scheduleFailure(code string, temporary bool) error {
	return &jobs.HandlerError{Failure: jobs.Failure{Code: code, Detail: strings.ReplaceAll(code, "_", " "), Temporary: temporary}}
}
func selectionID(s Selection) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%v\x00%v\x00%s\x00%s\x00%s", s.SourceID, s.ErrorCode, s.From, s.Through, s.Actor, s.At.UTC().Format(time.RFC3339Nano), s.ProcessingEvaluationID)))
	return "reprocess-" + hex.EncodeToString(h[:])[:20]
}
