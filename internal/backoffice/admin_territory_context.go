package backoffice

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

type territoryReturnKey struct{}
type sourceTerritoryReader interface {
	SourceTerritory(context.Context, string) (string, string, error)
}

func adminTerritorialContext(next http.Handler, a AdminRuntime) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/admin/sources/") {
			next.ServeHTTP(w, r)
			return
		}
		requested := r.URL.Query().Get("return_to")
		reader, ok := a.Registry.(sourceTerritoryReader)
		if !ok || a.Territories == nil {
			if requested != "" {
				http.Error(w, "Contesto territoriale non valido", 400)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.EscapedPath(), "/admin/sources/")
		part, _, _ := strings.Cut(path, "/")
		id, err := url.PathUnescape(part)
		if err != nil {
			http.Error(w, "Fonte non valida", 400)
			return
		}
		region, istat, err := reader.SourceTerritory(r.Context(), id)
		if err != nil {
			if requested != "" {
				http.Error(w, "Contesto territoriale non valido", 400)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		back := "/admin/regions/" + url.PathEscape(region)
		if istat != "" {
			back += "/municipalities/" + url.PathEscape(istat)
		}
		back += "?tab=configuration"
		if (requested != "" && requested != back) || (r.URL.Query().Get("region") != "" && r.URL.Query().Get("region") != region) || (r.URL.Query().Get("municipality") != "" && r.URL.Query().Get("municipality") != istat) || len(r.URL.Query()["return_to"]) > 1 {
			http.Error(w, "Contesto territoriale non valido", 400)
			return
		}
		*r = *r.WithContext(context.WithValue(r.Context(), territoryReturnKey{}, back))
		next.ServeHTTP(w, r)
	})
}
