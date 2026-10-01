package backoffice

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type reportLink struct{ Label, URL string }
type reportPage struct {
	Shell  shellData
	Result any
	Links  []reportLink
}

func reportRows(value any) []any {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var decoded any
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if decoder.Decode(&decoded) != nil {
		return nil
	}
	if rows, ok := decoded.([]any); ok {
		return rows
	}
	if decoded == nil {
		return nil
	}
	return []any{decoded}
}
func recordTitle(value any) string {
	if row, ok := value.(map[string]any); ok {
		for _, key := range []string{"id", "ID", "source_id", "campaign_id", "category", "status", "state"} {
			if v, ok := row[key]; ok && v != nil && fmt.Sprint(v) != "" {
				return fmt.Sprint(v)
			}
		}
	}
	return "Dettaglio record"
}

var reportTemplates = template.Must(uiTemplates("reports", template.FuncMap{"rows": reportRows, "recordTitle": recordTitle, "json": adminJSON, "label": columnLabel, "path": url.PathEscape}).ParseFS(adminUI, "ui/reports.html"))

func renderReport(w http.ResponseWriter, r *http.Request, result any) {
	title, section := "Rapporto operativo", "system"
	p := r.URL.Path
	switch {
	case strings.Contains(p, "processing-evaluations"):
		title, section = "Valutazioni delle elaborazioni", "findings"
	case strings.Contains(p, "regressions"):
		title, section = "Rapporti di regressione", "findings"
	case strings.Contains(p, "acceptance-reviews"):
		title, section = "Revisioni di accettazione", "sources"
	case strings.Contains(p, "diagnostics/sources"):
		title, section = "Diagnostica fonte", "sources"
	case strings.Contains(p, "cost-report"):
		title = "Costi del trial"
	case strings.Contains(p, "budget-proposals"):
		title = "Proposte di budget"
	case strings.HasSuffix(p, "/report"):
		title = "Rapporto osservativo"
	case strings.Contains(p, "observation-campaigns"):
		title = "Campagne osservative"
	}
	links := []reportLink{}
	if id := r.PathValue("id"); id != "" && strings.Contains(p, "observation-campaigns") {
		base := "/admin/observation-campaigns/" + url.PathEscape(id)
		links = []reportLink{{"Campagna", base}, {"Rapporto osservativo", base + "/report"}, {"Costi del trial", base + "/cost-report?through=" + url.QueryEscape(time.Now().UTC().Format(time.RFC3339))}, {"Proposte di budget", base + "/budget-proposals"}}
	}
	if strings.HasPrefix(p, "/admin/sources/") {
		links = append(links, reportLink{"Torna alla fonte", actionBack(r)})
	}
	renderUI(w, reportTemplates, "report", reportPage{Shell: shell(title, section), Result: result, Links: links})
}
