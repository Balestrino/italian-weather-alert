package interpretation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"github.com/jackc/pgx/v5"
)

type ArchiveRecovery struct {
	ID         string            `json:"id"`
	SourceID   string            `json:"source_id"`
	Revision   int               `json:"revision"`
	Actor      string            `json:"actor"`
	VersionIDs []int64           `json:"version_ids"`
	Regression string            `json:"regression_id"`
	Evidence   registry.Evidence `json:"evidence"`
	CreatedAt  time.Time         `json:"created_at"`
}

// RecoverArchived authorizes only a new, exact reprocessing selection. The
// archival marker remains effective for automatic and unrelated queued work.
func (s *Scheduler) RecoverArchived(ctx context.Context, r ArchiveRecovery) (ArchiveRecovery, error) {
	u, err := url.Parse(r.Evidence.URL)
	validName := func(v string) bool {
		return v != "" && v == strings.TrimSpace(v) && len(v) <= 180 && !strings.ContainsAny(v, "\r\n")
	}
	if s == nil || s.pool == nil || s.Queue == nil || !validName(r.ID) || !validName(r.SourceID) || !validName(r.Actor) || !validName(r.Regression) || r.Revision < 1 || len(r.VersionIDs) < 1 || len(r.VersionIDs) > 100 || err != nil || u.Host == "" || u.User != nil || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") || strings.TrimSpace(r.Evidence.Locator) == "" || r.Evidence.ObservedAt.IsZero() || r.Evidence.ObservedAt.After(time.Now().UTC()) {
		return ArchiveRecovery{}, ErrInvalid
	}
	r.VersionIDs = slices.Clone(r.VersionIDs)
	slices.Sort(r.VersionIDs)
	for i, id := range r.VersionIDs {
		if id < 1 || i > 0 && id == r.VersionIDs[i-1] {
			return ArchiveRecovery{}, ErrInvalid
		}
	}
	r.CreatedAt = time.Time{}
	body, err := json.Marshal(r)
	if err != nil {
		return ArchiveRecovery{}, ErrInvalid
	}
	hash := sha256.Sum256(body)
	digest := hex.EncodeToString(hash[:])
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ArchiveRecovery{}, err
	}
	defer tx.Rollback(ctx)
	var active *int
	if err = tx.QueryRow(ctx, "SELECT active_revision FROM registry_sources WHERE id=$1 FOR UPDATE", r.SourceID).Scan(&active); err != nil {
		return ArchiveRecovery{}, err
	}
	if active == nil || *active != r.Revision {
		return ArchiveRecovery{}, registry.ErrConflict
	}
	var ready bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM registry_regressions r WHERE r.id=$1 AND r.source_id=$2 AND r.revision=$3 AND r.passed
 AND r.id=(SELECT latest.id FROM registry_regressions latest WHERE latest.source_id=r.source_id AND latest.suite=r.suite AND latest.revision<=$3 ORDER BY latest.revision DESC,latest.recorded_at DESC,latest.id DESC LIMIT 1))
 AND NOT EXISTS(SELECT 1 FROM (SELECT DISTINCT ON(suite) passed,revision FROM registry_regressions WHERE source_id=$2 AND revision<=$3 ORDER BY suite,revision DESC,recorded_at DESC,id DESC) latest WHERE NOT passed OR revision<>$3)`, r.Regression, r.SourceID, r.Revision).Scan(&ready)
	if err != nil {
		return ArchiveRecovery{}, err
	}
	if !ready {
		return ArchiveRecovery{}, registry.ErrPrerequisite
	}
	for _, id := range r.VersionIDs {
		if err = territory.CheckVersion(ctx, tx, id); err != nil {
			return ArchiveRecovery{}, err
		}
		var valid bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 JOIN interpretation_archives a ON a.document_version_id=v.id
 WHERE v.id=$1 AND d.source_id=$2 AND v.complete
 AND EXISTS(SELECT 1 FROM retained_acquisitions x WHERE x.version_id=v.id AND x.source_id=$2 AND x.configuration=$3 AND x.acquired_at<=$4))`, id, r.SourceID, r.Revision, r.Evidence.ObservedAt.UTC()).Scan(&valid)
		if err != nil {
			return ArchiveRecovery{}, err
		}
		if !valid {
			return ArchiveRecovery{}, registry.ErrPrerequisite
		}
	}
	var previous string
	err = tx.QueryRow(ctx, "SELECT selection_hash FROM interpretation_archive_recoveries WHERE request_id=$1", r.ID).Scan(&previous)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ArchiveRecovery{}, err
	}
	if err == nil && previous != digest {
		return ArchiveRecovery{}, registry.ErrConflict
	}
	if errors.Is(err, pgx.ErrNoRows) {
		r.CreatedAt = time.Now().UTC()
		tag, err := tx.Exec(ctx, `INSERT INTO interpretation_reprocessing_requests(id,source_id,actor,created_at) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, r.ID, r.SourceID, r.Actor, r.CreatedAt)
		if err != nil {
			return ArchiveRecovery{}, err
		}
		if tag.RowsAffected() == 0 {
			return ArchiveRecovery{}, registry.ErrConflict
		}
		evidence, _ := json.Marshal(r.Evidence)
		if _, err = tx.Exec(ctx, `INSERT INTO interpretation_archive_recoveries(request_id,revision,regression_id,evidence,selection_hash) VALUES($1,$2,$3,$4,$5)`, r.ID, r.Revision, r.Regression, evidence, digest); err != nil {
			return ArchiveRecovery{}, err
		}
		for _, id := range r.VersionIDs {
			if _, err = tx.Exec(ctx, "INSERT INTO interpretation_reprocessing_selections(request_id,document_version_id) VALUES($1,$2)", r.ID, id); err != nil {
				return ArchiveRecovery{}, err
			}
			if _, err = tx.Exec(ctx, "INSERT INTO interpretation_archive_recovery_versions(request_id,document_version_id) VALUES($1,$2)", r.ID, id); err != nil {
				return ArchiveRecovery{}, err
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return ArchiveRecovery{}, err
	}
	// A queue/storage failure can be retried with the same audited selection.
	for _, id := range r.VersionIDs {
		job, err := classification.Enqueue(ctx, s.Queue, s.Policy, classification.Payload{DocumentVersionID: id, Workload: "reprocessing", SelectionID: r.ID}, time.Now().UTC())
		if err != nil {
			return ArchiveRecovery{}, err
		}
		if _, err = s.pool.Exec(ctx, "UPDATE interpretation_reprocessing_selections SET classification_job_id=$3 WHERE request_id=$1 AND document_version_id=$2 AND classification_job_id IS NULL", r.ID, id, job.ID); err != nil {
			return ArchiveRecovery{}, err
		}
	}
	rows, err := s.ArchiveRecoveries(ctx, r.SourceID)
	if err != nil {
		return ArchiveRecovery{}, err
	}
	for _, stored := range rows {
		if stored.ID == r.ID {
			return stored, nil
		}
	}
	return ArchiveRecovery{}, ErrInvalid
}

func (s *Scheduler) ArchiveRecoveries(ctx context.Context, source string) ([]ArchiveRecovery, error) {
	rows, err := s.pool.Query(ctx, `SELECT r.id,r.source_id,r.actor,r.created_at,a.revision,a.regression_id,a.evidence,
 ARRAY(SELECT v.document_version_id FROM interpretation_archive_recovery_versions v WHERE v.request_id=r.id ORDER BY v.document_version_id)
 FROM interpretation_archive_recoveries a JOIN interpretation_reprocessing_requests r ON r.id=a.request_id WHERE r.source_id=$1 ORDER BY r.created_at,r.id`, source)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []ArchiveRecovery{}
	for rows.Next() {
		var r ArchiveRecovery
		var evidence []byte
		if err = rows.Scan(&r.ID, &r.SourceID, &r.Actor, &r.CreatedAt, &r.Revision, &r.Regression, &evidence, &r.VersionIDs); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(evidence, &r.Evidence); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s *Scheduler) archiveRecoveryAllowed(ctx context.Context, versionID int64, selection string, classificationRunID, extractionRunID int64) (bool, error) {
	var allowed bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM interpretation_archive_recovery_versions v
 JOIN interpretation_archive_recoveries recovery ON recovery.request_id=v.request_id
 JOIN interpretation_reprocessing_requests request ON request.id=v.request_id
 JOIN registry_sources source ON source.id=request.source_id AND source.active_revision=recovery.revision
 WHERE v.document_version_id=$1 AND
 (v.request_id=NULLIF($2,'') OR EXISTS(SELECT 1 FROM classification_results c JOIN processing_run_attempts a ON a.run_id=c.run_id
 JOIN processing_jobs j ON j.id=a.queue_job_id WHERE c.document_version_id=$1 AND c.status='classified' AND c.relevant AND j.payload->>'workload'='reprocessing' AND j.payload->>'selection_id'=v.request_id
 AND (c.run_id=$3 OR c.run_id=(SELECT e.classification_run_id FROM extraction_results e WHERE e.run_id=$4 AND e.document_version_id=$1 AND e.status='extracted' AND e.content_complete)))))`, versionID, selection, classificationRunID, extractionRunID).Scan(&allowed)
	return allowed, err
}
