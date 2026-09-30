package server

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backups"
)

type backupReportFake struct{ fail bool }

func (f *backupReportFake) Report(context.Context) (backups.Report, error) {
	if f.fail {
		return backups.Report{}, errors.New("backup-password-secret")
	}
	return backups.Report{Total: 1, Runs: []backups.Run{{ID: "<script>unsafe</script>", ErrorCode: "transfer_failed"}}}, nil
}
func TestPrivateBackupReport(t *testing.T) {
	f := &backupReportFake{}
	h := HandlerWithAdministration(nil, AdminRuntime{Backups: f})
	for _, accept := range []string{"text/html", "application/json"} {
		r := httptest.NewRequest("GET", "http://127.0.0.1/admin/backups", nil)
		r.Header.Set("Accept", accept)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 || !strings.Contains(w.Body.String(), "transfer_failed") || strings.Contains(w.Body.String(), "<script>") || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("private report or escaping failed")
		}
	}
	f.fail = true
	r := httptest.NewRequest("GET", "http://127.0.0.1/admin/backups", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 503 || strings.Contains(w.Body.String(), "backup-password-secret") {
		t.Fatal("failed report exposed private error")
	}
}
