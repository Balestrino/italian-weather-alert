package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
)

// reviewedOCRAdapter is opt-in, local evidence for a caller-selected reviewed
// fixture. It never writes HTTP headers, endpoints, credentials or error bodies.
// The destination is reserved before the paid request and cannot be overwritten.
type reviewedOCRAdapter struct {
	adapter inference.Adapter
	file    *os.File
}

func (a *reviewedOCRAdapter) Name() string { return a.adapter.Name() }
func (a *reviewedOCRAdapter) Complete(ctx context.Context, req inference.Request) (inference.Response, error) {
	response, err := a.adapter.Complete(ctx, req)
	if err != nil {
		return response, err
	}
	evidence := struct {
		Model        string          `json:"model"`
		Text         string          `json:"text"`
		FinishReason string          `json:"finish_reason"`
		Usage        inference.Usage `json:"usage"`
	}{response.Model, response.Content, response.FinishReason, response.Usage}
	if e := json.NewEncoder(a.file).Encode(evidence); e != nil {
		return response, errors.New("ocr_review_write_failed")
	}
	if e := a.file.Sync(); e != nil {
		return response, errors.New("ocr_review_write_failed")
	}
	return response, nil
}
func reserveOCRReview(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, errors.New("ocr_review_file_unavailable")
	}
	return f, nil
}
