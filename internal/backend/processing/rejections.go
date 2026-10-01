package processing

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Rejection struct {
	Scope, Model, ConfigurationID, InputSHA256, ErrorCode, Category string
	HTTPStatus                                                      int
	CreatedAt                                                       time.Time
}

func (s *Store) Rejected(ctx context.Context, scope, model, configuration, hash string) (Rejection, bool, error) {
	r := Rejection{Scope: scope, Model: model, ConfigurationID: configuration, InputSHA256: hash}
	err := s.pool.QueryRow(ctx, `SELECT error_code,http_status,category,created_at FROM processing_provider_rejections WHERE scope=$1 AND model=$2 AND configuration_id=$3 AND input_sha256=$4`, scope, model, configuration, hash).Scan(&r.ErrorCode, &r.HTTPStatus, &r.Category, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, false, nil
	}
	return r, err == nil, err
}
func (s *Store) Reject(ctx context.Context, r Rejection) error {
	if !validName(r.Scope) || !validName(r.Model) || !validName(r.ConfigurationID) || !callHash.MatchString(r.InputSHA256) || !callIdentifier.MatchString(r.ErrorCode) || r.CreatedAt.IsZero() || r.HTTPStatus < 100 || r.HTTPStatus > 599 || (r.Category != "request" && r.Category != "permission" && r.Category != "rejected") {
		return ErrInvalid
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO processing_provider_rejections(scope,model,configuration_id,input_sha256,error_code,http_status,category,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT DO NOTHING`, r.Scope, r.Model, r.ConfigurationID, r.InputSHA256, r.ErrorCode, r.HTTPStatus, r.Category, r.CreatedAt.UTC())
	return err
}
func (s *Store) ClearRejection(ctx context.Context, scope, model, configuration, hash, actor string, now time.Time) error {
	if !validName(scope) || !validName(model) || !validName(configuration) || !callHash.MatchString(hash) || !validName(actor) || now.IsZero() {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `DELETE FROM processing_provider_rejections WHERE scope=$1 AND model=$2 AND configuration_id=$3 AND input_sha256=$4`, scope, model, configuration, hash)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO processing_provider_gate_events(scope,model,action,reason,actor,created_at) VALUES($1,$2,'clear_rejection',$3,$4,$5)`, scope, model, configuration+":"+hash, actor, now.UTC())
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type GateBinding struct {
	Kind, Scope, Model string
	Fallback           *FallbackBinding
}

func (s *Store) DeferHeldJobs(ctx context.Context, bindings []GateBinding, now time.Time) error {
	for _, b := range bindings {
		fallback := FallbackBinding{}
		if b.Fallback != nil {
			fallback = *b.Fallback
		}
		fallbackFilter := ""
		if b.Fallback != nil {
			fallbackFilter = ` AND NOT (` + fallbackSourceSQL + ` AND ` + fallbackPrimaryBlockedSQL + ` AND NOT ` + fallbackLocalBlockedSQL + `)`
		}
		arguments := []any{b.Kind, b.Scope, b.Model, now.UTC()}
		if b.Fallback != nil {
			arguments = append(arguments, fallback.Scope, fallback.Model, fallback.Sources)
		}
		_, err := s.pool.Exec(ctx, `UPDATE processing_jobs j SET available_at=GREATEST(j.available_at,g.wait_until),last_error_code='provider_scope_held',last_error_detail='waiting for provider recovery',updated_at=$4
 FROM (SELECT max(CASE WHEN state='held' THEN $4::timestamptz+interval '1 minute' ELSE GREATEST(available_at,COALESCE(probe_expires_at,available_at)) END) wait_until
 FROM processing_provider_gates WHERE scope=$2 AND model IN ('*',$3) AND state<>'closed') g
 WHERE j.archived_at IS NULL AND j.queue='inference' AND j.kind=$1 AND j.state IN ('queued','retry_wait') AND g.wait_until>$4`+fallbackFilter, arguments...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Gates(ctx context.Context) ([]GateState, error) {
	rows, err := s.pool.Query(ctx, `SELECT scope,model,state,reason,failures,generation,cooldown_ms,available_at,updated_at FROM processing_provider_gates ORDER BY scope,model`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []GateState{}
	for rows.Next() {
		var g GateState
		if err = rows.Scan(&g.Scope, &g.Model, &g.State, &g.Reason, &g.Failures, &g.Generation, &g.CooldownMS, &g.AvailableAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, rows.Err()
}

func (s *Store) Rejections(ctx context.Context) ([]Rejection, error) {
	rows, err := s.pool.Query(ctx, `SELECT scope,model,configuration_id,input_sha256,error_code,http_status,category,created_at FROM processing_provider_rejections ORDER BY created_at DESC LIMIT 1000`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Rejection{}
	for rows.Next() {
		var r Rejection
		if err = rows.Scan(&r.Scope, &r.Model, &r.ConfigurationID, &r.InputSHA256, &r.ErrorCode, &r.HTTPStatus, &r.Category, &r.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
