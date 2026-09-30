package server

import (
	"context"
	"net/http"

	"github.com/Balestrino/italian-weather-alert/internal/notifications"
)

type AdminNotifications interface {
	Report(context.Context) (notifications.Report, error)
}

func adminNotificationRoutes(mux *http.ServeMux, a AdminRuntime) {
	mux.HandleFunc("GET /admin/notifications", func(w http.ResponseWriter, r *http.Request) {
		if a.Notifications == nil {
			adminFailure(w, registryUnavailable)
			return
		}
		report, err := a.Notifications.Report(r.Context())
		if err != nil {
			adminFailure(w, registryUnavailable)
			return
		}
		if wantsAdminJSON(r) {
			reply(w, 200, report)
			return
		}
		adminRender(w, "notifications", report)
	})
}
