package httpapi

import (
	_ "embed"
	"net/http"

	contract "github.com/Balestrino/italian-weather-alert/api/public"
)

//go:embed public_docs.html
var publicDocumentation []byte

func publicDocumentationRoutes(mux *http.ServeMux) {
	for path, asset := range map[string]struct {
		contentType string
		body        []byte
	}{
		"/docs":                  {"text/html; charset=utf-8", publicDocumentation},
		"/openapi.json":          {"application/json", contract.OpenAPI},
		"/contratto.schema.json": {"application/schema+json", contract.Schema},
	} {
		mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", asset.contentType)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", "no-cache")
			_, _ = w.Write(asset.body)
		})
	}
	mux.HandleFunc("GET /docs/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusPermanentRedirect)
	})
}
