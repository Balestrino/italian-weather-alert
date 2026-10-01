package backoffice

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/backend/transport/httpapi"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/trialcost"
)

type trialCostFixture struct {
	reportCalls, recordCalls, proposeCalls, listCalls int
	incomplete                                        bool
}

func (f *trialCostFixture) RecordInput(_ context.Context, campaign, actor string, input trialcost.Input) (trialcost.Input, error) {
	f.recordCalls++
	input.CampaignID, input.Actor = campaign, actor
	return input, nil
}
func (f *trialCostFixture) Report(_ context.Context, campaign string, through time.Time) (trialcost.Report, error) {
	f.reportCalls++
	return trialcost.Report{CampaignID: campaign, Through: through}, nil
}
func (f *trialCostFixture) Propose(_ context.Context, campaign, actor string, through time.Time, evidence registry.Evidence) (trialcost.Proposal, error) {
	f.proposeCalls++
	if f.incomplete {
		return trialcost.Proposal{}, trialcost.ErrIncomplete
	}
	return trialcost.Proposal{CampaignID: campaign, Actor: actor, Through: through, Evidence: evidence}, nil
}
func (f *trialCostFixture) Proposals(context.Context, string) ([]trialcost.Proposal, error) {
	f.listCalls++
	return []trialcost.Proposal{}, nil
}

func TestTrialCostRoutesAreLocalAndPreserveIncompleteGate(t *testing.T) {
	fixture := &trialCostFixture{}
	admin := HandlerWithAdministration(nil, AdminRuntime{TrialCosts: fixture})
	through := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339Nano)
	paths := []string{
		"/admin/observation-campaigns/trial/cost-report?through=" + through,
		"/admin/observation-campaigns/trial/cost-inputs",
		"/admin/observation-campaigns/trial/budget-proposals",
	}
	for _, path := range paths {
		recorder := httptest.NewRecorder()
		httpapi.Handler(nil).ServeHTTP(recorder, httptest.NewRequest("GET", "http://127.0.0.1"+path, nil))
		if recorder.Code != 404 {
			t.Fatalf("public trial cost route %s = %d", path, recorder.Code)
		}
	}
	recorder := httptest.NewRecorder()
	admin.ServeHTTP(recorder, httptest.NewRequest("GET", "http://127.0.0.1/admin/observation-campaigns/trial/cost-report", nil))
	if recorder.Code != 400 || fixture.reportCalls != 0 {
		t.Fatal("missing through boundary reached report")
	}
	recorder = httptest.NewRecorder()
	request := httptest.NewRequest("GET", "http://127.0.0.1/admin/observation-campaigns/trial/cost-report?through="+through, nil)
	request.Header.Set("Accept", "application/json")
	admin.ServeHTTP(recorder, request)
	if recorder.Code != 200 || fixture.reportCalls != 1 || !strings.Contains(recorder.Body.String(), `"campaign_id":"trial"`) {
		t.Fatalf("report route: %d %s", recorder.Code, recorder.Body)
	}
	fixture.incomplete = true
	body := `{"actor":"operator","through":"` + through + `","evidence":{"url":"https://evidence.example/report","locator":"review","observed_at":"` + time.Now().UTC().Format(time.RFC3339Nano) + `"}}`
	request = httptest.NewRequest("POST", "http://127.0.0.1/admin/observation-campaigns/trial/budget-proposals", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	recorder = httptest.NewRecorder()
	admin.ServeHTTP(recorder, request)
	if recorder.Code != 409 || fixture.proposeCalls != 1 || !strings.Contains(recorder.Body.String(), "trial_cost_report_incomplete") {
		t.Fatalf("incomplete gate: %d %s", recorder.Code, recorder.Body)
	}
}
