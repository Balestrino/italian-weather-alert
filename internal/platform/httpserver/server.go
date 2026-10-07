// Package httpserver contains HTTP utilities shared by service entry points.
package httpserver

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/health"
)

// HealthRoutes discloses dependency identities only on the private listener.
func HealthRoutes(mux *http.ServeMux, checks map[string]health.Check, private bool) {
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		Reply(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		states := Check(r.Context(), checks)
		code := http.StatusOK
		for _, ok := range states {
			if !ok {
				code = http.StatusServiceUnavailable
			}
		}
		if private {
			Reply(w, code, states)
			return
		}
		status := "ready"
		if code != http.StatusOK {
			status = "unavailable"
		}
		Reply(w, code, map[string]string{"status": status})
	})
}

func Check(ctx context.Context, checks map[string]health.Check) map[string]bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	states := make(map[string]bool, len(checks))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for name, check := range checks {
		wg.Add(1)
		go func() { defer wg.Done(); ok := check(ctx) == nil; mu.Lock(); states[name] = ok; mu.Unlock() }()
	}
	wg.Wait()
	return states
}

func Reply(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// Serve owns the listener and drains in-flight requests on cancellation.
func Serve(ctx context.Context, ln net.Listener, handler http.Handler) error {
	// A complete saved situation can exceed ten seconds while gathering retained
	// evidence. Give the response a bounded budget separate from request reads.
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 16}
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ln) }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = srv.Close()
			return err
		}
		return nil
	}
}
