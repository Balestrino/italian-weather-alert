package server

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/notifications"
)

type notificationReportFake struct{ fail bool }

func (f *notificationReportFake) Report(context.Context) (notifications.Report, error) {
	if f.fail {
		return notifications.Report{}, errors.New("smtp-password-secret")
	}
	return notifications.Report{OpenIncidents: 1, PendingMessages: 1, TotalMessages: 2, Limit: 100, Messages: []notifications.Delivery{{Message: notifications.Message{Scope: "<script>unsafe</script>", Category: "backup_failed"}, ErrorCode: "smtp_delivery_failed"}}}, nil
}
func TestPrivateNotificationReport(t *testing.T) {
	f := &notificationReportFake{}
	h := HandlerWithAdministration(nil, AdminRuntime{Notifications: f})
	for _, accept := range []string{"text/html", "application/json"} {
		r := httptest.NewRequest("GET", "http://127.0.0.1/admin/notifications", nil)
		r.Header.Set("Accept", accept)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 || !strings.Contains(w.Body.String(), "smtp_delivery_failed") || strings.Contains(w.Body.String(), "<script>") || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("private report or escaping failed")
		}
	}
	f.fail = true
	r := httptest.NewRequest("GET", "http://127.0.0.1/admin/notifications", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 503 || strings.Contains(w.Body.String(), "smtp-password-secret") {
		t.Fatal("failed report exposed private error")
	}
}
