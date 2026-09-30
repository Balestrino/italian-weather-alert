package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicDocumentation(t *testing.T) {
	public := Handler("public", nil)
	for path, contentType := range map[string]string{
		"/docs": "text/html", "/openapi.json": "application/json", "/contratto.schema.json": "application/schema+json",
	} {
		w := httptest.NewRecorder()
		public.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Type"), contentType) {
			t.Fatalf("%s: status=%d, content-type=%s", path, w.Code, w.Header().Get("Content-Type"))
		}
		if path != "/docs" && !json.Valid(w.Body.Bytes()) {
			t.Fatalf("%s is not valid JSON", path)
		}
		for _, admin := range []http.Handler{Handler("admin", nil), HandlerWithAdministration(nil, AdminRuntime{})} {
			w = httptest.NewRecorder()
			admin.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1:8081"+path, nil))
			if w.Code != http.StatusNotFound {
				t.Fatalf("admin exposes %s: %d", path, w.Code)
			}
		}
		w = httptest.NewRecorder()
		public.ServeHTTP(w, httptest.NewRequest("POST", path, nil))
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("POST %s: %d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	public.ServeHTTP(w, httptest.NewRequest("GET", "/docs/", nil))
	if w.Code != http.StatusPermanentRedirect || w.Header().Get("Location") != "/docs" {
		t.Fatal("documentation trailing slash does not redirect")
	}
}
