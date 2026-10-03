package main

import (
	"strings"
	"testing"
)

func TestVerificationInputRejectsFactsAndExtraPayloads(t *testing.T) {
	for _, body := range []string{`{"candidate":{"fields":{"level":{"value":"green"}}}}`, `{} {}`, `{"unknown":true}`, strings.Repeat(" ", 1048577)} {
		if _, err := readVerificationRequest(strings.NewReader(body)); err == nil {
			t.Fatal("unsupported or unbounded verification input accepted")
		}
	}
	request, err := readVerificationRequest(strings.NewReader(`{"request_id":"fixture","candidate":{"source_id":"source","version_id":1,"fields":{"kind":{"resource_url":"https://municipal.example/act","json_pointer":"/kind","start_byte":0,"end_byte":8,"locator":"kind"}}}}`))
	if err != nil || request.Candidate.Fields["kind"].EndByte != 8 {
		t.Fatalf("valid selectors rejected: %v", err)
	}
}
