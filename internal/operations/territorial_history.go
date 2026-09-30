package operations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/domain"
	"slices"
	"time"
)

type TerritorialHistoryFilter struct {
	Source, Kind, After string
	From, Through       *time.Time
	Limit               int
}
type TerritorialHistoryEntry struct {
	ID, Kind, Source, Actor, Outcome string
	RecordedAt                       time.Time
	Revision                         int
	VersionID, JobID                 int64
	Evidence                         json.RawMessage
}
type TerritorialHistory struct {
	Entries         []TerritorialHistoryEntry
	Filter          TerritorialHistoryFilter
	Next            string
	ObservedAt      time.Time
	RetentionKnown  bool
	DeletedVersions int64
	Limitations     []string
}

const territorialHistorySQL = `WITH events AS(
 SELECT 'region-'||e.id id,'configuration' kind,''::text source_id,e.actor,e.recorded_at,e.revision,0::bigint version_id,0::bigint job_id,e.kind outcome,jsonb_build_object('configuration',v.configuration,'enabled',v.enabled) evidence,e.region_code region
 FROM territorial_region_events e JOIN territorial_region_versions v ON v.region_code=e.region_code AND v.revision=e.revision WHERE e.region_code=$1
 UNION ALL
 SELECT 'config-'||c.source_id||'-'||c.revision,'configuration',c.source_id,c.actor,c.created_at,c.revision,0,0,'configuration',c.body,NULL FROM registry_configurations c
 UNION ALL
 SELECT 'source-'||e.id,'source',e.source_id,e.actor,e.created_at,e.revision,0,0,e.kind,e.evidence,NULL FROM registry_events e
 UNION ALL
 SELECT 'document-'||v.id,'document',d.source_id,'',v.first_acquired_at,COALESCE((SELECT configuration FROM retained_acquisitions WHERE version_id=v.id ORDER BY acquired_at LIMIT 1),0),v.id,0,'acquired',jsonb_build_object('url',d.official_url,'metadata',v.metadata),NULL FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 UNION ALL
 SELECT 'check-'||c.id,'source',c.source_id,c.worker_id,c.finished_at,c.configuration,0,0,CASE WHEN c.complete THEN 'succeeded' ELSE COALESCE(c.error_code,'failed') END,jsonb_build_object('started_at',c.started_at,'reachable',c.reachable,'content_recognized',c.content_recognized),NULL FROM acquisition_checks c
 UNION ALL
 SELECT 'processing-'||r.id||'-'||a.number,'processing',COALESCE(r.source_id,d.source_id),COALESCE((SELECT worker_id FROM processing_attempts WHERE job_id=a.queue_job_id AND number=a.queue_attempt_number),''),COALESCE(a.finished_at,a.started_at),0,COALESCE(r.document_version_id,0),COALESCE(a.queue_job_id,0),a.outcome,jsonb_build_object('run_id',r.id,'stage',r.stage,'error_code',a.error_code,'attempt',a.number),NULL FROM processing_runs r JOIN processing_run_attempts a ON a.run_id=r.id LEFT JOIN retained_versions v ON v.id=r.document_version_id LEFT JOIN retained_documents d ON d.id=v.document_id
 UNION ALL
 SELECT 'job-'||j.id||'-'||a.number,'processing',COALESCE(d.source_id,j.payload->>'source_id'),a.worker_id,COALESCE(a.finished_at,a.started_at),0,COALESCE(v.id,0),j.id,a.outcome,jsonb_build_object('kind',j.kind,'attempt',a.number,'error_code',a.error_code),NULL
 FROM processing_jobs j JOIN processing_attempts a ON a.job_id=j.id
 LEFT JOIN extraction_results er ON er.run_id::text=j.payload->>'extraction_run_id'
 LEFT JOIN retained_acquisitions ac ON ac.id=j.payload->>'acquisition_id'
 LEFT JOIN retained_versions v ON v.id::text=COALESCE(j.payload->>'document_version_id',er.document_version_id::text,ac.version_id::text)
 LEFT JOIN retained_documents d ON d.id=v.document_id
 WHERE NOT EXISTS(SELECT 1 FROM processing_run_attempts pa WHERE pa.queue_job_id=j.id AND pa.queue_attempt_number=a.number)
 ), scoped AS(
 SELECT e.* FROM events e LEFT JOIN LATERAL(SELECT a.region_code,a.municipality_istat FROM territorial_source_associations a WHERE a.source_id=e.source_id AND a.recorded_at<=e.recorded_at ORDER BY a.recorded_at DESC,a.id DESC LIMIT 1)a ON e.source_id<>''
 WHERE (e.region=$1 OR (a.region_code=$1 AND ($2='' OR a.municipality_istat=$2))) AND ($3='' OR e.source_id=$3) AND ($4='' OR e.kind=$4) AND ($5::timestamptz IS NULL OR e.recorded_at>=$5) AND ($6::timestamptz IS NULL OR e.recorded_at<=$6) AND e.recorded_at<=$7)
 SELECT id,kind,COALESCE(source_id,''),actor,recorded_at,revision,version_id,job_id,outcome,evidence FROM scoped WHERE ($8::timestamptz IS NULL OR (recorded_at,id)<($8,$9)) ORDER BY recorded_at DESC,id DESC LIMIT $10`

func (s *Store) TerritorialHistory(ctx context.Context, region, istat string, f TerritorialHistoryFilter, at time.Time) (TerritorialHistory, error) {
	result := TerritorialHistory{Entries: []TerritorialHistoryEntry{}, Filter: f, ObservedAt: at, Limitations: []string{"Lo storico contiene le evidenze conservate; non certifica la completezza delle pubblicazioni."}}
	if f.Limit == 0 {
		f.Limit = 50
	}
	if f.Limit < 1 || f.Limit > 100 || !slices.Contains([]string{"", "configuration", "source", "document", "processing"}, f.Kind) || len(f.Source) > 200 || (f.From != nil && f.Through != nil && f.Through.Before(*f.From)) {
		return result, ErrInvalid
	}
	result.Filter = f
	if _, err := domain.New(s.pool).Region(ctx, region); err != nil {
		return result, err
	}
	if istat != "" {
		if _, err := s.Municipality(ctx, region, istat); err != nil {
			return result, err
		}
	}
	var cursorTime *time.Time
	cursorID := ""
	if f.After != "" {
		var c struct {
			At time.Time
			ID string
		}
		raw, err := base64.RawURLEncoding.DecodeString(f.After)
		if len(f.After) > 2048 || err != nil || json.Unmarshal(raw, &c) != nil || c.At.IsZero() || c.ID == "" {
			return result, ErrInvalid
		}
		cursorTime = &c.At
		cursorID = c.ID
	}
	rows, err := s.pool.Query(ctx, territorialHistorySQL, region, istat, f.Source, f.Kind, f.From, f.Through, at, cursorTime, cursorID, f.Limit+1)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var e TerritorialHistoryEntry
		if err = rows.Scan(&e.ID, &e.Kind, &e.Source, &e.Actor, &e.RecordedAt, &e.Revision, &e.VersionID, &e.JobID, &e.Outcome, &e.Evidence); err != nil {
			rows.Close()
			return result, err
		}
		result.Entries = append(result.Entries, e)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	if len(result.Entries) > f.Limit {
		result.Entries = result.Entries[:f.Limit]
		e := result.Entries[len(result.Entries)-1]
		raw, _ := json.Marshal(struct {
			At time.Time
			ID string
		}{e.RecordedAt, e.ID})
		result.Next = base64.RawURLEncoding.EncodeToString(raw)
	}
	if err = s.pool.QueryRow(ctx, `SELECT COALESCE(sum(deleted_versions),0) FROM retention_cleanup_runs`).Scan(&result.DeletedVersions); err == nil {
		result.RetentionKnown = true
		if result.DeletedVersions > 0 {
			result.Limitations = append(result.Limitations, "Sono state eliminate versioni per retention; il conteggio riguarda l'intero servizio.")
		}
	} else {
		result.Limitations = append(result.Limitations, "Informazioni sulla retention non disponibili.")
	}
	return result, nil
}
