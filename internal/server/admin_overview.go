package server

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/operations"
	"net/http"
	"strconv"
	"time"
)

type adminOverview interface {
	Overview(context.Context, time.Time) (operations.Overview, error)
}
type overviewCard struct {
	Title, Value, URL string
	ObservedAt        time.Time
}

func adminOverviewRoutes(mux *http.ServeMux, a AdminRuntime) {
	mux.HandleFunc("GET /admin/operations/{$}", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		cards := []overviewCard{{Title: "Fonti da verificare", Value: "Non disponibile", URL: "/admin/operations/sources?issue=attention"}, {Title: "Job falliti", Value: "Non disponibile", URL: "/admin/operations/jobs?state=failed"}, {Title: "Documenti in attesa", Value: "Non disponibile", URL: "/admin/operations/documents"}, {Title: "Incidenti aperti", Value: "Non disponibile", URL: "/admin/notifications"}}
		if reader, ok := a.Operations.(adminOverview); ok {
			if report, err := reader.Overview(r.Context(), now); err == nil {
				for i, count := range []int64{report.SourceIssues, report.FailedJobs, report.PendingDocuments} {
					cards[i].Value = strconv.FormatInt(count, 10)
					cards[i].ObservedAt = report.ObservedAt
				}
			}
		}
		if a.Notifications != nil {
			if report, err := a.Notifications.Report(r.Context()); err == nil {
				cards[3].Value = strconv.Itoa(report.OpenIncidents)
				cards[3].ObservedAt = time.Now().UTC()
			}
		}
		renderUI(w, landingTemplates, "home", cards)
	})
}
