package embedding

import (
	"errors"
	"time"
)

var ErrInvalid = errors.New("invalid embedding")

const Kind = "embed_measure"

type Result struct {
	RunID, ExtractionRunID                int64
	MeasureOrdinal, Dimensions            int
	ConfigurationVersionID, ReturnedModel string
	Vector                                []float32
	InputTokens                           *int64
	CreatedAt                             time.Time
}
