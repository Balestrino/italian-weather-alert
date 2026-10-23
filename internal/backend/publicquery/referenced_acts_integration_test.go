//go:build integration

package publicquery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/acts"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	alboOrigin     = "https://albo.example.it"
	alboDetail     = alboOrigin + "/web/trasparenza/papca-ap/-/papca/display/9000660?p_p_state=pop_up"
	alboDownload   = alboOrigin + "/web/trasparenza/papca-ap?p_p_id=jcitygovalbopubblicazioni_WAR_jcitygovalbiportlet&p_p_resource_id=downloadAllegato&_jcitygovalbopubblicazioni_WAR_jcitygovalbiportlet_downloadSigned=false&_jcitygovalbopubblicazioni_WAR_jcitygovalbiportlet_id=8000038"
	object38       = "EMERGENZA DI PROTEZIONE CIVILE - ORDINANZA CONTINGIBILE E URGENTE DI CHIUSURE TEMPORANEE DEI PARCHI, AREE DI SGAMBAMENTO, CIMITERI, DAL 10 SETTEMBRE 2026"
	ordinancePage1 = `COMUNE DI ESEMPIO ORDINANZA N. 38 / 2026 IL SINDACO Ritenuto che debbano essere assunte misure, disponendo: • la chiusura al pubblico: ◦ dell'area di sgambamento cani; ◦ dei cimiteri comunali;`
	ordinancePage2 = `ORDINA in via contingibile e urgente, a partire dalle ore 00.00 del giorno 10 settembre 2026, fino al perdurare dell'emergenza: • la chiusura al pubblico: ◦ dell'area di sgambamento cani; ◦ dei cimiteri comunali; • il divieto di qualunque attività all'aperto: ◦ nei giardini/parchi pubblici; ◦ nelle aree verdi; • la sospensione delle attività didattiche educative di tutte le scuole di ogni ordine e grado; AVVERTE che si riserva ulteriori misure.`
)

// syntheticAlbo serves one ordinance and a decree with the same object.
type syntheticAlbo struct{ details, downloads int }

func (a *syntheticAlbo) Window(_ context.Context, from, to time.Time, _ int) ([]acquisition.AlboRecord, error) {
	all := []acquisition.AlboRecord{
		{RecordID: "9000660", Category: "ORDINANZE", Subcategory: "ORDINANZA", ActNumber: 38, ActYear: 2026, RegistryYear: 2026, RegistryNumber: 660, Object: object38, PublishedFrom: time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)},
		{RecordID: "9000658", Category: "DECRETI", Subcategory: "DECRETO", ActNumber: 9, ActYear: 2026, RegistryYear: 2026, RegistryNumber: 658, Object: object38, PublishedFrom: time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)},
	}
	var result []acquisition.AlboRecord
	for _, r := range all {
		if !r.PublishedFrom.Before(from) && !r.PublishedFrom.After(to) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (a *syntheticAlbo) Detail(_ context.Context, id string) (acquisition.AlboDetail, []byte, error) {
	a.details++
	if id != "9000660" {
		return acquisition.AlboDetail{}, nil, fmt.Errorf("%w: unknown record", acquisition.ErrAlboUnavailable)
	}
	return acquisition.AlboDetail{AlboRecord: acquisition.AlboRecord{RecordID: id, Category: "ORDINANZE", Subcategory: "ORDINANZA", ActNumber: 38, ActYear: 2026, ActDate: time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC), RegistryYear: 2026, RegistryNumber: 660, Object: object38, DetailURL: alboDetail},
		Attachments: []acquisition.AlboAttachment{{Key: "7000040", MediaType: "application/x-pkcs7-mime", Title: "Ordinanza n. 00038/2026", Description: "Documento principale", UnsignedURL: alboDownload}}}, []byte(`<html><body><table data-id-pubblicazione="9000660"></table></body></html>`), nil
}

func (a *syntheticAlbo) Download(_ context.Context, raw string) ([]byte, string, error) {
	a.downloads++
	return []byte("%PDF-1.7 synthetic ordinance 38"), "application/pdf", nil
}

func testReferencedActs(ctx context.Context, t *testing.T, pool *pgxpool.Pool, createSource func(registry.Source, string, bool)) {
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{ocr.Migrate, interpretation.Migrate, acts.Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	const primary, alboSource = "acts-primary", "acts-albo"
	createSource(registry.Source{ID: primary, AuthorityID: "comune-calcinaia", ChannelID: "calcinaia", ProductID: "municipal", Territory: "050004"}, "https://calcinaia.example/acts", true)
	reg := registry.New(pool)
	now := time.Now().UTC()
	evidence := registry.Evidence{URL: "https://calcinaia.example/", Locator: "synthetic footer referral to the albo pretorio", ObservedAt: now}
	contract := &registry.JCityGovAlboContract{MunicipalityISTAT: "050004", ServesSource: primary, Origin: alboOrigin, ListPath: "/web/trasparenza/papca-ap/-/papca/igrid/1/2", DetailPathPrefix: "/web/trasparenza/papca-ap/-/papca/display/", Categories: map[string]string{"ordinanza": "ORDINANZE", "decreto": "DECRETI"}, PageBound: 12}
	albo := registry.Configuration{JCityGovAlbo: contract, URL: contract.ListURL(), Sections: []string{contract.ListURL()}, AccessMethod: registry.JCityGovAlboAccess, Attribution: "Comune di Calcinaia, albo pretorio",
		Referral: &registry.Referral{Evidence: evidence, Destination: contract.ListURL(), Context: "footer link of the municipal site", Sections: []string{contract.ListURL()}, ProductID: "municipal", Territory: "050004"},
		Policy:   registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true, PublicationPermitted: true}, Limitations: []string{"referenced acts only"}}
	for _, err := range []error{
		reg.CreateSource(ctx, registry.Source{ID: alboSource, AuthorityID: "comune-calcinaia", ChannelID: "calcinaia", ProductID: "municipal", Territory: "050004"}, albo, "test"),
		reg.RecordPreview(ctx, alboSource, 1, "test", evidence),
		reg.EnableCollection(ctx, alboSource, 1, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	objects := &memoryObjects{}
	retained := documents.New(pool, objects)
	issuer := "comune-calcinaia"
	retain := func(path, date, text string) documents.Version {
		t.Helper()
		url := "https://calcinaia.example/acts/" + path
		v, err := retained.Retain(ctx, documents.Acquisition{ID: "acts-" + path, SourceID: primary, Configuration: 1, URL: url, IssuerID: &issuer, Metadata: json.RawMessage(`{"source_publication_date":"` + date + `"}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: primary, Configuration: 1, MediaType: "text/html", Bytes: []byte("<main><p>" + text + "</p></main>")}}})
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	noticeA := retain("criticita", "2026-09-09", "Emessa ordinanza di chiusura cimiteri e parchi. Il Sindaco ha predisposto mediante apposita ordinanza, a partire dalle ore 0.00 di giovedì 10 agosto, la chiusura al pubblico dei cimiteri comunali e dell'area sgambamento cani, il divieto di attività all'aperto e la sospensione delle attività educative di tutte le scuole; chiuso anche il sottopasso di via Roma.")
	noticeB := retain("monitoraggio", "2026-09-10", "Restano in vigore le disposizioni adottate con ordinanza comunale: chiusura al pubblico dei cimiteri comunali.")

	process := processing.New(pool)
	classCatalog, err := classification.RegisterCatalog(ctx, process, "openai-chat", "synthetic-model", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	extractCatalog, err := extraction.RegisterCatalog(ctx, process, "openai-chat", "synthetic-model", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	type measure struct{ kind, subject, from string }
	interpret := func(v documents.Version, measures ...measure) int64 {
		t.Helper()
		source := primary
		classRun, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: fmt.Sprintf("acts-class-%d", v.ID), Workload: "ordinary", Stage: "classification", ConfigurationVersionID: classCatalog.ConfigurationVersionID, SourceID: &source, DocumentVersionID: &v.ID, Subject: json.RawMessage(`{}`), CreatedAt: time.Now()})
		if err != nil {
			t.Fatal(err)
		}
		relevant := true
		if err = classification.NewStore(pool).Put(ctx, classification.Result{RunID: classRun.ID, DocumentVersionID: v.ID, Status: "classified", Relevant: &relevant, ReasonCode: "local_weather_measure", EvidenceQuote: "ordinanza", ContentSHA256: strings.Repeat("a", 64), ContentComplete: true, ProviderResponseID: "synthetic", ReturnedModel: "synthetic-model", CreatedAt: time.Now()}); err != nil {
			t.Fatal(err)
		}
		run, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: fmt.Sprintf("acts-extract-%d", v.ID), Workload: "ordinary", Stage: "extraction", ConfigurationVersionID: extractCatalog.ConfigurationVersionID, SourceID: &source, DocumentVersionID: &v.ID, Subject: json.RawMessage(`{}`), CreatedAt: time.Now()})
		if err != nil {
			t.Fatal(err)
		}
		url := v.Resources[0].URL
		var extracted []extraction.Measure
		for i, m := range measures {
			e := extraction.Measure{Ordinal: i + 1, Kind: m.kind, Subject: m.subject, IndeterminateFields: []string{"place", "valid_until"}, Evidence: []extraction.Evidence{{Field: "kind", ResourceURL: url, Quote: "chiusura"}, {Field: "subject", ResourceURL: url, Quote: m.subject}}}
			if m.from != "" {
				from := m.from
				e.ValidFrom = &from
				e.Evidence = append(e.Evidence, extraction.Evidence{Field: "valid_from", ResourceURL: url, Quote: m.from})
			} else {
				e.IndeterminateFields = append(e.IndeterminateFields, "valid_from")
			}
			extracted = append(extracted, e)
		}
		if err = extraction.NewStore(pool).Put(ctx, extraction.Result{RunID: run.ID, DocumentVersionID: v.ID, ClassificationRunID: classRun.ID, Status: "extracted", ReasonCode: "measures_extracted", ContentSHA256: strings.Repeat("b", 64), ContentComplete: true, ProviderResponseID: "synthetic", ReturnedModel: "synthetic-model", CreatedAt: time.Now(), Measures: extracted}); err != nil {
			t.Fatal(err)
		}
		if n, err := domain.New(pool).ProjectMunicipalExtraction(ctx, run.ID, time.Now()); err != nil || n != len(measures) {
			t.Fatalf("projection %d of %d: %v", n, len(measures), err)
		}
		return run.ID
	}
	runA := interpret(noticeA,
		measure{"closure", "cimiteri comunali", "a partire dalle ore 0.00 di giovedì 10 agosto"},
		measure{"closure", "area sgambamento cani", ""},
		measure{"suspension", "attività educative di tutte le scuole", ""},
		measure{"closure", "sottopasso di via Roma", ""},
		measure{"prohibition", "attività all'aperto", "2026-09-11T00:00:00+02:00"})
	runB := interpret(noticeB, measure{"closure", "cimiteri comunali", ""})

	queue := jobs.New(pool)
	synthetic := &syntheticAlbo{}
	clock := time.Now().UTC()
	runner := &acts.Runner{Pool: pool, Store: acts.NewStore(pool), Documents: retained, OCR: ocr.NewStore(pool), Queue: queue, Policy: inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second},
		NewAlbo: func(c *registry.JCityGovAlboContract) acts.AlboClient {
			if c.ServesSource != primary {
				t.Fatalf("albo contract for %s", c.ServesSource)
			}
			return synthetic
		}, Now: func() time.Time { return clock }}
	handlers := runner.Handlers()
	payload := func(v documents.Version, run int64, resolution int64) jobs.Job {
		body, _ := json.Marshal(acts.Payload{DocumentVersionID: v.ID, ExtractionRunID: run, Workload: "ordinary", ResolutionID: resolution})
		return jobs.Job{ID: 1, Attempt: 1, Payload: body}
	}
	count := func(query string, args ...any) int {
		t.Helper()
		var n int
		if err := pool.QueryRow(ctx, query, args...).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	t.Run("projection schedules resolution", func(t *testing.T) {
		if err := runner.AfterProjection(ctx, runA, 5, "ordinary", clock); err != nil {
			t.Fatal(err)
		}
		if err := runner.AfterProjection(ctx, runA, 0, "ordinary", clock); err != nil {
			t.Fatal(err)
		}
		if n := count(`SELECT count(*) FROM processing_jobs WHERE kind=$1 AND (payload->>'document_version_id')::bigint=$2`, acts.KindResolve, noticeA.ID); n != 1 {
			t.Fatalf("resolution jobs %d", n)
		}
	})
	t.Run("both notices resolve to the ordinance, not the decree", func(t *testing.T) {
		for _, job := range []jobs.Job{payload(noticeA, runA, 0), payload(noticeB, runB, 0)} {
			if _, err := handlers[acts.KindResolve](ctx, job); err != nil {
				t.Fatal(err)
			}
		}
		if n := count(`SELECT count(*) FROM act_resolutions WHERE outcome='unique_candidate' AND albo_record_id='9000660' AND albo_source_id=$1`, alboSource); n != 2 {
			t.Fatalf("unique resolutions %d", n)
		}
		if n := count(`SELECT count(*) FROM processing_jobs WHERE kind=$1`, acts.KindAcquire); n != 2 {
			t.Fatalf("acquisition jobs %d", n)
		}
		if _, err := handlers[acts.KindResolve](ctx, payload(noticeA, runA, 0)); err != nil {
			t.Fatal(err)
		}
		if n := count(`SELECT count(*) FROM act_resolutions`); n != 2 {
			t.Fatalf("final resolution repeated: %d", n)
		}
	})
	var actVersion int64
	resolutionOf := func(v documents.Version) int64 {
		var id int64
		if err := pool.QueryRow(ctx, `SELECT r.id FROM act_resolutions r JOIN act_references f ON f.id=r.reference_id WHERE f.document_version_id=$1`, v.ID).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	t.Run("the shared act is retained once and only OCR runs on it", func(t *testing.T) {
		for _, job := range []jobs.Job{payload(noticeA, runA, resolutionOf(noticeA)), payload(noticeB, runB, resolutionOf(noticeB))} {
			if _, err := handlers[acts.KindAcquire](ctx, job); err != nil {
				t.Fatal(err)
			}
		}
		if err := pool.QueryRow(ctx, `SELECT document_version_id FROM act_albo_records WHERE source_id=$1 AND record_id='9000660'`, alboSource).Scan(&actVersion); err != nil {
			t.Fatal(err)
		}
		if synthetic.downloads != 1 || count(`SELECT count(*) FROM act_albo_records`) != 1 {
			t.Fatalf("act retained %d times", synthetic.downloads)
		}
		if n := count(`SELECT count(*) FROM processing_jobs WHERE kind=$1 AND (payload->>'document_version_id')::bigint=$2`, ocr.Kind, actVersion); n != 1 {
			t.Fatalf("OCR jobs for the act %d", n)
		}
		scheduler := interpretation.New(pool, queue, inference.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Second}, false)
		if scheduled, err := scheduler.AfterOCR(ctx, actVersion, "ordinary", clock); err != nil || scheduled {
			t.Fatalf("classification scheduled for an albo act: %v %v", scheduled, err)
		}
		if n := count(`SELECT count(*) FROM processing_jobs WHERE kind=$1 AND (payload->>'document_version_id')::bigint=$2`, classification.Kind, actVersion); n != 0 {
			t.Fatalf("classification jobs for the act %d", n)
		}
		if err := acquisition.NewScheduleStore(pool).SyncEnabled(ctx, clock); err != nil {
			t.Fatal(err)
		}
		if n := count(`SELECT count(*) FROM acquisition_source_status WHERE source_id=$1`, alboSource); n != 0 {
			t.Fatal("albo source scheduled for listing checks")
		}
		coverage, err := New(pool).Coverage(ctx, CoverageQuery{MunicipalityISTAT: "050004"})
		if err != nil {
			t.Fatal(err)
		}
		for _, source := range coverage.Sources {
			if source.SourceID == alboSource {
				t.Fatal("albo source counted in coverage")
			}
		}
		_, err = handlers[acts.KindClarify](ctx, payload(noticeA, runA, resolutionOf(noticeA)))
		var deferred *jobs.DeferredError
		if !errors.As(err, &deferred) || deferred.Code != "act_content_pending" {
			t.Fatalf("clarification before OCR: %v", err)
		}
	})

	// Simulate the completed OCR of the retained act's two pages.
	ocrCatalog, err := ocr.RegisterCatalog(ctx, process, "openai-chat", "synthetic-ocr", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	ocrRun, err := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "acts-ocr", Workload: "ordinary", Stage: "ocr", ConfigurationVersionID: ocrCatalog.ConfigurationVersionID, DocumentVersionID: &actVersion, Subject: json.RawMessage(`{}`), CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	ocrResults := ocr.NewStore(pool)
	for i, text := range []string{ordinancePage1, ordinancePage2} {
		sum := sha256.Sum256([]byte(text))
		if err = ocrResults.PutPage(ctx, ocr.PageResult{RunID: ocrRun.ID, DocumentVersionID: actVersion, PageNumber: i + 1, ResourceURL: alboDownload, Status: "complete", MediaType: "image/png", InputSHA256: strings.Repeat(fmt.Sprint(i+1), 64), OutputSHA256: hex.EncodeToString(sum[:]), ExtractedText: text, ProviderResponseID: "synthetic", ReturnedModel: "synthetic-ocr", CreatedAt: time.Now()}); err != nil {
			t.Fatal(err)
		}
	}
	if err = ocrResults.PutResource(ctx, ocr.ResourceResult{RunID: ocrRun.ID, DocumentVersionID: actVersion, ResourceURL: alboDownload, Status: "complete", PageCount: 2, CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	clock = time.Now().UTC().Add(time.Second).Truncate(time.Microsecond)
	beforeClarification := clock.Add(-time.Microsecond)
	measureID := func(run int64, ordinal int) string { return fmt.Sprintf("municipal-extraction-v1:%d:%d", run, ordinal) }
	validity := func(id string) (string, *time.Time, string) {
		t.Helper()
		var precision string
		var instant *time.Time
		var condition *string
		if err := pool.QueryRow(ctx, `SELECT precision,instant,condition FROM domain_temporal_values WHERE entity_kind='local_measure' AND entity_id=$1 AND meaning='validity' ORDER BY created_at DESC,id DESC LIMIT 1`, id).Scan(&precision, &instant, &condition); err != nil {
			t.Fatal(err)
		}
		if condition == nil {
			return precision, instant, ""
		}
		return precision, instant, *condition
	}

	t.Run("clarification of unknown validity from the ordinance", func(t *testing.T) {
		for _, job := range []jobs.Job{payload(noticeA, runA, resolutionOf(noticeA)), payload(noticeB, runB, resolutionOf(noticeB))} {
			if _, err := handlers[acts.KindClarify](ctx, job); err != nil {
				t.Fatal(err)
			}
		}
		start := time.Date(2026, 9, 9, 22, 0, 0, 0, time.UTC)
		for _, id := range []string{measureID(runA, 1), measureID(runA, 2), measureID(runA, 3), measureID(runB, 1)} {
			precision, instant, condition := validity(id)
			if precision != "conditional" || instant == nil || !instant.Equal(start) || condition != "fino al perdurare dell'emergenza" {
				t.Fatalf("%s not clarified: %s %v %q", id, precision, instant, condition)
			}
		}
		if precision, _, _ := validity(measureID(runA, 4)); precision != "unknown" {
			t.Fatalf("unmatched measure changed: %s", precision)
		}
		if precision, instant, _ := validity(measureID(runA, 5)); precision != "instant" || instant.Equal(start) {
			t.Fatalf("determinate measure changed: %s %v", precision, instant)
		}
		outcomes := map[string]string{}
		rows, err := pool.Query(ctx, `SELECT local_measure_id,outcome FROM act_clarifications`)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var id, outcome string
			if err = rows.Scan(&id, &outcome); err != nil {
				t.Fatal(err)
			}
			outcomes[id] = outcome
		}
		rows.Close()
		if outcomes[measureID(runA, 1)] != "clarified" || outcomes[measureID(runA, 4)] != "not_matched" || outcomes[measureID(runA, 5)] != "conflict" || outcomes[measureID(runB, 1)] != "clarified" || len(outcomes) != 6 {
			t.Fatalf("clarification outcomes %v", outcomes)
		}
		var actPassage, noticePassage string
		if err = pool.QueryRow(ctx, `SELECT act_passage,notice_passage FROM act_clarifications WHERE local_measure_id=$1`, measureID(runA, 5)).Scan(&actPassage, &noticePassage); err != nil || !strings.Contains(actPassage, "10 settembre 2026") || noticePassage != "2026-09-11T00:00:00+02:00" {
			t.Fatalf("conflict passages %q %q %v", actPassage, noticePassage, err)
		}
		var state, reason, actor string
		var evidence int64
		if err = pool.QueryRow(ctx, `SELECT state,reason,actor,evidence_document_version_id FROM domain_interpretation_events WHERE local_measure_id=$1 ORDER BY id DESC LIMIT 1`, measureID(runA, 1)).Scan(&state, &reason, &actor, &evidence); err != nil || reason != "clarified_by_referenced_act" || evidence != actVersion || actor != acts.ClarificationActor {
			t.Fatalf("interpretation event %s %s %d %v", state, reason, evidence, err)
		}
		temporal := count(`SELECT count(*) FROM domain_temporal_values`)
		clock = clock.Add(time.Minute)
		if _, err = handlers[acts.KindClarify](ctx, payload(noticeA, runA, resolutionOf(noticeA))); err != nil {
			t.Fatal(err)
		}
		if count(`SELECT count(*) FROM domain_temporal_values`) != temporal || count(`SELECT count(*) FROM act_clarifications`) != 6 {
			t.Fatal("repeated clarification appended rows")
		}
	})
	t.Run("attached act without a recognizable clause", func(t *testing.T) {
		url := "https://calcinaia.example/acts/allerta"
		attachment := url + "/ordinanza.html"
		noticeC, err := retained.Retain(ctx, documents.Acquisition{ID: "acts-allerta", SourceID: primary, Configuration: 1, URL: url, IssuerID: &issuer, Metadata: json.RawMessage(`{"source_publication_date":"2026-08-20"}`), Resources: []documents.Resource{
			{URL: url, Role: "original", Required: true, SourceID: primary, Configuration: 1, MediaType: "text/html", Bytes: []byte("<main><p>È stata emessa apposita ordinanza per la chiusura dei cimiteri comunali.</p></main>")},
			{URL: attachment, Role: "attachment", Required: true, SourceID: primary, Configuration: 1, MediaType: "text/html", Bytes: []byte("<main><p>ORDINANZA N. 40 / 2026 Il Sindaco dispone quanto in premessa.</p></main>")},
		}})
		if err != nil {
			t.Fatal(err)
		}
		runC := interpret(noticeC, measure{"closure", "cimiteri comunali", ""})
		if _, err = handlers[acts.KindResolve](ctx, payload(noticeC, runC, 0)); err != nil {
			t.Fatal(err)
		}
		var outcome, resource string
		if err = pool.QueryRow(ctx, `SELECT outcome,attached_resource_url FROM act_resolutions WHERE id=$1`, resolutionOf(noticeC)).Scan(&outcome, &resource); err != nil || outcome != "attached" || resource != attachment {
			t.Fatalf("attached resolution %s %s %v", outcome, resource, err)
		}
		if n := count(`SELECT count(*) FROM processing_jobs WHERE kind=$1 AND (payload->>'resolution_id')::bigint=$2`, acts.KindClarify, resolutionOf(noticeC)); n != 1 {
			t.Fatalf("clarification jobs for the attached act %d", n)
		}
		if _, err = handlers[acts.KindClarify](ctx, payload(noticeC, runC, resolutionOf(noticeC))); err != nil {
			t.Fatal(err)
		}
		var reason string
		if err = pool.QueryRow(ctx, `SELECT reason FROM act_clarifications WHERE resolution_id=$1 AND outcome='clarification_unavailable' AND local_measure_id IS NULL`, resolutionOf(noticeC)).Scan(&reason); err != nil || reason != "clause_not_recognized" {
			t.Fatalf("unrecognized clause not recorded: %q %v", reason, err)
		}
		if precision, _, _ := validity(measureID(runC, 1)); precision != "unknown" {
			t.Fatalf("measure changed by an unparseable act: %s", precision)
		}
	})
	_ = beforeClarification
}
