package backoffice

import (
	"context"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

type territoryOverviewFixture struct {
	AdminTerritories
	fail    bool
	regions []operations.RegionSummary
}

func (f territoryOverviewFixture) TerritorialOverview(context.Context, time.Time) (operations.TerritorialOverview, error) {
	if f.fail {
		return operations.TerritorialOverview{}, errors.New("private database detail")
	}
	if f.regions != nil {
		return operations.TerritorialOverview{ObservedAt: time.Now(), Regions: f.regions}, nil
	}
	n := int64(273)
	zero := int64(0)
	return operations.TerritorialOverview{ObservedAt: time.Now(), Regions: []operations.RegionSummary{{Region: domain.Region{Code: "09", Name: "Toscana", Revision: 1, Enabled: true, Configuration: domain.RegionConfiguration{MunicipalityDataset: "fixture"}}, Municipalities: &n, EnabledMunicipalities: &zero, Sources: &zero, Collecting: &zero, Issues: &zero}, {Region: domain.Region{Code: "03", Name: "Lombardia"}}}}, nil
}

func TestRegionalEntryOrdering(t *testing.T) {
	region := func(code, name string, enabled bool, count *int64) operations.RegionSummary {
		total := int64(100)
		return operations.RegionSummary{Region: domain.Region{Code: code, Name: name, Enabled: enabled, Configuration: domain.RegionConfiguration{MunicipalityDataset: "fixture"}}, Municipalities: &total, EnabledMunicipalities: count}
	}
	zero, two, ten := int64(0), int64(2), int64(10)
	f := territoryOverviewFixture{regions: []operations.RegionSummary{
		region("01", "Alfa", false, &ten),
		region("02", "Beta", true, &two),
		region("03", "Gamma", true, &ten),
		region("04", "Alfa", true, &two),
		region("05", "Senza conteggio", true, nil),
		region("06", "Zero", true, &zero),
		region("07", "Ignoto", false, nil),
		region("08", "Zero", false, &zero),
		region("09", "Alfa", true, &two),
	}}
	h := HandlerWithAdministration(nil, AdminRuntime{Territories: f})
	links := regexp.MustCompile(`href="/admin/regions/([^"?/]+)"`)
	for _, tt := range []struct {
		path string
		want []string
	}{
		{"/admin/", []string{"03", "04", "09", "02", "06", "05", "01", "08", "07"}},
		{"/admin/regions", []string{"03", "04", "09", "02", "06", "05", "01", "08", "07"}},
		{"/admin/regions?state=all", []string{"03", "04", "09", "02", "06", "05", "01", "08", "07"}},
		{"/admin/regions?state=enabled", []string{"03", "04", "09", "02", "06", "05"}},
		{"/admin/regions?state=disabled", []string{"01", "08", "07"}},
	} {
		t.Run(tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1"+tt.path, nil))
			var got []string
			for _, match := range links.FindAllStringSubmatch(w.Body.String(), -1) {
				got = append(got, match[1])
			}
			if w.Code != 200 || !slices.Equal(got, tt.want) {
				t.Fatalf("status %d, order %v; want %v", w.Code, got, tt.want)
			}
			for _, text := range []string{"Comuni abilitati / totali", "10 / 100", "0 / 100", "Non disponibile / 100"} {
				if !strings.Contains(w.Body.String(), text) {
					t.Fatalf("missing %q", text)
				}
			}
		})
	}
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

func TestTerritorialControlReturnScope(t *testing.T) {
	for _, raw := range []string{"", "/admin/regions", "/admin/regions?state=disabled", "/admin/regions/09?tab=configuration", "/admin/regions/09?tab=municipalities&search=Comune&province=PR&coverage=none&after=cursor"} {
		if _, err := territoryControlReturn(raw, "09", "municipalities"); err != nil {
			t.Fatalf("valid return %q: %v", raw, err)
		}
	}
	for _, raw := range []string{"https://example.test/admin/regions", "//example.test/admin/regions", "/admin/regions/03?tab=municipalities", "/admin/regions?state=unknown", "/admin/regions?state=all&state=disabled", "/admin/regions?extra=value", "/admin/regions/09?tab=municipalities&enabled=true", "/admin/regions/09?tab=results", "/admin/regions/09?tab=configuration&search=Comune", "/admin/regions/09?tab=municipalities&coverage=wrong", "/admin/regions#fragment", "/admin/%72egions"} {
		if _, err := territoryControlReturn(raw, "09", "municipalities"); err == nil {
			t.Fatalf("invalid return accepted %q", raw)
		}
	}
}
