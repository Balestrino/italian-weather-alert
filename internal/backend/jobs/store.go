package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Enqueue(ctx context.Context, req EnqueueRequest) (Job, error) {
	if !validName(req.Queue) || !validName(req.Kind) || !validName(req.IdempotencyKey) || req.MaxAttempts < 1 || req.MaxAttempts > 100 || req.RetryBase < time.Millisecond || req.RetryBase > 24*time.Hour || req.AvailableAt.IsZero() {
		return Job{}, ErrInvalid
	}
	payload, hash, err := canonicalObject(req.Payload)
	if err != nil {
		return Job{}, err
	}
	var job Job
	err = s.pool.QueryRow(ctx, `INSERT INTO processing_jobs(queue,kind,idempotency_key,payload,payload_hash,state,max_attempts,attempt_budget,retry_base_ms,available_at,created_at,updated_at)
 VALUES($1,$2,$3,$4,$5,'queued',$6,$6,$7,$8,$8,$8)
 ON CONFLICT(queue,kind,idempotency_key) DO NOTHING
 RETURNING id,queue,kind,idempotency_key,payload,state,max_attempts,attempt_count,available_at`, req.Queue, req.Kind, req.IdempotencyKey, payload, hash, req.MaxAttempts, req.RetryBase.Milliseconds(), req.AvailableAt.UTC()).Scan(&job.ID, &job.Queue, &job.Kind, &job.IdempotencyKey, &job.Payload, &job.State, &job.MaxAttempts, &job.Attempt, &job.AvailableAt)
	if err == nil {
		return job, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Job{}, err
	}
	var oldHash string
	var oldBase int64
	var oldBudget int
	err = s.pool.QueryRow(ctx, `SELECT id,queue,kind,idempotency_key,payload,state,max_attempts,attempt_count,available_at,payload_hash,retry_base_ms,attempt_budget
 FROM processing_jobs WHERE queue=$1 AND kind=$2 AND idempotency_key=$3`, req.Queue, req.Kind, req.IdempotencyKey).Scan(&job.ID, &job.Queue, &job.Kind, &job.IdempotencyKey, &job.Payload, &job.State, &job.MaxAttempts, &job.Attempt, &job.AvailableAt, &oldHash, &oldBase, &oldBudget)
	if err != nil {
		return Job{}, err
	}
	if oldHash != hash || oldBudget != req.MaxAttempts || oldBase != req.RetryBase.Milliseconds() {
		return Job{}, ErrConflict
	}
	return job, nil
}

// Claim selects one available job without blocking other workers. An expired
// lease becomes an abandoned persisted attempt before a new claim is created.
func (s *Store) Claim(ctx context.Context, queue, workerID string, now time.Time, lease time.Duration) (Claim, error) {
	if !validName(queue) || !validName(workerID) || now.IsZero() || lease < time.Second || lease > time.Hour {
		return Claim{}, ErrInvalid
	}
	claimToken, err := token()
	if err != nil {
		return Claim{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Claim{}, err
	}
	defer tx.Rollback(ctx)
	// A lease that consumed the last permitted attempt cannot remain running
	// forever merely because no later claim is possible.
	if _, err = tx.Exec(ctx, `UPDATE processing_attempts a SET outcome='failed',finished_at=$2,error_code='lease_expired',error_detail='worker lease expired'
FROM processing_jobs j WHERE j.id=a.job_id AND a.number=j.attempt_count AND a.outcome='running'
 AND j.queue=$1 AND j.state='running' AND j.lease_expires_at <= $2 AND j.attempt_count >= j.max_attempts`, queue, now.UTC()); err != nil {
		return Claim{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE processing_jobs SET state='failed',claimed_by=NULL,claim_token=NULL,lease_expires_at=NULL,
 last_error_code='lease_expired',last_error_detail='worker lease expired',updated_at=$2,completed_at=$2
WHERE queue=$1 AND state='running' AND lease_expires_at <= $2 AND attempt_count >= max_attempts`, queue, now.UTC()); err != nil {
		return Claim{}, err
	}
	gate, err := territory.Predicate(ctx, tx, "payload", true)
	if err != nil {
		return Claim{}, err
	}
	var c Claim
	var previousState string
	err = tx.QueryRow(ctx, `SELECT id,queue,kind,idempotency_key,payload,state,max_attempts,attempt_count,available_at
 FROM processing_jobs
 WHERE queue=$1 `+gate+` AND archived_at IS NULL AND attempt_count < max_attempts AND (
   (state IN ('queued','retry_wait') AND available_at <= $2)
   OR (state='running' AND lease_expires_at <= $2)
 )
	 ORDER BY available_at,id FOR UPDATE SKIP LOCKED LIMIT 1`, queue, now.UTC()).Scan(&c.ID, &c.Queue, &c.Kind, &c.IdempotencyKey, &c.Payload, &previousState, &c.MaxAttempts, &c.Attempt, &c.AvailableAt)
	if errors.Is(err, pgx.ErrNoRows) {
		if err = tx.Commit(ctx); err != nil {
			return Claim{}, err
		}
		return Claim{}, ErrNoJob
	}
	if err != nil {
		return Claim{}, err
	}
	if previousState == "running" {
		if _, err = tx.Exec(ctx, `UPDATE processing_attempts SET outcome='abandoned',finished_at=$3,error_code='lease_expired',error_detail='worker lease expired'
 WHERE job_id=$1 AND number=$2 AND outcome='running'`, c.ID, c.Attempt, now.UTC()); err != nil {
			return Claim{}, err
		}
	}
	c.Attempt++
	c.State = "running"
	c.WorkerID = workerID
	c.Token = claimToken
	c.LeaseExpiresAt = now.UTC().Add(lease)
	if _, err = tx.Exec(ctx, `UPDATE processing_jobs SET state='running',attempt_count=$2,claimed_by=$3,claim_token=$4,lease_expires_at=$5,updated_at=$6
 WHERE id=$1`, c.ID, c.Attempt, workerID, claimToken, c.LeaseExpiresAt, now.UTC()); err != nil {
		return Claim{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO processing_attempts(job_id,number,worker_id,claim_token,started_at,lease_expires_at,outcome)
 VALUES($1,$2,$3,$4,$5,$6,'running')`, c.ID, c.Attempt, workerID, claimToken, now.UTC(), c.LeaseExpiresAt); err != nil {
		return Claim{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Claim{}, err
	}
	return c, nil
}

func (s *Store) Heartbeat(ctx context.Context, c Claim, now time.Time, lease time.Duration) error {
	if now.IsZero() || lease < time.Second || lease > time.Hour {
		return ErrInvalid
	}
	expires := now.UTC().Add(lease)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE processing_jobs SET lease_expires_at=$4,updated_at=$3
 WHERE id=$1 AND state='running' AND attempt_count=$2 AND claim_token=$5 AND lease_expires_at>$3`, c.ID, c.Attempt, now.UTC(), expires, c.Token)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrStaleClaim
	}
	tag, err = tx.Exec(ctx, `UPDATE processing_attempts SET lease_expires_at=$3 WHERE job_id=$1 AND number=$2 AND outcome='running'`, c.ID, c.Attempt, expires)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrStaleClaim
	}
	return tx.Commit(ctx)
}

func (s *Store) Complete(ctx context.Context, c Claim, result Result, now time.Time) error {
	if now.IsZero() {
		return ErrInvalid
	}
	payload, _, err := canonicalObject(result.Payload)
	if err != nil {
		return err
	}
	type preparedEffect struct {
		Effect
		payload json.RawMessage
		hash    string
	}
	prepared := make([]preparedEffect, 0, len(result.Effects))
	seen := map[string]bool{}
	for _, effect := range result.Effects {
		if !validName(effect.Key) || !validName(effect.Kind) || seen[effect.Key] {
			return ErrInvalid
		}
		seen[effect.Key] = true
		body, hash, e := canonicalObject(effect.Payload)
		if e != nil {
			return e
		}
		prepared = append(prepared, preparedEffect{Effect: effect, payload: body, hash: hash})
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = lockClaim(ctx, tx, c, now); err != nil {
		return err
	}
	for _, effect := range prepared {
		tag, e := tx.Exec(ctx, `INSERT INTO processing_effects(effect_key,job_id,kind,payload,payload_hash,created_at)
 VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(effect_key) DO NOTHING`, effect.Key, c.ID, effect.Kind, effect.payload, effect.hash, now.UTC())
		if e != nil {
			return e
		}
		if tag.RowsAffected() == 0 {
			var kind, hash string
			if e = tx.QueryRow(ctx, "SELECT kind,payload_hash FROM processing_effects WHERE effect_key=$1", effect.Key).Scan(&kind, &hash); e != nil {
				return e
			}
			if kind != effect.Kind || hash != effect.hash {
				return ErrConflict
			}
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE processing_attempts SET outcome='succeeded',finished_at=$3
 WHERE job_id=$1 AND number=$2`, c.ID, c.Attempt, now.UTC()); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE processing_jobs SET state='succeeded',result=$2,claimed_by=NULL,claim_token=NULL,lease_expires_at=NULL,
 last_error_code=NULL,last_error_detail=NULL,updated_at=$3,completed_at=$3 WHERE id=$1`, c.ID, payload, now.UTC()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) Fail(ctx context.Context, c Claim, failure Failure, now time.Time) error {
	if now.IsZero() {
		return ErrInvalid
	}
	failure, err := cleanFailure(failure)
	if err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var maxAttempts, cycleStart int
	var baseMS int64
	if err = tx.QueryRow(ctx, `SELECT max_attempts,retry_base_ms,COALESCE((SELECT max(after_attempt) FROM processing_job_relaunches WHERE job_id=processing_jobs.id),0) FROM processing_jobs
 WHERE id=$1 AND state='running' AND attempt_count=$2 AND claim_token=$3 AND lease_expires_at>$4 FOR UPDATE`, c.ID, c.Attempt, c.Token, now.UTC()).Scan(&maxAttempts, &baseMS, &cycleStart); errors.Is(err, pgx.ErrNoRows) {
		return ErrStaleClaim
	} else if err != nil {
		return err
	}
	retry := failure.Temporary && c.Attempt < maxAttempts
	outcome, state := "failed", "failed"
	available, completed := now.UTC(), any(now.UTC())
	if retry {
		outcome, state, completed = "retry", "retry_wait", nil
		var deferred int
		if err = tx.QueryRow(ctx, `SELECT count(*) FROM processing_attempts WHERE job_id=$1 AND number>$2 AND number<$3 AND outcome='deferred'`, c.ID, cycleStart, c.Attempt).Scan(&deferred); err != nil {
			return err
		}
		multiplier := int64(1) << min(max(c.Attempt-cycleStart-deferred-1, 0), 20)
		delayMS := baseMS * multiplier
		if delayMS > int64((24 * time.Hour).Milliseconds()) {
			delayMS = int64((24 * time.Hour).Milliseconds())
		}
		available = now.UTC().Add(time.Duration(delayMS) * time.Millisecond)
	}
	if _, err = tx.Exec(ctx, `UPDATE processing_attempts SET outcome=$3,finished_at=$4,error_code=$5,error_detail=$6
 WHERE job_id=$1 AND number=$2`, c.ID, c.Attempt, outcome, now.UTC(), failure.Code, failure.Detail); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE processing_jobs SET state=$2,available_at=$3,claimed_by=NULL,claim_token=NULL,lease_expires_at=NULL,
 last_error_code=$4,last_error_detail=$5,updated_at=$6,completed_at=$7 WHERE id=$1`, c.ID, state, available, failure.Code, failure.Detail, now.UTC(), completed); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func lockClaim(ctx context.Context, tx pgx.Tx, c Claim, now time.Time) error {
	var one int
	err := tx.QueryRow(ctx, `SELECT 1 FROM processing_jobs WHERE id=$1 AND state='running' AND attempt_count=$2 AND claim_token=$3 AND lease_expires_at>$4 FOR UPDATE`, c.ID, c.Attempt, c.Token, now.UTC()).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrStaleClaim
	}
	return err
}

func (s *Store) Attempts(ctx context.Context, jobID int64) ([]Attempt, error) {
	rows, err := s.pool.Query(ctx, `SELECT job_id,number,worker_id,outcome,COALESCE(error_code,''),COALESCE(error_detail,''),started_at,lease_expires_at,finished_at
 FROM processing_attempts WHERE job_id=$1 ORDER BY number`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var attempts []Attempt
	for rows.Next() {
		var a Attempt
		if err = rows.Scan(&a.JobID, &a.Number, &a.WorkerID, &a.Outcome, &a.ErrorCode, &a.ErrorDetail, &a.StartedAt, &a.LeaseExpiresAt, &a.FinishedAt); err != nil {
			return nil, err
		}
		attempts = append(attempts, a)
	}
	return attempts, rows.Err()
}
