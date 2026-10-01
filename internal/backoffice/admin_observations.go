package backoffice

import (
	"net/http"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/observation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func adminObservationRoutes(mux *http.ServeMux, runtime AdminRuntime) {
	mux.HandleFunc("GET /admin/observation-campaigns", func(w http.ResponseWriter, r *http.Request) {
		if runtime.Observations == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		campaigns, err := runtime.Observations.List(r.Context())
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, campaigns)
	})
	mux.HandleFunc("POST /admin/observation-campaigns", func(w http.ResponseWriter, r *http.Request) {
		var request observation.StartRequest
		if !adminDecode(w, r, &request) {
			return
		}
		if runtime.Observations == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		campaign, err := runtime.Observations.Start(r.Context(), request)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, campaign)
	})
	mux.HandleFunc("GET /admin/observation-campaigns/{id}", func(w http.ResponseWriter, r *http.Request) {
		if runtime.Observations == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		campaign, err := runtime.Observations.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, campaign)
	})
	mux.HandleFunc("GET /admin/observation-campaigns/{id}/report", func(w http.ResponseWriter, r *http.Request) {
		if runtime.Observations == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		through := time.Now().UTC()
		if value := r.URL.Query().Get("through"); value != "" {
			parsed, err := time.Parse(time.RFC3339Nano, value)
			if err != nil {
				adminFailure(w, observation.ErrInvalid)
				return
			}
			through = parsed
		}
		report, err := runtime.Observations.Report(r.Context(), r.PathValue("id"), through)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, report)
	})
	mux.HandleFunc("POST /admin/observation-campaigns/{id}/reviews", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Actor  string             `json:"actor"`
			Review observation.Review `json:"review"`
		}
		if !adminDecode(w, r, &input) {
			return
		}
		if runtime.Observations == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		review, err := runtime.Observations.RecordReview(r.Context(), r.PathValue("id"), input.Actor, input.Review)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, review)
	})
	mux.HandleFunc("POST /admin/observation-campaigns/{id}/assess", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Actor    string            `json:"actor"`
			Through  time.Time         `json:"through"`
			Evidence registry.Evidence `json:"evidence"`
		}
		if !adminDecode(w, r, &input) {
			return
		}
		if runtime.Observations == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		assessment, err := runtime.Observations.Assess(r.Context(), r.PathValue("id"), input.Actor, input.Through, input.Evidence)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, assessment)
	})
}
