// Package linking selects prior measure candidates and records evidence-backed
// update relationships without applying domain-state changes.
package linking

import (
	"errors"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
)

var (
	ErrInvalid  = errors.New("invalid linking request")
	ErrConflict = errors.New("linking result conflict")
)

const Kind = "link_measure_update"

type MeasureContext struct {
	RunID             int64                 `json:"extraction_run_id"`
	DocumentVersionID int64                 `json:"document_version_id"`
	Ordinal           int                   `json:"measure_ordinal"`
	SourceID          string                `json:"source_id"`
	Municipality      string                `json:"municipality"`
	Kind              string                `json:"kind"`
	Subject           string                `json:"subject"`
	Place             string                `json:"place,omitempty"`
	OfficialURL       string                `json:"official_url"`
	FirstAcquiredAt   time.Time             `json:"first_acquired_at"`
	Evidence          []extraction.Evidence `json:"evidence"`
}

type Candidate struct {
	MeasureContext
	TextRank        float32  `json:"text_rank"`
	ExactPlace      bool     `json:"exact_place"`
	RetrievalMethod string   `json:"retrieval_method"`
	SemanticScore   *float32 `json:"semantic_score,omitempty"`
}

type Result struct {
	RunID, CurrentRunID                        int64
	CurrentOrdinal                             int
	Status, ReasonCode                         string
	Relation                                   *string
	CandidateRunID                             *int64
	CandidateOrdinal                           *int
	Candidates                                 []Candidate
	CurrentEvidence                            []extraction.Evidence
	CandidateEvidence                          []extraction.Evidence
	ProviderResponseID, ReturnedModel          string
	InputTokens, OutputTokens, CacheReadTokens *int64
	CreatedAt                                  time.Time
}
