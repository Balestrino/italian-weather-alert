package backoffice

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
)

//go:embed ui/*
var adminUI embed.FS

type shellData struct {
	Title, Section string
	Refresh        bool
}

func shell(title, section string) shellData {
	return shellData{Title: title, Section: section, Refresh: true}
}
func uiTemplates(name string, funcs template.FuncMap) *template.Template {
	if funcs == nil {
		funcs = template.FuncMap{}
	}
	funcs["shell"] = shell
	funcs["outcomeShell"] = func(title string) shellData { return shellData{Title: title, Section: "system"} }
	return template.Must(template.New(name).Funcs(funcs).ParseFS(adminUI, "ui/shell.html"))
}

var landingTemplates = template.Must(uiTemplates("landing", nil).ParseFS(adminUI, "ui/home.html"))

func renderUI(w http.ResponseWriter, t *template.Template, name string, data any) {
	var body bytes.Buffer
	if err := t.ExecuteTemplate(&body, name, data); err != nil {
		http.Error(w, "Interfaccia non disponibile", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body.Bytes())
}
func adminAssetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/assets/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		types := map[string]string{"admin.css": "text/css; charset=utf-8", "admin.js": "text/javascript; charset=utf-8"}
		mime, ok := types[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		body, err := adminUI.ReadFile("ui/" + name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", mime)
		_, _ = w.Write(body)
	})
}
