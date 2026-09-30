//go:build integration

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/domain"
	"github.com/Balestrino/italian-weather-alert/internal/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/linking"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

type lifecycleAdapter struct {
	calls    int
	relevant bool
}

func (a *lifecycleAdapter) Name() string { return "fixture" }
func (a *lifecycleAdapter) Complete(context.Context, inference.Request) (inference.Response, error) {
	a.calls++
	if a.relevant {
		return inference.Response{ID: "test", Model: "qwen3.8-27b", Content: `{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"Avviso comunale"}`}, nil
	}
	return inference.Response{ID: "test", Model: "qwen3.8-27b", Content: `{"relevant":false,"reason_code":"not_relevant","evidence_quote":"Avviso comunale"}`}, nil
}

// These outputs exercise lifecycle gates, not semantic source acceptance.
type lifecycleDownstream struct{ stage string }

func (a lifecycleDownstream) Name() string { return "fixture" }
func (a lifecycleDownstream) Complete(_ context.Context, request inference.Request) (inference.Response, error) {
	response := inference.Response{ID: "fixture", Model: "qwen3.8-27b"}
	if a.stage == "linking" {
		response.Content = `{"status":"no_relation","reason_code":"insufficient_evidence","relation":null,"candidate_index":null,"current_evidence_indices":[],"candidate_evidence_indices":[]}`
		return response, nil
	}
	var raw string
	if err := json.Unmarshal(request.Messages[1].Content, &raw); err != nil {
		return response, err
	}
	var input struct {
		Sections []classification.ContentSection `json:"sections"`
		Segment  struct {
			ResourceURL string `json:"resource_url"`
			Text        string `json:"text"`
		} `json:"segment"`
		Window struct {
			Ordinal     int    `json:"ordinal"`
			ResourceURL string `json:"resource_url"`
			Text        string `json:"text"`
		} `json:"window"`
	}
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		return response, err
	}
	resourceURL := input.Segment.ResourceURL
	if resourceURL == "" {
		resourceURL = input.Window.ResourceURL
	}
	if resourceURL == "" && len(input.Sections) > 0 {
		resourceURL = input.Sections[0].ResourceURL
	}
	if input.Window.Ordinal > 0 {
		body, err := json.Marshal(map[string]any{
			"envelope_version": "compact-evidence-v1", "window_ordinal": input.Window.Ordinal,
			"evidence": []string{"Avviso comunale"},
			"measures": []any{map[string]any{"kind": "observation", "subject": "Avviso comunale", "place": nil, "valid_from": nil, "valid_until": nil, "evidence_refs": map[string]any{"kind": []int{0}, "subject": []int{0}, "place": []int{}, "valid_from": []int{}, "valid_until": []int{}}}},
		})
		response.Content = string(body)
		return response, err
	}
	evidence := []map[string]any{{"field": "kind", "resource_url": resourceURL, "page": nil, "quote": "Avviso comunale"}, {"field": "subject", "resource_url": resourceURL, "page": nil, "quote": "Avviso comunale"}}
	body, err := json.Marshal(map[string]any{"measures": []any{map[string]any{"kind": "observation", "subject": "Avviso comunale", "place": nil, "valid_from": nil, "valid_until": nil, "indeterminate_fields": []string{"place", "valid_from", "valid_until"}, "evidence": evidence}}})
	response.Content = string(body)
	return response, err
}

func TestAdminAcceptanceAndValidatedInterpretationRecovery(t *testing.T) { testAdminRecovery(t, false) }
func TestAdminGuidedAcceptanceAndRecovery(t *testing.T)                  { testAdminRecovery(t, true) }
func testAdminRecovery(t *testing.T, guided bool) {
	ctx := context.Background()
	pool := adminTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, acquisition.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, linking.MigrateSemantic, interpretation.Migrate, domain.Migrate, domain.MigrateTemporal, domain.MigrateQuality, publicquery.Migrate} {
		must(m(ctx, pool))
	}
	reg := registry.New(pool)
	queue := jobs.New(pool)
	scheduler := interpretation.New(pool, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, false)
	retained := documents.New(pool, adminObjects{})
	host := httptest.NewServer(HandlerWithAdministration(nil, AdminRuntime{Registry: reg, Interpretation: scheduler}))
	defer host.Close()
	post := func(action string, payload any, want int) []byte {
		t.Helper()
		b, err := json.Marshal(payload)
		must(err)
		media := "application/json"
		if guided {
			var payloadMap map[string]any
			must(json.Unmarshal(b, &payloadMap))
			if action == "reprocess-interpretation" {
				payloadMap["from"] = time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
			}
			values := url.Values{}
			flattenGuided(values, "", payloadMap)
			b = []byte(values.Encode())
			media = "application/x-www-form-urlencoded"
		}
		req, err := http.NewRequest("POST", host.URL+"/admin/sources/s/"+action, bytes.NewReader(b))
		must(err)
		req.Header.Set("Content-Type", media)
		req.Header.Set("Accept", "application/json")
		res, err := http.DefaultClient.Do(req)
		must(err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		must(err)
		if res.StatusCode != want {
			t.Fatalf("%s: got %d want %d: %s", action, res.StatusCode, want, body)
		}
		return body
	}
	now := time.Now().UTC()
	evidence := registry.Evidence{URL: "https://source.example/report", Locator: "synthetic lifecycle fixture", ObservedAt: now}
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "Comune", OfficialURL: "https://source.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "official", URL: "https://source.example"}))
	cfg := registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/notices"}, AccessMethod: "fixture", Attribution: "Comune", Provenance: &evidence, Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true, PublicationPermitted: true}}
	must(reg.CreateSource(ctx, registry.Source{ID: "s", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "050004"}, cfg, "test"))
	must(reg.RecordPreview(ctx, "s", 1, "test", evidence))
	must(reg.EnableCollection(ctx, "s", 1, "test"))
	retain := func(id string) documents.Version {
		t.Helper()
		u := "https://source.example/notices/" + id
		v, e := retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: "s", Configuration: 1, URL: u, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: u, Role: "original", Required: true, SourceID: "s", Configuration: 1, MediaType: "text/html", Bytes: []byte("Avviso comunale senza misure meteo.")}}})
		must(e)
		return v
	}
	first := retain("trial")
	public := publicquery.New(pool)
	visible := func(count int, unreliable bool) {
		t.Helper()
		at := time.Now().UTC()
		r, e := public.Search(ctx, publicquery.SearchQuery{QueryTime: publicquery.QueryTime{EvaluationTime: at, KnownAt: at}, Kind: "document"})
		must(e)
		if len(r.Documents) != count {
			t.Fatalf("visible documents: %d want %d", len(r.Documents), count)
		}
		for _, d := range r.Documents {
			if unreliable && (d.Quality.Interpretation.State != "unreliable" || d.OfficialURL == "") {
				t.Fatalf("missing defect warning or official link: %#v", d)
			}
		}
	}
	action := map[string]any{"revision": 1, "actor": "operator"}
	visible(0, false)
	post("enable-public", action, 409)
	acceptance := registry.Acceptance{Report: evidence, PeriodStart: now.Add(-8 * 24 * time.Hour), PeriodEnd: now, Sections: cfg.Sections, ExtractionVerified: true, UpdatesVerified: true, AttachmentsVerified: true, ScannedAttachmentsVerified: true, HistoryVerified: true, FailureBehaviorVerified: true, InterfacesEquivalent: true, CoverageStatus: "accepted_declared_scope", CoverageLimitations: []string{}}
	acceptance.KnownOmissions = 1
	post("accept", map[string]any{"revision": 1, "actor": "operator", "acceptance": acceptance}, 400)
	acceptance.KnownOmissions = 0
	post("accept", map[string]any{"revision": 1, "actor": "operator", "acceptance": acceptance}, 200)
	visible(0, false)
	post("enable-public", action, 409)
	post("review-acceptance", map[string]any{"revision": 1, "actor": "release-reviewer", "evidence": registry.Evidence{URL: "https://source.example/release", Locator: "final acceptance review", ObservedAt: time.Now().UTC()}}, 200)
	post("enable-public", action, 200)
	visible(1, false)
	// A subsequent acquisition is public without another acceptance/notice gate.
	second := retain("automatic")
	visible(2, false)
	scheduled, err := scheduler.Automatic(ctx, interpretation.AcquisitionEvent{DocumentVersionID: first.ID, EvidenceHash: first.Hash, ContentChanged: true, At: time.Now()})
	must(err)
	if !scheduled {
		t.Fatal("ordinary job not scheduled")
	}
	post("suspend-interpretation", map[string]any{"revision": 2, "actor": "operator", "defect": evidence}, 404)
	post("suspend-interpretation", map[string]any{"revision": 1, "actor": "operator", "defect": evidence}, 200)
	post("suspend-interpretation", map[string]any{"revision": 1, "actor": "operator", "defect": evidence}, 409)
	state, err := registry.New(pool).State(ctx, "s")
	must(err)
	if state.InterpretationSuspendedAt == nil || !state.CollectionEnabled || !state.PublicEnabled {
		t.Fatalf("suspension changed other controls: %#v", state)
	}
	visible(2, true)
	// The acquisition scheduler still claims this source and new bytes are retained.
	schedule := acquisition.NewScheduleStore(pool)
	must(schedule.SyncEnabled(ctx, time.Now()))
	claim, err := schedule.ClaimDue(ctx, "worker", time.Now(), time.Minute)
	must(err)
	third := retain("during-suspension")
	must(schedule.Finish(ctx, claim, acquisition.CheckOutcome{SourceID: "s", Configuration: 1, StartedAt: claim.StartedAt, FinishedAt: time.Now(), Reachable: true, ContentRecognized: true, Complete: true}))
	visible(3, true)
	scheduled, err = scheduler.Automatic(ctx, interpretation.AcquisitionEvent{DocumentVersionID: third.ID, EvidenceHash: third.Hash, ContentChanged: true, At: time.Now()})
	must(err)
	if scheduled {
		t.Fatal("suspended source scheduled inference")
	}
	// Already queued jobs are blocked before a provider call, preserving an outcome.
	called := false
	guard := scheduler.Guard(func(context.Context, jobs.Job) (jobs.Result, error) { called = true; return jobs.Result{}, nil })
	worker := &jobs.Worker{Store: queue, Queue: inference.Queue, ID: "blocked-worker", Lease: time.Minute, PollInterval: time.Second, Handlers: map[string]jobs.Handler{classification.Kind: guard}}
	worked, err := worker.RunOne(ctx)
	must(err)
	if !worked || called {
		t.Fatal("queued job bypassed suspension")
	}
	history, err := reg.InterpretationHistory(ctx, "s")
	must(err)
	correction := evidence
	correction.ObservedAt = time.Now().UTC()
	recovery := interpretation.Recovery{SuspensionID: history[0].ID, Correction: correction, Validation: correction, Verified: true, ReprocessingID: "not-real"}
	resume := func(code int) {
		post("resume-interpretation", map[string]any{"revision": 1, "actor": "operator", "recovery": recovery}, code)
	}
	resume(409)
	body := post("reprocess-interpretation", action, 200)
	var selection struct {
		ID    string `json:"reprocessing_id"`
		Count int    `json:"selected_versions"`
	}
	must(json.Unmarshal(body, &selection))
	if selection.Count != 3 {
		t.Fatalf("selection: %s", body)
	}
	recovery.ReprocessingID = selection.ID
	recovery.Validation.ObservedAt = time.Now().UTC()
	resume(409)
	process := processing.New(pool)
	catalog, err := classification.RegisterCatalog(ctx, process, "fixture", "qwen3.8-27b", time.Now())
	must(err)
	adapter := &lifecycleAdapter{}
	runner := &classification.Runner{Documents: retained, OCR: ocr.NewStore(pool), Processing: process, Results: classification.NewStore(pool), Adapter: adapter, Model: "qwen3.8-27b", ConfigurationVersion: catalog.ConfigurationVersionID}
	worker.Handlers[classification.Kind] = scheduler.Guard(scheduler.ClassificationHandler(runner.Handler()))
	for range 3 {
		worked, err = worker.RunOne(ctx)
		must(err)
		if !worked {
			t.Fatal("selected job missing")
		}
	}
	if adapter.calls != 3 {
		t.Fatalf("reprocessing calls: %d", adapter.calls)
	}
	visible(3, true)
	recovery.Validation.ObservedAt = time.Now().UTC()
	recovery.UnsupportedAssertions = 1
	resume(400)
	recovery.UnsupportedAssertions = 0
	recovery.SuspensionID++
	resume(409)
	recovery.SuspensionID--
	// Resumption also waits for jobs already in flight, even outside this selection.
	_, err = classification.Enqueue(ctx, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, classification.Payload{DocumentVersionID: second.ID, Workload: "evaluation", SelectionID: "in-flight-fixture"}, time.Now())
	must(err)
	inFlight, err := queue.Claim(ctx, inference.Queue, "in-flight", time.Now(), time.Minute)
	must(err)
	recovery.Validation.ObservedAt = time.Now().UTC()
	resume(409)
	must(queue.Complete(ctx, inFlight, jobs.Result{Payload: json.RawMessage(`{"status":"interpretation_suspended"}`)}, time.Now()))
	recovery.Validation.ObservedAt = time.Now().UTC()
	resume(200)
	state, err = reg.State(ctx, "s")
	must(err)
	if state.InterpretationSuspendedAt != nil || !state.CollectionEnabled || !state.PublicEnabled {
		t.Fatalf("bad recovery state: %#v", state)
	}
	history, err = reg.InterpretationHistory(ctx, "s")
	must(err)
	if len(history) != 1 || history[0].ResumedAt == nil || history[0].Recovery == nil {
		t.Fatal("recovery audit missing")
	}
	scheduled, err = scheduler.Automatic(ctx, interpretation.AcquisitionEvent{DocumentVersionID: second.ID, EvidenceHash: second.Hash, ContentChanged: true, At: time.Now()})
	must(err)
	if !scheduled {
		t.Fatal("ordinary interpretation did not resume")
	}
	post("suspend-public", action, 200)
	visible(0, false)
	post("enable-public", action, 200)
	visible(3, false)
	post("resume-interpretation", map[string]any{"revision": 1, "actor": "operator", "recovery": recovery, "measures": []any{}}, 400)
	// A new suspension cannot reuse an earlier correction request. Relevant
	// classifications alone do not satisfy recovery: extraction/linking must finish.
	post("suspend-interpretation", map[string]any{"revision": 1, "actor": "operator", "defect": evidence}, 200)
	recovery.Validation.ObservedAt = time.Now().UTC()
	resume(409)
	history, err = reg.InterpretationHistory(ctx, "s")
	must(err)
	recovery.SuspensionID = history[1].ID
	recovery.Correction.ObservedAt = time.Now().UTC()
	body = post("reprocess-interpretation", action, 200)
	must(json.Unmarshal(body, &selection))
	recovery.ReprocessingID = selection.ID
	adapter.relevant = true
	for range 4 {
		worked, err = worker.RunOne(ctx)
		must(err)
		if !worked {
			t.Fatal("relevant recovery job missing")
		}
	}

	recovery.Validation.ObservedAt = time.Now().UTC()
	resume(409)
	extractCatalog, err := extraction.RegisterCatalog(ctx, process, "fixture", "qwen3.8-27b", time.Now())
	must(err)
	extractRunner := &extraction.Runner{Documents: retained, OCR: ocr.NewStore(pool), Classifications: classification.NewStore(pool), Processing: process, Results: extraction.NewStore(pool), Adapter: lifecycleDownstream{stage: "extraction"}, Model: "qwen3.8-27b", ConfigurationVersion: extractCatalog.ConfigurationVersionID}
	worker.Handlers[extraction.Kind] = scheduler.Guard(scheduler.ExtractionHandler(extractRunner.Handler()))
	for range 3 {
		worked, err = worker.RunOne(ctx)
		must(err)
		if !worked {
			t.Fatal("extraction recovery job missing")
		}
	}
	// Extracted measures still require completed linking (including no-relation).
	recovery.Validation.ObservedAt = time.Now().UTC()
	resume(409)
	linkCatalog, err := linking.RegisterCatalog(ctx, process, "fixture", "qwen3.8-27b", time.Now())
	must(err)
	linkRunner := &linking.Runner{Store: linking.NewStore(pool), Processing: process, Adapter: lifecycleDownstream{stage: "linking"}, Model: "qwen3.8-27b", ConfigurationVersion: linkCatalog.ConfigurationVersionID}
	worker.Handlers[linking.Kind] = scheduler.Guard(linkRunner.Handler())
	for range 3 {
		worked, err = worker.RunOne(ctx)
		must(err)
		if !worked {
			t.Fatal("linking recovery job missing")
		}
	}
	recovery.Validation.ObservedAt = time.Now().UTC()
	resume(200)
}

func flattenGuided(values url.Values, prefix string, v any) {
	switch value := v.(type) {
	case map[string]any:
		for k, item := range value {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flattenGuided(values, key, item)
		}
	case []any:
		lines := []string{}
		for _, item := range value {
			lines = append(lines, fmt.Sprint(item))
		}
		values.Set(prefix, strings.Join(lines, "\n"))
	case nil:
		values.Set(prefix, "")
	default:
		values.Set(prefix, fmt.Sprint(value))
	}
}
