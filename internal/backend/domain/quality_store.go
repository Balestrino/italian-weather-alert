package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Store) RecordProvenance(ctx context.Context, value ProvenanceAssessment) error {
	if value.SourceID == "" || value.Configuration < 1 || !slicesContains([]string{"verified", "unresolved"}, value.State) || strings.TrimSpace(value.EvidenceURL) == "" || strings.TrimSpace(value.EvidenceLocator) == "" || value.AssessedAt.IsZero() {
		return ErrInvalid
	}
	limitations, err := json.Marshal(value.Limitations)
	if err != nil {
		return ErrInvalid
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO domain_provenance_events(source_id,configuration,state,evidence_url,evidence_locator,limitations,assessed_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, value.SourceID, value.Configuration, value.State, value.EvidenceURL, value.EvidenceLocator, limitations, value.AssessedAt.UTC())
	return err
}

func (s *Store) RecordInterpretation(ctx context.Context, value InterpretationAssessment) error {
	if value.MeasureID == "" || !slicesContains([]string{"supported", "partial", "failed", "unreliable", "not_processed"}, value.State) || value.EvidenceDocumentVersionID < 1 || strings.TrimSpace(value.Reason) == "" || strings.TrimSpace(value.Actor) == "" || value.RecordedAt.IsZero() {
		return ErrInvalid
	}
	var versionID int64
	if err := s.pool.QueryRow(ctx, "SELECT document_version_id FROM domain_local_measures WHERE id=$1", value.MeasureID).Scan(&versionID); err != nil {
		return err
	}
	if versionID != value.EvidenceDocumentVersionID {
		return ErrInvalid
	}
	limitations, err := json.Marshal(value.Limitations)
	if err != nil {
		return ErrInvalid
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO domain_interpretation_events(local_measure_id,state,evidence_document_version_id,reason,limitations,actor,recorded_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, value.MeasureID, value.State, value.EvidenceDocumentVersionID, value.Reason, limitations, value.Actor, value.RecordedAt.UTC())
	return err
}

func (s *Store) MarkInterpretationsUnreliable(ctx context.Context, measureIDs []string, reason, actor string, at time.Time) error {
	if len(measureIDs) == 0 || strings.TrimSpace(reason) == "" || strings.TrimSpace(actor) == "" || at.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	seen := map[string]bool{}
	for _, measureID := range measureIDs {
		if measureID == "" || seen[measureID] {
			return ErrInvalid
		}
		seen[measureID] = true
		var versionID int64
		if err = tx.QueryRow(ctx, "SELECT document_version_id FROM domain_local_measures WHERE id=$1", measureID).Scan(&versionID); err != nil {
			return err
		}
		limitations, _ := json.Marshal([]string{"confirmed interpretation defect: " + reason})
		if _, err = tx.Exec(ctx, `INSERT INTO domain_interpretation_events(local_measure_id,state,evidence_document_version_id,reason,limitations,actor,recorded_at) VALUES($1,'unreliable',$2,$3,$4,$5,$6)`, measureID, versionID, reason, limitations, actor, at.UTC()); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) RecordPrecedingStateWarning(ctx context.Context, measureID string, newerVersionID int64, reason string, at time.Time) error {
	if measureID == "" || newerVersionID < 1 || strings.TrimSpace(reason) == "" || at.IsZero() {
		return ErrInvalid
	}
	var precedingAcquired, newerAcquired time.Time
	var status string
	err := s.pool.QueryRow(ctx, `SELECT preceding.first_acquired_at,newer.first_acquired_at,COALESCE(latest.status,'uninterpreted')
 FROM domain_local_measures m
 JOIN retained_versions preceding ON preceding.id=m.document_version_id
 JOIN retained_versions newer ON newer.id=$2
 LEFT JOIN LATERAL (SELECT status FROM extraction_results WHERE document_version_id=newer.id ORDER BY created_at DESC,run_id DESC LIMIT 1) latest ON true
 WHERE m.id=$1`, measureID, newerVersionID).Scan(&precedingAcquired, &newerAcquired, &status)
	if err != nil {
		return err
	}
	if !newerAcquired.After(precedingAcquired) || status == "extracted" {
		return ErrInvalid
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO domain_preceding_state_warnings(local_measure_id,newer_document_version_id,reason,recorded_at) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, measureID, newerVersionID, reason, at.UTC())
	return err
}

func (s *Store) MeasureView(ctx context.Context, measureID string, evaluatedAt time.Time) (MeasureView, error) {
	measure, ok, err := s.LocalMeasure(ctx, measureID)
	if err != nil {
		return MeasureView{}, err
	}
	if !ok {
		return MeasureView{}, pgx.ErrNoRows
	}
	state, err := s.MeasureState(ctx, measureID, evaluatedAt)
	if err != nil {
		return MeasureView{}, err
	}
	view := MeasureView{Measure: measure, State: state}
	if err = s.loadProvenance(ctx, measure.SourceID, &view.Quality.Provenance); err != nil {
		return MeasureView{}, err
	}
	if err = s.loadInterpretation(ctx, measureID, &view.Quality.Interpretation); err != nil {
		return MeasureView{}, err
	}
	if err = s.loadUpdating(ctx, measure.SourceID, evaluatedAt, &view.Quality.Updating); err != nil {
		return MeasureView{}, err
	}
	view.Warnings, err = s.precedingWarnings(ctx, measureID)
	if err != nil {
		return MeasureView{}, err
	}
	if len(view.Warnings) > 0 {
		view.Quality.Interpretation.Limitations = append(view.Quality.Interpretation.Limitations, "newer potentially superseding document is not interpreted")
	}
	return view, nil
}

func (s *Store) loadProvenance(ctx context.Context, sourceID string, target *Dimension) error {
	var raw []byte
	var evidence QualityEvidence
	err := s.pool.QueryRow(ctx, `SELECT state,evidence_url,evidence_locator,limitations,assessed_at FROM domain_provenance_events WHERE source_id=$1 ORDER BY id DESC LIMIT 1`, sourceID).Scan(&target.State, &evidence.URL, &evidence.Locator, &raw, &evidence.ObservedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		target.State = "unresolved"
		target.Limitations = []string{"no provenance assessment recorded"}
		return nil
	}
	if err != nil {
		return err
	}
	if json.Unmarshal(raw, &target.Limitations) != nil {
		return ErrInvalid
	}
	target.Evidence = []QualityEvidence{evidence}
	return nil
}

func (s *Store) loadInterpretation(ctx context.Context, measureID string, target *Dimension) error {
	var raw []byte
	var evidence QualityEvidence
	var reason string
	err := s.pool.QueryRow(ctx, `SELECT state,evidence_document_version_id,reason,limitations,recorded_at FROM domain_interpretation_events WHERE local_measure_id=$1 ORDER BY id DESC LIMIT 1`, measureID).Scan(&target.State, &evidence.DocumentVersionID, &reason, &raw, &evidence.ObservedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		target.State = "not_processed"
		target.Limitations = []string{"no interpretation assessment recorded"}
		return nil
	}
	if err != nil {
		return err
	}
	if json.Unmarshal(raw, &target.Limitations) != nil {
		return ErrInvalid
	}
	evidence.Locator = reason
	target.Evidence = []QualityEvidence{evidence}
	return nil
}

func (s *Store) loadUpdating(ctx context.Context, sourceID string, at time.Time, target *UpdatingDimension) error {
	var collectionEnabled bool
	var delaySeconds int
	var lastErrorAt *time.Time
	var lastErrorCode string
	err := s.pool.QueryRow(ctx, `SELECT r.collection_enabled,a.last_started_at,a.last_complete_at,a.last_error_at,COALESCE(a.last_error_code,''),COALESCE(a.delay_seconds,1)
 FROM registry_sources r LEFT JOIN acquisition_source_status a ON a.source_id=r.id WHERE r.id=$1`, sourceID).Scan(&collectionEnabled, &target.LastAttemptAt, &target.LastCompleteAt, &lastErrorAt, &lastErrorCode, &delaySeconds)
	if err != nil {
		return err
	}
	if target.LastAttemptAt == nil && target.LastCompleteAt == nil {
		target.State = "not_verified"
		target.Limitations = []string{"no complete source check recorded"}
		return nil
	}
	if !collectionEnabled {
		target.State = "suspended"
		target.Limitations = []string{"collection is suspended; retained validity is unchanged"}
		return nil
	}
	if target.LastCompleteAt != nil && !at.Before(target.LastCompleteAt.Add(time.Duration(delaySeconds)*time.Second)) {
		target.State = "delayed"
		target.Limitations = []string{"complete-check delay threshold reached; measure validity is evaluated separately"}
		return nil
	}
	if lastErrorAt != nil {
		target.State = "failed"
		target.Limitations = []string{fmt.Sprintf("latest source check failed (%s); measure validity is unchanged", lastErrorCode)}
		return nil
	}
	target.State = "ok"
	return nil
}

func (s *Store) precedingWarnings(ctx context.Context, measureID string) ([]AttentionDocument, error) {
	rows, err := s.pool.Query(ctx, `SELECT w.newer_document_version_id,d.official_url,v.first_acquired_at,COALESCE(latest.status,'uninterpreted'),COALESCE(latest.reason_code,'not_processed')
 FROM domain_preceding_state_warnings w
 JOIN retained_versions v ON v.id=w.newer_document_version_id
 JOIN retained_documents d ON d.id=v.document_id
 LEFT JOIN LATERAL (SELECT status,reason_code FROM extraction_results WHERE document_version_id=v.id ORDER BY created_at DESC,run_id DESC LIMIT 1) latest ON true
 WHERE w.local_measure_id=$1 AND NOT EXISTS(
 SELECT 1 FROM interpretation_reuse reuse JOIN processing_runs current ON current.id=reuse.run_id
 JOIN extraction_results verified ON verified.run_id=current.id AND verified.status='extracted'
 JOIN processing_runs original ON original.id=reuse.original_run_id
 JOIN domain_local_measures measure ON measure.id=w.local_measure_id
 WHERE current.document_version_id=v.id AND original.document_version_id=measure.document_version_id) ORDER BY v.first_acquired_at,v.id`, measureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []AttentionDocument
	for rows.Next() {
		var value AttentionDocument
		if err = rows.Scan(&value.DocumentVersionID, &value.OfficialURL, &value.FirstAcquiredAt, &value.Status, &value.Reason); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
