// Package httpapi exposes the read-only public JSON API and MCP transports.
package httpapi

import (
	"net/http"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/health"
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
)

func Handler(checks map[string]health.Check, queries ...PublicQueries) http.Handler {
	return handler(checks, time.Now, queries...)
}

func handler(checks map[string]health.Check, now func() time.Time, queries ...PublicQueries) http.Handler {
	return handlerWithLimits(checks, now, PublicLimits{Allowance: 120, Window: time.Minute, MaxPageSize: 100}, queries...)
}

func HandlerWithPublicLimits(checks map[string]health.Check, queries PublicQueries, limits PublicLimits) http.Handler {
	return handlerWithLimits(checks, time.Now, limits, queries)
}

func HandlerWithPublicRuntime(checks map[string]health.Check, queries PublicQueries, runtime PublicRuntime) http.Handler {
	return handlerWithRuntime(checks, time.Now, runtime, queries)
}

func handlerWithLimits(checks map[string]health.Check, now func() time.Time, limits PublicLimits, queries ...PublicQueries) http.Handler {
	return handlerWithRuntime(checks, now, PublicRuntime{Limits: limits}, queries...)
}

func handlerWithRuntime(checks map[string]health.Check, now func() time.Time, runtime PublicRuntime, queries ...PublicQueries) http.Handler {
	mux := http.NewServeMux()
	httpserver.HealthRoutes(mux, checks, false)
	publicDocumentationRoutes(mux)
	if len(queries) == 1 && queries[0] != nil {
		publicRoutes(mux, queries[0], now, runtime)
	}
	return mux
}
