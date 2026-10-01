// Package extraction turns relevant retained documents into evidence-backed
// candidate measures. It records model output as interpretation, never as a
// correction of the retained source.
package extraction

import (
	"errors"
	"time"
)

var (
	ErrInvalid        = errors.New("invalid extraction request")
	ErrEvidence       = errors.New("invalid extraction evidence")
	ErrConflict       = errors.New("extraction result conflict")
	ErrMergeDuplicate = errors.New("duplicate extracted fact")
	ErrMergeConflict  = errors.New("conflicting extracted facts")
)

const Kind = "extract_measures"

type Evidence struct {
	Field          string `json:"field"`
	ResourceURL    string `json:"resource_url"`
	Page           *int   `json:"page"`
	Quote          string `json:"quote"`
	SegmentOrdinal int    `json:"-"`
}

// TemporalCandidate preserves one supported original expression before the
// resolved measure view is projected. Conflicting candidates share one stable
// identity while their evidence remains independently attributable.
type TemporalCandidate struct {
	Ordinal            int
	Field              string
	OriginalExpression string
	ConflictIdentity   *string
	Evidence           []Evidence
}

type Measure struct {
	Ordinal                      int
	Kind, Subject                string
	Place, ValidFrom, ValidUntil *string
	IndeterminateFields          []string
	Evidence                     []Evidence
	TemporalCandidates           []TemporalCandidate
}

type Result struct {
	RunID, DocumentVersionID, ClassificationRunID int64
	Status, ReasonCode, ContentSHA256             string
	ContentComplete                               bool
	ProviderResponseID, ReturnedModel             string
	InputTokens, OutputTokens, CacheReadTokens    *int64
	Measures                                      []Measure
	Segments                                      []SegmentResult
	CreatedAt                                     time.Time
}

type SegmentResult struct {
	Ordinal, Total                             int
	DocumentVersionID                          int64
	ResourceURL, Role                          string
	Page, StartByte, EndByte                   int
	CoreStartByte, CoreEndByte                 int
	SourceStartByte, SourceEndByte             int
	ContentSHA256, ResponseSHA256              string
	ProviderResponseID, ReturnedModel          string
	MeasureCount                               int
	InputTokens, OutputTokens, CacheReadTokens *int64
	NormalizationMap                           []WindowOffset
}

type Visibility struct {
	DocumentID, DocumentVersionID, RunID int64
	OfficialURL, Status, ReasonCode      string
	FirstAcquiredAt                      time.Time
	InterpretedAt                        *time.Time
}
