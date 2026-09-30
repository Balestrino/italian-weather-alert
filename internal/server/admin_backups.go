package server

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/backups"
	"net/http"
)

type AdminBackups interface {
	Report(context.Context) (backups.Report, error)
}

func adminBackupRoutes(mux *http.ServeMux, a AdminRuntime) {
	mux.HandleFunc("GET /admin/backups", func(w http.ResponseWriter, r *http.Request) {
		if a.Backups == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		report, err := a.Backups.Report(r.Context())
		if err != nil {
			adminFailure(w, registryUnavailable)
			return
		}
		if wantsAdminJSON(r) {
			reply(w, 200, report)
			return
		}
		adminRender(w, "backups", report)
	})
}
