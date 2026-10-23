package acts

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalid = errors.New("invalid referenced act record")

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

type StoredReference struct {
	ID int64
	Reference
}

// RecordReferences appends a version's detected references once per
// detection version and returns the stored ones in ordinal order.
func (s *Store) RecordReferences(ctx context.Context, versionID int64, refs []Reference, at time.Time) ([]StoredReference, error) {
	if versionID < 1 || at.IsZero() || len(refs) > MaxReferences {
		return nil, ErrInvalid
	}
	for _, r := range refs {
		var number, year, page any
		if r.Number > 0 {
			number = r.Number
		}
		if r.Year > 0 {
			year = r.Year
		}
		if r.Page > 0 {
			page = r.Page
		}
		if _, err := s.pool.Exec(ctx, `INSERT INTO act_references(document_version_id,ordinal,kind,act_number,act_year,act_date,passage,resource_url,page_number,detection_version,detected_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT(document_version_id,detection_version,ordinal) DO NOTHING`,
			versionID, r.Ordinal, r.Kind, number, year, r.Date, r.Passage, r.ResourceURL, page, DetectionVersion, at.UTC()); err != nil {
			return nil, err
		}
	}
	return s.References(ctx, versionID)
}

func (s *Store) References(ctx context.Context, versionID int64) ([]StoredReference, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,ordinal,kind,COALESCE(act_number,0),COALESCE(act_year,0),act_date,passage,resource_url,COALESCE(page_number,0)
 FROM act_references WHERE document_version_id=$1 AND detection_version=$2 ORDER BY ordinal`, versionID, DetectionVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []StoredReference
	for rows.Next() {
		var r StoredReference
		if err = rows.Scan(&r.ID, &r.Ordinal, &r.Kind, &r.Number, &r.Year, &r.Date, &r.Passage, &r.ResourceURL, &r.Page); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

type StoredResolution struct {
	ID, ReferenceID  int64
	Outcome          string
	AlboSourceID     string
	AlboRecordID     string
	AttachedResource string
	Candidates       []Candidate
	Reason           string
	RecordedAt       time.Time
}

// RecordResolution appends one outcome; a later outcome supersedes it.
func (s *Store) RecordResolution(ctx context.Context, referenceID int64, alboSource string, r Resolution, at time.Time) (int64, error) {
	if referenceID < 1 || at.IsZero() {
		return 0, ErrInvalid
	}
	candidates := r.Candidates
	if candidates == nil {
		candidates = []Candidate{}
	}
	raw, err := json.Marshal(candidates)
	if err != nil {
		return 0, err
	}
	var source, record, attached any
	if r.AlboRecordID != "" {
		source, record = alboSource, r.AlboRecordID
	}
	if r.AttachedResource != "" {
		attached = r.AttachedResource
	}
	var id int64
	err = s.pool.QueryRow(ctx, `INSERT INTO act_resolutions(reference_id,outcome,albo_source_id,albo_record_id,attached_resource_url,candidates,reason,recorded_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, referenceID, r.Outcome, source, record, attached, raw, r.Reason, at.UTC()).Scan(&id)
	return id, err
}

// LatestResolution returns the reference's latest outcome recorded at or before at.
func (s *Store) LatestResolution(ctx context.Context, referenceID int64, at time.Time) (StoredResolution, bool, error) {
	var r StoredResolution
	var raw []byte
	err := s.pool.QueryRow(ctx, `SELECT id,reference_id,outcome,COALESCE(albo_source_id,''),COALESCE(albo_record_id,''),COALESCE(attached_resource_url,''),candidates,reason,recorded_at
 FROM act_resolutions WHERE reference_id=$1 AND recorded_at<=$2 ORDER BY recorded_at DESC,id DESC LIMIT 1`, referenceID, at.UTC()).
		Scan(&r.ID, &r.ReferenceID, &r.Outcome, &r.AlboSourceID, &r.AlboRecordID, &r.AttachedResource, &raw, &r.Reason, &r.RecordedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, false, nil
	}
	if err != nil {
		return r, false, err
	}
	return r, true, json.Unmarshal(raw, &r.Candidates)
}

// ReferenceView is the read-only operator view of one notice reference.
type ReferenceView struct {
	DocumentID     int64               `json:"document_id"`
	VersionID      int64               `json:"version_id"`
	NoticeURL      string              `json:"notice_url"`
	Ordinal        int                 `json:"ordinal"`
	Kind           string              `json:"kind"`
	Number         int                 `json:"number,omitempty"`
	Year           int                 `json:"year,omitempty"`
	Passage        string              `json:"passage"`
	DetectedAt     time.Time           `json:"detected_at"`
	Outcome        string              `json:"outcome"`
	Reason         string              `json:"reason,omitempty"`
	ResolvedAt     *time.Time          `json:"resolved_at"`
	Candidates     []Candidate         `json:"candidates"`
	ActURL         string              `json:"act_url,omitempty"`
	ActVersionID   *int64              `json:"act_document_version_id"`
	Clarifications []ClarificationView `json:"clarifications"`
}

type ClarificationView struct {
	MeasureID  string    `json:"local_measure_id,omitempty"`
	Outcome    string    `json:"outcome"`
	Reason     string    `json:"reason,omitempty"`
	RecordedAt time.Time `json:"recorded_at"`
}

// SourceReferences lists the latest references of a source's notices with
// their latest outcome, acquired act and clarifications, newest first.
func (s *Store) SourceReferences(ctx context.Context, source string, limit int) ([]ReferenceView, error) {
	if source == "" || limit < 1 || limit > 500 {
		return nil, ErrInvalid
	}
	rows, err := s.pool.Query(ctx, `SELECT d.id,v.id,d.official_url,f.ordinal,f.kind,COALESCE(f.act_number,0),COALESCE(f.act_year,0),f.passage,f.detected_at,
 COALESCE(r.outcome,'pending'),COALESCE(r.reason,''),r.recorded_at,COALESCE(r.candidates,'[]'::jsonb),COALESCE(a.detail_url,r.attached_resource_url,''),
 COALESCE(a.document_version_id,CASE WHEN r.outcome='attached' THEN v.id END),r.id
 FROM act_references f JOIN retained_versions v ON v.id=f.document_version_id JOIN retained_documents d ON d.id=v.document_id
 LEFT JOIN LATERAL (SELECT * FROM act_resolutions x WHERE x.reference_id=f.id ORDER BY x.recorded_at DESC,x.id DESC LIMIT 1) r ON true
 LEFT JOIN act_albo_records a ON a.source_id=r.albo_source_id AND a.record_id=r.albo_record_id
 WHERE d.source_id=$1 ORDER BY f.detected_at DESC,f.id DESC LIMIT $2`, source, limit)
	if err != nil {
		return nil, err
	}
	var result []ReferenceView
	var resolutions []*int64
	for rows.Next() {
		var v ReferenceView
		var raw []byte
		var resolution *int64
		if err = rows.Scan(&v.DocumentID, &v.VersionID, &v.NoticeURL, &v.Ordinal, &v.Kind, &v.Number, &v.Year, &v.Passage, &v.DetectedAt, &v.Outcome, &v.Reason, &v.ResolvedAt, &raw, &v.ActURL, &v.ActVersionID, &resolution); err != nil {
			rows.Close()
			return nil, err
		}
		if err = json.Unmarshal(raw, &v.Candidates); err != nil {
			rows.Close()
			return nil, err
		}
		v.Clarifications = []ClarificationView{}
		result, resolutions = append(result, v), append(resolutions, resolution)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	for i, resolution := range resolutions {
		if resolution == nil {
			continue
		}
		clarifications, err := s.pool.Query(ctx, `SELECT COALESCE(local_measure_id,''),outcome,reason,recorded_at FROM act_clarifications WHERE resolution_id=$1 ORDER BY recorded_at,id`, *resolution)
		if err != nil {
			return nil, err
		}
		for clarifications.Next() {
			var c ClarificationView
			if err = clarifications.Scan(&c.MeasureID, &c.Outcome, &c.Reason, &c.RecordedAt); err != nil {
				clarifications.Close()
				return nil, err
			}
			result[i].Clarifications = append(result[i].Clarifications, c)
		}
		clarifications.Close()
		if err = clarifications.Err(); err != nil {
			return nil, err
		}
	}
	if result == nil {
		result = []ReferenceView{}
	}
	return result, nil
}
