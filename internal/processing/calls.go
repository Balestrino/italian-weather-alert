package processing

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

type CallStart struct {
	RunID                                 int64
	AttemptNumber, Ordinal                int
	InputSHA256, Provider, RequestedModel string
	StartedAt                             time.Time
}

type CallFinish struct {
	ID                                                                                    int64
	FinishedAt                                                                            time.Time
	State                                                                                 string
	HTTPStatus                                                                            int
	ErrorCode, Category, ProviderCode, RequestID, ResponseID, ReturnedModel, FinishReason string
	Usage                                                                                 Usage
	RetryAfter                                                                            time.Duration
}

var callHash = regexp.MustCompile(`^[0-9a-f]{64}$`)
var callIdentifier = regexp.MustCompile(`^[A-Za-z0-9_.:/-]{0,200}$`)

// StartCall durably records intent. A duplicate ordinal is an error, not permission
// to transmit again after an uncertain commit.
func (s *Store) StartCall(ctx context.Context, input CallStart) (int64, error) {
	if input.RunID < 1 || input.AttemptNumber < 1 || input.Ordinal < 1 || !callHash.MatchString(input.InputSHA256) ||
		!validName(input.Provider) || !validName(input.RequestedModel) || input.StartedAt.IsZero() {
		return 0, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	if err = admitRecoveryCall(ctx, tx, input); err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRow(ctx, `INSERT INTO processing_provider_calls
 (run_id,attempt_number,ordinal,input_sha256,provider,requested_model,started_at)
 VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, input.RunID, input.AttemptNumber, input.Ordinal, input.InputSHA256, input.Provider, input.RequestedModel, input.StartedAt.UTC()).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit(ctx)
}

func (s *Store) FinishCall(ctx context.Context, input CallFinish) error {
	if input.ID < 1 || input.FinishedAt.IsZero() || (input.State != "received" && input.State != "failed" && input.State != "uncertain") ||
		(input.HTTPStatus != 0 && (input.HTTPStatus < 100 || input.HTTPStatus > 599)) || input.RetryAfter < 0 {
		return ErrInvalid
	}
	for _, v := range []string{input.ErrorCode, input.ProviderCode, input.RequestID, input.ResponseID, input.ReturnedModel, input.FinishReason} {
		if !callIdentifier.MatchString(v) {
			return ErrInvalid
		}
	}
	if !map[string]bool{"": true, "authentication": true, "permission": true, "quota": true, "rate_limit": true, "request": true, "availability": true, "rejected": true, "transport": true, "invalid_response": true}[input.Category] {
		return ErrInvalid
	}
	if input.Usage.Status != "reported" && input.Usage.Status != "partial" && input.Usage.Status != "unavailable" {
		return ErrInvalid
	}
	if _, _, err := validateUsage(input.Usage); err != nil {
		return err
	}
	if input.Usage.CacheWriteTokens != nil || len(input.Usage.OtherUnits) > 0 {
		return ErrInvalid
	}
	tag, err := s.pool.Exec(ctx, `UPDATE processing_provider_calls SET finished_at=$2,state=$3,http_status=NULLIF($4,0),
 error_code=$5,category=$6,provider_code=$7,request_id=$8,response_id=$9,returned_model=$10,finish_reason=$11,
 usage_status=$12,input_tokens=$13,output_tokens=$14,cache_read_tokens=$15,retry_after_ms=$16 WHERE id=$1 AND state='started'`,
		input.ID, input.FinishedAt.UTC(), input.State, input.HTTPStatus, input.ErrorCode, input.Category, input.ProviderCode, input.RequestID, input.ResponseID, input.ReturnedModel, input.FinishReason,
		input.Usage.Status, input.Usage.InputTokens, input.Usage.OutputTokens, input.Usage.CacheReadTokens, input.RetryAfter.Milliseconds())
	if err == nil && tag.RowsAffected() != 1 {
		return ErrFinished
	}
	return err
}

// CallUsage aggregates receipts for a single attempt, never legacy attempt totals.
// The bool distinguishes no ledger (legacy) from ledger with unknown consumption.
func (s *Store) CallUsage(ctx context.Context, runID int64, attempt int) (Usage, bool, error) {
	var u Usage
	var count, unknown, partial int
	err := s.pool.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE state IN ('started','uncertain') OR usage_status='unavailable'),
 count(*) FILTER(WHERE usage_status='partial'),sum(input_tokens),sum(output_tokens),sum(cache_read_tokens)
 FROM processing_provider_calls WHERE run_id=$1 AND attempt_number=$2`, runID, attempt).Scan(&count, &unknown, &partial, &u.InputTokens, &u.OutputTokens, &u.CacheReadTokens)
	if errors.Is(err, pgx.ErrNoRows) || count == 0 {
		return u, false, err
	}
	u.Status = "reported"
	if unknown > 0 || partial > 0 {
		u.Status = "partial"
	}
	if u.InputTokens == nil && u.OutputTokens == nil && u.CacheReadTokens == nil {
		u.Status = "unavailable"
	}
	u.OtherUnits = map[string]int64{"requests": int64(count)}
	if u.Status == "unavailable" {
		u.OtherUnits = nil
	}
	return u, true, err
}

// ResolveInterruptedCall requires proof the owner is no longer active from the
// caller (queue recovery), not an arbitrary age timeout on an active request.
func (s *Store) ResolveInterruptedCall(ctx context.Context, id int64, at time.Time) error {
	return s.FinishCall(ctx, CallFinish{ID: id, FinishedAt: at, State: "uncertain", ErrorCode: "call_interrupted", Usage: Usage{Status: "unavailable"}})
}

// RecoverInterruptedAttempts only closes work whose durable queue attempt is
// already terminal. Wall-clock age alone never revokes a live provider call.
func (s *Store) RecoverInterruptedAttempts(ctx context.Context, runID int64, at time.Time) error {
	rows, err := s.pool.Query(ctx, `SELECT a.run_id,a.number FROM processing_run_attempts a
 JOIN processing_attempts q ON q.job_id=a.queue_job_id AND q.number=a.queue_attempt_number
 WHERE a.outcome='running' AND q.outcome<>'running' AND ($1=0 OR a.run_id=$1)
 ORDER BY a.run_id,a.number`, runID)
	if err != nil {
		return err
	}
	type abandoned struct {
		run    int64
		number int
	}
	var attempts []abandoned
	for rows.Next() {
		var a abandoned
		if err = rows.Scan(&a.run, &a.number); err != nil {
			rows.Close()
			return err
		}
		attempts = append(attempts, a)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, a := range attempts {
		_, err = s.pool.Exec(ctx, `UPDATE processing_provider_calls SET state='uncertain',finished_at=GREATEST(started_at,$3),error_code='call_interrupted'
 WHERE run_id=$1 AND attempt_number=$2 AND state='started'`, a.run, a.number, at.UTC())
		if err != nil {
			return err
		}
		_, err = s.FinishAttempt(ctx, AttemptFinish{RunID: a.run, Number: a.number, FinishedAt: at.UTC(), Outcome: "failed", ErrorCode: "processing_interrupted", Usage: Usage{Status: "unavailable"}})
		if err != nil && !errors.Is(err, ErrFinished) {
			return err
		}
	}
	return nil
}
