package interpretation

import (
	"context"
	"strings"
	"time"
)

func (s *Scheduler) Archived(ctx context.Context, versionID int64) (bool, error) {
	var archived bool
	err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM interpretation_archives WHERE document_version_id=$1)", versionID).Scan(&archived)
	return archived, err
}

// ArchivePending retires unfinished interpretations at one snapshot boundary.
// Queue work must already be retired; no originals, results or usage are edited.
func (s *Scheduler) ArchivePending(ctx context.Context, cutoff time.Time, actor string, at time.Time) (int64, error) {
	if cutoff.IsZero() || at.IsZero() || cutoff.After(at) || strings.TrimSpace(actor) == "" || len(actor) > 200 {
		return 0, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "LOCK TABLE processing_jobs IN SHARE ROW EXCLUSIVE MODE"); err != nil {
		return 0, err
	}
	var unfinished bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM processing_jobs WHERE created_at<=$1 AND archived_at IS NULL AND state<>'succeeded')", cutoff.UTC()).Scan(&unfinished); err != nil {
		return 0, err
	}
	if unfinished {
		return 0, ErrInvalid
	}
	tag, err := tx.Exec(ctx, `INSERT INTO interpretation_archives(document_version_id,archived_at,cutoff,actor)
 SELECT v.id,$2,$1,$3 FROM retained_versions v
 LEFT JOIN LATERAL (SELECT id FROM processing_runs WHERE document_version_id=v.id AND stage='classification' ORDER BY created_at DESC,id DESC LIMIT 1) cr ON true
 LEFT JOIN classification_results c ON c.run_id=cr.id
 LEFT JOIN LATERAL (SELECT id FROM processing_runs WHERE document_version_id=v.id AND stage='extraction' ORDER BY created_at DESC,id DESC LIMIT 1) er ON true
 LEFT JOIN extraction_results e ON e.run_id=er.id
 WHERE v.first_acquired_at<=$1 AND (c.status IS NULL OR c.status='undetermined'
 OR (c.relevant AND (e.status IS DISTINCT FROM 'extracted' OR e.classification_run_id IS DISTINCT FROM c.run_id)))
 ON CONFLICT(document_version_id) DO NOTHING`, cutoff.UTC(), at.UTC(), actor)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
