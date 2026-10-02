package operations

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type JobCounts struct {
	Queued, Running, RetryWait, Failed, Succeeded, Archived int64
	Due, ExpiredLeases                                      int64
}

type JobGroup struct {
	Queue, Kind string
	JobCounts
	OldestPending, LastFinished *time.Time
}

type JobReason struct {
	Queue, Kind, State, Code string
	Count                    int64
	LastObserved             time.Time
}

type DocumentBreakdown struct {
	SourceID                                                                                 string
	Total, Unscheduled, Preparing, Queued, Running, RetryWait, Failed, Incomplete, Suspended int64
}

type HistoryBucket struct {
	At                                  time.Time
	Succeeded, Retry, Failed, Abandoned int64
}

type Dashboard struct {
	Overview
	Period      string
	HistoryFrom time.Time
	Jobs        JobCounts
	Groups      []JobGroup
	Reasons     []JobReason
	Documents   []DocumentBreakdown
	History     []HistoryBucket
}

// HistoryWindow uses UTC calendar buckets; the current bucket is partial.
func HistoryWindow(period string, at time.Time) (time.Time, time.Duration, error) {
	if at.IsZero() {
		return time.Time{}, 0, ErrInvalid
	}
	switch period {
	case "24h":
		return at.UTC().Truncate(time.Hour).Add(-23 * time.Hour), time.Hour, nil
	case "7d":
		return at.UTC().Truncate(24 * time.Hour).Add(-6 * 24 * time.Hour), 24 * time.Hour, nil
	case "30d":
		return at.UTC().Truncate(24 * time.Hour).Add(-29 * 24 * time.Hour), 24 * time.Hour, nil
	default:
		return time.Time{}, 0, ErrInvalid
	}
}

// Dashboard reads persisted queue facts only, without contacting workers/providers.
// All aggregates share a read-only snapshot, including the document list predicate.
func (s *Store) Dashboard(ctx context.Context, period string, at time.Time) (Dashboard, error) {
	from, step, err := HistoryWindow(period, at)
	result := Dashboard{Period: period, HistoryFrom: from}
	if err != nil {
		return result, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	result.Overview, err = overview(ctx, tx, at)
	if err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, `SELECT queue,kind,
 count(*) FILTER(WHERE archived_at IS NULL AND state='queued'),
 count(*) FILTER(WHERE archived_at IS NULL AND state='running'),
 count(*) FILTER(WHERE archived_at IS NULL AND state='retry_wait'),
 count(*) FILTER(WHERE archived_at IS NULL AND state='failed'),
 count(*) FILTER(WHERE archived_at IS NULL AND state='succeeded'),
 count(*) FILTER(WHERE archived_at IS NOT NULL),
 count(*) FILTER(WHERE archived_at IS NULL AND state IN ('queued','retry_wait') AND available_at<=$1),
 count(*) FILTER(WHERE archived_at IS NULL AND state='running' AND lease_expires_at<=$1),
 min(created_at) FILTER(WHERE archived_at IS NULL AND state IN ('queued','retry_wait')),
 max(completed_at) FILTER(WHERE completed_at<=$1)
 FROM processing_jobs GROUP BY queue,kind ORDER BY queue,kind`, at.UTC())
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var g JobGroup
		if err = rows.Scan(&g.Queue, &g.Kind, &g.Queued, &g.Running, &g.RetryWait, &g.Failed, &g.Succeeded, &g.Archived, &g.Due, &g.ExpiredLeases, &g.OldestPending, &g.LastFinished); err != nil {
			rows.Close()
			return result, err
		}
		result.Groups = append(result.Groups, g)
		result.Jobs.Queued += g.Queued
		result.Jobs.Running += g.Running
		result.Jobs.RetryWait += g.RetryWait
		result.Jobs.Failed += g.Failed
		result.Jobs.Succeeded += g.Succeeded
		result.Jobs.Archived += g.Archived
		result.Jobs.Due += g.Due
		result.Jobs.ExpiredLeases += g.ExpiredLeases
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return result, err
	}
	rows.Close()
	rows, err = tx.Query(ctx, `SELECT queue,kind,state,last_error_code,count(*),max(updated_at)
 FROM processing_jobs WHERE archived_at IS NULL AND state<>'succeeded' AND last_error_code IS NOT NULL
 GROUP BY queue,kind,state,last_error_code ORDER BY count(*) DESC,queue,kind,state,last_error_code`)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var r JobReason
		if err = rows.Scan(&r.Queue, &r.Kind, &r.State, &r.Code, &r.Count, &r.LastObserved); err != nil {
			rows.Close()
			return result, err
		}
		result.Reasons = append(result.Reasons, r)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return result, err
	}
	rows.Close()
	// Count each pending version once. Multiple OCR jobs cannot inflate documents.
	pending := `WITH filter AS (SELECT ''::text source_id,0::bigint version_id,0::bigint run_id,$1::timestamptz observed_at,0::bigint job_id,''::text state), pending AS (` + queries["documents"] + `), work AS (
 SELECT coalesce(j.payload->>'document_version_id',e.document_version_id::text) version_id,
 bool_or(j.state='running') running,bool_or(j.state='retry_wait') retry_wait,
 bool_or(j.state='queued') queued,bool_or(j.state='failed') failed
 FROM processing_jobs j LEFT JOIN extraction_results e ON e.run_id::text=j.payload->>'extraction_run_id'
 WHERE j.archived_at IS NULL AND j.kind IN ('classify_relevance','ocr_resource','extract_measures','embed_measure','link_measure_update')
 GROUP BY 1), categorized AS (
 SELECT p.source_id, CASE
 WHEN p.interpretation_suspended_at IS NOT NULL THEN 'suspended'
 WHEN work.running THEN 'running' WHEN work.retry_wait THEN 'retry_wait' WHEN work.queued THEN 'queued'
 WHEN work.failed THEN 'failed'
 WHEN p.classification_status<>'not_processed' THEN 'incomplete'
 WHEN t.document_version_id IS NULL THEN 'unscheduled' ELSE 'preparing' END category
 FROM pending p LEFT JOIN interpretation_triggers t ON t.document_version_id=p.document_version_id
 LEFT JOIN work ON work.version_id=p.document_version_id::text)
 SELECT source_id,count(*),
 count(*) FILTER(WHERE category='unscheduled'),count(*) FILTER(WHERE category='preparing'),
 count(*) FILTER(WHERE category='queued'),count(*) FILTER(WHERE category='running'),count(*) FILTER(WHERE category='retry_wait'),
 count(*) FILTER(WHERE category='failed'),count(*) FILTER(WHERE category='incomplete'),count(*) FILTER(WHERE category='suspended')
 FROM categorized GROUP BY source_id ORDER BY count(*) DESC,source_id`
	rows, err = tx.Query(ctx, pending, at.UTC())
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var d DocumentBreakdown
		if err = rows.Scan(&d.SourceID, &d.Total, &d.Unscheduled, &d.Preparing, &d.Queued, &d.Running, &d.RetryWait, &d.Failed, &d.Incomplete, &d.Suspended); err != nil {
			rows.Close()
			return result, err
		}
		result.Documents = append(result.Documents, d)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return result, err
	}
	rows.Close()
	rows, err = tx.Query(ctx, `WITH buckets AS (SELECT generate_series($1::timestamptz,$2::timestamptz,$3::interval) at),
 totals AS (SELECT date_bin($3::interval,finished_at,$1::timestamptz) at,
 count(*) FILTER(WHERE outcome='succeeded') succeeded,count(*) FILTER(WHERE outcome='retry') retry,
 count(*) FILTER(WHERE outcome='failed') failed,count(*) FILTER(WHERE outcome='abandoned') abandoned
 FROM processing_attempts WHERE finished_at>=$1 AND finished_at<$2 GROUP BY 1)
 SELECT b.at,coalesce(t.succeeded,0),coalesce(t.retry,0),coalesce(t.failed,0),coalesce(t.abandoned,0)
 FROM buckets b LEFT JOIN totals t ON t.at=b.at WHERE b.at<=$2 ORDER BY b.at`, from, at.UTC(), step.String())
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var h HistoryBucket
		if err = rows.Scan(&h.At, &h.Succeeded, &h.Retry, &h.Failed, &h.Abandoned); err != nil {
			rows.Close()
			return result, err
		}
		result.History = append(result.History, h)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return result, err
	}
	rows.Close()
	return result, tx.Commit(ctx)
}
