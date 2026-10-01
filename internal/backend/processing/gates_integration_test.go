//go:build integration

package processing

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestProviderGatesPersistAndIsolateRecovery(t *testing.T) {
	ctx := context.Background()
	pool := processingTestDB(t)
	// Gates have no foreign-key dependency on the legacy catalog.
	if _, err := pool.Exec(ctx, gatesSchema); err != nil {
		t.Fatal(err)
	}
	s := New(pool)
	now := time.Now().UTC()
	policy := DefaultGatePolicy()
	for i := 0; i < 3; i++ {
		p, err := s.AcquireGate(ctx, "account", "ocr", now, policy)
		if err != nil {
			t.Fatal(err)
		}
		if err = s.RecordGate(ctx, p, "availability", 0, now, policy); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := New(pool).AcquireGate(ctx, "account", "ocr", now.Add(59*time.Second), policy); err == nil {
		t.Fatal("gate lost across store recreation")
	}
	if _, err := s.AcquireGate(ctx, "account", "qwen", now, policy); err != nil {
		t.Fatal("healthy model blocked")
	}
	var wg sync.WaitGroup
	permits := make(chan GatePermit, 2)
	failures := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := s.AcquireGate(ctx, "account", "ocr", now.Add(time.Minute), policy)
			if err == nil {
				permits <- p
			} else {
				failures <- err
			}
		}()
	}
	wg.Wait()
	if len(permits) != 1 || len(failures) != 1 {
		t.Fatal("multiple half-open probes")
	}
	p := <-permits
	if err := s.RecordGate(ctx, p, "rate_limit", 30*time.Minute, now.Add(time.Minute), policy); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AcquireGate(ctx, "account", "ocr", now.Add(20*time.Minute), policy); err == nil {
		t.Fatal("Retry-After shortened by cooldown cap")
	}
	p, err := s.AcquireGate(ctx, "account", "qwen", now, policy)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.RecordGate(ctx, p, "authentication", 0, now, policy); err != nil {
		t.Fatal(err)
	}
	_, err = s.AcquireGate(ctx, "account", "other", now.Add(time.Hour), policy)
	var blocked *GateBlocked
	if !errors.As(err, &blocked) || blocked.State.Reason != "authentication" {
		t.Fatal("account-wide hold missing")
	}
	if err = s.ResumeGate(ctx, "account", "*", "operator", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	// A late old response must not revoke explicit recovery.
	if err = s.RecordGate(ctx, p, "authentication", 0, now.Add(time.Hour), policy); err != nil {
		t.Fatal(err)
	}
	probe, err := s.AcquireGate(ctx, "account", "qwen", now.Add(time.Hour), policy)
	if err != nil {
		t.Fatal(err)
	}
	if probe.Token == "" {
		t.Fatal("resume released unrestricted work")
	}
	if err = s.RecordGate(ctx, probe, "", 0, now.Add(time.Hour), policy); err != nil {
		t.Fatal(err)
	}
	if _, err = s.AcquireGate(ctx, "account", "qwen", now.Add(time.Hour), policy); err != nil {
		t.Fatal("successful probe did not close gate")
	}
}

func TestPermanentInputErrorDoesNotTripHealthyGate(t *testing.T) {
	ctx := context.Background()
	pool := processingTestDB(t)
	if _, err := pool.Exec(ctx, gatesSchema); err != nil {
		t.Fatal(err)
	}
	s := New(pool)
	now := time.Now().UTC()
	policy := DefaultGatePolicy()
	for i := 0; i < 5; i++ {
		p, err := s.AcquireGate(ctx, "account", "ocr", now, policy)
		if err != nil {
			t.Fatal(err)
		}
		if err = s.RecordGate(ctx, p, "request", 0, now, policy); err != nil {
			t.Fatal(err)
		}
	}
}
