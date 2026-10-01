package main

import (
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"testing"
)

func TestOnlyExactHistoricalRequestCanBeReplayed(t *testing.T) {
	s := classification.Segment{DocumentVersionID: 3419, Ordinal: 1, Total: 1, ResourceURL: "https://example.org", Role: "original", Text: "test", StartByte: 0, EndByte: 4}
	r, e := classification.SegmentRequest("fixture", s)
	if e != nil {
		t.Fatal(e)
	}
	r.ChatTemplateKwargs = &inference.ChatTemplateOptions{EnableThinking: false}
	expected := hashRequest(r)
	if _, e = matchedRequest("fixture", s, expected); e != nil {
		t.Fatal(e)
	}
	s.Text = "changed"
	s.EndByte = len(s.Text)
	if _, e = matchedRequest("fixture", s, expected); e == nil {
		t.Fatal("changed input accepted")
	}
	s.Text = "test"
	s.EndByte = 4
	if _, e = matchedRequest("other-model", s, expected); e == nil {
		t.Fatal("changed model accepted")
	}
}
