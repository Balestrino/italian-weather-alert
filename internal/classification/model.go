// Package classification determines whether complete retained content is in
// the service's weather/hydro scope. It does not extract measures.
package classification

import (
	"errors"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid classification request")
	ErrConflict = errors.New("classification result conflict")
)

const Kind = "classify_relevance"

type Result struct {
	RunID, DocumentVersionID                   int64
	Status                                     string
	Relevant                                   *bool
	ReasonCode, EvidenceQuote, ContentSHA256   string
	ContentComplete                            bool
	ProviderResponseID, ReturnedModel          string
	InputTokens, OutputTokens, CacheReadTokens *int64
	CreatedAt                                  time.Time
	Segments                                   []SegmentResult
}

type SegmentResult struct {
	Ordinal, Total                             int
	DocumentVersionID                          int64
	ResourceURL, Role                          string
	Page, StartByte, EndByte                   int
	ContentSHA256, ResponseSHA256              string
	Relevant                                   bool
	ReasonCode, EvidenceQuote                  string
	ProviderResponseID, ReturnedModel          string
	InputTokens, OutputTokens, CacheReadTokens *int64
}
