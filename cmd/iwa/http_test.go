package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/publicview"
	"github.com/Balestrino/italian-weather-alert/internal/backend/transport/httpapi"
	"github.com/Balestrino/italian-weather-alert/internal/backoffice"
)

func TestListenerRoleSelectsOnlyItsInterface(t *testing.T) {
	for _, role := range []string{"public", "admin", "worker"} {
		t.Run(role, func(t *testing.T) {
			var public httpapi.PublicRuntime
			if role != "public" {
				// Private roles have no public cursor key. Constructing the public
				// router first would panic even though it would later be replaced.
				public.Views = publicview.New(nil)
			}
			h := listenerHandler(role, nil, nil, public, backoffice.AdminRuntime{})
			for _, path := range []string{"/health/ready", "/openapi.json", "/admin/status"} {
				want := http.StatusNotFound
				if path == "/health/ready" || (path == "/openapi.json" && role == "public") || (path == "/admin/status" && role == "admin") {
					want = http.StatusOK
				}
				w := httptest.NewRecorder()
				h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1:8081"+path, nil))
				if w.Code != want {
					t.Errorf("%s returned %d, want %d", path, w.Code, want)
				}
			}
		})
	}
}
