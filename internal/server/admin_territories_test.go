package server

import (
	"context"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/domain"
	"github.com/Balestrino/italian-weather-alert/internal/operations"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type territoryOverviewFixture struct {
	AdminTerritories
	fail bool
}

func (f territoryOverviewFixture) TerritorialOverview(context.Context, time.Time) (operations.TerritorialOverview, error) {
	if f.fail {
		return operations.TerritorialOverview{}, errors.New("private database detail")
	}
	n := int64(273)
	zero := int64(0)
	return operations.TerritorialOverview{ObservedAt: time.Now(), Regions: []operations.RegionSummary{{Region: domain.Region{Code: "09", Name: "Toscana", Revision: 1, Enabled: true, Configuration: domain.RegionConfiguration{MunicipalityDataset: "fixture"}}, Municipalities: &n, Sources: &zero, Collecting: &zero, Issues: &zero}, {Region: domain.Region{Code: "03", Name: "Lombardia"}}}}, nil
}
func TestRegionalEntryAndLegacyOperations(t *testing.T) {
	for _, fail := range []bool{false, true} {
		h := HandlerWithAdministration(nil, AdminRuntime{Territories: territoryOverviewFixture{fail: fail}, Operations: &operationFixture{}})
		for _, path := range []string{"/admin/", "/admin/regions", "/admin/regions?state=enabled", "/admin/regions?state=disabled", "/admin/operations/", "/admin/operations/jobs"} {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1"+path, nil))
			if w.Code != 200 || strings.Contains(w.Body.String(), "private database detail") {
				t.Fatal(path, w.Code, w.Body)
			}
			if strings.HasPrefix(path, "/admin/regions") || path == "/admin/" {
				if fail && !strings.Contains(w.Body.String(), "Regioni non disponibili") {
					t.Fatal(w.Body)
				}
				if !fail && path == "/admin/regions" && (!strings.Contains(w.Body.String(), "273") || !strings.Contains(w.Body.String(), "Anagrafica da adottare")) {
					t.Fatal(w.Body)
				}
			}
			if !fail && strings.HasSuffix(path, "state=enabled") && strings.Contains(w.Body.String(), "Lombardia") {
				t.Fatal("disabled region shown")
			}
			if !fail && strings.HasSuffix(path, "state=disabled") && strings.Contains(w.Body.String(), ">Toscana<") {
				t.Fatal("enabled region shown")
			}
		}
	}
}
