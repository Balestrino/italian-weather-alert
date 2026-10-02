package backoffice

import (
	"context"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"net/http"
	"net/url"
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

type adminDashboard interface {
	Dashboard(context.Context, string, time.Time) (operations.Dashboard, error)
}

type overviewPage struct {
	Cards           []overviewCard
	Dashboard       *operations.Dashboard
	Period          string
	Chart           historyChart
	Gates           []processing.GateState
	GatesAvailable  bool
	GatesObservedAt time.Time
}

type historyRect struct {
	X, Y, Width, Height float64
	Class               string
}
type historyBar struct {
	Label, Detail string
	X             float64
	LabelVisible  bool
	Rects         []historyRect
}
type historyChart struct {
	Bars                                         []historyBar
	Maximum, Succeeded, Retry, Failed, Abandoned int64
}

func makeHistoryChart(history []operations.HistoryBucket, period string) historyChart {
	c := historyChart{}
	for _, h := range history {
		total := h.Succeeded + h.Retry + h.Failed + h.Abandoned
		if total > c.Maximum {
			c.Maximum = total
		}
		c.Succeeded += h.Succeeded
		c.Retry += h.Retry
		c.Failed += h.Failed
		c.Abandoned += h.Abandoned
	}
	if len(history) == 0 {
		return c
	}
	step := 840 / float64(len(history))
	for i, h := range history {
		label := h.At.Format("02/01")
		if period == "24h" {
			label = h.At.Format("15:04")
		}
		bar := historyBar{Label: label, X: 60 + (float64(i)+.5)*step, LabelVisible: len(history) <= 7 || i%3 == 0 || i == len(history)-1,
			Detail: fmt.Sprintf("%s UTC: %d riusciti, %d da riprovare, %d falliti, %d interrotti", h.At.Format("02/01 15:04"), h.Succeeded, h.Retry, h.Failed, h.Abandoned)}
		y := 220.0
		for n, count := range []int64{h.Succeeded, h.Retry, h.Failed, h.Abandoned} {
			height := 0.0
			if c.Maximum > 0 {
				height = 180 * float64(count) / float64(c.Maximum)
			}
			y -= height
			bar.Rects = append(bar.Rects, historyRect{X: 60 + float64(i)*step + step*.15, Y: y, Width: step * .7, Height: height, Class: []string{"chart-success", "chart-retry", "chart-failed", "chart-abandoned"}[n]})
		}
		c.Bars = append(c.Bars, bar)
	}
	return c
}

func jobOverviewURL(kind, state string, scope ...string) string {
	q := url.Values{}
	if kind != "" {
		q.Set("kind", kind)
	}
	if state != "" {
		q.Set("state", state)
	}
	if len(scope) > 0 && scope[0] != "" {
		q.Set("queue", scope[0])
	}
	if len(scope) > 1 && scope[1] != "" {
		q.Set("error_code", scope[1])
	}
	return "/admin/operations/jobs?" + q.Encode()
}

func jobKindLabel(kind string) string {
	if label, ok := map[string]string{"ocr_resource": "OCR", "classify_relevance": "Classificazione", "extract_measures": "Estrazione", "embed_measure": "Embedding", "link_measure_update": "Collegamento", "retain_acquisition": "Conservazione"}[kind]; ok {
		return label
	}
	return kind
}

func jobReasonLabel(code string) string {
	if label, ok := map[string]string{
		"provider_scope_held":             "Rinvio per controllo del provider",
		"classification_output_quotation": "Citazione del modello non valida",
		"equivalent_reuse_unavailable":    "Risultato equivalente non disponibile per il riuso",
		"equivalent_input_waiting":        "In attesa del risultato della versione di riferimento",
		"territory_unavailable":           "Elaborazione territoriale non ammessa",
		"lease_expired":                   "Tempo di presa in carico scaduto",
		"quota":                           "Quota del provider esaurita",
		"auth":                            "Accesso al provider da ripristinare",
		"rate_limit":                      "Limite di richieste del provider",
	}[code]; ok {
		return label
	}
	return "Causa registrata"
}

func adminOverviewRoutes(mux *http.ServeMux, a AdminRuntime) {
	mux.HandleFunc("GET /admin/operations/{$}", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		period := r.URL.Query().Get("period")
		if period == "" {
			period = "24h"
		}
		for key, values := range r.URL.Query() {
			if key != "period" || len(values) != 1 {
				http.Error(w, "Filtro non valido", 400)
				return
			}
		}
		if _, _, err := operations.HistoryWindow(period, now); err != nil {
			http.Error(w, "Periodo non valido", 400)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		page := overviewPage{Period: period}
		cards := []overviewCard{{Title: "Fonti da verificare", Value: "Non disponibile", URL: "/admin/operations/sources?issue=attention"}, {Title: "Job falliti", Value: "Non disponibile", URL: "/admin/operations/jobs?state=failed"}, {Title: "Documenti senza interpretazione completa", Value: "Non disponibile", URL: "/admin/operations/documents"}, {Title: "Incidenti aperti", Value: "Non disponibile", URL: "/admin/notifications"}}
		var report operations.Overview
		available := false
		if reader, ok := a.Operations.(adminDashboard); ok {
			if dashboard, err := reader.Dashboard(ctx, period, now); err == nil {
				page.Dashboard = &dashboard
				page.Chart = makeHistoryChart(dashboard.History, period)
				report = dashboard.Overview
				available = true
			}
		}
		if !available {
			if reader, ok := a.Operations.(adminOverview); ok {
				var err error
				report, err = reader.Overview(ctx, now)
				available = err == nil
			}
		}
		if available {
			for i, count := range []int64{report.SourceIssues, report.FailedJobs, report.PendingDocuments} {
				cards[i].Value = strconv.FormatInt(count, 10)
				cards[i].ObservedAt = report.ObservedAt
			}
		}
		if a.Provider != nil {
			providerCtx, providerCancel := context.WithTimeout(r.Context(), 5*time.Second)
			gates, err := a.Provider.Gates(providerCtx)
			providerCancel()
			if err == nil {
				page.Gates = gates
				page.GatesAvailable = true
				page.GatesObservedAt = time.Now().UTC()
			}
		}
		if a.Notifications != nil {
			notificationCtx, notificationCancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer notificationCancel()
			if report, err := a.Notifications.Report(notificationCtx); err == nil {
				cards[3].Value = strconv.Itoa(report.OpenIncidents)
				cards[3].ObservedAt = time.Now().UTC()
			}
		}
		page.Cards = cards
		renderUI(w, landingTemplates, "home", page)
	})
}
