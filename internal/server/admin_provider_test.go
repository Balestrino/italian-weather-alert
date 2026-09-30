package server

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type providerFixture struct{ resumed, cleared int }

func (p *providerFixture) Gates(context.Context) ([]processing.GateState, error) {
	return []processing.GateState{{Scope: "opaque", State: "held", Reason: "quota"}}, nil
}
func (p *providerFixture) Rejections(context.Context) ([]processing.Rejection, error) {
	return []processing.Rejection{}, nil
}
func (p *providerFixture) ResumeGate(_ context.Context, scope, model, actor string, _ time.Time) error {
	if actor == "" {
		return processing.ErrInvalid
	}
	p.resumed++
	return nil
}
func (p *providerFixture) ClearRejection(_ context.Context, scope, model, configuration, hash, actor string, _ time.Time) error {
	if actor == "" {
		return processing.ErrInvalid
	}
	p.cleared++
	return nil
}
func TestProviderRecoveryStaysPrivate(t *testing.T) {
	p := &providerFixture{}
	handler := HandlerWithAdministration(nil, AdminRuntime{Provider: p})
	for _, path := range []string{"/admin/inference/gates", "/admin/inference/rejections", "/admin/inference/resume", "/admin/inference/clear-rejection"} {
		r := httptest.NewRecorder()
		Handler("public", nil).ServeHTTP(r, httptest.NewRequest("GET", "http://127.0.0.1"+path, nil))
		if r.Code != 404 {
			t.Fatal("provider controls public")
		}
	}
	for _, action := range []string{"resume", "clear-rejection"} {
		request := httptest.NewRequest("POST", "http://127.0.0.1/admin/inference/"+action, strings.NewReader(`{"scope":"opaque","model":"*","actor":"operator"}`))
		request.Header.Set("Content-Type", "application/json")
		r := httptest.NewRecorder()
		handler.ServeHTTP(r, request)
		if r.Code != 200 {
			t.Fatalf("recovery failed: %d %s", r.Code, r.Body.String())
		}
	}
	if p.resumed != 1 || p.cleared != 1 {
		t.Fatal("recovery operation missing")
	}
}
