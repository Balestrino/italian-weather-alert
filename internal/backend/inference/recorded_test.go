package inference

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
)

type ledgerFixture struct {
	starts              []processing.CallStart
	finishes            []processing.CallFinish
	startErr, finishErr error
}

func (l *ledgerFixture) StartCall(_ context.Context, s processing.CallStart) (int64, error) {
	l.starts = append(l.starts, s)
	return int64(len(l.starts)), l.startErr
}
func (l *ledgerFixture) FinishCall(ctx context.Context, f processing.CallFinish) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	l.finishes = append(l.finishes, f)
	return l.finishErr
}

type adapterFunc func(context.Context, Request) (Response, error)

func (a adapterFunc) Name() string                                              { return "fixture" }
func (a adapterFunc) Complete(ctx context.Context, r Request) (Response, error) { return a(ctx, r) }

func TestRecordedCallWritesUsageBeforeConsumerValidation(t *testing.T) {
	l := &ledgerFixture{}
	tokens := int64(21)
	ctx, cancel := context.WithCancel(WithAttempt(context.Background(), processing.Attempt{RunID: 1, Number: 2}))
	a := &RecordedAdapter{Ledger: l, Adapter: adapterFunc(func(context.Context, Request) (Response, error) {
		if len(l.starts) != 1 {
			t.Fatal("provider called without durable intent")
		}
		cancel()
		return Response{Content: "invalid structured output", Usage: Usage{InputTokens: &tokens}, Diagnostic: Diagnostic{HTTPStatus: 200}}, nil
	})}
	r := Request{Model: "fixture", Messages: []Message{{Role: "user", Content: json.RawMessage(`"confidential input"`)}}}
	response, err := a.Complete(ctx, r)
	if err != nil || response.Content == "" || len(l.finishes) != 1 || l.finishes[0].Usage.Status != "partial" || *l.finishes[0].Usage.InputTokens != 21 {
		t.Fatalf("receipt missing: %#v %v", l.finishes, err)
	}
	if len(l.starts[0].InputSHA256) != 64 || l.starts[0].AttemptNumber != 2 {
		t.Fatal("input identity missing")
	}
}

func TestRecordedCallFailsClosedWithoutIntentOrReceipt(t *testing.T) {
	for _, stage := range []string{"intent", "receipt", "context"} {
		t.Run(stage, func(t *testing.T) {
			l := &ledgerFixture{}
			calls := 0
			if stage == "intent" {
				l.startErr = errors.New("db unavailable")
			}
			if stage == "receipt" {
				l.finishErr = errors.New("db unavailable")
			}
			ctx := WithAttempt(context.Background(), processing.Attempt{RunID: 1, Number: 1})
			if stage == "context" {
				ctx = context.Background()
			}
			a := &RecordedAdapter{Ledger: l, Adapter: adapterFunc(func(context.Context, Request) (Response, error) { calls++; return Response{}, nil })}
			_, err := a.Complete(ctx, Request{Model: "fixture", Messages: []Message{{Role: "user", Content: json.RawMessage(`"hello"`)}}})
			if err == nil {
				t.Fatal("storage failure hidden")
			}
			if stage != "receipt" && calls != 0 {
				t.Fatal("called without durable context")
			}
		})
	}
}

func TestTransportFailureRemainsUncertain(t *testing.T) {
	l := &ledgerFixture{}
	ctx := WithAttempt(context.Background(), processing.Attempt{RunID: 1, Number: 1})
	a := &RecordedAdapter{Ledger: l, Adapter: adapterFunc(func(context.Context, Request) (Response, error) {
		return Response{}, &CallError{Code: "request_timeout", Temporary: true, RetryAfter: time.Second}
	})}
	_, err := a.Complete(ctx, Request{Model: "fixture", Messages: []Message{{Role: "user", Content: json.RawMessage(`"hello"`)}}})
	if err == nil || len(l.finishes) != 1 || l.finishes[0].State != "uncertain" || l.finishes[0].Usage.Status != "unavailable" {
		t.Fatal("uncertain call recorded as free")
	}
}
