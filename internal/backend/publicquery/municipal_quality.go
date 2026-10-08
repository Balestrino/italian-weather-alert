package publicquery

import (
	"context"
	"fmt"
)

// Typed section containers and their pagination are acquisition evidence,
// not individual notices requiring a municipal measure interpretation.
const municipalNoticeSQL = `NOT EXISTS(SELECT 1 FROM registry_configurations notice_configuration,
 jsonb_array_elements_text(notice_configuration.body->'sections') section_url
 WHERE notice_configuration.source_id=s.id AND notice_configuration.revision=s.active_revision
 AND rtrim(split_part(d.official_url,'?',1),'/')=rtrim(split_part(section_url,'?',1),'/'))`

// Source acceptance and interpretation progress answer different questions.
// Summarize only latest visible primary documents at the requested boundary.
func (s *Store) municipalInterpretation(ctx context.Context, sourceID string, qt QueryTime) (Dimension, error) {
	rows, err := s.pool.Query(ctx, `SELECT v.id FROM retained_documents d
 JOIN LATERAL (SELECT * FROM retained_versions candidate WHERE candidate.document_id=d.id AND candidate.first_acquired_at<=$2 ORDER BY candidate.first_acquired_at DESC,candidate.id DESC LIMIT 1) v ON true
 JOIN `+s.sourcesSQL()+` s ON s.id=d.source_id
 WHERE s.id=$1 AND (`+s.visibilitySQL()+`) AND `+municipalNoticeSQL+` ORDER BY d.id`, sourceID, qt.KnownAt)
	if err != nil {
		return Dimension{}, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return Dimension{}, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return Dimension{}, err
	}
	result := Dimension{State: "not_processed", Limitations: []string{}, Evidence: []Evidence{}}
	if len(ids) == 0 {
		result.Limitations = append(result.Limitations, "no visible municipal document was available at the knowledge boundary")
		return result, nil
	}
	if err = s.preloadDocumentInterpretations(ctx, ids, qt.KnownAt); err != nil {
		return Dimension{}, err
	}
	supported, incomplete := 0, 0
	for _, id := range ids {
		assessment, _, _, e := s.documentInterpretation(ctx, id, qt.KnownAt)
		if e != nil {
			return Dimension{}, e
		}
		if assessment.State == "supported" {
			supported++
		} else {
			incomplete++
		}
	}
	if supported > 0 {
		result.State = "supported"
	}
	if incomplete > 0 {
		result.State = "partial"
	}
	result.Limitations = append(result.Limitations, fmt.Sprintf("latest visible municipal documents: %d interpreted; %d require interpretation or projection", supported, incomplete))
	return result, nil
}
