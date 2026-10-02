package backoffice

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
	"html/template"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/diagnostics"
	"github.com/Balestrino/italian-weather-alert/internal/backend/evaluation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/observation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"github.com/Balestrino/italian-weather-alert/internal/backend/trialcost"
	"github.com/jackc/pgx/v5/pgconn"
)

type AdminRegistry interface {
	Sources(context.Context) ([]registry.State, error)
	Accept(context.Context, string, int, string, registry.Acceptance) error
	EnablePublic(context.Context, string, int, string) error
	ReviewAcceptance(context.Context, string, int, string, registry.Evidence) (registry.AcceptanceReview, error)
	AcceptanceReviews(context.Context, string) ([]registry.AcceptanceReview, error)
	ReleaseReadiness(context.Context) (registry.ReleaseReadiness, error)
	SuspendInterpretation(context.Context, string, int, string, registry.Evidence) error
	InterpretationHistory(context.Context, string) ([]registry.InterpretationSuspension, error)
	State(context.Context, string) (registry.State, error)
	Versions(context.Context, string) ([]registry.Version, error)
	Events(context.Context, string) ([]registry.Event, error)
	Intervals(context.Context, string) (registry.Intervals, error)
	CreateAuthority(context.Context, registry.Authority) error
	CreateChannel(context.Context, registry.Channel) error
	CreateSource(context.Context, registry.Source, registry.Configuration, string) error
	AppendConfiguration(context.Context, string, int, registry.Configuration, string) (int, error)
	EnableCollection(context.Context, string, int, string) error
	Disable(context.Context, string, int, string, bool) error
	SetIntervals(context.Context, string, int, int, int, string) error
}
type AdminPreview interface {
	Preview(context.Context, string, int, string) (acquisition.Preview, error)
}
type AdminInterpretation interface {
	ResumeInterpretation(context.Context, string, int, string, interpretation.Recovery) error
	Reprocess(context.Context, interpretation.Selection) (string, int, error)
}

type AdminRuntime struct {
	DevelopmentPublication interface {
		State(context.Context, string, string) (domain.DevelopmentPublicationState, error)
		Set(context.Context, string, string, int, bool, string) (int, error)
	}
	Territories     AdminTerritories
	TerritoryConfig AdminTerritoryConfiguration
	Alerts          AdminAlerts
	TailscaleOrigin string
	Provider        AdminProvider
	Backups         AdminBackups
	Notifications   AdminNotifications
	Operations      AdminOperations
	Diagnostics     interface {
		Snapshot(context.Context, string, time.Time) (diagnostics.Snapshot, error)
		Compare(context.Context, diagnostics.Comparison) (diagnostics.Comparison, error)
	}
	Jobs           AdminJobs
	Interpretation AdminInterpretation
	Evaluations    interface {
		RecordComparison(context.Context, string, evaluation.ProcessingComparison) (evaluation.ComparisonReport, error)
		List(context.Context) ([]evaluation.ComparisonReport, error)
	}
	Observations interface {
		Start(context.Context, observation.StartRequest) (observation.Campaign, error)
		RecordReview(context.Context, string, string, observation.Review) (observation.Review, error)
		Assess(context.Context, string, string, time.Time, registry.Evidence) (observation.Assessment, error)
		Report(context.Context, string, time.Time) (observation.Report, error)
		Get(context.Context, string) (observation.Campaign, error)
		List(context.Context) ([]observation.Campaign, error)
	}
	TrialCosts interface {
		RecordInput(context.Context, string, string, trialcost.Input) (trialcost.Input, error)
		Report(context.Context, string, time.Time) (trialcost.Report, error)
		Propose(context.Context, string, string, time.Time, registry.Evidence) (trialcost.Proposal, error)
		Proposals(context.Context, string) ([]trialcost.Proposal, error)
	}
	Registry AdminRegistry
	Preview  AdminPreview
}

func adminSourceRoutes(mux *http.ServeMux, a AdminRuntime) {
	adminNotificationRoutes(mux, a)
	adminBackupRoutes(mux, a)
	adminObservationRoutes(mux, a)
	adminTrialCostRoutes(mux, a)
	adminReleaseRoutes(mux, a)
	adminLifecycleRoutes(mux, a)
	adminOperationRoutes(mux, a)
	adminDiagnosticRoutes(mux, a)
	mux.HandleFunc("GET /admin/sources", func(w http.ResponseWriter, r *http.Request) {
		if a.Registry == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		sources, err := a.Registry.Sources(r.Context())
		if err != nil {
			adminFailure(w, err)
			return
		}
		if wantsAdminJSON(r) {
			httpserver.Reply(w, 200, sources)
			return
		}
		adminRender(w, "sources", sources)
	})
	mux.HandleFunc("GET /admin/sources/{id}", func(w http.ResponseWriter, r *http.Request) {
		if a.Registry == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		id := r.PathValue("id")
		st, err := a.Registry.State(r.Context(), id)
		if err != nil {
			adminFailure(w, err)
			return
		}
		versions, err := a.Registry.Versions(r.Context(), id)
		if err != nil {
			adminFailure(w, err)
			return
		}
		events, err := a.Registry.Events(r.Context(), id)
		if err != nil {
			adminFailure(w, err)
			return
		}
		intervals, err := a.Registry.Intervals(r.Context(), id)
		if err != nil {
			adminFailure(w, err)
			return
		}
		configuration := versions[len(versions)-1].Configuration
		if st.ActiveRevision != nil {
			configuration = versions[*st.ActiveRevision-1].Configuration
		}
		check, delay := configuration.CheckSeconds, configuration.DelaySeconds
		if intervals.CheckSeconds != nil {
			check = *intervals.CheckSeconds
		}
		if intervals.DelaySeconds != nil {
			delay = *intervals.DelaySeconds
		}
		history, err := a.Registry.InterpretationHistory(r.Context(), id)
		if err != nil {
			adminFailure(w, err)
			return
		}
		back, _ := r.Context().Value(territoryReturnKey{}).(string)
		view := adminSourceView{Back: back, InterpretationHistory: history, State: st, Versions: versions, Events: events, Intervals: intervals, EffectiveCheckSeconds: check, EffectiveDelaySeconds: delay}
		if wantsAdminJSON(r) {
			httpserver.Reply(w, 200, view)
			return
		}
		adminRender(w, "source", view)
	})
	// Form and JSON clients use the same validation and persisted operations.
	mux.HandleFunc("POST /admin/authorities", func(w http.ResponseWriter, r *http.Request) {
		var input registry.Authority
		if !adminDecode(w, r, &input) {
			return
		}
		if a.Registry == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		if err := a.Registry.CreateAuthority(r.Context(), input); err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, map[string]string{"authority_id": input.ID})
	})
	mux.HandleFunc("POST /admin/channels", func(w http.ResponseWriter, r *http.Request) {
		var input registry.Channel
		if !adminDecode(w, r, &input) {
			return
		}
		if a.Registry == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		if err := a.Registry.CreateChannel(r.Context(), input); err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, map[string]string{"channel_id": input.ID})
	})
	mux.HandleFunc("POST /admin/sources", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Source        registry.Source        `json:"source"`
			Configuration registry.Configuration `json:"configuration"`
			Actor         string                 `json:"actor"`
		}
		if !adminDecode(w, r, &input) {
			return
		}
		if a.Registry == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		if err := a.Registry.CreateSource(r.Context(), input.Source, input.Configuration, input.Actor); err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, map[string]any{"source_id": input.Source.ID, "revision": 1})
	})
	mux.HandleFunc("POST /admin/sources/{id}/configuration", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ExpectedRevision int                    `json:"expected_revision"`
			Configuration    registry.Configuration `json:"configuration"`
			Actor            string                 `json:"actor"`
		}
		if !adminDecode(w, r, &input) {
			return
		}
		if a.Registry == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		rev, err := a.Registry.AppendConfiguration(r.Context(), r.PathValue("id"), input.ExpectedRevision, input.Configuration, input.Actor)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, map[string]int{"revision": rev})
	})
	mux.HandleFunc("POST /admin/sources/{id}/intervals", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ExpectedRevision int    `json:"expected_revision"`
			CheckSeconds     int    `json:"check_seconds"`
			DelaySeconds     int    `json:"delay_seconds"`
			Actor            string `json:"actor"`
		}
		if !adminDecode(w, r, &input) {
			return
		}
		if a.Registry == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		err := a.Registry.SetIntervals(r.Context(), r.PathValue("id"), input.ExpectedRevision, input.CheckSeconds, input.DelaySeconds, input.Actor)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, map[string]int{"interval_revision": input.ExpectedRevision + 1})
	})
	for _, action := range []string{"preview", "enable-collection", "suspend-collection"} {
		mux.HandleFunc("POST /admin/sources/{id}/"+action, func(w http.ResponseWriter, r *http.Request) {
			var input struct {
				Revision int    `json:"revision"`
				Actor    string `json:"actor"`
			}
			if !adminDecode(w, r, &input) {
				return
			}
			if a.Registry == nil {
				adminFailure(w, registryUnavailable)
				return
			}
			if input.Revision < 1 || strings.TrimSpace(input.Actor) == "" {
				adminFailure(w, registry.ErrInvalid)
				return
			}
			id := r.PathValue("id")
			var err error
			switch action {
			case "preview":
				if a.Preview == nil {
					adminFailure(w, registryUnavailable)
					return
				}
				// Bounded explicit acquisition may outlive the normal ten-second write
				// deadline; do not extend deadlines on other admin/public requests.
				_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(16 * time.Minute))
				ctx, cancel := context.WithTimeout(r.Context(), 15*time.Minute)
				defer cancel()
				var report acquisition.Preview
				report, err = a.Preview.Preview(ctx, id, input.Revision, input.Actor)
				if err == nil {
					adminSuccess(w, r, report)
					return
				}
			case "enable-collection":
				err = a.Registry.EnableCollection(r.Context(), id, input.Revision, input.Actor)
			case "suspend-collection":
				err = a.Registry.Disable(r.Context(), id, input.Revision, input.Actor, false)
			}
			if err != nil {
				adminFailure(w, err)
				return
			}
			adminSuccess(w, r, map[string]string{"result": action})
		})
	}
}

var registryUnavailable = errors.New("administration unavailable")

type adminSourceView struct {
	Back                  string `json:"-"`
	InterpretationHistory []registry.InterpretationSuspension
	State                 registry.State
	Versions              []registry.Version
	Events                []registry.Event
	Intervals             registry.Intervals
	EffectiveCheckSeconds int
	EffectiveDelaySeconds int
}

func wantsAdminJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json") || strings.HasPrefix(r.Header.Get("Content-Type"), "application/json")
}
func adminDecode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		adminFailure(w, registry.ErrInvalid)
		return false
	}
	var body io.Reader = r.Body
	switch media {
	case "application/json":
	case "application/x-www-form-urlencoded":
		if r.ParseForm() != nil {
			adminFailure(w, registry.ErrInvalid)
			return false
		}
		if _, ok := r.PostForm["payload"]; ok {
			if len(r.PostForm["payload"]) != 1 {
				adminFailure(w, registry.ErrInvalid)
				return false
			}
			body = strings.NewReader(r.PostForm.Get("payload"))
		} else {
			payload, err := guidedPayload(r)
			if err != nil {
				adminFailure(w, err)
				return false
			}
			body = strings.NewReader(payload)
		}
	default:
		http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
		return false
	}
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()
	if dec.Decode(v) != nil || dec.Decode(new(any)) != io.EOF {
		adminFailure(w, registry.ErrInvalid)
		return false
	}
	return true
}
func adminFailure(w http.ResponseWriter, err error) {
	status, code := http.StatusServiceUnavailable, "administration_unavailable"
	var pg *pgconn.PgError
	switch {
	case errors.Is(err, territory.ErrAssociation), errors.Is(err, territory.ErrUnsupported), errors.Is(err, territory.ErrDisabled), errors.Is(err, territory.ErrPublicScope):
		status, code = 409, "territorial_prerequisite_not_met"
	case errors.Is(err, jobs.ErrNoJob):
		status, code = 404, "job_not_found"
	case errors.Is(err, jobs.ErrConflict):
		status, code = 409, "job_state_conflict"
	case errors.Is(err, jobs.ErrInvalid), errors.Is(err, operations.ErrInvalid):
		status, code = 400, "invalid_operation"
	case errors.Is(err, diagnostics.ErrInvalid):
		status, code = 400, "invalid_diagnostic"
	case errors.Is(err, diagnostics.ErrNotFound):
		status, code = 404, "diagnostic_evidence_not_found"
	case errors.Is(err, diagnostics.ErrConflict):
		status, code = 409, "diagnostic_conflict"
	case errors.Is(err, observation.ErrInvalid):
		status, code = 400, "invalid_observation"
	case errors.Is(err, observation.ErrNotFound):
		status, code = 404, "observation_not_found"
	case errors.Is(err, observation.ErrConflict):
		status, code = 409, "observation_conflict"
	case errors.Is(err, trialcost.ErrInvalid):
		status, code = 400, "invalid_trial_cost"
	case errors.Is(err, trialcost.ErrNotFound):
		status, code = 404, "trial_cost_campaign_not_found"
	case errors.Is(err, trialcost.ErrIncomplete):
		status, code = 409, "trial_cost_report_incomplete"
	case errors.Is(err, trialcost.ErrConflict):
		status, code = 409, "trial_cost_conflict"
	case errors.Is(err, registry.ErrNotFound):
		status, code = 404, "source_not_found"
	case errors.Is(err, registry.ErrInvalid), errors.Is(err, acquisition.ErrInvalidConfiguration), errors.Is(err, interpretation.ErrInvalid), errors.Is(err, evaluation.ErrComparisonInvalid):
		status, code = 400, "invalid_configuration"
	case errors.Is(err, registry.ErrConflict):
		status, code = 409, "revision_conflict"
	case errors.Is(err, registry.ErrPrerequisite):
		status, code = 409, "source_prerequisite_missing"
	case errors.As(err, &pg) && pg.Code == "23505":
		status, code = 409, "identifier_exists"
	case errors.As(err, &pg) && pg.Code == "23503":
		status, code = 400, "unknown_registry_reference"
	}
	// Never return database, storage, crawler response or credential details.
	if browser, ok := w.(*adminBrowserWriter); ok {
		message := "Operazione non eseguita. Verifica i dati inseriti e riprova."
		if status == 409 {
			message = "Lo stato o i prerequisiti sono cambiati. Ricarica la pagina e verifica la selezione prima di riprovare."
		}
		if status == 503 {
			message = "Dati temporaneamente non disponibili. Riprova più tardi."
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		renderUI(w, outcomeTemplates, "outcome", actionOutcome{Title: "Operazione non disponibile", Message: message, Code: code, Back: actionBack(browser.request)})
		return
	}
	httpserver.Reply(w, status, map[string]string{"error": code})
}
func adminSuccess(w http.ResponseWriter, r *http.Request, result any) {
	if wantsAdminJSON(r) {
		httpserver.Reply(w, 200, result)
		return
	}
	if r.Method == http.MethodGet {
		renderReport(w, r, result)
		return
	}
	renderUI(w, outcomeTemplates, "outcome", actionOutcome{Title: "Operazione completata", Message: "Richiesta registrata. Consulta lo stato aggiornato prima di eseguire altre operazioni.", Back: actionBack(r), Result: result})
}
func adminRender(w http.ResponseWriter, name string, value any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	renderUI(w, adminTemplates, name, value)
}
func adminJSON(v any) string { b, _ := json.MarshalIndent(v, "", "  "); return string(b) }

var adminTemplates = template.Must(uiTemplates("admin", template.FuncMap{
	"rows": reportRows, "recordTitle": recordTitle, "label": columnLabel, "fields": sourceFields, "json": adminJSON, "path": url.PathEscape,
	"edit": func(v registry.Version) string {
		return adminJSON(map[string]any{"actor": "", "expected_revision": v.Revision, "configuration": v.Configuration})
	},
	"event":  func(e registry.Event) string { return string(e.Evidence) },
	"latest": func(v []registry.Version) registry.Version { return v[len(v)-1] },
}).ParseFS(adminUI, "ui/sources.html", "ui/fields.html", "ui/reports.html"))
