package acquisition

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DecisionExclude = "exclude_stale_unavailable"
	DecisionResume  = "resume"
)

// Stable, private-detail-free disposition failures shared by administration
// and CLI.
var (
	ErrDispositionInvalid             = errors.New("invalid_disposition")
	ErrTargetNotFound                 = errors.New("target_not_found")
	ErrTargetChanged                  = errors.New("target_changed")
	ErrTargetRecovered                = errors.New("target_recovered")
	ErrTargetReferenced               = errors.New("target_referenced")
	ErrTargetAlreadyExcluded          = errors.New("target_already_excluded")
	ErrConfigurationInactive          = errors.New("configuration_inactive")
	ErrMeasureAcknowledgementRequired = errors.New("measure_acknowledgement_required")
	ErrNothingToResume                = errors.New("nothing_to_resume")
)

// DispositionErrorCode maps a disposition failure to its stable code.
func DispositionErrorCode(err error) string {
	for _, known := range []error{ErrDispositionInvalid, ErrTargetNotFound, ErrTargetChanged, ErrTargetRecovered, ErrTargetReferenced, ErrTargetAlreadyExcluded, ErrConfigurationInactive, ErrMeasureAcknowledgementRequired, ErrNothingToResume} {
		if errors.Is(err, known) {
			return known.Error()
		}
	}
	return "disposition_failed"
}

type TargetDisposition struct {
	ID                  int64           `json:"id"`
	Decision            string          `json:"decision"`
	Actor               string          `json:"actor"`
	Reason              string          `json:"reason"`
	Evidence            json.RawMessage `json:"evidence"`
	ObservedLastSeenAt  time.Time       `json:"observed_last_seen_at"`
	MeasureAcknowledged bool            `json:"measure_acknowledged"`
	RecordedAt          time.Time       `json:"recorded_at"`
}

// UnavailableTarget is an unrecovered 404/410 target of a source's active
// configuration, classified as on the latest check that traversed listings.
type UnavailableTarget struct {
	URL                 string              `json:"url"`
	Configuration       int                 `json:"configuration"`
	HTTPStatus          int                 `json:"http_status"`
	ConsecutiveFailures int                 `json:"consecutive_failures"`
	FirstFailedAt       time.Time           `json:"first_failed_at"`
	LastFailedAt        time.Time           `json:"last_failed_at"`
	NextAttemptAt       time.Time           `json:"next_attempt_at"`
	LastSeenAt          time.Time           `json:"last_seen_at"`
	ExplicitReference   bool                `json:"explicit_reference"`
	SupportsMeasure     bool                `json:"supports_measure"`
	Classification      string              `json:"classification"`
	Dispositions        []TargetDisposition `json:"dispositions"`
}

type DispositionRequest struct {
	SourceID            string
	Configuration       int
	URL                 string
	Decision            string
	ObservedLastSeenAt  time.Time
	Actor               string
	Reason              string
	Evidence            json.RawMessage
	AcknowledgeMeasures bool
}

type DispositionStore struct{ pool *pgxpool.Pool }

func NewDispositionStore(pool *pgxpool.Pool) *DispositionStore { return &DispositionStore{pool: pool} }

func (s *DispositionStore) ListUnavailable(ctx context.Context, sourceID string) ([]UnavailableTarget, error) {
	if strings.TrimSpace(sourceID) == "" {
		return nil, ErrDispositionInvalid
	}
	measure, err := measureSupportSQL(ctx, s.pool)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `WITH traversal AS (
   SELECT max(c.started_at) AS started_at FROM acquisition_checks c JOIN registry_sources r ON r.id=c.source_id AND r.active_revision=c.configuration
   WHERE c.source_id=$1 AND c.listing_count>0)
 SELECT t.url,t.configuration,u.http_status,u.consecutive_failures,u.first_failed_at,u.last_failed_at,u.next_attempt_at,t.last_seen_at,
 t.explicit_reference,`+measure+`,
 COALESCE(t.last_seen_at>=(SELECT started_at FROM traversal),true),
 COALESCE(latest.decision='exclude_stale_unavailable' AND t.last_seen_at<=latest.observed_last_seen_at,false)
 FROM acquisition_targets t
 JOIN registry_sources r ON r.id=t.source_id AND r.active_revision=t.configuration
 JOIN acquisition_unavailable_targets u ON u.source_id=t.source_id AND u.configuration=t.configuration AND u.url=t.url AND u.recovered_at IS NULL
 LEFT JOIN LATERAL (SELECT decision,observed_last_seen_at FROM acquisition_target_dispositions d
   WHERE d.source_id=t.source_id AND d.configuration=t.configuration AND d.url=t.url ORDER BY d.id DESC LIMIT 1) latest ON true
 WHERE t.source_id=$1 ORDER BY u.first_failed_at,t.url`, sourceID)
	if err != nil {
		return nil, err
	}
	var result []UnavailableTarget
	for rows.Next() {
		var item UnavailableTarget
		var listed, excluded bool
		if err = rows.Scan(&item.URL, &item.Configuration, &item.HTTPStatus, &item.ConsecutiveFailures, &item.FirstFailedAt, &item.LastFailedAt, &item.NextAttemptAt, &item.LastSeenAt, &item.ExplicitReference, &item.SupportsMeasure, &listed, &excluded); err != nil {
			rows.Close()
			return nil, err
		}
		switch {
		case excluded:
			item.Classification = "excluded"
		case listed || item.ExplicitReference || item.SupportsMeasure:
			item.Classification = "blocking"
		default:
			item.Classification = "withdrawn"
		}
		result = append(result, item)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	for i := range result {
		if result[i].Dispositions, err = s.dispositions(ctx, sourceID, result[i].Configuration, result[i].URL); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *DispositionStore) dispositions(ctx context.Context, sourceID string, configuration int, url string) ([]TargetDisposition, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,decision,actor,reason,evidence,observed_last_seen_at,measure_acknowledged,recorded_at
 FROM acquisition_target_dispositions WHERE source_id=$1 AND configuration=$2 AND url=$3 ORDER BY id`, sourceID, configuration, url)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []TargetDisposition{}
	for rows.Next() {
		var item TargetDisposition
		if err = rows.Scan(&item.ID, &item.Decision, &item.Actor, &item.Reason, &item.Evidence, &item.ObservedLastSeenAt, &item.MeasureAcknowledged, &item.RecordedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// RecordDisposition records one reviewed decision for an exact source,
// configuration and URL, provided the target is still in the reviewed state.
// It neither starts a check nor changes publication.
func (s *DispositionStore) RecordDisposition(ctx context.Context, request DispositionRequest) (TargetDisposition, error) {
	var evidence map[string]any
	if strings.TrimSpace(request.SourceID) == "" || request.Configuration < 1 || strings.TrimSpace(request.URL) == "" ||
		(request.Decision != DecisionExclude && request.Decision != DecisionResume) || request.ObservedLastSeenAt.IsZero() ||
		strings.TrimSpace(request.Actor) == "" || strings.TrimSpace(request.Reason) == "" ||
		json.Unmarshal(request.Evidence, &evidence) != nil || len(evidence) == 0 {
		return TargetDisposition{}, ErrDispositionInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return TargetDisposition{}, err
	}
	defer tx.Rollback(ctx)
	var active *int
	if err = tx.QueryRow(ctx, "SELECT active_revision FROM registry_sources WHERE id=$1", request.SourceID).Scan(&active); errors.Is(err, pgx.ErrNoRows) {
		return TargetDisposition{}, ErrTargetNotFound
	} else if err != nil {
		return TargetDisposition{}, err
	}
	if active == nil || *active != request.Configuration {
		return TargetDisposition{}, ErrConfigurationInactive
	}
	measure, err := measureSupportSQL(ctx, tx)
	if err != nil {
		return TargetDisposition{}, err
	}
	var lastSeen time.Time
	var referenced, supportsMeasure bool
	err = tx.QueryRow(ctx, `SELECT t.last_seen_at,t.explicit_reference,`+measure+` FROM acquisition_targets t
 WHERE t.source_id=$1 AND t.configuration=$2 AND t.url=$3 FOR UPDATE OF t`, request.SourceID, request.Configuration, request.URL).Scan(&lastSeen, &referenced, &supportsMeasure)
	if errors.Is(err, pgx.ErrNoRows) {
		return TargetDisposition{}, ErrTargetNotFound
	} else if err != nil {
		return TargetDisposition{}, err
	}
	var recovered *time.Time
	err = tx.QueryRow(ctx, `SELECT recovered_at FROM acquisition_unavailable_targets WHERE source_id=$1 AND configuration=$2 AND url=$3 FOR UPDATE`, request.SourceID, request.Configuration, request.URL).Scan(&recovered)
	if errors.Is(err, pgx.ErrNoRows) {
		return TargetDisposition{}, ErrTargetNotFound
	} else if err != nil {
		return TargetDisposition{}, err
	}
	if recovered != nil {
		return TargetDisposition{}, ErrTargetRecovered
	}
	if !lastSeen.Equal(request.ObservedLastSeenAt.Truncate(time.Microsecond)) {
		return TargetDisposition{}, ErrTargetChanged
	}
	var latest string
	var latestSeen time.Time
	err = tx.QueryRow(ctx, `SELECT decision,observed_last_seen_at FROM acquisition_target_dispositions
 WHERE source_id=$1 AND configuration=$2 AND url=$3 ORDER BY id DESC LIMIT 1`, request.SourceID, request.Configuration, request.URL).Scan(&latest, &latestSeen)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return TargetDisposition{}, err
	}
	excluded := latest == DecisionExclude && !lastSeen.After(latestSeen)
	acknowledged := false
	switch request.Decision {
	case DecisionExclude:
		if excluded {
			return TargetDisposition{}, ErrTargetAlreadyExcluded
		}
		if referenced {
			return TargetDisposition{}, ErrTargetReferenced
		}
		if supportsMeasure && !request.AcknowledgeMeasures {
			return TargetDisposition{}, ErrMeasureAcknowledgementRequired
		}
		acknowledged = supportsMeasure
	case DecisionResume:
		if !excluded {
			return TargetDisposition{}, ErrNothingToResume
		}
	}
	item := TargetDisposition{Decision: request.Decision, Actor: strings.TrimSpace(request.Actor), Reason: strings.TrimSpace(request.Reason), Evidence: request.Evidence, ObservedLastSeenAt: lastSeen, MeasureAcknowledged: acknowledged}
	err = tx.QueryRow(ctx, `INSERT INTO acquisition_target_dispositions(source_id,configuration,url,decision,actor,reason,evidence,observed_last_seen_at,measure_acknowledged)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id,recorded_at`, request.SourceID, request.Configuration, request.URL, item.Decision, item.Actor, item.Reason, item.Evidence, lastSeen, acknowledged).Scan(&item.ID, &item.RecordedAt)
	if err != nil {
		return TargetDisposition{}, err
	}
	return item, tx.Commit(ctx)
}
