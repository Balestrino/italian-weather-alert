package inference

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
)

type gateFixture struct {
	rejected map[string]processing.Rejection
	held     bool
}

func (g *gateFixture) AcquireGate(_ context.Context, scope, model string, _ time.Time, _ processing.GatePolicy) (processing.GatePermit, error) {
	if g.held {
		return processing.GatePermit{}, &processing.GateBlocked{State: processing.GateState{State: "held", Reason: "quota"}}
	}
	return processing.GatePermit{Scope: scope, Model: model, Generation: 1, GlobalGeneration: 1}, nil
}
func (g *gateFixture) RecordGate(context.Context, processing.GatePermit, string, time.Duration, time.Time, processing.GatePolicy) error {
	return nil
}
func (g *gateFixture) Rejected(_ context.Context, scope, model, configuration, hash string) (processing.Rejection, bool, error) {
	r, ok := g.rejected[configuration+hash]
	return r, ok, nil
}
func (g *gateFixture) Reject(_ context.Context, r processing.Rejection) error {
	g.rejected[r.ConfigurationID+r.InputSHA256] = r
	return nil
}
func TestPermanentRequestRejectionSharedAcrossVersions(t *testing.T) {
	g := &gateFixture{rejected: map[string]processing.Rejection{}}
	calls := 0
	a := &GatedAdapter{Gate: Gate{Store: g, Scope: "fixture", Policy: processing.DefaultGatePolicy()}, Adapter: adapterFunc(func(context.Context, Request) (Response, error) {
		calls++
		return Response{}, &CallError{Code: "provider_rejected", StatusCode: 400, Diagnostic: Diagnostic{HTTPStatus: 400, Category: "request"}}
	})}
	r := Request{Model: "ocr", Messages: []Message{{Role: "user", Content: json.RawMessage(`"same image"`)}}}
	for _, run := range []int64{11, 22} {
		_, err := a.Complete(WithAttempt(context.Background(), processing.Attempt{RunID: run, Number: 1}, "config-v1"), r)
		if err == nil {
			t.Fatal("rejection hidden")
		}
	}
	if calls != 1 {
		t.Fatal("new version repeated rejected identical request")
	}
	_, _ = a.Complete(WithAttempt(context.Background(), processing.Attempt{RunID: 33, Number: 1}, "config-v2"), r)
	if calls != 2 {
		t.Fatal("corrected configuration cannot run")
	}
	g.held = true
	_, err := a.Complete(WithAttempt(context.Background(), processing.Attempt{RunID: 44, Number: 1}, "config-v3"), r)
	var deferred *jobs.DeferredError
	if !errors.As(err, &deferred) || calls != 2 {
		t.Fatal("held scope called provider")
	}
}

func TestEquivalentCopyForbidsProviderFallback(t *testing.T) {
	calls := 0
	a := &GatedAdapter{Adapter: adapterFunc(func(context.Context, Request) (Response, error) { calls++; return Response{}, nil })}
	ctx := WithAttempt(LocalReuseOnly(context.Background()), processing.Attempt{RunID: 1, Number: 1}, "cfg")
	_, err := a.Complete(ctx, Request{Model: "qwen", Messages: []Message{{Role: "user", Content: json.RawMessage(`"unchanged"`)}}})
	var call *CallError
	if !errors.As(err, &call) || call.Code != "equivalent_reuse_unavailable" || calls != 0 {
		t.Fatalf("paid fallback allowed: calls=%d err=%v", calls, err)
	}
	// No gate or embedder is configured: the local-only check must precede both.
	_, err = (&GatedEmbedder{}).Embed(ctx, "embed", []string{"unchanged"}, 10)
	if !errors.As(err, &call) || call.Code != "equivalent_reuse_unavailable" {
		t.Fatal(err)
	}
}

func TestRecoveryDeferralMakesNoTransportOrFailureReceipt(t *testing.T) {
	d := &jobs.DeferredError{Until: time.Now().Add(time.Minute), Code: "recovery_call_limit"}
	l := &ledgerFixture{startErr: d}
	calls := 0
	g := &gateFixture{rejected: map[string]processing.Rejection{}}
	a := &GatedAdapter{Gate: Gate{Store: g, Scope: "fixture", Policy: processing.DefaultGatePolicy()}, Adapter: &RecordedAdapter{Ledger: l, Adapter: adapterFunc(func(context.Context, Request) (Response, error) { calls++; return Response{}, nil })}}
	ctx := WithAttempt(context.Background(), processing.Attempt{RunID: 1, Number: 1}, "config")
	_, err := a.Complete(ctx, Request{Model: "fixture", Messages: []Message{{Role: "user", Content: json.RawMessage(`"hello"`)}}})
	var got *jobs.DeferredError
	if !errors.As(err, &got) || got.Code != d.Code || calls != 0 || len(l.finishes) != 0 {
		t.Fatalf("recovery deferral changed: %v", err)
	}
}
