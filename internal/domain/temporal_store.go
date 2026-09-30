package domain

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Store) BindExtraction(ctx context.Context, measureID string, extractionRunID int64, ordinal int) error {
	if measureID == "" || extractionRunID < 1 || ordinal < 1 {
		return ErrInvalid
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO domain_measure_bindings(local_measure_id,extraction_run_id,extraction_ordinal) VALUES($1,$2,$3)`, measureID, extractionRunID, ordinal)
	return err
}

func (s *Store) ApplyLinkedUpdate(ctx context.Context, linkingRunID int64, appliedAt time.Time) (MeasureUpdate, error) {
	if linkingRunID < 1 || appliedAt.IsZero() {
		return MeasureUpdate{}, ErrInvalid
	}
	var value MeasureUpdate
	err := s.pool.QueryRow(ctx, `SELECT l.run_id,current_binding.local_measure_id,target_binding.local_measure_id,l.relation
 FROM linking_results l
 JOIN domain_measure_bindings current_binding ON current_binding.extraction_run_id=l.current_extraction_run_id AND current_binding.extraction_ordinal=l.current_measure_ordinal
 JOIN domain_measure_bindings target_binding ON target_binding.extraction_run_id=l.candidate_extraction_run_id AND target_binding.extraction_ordinal=l.candidate_measure_ordinal
 WHERE l.run_id=$1 AND l.status='linked'`, linkingRunID).Scan(&value.LinkingRunID, &value.UpdateMeasureID, &value.TargetMeasureID, &value.Relation)
	if errors.Is(err, pgx.ErrNoRows) {
		return MeasureUpdate{}, ErrInvalid
	}
	if err != nil {
		return MeasureUpdate{}, err
	}
	value.AppliedAt = appliedAt.UTC()
	tag, err := s.pool.Exec(ctx, `INSERT INTO domain_measure_updates(linking_run_id,update_measure_id,target_measure_id,relation,applied_at) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, value.LinkingRunID, value.UpdateMeasureID, value.TargetMeasureID, value.Relation, value.AppliedAt)
	if err != nil {
		return MeasureUpdate{}, err
	}
	if tag.RowsAffected() == 0 {
		var existing MeasureUpdate
		if err = s.pool.QueryRow(ctx, `SELECT linking_run_id,update_measure_id,target_measure_id,relation,applied_at FROM domain_measure_updates WHERE linking_run_id=$1`, linkingRunID).Scan(&existing.LinkingRunID, &existing.UpdateMeasureID, &existing.TargetMeasureID, &existing.Relation, &existing.AppliedAt); err != nil {
			return MeasureUpdate{}, err
		}
		if existing.UpdateMeasureID != value.UpdateMeasureID || existing.TargetMeasureID != value.TargetMeasureID || existing.Relation != value.Relation {
			return MeasureUpdate{}, ErrConflict
		}
		return existing, nil
	}
	return value, nil
}

func (s *Store) PutTemporal(ctx context.Context, value TemporalValue) (int64, error) {
	if !validTemporal(value) {
		return 0, ErrInvalid
	}
	var versionID int64
	switch value.EntityKind {
	case "local_measure":
		if err := s.pool.QueryRow(ctx, "SELECT document_version_id FROM domain_local_measures WHERE id=$1", value.EntityID).Scan(&versionID); err != nil {
			return 0, err
		}
	case "operational_phase":
		if err := s.pool.QueryRow(ctx, "SELECT document_version_id FROM domain_operational_phases WHERE id=$1", value.EntityID).Scan(&versionID); err != nil {
			return 0, err
		}
	case "regional_record":
		if err := s.pool.QueryRow(ctx, "SELECT document_version_id FROM domain_regional_records WHERE id=$1", value.EntityID).Scan(&versionID); err != nil {
			return 0, err
		}
	}
	if versionID != value.EvidenceDocumentVersionID {
		return 0, ErrInvalid
	}
	var id int64
	err := s.pool.QueryRow(ctx, `INSERT INTO domain_temporal_values(entity_kind,entity_id,meaning,original_expression,precision,instant,date_value,end_instant,timezone,assumption,condition,conflict_group,evidence_document_version_id,created_at)
 VALUES($1,$2,$3,NULLIF($4,''),$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id`, value.EntityKind, value.EntityID, value.Meaning, value.Original, value.Precision, utcValue(value.Instant), dateValue(value.Date), utcValue(value.EndInstant), value.Timezone, value.Assumption, value.Condition, value.ConflictGroup, value.EvidenceDocumentVersionID, value.CreatedAt.UTC()).Scan(&id)
	return id, err
}

func utcValue(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}

func (s *Store) TemporalValues(ctx context.Context, entityKind, entityID string) ([]TemporalValue, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,entity_kind,entity_id,meaning,COALESCE(original_expression,''),precision,instant,date_value,end_instant,timezone,assumption,condition,conflict_group,evidence_document_version_id,created_at FROM domain_temporal_values WHERE entity_kind=$1 AND entity_id=$2 ORDER BY id`, entityKind, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []TemporalValue
	for rows.Next() {
		var value TemporalValue
		if err = rows.Scan(&value.ID, &value.EntityKind, &value.EntityID, &value.Meaning, &value.Original, &value.Precision, &value.Instant, &value.Date, &value.EndInstant, &value.Timezone, &value.Assumption, &value.Condition, &value.ConflictGroup, &value.EvidenceDocumentVersionID, &value.CreatedAt); err != nil {
			return nil, err
		}
		if value.Instant != nil {
			instant := value.Instant.UTC()
			value.Instant = &instant
		}
		if value.EndInstant != nil {
			instant := value.EndInstant.UTC()
			value.EndInstant = &instant
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func (s *Store) MeasureState(ctx context.Context, measureID string, evaluatedAt time.Time) (MeasureState, error) {
	if measureID == "" || evaluatedAt.IsZero() {
		return MeasureState{}, ErrInvalid
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM domain_local_measures WHERE id=$1)", measureID).Scan(&exists); err != nil {
		return MeasureState{}, err
	}
	if !exists {
		return MeasureState{}, pgx.ErrNoRows
	}
	state := MeasureState{MeasureID: measureID, Status: "undetermined"}
	rows, err := s.pool.Query(ctx, `SELECT update_measure_id,relation FROM domain_measure_updates WHERE target_measure_id=$1 AND applied_at<=$2 ORDER BY applied_at,linking_run_id`, measureID, evaluatedAt.UTC())
	if err != nil {
		return MeasureState{}, err
	}
	for rows.Next() {
		var updatedBy, relation string
		if err = rows.Scan(&updatedBy, &relation); err != nil {
			rows.Close()
			return MeasureState{}, err
		}
		state.UpdatedBy = append(state.UpdatedBy, updatedBy)
		state.UpdateRelations = append(state.UpdateRelations, relation)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return MeasureState{}, err
	}
	rows.Close()
	if len(state.UpdatedBy) > 0 {
		state.Status = "superseded"
		return state, nil
	}
	var precision string
	var start, end *time.Time
	err = s.pool.QueryRow(ctx, `SELECT precision,instant,end_instant FROM domain_temporal_values WHERE entity_kind='local_measure' AND entity_id=$1 AND meaning='validity' ORDER BY id DESC LIMIT 1`, measureID).Scan(&precision, &start, &end)
	if errors.Is(err, pgx.ErrNoRows) {
		return state, nil
	}
	if err != nil {
		return MeasureState{}, err
	}
	switch precision {
	case "interval":
		if evaluatedAt.Before(*start) {
			state.Status = "future"
		} else if !evaluatedAt.Before(*end) {
			state.Status = "expired"
		} else {
			state.Status = "current"
		}
	case "instant":
		if evaluatedAt.Before(*start) {
			state.Status = "future"
		}
	}
	return state, nil
}
