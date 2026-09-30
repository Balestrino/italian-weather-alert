package acquisition

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/territory"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoDueCheck = errors.New("no source check is due")
	ErrStaleCheck = errors.New("source check lease is stale")
)

type ScheduleStore struct{ pool *pgxpool.Pool }

func NewScheduleStore(pool *pgxpool.Pool) *ScheduleStore { return &ScheduleStore{pool: pool} }

type CheckClaim struct {
	SourceID      string
	Configuration int
	WorkerID      string
	Token         string
	StartedAt     time.Time
	LeaseExpires  time.Time
}

type SourceStatus struct {
	SourceID            string
	Configuration       int
	UpdatingState       string
	NextCheckAt         time.Time
	ConsecutiveFailures int
	LastStartedAt       *time.Time
	LastReachableAt     *time.Time
	LastContentAt       *time.Time
	LastCompleteAt      *time.Time
	FirstErrorAt        *time.Time
	LastErrorAt         *time.Time
	LastErrorCode       string
	RetryAfter          *time.Time
	LastCheckState      string
	PublicationState    string
	LastPublicationAt   *time.Time
}

type StoredCheck struct {
	ID int64
	CheckOutcome
	WorkerID         string
	CheckState       string
	PublicationState string
}

func (s *ScheduleStore) SyncEnabled(ctx context.Context, now time.Time) error {
	if s == nil || s.pool == nil || now.IsZero() {
		return ErrInvalidConfiguration
	}
	gate, err := territory.Predicate(ctx, s.pool, "s.id", false)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO acquisition_source_status(source_id,configuration,check_seconds,delay_seconds,backoff_base_seconds,backoff_max_seconds,next_check_at,expected_anchor,expected_interval_seconds,expected_tolerance_seconds,publication_state)
 SELECT s.id,s.active_revision,
        COALESCE(s.check_seconds_override,(c.body->>'check_seconds')::integer,600),
        COALESCE(s.delay_seconds_override,(c.body->>'delay_seconds')::integer,1800),
        COALESCE((c.body->>'backoff_base_seconds')::integer,60),
        COALESCE((c.body->>'backoff_max_seconds')::integer,3600),$1,
        NULLIF(c.body#>>'{expected_publication,anchor}','')::timestamptz,
        NULLIF(c.body#>>'{expected_publication,interval_seconds}','')::integer,
        NULLIF(c.body#>>'{expected_publication,tolerance_seconds}','')::integer,
        CASE WHEN c.body->'expected_publication' IS NULL THEN 'not_expected' ELSE 'awaiting' END
 FROM registry_sources s JOIN registry_configurations c ON c.source_id=s.id AND c.revision=s.active_revision
 WHERE s.collection_enabled `+gate+`
 ON CONFLICT(source_id) DO UPDATE SET
   configuration=EXCLUDED.configuration,check_seconds=EXCLUDED.check_seconds,delay_seconds=EXCLUDED.delay_seconds,
   backoff_base_seconds=EXCLUDED.backoff_base_seconds,backoff_max_seconds=EXCLUDED.backoff_max_seconds,
   expected_anchor=EXCLUDED.expected_anchor,expected_interval_seconds=EXCLUDED.expected_interval_seconds,expected_tolerance_seconds=EXCLUDED.expected_tolerance_seconds,
   next_check_at=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN EXCLUDED.next_check_at ELSE acquisition_source_status.next_check_at END,
   consecutive_failures=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN 0 ELSE acquisition_source_status.consecutive_failures END,
   last_started_at=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.last_started_at END,
   last_reachable_at=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.last_reachable_at END,
   last_content_at=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.last_content_at END,
   last_complete_at=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.last_complete_at END,
   first_error_at=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.first_error_at END,
   last_error_at=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.last_error_at END,
   last_error_code=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.last_error_code END,
   retry_after=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.retry_after END,
   last_publication_at=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.last_publication_at END,
   last_check_state=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN 'never' ELSE acquisition_source_status.last_check_state END,
   publication_state=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN CASE WHEN EXCLUDED.expected_anchor IS NULL THEN 'not_expected' ELSE 'awaiting' END ELSE acquisition_source_status.publication_state END,
   claimed_by=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.claimed_by END,
   claim_token=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.claim_token END,
   lease_expires_at=CASE WHEN acquisition_source_status.configuration<>EXCLUDED.configuration THEN NULL ELSE acquisition_source_status.lease_expires_at END`, now.UTC())
	return err
}

func (s *ScheduleStore) ClaimDue(ctx context.Context, workerID string, now time.Time, lease time.Duration) (CheckClaim, error) {
	if s == nil || s.pool == nil || !validWorker(workerID) || now.IsZero() || lease < 3*time.Second || lease > time.Hour {
		return CheckClaim{}, ErrInvalidConfiguration
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return CheckClaim{}, err
	}
	if _, err := s.RecoverExpired(ctx, now); err != nil {
		return CheckClaim{}, err
	}
	claim := CheckClaim{WorkerID: workerID, Token: hex.EncodeToString(tokenBytes), StartedAt: now.UTC(), LeaseExpires: now.Add(lease).UTC()}
	gate, err := territory.Predicate(ctx, s.pool, "r.id", false)
	if err != nil {
		return CheckClaim{}, err
	}
	err = s.pool.QueryRow(ctx, `WITH candidate AS (
 SELECT a.source_id FROM acquisition_source_status a JOIN registry_sources r ON r.id=a.source_id
 WHERE r.collection_enabled `+gate+` AND r.active_revision=a.configuration AND a.next_check_at<=$1
   AND a.claim_token IS NULL
 ORDER BY a.next_check_at,a.source_id FOR UPDATE OF a SKIP LOCKED LIMIT 1)
 UPDATE acquisition_source_status a SET claimed_by=$2,claim_token=$3,lease_expires_at=$4,last_started_at=$1
 FROM candidate c WHERE a.source_id=c.source_id
 RETURNING a.source_id,a.configuration`, claim.StartedAt, workerID, claim.Token, claim.LeaseExpires).Scan(&claim.SourceID, &claim.Configuration)
	if errors.Is(err, pgx.ErrNoRows) {
		return CheckClaim{}, ErrNoDueCheck
	}
	return claim, err
}

func (s *ScheduleStore) RecoverExpired(ctx context.Context, now time.Time) (int64, error) {
	if s == nil || s.pool == nil || now.IsZero() {
		return 0, ErrInvalidConfiguration
	}
	tag, err := s.pool.Exec(ctx, `WITH expired AS (
 UPDATE acquisition_source_status SET
   next_check_at=$1 + (LEAST(backoff_max_seconds::numeric,backoff_base_seconds::numeric*power(2,LEAST(consecutive_failures,30)))::double precision * interval '1 second'),
   consecutive_failures=consecutive_failures+1,
   first_error_at=COALESCE(first_error_at,$1),last_error_at=$1,last_error_code='check_lease_expired',retry_after=NULL,last_check_state='collection_failure',
   claimed_by=NULL,claim_token=NULL,lease_expires_at=NULL
 WHERE claim_token IS NOT NULL AND lease_expires_at<=$1
 RETURNING source_id,configuration,claimed_by,last_started_at)
 INSERT INTO acquisition_checks(source_id,configuration,worker_id,started_at,finished_at,reachable,content_recognized,complete,error_code,listing_count,document_count,check_state,publication_state)
 SELECT e.source_id,e.configuration,COALESCE(e.claimed_by,'expired-worker'),COALESCE(e.last_started_at,$1),$1,false,false,false,'check_lease_expired',0,0,'collection_failure',s.publication_state
 FROM expired e JOIN acquisition_source_status s ON s.source_id=e.source_id`, now.UTC())
	return tag.RowsAffected(), err
}

func (s *ScheduleStore) Heartbeat(ctx context.Context, claim CheckClaim, now time.Time, lease time.Duration) error {
	if now.IsZero() || lease < 3*time.Second || lease > time.Hour {
		return ErrInvalidConfiguration
	}
	expires := now.Add(lease).UTC()
	tag, err := s.pool.Exec(ctx, `UPDATE acquisition_source_status SET lease_expires_at=$4
 WHERE source_id=$1 AND configuration=$2 AND claim_token=$3 AND lease_expires_at>$5`, claim.SourceID, claim.Configuration, claim.Token, expires, now.UTC())
	if err == nil && tag.RowsAffected() == 0 {
		err = ErrStaleCheck
	}
	return err
}

func (s *ScheduleStore) Finish(ctx context.Context, claim CheckClaim, outcome CheckOutcome) error {
	if outcome.SourceID != claim.SourceID || outcome.Configuration != claim.Configuration || !outcome.StartedAt.Equal(claim.StartedAt) || outcome.StartedAt.IsZero() || outcome.FinishedAt.Before(outcome.StartedAt) || outcome.NewDocuments < 0 || (outcome.Complete && (!outcome.Reachable || !outcome.ContentRecognized || outcome.ErrorCode != "" || outcome.RetryAfter != nil)) || (!outcome.Complete && strings.TrimSpace(outcome.ErrorCode) == "") {
		return ErrInvalidConfiguration
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var failures, checkSeconds, baseSeconds, maxSeconds int
	var expectedAnchor, lastPublication *time.Time
	var expectedInterval, expectedTolerance *int
	var publicationState string
	err = tx.QueryRow(ctx, `SELECT consecutive_failures,check_seconds,backoff_base_seconds,backoff_max_seconds,
 expected_anchor,expected_interval_seconds,expected_tolerance_seconds,last_publication_at,publication_state
 FROM acquisition_source_status WHERE source_id=$1 AND configuration=$2 AND claim_token=$3 AND lease_expires_at>$4 FOR UPDATE`, claim.SourceID, claim.Configuration, claim.Token, outcome.FinishedAt.UTC()).Scan(&failures, &checkSeconds, &baseSeconds, &maxSeconds, &expectedAnchor, &expectedInterval, &expectedTolerance, &lastPublication, &publicationState)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrStaleCheck
	}
	if err != nil {
		return err
	}
	checkState := "collection_failure"
	if outcome.Complete {
		checkState = "complete_unchanged"
		if outcome.NewDocuments > 0 || outcome.Revisions > 0 {
			checkState = "complete_changed"
		}
		if outcome.PublicationObservedAt != nil && (lastPublication == nil || outcome.PublicationObservedAt.After(*lastPublication)) {
			observed := outcome.PublicationObservedAt.UTC()
			lastPublication = &observed
		}
		publicationState = publicationStateAt(expectedAnchor, expectedInterval, expectedTolerance, lastPublication, outcome.FinishedAt)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO acquisition_checks(source_id,configuration,worker_id,started_at,finished_at,reachable,content_recognized,complete,error_code,retry_after,listing_count,document_count,revision_count,new_document_count,publication_observed_at,check_state,publication_state)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`, outcome.SourceID, outcome.Configuration, claim.WorkerID, outcome.StartedAt.UTC(), outcome.FinishedAt.UTC(), outcome.Reachable, outcome.ContentRecognized, outcome.Complete, nullable(outcome.ErrorCode), outcome.RetryAfter, outcome.Listings, outcome.Documents, outcome.Revisions, outcome.NewDocuments, outcome.PublicationObservedAt, checkState, publicationState); err != nil {
		return err
	}
	if outcome.Complete {
		_, err = tx.Exec(ctx, `UPDATE acquisition_source_status SET next_check_at=$2,consecutive_failures=0,
 last_reachable_at=$3,last_content_at=$3,last_complete_at=$3,first_error_at=NULL,last_error_at=NULL,last_error_code=NULL,retry_after=NULL,
 last_check_state=$4,publication_state=$5,last_publication_at=$6,
 claimed_by=NULL,claim_token=NULL,lease_expires_at=NULL WHERE source_id=$1`, claim.SourceID, outcome.FinishedAt.Add(time.Duration(checkSeconds)*time.Second).UTC(), outcome.FinishedAt.UTC(), checkState, publicationState, lastPublication)
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	failures++
	delay := exponentialBackoff(baseSeconds, maxSeconds, failures)
	next := outcome.FinishedAt.Add(delay).UTC()
	if outcome.RetryAfter != nil && outcome.RetryAfter.After(next) {
		next = outcome.RetryAfter.UTC()
	}
	_, err = tx.Exec(ctx, `UPDATE acquisition_source_status SET next_check_at=$2,consecutive_failures=$3,
 last_reachable_at=CASE WHEN $4 THEN $5 ELSE last_reachable_at END,
 last_content_at=CASE WHEN $6 THEN $5 ELSE last_content_at END,
 first_error_at=COALESCE(first_error_at,$5),last_error_at=$5,last_error_code=$7,retry_after=$8,
 last_check_state='collection_failure',claimed_by=NULL,claim_token=NULL,lease_expires_at=NULL WHERE source_id=$1`, claim.SourceID, next, failures, outcome.Reachable, outcome.FinishedAt.UTC(), outcome.ContentRecognized, outcome.ErrorCode, outcome.RetryAfter)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *ScheduleStore) Status(ctx context.Context, sourceID string, now time.Time) (SourceStatus, error) {
	var status SourceStatus
	if now.IsZero() {
		return status, ErrInvalidConfiguration
	}
	var delaySeconds int
	err := s.pool.QueryRow(ctx, `SELECT source_id,configuration,next_check_at,consecutive_failures,last_started_at,last_reachable_at,last_content_at,last_complete_at,first_error_at,last_error_at,COALESCE(last_error_code,''),retry_after,delay_seconds,last_check_state,publication_state,last_publication_at
 FROM acquisition_source_status WHERE source_id=$1`, sourceID).Scan(&status.SourceID, &status.Configuration, &status.NextCheckAt, &status.ConsecutiveFailures, &status.LastStartedAt, &status.LastReachableAt, &status.LastContentAt, &status.LastCompleteAt, &status.FirstErrorAt, &status.LastErrorAt, &status.LastErrorCode, &status.RetryAfter, &delaySeconds, &status.LastCheckState, &status.PublicationState, &status.LastPublicationAt)
	if err != nil {
		return status, err
	}
	status.UpdatingState = "current"
	if status.LastCompleteAt == nil {
		status.UpdatingState = "not_yet_verified"
	} else if !now.Before(status.LastCompleteAt.Add(time.Duration(delaySeconds) * time.Second)) {
		status.UpdatingState = "delayed"
	}
	return status, nil
}

func (s *ScheduleStore) Checks(ctx context.Context, sourceID string) ([]StoredCheck, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,configuration,worker_id,started_at,finished_at,reachable,content_recognized,complete,COALESCE(error_code,''),retry_after,listing_count,document_count,revision_count,new_document_count,publication_observed_at,check_state,publication_state
 FROM acquisition_checks WHERE source_id=$1 ORDER BY id`, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []StoredCheck
	for rows.Next() {
		item := StoredCheck{CheckOutcome: CheckOutcome{SourceID: sourceID}}
		if err = rows.Scan(&item.ID, &item.Configuration, &item.WorkerID, &item.StartedAt, &item.FinishedAt, &item.Reachable, &item.ContentRecognized, &item.Complete, &item.ErrorCode, &item.RetryAfter, &item.Listings, &item.Documents, &item.Revisions, &item.NewDocuments, &item.PublicationObservedAt, &item.CheckState, &item.PublicationState); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func publicationStateAt(anchor *time.Time, intervalSeconds, toleranceSeconds *int, lastPublication *time.Time, checkedAt time.Time) string {
	if anchor == nil || intervalSeconds == nil || toleranceSeconds == nil || *intervalSeconds < 1 {
		return "not_expected"
	}
	if checkedAt.Before(*anchor) {
		return "awaiting"
	}
	interval := time.Duration(*intervalSeconds) * time.Second
	due := anchor.Add(time.Duration(checkedAt.Sub(*anchor)/interval) * interval)
	if lastPublication != nil && !lastPublication.Before(due) {
		return "observed"
	}
	if !checkedAt.After(due.Add(time.Duration(*toleranceSeconds) * time.Second)) {
		return "awaiting"
	}
	return "missing"
}

func exponentialBackoff(baseSeconds, maxSeconds, failures int) time.Duration {
	delay := int64(baseSeconds)
	for i := 1; i < failures && delay < int64(maxSeconds); i++ {
		delay *= 2
		if delay > int64(maxSeconds) {
			delay = int64(maxSeconds)
		}
	}
	return time.Duration(delay) * time.Second
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func validWorker(value string) bool {
	if value == "" || len(value) > 200 {
		return false
	}
	for _, r := range value {
		if !(r == '-' || r == '_' || r == ':' || r == '.' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return false
		}
	}
	return true
}
