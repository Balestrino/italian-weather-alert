package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/config"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
)

func main() {
	if err := run(); err != nil {
		// Errors are deliberately stable codes; never print provider bodies,
		// prompts, endpoints or credentials.
		fmt.Fprintln(os.Stderr, "inference probe failed:", safeCode(err))
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) == 3 && os.Args[1] == "diagnose" {
		return diagnose(os.Args[2])
	}
	if len(os.Args) < 2 || (os.Args[1] != "qwen" && os.Args[1] != "ocr" && os.Args[1] != "classification") {
		return errors.New("usage")
	}
	if ((os.Args[1] == "qwen" || os.Args[1] == "classification") && len(os.Args) != 2) || (os.Args[1] == "ocr" && len(os.Args) != 3 && len(os.Args) != 4) {
		return errors.New("usage")
	}
	var endpoint config.InferenceEndpoint
	var err error
	endpoint, err = config.LoadQwen()
	if os.Args[1] == "ocr" {
		endpoint, err = config.LoadOCR()
	}
	if err != nil {
		return err
	}
	if endpoint.Adapter != "openai-chat" {
		return errors.New("unsupported_adapter")
	}
	adapter, err := inference.NewOpenAIChat(os.Args[1], endpoint.URL, endpoint.APIKey, &http.Client{Timeout: 90 * time.Second})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if os.Args[1] == "qwen" {
		result, probeErr := inference.ProbeStructuredOutput(ctx, adapter, endpoint.Model, time.Now())
		if probeErr != nil {
			return probeErr
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	if os.Args[1] == "classification" {
		result, probeErr := classification.ProbeReviewedCases(ctx, adapter, endpoint.Model, time.Now())
		if probeErr != nil {
			return probeErr
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	image, err := os.ReadFile(os.Args[2])
	if err != nil {
		return errors.New("image_read_failed")
	}
	checks := []inference.OCRCheck{
		{Name: "order_heading", Text: "ORDINA"},
		{Name: "start_date", Text: "20 agosto 2026"},
		{Name: "start_time", Text: "18.00"},
		{Name: "conditional_end", Text: "fino al perdurare dell’emergenza"},
		{Name: "dog_area", Text: "area di sgambamento cani"},
		{Name: "cemeteries", Text: "cimiteri comunali"},
		{Name: "river_cycle_paths", Text: "ciclopiste fluviali"},
		{Name: "fornacette", Text: "Fornacette"},
	}
	var ocrAdapter inference.Adapter = adapter
	if len(os.Args) == 4 {
		review, err := reserveOCRReview(os.Args[3])
		if err != nil {
			return err
		}
		defer review.Close()
		ocrAdapter = &reviewedOCRAdapter{adapter: adapter, file: review}
	}
	result, err := inference.ProbeOCR(ctx, ocrAdapter, endpoint.Model, "image/png", image, checks, time.Now())
	if err != nil {
		return err
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		return err
	}
	if len(result.MissingChecks) > 0 {
		return errors.New("ocr_content_mismatch")
	}
	return nil
}

func safeCode(err error) string {
	var call *inference.CallError
	if errors.As(err, &call) {
		return call.Code
	}
	if errors.Is(err, inference.ErrInvalid) {
		return "invalid_configuration"
	}
	if err.Error() == "usage" || err.Error() == "unsupported_adapter" || err.Error() == "image_read_failed" || err.Error() == "ocr_content_mismatch" || err.Error() == "classification_mismatch" || err.Error() == "ocr_review_file_unavailable" || err.Error() == "ocr_review_write_failed" {
		return err.Error()
	}
	return "configuration_error"
}
