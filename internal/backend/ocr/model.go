// Package ocr extracts page text from retained image/PDF evidence without
// treating model output as verified interpretation.
package ocr

import (
	"errors"
	"time"
)

var (
	ErrInvalid             = errors.New("invalid OCR request")
	ErrConflict            = errors.New("OCR evidence conflict")
	ErrRasterization       = errors.New("PDF rasterization failed")
	ErrRendererUnavailable = errors.New("PDF renderer unavailable")
)

const (
	Queue = "inference"
	Kind  = "ocr_resource"
)

type PageImage struct {
	Number    int
	MediaType string
	Bytes     []byte
}

type PageResult struct {
	RunID, DocumentVersionID                   int64
	PageNumber                                 int
	ResourceURL, Status, MediaType             string
	InputSHA256, OutputSHA256, ExtractedText   string
	ProviderResponseID, ReturnedModel          string
	InputTokens, OutputTokens, CacheReadTokens *int64
	CreatedAt                                  time.Time
}

type ResourceResult struct {
	RunID, DocumentVersionID int64
	ResourceURL, Status      string
	PageCount                int
	ErrorCode                string
	CreatedAt                time.Time
}
