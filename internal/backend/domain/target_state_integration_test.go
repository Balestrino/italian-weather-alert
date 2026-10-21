//go:build integration

package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Tracked target states follow projected measures, validity, applied linked
// updates and classification, evaluated at each refresh time.
func TestTrackedTargetStatesFollowValidatedInterpretation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := domainTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, acquisition.Migrate, Migrate, MigrateTemporal} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://www.calcinaia.example/note-legali", Locator: "fixture", ObservedAt: now}
	sourceID := "state-source"
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "state-authority", Name: "Comune di Calcinaia", OfficialURL: "https://www.calcinaia.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "state-channel", PublisherID: "state-authority", Platform: "municipal-site", URL: "https://www.calcinaia.example"}),
		reg.CreateSource(ctx, registry.Source{ID: sourceID, AuthorityID: "state-authority", ChannelID: "state-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://www.calcinaia.example", Sections: []string{"https://www.calcinaia.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	tracker := acquisition.NewTrackingStore(pool)
	documentStore := documents.New(pool, &domainObjects{})
	process := processing.New(pool)
	classCatalog, err := classification.RegisterCatalog(ctx, process, "fixture-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	extractCatalog, err := extraction.RegisterCatalog(ctx, process, "fixture-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	linkCatalog, err := linking.RegisterCatalog(ctx, process, "fixture-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	store := New(pool)
	measureRuns := map[string]int64{}
	url := func(name string) string { return "https://www.calcinaia.example/notices/" + name }
	// Each notice is discovered with its publication date before retention.
	for name, age := range map[string]int{"none": 5, "interval": 100, "open": 40, "aged": 91, "custom": 40, "conflict": 20, "partial": 50, "reopening": 45, "cancelled": 70, "cancelling": 60, "extended": 30, "extending": 10, "revised": 10, "manual": 200} {
		published := now.AddDate(0, 0, -age)
		if _, err := tracker.Remember(ctx, sourceID, 1, []acquisition.DiscoveredDocument{{URL: url(name), PublicationDate: &published}}, now.AddDate(0, 0, -age)); err != nil {
			t.Fatal(err)
		}
	}
	sequence := 0
	retain := func(name, body string) documents.Version {
		t.Helper()
		sequence++
		version, err := documentStore.Retain(ctx, documents.Acquisition{ID: fmt.Sprintf("state-%s-%d", name, sequence), SourceID: sourceID, Configuration: 1, URL: url(name), Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url(name), Role: "original", Required: true, SourceID: sourceID, Configuration: 1, MediaType: "text/html", Bytes: []byte(body)}}})
		if err != nil {
			t.Fatal(err)
		}
		// Keep acquisition chronology observable even on very fast hosts.
		time.Sleep(time.Millisecond)
		return version
	}
	classify := func(version documents.Version, relevant bool) int64 {
		t.Helper()
		sequence++
		run, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: fmt.Sprintf("state-class-%d", sequence), Workload: "evaluation", Stage: "classification", ConfigurationVersionID: classCatalog.ConfigurationVersionID, SourceID: &sourceID, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		reason := "local_weather_measure"
		if !relevant {
			reason = "not_relevant"
		}
		if err = classification.NewStore(pool).Put(ctx, classification.Result{RunID: run.ID, DocumentVersionID: version.ID, Status: "classified", Relevant: &relevant, ReasonCode: reason, EvidenceQuote: "fixture", ContentSHA256: strings.Repeat("c", 64), ContentComplete: true, ProviderResponseID: "fixture", ReturnedModel: "qwen3.8-27b", CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
		return run.ID
	}
	// measure projects one bound municipal measure with the given validity.
	measure := func(id string, version documents.Version, validity TemporalValue) {
		t.Helper()
		classRun := classify(version, true)
		sequence++
		run, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: fmt.Sprintf("state-extract-%d", sequence), Workload: "evaluation", Stage: "extraction", ConfigurationVersionID: extractCatalog.ConfigurationVersionID, SourceID: &sourceID, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		if err = extraction.NewStore(pool).Put(ctx, extraction.Result{RunID: run.ID, DocumentVersionID: version.ID, ClassificationRunID: classRun, Status: "extracted", ReasonCode: "measures_extracted", ContentSHA256: strings.Repeat("d", 64), ContentComplete: true, ProviderResponseID: "fixture", ReturnedModel: "qwen3.8-27b",
			Measures: []extraction.Measure{{Ordinal: 1, Kind: "closure", Subject: id, IndeterminateFields: []string{"place", "valid_from", "valid_until"}, Evidence: []extraction.Evidence{{Field: "subject", ResourceURL: "fixture", Quote: id}}}}, CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
		if err = store.PutLocalMeasure(ctx, LocalMeasure{ID: id, DocumentVersionID: version.ID, SourceID: sourceID, MunicipalityISTAT: "050004", Kind: "closure", Subject: id}); err != nil {
			t.Fatal(err)
		}
		if err = store.BindExtraction(ctx, id, run.ID, 1); err != nil {
			t.Fatal(err)
		}
		validity.EntityKind, validity.EntityID, validity.Meaning, validity.EvidenceDocumentVersionID, validity.CreatedAt = "local_measure", id, "validity", version.ID, now
		if _, err = store.PutTemporal(ctx, validity); err != nil {
			t.Fatal(err)
		}
		measureRuns[id] = run.ID
	}
	link := func(current, candidate, relation string, version documents.Version) {
		t.Helper()
		sequence++
		run, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: fmt.Sprintf("state-link-%d", sequence), Workload: "evaluation", Stage: "linking", ConfigurationVersionID: linkCatalog.ConfigurationVersionID, SourceID: &sourceID, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{"fixture":true}`), CreatedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		candidateRun, ordinal := measureRuns[candidate], 1
		if err = linking.NewStore(pool).Put(ctx, linking.Result{RunID: run.ID, CurrentRunID: measureRuns[current], CurrentOrdinal: 1, Status: "linked", ReasonCode: "explicit_reference", Relation: &relation, CandidateRunID: &candidateRun, CandidateOrdinal: &ordinal, ProviderResponseID: "fixture", ReturnedModel: "qwen3.8-27b", CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
		if _, err = store.ApplyLinkedUpdate(ctx, run.ID, now); err != nil {
			t.Fatal(err)
		}
	}
	unknown := TemporalValue{Precision: "unknown"}
	end := now.Add(24 * time.Hour)
	start := now.AddDate(0, 0, -100)
	conflict := "fixture-conflict"

	classify(retain("none", "avviso senza misure"), false)
	measure("interval-closure", retain("interval", "chiuso fino a domani"), TemporalValue{Precision: "interval", Original: "fino a domani", Instant: &start, EndInstant: &end})
	measure("open-coc", retain("open", "COC aperto"), TemporalValue{Precision: "instant", Original: "dalle 17", Instant: &start})
	measure("aged-coc", retain("aged", "COC aperto da tempo"), unknown)
	measure("custom-coc", retain("custom", "COC con limite ridotto"), unknown)
	measure("conflict-closure", retain("conflict", "date discordanti"), TemporalValue{Precision: "unknown", ConflictGroup: &conflict})
	measure("partial-closure", retain("partial", "chiusi sottopasso e parchi"), TemporalValue{Precision: "conditional", Original: "fino a nuova comunicazione", Condition: stringPointer("fino a nuova comunicazione")})
	reopening := retain("reopening", "riaperto il sottopasso")
	measure("reopening-measure", reopening, unknown)
	link("reopening-measure", "partial-closure", "reopens", reopening)
	measure("cancelled-closure", retain("cancelled", "chiusura"), unknown)
	cancelling := retain("cancelling", "revocata la chiusura")
	measure("cancelling-measure", cancelling, unknown)
	link("cancelling-measure", "cancelled-closure", "cancels", cancelling)
	measure("extended-closure", retain("extended", "chiusura fino a venerdì"), unknown)
	extending := retain("extending", "prorogata la chiusura")
	measure("extending-measure", extending, unknown)
	link("extending-measure", "extended-closure", "extends", extending)
	measure("revised-closure", retain("revised", "prima versione con chiusura"), unknown)
	classify(retain("revised", "seconda versione senza chiusura"), false)
	if _, err := pool.Exec(ctx, "UPDATE acquisition_targets SET explicit_reference=true WHERE url=$1", url("manual")); err != nil {
		t.Fatal(err)
	}

	states := func() map[string][3]string {
		t.Helper()
		rows, err := pool.Query(ctx, "SELECT url,measure_state,relevance_state,explicit_reference::text FROM acquisition_targets WHERE source_id=$1", sourceID)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		result := map[string][3]string{}
		for rows.Next() {
			var target, measureState, relevance, reference string
			if err = rows.Scan(&target, &measureState, &relevance, &reference); err != nil {
				t.Fatal(err)
			}
			result[strings.TrimPrefix(target, url(""))] = [3]string{measureState, relevance, reference}
		}
		return result
	}
	expect := func(got map[string][3]string, want map[string][3]string) {
		t.Helper()
		for name, state := range want {
			if got[name] != state {
				t.Errorf("%s: got %v, want %v", name, got[name], state)
			}
		}
	}
	if err := tracker.RefreshInterpretationState(ctx, sourceID, now, 90); err != nil {
		t.Fatal(err)
	}
	expect(states(), map[string][3]string{
		"none":       {"none", "irrelevant", "false"},
		"interval":   {"ongoing", "relevant", "false"},
		"open":       {"unresolved", "relevant", "false"},
		"aged":       {"aged_unresolved", "relevant", "false"},
		"conflict":   {"unresolved", "relevant", "false"},
		"partial":    {"unresolved", "relevant", "true"},
		"cancelled":  {"ceased", "relevant", "true"},
		"cancelling": {"unresolved", "relevant", "false"},
		"extended":   {"unresolved", "relevant", "true"},
		"revised":    {"unresolved", "irrelevant", "false"},
		"manual":     {"none", "unknown", "true"},
	})
	var derivedAt time.Time
	if err := pool.QueryRow(ctx, "SELECT interpretation_state_at FROM acquisition_targets WHERE url=$1", url("open")).Scan(&derivedAt); err != nil || !derivedAt.Equal(now) {
		t.Fatalf("derivation time: %v %v", derivedAt, err)
	}

	// The recurring plan follows the maintained state beyond the 30-day window.
	if err := tracker.CompleteBootstrap(ctx, sourceID, 1, now); err != nil {
		t.Fatal(err)
	}
	planned := func() map[string]bool {
		t.Helper()
		plan, err := tracker.Plan(ctx, sourceID, 1, now, 30)
		if err != nil {
			t.Fatal(err)
		}
		result := map[string]bool{}
		for _, document := range plan.Documents {
			result[strings.TrimPrefix(document.URL, url(""))] = true
		}
		return result
	}
	inPlan := planned()
	for name, want := range map[string]bool{"interval": true, "open": true, "partial": true, "aged": false, "cancelled": false, "manual": false} {
		if inPlan[name] != want {
			t.Errorf("%s planned=%v, want %v", name, inPlan[name], want)
		}
	}

	// Withdrawn protection follows the maintained state: the aged notice can be
	// withdrawn, recent open-ended ones stay blocking, and an operator exclusion
	// of a notice that ever produced a measure still needs the acknowledgement.
	for _, name := range []string{"aged", "open", "partial"} {
		if err := tracker.RecordUnavailable(ctx, sourceID, 1, url(name), 404, now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := pool.Exec(ctx, `INSERT INTO acquisition_checks(source_id,configuration,worker_id,started_at,finished_at,reachable,content_recognized,complete,error_code,listing_count,document_count,check_state)
 VALUES($1,1,'test',$2,$2,true,false,false,'source_http_error',1,0,'collection_failure')`, sourceID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "UPDATE registry_sources SET active_revision=1 WHERE id=$1", sourceID); err != nil {
		t.Fatal(err)
	}
	dispositions := acquisition.NewDispositionStore(pool)
	unavailable, err := dispositions.ListUnavailable(ctx, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	classes := map[string]string{}
	for _, target := range unavailable {
		classes[strings.TrimPrefix(target.URL, url(""))] = target.Classification + "/" + target.MeasureState
	}
	for name, want := range map[string]string{"aged": "withdrawn/aged_unresolved", "open": "blocking/unresolved", "partial": "blocking/unresolved"} {
		if classes[name] != want {
			t.Errorf("%s classified %q, want %q", name, classes[name], want)
		}
	}
	var agedSeen time.Time
	if err := pool.QueryRow(ctx, "SELECT last_seen_at FROM acquisition_targets WHERE url=$1", url("aged")).Scan(&agedSeen); err != nil {
		t.Fatal(err)
	}
	exclusion := acquisition.DispositionRequest{SourceID: sourceID, Configuration: 1, URL: url("aged"), Decision: acquisition.DecisionExclude, ObservedLastSeenAt: agedSeen, Actor: "operator", Reason: "aged notice withdrawn", Evidence: json.RawMessage(`{"pages": 1}`)}
	if _, err := dispositions.RecordDisposition(ctx, exclusion); err != acquisition.ErrMeasureAcknowledgementRequired {
		t.Fatalf("aged measure exclusion without acknowledgement: %v", err)
	}

	followed, err := tracker.FollowedTargets(ctx, sourceID, now)
	if err != nil {
		t.Fatal(err)
	}
	reasons := map[string]string{}
	for _, target := range followed {
		if target.LimitDays != 90 || target.InterpretationStateAt == nil {
			t.Fatalf("followed target details: %+v", target)
		}
		reasons[strings.TrimPrefix(target.URL, url(""))] = target.Recheck
	}
	if _, listed := reasons["none"]; listed {
		t.Error("notice without measure or reference listed as followed")
	}
	for name, want := range map[string]string{"interval": "ongoing", "open": "unresolved", "aged": "not_rechecked", "cancelled": "not_rechecked", "manual": "not_rechecked", "revised": "window"} {
		if reasons[name] != want {
			t.Errorf("%s recheck %q, want %q", name, reasons[name], want)
		}
	}

	// During a bootstrap, an explicitly referenced old notice is rechecked.
	if _, err := pool.Exec(ctx, "UPDATE acquisition_bootstrap SET completed_at=NULL WHERE source_id=$1", sourceID); err != nil {
		t.Fatal(err)
	}
	if followed, err = tracker.FollowedTargets(ctx, sourceID, now); err != nil {
		t.Fatal(err)
	}
	for _, target := range followed {
		if target.URL == url("manual") && target.Recheck != "bootstrap_reference" {
			t.Errorf("referenced notice during bootstrap: %+v", target)
		}
	}
	if err := tracker.CompleteBootstrap(ctx, sourceID, 1, now); err != nil {
		t.Fatal(err)
	}

	// A shorter source limit ages the 40-day open-ended notice.
	if err := tracker.RefreshInterpretationState(ctx, sourceID, now, 30); err != nil {
		t.Fatal(err)
	}
	expect(states(), map[string][3]string{"custom": {"aged_unresolved", "relevant", "false"}, "open": {"aged_unresolved", "relevant", "false"}})
	if planned()["open"] {
		t.Error("aged open-ended notice still planned")
	}

	// Two days later the supported end has passed and the references remain.
	later := now.Add(48 * time.Hour)
	if err := tracker.RefreshInterpretationState(ctx, sourceID, later, 90); err != nil {
		t.Fatal(err)
	}
	expect(states(), map[string][3]string{
		"interval": {"ceased", "relevant", "false"},
		"open":     {"unresolved", "relevant", "false"},
		"extended": {"unresolved", "relevant", "true"},
		"manual":   {"none", "unknown", "true"},
	})
}
