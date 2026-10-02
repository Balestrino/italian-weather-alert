package backoffice

import (
	"cmp"
	"context"
	"encoding/base64"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"html/template"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"
)

type AdminTerritories interface {
	TerritorialOverview(context.Context, time.Time) (operations.TerritorialOverview, error)
	Municipalities(context.Context, string, operations.MunicipalityFilter, time.Time) (operations.MunicipalityPage, error)
	Municipality(context.Context, string, string) (operations.TerritorialMunicipality, error)
	TerritorialSources(context.Context, string, string) ([]operations.TerritorialSource, error)
	TerritorialResults(context.Context, string, string, time.Time) (operations.TerritorialResults, error)
	TerritorialHistory(context.Context, string, string, operations.TerritorialHistoryFilter, time.Time) (operations.TerritorialHistory, error)
}
type AdminTerritoryConfiguration interface {
	Regions(context.Context) ([]domain.Region, error)
	Region(context.Context, string) (domain.Region, error)
	ConfigureRegion(context.Context, string, int, domain.RegionConfiguration, string) (int, error)
	SetRegionEnabled(context.Context, string, int, bool, string) (int, error)
	SetMunicipalityEnabled(context.Context, string, string, int, bool, string) (int, error)
	AdoptMunicipalities(context.Context, []byte, domain.MunicipalityImport, string, int, string) (string, error)
	TerritoryMigrationReport(context.Context) (domain.TerritoryMigrationReport, error)
}

var territoryTemplates = template.Must(uiTemplates("territories", template.FuncMap{
	"count": func(n *int64) any {
		if n == nil {
			return "Non disponibile"
		}
		return *n
	},
	"regionState": func(r domain.Region) string {
		if r.Enabled {
			return "Abilitata"
		}
		if r.Revision == 0 {
			return "Da configurare"
		}
		return "Disabilitata"
	},
	"blockedReason": func(reason string) string {
		switch reason {
		case "region_disabled":
			return "Sospeso: regione disabilitata"
		case "municipality_disabled":
			return "Sospeso: comune disabilitato"
		case "municipality_retired":
			return "Sospeso: comune fuori dall’anagrafica attuale"
		default:
			return "Territorio abilitato"
		}
	},
	"base64":     func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) },
	"hasProfile": func(p []string, v string) bool { return slices.Contains(p, v) }, "json": adminJSON, "alertLabel": alertLabel, "path": url.PathEscape,
}).ParseFS(adminUI, "ui/territories.html", "ui/territory.html", "ui/territory_setup.html", "ui/alerts.html"))

type regionsPage struct {
	Overview    operations.TerritorialOverview
	State       string
	Unavailable bool
	Controls    bool
}

func adminTerritoryRoutes(mux *http.ServeMux, a AdminRuntime) {
	adminTerritorySetupRoutes(mux, a)
	handler := func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state == "" {
			state = "all"
		}
		if !slices.Contains([]string{"all", "enabled", "disabled"}, state) {
			http.Error(w, "Filtro non valido", 400)
			return
		}
		page := regionsPage{State: state, Controls: a.TerritoryConfig != nil}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if a.Territories == nil {
			page.Unavailable = true
		} else {
			var err error
			page.Overview, err = a.Territories.TerritorialOverview(ctx, time.Now())
			page.Unavailable = err != nil
		}
		if !page.Unavailable && state != "all" {
			all := page.Overview.Regions
			page.Overview.Regions = []operations.RegionSummary{}
			for _, region := range all {
				if region.Enabled == (state == "enabled") {
					page.Overview.Regions = append(page.Overview.Regions, region)
				}
			}
		}
		if !page.Unavailable {
			page.Overview.Regions = slices.Clone(page.Overview.Regions)
			slices.SortFunc(page.Overview.Regions, func(a, b operations.RegionSummary) int {
				if a.Enabled != b.Enabled {
					if a.Enabled {
						return -1
					}
					return 1
				}
				if (a.EnabledMunicipalities == nil) != (b.EnabledMunicipalities == nil) {
					if a.EnabledMunicipalities == nil {
						return 1
					}
					return -1
				}
				if a.EnabledMunicipalities != nil && b.EnabledMunicipalities != nil {
					if order := cmp.Compare(*b.EnabledMunicipalities, *a.EnabledMunicipalities); order != 0 {
						return order
					}
				}
				if order := cmp.Compare(a.Name, b.Name); order != 0 {
					return order
				}
				return cmp.Compare(a.Code, b.Code)
			})
		}
		renderUI(w, territoryTemplates, "regions", page)
	}
	mux.HandleFunc("GET /admin/{$}", handler)
	mux.HandleFunc("GET /admin/regions", handler)
	mux.HandleFunc("GET /admin/regions/{region}", func(w http.ResponseWriter, r *http.Request) { territoryDetail(w, r, a, "") })
	mux.HandleFunc("GET /admin/regions/{region}/municipalities/{istat}", func(w http.ResponseWriter, r *http.Request) { territoryDetail(w, r, a, r.PathValue("istat")) })
}

type territoryPage struct {
	DevelopmentPublication *domain.DevelopmentPublicationState
	Municipality           *operations.TerritorialMunicipality
	Region                 domain.Region
	Title, Base, Tab       string
	ReturnTo               string
	Municipalities         operations.MunicipalityPage
	Sources                []operations.TerritorialSource
	Results                operations.TerritorialResults
	History                operations.TerritorialHistory
	NextURL, From, Through string
	Unavailable            bool
}

func territoryFailure(w http.ResponseWriter, r *http.Request, err error) {
	status, message := 503, "Dati temporaneamente non disponibili. Riprova aggiornando la pagina."
	if errors.Is(err, domain.ErrTerritoryNotFound) {
		status, message = 404, "Territorio non trovato."
	}
	if errors.Is(err, domain.ErrInvalid) || errors.Is(err, operations.ErrInvalid) {
		status, message = 400, "Parametri non validi. Controlla i campi e riprova."
	}
	if errors.Is(err, domain.ErrConflict) {
		status, message = 409, "La configurazione è cambiata. Ricarica la pagina prima di riprovare."
	}
	if errors.Is(err, domain.ErrRegionIncomplete) {
		status, message = 409, "Adotta un'anagrafica completa dei comuni prima di abilitare la regione."
	}
	if errors.Is(err, domain.ErrProfileUnsupported) {
		status, message = 400, "Profilo di elaborazione non supportato per questa regione."
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	renderUI(w, outcomeTemplates, "outcome", actionOutcome{Title: "Operazione non completata", Message: message, Back: "/admin/regions"})
}
func territoryDetail(w http.ResponseWriter, r *http.Request, a AdminRuntime, istat string) {
	if a.Territories == nil || a.TerritoryConfig == nil {
		territoryFailure(w, r, registryUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	region, err := a.TerritoryConfig.Region(ctx, r.PathValue("region"))
	if err != nil {
		territoryFailure(w, r, err)
		return
	}
	q := r.URL.Query()
	tab := q.Get("tab")
	if tab == "" {
		tab = "results"
	}
	if !slices.Contains([]string{"results", "municipalities", "configuration", "history"}, tab) {
		territoryFailure(w, r, operations.ErrInvalid)
		return
	}
	page := territoryPage{Region: region, Title: region.Name, Base: "/admin/regions/" + url.PathEscape(region.Code), Tab: tab, From: q.Get("from"), Through: q.Get("through")}
	if istat != "" {
		if tab == "municipalities" {
			territoryFailure(w, r, operations.ErrInvalid)
			return
		}
		m, e := a.Territories.Municipality(ctx, region.Code, istat)
		if e != nil {
			territoryFailure(w, r, e)
			return
		}
		if a.DevelopmentPublication != nil && region.Code == "09" {
			state, e := a.DevelopmentPublication.State(ctx, region.Code, istat)
			if e != nil {
				territoryFailure(w, r, e)
				return
			}
			page.DevelopmentPublication = &state
		}
		page.Municipality = &m
		page.Title = m.Name
		page.Base += "/municipalities/" + url.PathEscape(m.ISTAT)
	}
	if tab == "municipalities" {
		ret := url.Values{"tab": {"municipalities"}}
		for _, key := range []string{"search", "province", "coverage", "after"} {
			if q.Get(key) != "" {
				ret.Set(key, q.Get(key))
			}
		}
		page.ReturnTo = page.Base + "?" + ret.Encode()
	}
	next := ""
	switch tab {
	case "municipalities":
		page.Municipalities, err = a.Territories.Municipalities(ctx, region.Code, operations.MunicipalityFilter{Search: q.Get("search"), Province: q.Get("province"), Coverage: q.Get("coverage"), After: q.Get("after")}, time.Now())
		next = page.Municipalities.Next
	case "results":
		page.Results, err = a.Territories.TerritorialResults(ctx, region.Code, istat, time.Now())
	case "configuration":
		page.Sources, err = a.Territories.TerritorialSources(ctx, region.Code, istat)
	case "history":
		f := operations.TerritorialHistoryFilter{Source: q.Get("source"), Kind: q.Get("kind"), After: q.Get("after")}
		f.From, err = territoryDate(q.Get("from"), false)
		if err == nil {
			f.Through, err = territoryDate(q.Get("through"), true)
		}
		if err == nil {
			page.History, err = a.Territories.TerritorialHistory(ctx, region.Code, istat, f, time.Now())
			next = page.History.Next
		}
	}
	if errors.Is(err, operations.ErrInvalid) || errors.Is(err, domain.ErrTerritoryNotFound) {
		territoryFailure(w, r, err)
		return
	}
	page.Unavailable = err != nil
	if next != "" {
		q.Set("after", next)
		q.Set("tab", tab)
		page.NextURL = page.Base + "?" + q.Encode()
	}
	renderUI(w, territoryTemplates, "territory", page)
}
func territoryDate(raw string, end bool) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	zone, _ := time.LoadLocation("Europe/Rome")
	at, err := time.ParseInLocation(time.DateOnly, raw, zone)
	if err != nil {
		return nil, operations.ErrInvalid
	}
	if end {
		at = at.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}
	return &at, nil
}

// Keep integer field parsing strict for native forms added by the setup workflow.
func territoryRevision(raw string) (int, error) {
	n, e := strconv.Atoi(raw)
	if e != nil || n < 0 {
		return 0, domain.ErrInvalid
	}
	return n, nil
}
