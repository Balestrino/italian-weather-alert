// Package backoffice serves the private administrative HTML and JSON interface.
package backoffice

import (
	"net/http"

	"github.com/Balestrino/italian-weather-alert/internal/backend/health"
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
)

func Handler(checks map[string]health.Check) http.Handler {
	mux := http.NewServeMux()
	httpserver.HealthRoutes(mux, checks, true)
	adminRoutes(mux)
	return adminBoundary(mux)
}

// HandlerWithAdministration is constructed only for the private admin listener.
func HandlerWithAdministration(checks map[string]health.Check, runtime AdminRuntime) http.Handler {
	mux := http.NewServeMux()
	adminSourceRoutes(mux, runtime)
	adminOverviewRoutes(mux, runtime)
	adminTerritoryRoutes(mux, runtime)
	adminAlertRoutes(mux, runtime)
	mux.Handle("/", adminMux(checks))
	return adminBoundaryWithOrigin(adminTerritorialContext(mux, runtime), runtime.TailscaleOrigin)
}

func adminMux(checks map[string]health.Check) http.Handler {
	mux := http.NewServeMux()
	httpserver.HealthRoutes(mux, checks, true)
	adminRoutes(mux)
	return mux
}
