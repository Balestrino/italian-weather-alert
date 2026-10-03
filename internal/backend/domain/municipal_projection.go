package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

const MunicipalProjectionLogic = "municipal-extraction-v1"

// Carry the official-provenance evidence already reviewed in the active source
// configuration into the domain's separate quality history. Existing explicit
// assessments (including unresolved ones) are never replaced automatically.
func (s *Store) RecordConfiguredProvenance(ctx context.Context, sourceID string, at time.Time) error {
	if sourceID == "" || at.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var revision int
	var body []byte
	if err = tx.QueryRow(ctx, `SELECT c.revision,c.body FROM registry_sources s JOIN registry_configurations c ON c.source_id=s.id AND c.revision=s.active_revision WHERE s.id=$1 FOR SHARE OF s`, sourceID).Scan(&revision, &body); err != nil {
		return err
	}
	var configuration registry.Configuration
	if json.Unmarshal(body, &configuration) != nil {
		return ErrInvalid
	}
	e := configuration.Provenance
	if e == nil || e.URL == "" || e.Locator == "" || e.ObservedAt.IsZero() || e.ObservedAt.After(at) {
		return nil
	}
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,730071))`, sourceID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO domain_provenance_events(source_id,configuration,state,evidence_url,evidence_locator,limitations,assessed_at)
 SELECT $1,$2,'verified',$3,$4,'[]'::jsonb,$5 WHERE NOT EXISTS(SELECT 1 FROM domain_provenance_events WHERE source_id=$1 AND configuration=$2)`, sourceID, revision, e.URL, e.Locator, at.UTC())
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ProjectMunicipalExtraction publishes already validated primary extraction
// evidence atomically. It makes no model calls and grants no source acceptance.
// Evaluation runs and platform-only candidates cannot enter through this path.
func (s *Store) ProjectMunicipalExtraction(ctx context.Context, runID int64, at time.Time) (int, error) {
	if runID < 1 || at.IsZero() {
		return 0, ErrInvalid
	}
	r, found, err := extraction.NewStore(s.pool).Get(ctx, runID)
	if err != nil || !found {
		if err == nil {
			err = ErrInvalid
		}
		return 0, err
	}
	if r.Status != "extracted" || !r.ContentComplete || r.CreatedAt.After(at) {
		return 0, nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var sourceID, municipality, authority, publisher, platform, product, workload, url string
	var suspended *time.Time
	var allowed, complete, relevant bool
	err = tx.QueryRow(ctx, `SELECT s.id,s.territory,s.authority_id,c.publisher_id,c.platform,s.product_id,p.workload,d.official_url,
 s.interpretation_suspended_at,territorial_source_allowed(s.id),v.complete,COALESCE(cl.relevant,false)
 FROM extraction_results e JOIN processing_runs p ON p.id=e.run_id
 JOIN classification_results cl ON cl.run_id=e.classification_run_id AND cl.document_version_id=e.document_version_id
 JOIN retained_versions v ON v.id=e.document_version_id JOIN retained_documents d ON d.id=v.document_id
 JOIN registry_sources s ON s.id=d.source_id JOIN registry_channels c ON c.id=s.channel_id
 WHERE e.run_id=$1 AND p.source_id=s.id AND p.document_version_id=e.document_version_id FOR SHARE OF s`, runID).Scan(&sourceID, &municipality, &authority, &publisher, &platform, &product, &workload, &url, &suspended, &allowed, &complete, &relevant)
	if err != nil {
		return 0, err
	}
	if product != "municipal" || platform == "cittadino-informato" || workload == "evaluation" || suspended != nil || !allowed || !complete || !relevant {
		return 0, nil
	}
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,730070))`, fmt.Sprintf("%s:%d", MunicipalProjectionLogic, r.DocumentVersionID)); err != nil {
		return 0, err
	}
	count := 0
	for _, m := range r.Measures {
		if !slices.Contains([]string{"closure", "reopening", "restriction", "prohibition", "suspension", "activation", "deactivation", "operational_update", "observation"}, m.Kind) || strings.TrimSpace(m.Subject) == "" || !municipalFieldEvidence(m, "kind") || !municipalFieldEvidence(m, "subject") {
			continue
		}
		if (m.Place != nil && !municipalFieldEvidence(m, "place")) || (m.ValidFrom != nil && !municipalFieldEvidence(m, "valid_from")) || (m.ValidUntil != nil && !municipalFieldEvidence(m, "valid_until")) {
			continue
		}
		owned := true
		for _, e := range m.Evidence {
			var retained bool
			if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM retained_resources WHERE version_id=$1 AND url=$2 AND missing='')`, r.DocumentVersionID, e.ResourceURL).Scan(&retained); err != nil {
				return 0, err
			}
			if !retained {
				owned = false
				break
			}
		}
		if !owned {
			continue
		}
		// Do not convert an operational activation into a named phase by inference.
		var bound bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM domain_measure_bindings WHERE extraction_run_id=$1 AND extraction_ordinal=$2)`, runID, m.Ordinal).Scan(&bound); err != nil {
			return 0, err
		}
		if bound {
			count++
			continue
		}
		key := fmt.Sprintf("%s:%d:%d", MunicipalProjectionLogic, runID, m.Ordinal)
		tag, e := tx.Exec(ctx, `INSERT INTO domain_local_measures(id,document_version_id,source_id,municipality_istat,issuing_authority_id,publisher_id,platform,kind,subject,place,recorded_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING`, key, r.DocumentVersionID, sourceID, municipality, authority, publisher, platform, m.Kind, m.Subject, m.Place, at.UTC())
		if e != nil {
			return 0, e
		}
		count++
		if tag.RowsAffected() == 0 {
			continue
		}
		if _, err = tx.Exec(ctx, `INSERT INTO domain_measure_bindings(local_measure_id,extraction_run_id,extraction_ordinal) VALUES($1,$2,$3)`, key, runID, m.Ordinal); err != nil {
			return 0, err
		}
		v := municipalValidity(m, r.DocumentVersionID, key, at)
		if _, err = tx.Exec(ctx, `INSERT INTO domain_temporal_values(entity_kind,entity_id,meaning,original_expression,precision,instant,date_value,end_instant,timezone,assumption,condition,evidence_document_version_id,created_at)
 VALUES('local_measure',$1,'validity',NULLIF($2,''),$3,$4,$5,$6,$7,$8,$9,$10,$11)`, key, v.Original, v.Precision, v.Instant, dateValue(v.Date), v.EndInstant, v.Timezone, v.Assumption, v.Condition, r.DocumentVersionID, at.UTC()); err != nil {
			return 0, err
		}
		limits := append([]string{}, m.IndeterminateFields...)
		state := "supported"
		if v.Precision == "unknown" {
			limits = append(limits, "validity_not_established")
		}
		if m.Place == nil {
			limits = append(limits, "place_not_established")
		}
		if len(limits) > 0 {
			state = "partial"
		}
		raw, _ := json.Marshal(limits)
		if _, err = tx.Exec(ctx, `INSERT INTO domain_interpretation_events(local_measure_id,state,evidence_document_version_id,reason,limitations,actor,recorded_at)
 VALUES($1,$2,$3,'validated_primary_extraction',$4,$5,$6)`, key, state, r.DocumentVersionID, raw, MunicipalProjectionLogic, at.UTC()); err != nil {
			return 0, err
		}
		// Reprocessing/revisions of the same explicit action/scope replace only
		// that projection at the new knowledge boundary; unrelated acts stay separate.
		identity := map[string]string{"publication": url, "kind": m.Kind, "subject": cleanLiteral(m.Subject)}
		if m.ValidFrom != nil {
			identity["valid_from"] = cleanLiteral(*m.ValidFrom)
		}
		if m.ValidUntil != nil {
			identity["valid_until"] = cleanLiteral(*m.ValidUntil)
		}
		if m.Place != nil {
			identity["place"] = cleanLiteral(*m.Place)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO domain_verification_projections(kind,source_id,logical_key,document_version_id,domain_record_id,recorded_at,acquisition_at)
 SELECT 'local_measure',$1,$2,$3,$4,$5,COALESCE((SELECT max(acquired_at) FROM retained_acquisitions WHERE version_id=$3 AND acquired_at<=$5),(SELECT first_acquired_at FROM retained_versions WHERE id=$3))`, sourceID, verificationHash(identity), r.DocumentVersionID, key, at.UTC()); err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return count, s.ProjectMunicipalLinks(ctx, runID, at)
}

func municipalFieldEvidence(m extraction.Measure, field string) bool {
	return slices.ContainsFunc(m.Evidence, func(e extraction.Evidence) bool {
		return e.Field == field && e.ResourceURL != "" && strings.TrimSpace(e.Quote) != ""
	})
}

func municipalValidity(m extraction.Measure, version int64, key string, at time.Time) TemporalValue {
	v := TemporalValue{EntityKind: "local_measure", EntityID: key, Meaning: "validity", Precision: "unknown", EvidenceDocumentVersionID: version, CreatedAt: at.UTC()}
	var parts []string
	if m.ValidFrom != nil {
		parts = append(parts, *m.ValidFrom)
	}
	if m.ValidUntil != nil {
		parts = append(parts, *m.ValidUntil)
	}
	v.Original = strings.Join(parts, "; ")
	// Conflicting/unresolved candidates remain literal evidence, never a guessed interval.
	for _, candidate := range m.TemporalCandidates {
		if candidate.ConflictIdentity != nil {
			return v
		}
	}
	if m.ValidFrom != nil && m.ValidUntil != nil {
		start, e1 := time.Parse(time.RFC3339, *m.ValidFrom)
		end, e2 := time.Parse(time.RFC3339, *m.ValidUntil)
		if e1 == nil && e2 == nil && !end.Before(start) {
			v.Precision = "interval"
			v.Instant = &start
			v.EndInstant = &end
			return v
		}
	}
	if m.ValidUntil != nil {
		end := verificationTemporal("local_measure", key, version, VerifiedField{Passage: *m.ValidUntil}, at)
		if end.Precision == "conditional" {
			v.Precision = "conditional"
			v.Condition = end.Condition
			if m.ValidFrom != nil {
				if start, e := time.Parse(time.RFC3339, *m.ValidFrom); e == nil {
					v.Instant = &start
				}
			}
			return v
		}
	}
	return v
}

// Linking results are applied only after both primary measures have bindings.
func (s *Store) ProjectMunicipalLinks(ctx context.Context, extractionRunID int64, at time.Time) error {
	rows, err := s.pool.Query(ctx, `SELECT l.run_id FROM linking_results l
 JOIN domain_measure_bindings a ON a.extraction_run_id=l.current_extraction_run_id AND a.extraction_ordinal=l.current_measure_ordinal
 JOIN domain_measure_bindings b ON b.extraction_run_id=l.candidate_extraction_run_id AND b.extraction_ordinal=l.candidate_measure_ordinal
 JOIN domain_local_measures current ON current.id=a.local_measure_id
 JOIN domain_local_measures target ON target.id=b.local_measure_id
 JOIN registry_sources s ON s.id=current.source_id JOIN registry_sources t ON t.id=target.source_id
 JOIN processing_runs p ON p.id=l.run_id
 WHERE l.status='linked' AND p.workload<>'evaluation' AND l.created_at<=$2
 AND current.municipality_istat=target.municipality_istat AND s.interpretation_suspended_at IS NULL AND t.interpretation_suspended_at IS NULL
 AND territorial_source_allowed(s.id) AND territorial_source_allowed(t.id)
 AND EXISTS(SELECT 1 FROM linking_evidence WHERE run_id=l.run_id AND side='current')
 AND EXISTS(SELECT 1 FROM linking_evidence WHERE run_id=l.run_id AND side='candidate')
 AND (l.current_extraction_run_id=$1 OR l.candidate_extraction_run_id=$1) ORDER BY l.created_at,l.run_id`, extractionRunID, at.UTC())
	if err != nil {
		return err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, id := range ids {
		if _, err = s.ApplyLinkedUpdate(ctx, id, at); err != nil {
			return err
		}
	}
	return nil
}
