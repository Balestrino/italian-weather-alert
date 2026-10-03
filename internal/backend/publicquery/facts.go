package publicquery

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Store) temporal(ctx context.Context, entityKind, entityID, meaning string, knownAt time.Time) (Temporal, error) {
	var value Temporal
	var date *time.Time
	var condition *string
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(original_expression,''),precision,instant,date_value,end_instant,timezone,assumption,condition
 FROM domain_temporal_values WHERE entity_kind=$1 AND entity_id=$2 AND meaning=$3 AND created_at<=registry_interpretation_cutoff(COALESCE(
 (SELECT source_id FROM domain_local_measures WHERE id=$2 AND $1='local_measure'),
 (SELECT source_id FROM domain_operational_phases WHERE id=$2 AND $1='operational_phase'),
 (SELECT source_id FROM domain_regional_records WHERE id=$2 AND $1='regional_record')),$4) ORDER BY created_at DESC,id DESC LIMIT 1`, entityKind, entityID, meaning, knownAt).Scan(&value.Original, &value.Precision, &value.Instant, &date, &value.EndInstant, &value.Timezone, &value.Assumption, &condition)
	if errors.Is(err, pgx.ErrNoRows) {
		return unknownTemporal(), nil
	}
	if err != nil {
		return Temporal{}, err
	}
	value.Instant, value.EndInstant = utcPointer(value.Instant), utcPointer(value.EndInstant)
	if date != nil {
		formatted := date.Format(time.DateOnly)
		value.Date = &formatted
	}
	value.Condition = condition
	return value, nil
}

func (s *Store) versionEvidence(ctx context.Context, versionID int64, locator string) (Evidence, error) {
	var documentID int64
	var value Evidence
	err := s.pool.QueryRow(ctx, `SELECT d.id,v.id,d.official_url FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id WHERE v.id=$1`, versionID).Scan(&documentID, &versionID, &value.SourceURL)
	if err != nil {
		return Evidence{}, err
	}
	value.DocumentID, value.VersionID = stringID(documentID), stringID(versionID)
	value.Locator = &locator
	return value, nil
}

func (s *Store) measureEvidence(ctx context.Context, measureID string) ([]Evidence, error) {
	rows, err := s.pool.Query(ctx, `SELECT d.id,v.id,e.resource_url,e.page_number,e.field_name,e.quote
 FROM domain_measure_bindings b JOIN extraction_evidence e ON e.run_id=b.extraction_run_id AND e.measure_ordinal=b.extraction_ordinal
 JOIN domain_local_measures m ON m.id=b.local_measure_id JOIN retained_versions v ON v.id=m.document_version_id JOIN retained_documents d ON d.id=v.document_id
 WHERE b.local_measure_id=$1 ORDER BY e.ordinal`, measureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Evidence{}
	for rows.Next() {
		var documentID, versionID int64
		var value Evidence
		var locator, passage string
		if err = rows.Scan(&documentID, &versionID, &value.SourceURL, &value.Page, &locator, &passage); err != nil {
			return nil, err
		}
		value.DocumentID, value.VersionID, value.Locator, value.Passage = stringID(documentID), stringID(versionID), &locator, &passage
		result = append(result, value)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return result, nil
	}
	var versionID int64
	if err = s.pool.QueryRow(ctx, `SELECT document_version_id FROM domain_local_measures WHERE id=$1`, measureID).Scan(&versionID); err != nil {
		return nil, err
	}
	value, err := s.versionEvidence(ctx, versionID, "retained source document")
	if err != nil {
		return nil, err
	}
	return []Evidence{value}, nil
}

func (s *Store) measure(ctx context.Context, id string, qt QueryTime) (Measure, error) {
	var value Measure
	var versionID int64
	err := s.pool.QueryRow(ctx, `SELECT m.id,m.document_version_id,m.municipality_istat,a.name,m.kind,m.place
 FROM domain_local_measures m JOIN retained_versions v ON v.id=m.document_version_id
 JOIN `+s.sourcesSQL()+` s ON s.id=m.source_id LEFT JOIN registry_authorities a ON a.id=m.issuing_authority_id
	 WHERE m.id=$1 AND v.first_acquired_at<=$2 AND m.recorded_at<=registry_interpretation_cutoff(s.id,$2) AND `+s.visibilitySQL()+``, id, qt.KnownAt).Scan(&value.ID, &versionID, &value.MunicipalityISTAT, &value.Issuer, &value.Action, &value.Place)
	if err != nil {
		return Measure{}, err
	}
	value.RevisionID = stringID(versionID)
	// A missing place is an unresolved extraction field. The municipality is
	// the source's jurisdiction, not evidence that a measure covers it all.
	value.TerritorialScope = "undetermined"
	if value.Place != nil {
		value.TerritorialScope = "place"
	}
	value.Validity, err = s.temporal(ctx, "local_measure", value.ID, "validity", qt.KnownAt)
	if err != nil {
		return Measure{}, err
	}
	value.Status = statusForTemporal(value.Validity, qt.EvaluationTime)
	value.Evidence, err = s.measureEvidence(ctx, value.ID)
	if err != nil {
		return Measure{}, err
	}
	interpretation, err := s.measureInterpretation(ctx, value.ID, qt.KnownAt)
	if err != nil {
		return Measure{}, err
	}
	var sourceID string
	if err = s.pool.QueryRow(ctx, `SELECT source_id FROM domain_local_measures WHERE id=$1`, value.ID).Scan(&sourceID); err != nil {
		return Measure{}, err
	}
	value.Quality, err = s.quality(ctx, sourceID, qt, interpretation)
	if err != nil {
		return Measure{}, err
	}
	value.RelatedMeasureIDs, value.RelationshipStatus, err = s.measureRelations(ctx, value.ID, qt)
	if err != nil {
		return Measure{}, err
	}
	if value.Status != "superseded" {
		var superseded bool
		if err = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM domain_measure_updates WHERE target_measure_id=$1 AND applied_at<=LEAST($2::timestamptz,registry_interpretation_cutoff((SELECT source_id FROM domain_local_measures WHERE id=$1),$3)))`, value.ID, qt.EvaluationTime, qt.KnownAt).Scan(&superseded); err != nil {
			return Measure{}, err
		}
		if superseded {
			value.Status = "superseded"
		}
	}
	value.NewerUninterpretedDocumentIDs, err = s.newerWarnings(ctx, value.ID, qt.KnownAt)
	if err == nil && len(value.NewerUninterpretedDocumentIDs) > 0 {
		value.Quality.Interpretation.Limitations = append(value.Quality.Interpretation.Limitations, "newer potentially superseding document is not interpreted")
	}
	if err != nil {
		return Measure{}, err
	}
	value.Verifications, err = s.forMunicipality(value.MunicipalityISTAT).verifications(ctx, "local_measure", value.ID, 0, qt)
	return value, err
}

func (s *Store) measureInterpretation(ctx context.Context, id string, knownAt time.Time) (Dimension, error) {
	var result Dimension
	var raw []byte
	var versionID int64
	var reason string
	var assessedAt time.Time
	err := s.pool.QueryRow(ctx, `SELECT state,evidence_document_version_id,reason,limitations,recorded_at FROM domain_interpretation_events
 WHERE local_measure_id=$1 AND recorded_at<=registry_interpretation_cutoff((SELECT source_id FROM domain_local_measures WHERE id=$1),$2) ORDER BY recorded_at DESC,id DESC LIMIT 1`, id, knownAt).Scan(&result.State, &versionID, &reason, &raw, &assessedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Dimension{State: "not_processed", Limitations: []string{"no interpretation assessment was available at the knowledge boundary"}, Evidence: []Evidence{}}, nil
	}
	if err != nil {
		return Dimension{}, err
	}
	if json.Unmarshal(raw, &result.Limitations) != nil {
		return Dimension{}, ErrInvalidParameters
	}
	evidence, err := s.versionEvidence(ctx, versionID, reason)
	if err != nil {
		return Dimension{}, err
	}
	result.Evidence = []Evidence{evidence}
	return s.preserveDefectWarning(ctx, versionID, assessedAt, result)
}

func (s *Store) measureRelations(ctx context.Context, id string, qt QueryTime) ([]string, string, error) {
	rows, err := s.pool.Query(ctx, `SELECT CASE WHEN target_measure_id=$1 THEN update_measure_id ELSE target_measure_id END
	 FROM domain_measure_updates WHERE (target_measure_id=$1 OR update_measure_id=$1) AND applied_at<=LEAST($2::timestamptz,registry_interpretation_cutoff((SELECT source_id FROM domain_local_measures WHERE id=$1),$3))
 ORDER BY applied_at,linking_run_id`, id, qt.EvaluationTime, qt.KnownAt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var related string
		if err = rows.Scan(&related); err != nil {
			return nil, "", err
		}
		result = append(result, related)
	}
	if err = rows.Err(); err != nil {
		return nil, "", err
	}
	if len(result) > 0 {
		return result, "supported", nil
	}
	return result, "not_applicable", nil
}

func (s *Store) newerWarnings(ctx context.Context, id string, knownAt time.Time) ([]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT d.id FROM domain_preceding_state_warnings w JOIN retained_versions v ON v.id=w.newer_document_version_id JOIN retained_documents d ON d.id=v.document_id
 WHERE w.local_measure_id=$1 AND v.first_acquired_at<=$2 AND NOT EXISTS(
 SELECT 1 FROM interpretation_reuse reuse JOIN processing_runs current ON current.id=reuse.run_id
 JOIN extraction_results verified ON verified.run_id=current.id AND verified.status='extracted'
 JOIN processing_runs original ON original.id=reuse.original_run_id
 JOIN domain_local_measures measure ON measure.id=w.local_measure_id
 WHERE current.document_version_id=v.id AND original.document_version_id=measure.document_version_id AND reuse.created_at<=$2 AND verified.created_at<=$2) ORDER BY v.first_acquired_at,v.id`, id, knownAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var documentID int64
		if err = rows.Scan(&documentID); err != nil {
			return nil, err
		}
		result = append(result, stringID(documentID))
	}
	return result, rows.Err()
}

func (s *Store) measures(ctx context.Context, municipality string, qt QueryTime) ([]Measure, error) {
	rows, err := s.pool.Query(ctx, `SELECT m.id FROM domain_local_measures m JOIN retained_versions v ON v.id=m.document_version_id JOIN `+s.sourcesSQL()+` s ON s.id=m.source_id
	 WHERE ($1='' OR m.municipality_istat=$1) AND v.first_acquired_at<=$2 AND m.recorded_at<=registry_interpretation_cutoff(s.id,$2) AND `+s.visibilitySQL()+`
	   AND domain_verification_record_current('local_measure',m.id,registry_interpretation_cutoff(s.id,$2))
	   AND COALESCE((SELECT state FROM domain_interpretation_events i WHERE i.local_measure_id=m.id AND i.recorded_at<=registry_interpretation_cutoff(s.id,$2) ORDER BY i.recorded_at DESC,i.id DESC LIMIT 1),'not_processed') IN ('supported','partial','unreliable')
	 ORDER BY m.id`, municipality, qt.KnownAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	result := make([]Measure, 0, len(ids))
	for _, id := range ids {
		value, loadErr := s.measure(ctx, id, qt)
		if loadErr != nil {
			return nil, loadErr
		}
		result = append(result, value)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if sortStatus(result[i].Status) != sortStatus(result[j].Status) {
			return sortStatus(result[i].Status) < sortStatus(result[j].Status)
		}
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (s *Store) phases(ctx context.Context, municipality string, qt QueryTime) ([]OperationalPhase, error) {
	rows, err := s.pool.Query(ctx, `SELECT p.id,p.document_version_id,p.municipality_istat,a.name,p.phase,p.source_id,p.recorded_at
 FROM domain_operational_phases p JOIN retained_versions v ON v.id=p.document_version_id JOIN `+s.sourcesSQL()+` s ON s.id=p.source_id
 JOIN registry_authorities a ON a.id=p.authority_id
	 WHERE ($1='' OR p.municipality_istat=$1) AND v.first_acquired_at<=$2 AND p.recorded_at<=registry_interpretation_cutoff(s.id,$2) AND domain_verification_record_current('operational_phase',p.id,registry_interpretation_cutoff(s.id,$2)) AND `+s.visibilitySQL()+` ORDER BY p.id`, municipality, qt.KnownAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type row struct {
		value      OperationalPhase
		versionID  int64
		sourceID   string
		recordedAt time.Time
	}
	var pending []row
	for rows.Next() {
		var item row
		if err = rows.Scan(&item.value.ID, &item.versionID, &item.value.MunicipalityISTAT, &item.value.Authority, &item.value.Phase, &item.sourceID, &item.recordedAt); err != nil {
			return nil, err
		}
		pending = append(pending, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	result := make([]OperationalPhase, 0, len(pending))
	for _, item := range pending {
		item.value.Validity, err = s.temporal(ctx, "operational_phase", item.value.ID, "validity", qt.KnownAt)
		if err != nil {
			return nil, err
		}
		evidence, evidenceErr := s.versionEvidence(ctx, item.versionID, "retained operational-phase evidence")
		if evidenceErr != nil {
			return nil, evidenceErr
		}
		item.value.Evidence = []Evidence{evidence}
		assessment, assessmentErr := s.preserveDefectWarning(ctx, item.versionID, item.recordedAt, Dimension{State: "supported", Limitations: []string{}, Evidence: []Evidence{evidence}})
		if assessmentErr != nil {
			return nil, assessmentErr
		}
		item.value.Quality, err = s.quality(ctx, item.sourceID, qt, assessment)
		if err != nil {
			return nil, err
		}
		item.value.Verifications, err = s.forMunicipality(item.value.MunicipalityISTAT).verifications(ctx, "operational_phase", item.value.ID, 0, qt)
		if err != nil {
			return nil, err
		}
		result = append(result, item.value)
	}
	return result, nil
}

func (s *Store) regional(ctx context.Context, municipality, zone, product, risk, sourceID string, qt QueryTime, history ...bool) ([]RegionalWarning, error) {
	zoneSet := map[string]bool{}
	mappingVersion := "unavailable"
	selectedMapping, mappingErr := s.selectedDatasetAt(ctx, "zone_mapping", qt.KnownAt)
	if mappingErr == nil {
		mappingVersion = selectedMapping
	} else if !errors.Is(mappingErr, ErrMappingUnavailable) {
		return nil, mappingErr
	}
	if municipality != "" {
		selected, err := s.municipality(ctx, municipality, qt.KnownAt)
		if err != nil {
			return nil, err
		}
		for _, candidate := range selected.Zones {
			zoneSet[candidate] = true
		}
		if selected.MappingVersion != nil {
			mappingVersion = *selected.MappingVersion
		}
	}
	rows, err := s.pool.Query(ctx, `SELECT r.id,r.document_version_id,r.source_id,r.product,f.ordinal,f.risk,f.official_risk_label,f.zone,f.level,r.recorded_at,COALESCE(r.evidence_locator,'retained regional product evidence'),COALESCE(p.limitations,'[]'::jsonb),EXISTS(SELECT 1 FROM domain_regional_projections failed JOIN retained_versions fv ON fv.id=failed.document_version_id WHERE failed.status='unsupported' AND fv.document_id=v.document_id AND (fv.first_acquired_at,fv.id)>(v.first_acquired_at,v.id) AND failed.projected_at<=registry_interpretation_cutoff(s.id,$1) AND fv.first_acquired_at<=$1)
 FROM domain_regional_records r JOIN domain_regional_facts f ON f.regional_record_id=r.id
 LEFT JOIN domain_regional_projections p ON p.document_version_id=r.document_version_id AND p.logic_version=r.projection_logic
 JOIN retained_versions v ON v.id=r.document_version_id JOIN `+s.sourcesSQL()+` s ON s.id=r.source_id
	 WHERE v.first_acquired_at<=$1 AND r.recorded_at<=registry_interpretation_cutoff(s.id,$1) AND (`+s.visibilitySQL()+`) AND (`+s.regionalDevelopmentScopeSQL(selectedMapping)+`) AND ($2='' OR r.product=$2) AND ($3='' OR f.risk=$3) AND ($4='' OR r.source_id=$4)
 AND ($5 OR domain_verification_record_current('regional_record',r.id,registry_interpretation_cutoff(s.id,$1)))
 AND (r.projection_logic IS NULL OR NOT EXISTS(
 SELECT 1 FROM domain_regional_projections replacement
 WHERE replacement.document_version_id=r.document_version_id AND replacement.status<>'unsupported'
 AND replacement.projected_at<=registry_interpretation_cutoff(s.id,$1)
 AND (replacement.projected_at,replacement.logic_version)>(p.projected_at,p.logic_version)))
 AND (r.projection_logic IS NULL OR $5 OR NOT EXISTS(
 SELECT 1 FROM domain_regional_projections newer JOIN retained_versions nv ON nv.id=newer.document_version_id
 WHERE nv.document_id=v.document_id AND newer.status<>'unsupported'
 AND newer.projected_at<=registry_interpretation_cutoff(s.id,$1) AND nv.first_acquired_at<=$1
 AND (nv.first_acquired_at,nv.id)>(v.first_acquired_at,v.id)))
 ORDER BY v.first_acquired_at,r.id,f.ordinal`, qt.KnownAt, product, risk, sourceID, len(history) > 0 && history[0])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type regionalRow struct {
		recordID, sourceID string
		locator            string
		newerUnsupported   bool
		limitations        []string
		versionID          int64
		value              RegionalWarning
		recordedAt         time.Time
	}
	var pending []regionalRow
	for rows.Next() {
		var item regionalRow
		var ordinal int
		var level *string
		if err = rows.Scan(&item.recordID, &item.versionID, &item.sourceID, &item.value.Product, &ordinal, &item.value.Risk, &item.value.OfficialRiskLabel, &item.value.Zone, &level, &item.recordedAt, &item.locator, &item.limitations, &item.newerUnsupported); err != nil {
			return nil, err
		}
		if zone != "" && item.value.Zone != zone {
			continue
		}
		if municipality != "" && !zoneSet[item.value.Zone] {
			continue
		}
		item.value.ID = item.recordID + ":" + stringID(int64(ordinal))
		item.value.MappingVersion = mappingVersion
		if level == nil {
			item.value.Level = "not_applicable"
		} else {
			item.value.Level = *level
		}
		pending = append(pending, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	result := make([]RegionalWarning, 0, len(pending))
	for _, item := range pending {
		item.value.Validity, err = s.temporal(ctx, "regional_record", item.recordID, "validity", qt.KnownAt)
		if err != nil {
			return nil, err
		}
		item.value.Status = statusForTemporal(item.value.Validity, qt.EvaluationTime)
		if item.value.Validity.Precision == "date" && item.value.Validity.Date != nil {
			loc, e := time.LoadLocation("Europe/Rome")
			if e != nil {
				return nil, e
			}
			day := qt.EvaluationTime.In(loc).Format(time.DateOnly)
			item.value.Status = "current"
			if day < *item.value.Validity.Date {
				item.value.Status = "future"
			} else if day > *item.value.Validity.Date {
				item.value.Status = "expired"
			}
		}
		evidence, evidenceErr := s.versionEvidence(ctx, item.versionID, item.locator)
		if evidenceErr != nil {
			return nil, evidenceErr
		}
		item.value.Evidence = []Evidence{evidence}
		if item.newerUnsupported {
			item.limitations = append(item.limitations, "A newer retained bulletin could not be interpreted; these are the preceding documented facts, not a fully updated situation.")
		}
		state := "supported"
		if len(item.limitations) > 0 {
			state = "partial"
		}
		assessment, assessmentErr := s.preserveDefectWarning(ctx, item.versionID, item.recordedAt, Dimension{State: state, Limitations: item.limitations, Evidence: []Evidence{evidence}})
		if assessmentErr != nil {
			return nil, assessmentErr
		}
		item.value.Quality, err = s.quality(ctx, item.sourceID, qt, assessment)
		if err != nil {
			return nil, err
		}
		if item.value.MappingVersion == "unavailable" {
			item.value.Quality.Interpretation.Limitations = append(item.value.Quality.Interpretation.Limitations, "no municipality-zone mapping was available at the knowledge boundary")
		}
		item.value.Verifications, err = s.forMunicipality(municipality).verifications(ctx, "regional_record", item.recordID, 0, qt)
		if err != nil {
			return nil, err
		}
		result = append(result, item.value)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if sortStatus(result[i].Status) != sortStatus(result[j].Status) {
			return sortStatus(result[i].Status) < sortStatus(result[j].Status)
		}
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func intersects(value Temporal, from, to *time.Time) bool {
	if from == nil && to == nil {
		return true
	}
	if value.Precision == "unknown" || value.Precision == "conditional" || (value.Instant == nil && value.Date == nil) {
		return true
	}
	var start, end *time.Time
	if value.Instant != nil {
		start = value.Instant
		end = value.EndInstant
	} else if value.Date != nil {
		if parsed, err := time.Parse(time.DateOnly, *value.Date); err == nil {
			parsed = parsed.UTC()
			start = &parsed
			next := parsed.AddDate(0, 0, 1)
			end = &next
		}
	}
	if to != nil && start != nil && !start.Before(*to) {
		return false
	}
	if from != nil && end != nil && !end.After(*from) {
		return false
	}
	return true
}

func filterMeasure(values []Measure, query SearchQuery) []Measure {
	result := []Measure{}
	for _, value := range values {
		if query.Status != "" && value.Status != query.Status {
			continue
		}
		if !intersects(value.Validity, query.From, query.To) {
			continue
		}
		result = append(result, value)
	}
	return result
}

func filterRegional(values []RegionalWarning, query SearchQuery) []RegionalWarning {
	result := []RegionalWarning{}
	for _, value := range values {
		if query.Status != "" && value.Status != query.Status {
			continue
		}
		if !intersects(value.Validity, query.From, query.To) {
			continue
		}
		result = append(result, value)
	}
	return result
}

var validStatuses = []string{"current", "future", "expired", "cancelled", "superseded", "undetermined"}

func validStatus(value string) bool { return value == "" || slices.Contains(validStatuses, value) }

func validProduct(value string) bool {
	return value == "" || slices.Contains([]string{"vigilance", "criticality", "monitoring"}, value)
}

func validRisk(value string) bool {
	return value == "" || slices.Contains([]string{"minor_network_hydro", "main_network_hydraulic", "thunderstorms", "wind", "coastal_waves", "snow", "ice"}, value)
}
