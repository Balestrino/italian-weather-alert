package publicquery

import (
	"context"
	"strconv"
	"time"
)

const carriedForwardLocator = "equivalent_version_interpretation"

// Earlier versions of the same document whose manifest fingerprint matches a
// manifest of the target, both recorded within the source interpretation
// cutoff. Ordered by recency so the latest validated equivalent wins.
const equivalentCandidatesSQL = `WITH targets AS MATERIALIZED (
 SELECT v.id,v.document_id,v.first_acquired_at,registry_interpretation_cutoff(d.source_id,$2) cutoff
 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id WHERE v.id=ANY($1::bigint[])
), prints AS (
 SELECT DISTINCT t.id,e.fingerprint FROM targets t
 JOIN interpretation_input_equivalences e ON e.version_id=t.id
 JOIN processing_runs r ON r.id=e.run_id AND r.created_at<=t.cutoff
)
SELECT DISTINCT p.id,c.version_id,cv.first_acquired_at,d.id,d.official_url FROM prints p JOIN targets t ON t.id=p.id
 JOIN interpretation_input_equivalences c ON c.document_id=t.document_id AND c.fingerprint=p.fingerprint
 JOIN processing_runs cr ON cr.id=c.run_id AND cr.created_at<=t.cutoff
 JOIN retained_versions cv ON cv.id=c.version_id AND (cv.first_acquired_at,cv.id)<(t.first_acquired_at,t.id) AND cv.first_acquired_at<=$2
 JOIN retained_documents d ON d.id=cv.document_id
 ORDER BY p.id,cv.first_acquired_at DESC,c.version_id DESC`

// carryForwardEquivalent presents, for a failed or pending version, the most
// recent validated interpretation of an earlier equivalent version. It only
// changes the returned assessment; retained results are never written.
func (s *Store) carryForwardEquivalent(ctx context.Context, result map[int64]documentAssessment, own map[int64]presentedRun, knownAt time.Time) error {
	targets := []int64{}
	for id, value := range result {
		if value.dimension.State == "failed" || value.dimension.State == "not_processed" {
			targets = append(targets, id)
		}
	}
	if len(targets) == 0 {
		return nil
	}
	rows, err := s.pool.Query(ctx, equivalentCandidatesSQL, targets, knownAt)
	if err != nil {
		return err
	}
	type candidate struct {
		version, document int64
		url               string
	}
	candidates := map[int64][]candidate{}
	ids := []int64{}
	seen := map[int64]bool{}
	for rows.Next() {
		var target, version, document int64
		var acquired time.Time
		var url string
		if err = rows.Scan(&target, &version, &acquired, &document, &url); err != nil {
			rows.Close()
			return err
		}
		candidates[target] = append(candidates[target], candidate{version, document, url})
		if !seen[version] {
			seen[version] = true
			ids = append(ids, version)
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil || len(ids) == 0 {
		return err
	}
	assessed, origins, err := s.assessDocumentVersions(ctx, ids, knownAt)
	if err != nil {
		return err
	}
	for target, list := range candidates {
		for _, c := range list {
			earlier, found := assessed[c.version]
			if !found || (earlier.dimension.State != "supported" && earlier.dimension.State != "partial") {
				continue
			}
			result[target] = carriedForward(result[target], own[target], earlier, origins[c.version], c.document, c.version, c.url)
			break
		}
	}
	return nil
}

func carriedForward(current documentAssessment, own presentedRun, earlier documentAssessment, origin presentedRun, document, version int64, url string) documentAssessment {
	run, configuration := "none", "unknown"
	if origin.run != nil {
		run = strconv.FormatInt(*origin.run, 10)
	}
	if origin.configuration != nil {
		configuration = *origin.configuration
	}
	outcome := "own_interpretation:not_available"
	if current.dimension.State == "failed" {
		code := own.reason
		if code == "" {
			code = "unknown"
		}
		outcome = "own_interpretation:failed:" + code
	}
	dimension := cloneDimension(earlier.dimension)
	dimension.Limitations = append([]string{
		"carried_forward_from_equivalent_version:" + strconv.FormatInt(version, 10) + ":run:" + run + ":configuration:" + configuration,
		outcome,
	}, dimension.Limitations...)
	locator := carriedForwardLocator
	dimension.Evidence = append(dimension.Evidence, Evidence{DocumentID: strconv.FormatInt(document, 10), VersionID: strconv.FormatInt(version, 10), SourceURL: url, Locator: &locator})
	return documentAssessment{dimension: dimension, interpretedAt: earlier.interpretedAt, runID: earlier.runID}
}
