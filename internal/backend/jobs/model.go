// Package jobs provides a durable PostgreSQL work queue and transactional
// logical-effect outbox. Handlers remain responsible for idempotent external
// effects, using the job idempotency key when a transaction cannot span them.
package jobs

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

var (
	ErrInvalid    = errors.New("invalid job")
	ErrConflict   = errors.New("idempotency key reused with different job or effect")
	ErrNoJob      = errors.New("no job available")
	ErrStaleClaim = errors.New("job claim is no longer current")
)

const MaxPayloadBytes = 1 << 20

type EnqueueRequest struct {
	Queue, Kind, IdempotencyKey string
	Payload                     json.RawMessage
	MaxAttempts                 int
	RetryBase                   time.Duration
	AvailableAt                 time.Time
}

type Job struct {
	ID                          int64
	Queue, Kind, IdempotencyKey string
	Payload                     json.RawMessage
	State                       string
	MaxAttempts, Attempt        int
	AvailableAt                 time.Time
}

type Claim struct {
	Job
	WorkerID, Token string
	LeaseExpiresAt  time.Time
}

type Failure struct {
	Code, Detail string
	Temporary    bool
}

type Effect struct {
	Key, Kind string
	Payload   json.RawMessage
}

type Result struct {
	Payload json.RawMessage
	Effects []Effect
}

type Attempt struct {
	JobID                        int64
	Number                       int
	WorkerID, Outcome, ErrorCode string
	ErrorDetail                  string
	StartedAt, LeaseExpiresAt    time.Time
	FinishedAt                   *time.Time
}

func canonicalObject(raw json.RawMessage) (json.RawMessage, string, error) {
	if len(raw) == 0 || len(raw) > MaxPayloadBytes {
		return nil, "", ErrInvalid
	}
	var value map[string]any
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if dec.Decode(&value) != nil || value == nil {
		return nil, "", ErrInvalid
	}
	if err := dec.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, "", ErrInvalid
	}
	b, err := json.Marshal(value)
	if err != nil {
		return nil, "", ErrInvalid
	}
	h := sha256.Sum256(b)
	return b, hex.EncodeToString(h[:]), nil
}

func token() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func validName(s string) bool {
	return s != "" && s == strings.TrimSpace(s) && len(s) <= 200 && !strings.ContainsAny(s, "\r\n")
}

func cleanFailure(f Failure) (Failure, error) {
	f.Code = strings.TrimSpace(f.Code)
	f.Detail = strings.TrimSpace(f.Detail)
	if !validName(f.Code) || len(f.Detail) > 1000 || strings.ContainsAny(f.Detail, "\r\n") {
		return Failure{}, ErrInvalid
	}
	return f, nil
}
