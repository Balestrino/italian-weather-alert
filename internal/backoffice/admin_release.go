package backoffice

import (
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
	"net/http"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func adminReleaseRoutes(mux *http.ServeMux, runtime AdminRuntime) {
	mux.HandleFunc("GET /admin/release-readiness", func(w http.ResponseWriter, r *http.Request) {
		if runtime.Registry == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		report, err := runtime.Registry.ReleaseReadiness(r.Context())
		if err != nil {
			adminFailure(w, err)
			return
		}
		if wantsAdminJSON(r) {
			httpserver.Reply(w, http.StatusOK, report)
			return
		}
		adminRender(w, "release", report)
	})
	mux.HandleFunc("GET /admin/sources/{id}/acceptance-reviews", func(w http.ResponseWriter, r *http.Request) {
		if runtime.Registry == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		reviews, err := runtime.Registry.AcceptanceReviews(r.Context(), r.PathValue("id"))
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, reviews)
	})
	mux.HandleFunc("POST /admin/sources/{id}/review-acceptance", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Revision int               `json:"revision"`
			Actor    string            `json:"actor"`
			Evidence registry.Evidence `json:"evidence"`
		}
		if !adminDecode(w, r, &input) {
			return
		}
		if runtime.Registry == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		review, err := runtime.Registry.ReviewAcceptance(r.Context(), r.PathValue("id"), input.Revision, input.Actor, input.Evidence)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, review)
	})
}
