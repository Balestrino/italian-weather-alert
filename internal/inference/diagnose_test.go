package inference

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDiagnosticBudgetAndRecoveryActions(t *testing.T) {
	for _, category := range []string{"authentication", "quota", "request"} {
		t.Run(category, func(t *testing.T) {
			calls := 0
			a := adapterFunc(func(context.Context, Request) (Response, error) {
				calls++
				return Response{}, &CallError{Code: "provider_rejected", Diagnostic: Diagnostic{HTTPStatus: 400, Category: category}}
			})
			request := Request{Model: "fixture", Messages: []Message{{Role: "user", Content: json.RawMessage(`"secret prompt"`)}}, MaxTokens: 16}
			cases := []DiagnosticCase{{Class: "first", Adapter: a, Request: request}, {Class: "second", Adapter: a, Request: request}}
			results, err := Diagnose(context.Background(), cases)
			if err != nil || len(results) == 0 || results[0].NextAction == "" {
				t.Fatal("missing corrective action")
			}
			want := 1
			if category == "request" {
				want = 2
			}
			if calls != want {
				t.Fatalf("wrong bounded calls: %d", calls)
			}
			calls = 0
			cases[1].Class = "first"
			if _, err = Diagnose(context.Background(), cases); err == nil || calls != 0 {
				t.Fatal("duplicate class consumed tokens")
			}
			if _, err = Diagnose(context.Background(), append(cases, cases...)); err == nil || calls != 0 {
				t.Fatal("more than three calls permitted")
			}
		})
	}
}
