package backoffice

import (
	"context"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
)

type AdminEmbedding interface {
	EmbeddingState(context.Context) (jobs.EmbeddingControl, error)
	SourceEmbeddingStates(context.Context) ([]jobs.EmbeddingControl, error)
	SetEmbeddingEnabled(context.Context, int, bool, string, time.Time) (jobs.EmbeddingControl, error)
	SetSourceEmbeddingEnabled(context.Context, string, int, bool, string, time.Time) (jobs.EmbeddingControl, error)
}

type embeddingPage struct {
	Global  jobs.EmbeddingControl   `json:"global"`
	Sources []jobs.EmbeddingControl `json:"sources"`
}

var embeddingTemplates = template.Must(uiTemplates("embedding", template.FuncMap{"path": url.PathEscape}).ParseFS(adminUI, "ui/embedding.html"))

func adminEmbeddingRoutes(mux *http.ServeMux, a AdminRuntime) {
	mux.HandleFunc("GET /admin/embedding", func(w http.ResponseWriter, r *http.Request) {
		if a.Embedding == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		global, err := a.Embedding.EmbeddingState(ctx)
		if err != nil {
			adminFailure(w, err)
			return
		}
		sources, err := a.Embedding.SourceEmbeddingStates(ctx)
		if err != nil {
			adminFailure(w, err)
			return
		}
		page := embeddingPage{global, sources}
		if wantsAdminJSON(r) {
			httpserver.Reply(w, 200, page)
			return
		}
		renderUI(w, embeddingTemplates, "embedding", page)
	})
	update := func(w http.ResponseWriter, r *http.Request) {
		if a.Embedding == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		var input struct {
			Actor    string `json:"actor"`
			Revision *int   `json:"expected_revision"`
			Enabled  *bool  `json:"enabled"`
		}
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			if !adminDecode(w, r, &input) {
				return
			}
		} else {
			if err := territoryForm(w, r, "actor", "expected_revision", "enabled"); err != nil {
				adminFailure(w, jobs.ErrInvalid)
				return
			}
			revision, err := strconv.Atoi(r.PostForm.Get("expected_revision"))
			enabled := r.PostForm.Get("enabled") == "true"
			if err != nil || (r.PostForm.Get("enabled") != "true" && r.PostForm.Get("enabled") != "false") {
				adminFailure(w, jobs.ErrInvalid)
				return
			}
			input.Actor, input.Revision, input.Enabled = r.PostForm.Get("actor"), &revision, &enabled
		}
		if input.Revision == nil || input.Enabled == nil {
			adminFailure(w, jobs.ErrInvalid)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		var state jobs.EmbeddingControl
		var err error
		if source := r.PathValue("id"); source != "" {
			state, err = a.Embedding.SetSourceEmbeddingEnabled(ctx, source, *input.Revision, *input.Enabled, input.Actor, time.Now())
		} else {
			state, err = a.Embedding.SetEmbeddingEnabled(ctx, *input.Revision, *input.Enabled, input.Actor, time.Now())
		}
		if err != nil {
			adminFailure(w, err)
			return
		}
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") && !wantsAdminJSON(r) {
			http.Redirect(w, r, "/admin/embedding", http.StatusSeeOther)
			return
		}
		httpserver.Reply(w, 200, state)
	}
	mux.HandleFunc("POST /admin/embedding", update)
	mux.HandleFunc("POST /admin/sources/{id}/embedding", update)
}
