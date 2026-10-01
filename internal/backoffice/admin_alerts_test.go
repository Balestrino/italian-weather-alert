package backoffice

import (
	"context"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publicquery"
	"github.com/Balestrino/italian-weather-alert/internal/backend/transport/httpapi"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type alertsFixture struct{ fail bool }

func (f alertsFixture) Alerts(context.Context, time.Time) (operations.Alerts, error) {
	if f.fail {
		return operations.Alerts{}, errors.New("private-password")
	}
	return operations.Alerts{ObservedAt: time.Now(), Days: []operations.AlertDay{{Label: "Oggi", Date: "23/09/2026", Bulletins: []operations.AlertBulletin{{Product: "criticality", Source: "CFR", URL: "https://example.test", Issued: "<script>alert(1)</script>", Validity: []string{"23 settembre 2026"}, AcquiredAt: time.Now()}}, Regional: []publicquery.RegionalWarning{{OfficialRiskLabel: "Temporali", Zone: "A1", Level: "yellow", Status: "current"}}, Measures: []publicquery.Measure{{Action: "Chiusura parchi", MunicipalityISTAT: "050004", Status: "current"}}}, {Label: "Domani", Date: "24/09/2026"}}, UndatedBulletins: []operations.AlertBulletin{{Product: "monitoring", URL: "https://example.test"}}}, nil
}
func TestAdminAlerts(t *testing.T) {
	for _, fail := range []bool{false, true} {
		h := HandlerWithAdministration(nil, AdminRuntime{Alerts: alertsFixture{fail: fail}, TailscaleOrigin: "https://iwa.tail123.ts.net"})
		for _, base := range []string{"http://127.0.0.1", "https://iwa.tail123.ts.net"} {
			r := httptest.NewRequest("GET", base+"/admin/alerts", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			body := w.Body.String()
			if fail {
				if w.Code != 503 || strings.Contains(body, "private-password") {
					t.Fatal(body)
				}
			} else {
				if w.Code != 200 {
					t.Fatal(body)
				}
				for _, s := range []string{"Allerte meteo ufficiali", "Provvedimenti comunali", "23/09/2026", "24/09/2026", "Validità da verificare", "Giallo", "Chiusura parchi", "&lt;script&gt;"} {
					if !strings.Contains(body, s) {
						t.Fatal("missing", s)
					}
				}
				if strings.Contains(body, "<script>alert") {
					t.Fatal("unescaped")
				}
			}
			r.Header.Set("Origin", "https://evil.test")
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 403 {
				t.Fatal(w.Code)
			}
		}
	}
	w := httptest.NewRecorder()
	httpapi.Handler(nil).ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1/admin/alerts", nil))
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
}
