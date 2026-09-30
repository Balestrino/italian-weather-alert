// Package processing persists immutable inference/collection configuration and
// explicit per-attempt timing, usage and cost provenance.
package processing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid processing record")
	ErrConflict = errors.New("version or idempotency key reused with different content")
	ErrNotFound = errors.New("processing record not found")
	ErrFinished = errors.New("processing attempt already finished")
)

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

const maxJSONBytes = 1 << 20

var validStages = map[string]bool{"collection": true, "ocr": true, "classification": true, "extraction": true, "linking": true, "embedding": true}
var validWorkloads = map[string]bool{"bootstrap": true, "ordinary": true, "evaluation": true, "reprocessing": true}

type ModelVersion struct {
	ID, Provider, Model, Revision string
	Capabilities                  json.RawMessage
	CreatedAt                     time.Time
}

type PromptVersion struct {
	ID, Name, Stage, Revision string
	Body                      string
	CreatedAt                 time.Time
}

type Rate struct {
	Metric          string
	UnitSize        int64
	PriceMicrounits int64
}

type PriceVersion struct {
	ID, ModelVersionID, Currency, ProvenanceURL string
	ObservedAt, CreatedAt                       time.Time
	EffectiveFrom, EffectiveThrough             *time.Time
	Details                                     json.RawMessage
	Rates                                       []Rate
}

type ConfigurationVersion struct {
	ID, Name, Stage, Revision string
	ModelVersionID            *string
	PromptVersionID           *string
	LogicVersion              string
	Settings                  json.RawMessage
	CreatedAt                 time.Time
}

type RunRequest struct {
	IdempotencyKey, Workload, Stage, ConfigurationVersionID string
	SourceID                                                *string
	DocumentVersionID                                       *int64
	Subject                                                 json.RawMessage
	CreatedAt                                               time.Time
}

type Run struct {
	ID                                                      int64
	IdempotencyKey, Workload, Stage, ConfigurationVersionID string
	SourceID                                                *string
	DocumentVersionID                                       *int64
	Subject                                                 json.RawMessage
	CreatedAt                                               time.Time
	Attempts                                                []Attempt
}

type Attempt struct {
	RunID                                       int64
	Number                                      int
	QueueJobID                                  *int64
	QueueAttemptNumber                          *int
	StartedAt                                   time.Time
	FinishedAt                                  *time.Time
	DurationMS                                  *int64
	Outcome, UsageStatus, CostStatus, ErrorCode string
	InputTokens, OutputTokens                   *int64
	CacheReadTokens, CacheWriteTokens           *int64
	OtherUnits                                  map[string]int64
	PriceVersionID                              *string
	EstimatedCostMicrounits                     *int64
}

type AttemptStart struct {
	RunID              int64
	QueueJobID         *int64
	QueueAttemptNumber *int
	StartedAt          time.Time
}

type Usage struct {
	Status                            string
	InputTokens, OutputTokens         *int64
	CacheReadTokens, CacheWriteTokens *int64
	OtherUnits                        map[string]int64
}

type AttemptFinish struct {
	RunID          int64
	Number         int
	FinishedAt     time.Time
	Outcome        string
	Usage          Usage
	PriceVersionID *string
	ErrorCode      string
}

type Summary struct {
	Workload, Stage, Currency                          string
	Runs, Attempts, UnknownUsageAttempts, UnknownCosts int64
	InputTokens, OutputTokens                          *int64
	CacheReadTokens, CacheWriteTokens                  *int64
	EstimatedCostMicrounits                            *int64
}

func validName(value string) bool {
	return value != "" && value == strings.TrimSpace(value) && len(value) <= 200 && !strings.ContainsAny(value, "\r\n")
}

func validHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https") && u.User == nil && u.Fragment == ""
}

func canonicalObject(raw json.RawMessage) (json.RawMessage, string, error) {
	if len(raw) == 0 || len(raw) > maxJSONBytes {
		return nil, "", ErrInvalid
	}
	var value map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil || value == nil {
		return nil, "", ErrInvalid
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, "", ErrInvalid
	}
	body, err := json.Marshal(value)
	if err != nil {
		return nil, "", ErrInvalid
	}
	hash := sha256.Sum256(body)
	return body, hex.EncodeToString(hash[:]), nil
}

func digest(parts ...[]byte) string {
	hash := sha256.New()
	for _, part := range parts {
		hash.Write(part)
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}
