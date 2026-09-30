// Package operations provides read-only administrative projections. Queries
// select explicit diagnostic fields, never provider settings or job payloads.
package operations

import (
	"context"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/interpretation"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalid = errors.New("invalid operations filter")

type Filter struct {
	Issue     string `json:"issue,omitempty"`
	State     string `json:"state,omitempty"`
	SourceID  string `json:"source_id"`
	VersionID int64  `json:"document_version_id"`
	RunID     int64  `json:"run_id"`
	JobID     int64  `json:"job_id"`
	Page      int    `json:"page"`
}
type Table struct {
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
}
type Report struct {
	Section                   string    `json:"section"`
	Filter                    Filter    `json:"filter"`
	ObservedAt                time.Time `json:"observed_at"`
	Table                     Table     `json:"table"`
	Units                     *Table    `json:"other_unit_totals,omitempty"`
	UnavailableCostCategories []string  `json:"unavailable_cost_categories,omitempty"`
	Totals                    *Table    `json:"totals,omitempty"`
	More                      bool      `json:"more"`
}
type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func readTable(ctx context.Context, tx pgx.Tx, query string, args ...any) (Table, error) {
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return Table{}, err
	}
	defer rows.Close()
	result := Table{Rows: []map[string]any{}, Columns: []string{}}
	for _, f := range rows.FieldDescriptions() {
		result.Columns = append(result.Columns, f.Name)
	}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return Table{}, err
		}
		item := map[string]any{}
		for i, v := range values {
			item[result.Columns[i]] = v
		}
		result.Rows = append(result.Rows, item)
	}
	return result, rows.Err()
}

// Page uses one snapshot for the page and its totals. Pagination is for live
// diagnosis; totals cover the full selected scope, including every attempt.
func (s *Store) Page(ctx context.Context, section string, f Filter, at time.Time) (Report, error) {
	query, ok := queries[section]
	if !ok || f.Page < 1 || f.Page > 1000000 || f.VersionID < 0 || f.RunID < 0 || f.JobID < 0 || len(f.SourceID) > 200 || at.IsZero() {
		return Report{}, ErrInvalid
	}
	if f.Issue != "" && (section != "sources" || f.Issue != "attention") {
		return Report{}, ErrInvalid
	}
	if f.State != "" && (section != "jobs" || !ValidJobState(f.State)) {
		return Report{}, ErrInvalid
	}
	if section == "sources" && (f.VersionID != 0 || f.RunID != 0) || section != "usage" && f.RunID != 0 || section != "jobs" && f.JobID != 0 {
		return Report{}, ErrInvalid
	}
	query = filteredQuery(section, f)
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Report{}, err
	}
	defer tx.Rollback(ctx)
	result := Report{Section: section, Filter: f, ObservedAt: at.UTC()}
	// Typed parameters are shared by every projection so unused filters cannot
	// produce unknown-parameter errors or silently broaden document selection.
	prefix := `WITH filter AS (SELECT $1::text source_id,$2::bigint version_id,$3::bigint run_id,$4::timestamptz observed_at,$5::bigint job_id,$6::text state) `
	args := []any{f.SourceID, f.VersionID, f.RunID, at.UTC(), f.JobID, f.State}
	result.Table, err = readTable(ctx, tx, prefix+query+` LIMIT 101 OFFSET $7`, append(args, (f.Page-1)*100)...)
	if err != nil {
		return Report{}, err
	}
	if len(result.Table.Rows) > 100 {
		result.More = true
		result.Table.Rows = result.Table.Rows[:100]
	}
	if section == "usage" {
		totals, e := readTable(ctx, tx, prefix+usageTotals, args...)
		if e != nil {
			return Report{}, e
		}
		result.Totals = &totals
		units, e := readTable(ctx, tx, prefix+usageUnits, args...)
		if e != nil {
			return Report{}, e
		}
		result.Units = &units
		result.UnavailableCostCategories = []string{"hosting", "storage", "local_compute"}
	}
	if err = tx.Commit(ctx); err != nil {
		return Report{}, err
	}
	return result, nil
}

const usageFrom = ` FROM processing_runs r
 JOIN processing_configuration_versions c ON c.id=r.configuration_version_id
 LEFT JOIN processing_model_versions m ON m.id=c.model_version_id
 LEFT JOIN processing_run_attempts a ON a.run_id=r.id
 LEFT JOIN processing_price_versions p ON p.id=a.price_version_id
 CROSS JOIN filter f
 WHERE (f.source_id='' OR r.source_id=f.source_id)
 AND (f.version_id=0 OR r.document_version_id=f.version_id)
 AND (f.run_id=0 OR r.id=f.run_id) `

const usageUnits = `SELECT r.workload,r.stage,m.provider,m.model,u.key AS metric,sum(u.value::bigint)::bigint AS known_units
 FROM processing_runs r JOIN processing_configuration_versions c ON c.id=r.configuration_version_id
 LEFT JOIN processing_model_versions m ON m.id=c.model_version_id
 JOIN processing_run_attempts a ON a.run_id=r.id
 CROSS JOIN LATERAL jsonb_each_text(a.other_units) u CROSS JOIN filter f
 WHERE (f.source_id='' OR r.source_id=f.source_id) AND (f.version_id=0 OR r.document_version_id=f.version_id) AND (f.run_id=0 OR r.id=f.run_id)
 GROUP BY r.workload,r.stage,m.provider,m.model,u.key ORDER BY r.workload,r.stage,m.provider,m.model,u.key`

const usageTotals = `SELECT r.workload,r.stage,m.provider,m.model,p.currency,
 count(DISTINCT r.id) AS runs,count(a.number) AS attempts,
 count(*) FILTER(WHERE a.number>1) AS retries,
 count(*) FILTER(WHERE a.outcome='running') AS running_attempts,
 count(*) FILTER(WHERE a.usage_status IN ('partial','unavailable')) AS unavailable_usage_attempts,
 count(*) FILTER(WHERE a.cost_status='unknown') AS unknown_cost_attempts,
 sum(a.duration_ms)::bigint AS measured_duration_ms,
 sum(a.input_tokens)::bigint AS known_input_tokens,sum(a.output_tokens)::bigint AS known_output_tokens,
 sum(a.cache_read_tokens)::bigint AS known_cache_read_tokens,sum(a.cache_write_tokens)::bigint AS known_cache_write_tokens,
 sum(a.estimated_cost_microunits)::bigint AS known_cost_microunits,
 (count(*) FILTER(WHERE a.number IS NULL OR a.cost_status IS NULL OR a.cost_status='unknown')=0) AS cost_complete
 ` + usageFrom + ` GROUP BY r.workload,r.stage,m.provider,m.model,p.currency ORDER BY r.workload,r.stage,m.provider,m.model,p.currency`

var queries = map[string]string{
	"sources": `SELECT s.id AS source_id,s.latest_revision,s.active_revision,s.collection_enabled,s.public_enabled,s.interpretation_suspended_at,
 a.configuration AS checked_configuration,a.last_reachable_at,a.last_content_at,a.last_complete_at,
 CASE WHEN a.last_complete_at IS NULL OR s.active_revision IS DISTINCT FROM a.configuration THEN 'not_yet_verified'
 WHEN f.observed_at>=a.last_complete_at+make_interval(secs=>a.delay_seconds) THEN 'delayed' ELSE 'current' END AS updating_state,
 a.last_check_state,a.publication_state,a.last_error_code,a.first_error_at,a.last_error_at,a.consecutive_failures,a.next_check_at,a.retry_after,
 checks.checks,checks.measured_duration_ms
 FROM registry_sources s CROSS JOIN filter f LEFT JOIN acquisition_source_status a ON a.source_id=s.id
 LEFT JOIN LATERAL (SELECT count(*) AS checks,sum((extract(epoch FROM finished_at-started_at)*1000)::bigint)::bigint AS measured_duration_ms FROM acquisition_checks WHERE source_id=s.id) checks ON true
 WHERE (f.source_id='' OR s.id=f.source_id) ORDER BY s.id`,
	"jobs": `SELECT j.id AS job_id,j.queue,j.kind,j.state,j.attempt_count,j.max_attempts,j.attempt_budget,j.last_error_code,j.available_at,j.completed_at,
 history.attempts,history.measured_duration_ms,
 (SELECT jsonb_agg(jsonb_build_object('after_attempt',after_attempt,'actor',actor,'created_at',created_at) ORDER BY after_attempt) FROM processing_job_relaunches WHERE job_id=j.id) AS relaunches,
 d.source_id,d.id AS document_id,v.id AS document_version_id,d.official_url
 FROM processing_jobs j CROSS JOIN filter f
 LEFT JOIN extraction_results je ON je.run_id::text=j.payload->>'extraction_run_id'
 LEFT JOIN retained_acquisitions ja ON ja.id=j.payload->>'acquisition_id'
 LEFT JOIN retained_versions v ON v.id::text=COALESCE(j.payload->>'document_version_id',je.document_version_id::text,ja.version_id::text)
 LEFT JOIN retained_documents d ON d.id=COALESCE(v.document_id,ja.document_id)
 LEFT JOIN LATERAL (SELECT jsonb_agg(jsonb_build_object('number',number,'outcome',outcome,'error_code',error_code,'started_at',started_at,'finished_at',finished_at) ORDER BY number) AS attempts,
 sum((extract(epoch FROM finished_at-started_at)*1000)::bigint)::bigint AS measured_duration_ms FROM processing_attempts WHERE job_id=j.id) history ON true
 WHERE j.archived_at IS NULL AND (f.job_id=0 OR j.id=f.job_id)
 AND (f.state='' OR j.state=f.state)
 AND (f.source_id='' OR d.source_id=f.source_id) AND (f.version_id=0 OR v.id=f.version_id)
 ORDER BY j.updated_at DESC,j.id DESC`,
	"documents": `SELECT d.source_id,d.id AS document_id,v.id AS document_version_id,d.official_url,v.first_acquired_at,
 cr.id AS classification_run_id,coalesce(c.status,'not_processed') AS classification_status,c.reason_code AS classification_reason,
 er.id AS extraction_run_id,coalesce(e.status,'not_processed') AS extraction_status,e.reason_code AS extraction_reason,
 s.interpretation_suspended_at, CASE WHEN discovery.is_listing THEN 'discovery_listing' ELSE 'interpretation_input' END AS evidence_role,
 (SELECT representative_version_id FROM interpretation_preflight WHERE document_version_id=v.id AND representative_version_id<>v.id) AS equivalent_to_version,
 (SELECT count(*) FROM interpretation_preflight WHERE representative_version_id=v.id AND document_version_id<>v.id) AS equivalent_versions
 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id JOIN registry_sources s ON s.id=d.source_id CROSS JOIN filter f
 CROSS JOIN LATERAL (SELECT (d.source_id='calcinaia-municipal'
 AND EXISTS(SELECT 1 FROM retained_resources dr JOIN registry_configurations dc ON dc.source_id=dr.source_id AND dc.revision=dr.configuration,
 jsonb_array_elements_text(dc.body->'sections') section(url)
 WHERE dr.version_id=v.id AND dr.role='original' AND dc.body#>>'{discovery,pagination_parameter}'='page'
 AND regexp_replace(d.official_url,'[?]page=[0-9]+$','')=section.url)
 AND NOT EXISTS(SELECT 1 FROM interpretation_triggers t WHERE t.document_version_id=v.id)
 AND NOT EXISTS(SELECT 1 FROM acquisition_targets target WHERE target.source_id=d.source_id AND target.url=d.official_url)) AS is_listing) discovery
 LEFT JOIN LATERAL (SELECT id FROM processing_runs WHERE document_version_id=v.id AND stage='classification' ORDER BY created_at DESC,id DESC LIMIT 1) cr ON true
 LEFT JOIN classification_results c ON c.run_id=cr.id
 LEFT JOIN LATERAL (SELECT id FROM processing_runs WHERE document_version_id=v.id AND stage='extraction' ORDER BY created_at DESC,id DESC LIMIT 1) er ON true
 LEFT JOIN extraction_results e ON e.run_id=er.id
 WHERE (f.source_id='' OR d.source_id=f.source_id) AND (f.version_id=0 OR v.id=f.version_id)
 AND NOT EXISTS(SELECT 1 FROM interpretation_archives ia WHERE ia.document_version_id=v.id)
 AND (f.version_id<>0 OR NOT discovery.is_listing)
 AND (f.version_id<>0 OR NOT EXISTS(SELECT 1 FROM (` + interpretation.WaitingEquivalentSQL + `) w WHERE w.document_version_id=v.id))
 AND (s.interpretation_suspended_at IS NOT NULL OR c.status IS NULL OR c.status='undetermined'
 OR (c.relevant AND (e.status IS DISTINCT FROM 'extracted' OR e.classification_run_id IS DISTINCT FROM c.run_id)))
 ORDER BY v.first_acquired_at DESC,v.id DESC`,
	"findings": `SELECT q.* FROM (
 SELECT 'interpretation_assessment' AS kind,m.source_id,i.evidence_document_version_id AS document_version_id,NULL::bigint AS run_id,i.state,i.reason AS detail,i.limitations AS evidence,i.recorded_at
 FROM domain_interpretation_events i JOIN domain_local_measures m ON m.id=i.local_measure_id WHERE i.state<>'supported'
 UNION ALL SELECT 'source_defect',source_id,NULL,NULL,CASE WHEN resumed_at IS NULL THEN 'suspended' ELSE 'resumed' END,'confirmed_interpretation_defect',jsonb_build_object('defect',defect,'recovery',recovery),suspended_at FROM registry_interpretation_suspensions
 UNION ALL SELECT 'evaluation',source_id,NULL,NULL,'accepted','source_acceptance',evidence,created_at FROM registry_events WHERE kind='acceptance'
 UNION ALL SELECT 'regression',source_id,NULL,NULL,CASE WHEN passed THEN 'passed' ELSE 'blocked' END,suite,report,recorded_at FROM registry_regressions
 UNION ALL SELECT 'linking',r.source_id,r.document_version_id,l.run_id,l.status,l.reason_code,
 jsonb_build_object('extraction_run_id',l.current_extraction_run_id,'measure_ordinal',l.current_measure_ordinal),l.created_at
 FROM linking_results l JOIN processing_runs r ON r.id=l.run_id WHERE l.status='unresolved'
 UNION ALL SELECT 'indeterminate_fields',r.source_id,r.document_version_id,e.run_id,'partial','indeterminate_fields',
 jsonb_build_object('measure_ordinal',e.ordinal,'fields',e.indeterminate_fields),r.created_at
 FROM extracted_measures e JOIN processing_runs r ON r.id=e.run_id WHERE e.indeterminate_fields<>'[]'::jsonb
 UNION ALL SELECT 'ocr',r.source_id,o.document_version_id,o.run_id,o.status,coalesce(o.error_code,'partial_unreadable'),jsonb_build_object('resource_url',o.resource_url,'page_count',o.page_count),o.created_at FROM ocr_resource_results o JOIN processing_runs r ON r.id=o.run_id WHERE o.status<>'complete'
 UNION ALL SELECT 'extraction',r.source_id,e.document_version_id,e.run_id,e.status,e.reason_code,'{}'::jsonb,e.created_at FROM extraction_results e JOIN processing_runs r ON r.id=e.run_id WHERE e.status='uninterpreted'
 UNION ALL SELECT 'classification',r.source_id,c.document_version_id,c.run_id,c.status,c.reason_code,'{}'::jsonb,c.created_at FROM classification_results c JOIN processing_runs r ON r.id=c.run_id WHERE c.status='undetermined'
 ) q CROSS JOIN filter f WHERE (f.source_id='' OR q.source_id=f.source_id) AND (f.version_id=0 OR q.document_version_id=f.version_id)
 ORDER BY q.recorded_at DESC,q.kind,q.source_id,q.document_version_id,q.run_id,q.evidence::text`,
	"dpc-comparisons": `SELECT c.id,c.regional_source_id,c.regional_version_id,c.dpc_source_id,c.dpc_version_id,c.scope,c.dimensions,c.actor,c.compared_at
 FROM diagnostic_dpc_comparisons c CROSS JOIN filter f
 WHERE (f.source_id='' OR c.regional_source_id=f.source_id OR c.dpc_source_id=f.source_id)
 AND (f.version_id=0 OR c.regional_version_id=f.version_id OR c.dpc_version_id=f.version_id)
 ORDER BY c.compared_at DESC,c.id DESC`,
	"usage": `SELECT r.id AS run_id,r.source_id,r.document_version_id,r.workload,r.stage,r.configuration_version_id,c.prompt_version_id,c.logic_version,m.provider,m.model,c.model_version_id,
 a.number AS attempt,a.queue_job_id,a.queue_attempt_number,a.started_at,a.finished_at,a.duration_ms,a.outcome,a.error_code,
 a.usage_status,a.input_tokens,a.output_tokens,a.cache_read_tokens,a.cache_write_tokens,a.other_units,a.cost_status,a.estimated_cost_microunits,
 a.price_version_id,p.currency,p.provenance_url AS pricing_provenance_url,p.observed_at AS price_observed_at,p.effective_from,p.effective_through,
 (SELECT jsonb_agg(jsonb_build_object('metric',metric,'unit_size',unit_size,'price_microunits',price_microunits) ORDER BY metric) FROM processing_price_rates WHERE price_version_id=p.id) AS price_rates
 ` + usageFrom + ` ORDER BY r.id DESC,a.number`,
}

func ValidJobState(state string) bool {
	switch state {
	case "", "queued", "running", "retry_wait", "failed", "succeeded":
		return true
	}
	return false
}

func filteredQuery(section string, f Filter) string {
	q := queries[section]
	if section == "sources" && f.Issue == "attention" {
		q = "SELECT * FROM (" + q + ") source_issues WHERE updating_state='delayed' OR last_error_code IS NOT NULL ORDER BY source_id"
	}
	return q
}

type Overview struct {
	ObservedAt       time.Time `json:"observed_at"`
	SourceIssues     int64     `json:"source_issues"`
	FailedJobs       int64     `json:"failed_jobs"`
	PendingDocuments int64     `json:"pending_documents"`
}

func (s *Store) Overview(ctx context.Context, at time.Time) (Overview, error) {
	result := Overview{ObservedAt: at.UTC()}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	prefix := `WITH filter AS (SELECT $1::text source_id,$2::bigint version_id,$3::bigint run_id,$4::timestamptz observed_at,$5::bigint job_id,$6::text state) `
	for _, item := range []struct {
		section string
		filter  Filter
		count   *int64
	}{{"sources", Filter{Issue: "attention"}, &result.SourceIssues}, {"jobs", Filter{State: "failed"}, &result.FailedJobs}, {"documents", Filter{}, &result.PendingDocuments}} {
		q := prefix + "SELECT count(*) FROM (" + filteredQuery(item.section, item.filter) + ") records"
		if err := tx.QueryRow(ctx, q, "", int64(0), int64(0), at.UTC(), int64(0), item.filter.State).Scan(item.count); err != nil {
			return result, err
		}
	}
	return result, tx.Commit(ctx)
}
