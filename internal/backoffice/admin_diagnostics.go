package backoffice

import (
	"net/http"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/diagnostics"
)

func adminDiagnosticRoutes(mux *http.ServeMux, runtime AdminRuntime) {
	mux.HandleFunc("GET /admin/diagnostics/sources/{id}", func(w http.ResponseWriter, r *http.Request) {
		if runtime.Diagnostics == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		result, err := runtime.Diagnostics.Snapshot(r.Context(), r.PathValue("id"), time.Now().UTC())
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, result)
	})
	mux.HandleFunc("POST /admin/diagnostics/dpc-comparisons", func(w http.ResponseWriter, r *http.Request) {
		if runtime.Diagnostics == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		var input diagnostics.Comparison
		if !adminDecode(w, r, &input) {
			return
		}
		result, err := runtime.Diagnostics.Compare(r.Context(), input)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, result)
	})
}
