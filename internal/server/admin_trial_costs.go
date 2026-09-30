package server

import (
	"net/http"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/registry"
	"github.com/Balestrino/italian-weather-alert/internal/trialcost"
)

func adminTrialCostRoutes(mux *http.ServeMux, runtime AdminRuntime) {
	mux.HandleFunc("POST /admin/observation-campaigns/{id}/cost-inputs", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Actor string          `json:"actor"`
			Input trialcost.Input `json:"input"`
		}
		if !adminDecode(w, r, &request) {
			return
		}
		if runtime.TrialCosts == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		result, err := runtime.TrialCosts.RecordInput(r.Context(), r.PathValue("id"), request.Actor, request.Input)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, result)
	})
	mux.HandleFunc("GET /admin/observation-campaigns/{id}/cost-report", func(w http.ResponseWriter, r *http.Request) {
		if runtime.TrialCosts == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		through, ok := trialCostThrough(w, r)
		if !ok {
			return
		}
		result, err := runtime.TrialCosts.Report(r.Context(), r.PathValue("id"), through)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, result)
	})
	mux.HandleFunc("POST /admin/observation-campaigns/{id}/budget-proposals", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Actor    string            `json:"actor"`
			Through  time.Time         `json:"through"`
			Evidence registry.Evidence `json:"evidence"`
		}
		if !adminDecode(w, r, &request) {
			return
		}
		if runtime.TrialCosts == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		result, err := runtime.TrialCosts.Propose(r.Context(), r.PathValue("id"), request.Actor, request.Through, request.Evidence)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, result)
	})
	mux.HandleFunc("GET /admin/observation-campaigns/{id}/budget-proposals", func(w http.ResponseWriter, r *http.Request) {
		if runtime.TrialCosts == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		result, err := runtime.TrialCosts.Proposals(r.Context(), r.PathValue("id"))
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, result)
	})
}

func trialCostThrough(w http.ResponseWriter, r *http.Request) (time.Time, bool) {
	values, present := r.URL.Query()["through"]
	if !present || len(values) != 1 {
		adminFailure(w, trialcost.ErrInvalid)
		return time.Time{}, false
	}
	through, err := time.Parse(time.RFC3339Nano, values[0])
	if err != nil {
		adminFailure(w, trialcost.ErrInvalid)
		return time.Time{}, false
	}
	return through, true
}
