package interpretation

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/territory"
	"time"
)

// WaitingEquivalentSQL intentionally tests validated domain results rather than
// queue success (a guarded/skipped job is not a valid interpretation).
const WaitingEquivalentSQL = `SELECT p.document_version_id FROM interpretation_preflight p
 LEFT JOIN LATERAL (SELECT c.status,c.relevant,c.run_id FROM processing_runs r LEFT JOIN classification_results c ON c.run_id=r.id WHERE r.document_version_id=p.representative_version_id AND r.stage='classification' ORDER BY r.created_at DESC,r.id DESC LIMIT 1) c ON true
 LEFT JOIN LATERAL (SELECT e.status,e.classification_run_id FROM processing_runs r LEFT JOIN extraction_results e ON e.run_id=r.id WHERE r.document_version_id=p.representative_version_id AND r.stage='extraction' ORDER BY r.created_at DESC,r.id DESC LIMIT 1) e ON true
 WHERE p.document_version_id<>p.representative_version_id AND
 (c.status IS DISTINCT FROM 'classified' OR c.relevant IS NULL OR (c.relevant AND (e.status IS DISTINCT FROM 'extracted' OR e.classification_run_id IS DISTINCT FROM c.run_id)))`

func (s *Scheduler) Equivalent(ctx context.Context, id int64) (equivalent, waiting bool, err error) {
	err = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM interpretation_preflight WHERE document_version_id=$1 AND representative_version_id<>document_version_id)`, id).Scan(&equivalent)
	if err != nil || !equivalent {
		return
	}
	err = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM (`+WaitingEquivalentSQL+`) w WHERE w.document_version_id=$1)`, id).Scan(&waiting)
	return
}

// DeferEquivalent protects already queued copies without claiming attempts.
// Their version-specific jobs remain available to materialize validated reuse.
func (s *Scheduler) DeferEquivalent(ctx context.Context, at time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE processing_jobs j SET available_at=GREATEST(j.available_at,$1),updated_at=$2
 WHERE j.archived_at IS NULL AND j.state IN('queued','retry_wait') AND j.available_at<$1
 AND COALESCE(j.payload->>'workload','ordinary') NOT IN('evaluation','reprocessing')
 AND j.payload->>'document_version_id' IN (SELECT document_version_id::text FROM (`+WaitingEquivalentSQL+`) w)`, at.Add(time.Minute), at)
	return err
}

// ReleaseEquivalent materializes ready copies locally after their representative
// finishes. The durable trigger survives restarts; job identities make concurrent
// reconcilers harmless. Existing queued/failed jobs retain their original policy.
func (s *Scheduler) ReleaseEquivalent(ctx context.Context, at time.Time) error {
	gate, err := territory.Predicate(ctx, s.pool, "d.source_id", false)
	if err != nil {
		return err
	}
	rows, err := s.pool.Query(ctx, `SELECT t.document_version_id,t.evidence_hash,t.workload
 FROM interpretation_triggers t
 JOIN interpretation_preflight p ON p.document_version_id=t.document_version_id
 JOIN retained_versions v ON v.id=t.document_version_id
 JOIN retained_documents d ON d.id=v.document_id
 JOIN registry_sources src ON src.id=d.source_id
 WHERE t.classification_job_id IS NULL AND p.representative_version_id<>p.document_version_id
 AND src.interpretation_suspended_at IS NULL
 AND t.workload NOT IN ('evaluation','reprocessing')
 AND NOT EXISTS(SELECT 1 FROM interpretation_archives a WHERE a.document_version_id IN(p.document_version_id,p.representative_version_id))
 AND NOT EXISTS(SELECT 1 FROM (`+WaitingEquivalentSQL+`) w WHERE w.document_version_id=t.document_version_id)
 AND NOT EXISTS(SELECT 1 FROM processing_jobs j WHERE j.payload->>'document_version_id'=t.document_version_id::text AND j.kind IN ($1,$2))
 `+gate+` ORDER BY t.document_version_id LIMIT 25`, ocr.Kind, classification.Kind)
	if err != nil {
		return err
	}
	var pending []AcquisitionEvent
	for rows.Next() {
		var e AcquisitionEvent
		if err = rows.Scan(&e.DocumentVersionID, &e.EvidenceHash, &e.Workload); err != nil {
			rows.Close()
			return err
		}
		e.At = at
		e.ContentChanged = true
		pending = append(pending, e)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, e := range pending {
		if _, err = s.Automatic(ctx, e); err != nil {
			return err
		}
	}
	return nil
}
