package backoffice

import (
	"context"
	"net/http"

	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

type adminArchiveRecovery interface {
	RecoverArchived(context.Context, interpretation.ArchiveRecovery) (interpretation.ArchiveRecovery, error)
	ArchiveRecoveries(context.Context, string) ([]interpretation.ArchiveRecovery, error)
}

func adminArchiveRecoveryRoutes(mux *http.ServeMux, a AdminRuntime) {
	mux.HandleFunc("POST /admin/sources/{id}/archive-recoveries", func(w http.ResponseWriter, r *http.Request) {
		var input interpretation.ArchiveRecovery
		if !adminDecode(w, r, &input) {
			return
		}
		if input.SourceID != r.PathValue("id") {
			adminFailure(w, registry.ErrInvalid)
			return
		}
		store, ok := a.Interpretation.(adminArchiveRecovery)
		if !ok {
			adminFailure(w, registryUnavailable)
			return
		}
		recovery, err := store.RecoverArchived(r.Context(), input)
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, recovery)
	})
	mux.HandleFunc("GET /admin/sources/{id}/archive-recoveries", func(w http.ResponseWriter, r *http.Request) {
		store, ok := a.Interpretation.(adminArchiveRecovery)
		if !ok {
			adminFailure(w, registryUnavailable)
			return
		}
		recoveries, err := store.ArchiveRecoveries(r.Context(), r.PathValue("id"))
		if err != nil {
			adminFailure(w, err)
			return
		}
		adminSuccess(w, r, recoveries)
	})
}
