package frontend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const emptyCoverage = `{"data":[],"meta":{"served_at":"2026-10-01T12:00:00Z","next_cursor":null,"limitations":[]}}`

func TestStatusSeparatesAvailabilityCoverageAndFreshness(t *testing.T) {
	var paths []string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/health/ready":
			fmt.Fprint(w, `{"status":"ready"}`)
		case "/v1/sources/coverage":
			fmt.Fprint(w, `{"data":[{"source_id":"<script>alert(1)</script>","territory":"050004","product":"municipal","public_state":"pending","coverage_status":"accepted_with_limitations","coverage_limitations":["<b>Ambito parziale</b>"],"quality":{"updating":{"state":"delayed","last_complete_check_at":"2026-10-01T10:00:00Z","limitations":["Controllo in ritardo"]}}}],"meta":{"served_at":"2026-10-01T12:00:00Z","next_cursor":"next+page/=&","limitations":["Copertura parziale"]}}`)
		default:
			t.Errorf("unexpected backend path: %s", r.URL.Path)
		}
	}))
	defer backend.Close()
	h, err := Handler(backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 200 {
		t.Fatalf("page: %d %s", w.Code, w.Body)
	}
	for _, want := range []string{"Disponibile", "In attesa", "Verificata con limiti", "In ritardo", "10:00:00 UTC", "12:00:00 UTC", "Copertura parziale", "Controllo in ritardo", "pagina corrente", `/?cursor=next%2Bpage%2F%3D%26`, "&lt;script&gt;", "&lt;b&gt;"} {
		if !strings.Contains(w.Body.String(), want) {
			t.Errorf("missing %q in %s", want, w.Body)
		}
	}
	if strings.Contains(w.Body.String(), "<script>") || len(paths) != 2 {
		t.Fatal("unsafe output or unexpected reads")
	}
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'none'") {
		t.Fatal("missing content policy")
	}
}

func TestCoverageFailuresAreNotEmptyResults(t *testing.T) {
	for _, tc := range []struct {
		name       string
		code       int
		body, want string
	}{
		{"empty", 200, emptyCoverage, "Nessuna fonte restituita"},
		{"limited", 429, `{"internal":"private-response"}`, "Accesso temporaneamente limitato"},
		{"expired", 410, `private-response`, "Pagina scaduta"},
		{"invalid", 400, `private-response`, "Richiesta non valida"},
		{"unavailable", 503, `private-response`, "temporaneamente non disponibili"},
		{"malformed", 200, `not json private-response`, "temporaneamente non disponibili"},
		{"missing data", 200, `{"meta":{"served_at":"2026-10-01T12:00:00Z"}}`, "non verificabile"},
		{"missing metadata", 200, `{"data":[]}`, "non verificabile"},
		{"null data", 200, `{"data":null,"meta":{"served_at":"2026-10-01T12:00:00Z"}}`, "non verificabile"},
		{"error envelope", 200, `{"data":[],"error":{"message":"private-response"},"meta":{"served_at":"2026-10-01T12:00:00Z"}}`, "non verificabile"},
		{"oversized", 200, strings.Repeat("x", maxResponseBytes+1), "temporaneamente non disponibili"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/health/ready" {
					w.WriteHeader(503)
					fmt.Fprint(w, `{"status":"unavailable"}`)
					return
				}
				w.WriteHeader(tc.code)
				fmt.Fprint(w, tc.body)
			}))
			defer backend.Close()
			h, err := Handler(backend.URL)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
			body := w.Body.String()
			if w.Code != 200 || !strings.Contains(body, tc.want) || !strings.Contains(body, "Non disponibile") {
				t.Fatalf("unexpected page: %d %s", w.Code, body)
			}
			if tc.name != "empty" && strings.Contains(body, "Nessuna fonte restituita") {
				t.Fatal("failure became empty result")
			}
			if strings.Contains(body, "private-response") || strings.Contains(body, backend.URL) {
				t.Fatal("private backend detail leaked")
			}
		})
	}
}

func TestFrontendSurvivesBackendFailureAndRejectsPrivateRoutes(t *testing.T) {
	backend := httptest.NewServer(http.NotFoundHandler())
	backend.Close()
	h, err := Handler(backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Non verificabile") || !strings.Contains(w.Body.String(), "temporaneamente non disponibili") {
		t.Fatal("backend failure prevented usable page")
	}
	for _, path := range []string{"/admin/", "/admin/assets/admin.css", "/mcp", "/v1/sources/coverage", "/proxy?url=https://example.test", "/assets/missing.css"} {
		w = httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Errorf("exposed %s: %d", path, w.Code)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/health/live", nil))
	if w.Code != 200 {
		t.Fatal("frontend liveness depends on backend")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/", nil))
	if w.Code != 405 {
		t.Fatal("mutation method accepted")
	}
}

func TestCursorIsPreservedAndQueriesCannotChooseBackend(t *testing.T) {
	reads := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		if r.URL.Path == "/health/ready" {
			fmt.Fprint(w, `{"status":"ready"}`)
			return
		}
		if r.URL.Path != "/v1/sources/coverage" || r.URL.Query().Get("cursor") != "opaque+token/&=" || len(r.URL.Query()) != 1 {
			t.Errorf("query changed: %s", r.URL)
		}
		fmt.Fprint(w, emptyCoverage)
	}))
	defer backend.Close()
	h, err := Handler(backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/?cursor=opaque%2Btoken%2F%26%3D", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "successiva alla prima") {
		t.Fatal("cursor navigation unavailable")
	}
	for _, path := range []string{"/?url=https://example.test", "/?source_id=private", "/?cursor=", "/?cursor=a&cursor=b", "/?cursor=a&other=b", "/?cursor=%ZZ"} {
		w = httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 400 {
			t.Errorf("unexpected query accepted: %s", path)
		}
	}
	if reads != 2 {
		t.Fatal("invalid queries reached backend")
	}
}

func TestOriginsRedirectsAndCancelledReads(t *testing.T) {
	for _, origin := range []string{"", "file:///tmp/data", "https://user:secret@example.test", "http://example.test/admin", "http://example.test?url=x", "http://example.test#x"} {
		if _, err := Handler(origin); err == nil {
			t.Errorf("unsafe origin accepted: %s", origin)
		}
	}
	redirectTargetReads := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirectTargetReads++ }))
	defer target.Close()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/admin/", http.StatusFound)
	}))
	defer backend.Close()
	h, err := Handler(backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if redirectTargetReads != 0 || !strings.Contains(w.Body.String(), "Non verificabile") {
		t.Fatal("redirect expanded public read boundary")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := server{origin: backend.URL, client: backend.Client()}
	if _, err := s.read(ctx, "/health/ready", new(any)); err == nil {
		t.Fatal("read ignored cancellation")
	}
}
