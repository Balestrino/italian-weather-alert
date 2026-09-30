package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/health"
)

// Handler has separate route tables. No admin route is registered publicly.
func Handler(role string, checks map[string]health.Check, queries ...PublicQueries) http.Handler {
	return handler(role, checks, time.Now, queries...)
}

func handler(role string, checks map[string]health.Check, now func() time.Time, queries ...PublicQueries) http.Handler {
	return handlerWithLimits(role, checks, now, PublicLimits{Allowance: 120, Window: time.Minute, MaxPageSize: 100}, queries...)
}

// HandlerWithPublicLimits builds the role-specific server with configured
// anonymous limits. The limits are ignored when the role is not public.
func HandlerWithPublicLimits(role string, checks map[string]health.Check, queries PublicQueries, limits PublicLimits) http.Handler {
	return handlerWithLimits(role, checks, time.Now, limits, queries)
}

// HandlerWithPublicRuntime enables persistent views in addition to request
// limits. Production public listeners use this constructor.
func HandlerWithPublicRuntime(role string, checks map[string]health.Check, queries PublicQueries, runtime PublicRuntime) http.Handler {
	return handlerWithRuntime(role, checks, time.Now, runtime, queries)
}

// HandlerWithAdministration is only constructed for the local admin listener.
func HandlerWithAdministration(checks map[string]health.Check, runtime AdminRuntime) http.Handler {
	mux := http.NewServeMux()
	adminSourceRoutes(mux, runtime)
	adminOverviewRoutes(mux, runtime)
	adminTerritoryRoutes(mux, runtime)
	adminAlertRoutes(mux, runtime)
	mux.Handle("/", runtimeMux("admin", checks, time.Now, PublicRuntime{}))
	return adminBoundaryWithOrigin(adminTerritorialContext(mux, runtime), runtime.TailscaleOrigin)
}

func handlerWithLimits(role string, checks map[string]health.Check, now func() time.Time, limits PublicLimits, queries ...PublicQueries) http.Handler {
	return handlerWithRuntime(role, checks, now, PublicRuntime{Limits: limits}, queries...)
}

func handlerWithRuntime(role string, checks map[string]health.Check, now func() time.Time, runtime PublicRuntime, queries ...PublicQueries) http.Handler {
	mux := runtimeMux(role, checks, now, runtime, queries...)
	if role == "admin" {
		return adminBoundary(mux)
	}
	return mux
}

func runtimeMux(role string, checks map[string]health.Check, now func() time.Time, runtime PublicRuntime, queries ...PublicQueries) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		reply(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		states := Check(r.Context(), checks)
		code := http.StatusOK
		for _, ok := range states {
			if !ok {
				code = http.StatusServiceUnavailable
			}
		}
		if role == "admin" {
			reply(w, code, states)
			return
		}
		status := "ready"
		if code != http.StatusOK {
			status = "unavailable"
		}
		reply(w, code, map[string]string{"status": status})
	})
	if role == "public" {
		publicDocumentationRoutes(mux)
	}
	if role == "admin" {
		adminRoutes(mux)
	} else if role == "public" && len(queries) == 1 && queries[0] != nil {
		publicRoutes(mux, queries[0], now, runtime)
	}
	return mux
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
func reply(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// Serve owns the listener and drains in-flight requests on cancellation.
func Serve(ctx context.Context, ln net.Listener, handler http.Handler) error {
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 16}
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
