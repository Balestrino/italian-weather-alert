package notifications

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/jackc/pgx/v5"
	"time"
)

type ReportGenerator func(context.Context, time.Time) (ReportMail, error)
type ReportSender interface {
	SendReport(context.Context, string, string, string, time.Time) error
}
type DailyReportStatus struct {
	Date, Timezone, LocalTime            string
	ObservedAt, NextAttempt              time.Time
	SentAt                               *time.Time
	GenerationAttempts, DeliveryAttempts int
	Error                                *string
	Generated                            bool
}

func reportBackoff(attempt int) time.Duration {
	delay := time.Minute
	for i := 1; i < attempt && delay < time.Hour; i++ {
		delay *= 2
	}
	if delay > time.Hour {
		return time.Hour
	}
	return delay
}

// QueueDaily admits only today's missing snapshot. The daily row is also the
// durable generation failure record, and a generated body is never replaced.
func (s *Store) QueueDaily(ctx context.Context, cfg DailyReportConfig, at time.Time, generate ReportGenerator) (bool, error) {
	date, due := cfg.Due(at)
	if !due {
		return false, nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var locked bool
	if err = tx.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock(730065)").Scan(&locked); err != nil || !locked {
		return false, err
	}
	random := make([]byte, 16)
	if _, err = rand.Read(random); err != nil {
		return false, err
	}
	id := "report-" + date + "-" + hex.EncodeToString(random) + "@iwa.invalid"
	if _, err = tx.Exec(ctx, `INSERT INTO notification_daily_reports(report_date,timezone,local_time,message_id,observed_at,next_attempt_at) VALUES($1,$2,$3,$4,$5,$5) ON CONFLICT DO NOTHING`, date, cfg.Timezone, cfg.LocalTime, id, at); err != nil {
		return false, err
	}
	var generated bool
	var next time.Time
	var attempts int
	if err = tx.QueryRow(ctx, `SELECT body IS NOT NULL,next_attempt_at,generation_attempts FROM notification_daily_reports WHERE report_date=$1 FOR UPDATE`, date).Scan(&generated, &next, &attempts); err != nil {
		return false, err
	}
	if generated || at.Before(next) {
		return false, tx.Commit(ctx)
	}
	generationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	mail, generationErr := generate(generationCtx, at)
	cancel()
	attempts++
	if generationErr != nil {
		_, err = tx.Exec(ctx, `UPDATE notification_daily_reports SET generation_attempts=$2,next_attempt_at=$3,last_error_code='report_generation_failed' WHERE report_date=$1`, date, attempts, at.Add(reportBackoff(attempts)))
		if err != nil {
			return false, err
		}
		return false, tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, `UPDATE notification_daily_reports SET observed_at=$2,subject=$3,body=$4,generation_attempts=$5,last_error_code=NULL,next_attempt_at=$2 WHERE report_date=$1`, date, at, mail.Subject, mail.Body, attempts)
	if err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}
func (s *Store) DeliverReport(ctx context.Context, sender ReportSender, at time.Time) (bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var date, id, subject, body string
	var observed time.Time
	var attempts int
	err = tx.QueryRow(ctx, `SELECT report_date::text,message_id,subject,body,observed_at,delivery_attempts FROM notification_daily_reports WHERE body IS NOT NULL AND sent_at IS NULL AND next_attempt_at<=$1 ORDER BY report_date FOR UPDATE SKIP LOCKED LIMIT 1`, at).Scan(&date, &id, &subject, &body, &observed, &attempts)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if sender.SendReport(ctx, id, subject, body, observed) != nil {
		_, err = tx.Exec(ctx, `UPDATE notification_daily_reports SET delivery_attempts=delivery_attempts+1,next_attempt_at=$2,last_error_code='smtp_delivery_failed' WHERE report_date=$1`, date, at.Add(reportBackoff(attempts+1)))
	} else {
		_, err = tx.Exec(ctx, `UPDATE notification_daily_reports SET delivery_attempts=delivery_attempts+1,sent_at=$2,last_error_code=NULL WHERE report_date=$1`, date, at)
	}
	if err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}
func (s *Store) DailyReportStatuses(ctx context.Context) ([]DailyReportStatus, error) {
	rows, err := s.pool.Query(ctx, `SELECT report_date::text,timezone,local_time,observed_at,next_attempt_at,sent_at,generation_attempts,delivery_attempts,last_error_code,body IS NOT NULL FROM notification_daily_reports ORDER BY report_date DESC LIMIT 30`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []DailyReportStatus{}
	for rows.Next() {
		var r DailyReportStatus
		if err = rows.Scan(&r.Date, &r.Timezone, &r.LocalTime, &r.ObservedAt, &r.NextAttempt, &r.SentAt, &r.GenerationAttempts, &r.DeliveryAttempts, &r.Error, &r.Generated); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
