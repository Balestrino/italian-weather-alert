package publicquery

import (
	"context"
	"time"
)

// Read interpretations in one round trip while retaining each version's own
// source cutoff, latest ordinary result, reuse lineage and projection status.
func (s *Store) loadDocumentInterpretations(ctx context.Context, ids []int64, knownAt time.Time) (map[int64]documentAssessment, error) {
	result := map[int64]documentAssessment{}
	if len(ids) == 0 {
		return result, nil
	}
	rows, err := s.pool.Query(ctx, `WITH versions AS MATERIALIZED (
 SELECT v.id,d.source_id,s.product_id,registry_interpretation_cutoff(d.source_id,$2) cutoff
 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 JOIN registry_sources s ON s.id=d.source_id WHERE v.id=ANY($1::bigint[])
 ), assessments AS MATERIALIZED (
 SELECT v.*,e.status,e.reason_code,e.created_at,e.run_id,c.relevant
 FROM versions v LEFT JOIN LATERAL (
 SELECT e.status,e.reason_code,e.created_at,e.run_id FROM extraction_results e JOIN processing_runs p ON p.id=e.run_id
 WHERE p.workload<>'evaluation' AND e.document_version_id=v.id AND e.created_at<=v.cutoff
 ORDER BY e.created_at DESC,e.run_id DESC LIMIT 1) e ON true
 LEFT JOIN LATERAL (
 SELECT c.relevant FROM classification_results c JOIN processing_runs p ON p.id=c.run_id
 WHERE p.workload<>'evaluation' AND c.document_version_id=v.id AND c.created_at<=v.cutoff
 ORDER BY c.created_at DESC,c.run_id DESC LIMIT 1) c ON e.run_id IS NULL
 ) SELECT a.id,a.status,a.reason_code,a.created_at,a.run_id,COALESCE(NOT a.relevant,false),
 CASE WHEN a.status='extracted' AND a.product_id='municipal' THEN (
 WITH RECURSIVE lineage(id) AS (
 SELECT a.run_id UNION ALL SELECT u.original_run_id FROM interpretation_reuse u JOIN lineage l ON l.id=u.run_id WHERE u.created_at<=$2)
 SELECT EXISTS(SELECT 1 FROM extracted_measures m WHERE m.run_id=a.run_id AND NOT EXISTS(
 SELECT 1 FROM domain_measure_bindings b JOIN domain_local_measures projected ON projected.id=b.local_measure_id
 WHERE b.extraction_run_id IN(SELECT id FROM lineage) AND b.extraction_ordinal=m.ordinal AND projected.recorded_at<=$2))
 ) ELSE false END,
 EXISTS(SELECT 1 FROM registry_interpretation_suspensions x WHERE x.source_id=a.source_id AND x.suspended_at>=a.created_at)
 FROM assessments a`, ids, knownAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var status, reason *string
		var at *time.Time
		var run *int64
		var irrelevant, unprojected, defective bool
		if err = rows.Scan(&id, &status, &reason, &at, &run, &irrelevant, &unprojected, &defective); err != nil {
			return nil, err
		}
		value := documentAssessment{dimension: Dimension{State: "not_processed", Limitations: []string{"no interpretation was available at the knowledge boundary"}, Evidence: []Evidence{}}}
		if status == nil {
			if irrelevant {
				value.dimension.State = "supported"
				value.dimension.Limitations = []string{"classified_not_relevant"}
			}
		} else {
			state := "failed"
			if *status == "extracted" || *status == "not_applicable" {
				state = "supported"
			}
			limitation := ""
			if reason != nil {
				limitation = *reason
			}
			if unprojected {
				state = "partial"
				limitation = "validated extraction has municipal facts awaiting domain projection"
			}
			value.dimension = Dimension{State: state, Limitations: []string{limitation}, Evidence: []Evidence{}}
			if defective {
				value.dimension.State = "unreliable"
				value.dimension.Limitations = append(value.dimension.Limitations, "this interpretation predates a confirmed source defect; validated resumption does not rewrite earlier results")
			}
			value.interpretedAt = utcPointer(at)
			if run != nil {
				runID := stringID(*run)
				value.runID = &runID
			}
		}
		result[id] = value
	}
	return result, rows.Err()
}
func (s *Store) preloadDocumentInterpretations(ctx context.Context, ids []int64, knownAt time.Time) error {
	if s.reads == nil {
		return nil
	}
	missing := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, found := s.reads.interpretations[id]; !found {
			missing = append(missing, id)
		}
	}
	values, err := s.loadDocumentInterpretations(ctx, missing, knownAt)
	if err != nil {
		return err
	}
	for id, value := range values {
		s.reads.interpretations[id] = value
	}
	return nil
}

func (s *Store) loadDocumentMetadata(ctx context.Context, ids []int64, knownAt time.Time) (map[int64]Document, error) {
	result := map[int64]Document{}
	if len(ids) == 0 {
		return result, nil
	}
	rows, err := s.pool.Query(ctx, `SELECT d.id,v.id,d.source_id,issuer.name,publisher.name,d.official_url,v.content_hash,v.first_acquired_at,v.metadata
 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 JOIN `+s.sourcesSQL()+` s ON s.id=d.source_id JOIN registry_channels c ON c.id=s.channel_id
 JOIN registry_authorities publisher ON publisher.id=c.publisher_id
 LEFT JOIN registry_authorities issuer ON issuer.id=v.issuer_id
 WHERE v.id=ANY($1::bigint[]) AND v.first_acquired_at<=$2 AND (`+s.visibilitySQL()+`)`, ids, knownAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var docID, versionID int64
		var value Document
		var raw []byte
		if err = rows.Scan(&docID, &versionID, &value.SourceID, &value.Issuer, &value.Publisher, &value.OfficialURL, &value.SHA256, &value.AcquiredAt, &raw); err != nil {
			return nil, err
		}
		value.ID, value.VersionID = stringID(docID), stringID(versionID)
		value.AcquiredAt = value.AcquiredAt.UTC()
		value.Publication, value.Modification, value.Kind = metadataFields(raw)
		result[versionID] = value
	}
	return result, rows.Err()
}
func (s *Store) preloadDocumentVersions(ctx context.Context, ids []int64, knownAt time.Time) error {
	if s.reads == nil {
		return nil
	}
	if err := s.preloadDocumentInterpretations(ctx, ids, knownAt); err != nil {
		return err
	}
	missing := []int64{}
	for _, id := range ids {
		if _, found := s.reads.documents[documentReadKey{id, s.municipalityScope}]; !found {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	values, err := s.loadDocumentMetadata(ctx, missing, knownAt)
	if err != nil {
		return err
	}
	for id, value := range values {
		s.reads.documents[documentReadKey{id, s.municipalityScope}] = value
	}
	rows, err := s.pool.Query(ctx, `SELECT version_id,url,object_hash,missing FROM retained_resources WHERE version_id=ANY($1::bigint[]) AND role='attachment' ORDER BY version_id,url`, missing)
	if err != nil {
		return err
	}
	for _, id := range missing {
		s.reads.attachments[id] = []Attachment{}
	}
	for rows.Next() {
		var id int64
		var value Attachment
		var hash *string
		var absent string
		if err = rows.Scan(&id, &value.OfficialURL, &hash, &absent); err != nil {
			rows.Close()
			return err
		}
		switch {
		case hash != nil:
			value.Status = "acquired"
		case absent == "forbidden":
			value.Status = "restricted"
		default:
			value.Status = "unavailable"
		}
		s.reads.attachments[id] = append(s.reads.attachments[id], value)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	rows, err = s.pool.Query(ctx, `SELECT candidate_version_id FROM domain_verification_receipts WHERE candidate_version_id=ANY($1::bigint[])
 UNION SELECT evidence_version_id FROM domain_verification_dependencies WHERE evidence_version_id=ANY($1::bigint[])`, missing)
	if err != nil {
		return err
	}
	for _, id := range missing {
		s.reads.versionReceipts[id] = false
	}
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		s.reads.versionReceipts[id] = true
	}
	err = rows.Err()
	rows.Close()
	return err
}
