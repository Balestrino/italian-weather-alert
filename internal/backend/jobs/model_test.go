package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestCanonicalObject(t *testing.T) {
	a, ah, err := canonicalObject(json.RawMessage(`{"b":2,"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	b, bh, err := canonicalObject(json.RawMessage(" { \"a\" : 1, \"b\" : 2 } "))
	if err != nil || string(a) != string(b) || ah != bh {
		t.Fatal("equivalent payloads were not canonicalized")
	}
	for _, raw := range []string{"null", "[]", `{"a":1} trailing`, ""} {
		if _, _, err = canonicalObject(json.RawMessage(raw)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid payload accepted: %q", raw)
		}
	}
}

func TestSafeHandlerPanic(t *testing.T) {
	_, err := safeHandle(context.Background(), func(context.Context, Job) (Result, error) {
		panic("private detail")
	}, Job{})
	var handlerErr *HandlerError
	if !errors.As(err, &handlerErr) || handlerErr.Code != "handler_panic" || handlerErr.Detail != "handler panicked" {
		t.Fatal("panic was not converted to a safe failure")
	}
}
