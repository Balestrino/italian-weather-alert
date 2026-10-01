package processing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/jackc/pgx/v5"
)

type GatePolicy struct {
	FailureThreshold                         int
	InitialCooldown, MaxCooldown, ProbeLease time.Duration
}

func DefaultGatePolicy() GatePolicy {
	return GatePolicy{3, time.Minute, 15 * time.Minute, 2 * time.Minute}
}
func (p GatePolicy) Valid() bool {
	return p.FailureThreshold > 0 && p.FailureThreshold <= 100 && p.InitialCooldown >= time.Second && p.MaxCooldown >= p.InitialCooldown && p.MaxCooldown <= 24*time.Hour && p.ProbeLease >= time.Second && p.ProbeLease <= 10*time.Minute
}

type GatePermit struct {
	Scope, Model, Token string
	Generation          int64
	GlobalGeneration    int64
}
type GateState struct {
	Scope, Model, State, Reason string
	Failures                    int
	Generation                  int64
	CooldownMS                  int64
	AvailableAt                 time.Time
	ProbeToken                  string
	ProbeExpiresAt              *time.Time
	UpdatedAt                   time.Time
}
type GateBlocked struct{ State GateState }

func (e *GateBlocked) Error() string { return "provider_scope_held" }

func lockGate(ctx context.Context, tx pgx.Tx, scope, model string, now time.Time) (GateState, error) {
	_, err := tx.Exec(ctx, `INSERT INTO processing_provider_gates(scope,model,available_at,updated_at) VALUES($1,$2,$3,$3) ON CONFLICT DO NOTHING`, scope, model, now.UTC())
	if err != nil {
		return GateState{}, err
	}
	var g GateState
	err = tx.QueryRow(ctx, `SELECT scope,model,state,reason,failures,generation,cooldown_ms,available_at,probe_token,probe_expires_at,updated_at FROM processing_provider_gates WHERE scope=$1 AND model=$2 FOR UPDATE`, scope, model).Scan(&g.Scope, &g.Model, &g.State, &g.Reason, &g.Failures, &g.Generation, &g.CooldownMS, &g.AvailableAt, &g.ProbeToken, &g.ProbeExpiresAt, &g.UpdatedAt)
	return g, err
}

func (s *Store) AcquireGate(ctx context.Context, scope, model string, now time.Time, policy GatePolicy) (GatePermit, error) {
	if !validName(scope) || !validName(model) || model == "*" || now.IsZero() || !policy.Valid() {
		return GatePermit{}, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return GatePermit{}, err
	}
	defer tx.Rollback(ctx)
	global, err := lockGate(ctx, tx, scope, "*", now)
	if err != nil {
		return GatePermit{}, err
	}
	if global.State == "held" {
		return GatePermit{}, &GateBlocked{global}
	}
	g, err := lockGate(ctx, tx, scope, model, now)
	if err != nil {
		return GatePermit{}, err
	}
	p := GatePermit{Scope: scope, Model: model, Generation: g.Generation, GlobalGeneration: global.Generation}
	if g.State == "held" || g.State == "open" && (now.Before(g.AvailableAt) || g.ProbeExpiresAt != nil && now.Before(*g.ProbeExpiresAt)) {
		return GatePermit{}, &GateBlocked{g}
	}
	if g.State == "open" {
		var token [24]byte
		if _, err = rand.Read(token[:]); err != nil {
			return GatePermit{}, err
		}
		p.Token = hex.EncodeToString(token[:])
		_, err = tx.Exec(ctx, `UPDATE processing_provider_gates SET probe_token=$3,probe_expires_at=$4,updated_at=$5 WHERE scope=$1 AND model=$2`, scope, model, p.Token, now.Add(policy.ProbeLease), now.UTC())
		if err != nil {
			return GatePermit{}, err
		}
	}
	return p, tx.Commit(ctx)
}

// RecordGate accepts approved categories, never raw upstream text. A successful
// stale in-flight request cannot close a gate opened by a newer failure.
func (s *Store) RecordGate(ctx context.Context, p GatePermit, category string, retryAfter time.Duration, now time.Time, policy GatePolicy) error {
	if !validName(p.Scope) || !validName(p.Model) || p.Model == "*" || p.Generation < 1 || now.IsZero() || retryAfter < 0 || !policy.Valid() {
		return ErrInvalid
	}
	if !map[string]bool{"": true, "authentication": true, "quota": true, "permission": true, "request": true, "rate_limit": true, "availability": true, "transport": true, "invalid_response": true, "rejected": true}[category] {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	global, err := lockGate(ctx, tx, p.Scope, "*", now)
	if err != nil {
		return err
	}
	g, err := lockGate(ctx, tx, p.Scope, p.Model, now)
	if err != nil {
		return err
	}
	if category == "authentication" || category == "quota" {
		if global.Generation != p.GlobalGeneration {
			return tx.Commit(ctx)
		}
		if global.State != "held" || global.Reason != category {
			_, err = tx.Exec(ctx, `UPDATE processing_provider_gates SET state='held',reason=$2,generation=generation+1,updated_at=$3 WHERE scope=$1 AND model='*'`, p.Scope, category, now.UTC())
			if err != nil {
				return err
			}
			_, err = tx.Exec(ctx, `INSERT INTO processing_provider_gate_events(scope,model,action,reason,actor,created_at) VALUES($1,'*','held',$2,'provider',$3)`, p.Scope, category, now.UTC())
			if err != nil {
				return err
			}
		}
		return tx.Commit(ctx)
	}
	if global.State == "held" || g.Generation != p.Generation || g.ProbeToken != p.Token {
		return tx.Commit(ctx)
	}
	if category == "" {
		_, err = tx.Exec(ctx, `UPDATE processing_provider_gates SET state='closed',reason='',failures=0,cooldown_ms=0,probe_token='',probe_expires_at=NULL,generation=generation+1,updated_at=$3 WHERE scope=$1 AND model=$2`, p.Scope, p.Model, now.UTC())
	} else if category == "rate_limit" || category == "availability" || category == "transport" {
		failures := g.Failures + 1
		if failures > 1000000 {
			failures = 1000000
		}
		if failures >= policy.FailureThreshold || g.State == "open" || retryAfter > 0 {
			delay := policy.InitialCooldown
			if g.CooldownMS > 0 {
				delay = time.Duration(g.CooldownMS) * time.Millisecond * 2
			}
			if delay > policy.MaxCooldown {
				delay = policy.MaxCooldown
			}
			wait := delay
			if retryAfter > wait {
				wait = retryAfter
			}
			_, err = tx.Exec(ctx, `UPDATE processing_provider_gates SET state='open',reason=$3,failures=$4,cooldown_ms=$5,available_at=$6,probe_token='',probe_expires_at=NULL,generation=generation+1,updated_at=$7 WHERE scope=$1 AND model=$2`, p.Scope, p.Model, category, failures, delay.Milliseconds(), now.Add(wait).UTC(), now.UTC())
		} else {
			_, err = tx.Exec(ctx, `UPDATE processing_provider_gates SET failures=$3,reason=$4,updated_at=$5 WHERE scope=$1 AND model=$2`, p.Scope, p.Model, failures, category, now.UTC())
		}
	} else if g.State == "open" {
		// A permanent/model-specific failure in the half-open probe requires
		// correction; it must not leave an endless half-open retry loop.
		_, err = tx.Exec(ctx, `UPDATE processing_provider_gates SET state='held',reason=$3,probe_token='',probe_expires_at=NULL,generation=generation+1,updated_at=$4 WHERE scope=$1 AND model=$2`, p.Scope, p.Model, category, now.UTC())
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) ResumeGate(ctx context.Context, scope, model, actor string, now time.Time) error {
	if !validName(scope) || !validName(model) || !validName(actor) || now.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = lockGate(ctx, tx, scope, "*", now)
	if err != nil {
		return err
	}
	_, err = lockGate(ctx, tx, scope, model, now)
	if err != nil {
		return err
	}
	state := "open"
	if model == "*" {
		state = "closed"
	}
	_, err = tx.Exec(ctx, `UPDATE processing_provider_gates SET state=$3,reason='',failures=0,cooldown_ms=0,available_at=$4,probe_token='',probe_expires_at=NULL,generation=generation+1,updated_at=$4 WHERE scope=$1 AND model=$2`, scope, model, state, now.UTC())
	if err != nil {
		return err
	}
	if model == "*" {
		// Releasing credentials does not release an unrestricted backlog: each
		// previously seen model must first acquire its single recovery probe.
		_, err = tx.Exec(ctx, `UPDATE processing_provider_gates SET state='open',available_at=$2,probe_token='',probe_expires_at=NULL,generation=generation+1,updated_at=$2 WHERE scope=$1 AND model<>'*'`, scope, now.UTC())
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO processing_provider_gate_events(scope,model,action,reason,actor,created_at) VALUES($1,$2,'resume','operator_recovery',$3,$4)`, scope, model, actor, now.UTC())
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
