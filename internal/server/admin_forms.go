package server

import (
	"encoding/json"
	"html/template"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type adminBrowserWriter struct {
	http.ResponseWriter
	request *http.Request
}

var outcomeTemplates = templateOutcome()

func templateOutcome() *template.Template {
	return template.Must(uiTemplates("outcome", template.FuncMap{"json": adminJSON}).ParseFS(adminUI, "ui/outcome.html"))
}

type actionOutcome struct {
	Title, Message, Back, Code string
	Result                     any
}

func actionBack(r *http.Request) string {
	if back, ok := r.Context().Value(territoryReturnKey{}).(string); ok {
		return back
	}
	if strings.HasPrefix(r.URL.Path, "/admin/jobs/") {
		return "/admin/operations/jobs?job_id=" + r.PathValue("id")
	}
	if strings.HasPrefix(r.URL.Path, "/admin/sources/") {
		return "/admin/sources/" + url.PathEscape(r.PathValue("id"))
	}
	return "/admin/"
}
func guidedPayload(r *http.Request) (string, error) {
	if !strings.HasPrefix(r.URL.Path, "/admin/jobs/") || !strings.HasSuffix(r.URL.Path, "/relaunch") {
		return sourceGuidedPayload(r)
	}
	if len(r.PostForm) != 2 || len(r.PostForm["actor"]) != 1 || len(r.PostForm["after_attempt"]) != 1 {
		return "", registry.ErrInvalid
	}
	actor := strings.TrimSpace(r.PostForm.Get("actor"))
	attempt, err := strconv.Atoi(r.PostForm.Get("after_attempt"))
	if err != nil || attempt < 1 || actor == "" {
		return "", registry.ErrInvalid
	}
	b, err := json.Marshal(map[string]any{"actor": actor, "after_attempt": attempt})
	return string(b), err
}
