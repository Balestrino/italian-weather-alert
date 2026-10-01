package main

import (
	"net/http"

	"github.com/Balestrino/italian-weather-alert/internal/backend/health"
	"github.com/Balestrino/italian-weather-alert/internal/backend/transport/httpapi"
	"github.com/Balestrino/italian-weather-alert/internal/backoffice"
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
)

func listenerHandler(role string, checks map[string]health.Check, queries httpapi.PublicQueries, public httpapi.PublicRuntime, admin backoffice.AdminRuntime) http.Handler {
	switch role {
	case "public":
		return httpapi.HandlerWithPublicRuntime(checks, queries, public)
	case "admin":
		return backoffice.HandlerWithAdministration(checks, admin)
	default:
		// Preserve the health-only listener for the worker role without a subcommand.
		mux := http.NewServeMux()
		httpserver.HealthRoutes(mux, checks, false)
		return mux
	}
}
