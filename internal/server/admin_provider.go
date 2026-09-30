package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

type AdminProvider interface {
	Gates(context.Context) ([]processing.GateState, error)
	Rejections(context.Context) ([]processing.Rejection, error)
	ResumeGate(context.Context, string, string, string, time.Time) error
	ClearRejection(context.Context, string, string, string, string, string, time.Time) error
}

func adminProviderRoutes(mux *http.ServeMux, a AdminRuntime) {
	mux.HandleFunc("GET /admin/inference/gates", func(w http.ResponseWriter, r *http.Request) {
		if a.Provider == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		gates, err := a.Provider.Gates(r.Context())
		if err != nil {
			adminFailure(w, err)
			return
		}
		reply(w, 200, gates)
	})
	mux.HandleFunc("GET /admin/inference/rejections", func(w http.ResponseWriter, r *http.Request) {
		if a.Provider == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		rows, err := a.Provider.Rejections(r.Context())
		if err != nil {
			adminFailure(w, err)
			return
		}
		reply(w, 200, map[string]any{"latest_limit": 1000, "rejections": rows})
	})
	for _, action := range []string{"resume", "clear-rejection"} {
		mux.HandleFunc("POST /admin/inference/"+action, func(w http.ResponseWriter, r *http.Request) {
			var input struct {
				Scope         string `json:"scope"`
				Model         string `json:"model"`
				Actor         string `json:"actor"`
				Configuration string `json:"configuration"`
				InputSHA256   string `json:"input_sha256"`
			}
			if !adminDecode(w, r, &input) {
				return
			}
			if a.Provider == nil {
				adminFailure(w, registryUnavailable)
				return
			}
			var err error
			if action == "resume" {
				err = a.Provider.ResumeGate(r.Context(), input.Scope, input.Model, input.Actor, time.Now())
			} else {
				err = a.Provider.ClearRejection(r.Context(), input.Scope, input.Model, input.Configuration, input.InputSHA256, input.Actor, time.Now())
			}
			if errors.Is(err, processing.ErrInvalid) {
				err = registry.ErrInvalid
			}
			if err != nil {
				adminFailure(w, err)
				return
			}
			adminSuccess(w, r, map[string]string{"result": action})
		})
	}
}
