package backups

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/Balestrino/italian-weather-alert/internal/backend/notifications"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool} }

type Run struct {
	ID                              string
	StartedAt                       time.Time
	FinishedAt                      *time.Time
	NextDueAt                       time.Time
	State, ErrorCode, ArchiveSHA256 string
	ArchiveBytes                    int64
	Objects                         int
}
type Report struct {
	Configured, Enabled            bool
	IntervalSeconds, RetentionDays int
	ObservedAt                     *time.Time
	Total                          int
	Runs                           []Run
}

func (s *Store) Report(ctx context.Context) (Report, error) {
	r := Report{Runs: []Run{}}
	rows, err := s.pool.Query(ctx, `SELECT enabled,interval_seconds,retention_days,observed_at FROM backup_schedule`)
	if err != nil {
		return r, err
	}
	if rows.Next() {
		r.Configured = true
		err = rows.Scan(&r.Enabled, &r.IntervalSeconds, &r.RetentionDays, &r.ObservedAt)
	}
	rows.Close()
	if err != nil {
		return r, err
	}
	if rows.Err() != nil {
		return r, rows.Err()
	}
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM backup_runs`).Scan(&r.Total); err != nil {
		return r, err
	}
	rows, err = s.pool.Query(ctx, `SELECT id,started_at,finished_at,next_due_at,state,error_code,archive_sha256,archive_bytes,objects FROM backup_runs ORDER BY started_at DESC,id DESC LIMIT 100`)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		var run Run
		if err = rows.Scan(&run.ID, &run.StartedAt, &run.FinishedAt, &run.NextDueAt, &run.State, &run.ErrorCode, &run.ArchiveSHA256, &run.ArchiveBytes, &run.Objects); err != nil {
			return r, err
		}
		r.Runs = append(r.Runs, run)
	}
	return r, rows.Err()
}

type Worker struct {
	Store       *Store
	Config      Config
	Objects     ObjectReader
	Dump        Dumper
	Destination Destination
}

func (w *Worker) Run(ctx context.Context) error {
	if w.Store == nil || w.Config.Validate() != nil || (w.Config.Enabled && (w.Objects == nil || w.Dump == nil || w.Destination == nil)) {
		return ErrConfig
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return nil
		}
		if _, err := w.Tick(ctx, time.Now()); err != nil && ctx.Err() == nil {
			slog.Error("backup cycle failed; inspect local backup history")
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// Tick serializes schedules across processes and remembers due times across
// restarts. An interrupted run is a failure; only a new verified backup recovers.
func (w *Worker) Tick(ctx context.Context, now time.Time) (bool, error) {
	if w.Store == nil || w.Config.Validate() != nil || now.IsZero() || (w.Config.Enabled && (w.Objects == nil || w.Dump == nil || w.Destination == nil)) {
		return false, ErrConfig
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(max(w.Config.TimeoutSeconds, 10))*time.Second)
	defer cancel()
	conn, err := w.Store.pool.Acquire(ctx)
	if err != nil {
		return false, ErrBackup
	}
	defer func() {
		c, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_ = conn.Conn().Close(c)
		conn.Release()
	}()
	var locked bool
	if conn.QueryRow(ctx, `SELECT pg_try_advisory_lock(730027)`).Scan(&locked) != nil {
		return false, ErrBackup
	}
	if !locked {
		return false, nil
	}
	if _, err = conn.Exec(ctx, `INSERT INTO backup_schedule(singleton,enabled,interval_seconds,retention_days,observed_at) VALUES(true,$1,$2,$3,$4) ON CONFLICT(singleton) DO UPDATE SET enabled=EXCLUDED.enabled,interval_seconds=EXCLUDED.interval_seconds,retention_days=EXCLUDED.retention_days,observed_at=EXCLUDED.observed_at`, w.Config.Enabled, w.Config.IntervalSeconds, w.Config.RetentionDays, now.UTC()); err != nil {
		return false, ErrBackup
	}
	if _, err = conn.Exec(ctx, `UPDATE backup_runs SET state='failed',error_code='interrupted',finished_at=$1 WHERE state='running'`, now.UTC()); err != nil {
		return false, ErrBackup
	}
	if w.flush(ctx) != nil {
		return false, ErrBackup
	}
	if !w.Config.Enabled {
		return false, nil
	}
	var due bool
	if conn.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM backup_runs WHERE started_at + make_interval(secs=>$2) > $1)`, now.UTC(), w.Config.IntervalSeconds).Scan(&due) != nil {
		return false, ErrBackup
	}
	if !due {
		return false, nil
	}
	token := make([]byte, 16)
	if _, err = rand.Read(token); err != nil {
		return false, ErrBackup
	}
	id := hex.EncodeToString(token)
	if _, err = conn.Exec(ctx, `INSERT INTO backup_runs(id,started_at,next_due_at,state) VALUES($1,$2,$3,'running')`, id, now.UTC(), now.Add(time.Duration(w.Config.IntervalSeconds)*time.Second).UTC()); err != nil {
		return false, ErrBackup
	}
	code := "snapshot_failed"
	b, cleanup, buildErr := Build(ctx, w.Store.pool, w.Objects, w.Dump, w.Config.ScratchDirectory, id)
	defer cleanup()
	if buildErr == nil {
		code = "transfer_failed"
		if w.Destination.Publish(ctx, b, time.Now()) == nil {
			code = "retention_failed"
			if w.Destination.Prune(ctx, now.Add(-time.Duration(w.Config.RetentionDays)*24*time.Hour), id) == nil {
				code = ""
			}
		}
	}
	state := "failed"
	if code == "" {
		state = "succeeded"
	}
	// Record cancellation too, independently from the cancelled work context.
	finishCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	result, err := conn.Exec(finishCtx, `UPDATE backup_runs SET state=$2,error_code=$3,finished_at=$4,archive_sha256=$5,archive_bytes=$6,objects=$7 WHERE id=$1 AND state='running'`, id, state, code, time.Now().UTC(), b.SHA256, b.Size, len(b.Manifest.Objects))
	if err != nil || result.RowsAffected() != 1 {
		return true, ErrBackup
	}
	if w.flush(finishCtx) != nil {
		return true, ErrBackup
	}
	if code != "" {
		return true, ErrBackup
	}
	return true, nil
}
func (w *Worker) flush(ctx context.Context) error {
	rows, err := w.Store.pool.Query(ctx, `SELECT id,state,finished_at FROM backup_runs WHERE state<>'running' AND NOT notified ORDER BY finished_at,id`)
	if err != nil {
		return err
	}
	type outcome struct {
		id, state string
		at        time.Time
	}
	var pending []outcome
	for rows.Next() {
		var o outcome
		if err = rows.Scan(&o.id, &o.state, &o.at); err != nil {
			rows.Close()
			return err
		}
		pending = append(pending, o)
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}
	for _, o := range pending {
		if err = notifications.New(w.Store.pool).RecordBackup(ctx, "scheduled-backup", o.id, o.state == "succeeded", o.at); err != nil {
			return err
		}
		if _, err = w.Store.pool.Exec(ctx, `UPDATE backup_runs SET notified=true WHERE id=$1`, o.id); err != nil {
			return err
		}
	}
	return nil
}
