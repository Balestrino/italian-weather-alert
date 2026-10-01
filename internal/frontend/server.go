// Package frontend serves the public status UI using only public HTTP reads.
package frontend

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
)

//go:embed ui/*
var assets embed.FS

const maxResponseBytes = 2 << 20

var pageTemplate = template.Must(template.New("status.html").Funcs(template.FuncMap{
	"label": statusLabel,
	"timeLabel": func(at *time.Time) string {
		if at == nil || at.IsZero() {
			return "Non disponibile"
		}
		return at.UTC().Format("02/01/2006 15:04:05 UTC")
	},
}).ParseFS(assets, "ui/status.html"))

type source struct {
	ID                  string   `json:"source_id"`
	Product             string   `json:"product"`
	Territory           string   `json:"territory"`
	PublicState         string   `json:"public_state"`
	CoverageStatus      string   `json:"coverage_status"`
	CoverageLimitations []string `json:"coverage_limitations"`
	Quality             struct {
		Updating struct {
			State          string     `json:"state"`
			LastCompleteAt *time.Time `json:"last_complete_check_at"`
			Limitations    []string   `json:"limitations"`
		} `json:"updating"`
	} `json:"quality"`
}

type coverageResponse struct {
	Data *[]source `json:"data"`
	Meta *struct {
		ServedAt    time.Time `json:"served_at"`
		NextCursor  *string   `json:"next_cursor"`
		Limitations []string  `json:"limitations"`
	} `json:"meta"`
	Error *json.RawMessage `json:"error"`
}

type page struct {
	ObservedAt    *time.Time
	Availability  string
	CoverageError string
	Sources       []source
	ServedAt      *time.Time
	NextURL       string
	LaterPage     bool
	Limitations   []string
}

type server struct {
	origin string
	client *http.Client
}

// Handler accepts an operator-configured public origin, never a visitor URL.
func Handler(origin string) (http.Handler, error) {
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil ||
		(u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return nil, errors.New("invalid public backend origin")
	}
	u.Path = ""
	s := &server{origin: strings.TrimSuffix(u.String(), "/"), client: &http.Client{
		Timeout:       4 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.status)
	mux.HandleFunc("GET /assets/status.css", func(w http.ResponseWriter, r *http.Request) {
		body, _ := assets.ReadFile("ui/status.css")
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write(body)
	})
	// These endpoints describe the frontend process, independently of the backend.
	httpserver.HealthRoutes(mux, nil, false)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		mux.ServeHTTP(w, r)
	}), nil
}

func (s *server) status(w http.ResponseWriter, r *http.Request) {
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(query) > 1 || (len(query) == 1 && (len(query["cursor"]) != 1 || query.Get("cursor") == "" || len(query.Get("cursor")) > 8192)) {
		http.Error(w, "Richiesta non valida. Riparti dalla prima pagina.", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	observedAt := time.Now().UTC()
	p := page{ObservedAt: &observedAt, Availability: "Non verificabile", LaterPage: query.Get("cursor") != ""}
	var health struct {
		Status string `json:"status"`
	}
	code, err := s.read(ctx, "/health/ready", &health)
	if code == http.StatusServiceUnavailable {
		p.Availability = "Non disponibile"
	}
	if err == nil && code == http.StatusOK && health.Status == "ready" {
		p.Availability = "Disponibile"
	}
	path := "/v1/sources/coverage"
	if p.LaterPage {
		path += "?" + url.Values{"cursor": {query.Get("cursor")}}.Encode()
	}
	var coverage coverageResponse
	code, err = s.read(ctx, path, &coverage)
	switch {
	case code == http.StatusTooManyRequests:
		p.CoverageError = "Accesso temporaneamente limitato. Riprova più tardi."
	case code == http.StatusGone:
		p.CoverageError = "Pagina scaduta. Riparti dalla prima pagina."
	case code == http.StatusBadRequest:
		p.CoverageError = "Richiesta non valida. Riparti dalla prima pagina."
	case err != nil || code != http.StatusOK:
		p.CoverageError = "Informazioni sulla copertura temporaneamente non disponibili."
	case coverage.Error != nil || coverage.Data == nil || coverage.Meta == nil || coverage.Meta.ServedAt.IsZero():
		p.CoverageError = "Risposta sulla copertura non verificabile. Riprova più tardi."
	default:
		p.Sources = *coverage.Data
		p.ServedAt = &coverage.Meta.ServedAt
		p.Limitations = coverage.Meta.Limitations
		if coverage.Meta.NextCursor != nil && *coverage.Meta.NextCursor != "" {
			p.NextURL = "/?" + url.Values{"cursor": {*coverage.Meta.NextCursor}}.Encode()
		}
	}
	var body bytes.Buffer
	if err := pageTemplate.ExecuteTemplate(&body, "status.html", p); err != nil {
		http.Error(w, "Pagina temporaneamente non disponibile.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body.Bytes())
}

func (s *server) read(ctx context.Context, path string, target any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.origin+path, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")
	response, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return response.StatusCode, errors.New("public read failed")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil || len(body) > maxResponseBytes {
		return response.StatusCode, errors.New("public response unavailable")
	}
	return response.StatusCode, json.Unmarshal(body, target)
}

func statusLabel(value string) string {
	labels := map[string]string{
		"enabled": "Pubblicata", "pending": "In attesa", "suspended": "Sospesa",
		"accepted_declared_scope": "Ambito dichiarato verificato", "accepted_with_limitations": "Verificata con limiti",
		"ok": "Aggiornata", "delayed": "In ritardo", "failed": "Verifica fallita", "not_verified": "Non verificata",
		"municipal": "Fonte comunale", "criticality": "Criticità e allerte", "vigilance": "Vigilanza", "monitoring": "Monitoraggio",
	}
	if label, ok := labels[value]; ok {
		return label
	}
	return "Non verificato"
}
