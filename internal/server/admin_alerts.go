package server

import (
	"context"
	"html/template"
	"github.com/Balestrino/italian-weather-alert/internal/operations"
	"net/http"
	"time"
)

type AdminAlerts interface {
	Alerts(context.Context, time.Time) (operations.Alerts, error)
}

func alertLabel(s string) string {
	labels := map[string]string{"criticality": "Bollettino di criticità", "vigilance": "Vigilanza meteorologica", "monitoring": "Monitoraggio", "current": "In corso", "future": "Previsto", "expired": "Terminato", "undetermined": "Validità da verificare", "green": "Verde", "yellow": "Giallo", "orange": "Arancione", "red": "Rosso", "unknown": "Non determinato", "not_applicable": "Non applicabile"}
	if v, ok := labels[s]; ok {
		return v
	}
	return s
}

var alertsTemplate = template.Must(uiTemplates("alerts", template.FuncMap{"alertLabel": alertLabel}).ParseFS(adminUI, "ui/alerts.html"))

func adminAlertRoutes(mux *http.ServeMux, runtime AdminRuntime) {
	mux.HandleFunc("GET /admin/alerts", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if runtime.Alerts == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			renderUI(w, alertsTemplate, "alerts-unavailable", nil)
			return
		}
		data, err := runtime.Alerts.Alerts(ctx, time.Now())
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			renderUI(w, alertsTemplate, "alerts-unavailable", nil)
			return
		}
		renderUI(w, alertsTemplate, "alerts", data)
	})
}
