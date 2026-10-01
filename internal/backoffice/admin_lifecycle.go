package backoffice

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/backend/evaluation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"net/http"
	"strings"
	"time"
)

func adminLifecycleRoutes(mux *http.ServeMux, a AdminRuntime) {
	adminRegressionRoutes(mux, a)
	adminProcessingEvaluationRoutes(mux, a)
	for _, action := range []string{"accept", "enable-public", "suspend-public", "suspend-interpretation", "resume-interpretation", "reprocess-interpretation"} {
		mux.HandleFunc("POST /admin/sources/{id}/"+action, func(w http.ResponseWriter, r *http.Request) {
			// Each operation has a closed schema; no endpoint accepts edited measures.
			var revision int
			var actor string
			var acceptance registry.Acceptance
			var defect registry.Evidence
			var recovery interpretation.Recovery
			var from, through *time.Time
			var errorCode, processingEvaluationID string
			switch action {
			case "accept":
				var v struct {
					Revision   int                 `json:"revision"`
					Actor      string              `json:"actor"`
					Acceptance registry.Acceptance `json:"acceptance"`
				}
				if !adminDecode(w, r, &v) {
					return
				}
				revision, actor, acceptance = v.Revision, v.Actor, v.Acceptance
			case "suspend-interpretation":
				var v struct {
					Revision int               `json:"revision"`
					Actor    string            `json:"actor"`
					Defect   registry.Evidence `json:"defect"`
				}
				if !adminDecode(w, r, &v) {
					return
				}
				revision, actor, defect = v.Revision, v.Actor, v.Defect
			case "resume-interpretation":
				var v struct {
					Revision int                     `json:"revision"`
					Actor    string                  `json:"actor"`
					Recovery interpretation.Recovery `json:"recovery"`
				}
				if !adminDecode(w, r, &v) {
					return
				}
				revision, actor, recovery = v.Revision, v.Actor, v.Recovery
			case "reprocess-interpretation":
				var v struct {
					Revision               int        `json:"revision"`
					Actor                  string     `json:"actor"`
					From                   *time.Time `json:"from"`
					Through                *time.Time `json:"through"`
					ErrorCode              string     `json:"error_code"`
					ProcessingEvaluationID string     `json:"processing_evaluation_id"`
				}
				if !adminDecode(w, r, &v) {
					return
				}
				revision, actor, from, through, errorCode, processingEvaluationID = v.Revision, v.Actor, v.From, v.Through, v.ErrorCode, v.ProcessingEvaluationID
			default:
				var v struct {
					Revision int    `json:"revision"`
					Actor    string `json:"actor"`
				}
				if !adminDecode(w, r, &v) {
					return
				}
				revision, actor = v.Revision, v.Actor
			}
			if revision < 1 || strings.TrimSpace(actor) == "" {
				adminFailure(w, registry.ErrInvalid)
				return
			}
			if a.Registry == nil {
				adminFailure(w, registryUnavailable)
				return
			}
			id := r.PathValue("id")
			var err error
			switch action {
			case "accept":
				err = a.Registry.Accept(r.Context(), id, revision, actor, acceptance)
			case "enable-public":
				err = a.Registry.EnablePublic(r.Context(), id, revision, actor)
			case "suspend-public":
				err = a.Registry.Disable(r.Context(), id, revision, actor, true)
			case "suspend-interpretation":
				err = a.Registry.SuspendInterpretation(r.Context(), id, revision, actor, defect)
			case "resume-interpretation":
				if a.Interpretation == nil {
					adminFailure(w, registryUnavailable)
					return
				}
				err = a.Interpretation.ResumeInterpretation(r.Context(), id, revision, actor, recovery)
			case "reprocess-interpretation":
				if a.Interpretation == nil {
					adminFailure(w, registryUnavailable)
					return
				}
				st, e := a.Registry.State(r.Context(), id)
				if e != nil {
					adminFailure(w, e)
					return
				}
				if st.ActiveRevision == nil || *st.ActiveRevision != revision {
					adminFailure(w, registry.ErrConflict)
					return
				}
				var request string
				var count int
				request, count, err = a.Interpretation.Reprocess(r.Context(), interpretation.Selection{SourceID: id, Actor: actor, From: from, Through: through, ErrorCode: errorCode, ProcessingEvaluationID: processingEvaluationID, At: time.Now().UTC()})
				if err == nil {
					adminSuccess(w, r, map[string]any{"reprocessing_id": request, "selected_versions": count})
					return
				}
			}
			if err != nil {
				adminFailure(w, err)
				return
			}
			adminSuccess(w, r, map[string]string{"result": action})
		})
	}
}

func adminProcessingEvaluationRoutes(mux *http.ServeMux, a AdminRuntime) {
	mux.HandleFunc("GET /admin/processing-evaluations", func(w http.ResponseWriter, r *http.Request) {
		if a.Evaluations == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		reports, err := a.Evaluations.List(r.Context())
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, reports)
	})
	mux.HandleFunc("POST /admin/processing-evaluations", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Actor      string                          `json:"actor"`
			Comparison evaluation.ProcessingComparison `json:"comparison"`
		}
		if !adminDecode(w, r, &input) {
			return
		}
		if a.Evaluations == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		report, err := a.Evaluations.RecordComparison(r.Context(), input.Actor, input.Comparison)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, report)
	})
}

type adminRegressions interface {
	RecordRegression(context.Context, string, int, string, string, evaluation.Run) error
	Regressions(context.Context, string) ([]registry.Regression, error)
}

func adminRegressionRoutes(mux *http.ServeMux, a AdminRuntime) {
	mux.HandleFunc("GET /admin/sources/{id}/regressions", func(w http.ResponseWriter, r *http.Request) {
		store, ok := a.Registry.(adminRegressions)
		if !ok {
			adminFailure(w, registryUnavailable)
			return
		}
		reports, err := store.Regressions(r.Context(), r.PathValue("id"))
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, reports)
	})
	mux.HandleFunc("POST /admin/sources/{id}/regressions", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Revision   int            `json:"revision"`
			Actor      string         `json:"actor"`
			PreviousID string         `json:"previous_id"`
			Run        evaluation.Run `json:"run"`
		}
		if !adminDecode(w, r, &input) {
			return
		}
		store, ok := a.Registry.(adminRegressions)
		if !ok {
			adminFailure(w, registryUnavailable)
			return
		}
		if err := store.RecordRegression(r.Context(), r.PathValue("id"), input.Revision, input.Actor, input.PreviousID, input.Run); err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, map[string]string{"result": "regression_recorded"})
	})
}
