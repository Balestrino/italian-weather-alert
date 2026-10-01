package main

import (
	"context"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type reviewFixture struct{ calls int }

func (a *reviewFixture) Name() string { return "fixture" }
func (a *reviewFixture) Complete(context.Context, inference.Request) (inference.Response, error) {
	a.calls++
	return inference.Response{Model: "ocr", Content: "nelle ciclopiste fluviali", FinishReason: "stop", ID: "private-request-id", ReasoningContent: "not needed"}, nil
}
func TestOCRReviewIsPrivateAndCannotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review.json")
	f, err := reserveOCRReview(path)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &reviewFixture{}
	a := &reviewedOCRAdapter{adapter: fixture, file: f}
	if _, err = a.Complete(context.Background(), inference.Request{}); err != nil {
		t.Fatal(err)
	}
	f.Close()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if json.Unmarshal(b, &parsed) != nil || parsed["text"] != "nelle ciclopiste fluviali" || parsed["finish_reason"] != "stop" || strings.Contains(string(b), "private-request-id") {
		t.Fatal("incorrect capture")
	}
	st, _ := os.Stat(path)
	if st.Mode().Perm() != 0600 {
		t.Fatal("review exposed")
	}
	if _, err = reserveOCRReview(path); err == nil {
		t.Fatal("capture overwritten")
	}
	if fixture.calls != 1 {
		t.Fatal("unexpected retry")
	}
}
