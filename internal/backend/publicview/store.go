package publicview

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("public query view not found")
var ErrInvalid = errors.New("invalid public query view")

type Create struct {
	Operation               string
	RequestHash             string
	Request, Data, History  json.RawMessage
	EvaluationTime, KnownAt time.Time
	CreatedAt, ExpiresAt    time.Time
}

type View struct {
	ID                      string
	Operation               string
	RequestHash             string
	Request, Data, History  json.RawMessage
	EvaluationTime, KnownAt time.Time
	CreatedAt, ExpiresAt    time.Time
}

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Create(ctx context.Context, request Create) (View, error) {
	if !validCreate(request) {
		return View{}, ErrInvalid
	}
	identifier := make([]byte, 18)
	if _, err := rand.Read(identifier); err != nil {
		return View{}, err
	}
	view := View{ID: base64.RawURLEncoding.EncodeToString(identifier), Operation: request.Operation, RequestHash: request.RequestHash, Request: append(json.RawMessage(nil), request.Request...), Data: append(json.RawMessage(nil), request.Data...), History: append(json.RawMessage(nil), request.History...), EvaluationTime: request.EvaluationTime.UTC(), KnownAt: request.KnownAt.UTC(), CreatedAt: request.CreatedAt.UTC(), ExpiresAt: request.ExpiresAt.UTC()}
	_, err := s.pool.Exec(ctx, `WITH expired AS (DELETE FROM public_query_views WHERE expires_at<=$9)
 INSERT INTO public_query_views(id,operation,request_hash,request,data,history,evaluation_time,known_at,created_at,expires_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, view.ID, view.Operation, view.RequestHash, view.Request, view.Data, view.History, view.EvaluationTime, view.KnownAt, view.CreatedAt, view.ExpiresAt)
	return view, err
}

func validCreate(request Create) bool {
	validOperation := request.Operation == "discover_municipalities" || request.Operation == "get_municipality_situation" || request.Operation == "search_alerts" || request.Operation == "get_document" || request.Operation == "get_source_coverage"
	_, hashErr := hex.DecodeString(request.RequestHash)
	if !validOperation || len(request.RequestHash) != 64 || hashErr != nil || !json.Valid(request.Data) || request.EvaluationTime.IsZero() || request.KnownAt.IsZero() || request.CreatedAt.IsZero() || !request.ExpiresAt.After(request.CreatedAt) || request.KnownAt.After(request.EvaluationTime) {
		return false
	}
	var requestObject, historyObject map[string]any
	return json.Unmarshal(request.Request, &requestObject) == nil && requestObject != nil && json.Unmarshal(request.History, &historyObject) == nil && historyObject != nil
}

func (s *Store) Get(ctx context.Context, id string) (View, error) {
	var view View
	err := s.pool.QueryRow(ctx, `SELECT id,operation,request_hash,request,data,history,evaluation_time,known_at,created_at,expires_at FROM public_query_views WHERE id=$1`, id).Scan(&view.ID, &view.Operation, &view.RequestHash, &view.Request, &view.Data, &view.History, &view.EvaluationTime, &view.KnownAt, &view.CreatedAt, &view.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return View{}, ErrNotFound
	}
	view.EvaluationTime, view.KnownAt = view.EvaluationTime.UTC(), view.KnownAt.UTC()
	view.CreatedAt, view.ExpiresAt = view.CreatedAt.UTC(), view.ExpiresAt.UTC()
	return view, err
}
