package notifications

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalid = errors.New("invalid notification operation")

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// RecordBackup is the internal completion hook for backup runners. Replaying an
// identical run is harmless; a contradictory result cannot rewrite its history.
// Only identifiers and outcomes belong here, never destinations or credentials.
func (s *Store) RecordBackup(ctx context.Context, scope, runID string, succeeded bool, at time.Time) error {
	if !identifier(scope) || !identifier(runID) || at.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(730025)`); err != nil {
		return err
	}
	var id int64
	err = tx.QueryRow(ctx, `INSERT INTO notification_backup_results(scope,run_id,succeeded,finished_at)
        VALUES($1,$2,$3,$4) ON CONFLICT(scope,run_id) DO NOTHING RETURNING id`, scope, runID, succeeded, at.UTC()).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		var matched bool
		if err = tx.QueryRow(ctx, `SELECT succeeded=$3 AND finished_at=$4 FROM notification_backup_results WHERE scope=$1 AND run_id=$2`, scope, runID, succeeded, at.UTC()).Scan(&matched); err != nil {
			return err
		}
		if !matched {
			return ErrInvalid
		}
		return tx.Commit(ctx)
	}
	if err != nil {
		return err
	}
	var latest int64
	if err = tx.QueryRow(ctx, `SELECT id FROM notification_backup_results WHERE scope=$1 ORDER BY finished_at DESC,id DESC LIMIT 1`, scope).Scan(&latest); err != nil {
		return err
	}
	if latest == id {
		var recovery *time.Time
		if succeeded {
			recovery = &at
		}
		if err = s.observe(ctx, tx, observation{"backup_failed", scope, !succeeded, recovery}, at.UTC(), 0); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func identifier(s string) bool {
	if len(s) == 0 || len(s) > 200 {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("-_.:", r)) {
			return false
		}
	}
	return true
}

type observation struct {
	category, scope string
	active          bool
	recovery        *time.Time
}

// Reconcile takes one consistent snapshot. Missing/suspended sources and failed
// database reads never imply recovery. A relaunched job remains an incident until
// it succeeds, even while its fresh retry budget is being consumed.
func (s *Store) Reconcile(ctx context.Context, now time.Time, reminder time.Duration) error {
	if now.IsZero() || reminder < 0 {
		return ErrInvalid
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Serialize monitors before the first data snapshot; try-lock avoids waiting
	// on an older snapshot when several worker processes poll concurrently.
	var locked bool
	if err = tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(730025)`).Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return nil
	}
	_, err = tx.Exec(ctx, `INSERT INTO notification_source_watches(source_id,configuration,first_seen_at)
 SELECT s.id,s.active_revision,$1 FROM registry_sources s WHERE s.collection_enabled
 ON CONFLICT(source_id) DO UPDATE SET configuration=EXCLUDED.configuration,
 first_seen_at=CASE WHEN notification_source_watches.configuration<>EXCLUDED.configuration THEN EXCLUDED.first_seen_at ELSE notification_source_watches.first_seen_at END`, now.UTC())
	if err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT 'source_delay',r.id,
 $1::timestamptz >= COALESCE(a.last_complete_at,
 (SELECT min(started_at) FROM acquisition_checks c WHERE c.source_id=r.id AND c.configuration=r.active_revision),
 a.last_started_at,w.first_seen_at) + COALESCE(r.delay_seconds_override,(cfg.body->>'delay_seconds')::integer,1800)*interval '1 second', a.last_complete_at
 FROM registry_sources r JOIN registry_configurations cfg ON cfg.source_id=r.id AND cfg.revision=r.active_revision
 JOIN notification_source_watches w ON w.source_id=r.id
 LEFT JOIN acquisition_source_status a ON a.source_id=r.id AND a.configuration=r.active_revision
 WHERE r.collection_enabled
 UNION ALL
 SELECT 'processing_failed',jsonb_build_array(queue,kind)::text,
 bool_or(archived_at IS NULL AND state<>'succeeded' AND (state='failed' OR EXISTS(SELECT 1 FROM processing_job_relaunches l WHERE l.job_id=j.id))),
 max(completed_at) FILTER (WHERE state='succeeded')
 FROM processing_jobs j GROUP BY queue,kind
 UNION ALL
 SELECT 'backup_failed',scope,NOT succeeded,CASE WHEN succeeded THEN finished_at END
 FROM (SELECT DISTINCT ON(scope) scope,succeeded,finished_at FROM notification_backup_results WHERE finished_at<=$1 ORDER BY scope,finished_at DESC,id DESC) b`, now.UTC())
	if err != nil {
		return err
	}
	var observations []observation
	for rows.Next() {
		var o observation
		if err = rows.Scan(&o.category, &o.scope, &o.active, &o.recovery); err != nil {
			rows.Close()
			return err
		}
		observations = append(observations, o)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return err
	}
	for _, o := range observations {
		if err = s.observe(ctx, tx, o, now.UTC(), reminder); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) observe(ctx context.Context, tx pgx.Tx, o observation, now time.Time, reminder time.Duration) error {
	var id int64
	var opened time.Time
	var last *time.Time
	err := tx.QueryRow(ctx, `SELECT id,opened_at,last_notified_at FROM notification_incidents WHERE category=$1 AND scope=$2 AND recovered_at IS NULL FOR UPDATE`, o.category, o.scope).Scan(&id, &opened, &last)
	if errors.Is(err, pgx.ErrNoRows) {
		if !o.active {
			return nil
		}
		if err = tx.QueryRow(ctx, `INSERT INTO notification_incidents(category,scope,opened_at,last_observed_at) VALUES($1,$2,$3,$3) RETURNING id`, o.category, o.scope, now).Scan(&id); err != nil {
			return err
		}
		return enqueue(ctx, tx, id, "opened", now)
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE notification_incidents SET last_observed_at=$2 WHERE id=$1`, id, now); err != nil {
		return err
	}
	if !o.active {
		if o.recovery == nil || o.recovery.Before(opened) {
			return nil
		}
		if _, err = tx.Exec(ctx, `UPDATE notification_incidents SET recovered_at=$2 WHERE id=$1`, id, *o.recovery); err != nil {
			return err
		}
		return enqueue(ctx, tx, id, "recovered", now)
	}
	if reminder == 0 || last == nil || now.Before(last.Add(reminder)) {
		return nil
	}
	var pending bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM notification_messages WHERE incident_id=$1 AND sent_at IS NULL)`, id).Scan(&pending); err != nil {
		return err
	}
	if pending {
		return nil
	}
	return enqueue(ctx, tx, id, "reminder", now)
}

func enqueue(ctx context.Context, tx pgx.Tx, id int64, kind string, now time.Time) error {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `INSERT INTO notification_messages(message_id,incident_id,kind,created_at,next_attempt_at) VALUES($1,$2,$3,$4,$4)`, hex.EncodeToString(b)+"@iwa.invalid", id, kind, now)
	return err
}

type Message struct {
	ID, Category, Scope, Kind string
	IncidentID                int64
	CreatedAt, OpenedAt       time.Time
	RecoveredAt               *time.Time
}
type Sender interface {
	Send(context.Context, Message) error
}

// DeliverOne holds a row lock only for the bounded SMTP attempt. Earlier pending
// messages for the same incident must finish first. SMTP cannot guarantee exactly
// once delivery after a crash between server acceptance and the database commit;
// retries retain the same Message-ID to make that ambiguity identifiable.
func (s *Store) DeliverOne(ctx context.Context, sender Sender, now time.Time) (bool, error) {
	if sender == nil || now.IsZero() {
		return false, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var id int64
	var attempts int
	var m Message
	err = tx.QueryRow(ctx, `SELECT m.id,m.message_id,m.kind,m.created_at,m.attempts,i.id,i.category,i.scope,i.opened_at,i.recovered_at
 FROM notification_messages m JOIN notification_incidents i ON i.id=m.incident_id
 WHERE m.sent_at IS NULL AND m.next_attempt_at<=$1 AND NOT EXISTS(
 SELECT 1 FROM notification_messages earlier WHERE earlier.incident_id=m.incident_id AND earlier.id<m.id AND earlier.sent_at IS NULL)
 ORDER BY m.id FOR UPDATE OF m SKIP LOCKED LIMIT 1`, now.UTC()).Scan(&id, &m.ID, &m.Kind, &m.CreatedAt, &attempts, &m.IncidentID, &m.Category, &m.Scope, &m.OpenedAt, &m.RecoveredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	deliveryCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	sendErr := sender.Send(deliveryCtx, m)
	cancel()
	if sendErr != nil {
		delay := time.Minute
		for i := 0; i < attempts && delay < time.Hour; i++ {
			delay *= 2
		}
		if delay > time.Hour {
			delay = time.Hour
		}
		_, err = tx.Exec(ctx, `UPDATE notification_messages SET attempts=attempts+1,next_attempt_at=$2,last_error_code='smtp_delivery_failed' WHERE id=$1`, id, now.Add(delay).UTC())
	} else {
		_, err = tx.Exec(ctx, `UPDATE notification_messages SET attempts=attempts+1,sent_at=$2,last_error_code=NULL WHERE id=$1`, id, now.UTC())
		if err == nil {
			_, err = tx.Exec(ctx, `UPDATE notification_incidents SET last_notified_at=$2 WHERE id=$1`, m.IncidentID, now.UTC())
		}
	}
	if err != nil {
		return true, err
	}
	return true, tx.Commit(ctx)
}

func (m Message) Body() string {
	body := fmt.Sprintf("IWA operational incident %d\nCategory: %s\nScope identifier: %q\nNotification: %s\nIncident observed at: %s\nNotification queued at: %s\n", m.IncidentID, m.Category, m.Scope, m.Kind, m.OpenedAt.UTC().Format(time.RFC3339), m.CreatedAt.UTC().Format(time.RFC3339))
	if m.RecoveredAt != nil {
		body += "Recovery observed at: " + m.RecoveredAt.UTC().Format(time.RFC3339) + "\n"
	}
	return body + "Details and retained errors are available in local administration. This is a service diagnostic, not an official weather warning.\n"
}
