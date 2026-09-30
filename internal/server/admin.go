package server

import (
	"github.com/Balestrino/italian-weather-alert/internal/config"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// adminBoundary protects every administrative route, including future mutation
// handlers. Listener binding (or the Compose loopback port mapping) remains the
// access boundary; Host/Origin checks additionally protect an operator's browser
// from DNS rebinding and cross-origin access. Forwarded headers are not trusted.
func adminBoundary(next http.Handler) http.Handler {
	return adminBoundaryWithOrigin(next, "")
}

// The loopback port mapping and tailnet ACLs remain the access boundary.
// Trust only the configured external origin; forwarded headers grant no access.
func adminBoundaryWithOrigin(next http.Handler, origin string) http.Handler {
	externalHost := ""
	if config.ValidAdminTailscaleOrigin(origin) {
		externalHost = strings.TrimPrefix(origin, "https://")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'self'; script-src 'self'; frame-ancestors 'none'; form-action 'self'; base-uri 'none'")
		allowed := localAdminHost(r.Host) && adminOriginMatches(r)
		if externalHost != "" && r.Host == externalHost {
			origins := r.Header.Values("Origin")
			allowed = len(origins) == 0 || (len(origins) == 1 && origins[0] == origin)
		}
		if !allowed || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			http.Error(w, "administrative access requires an allowed origin", http.StatusForbidden)
			return
		}
		if strings.Contains(r.Header.Get("Accept"), "text/html") && !wantsAdminJSON(r) {
			w = &adminBrowserWriter{w, r}
		}
		next.ServeHTTP(w, r)
	})
}

func localAdminHost(authority string) bool {
	host := authority
	if strings.Contains(authority, ":") {
		if strings.HasPrefix(authority, "[") && strings.HasSuffix(authority, "]") {
			ip := net.ParseIP(authority[1 : len(authority)-1])
			return ip != nil && ip.IsLoopback()
		}
		var err error
		host, _, err = net.SplitHostPort(authority)
		if err != nil {
			return false
		}
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func adminOriginMatches(r *http.Request) bool {
	origins := r.Header.Values("Origin")
	if len(origins) == 0 {
		return true // Local CLI clients and direct browser navigation.
	}
	if len(origins) != 1 {
		return false
	}
	origin, err := url.Parse(origins[0])
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return err == nil && origin.Scheme == scheme && origin.Host == r.Host &&
		origin.User == nil && origin.Path == "" && origin.RawQuery == "" && !origin.ForceQuery && origin.Fragment == ""
}

func adminRoutes(mux *http.ServeMux) {
	adminAssetRoutes(mux)
	mux.HandleFunc("GET /admin/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		renderUI(w, landingTemplates, "home", nil)
	})
	mux.HandleFunc("GET /admin/status", func(w http.ResponseWriter, r *http.Request) {
		reply(w, http.StatusOK, map[string]string{"role": "admin", "access": "local"})
	})
}
