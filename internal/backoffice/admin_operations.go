package backoffice

import (
	"context"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

type AdminOperations interface {
	Page(context.Context, string, operations.Filter, time.Time) (operations.Report, error)
}
type AdminJobs interface {
	Relaunch(context.Context, int64, int, string, time.Time) error
}

func adminOperationRoutes(mux *http.ServeMux, a AdminRuntime) {
	adminProviderRoutes(mux, a)
	for _, section := range []string{"sources", "jobs", "documents", "findings", "dpc-comparisons", "usage"} {
		mux.HandleFunc("GET /admin/operations/"+section, func(w http.ResponseWriter, r *http.Request) {
			f := operations.Filter{SourceID: r.URL.Query().Get("source_id"), Page: 1}
			for key, values := range r.URL.Query() {
				if len(values) != 1 {
					adminFailure(w, operations.ErrInvalid)
					return
				}
				var err error
				switch key {
				case "source_id":
				case "queue", "error_code":
					if key == "queue" {
						f.Queue = values[0]
					} else {
						f.ErrorCode = values[0]
					}
					if section != "jobs" || len(values[0]) > 200 {
						err = operations.ErrInvalid
					}
				case "kind":
					f.Kind = values[0]
					if section != "jobs" || len(f.Kind) > 200 {
						err = operations.ErrInvalid
					}
				case "issue":
					f.Issue = values[0]
					if section != "sources" || f.Issue != "attention" {
						err = operations.ErrInvalid
					}
				case "state":
					f.State = values[0]
					if section != "jobs" || !operations.ValidJobState(f.State) {
						err = operations.ErrInvalid
					}
				case "page":
					f.Page, err = strconv.Atoi(values[0])
				case "document_version_id":
					f.VersionID, err = strconv.ParseInt(values[0], 10, 64)
					if f.VersionID < 1 {
						err = operations.ErrInvalid
					}
				case "job_id":
					f.JobID, err = strconv.ParseInt(values[0], 10, 64)
					if f.JobID < 1 {
						err = operations.ErrInvalid
					}
				case "run_id":
					f.RunID, err = strconv.ParseInt(values[0], 10, 64)
					if f.RunID < 1 {
						err = operations.ErrInvalid
					}
				default:
					err = operations.ErrInvalid
				}
				if err != nil {
					adminFailure(w, operations.ErrInvalid)
					return
				}
			}
			if a.Operations == nil {
				adminFailure(w, registryUnavailable)
				return
			}
			report, err := a.Operations.Page(r.Context(), section, f, time.Now().UTC())
			if err != nil {
				adminFailure(w, err)
				return
			}
			if wantsAdminJSON(r) {
				httpserver.Reply(w, 200, report)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = operationTemplates.ExecuteTemplate(w, "page", report)
		})
	}
	mux.HandleFunc("POST /admin/jobs/{id}/relaunch", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			adminFailure(w, registry.ErrInvalid)
			return
		}
		var input struct {
			Actor        string `json:"actor"`
			AfterAttempt int    `json:"after_attempt"`
		}
		if !adminDecode(w, r, &input) {
			return
		}
		if a.Jobs == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		if err = a.Jobs.Relaunch(r.Context(), id, input.AfterAttempt, input.Actor, time.Now().UTC()); err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, map[string]any{"job_id": id, "result": "relaunch_recorded"})
	})
}

var operationTemplates = template.Must(uiTemplates("operations", template.FuncMap{
	"title": func(s string) string {
		return map[string]string{"sources": "Stato delle fonti", "jobs": "Job", "documents": "Documenti da interpretare", "findings": "Riscontri di qualità e valutazioni", "dpc-comparisons": "Confronti interni CFR–DPC", "usage": "Consumi e durate"}[s]
	},
	"value": func(v any) string {
		if v == nil {
			return "non disponibile"
		}
		switch t := v.(type) {
		case time.Time:
			return t.UTC().Format("02/01/2006 15:04:05 UTC")
		case *time.Time:
			if t == nil {
				return "non disponibile"
			}
			return t.UTC().Format("02/01/2006 15:04:05 UTC")
		case bool:
			if t {
				return "Sì"
			}
			return "No"
		case map[string]any, []any:
			return adminJSON(v)
		}
		return fmt.Sprint(v)
	},
	"sectionNav": func(s string) string {
		if s == "dpc-comparisons" {
			return "findings"
		}
		return s
	},
	"label": columnLabel, "primary": primaryColumns, "status": statusLabel,
	"json": adminJSON, "path": url.PathEscape,
	"relaunch": func(row map[string]any) string {
		return adminJSON(map[string]any{"actor": "", "after_attempt": row["attempt_count"]})
	},
	"pageURL": func(r operations.Report, delta int) string {
		q := url.Values{"page": {strconv.Itoa(r.Filter.Page + delta)}}
		if r.Filter.Issue != "" {
			q.Set("issue", r.Filter.Issue)
		}
		if r.Filter.State != "" {
			q.Set("state", r.Filter.State)
		}
		if r.Filter.Queue != "" {
			q.Set("queue", r.Filter.Queue)
		}
		if r.Filter.ErrorCode != "" {
			q.Set("error_code", r.Filter.ErrorCode)
		}
		if r.Filter.Kind != "" {
			q.Set("kind", r.Filter.Kind)
		}
		if r.Filter.SourceID != "" {
			q.Set("source_id", r.Filter.SourceID)
		}
		if r.Filter.VersionID != 0 {
			q.Set("document_version_id", strconv.FormatInt(r.Filter.VersionID, 10))
		}
		if r.Filter.JobID != 0 {
			q.Set("job_id", strconv.FormatInt(r.Filter.JobID, 10))
		}
		if r.Filter.RunID != 0 {
			q.Set("run_id", strconv.FormatInt(r.Filter.RunID, 10))
		}
		return "?" + q.Encode()
	},
}).ParseFS(adminUI, "ui/operations.html"))
