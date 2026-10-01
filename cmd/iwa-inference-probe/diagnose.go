package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/config"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
)

// The input is a local reviewed request list. It is never echoed to stdout.
func diagnose(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("diagnostic_input_unavailable")
	}
	defer f.Close()
	var input []struct {
		Class    string            `json:"class"`
		Endpoint string            `json:"endpoint"`
		Request  inference.Request `json:"request"`
	}
	decoder := json.NewDecoder(io.LimitReader(f, 24<<20))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || len(input) < 1 || len(input) > 3 {
		return inference.ErrInvalid
	}
	if err = decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return inference.ErrInvalid
	}
	var cases []inference.DiagnosticCase
	for _, item := range input {
		var endpoint config.InferenceEndpoint
		switch item.Endpoint {
		case "ocr":
			endpoint, err = config.LoadOCR()
		case "qwen":
			endpoint, err = config.LoadQwen()
		default:
			return inference.ErrInvalid
		}
		if err != nil {
			return err
		}
		if endpoint.Adapter != "openai-chat" {
			return errors.New("unsupported_adapter")
		}
		adapter, e := inference.NewOpenAIChat(item.Endpoint, endpoint.URL, endpoint.APIKey, &http.Client{Timeout: 90 * time.Second})
		if e != nil {
			return e
		}
		item.Request.Model = endpoint.Model
		cases = append(cases, inference.DiagnosticCase{Class: item.Class, Adapter: adapter, Request: item.Request})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	results, err := inference.Diagnose(ctx, cases)
	if encodeErr := json.NewEncoder(os.Stdout).Encode(results); encodeErr != nil {
		return encodeErr
	}
	return err
}
