package notifications

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Delivery struct {
	Message
	Attempts      int
	SentAt        *time.Time
	NextAttemptAt time.Time
	ErrorCode     string
}
type Report struct {
	OpenIncidents, PendingMessages, TotalMessages int
	Limit                                         int
	Messages                                      []Delivery
	Incidents                                     []Incident
}

type Incident struct {
	ID                       int64
	Category, Scope          string
	OpenedAt, LastObservedAt time.Time
	Workload                 Workload
}

// Report exposes only operational history, never SMTP configuration or recipients.
func (s *Store) Report(ctx context.Context) (Report, error) {
	r := Report{Limit: 100, Messages: []Delivery{}, Incidents: []Incident{}}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return r, err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM notification_incidents WHERE recovered_at IS NULL),count(*) FILTER(WHERE sent_at IS NULL),count(*) FROM notification_messages`).Scan(&r.OpenIncidents, &r.PendingMessages, &r.TotalMessages)
	if err != nil {
		return r, err
	}
	rows, err := tx.Query(ctx, `SELECT m.message_id,m.kind,m.created_at,m.attempts,m.sent_at,m.next_attempt_at,COALESCE(m.last_error_code,''),i.id,i.category,i.scope,i.opened_at,i.recovered_at
 ,m.workload FROM notification_messages m JOIN notification_incidents i ON i.id=m.incident_id ORDER BY m.id DESC LIMIT 100`)
	if err != nil {
		return r, err
	}
	for rows.Next() {
		var d Delivery
		if err = rows.Scan(&d.ID, &d.Kind, &d.CreatedAt, &d.Attempts, &d.SentAt, &d.NextAttemptAt, &d.ErrorCode, &d.IncidentID, &d.Category, &d.Scope, &d.OpenedAt, &d.RecoveredAt, &d.Workload); err != nil {
			rows.Close()
			return r, err
		}
		r.Messages = append(r.Messages, d)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return r, err
	}
	rows, err = tx.Query(ctx, `SELECT id,category,scope,opened_at,last_observed_at,workload FROM notification_incidents WHERE recovered_at IS NULL ORDER BY id`)
	if err != nil {
		return r, err
	}
	for rows.Next() {
		var i Incident
		if err = rows.Scan(&i.ID, &i.Category, &i.Scope, &i.OpenedAt, &i.LastObservedAt, &i.Workload); err != nil {
			rows.Close()
			return r, err
		}
		r.Incidents = append(r.Incidents, i)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return r, err
	}
	return r, tx.Commit(ctx)
}
