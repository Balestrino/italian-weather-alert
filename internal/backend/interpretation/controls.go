package interpretation

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"net/url"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/embedding"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5"
)

type Recovery struct {
	SuspensionID          int64             `json:"suspension_id"`
	Correction            registry.Evidence `json:"correction"`
	Validation            registry.Evidence `json:"validation"`
	ReprocessingID        string            `json:"reprocessing_id"`
	KnownOmissions        int               `json:"known_omissions"`
	UnsupportedAssertions int               `json:"unsupported_assertions"`
	Verified              bool              `json:"verified"`
}

// ResumeInterpretation validates persisted reprocessing, not a client-supplied
// success flag. The report attests semantic evaluation which SQL cannot prove.
func (s *Scheduler) ResumeInterpretation(ctx context.Context, source string, revision int, actor string, recovery Recovery) error {
	validEvidence := func(e registry.Evidence) bool {
		u, err := url.Parse(e.URL)
		return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && u.User == nil && u.Fragment == "" && strings.TrimSpace(e.Locator) != "" && !e.ObservedAt.IsZero()
	}
	if strings.TrimSpace(actor) == "" || recovery.SuspensionID < 1 || !validEvidence(recovery.Correction) || !validEvidence(recovery.Validation) || !recovery.Verified || recovery.KnownOmissions != 0 || recovery.UnsupportedAssertions != 0 || recovery.Validation.ObservedAt.Before(recovery.Correction.ObservedAt) || recovery.Validation.ObservedAt.After(time.Now()) {
		return registry.ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var active *int
	var suspended *time.Time
	if err = tx.QueryRow(ctx, `SELECT active_revision,interpretation_suspended_at FROM registry_sources WHERE id=$1 FOR UPDATE`, source).Scan(&active, &suspended); errors.Is(err, pgx.ErrNoRows) {
		return registry.ErrNotFound
	} else if err != nil {
		return err
	}
	if active == nil || *active != revision || suspended == nil {
		return registry.ErrConflict
	}
	var suspensionID int64
	if err = tx.QueryRow(ctx, `SELECT id FROM registry_interpretation_suspensions WHERE source_id=$1 AND resumed_at IS NULL`, source).Scan(&suspensionID); err != nil {
		return err
	}
	if suspensionID != recovery.SuspensionID || recovery.Correction.ObservedAt.Before(*suspended) {
		return registry.ErrPrerequisite
	}
	var ready bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(
 SELECT 1 FROM interpretation_reprocessing_requests r WHERE r.id=$1 AND r.source_id=$2 AND r.created_at>=$3 AND r.created_at>=$4 AND r.created_at<=$5
 AND EXISTS(SELECT 1 FROM interpretation_reprocessing_selections x WHERE x.request_id=r.id)
 -- Do not release a source while an older in-flight job can still write results.
 AND NOT EXISTS(SELECT 1 FROM processing_jobs running WHERE running.state='running' AND EXISTS(
  SELECT 1 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id WHERE d.source_id=$2
  AND (running.payload->>'document_version_id'=v.id::text OR running.payload->>'extraction_run_id' IN(SELECT e.run_id::text FROM extraction_results e WHERE e.document_version_id=v.id))))
 AND NOT EXISTS(
  SELECT 1 FROM interpretation_reprocessing_selections x
  LEFT JOIN processing_jobs j ON j.id=x.classification_job_id
  LEFT JOIN classification_results c ON c.run_id=(j.result->>'run_id')::bigint
  WHERE x.request_id=r.id AND (j.state IS DISTINCT FROM 'succeeded' OR j.completed_at>$5 OR c.status IS DISTINCT FROM 'classified'
   OR EXISTS(SELECT 1 FROM processing_jobs child WHERE
    (child.payload->>'classification_run_id'=c.run_id::text OR child.payload->>'extraction_run_id' IN(SELECT e.run_id::text FROM extraction_results e WHERE e.classification_run_id=c.run_id))
    AND (child.state<>'succeeded' OR child.completed_at>$5))
   OR (c.relevant AND NOT EXISTS(SELECT 1 FROM extraction_results e WHERE e.classification_run_id=c.run_id AND e.status='extracted' AND e.created_at<=$5
    AND NOT EXISTS(SELECT 1 FROM extracted_measures m WHERE m.run_id=e.run_id AND NOT EXISTS(
     SELECT 1 FROM linking_results l WHERE l.current_extraction_run_id=m.run_id AND l.current_measure_ordinal=m.ordinal AND l.created_at<=$5 AND l.reason_code NOT IN('provider_error','attempts_exhausted','output_invalid'))))))))`, recovery.ReprocessingID, source, *suspended, recovery.Correction.ObservedAt, recovery.Validation.ObservedAt).Scan(&ready)
	if err != nil {
		return err
	}
	if !ready {
		return registry.ErrPrerequisite
	}
	raw, err := json.Marshal(recovery)
	if err != nil {
		return registry.ErrInvalid
	}
	if _, err = tx.Exec(ctx, `UPDATE registry_interpretation_suspensions SET resumed_at=clock_timestamp(),resumed_by=$2,recovery=$3 WHERE id=$1`, suspensionID, actor, raw); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE registry_sources SET interpretation_suspended_at=NULL WHERE id=$1`, source); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Scheduler) suspended(ctx context.Context, versionID int64, workload string) (bool, error) {
	// Operator-selected correction runs remain private until validated resumption.
	if workload == "reprocessing" || workload == "evaluation" {
		return false, nil
	}
	var suspended bool
	err := s.pool.QueryRow(ctx, `SELECT s.interpretation_suspended_at IS NOT NULL FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id JOIN registry_sources s ON s.id=d.source_id WHERE v.id=$1`, versionID).Scan(&suspended)
	return suspended, err
}

// Guard also covers jobs queued before suspension. A blocked job finishes with
// an explicit outcome and no inference/effect; the operator selects its replay.
func (s *Scheduler) Guard(next jobs.Handler) jobs.Handler {
	return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
		var input struct {
			DocumentVersionID   int64  `json:"document_version_id"`
			ClassificationRunID int64  `json:"classification_run_id"`
			ExtractionRunID     int64  `json:"extraction_run_id"`
			SelectionID         string `json:"selection_id"`
			Workload            string `json:"workload"`
		}
		if json.Unmarshal(job.Payload, &input) != nil {
			return jobs.Result{}, scheduleFailure("interpretation_payload_invalid", false)
		}
		if input.DocumentVersionID == 0 && input.ExtractionRunID > 0 {
			if err := s.pool.QueryRow(ctx, `SELECT document_version_id FROM extraction_results WHERE run_id=$1`, input.ExtractionRunID).Scan(&input.DocumentVersionID); err != nil {
				return jobs.Result{}, scheduleFailure("interpretation_source_unavailable", true)
			}
		}
		if err := territory.CheckVersion(ctx, s.pool, input.DocumentVersionID); err != nil {
			return jobs.Result{}, &jobs.DeferredError{Until: time.Now().Add(time.Minute), Code: "territory_unavailable"}
		}
		archived, archiveErr := s.Archived(ctx, input.DocumentVersionID)
		if archiveErr != nil {
			return jobs.Result{}, scheduleFailure("interpretation_archive_unavailable", true)
		}
		if archived && input.Workload == "reprocessing" {
			selection := ""
			var classificationRunID, extractionRunID int64
			if job.Kind == classification.Kind {
				selection = input.SelectionID
			} else if job.Kind == extraction.Kind {
				classificationRunID = input.ClassificationRunID
			} else if job.Kind == embedding.Kind || job.Kind == linking.Kind {
				extractionRunID = input.ExtractionRunID
			}
			allowed, err := s.archiveRecoveryAllowed(ctx, input.DocumentVersionID, selection, classificationRunID, extractionRunID)
			if err != nil {
				return jobs.Result{}, scheduleFailure("interpretation_archive_unavailable", true)
			}
			archived = !allowed
		}
		if archived {
			return jobs.Result{Payload: json.RawMessage(`{"status":"interpretation_archived"}`)}, nil
		}
		blocked, err := s.suspended(ctx, input.DocumentVersionID, input.Workload)
		if err != nil {
			return jobs.Result{}, scheduleFailure("interpretation_source_unavailable", true)
		}
		if blocked {
			return jobs.Result{Payload: json.RawMessage(`{"status":"interpretation_suspended"}`)}, nil
		}
		if input.Workload != "evaluation" && input.Workload != "reprocessing" {
			equivalent, waiting, err := s.Equivalent(ctx, input.DocumentVersionID)
			if err != nil {
				return jobs.Result{}, scheduleFailure("equivalence_unavailable", true)
			}
			if waiting {
				return jobs.Result{}, &jobs.DeferredError{Until: time.Now().Add(time.Minute), Code: "equivalent_input_waiting"}
			}
			if equivalent {
				ctx = inference.LocalReuseOnly(ctx)
			}
		}
		return next(ctx, job)
	}
}
