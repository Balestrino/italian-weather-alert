//go:build integration

package publicquery

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testEquivalentVersionMeasures runs inside the shared public-query fixture,
// which provides the territorial and dataset setup that projection requires.
func testEquivalentVersionMeasures(ctx context.Context, t *testing.T, pool *pgxpool.Pool, createSource func(registry.Source, string, bool)) {
	if err := ocr.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	const sourceID = "equivalent-measures"
	createSource(registry.Source{ID: sourceID, AuthorityID: "comune-calcinaia", ChannelID: "calcinaia", ProductID: "municipal", Territory: "050004"}, "https://calcinaia.example/equivalent", true)
	retained := documents.New(pool, &memoryObjects{})
	issuer := "comune-calcinaia"
	const notice = "Chiusura del ponte sintetico in Via Sintetica fino a revoca"
	retain := func(path, class, text string) documents.Version {
		t.Helper()
		url := "https://calcinaia.example/equivalent/" + path
		body := fmt.Sprintf(`<main class="%s"><h1>Avviso</h1><p>%s</p></main>`, class, text)
		v, err := retained.Retain(ctx, documents.Acquisition{ID: path + "-" + class, SourceID: sourceID, Configuration: 1, URL: url, IssuerID: &issuer, Metadata: json.RawMessage(`{"source_publication_date":"2026-10-07"}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: sourceID, Configuration: 1, MediaType: "text/html", Bytes: []byte(body)}}})
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	// x: template-only versions. w: the later version's fingerprint arrives late.
	// y: a genuine revision. z: later equivalents that cannot replace anything.
	xa, xb := retain("x", "v1", notice), retain("x", "template-2", notice)
	wa, wb := retain("w", "v1", notice), retain("w", "template-2", notice)
	ya, yb := retain("y", "v1", "Chiusura del ponte sintetico e del sottopasso sintetico in Via Sintetica fino a revoca"), retain("y", "v2", "Chiusura del sottopasso sintetico in Via Sintetica fino a revoca")
	za, zb, zc, zd := retain("z", "v1", notice), retain("z", "template-2", notice), retain("z", "template-3", notice), retain("z", "template-4", notice)

	process := processing.New(pool)
	classCatalog, err := classification.RegisterCatalog(ctx, process, "openai-chat", "synthetic-model", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	extractCatalog, err := extraction.RegisterCatalog(ctx, process, "openai-chat", "synthetic-model", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	manifests, ocrStore, domainStore, store := classification.NewStore(pool), ocr.NewStore(pool), domain.New(pool), New(pool)
	base := time.Now().UTC().Add(time.Second).Truncate(time.Microsecond)
	at := func(seconds int) time.Time { return base.Add(time.Duration(seconds) * time.Second) }
	startRun := func(key, stage, configuration string, v documents.Version, created time.Time, manifest bool) int64 {
		t.Helper()
		source := sourceID
		run, runErr := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: "equivalent-" + key, Workload: "ordinary", Stage: stage, ConfigurationVersionID: configuration, SourceID: &source, DocumentVersionID: &v.ID, Subject: json.RawMessage(`{"fixture":"equivalent-measures"}`), CreatedAt: created})
		if runErr != nil {
			t.Fatal(runErr)
		}
		if !manifest {
			return run.ID
		}
		content, gatherErr := classification.GatherContent(ctx, retained, ocrStore, v)
		if gatherErr != nil {
			t.Fatal(gatherErr)
		}
		m, buildErr := classification.BuildManifest(ctx, retained, ocrStore, v, content, configuration)
		if buildErr != nil {
			t.Fatal(buildErr)
		}
		if putErr := manifests.PutManifest(ctx, run.ID, m); putErr != nil {
			t.Fatal(putErr)
		}
		return run.ID
	}
	// interpret records a classification at seconds and an extraction one
	// second later, both with manifests unless the fixture withholds them.
	interpret := func(key string, v documents.Version, seconds int, status string, manifest bool, subjects ...string) int64 {
		t.Helper()
		created := at(seconds)
		relevant, classReason, extractReason := status != "not_applicable", "local_weather_measure", "measures_extracted"
		switch status {
		case "not_applicable":
			classReason, extractReason = "not_relevant", "not_relevant"
		case "uninterpreted":
			extractReason = "extraction_evidence_invalid"
		}
		classRun := startRun(key+"-class", "classification", classCatalog.ConfigurationVersionID, v, created, manifest)
		if putErr := manifests.Put(ctx, classification.Result{RunID: classRun, DocumentVersionID: v.ID, Status: "classified", Relevant: &relevant, ReasonCode: classReason, EvidenceQuote: "Chiusura", ContentSHA256: strings.Repeat("a", 64), ContentComplete: true, ProviderResponseID: "synthetic", ReturnedModel: "synthetic-model", CreatedAt: created}); putErr != nil {
			t.Fatal(putErr)
		}
		url := v.Resources[0].URL
		var measures []extraction.Measure
		for i, subject := range subjects {
			measures = append(measures, extraction.Measure{Ordinal: i + 1, Kind: "closure", Subject: subject, Place: stringPointer("Via Sintetica"), ValidUntil: stringPointer("fino a revoca"), IndeterminateFields: []string{"valid_from"}, Evidence: []extraction.Evidence{{Field: "kind", ResourceURL: url, Quote: "Chiusura"}, {Field: "subject", ResourceURL: url, Quote: subject}, {Field: "place", ResourceURL: url, Quote: "Via Sintetica"}, {Field: "valid_until", ResourceURL: url, Quote: "fino a revoca"}}})
		}
		result := extraction.Result{DocumentVersionID: v.ID, ClassificationRunID: classRun, Status: status, ReasonCode: extractReason, ContentSHA256: strings.Repeat("b", 64), ContentComplete: true, CreatedAt: created.Add(time.Second), Measures: measures}
		if status == "extracted" {
			result.ProviderResponseID, result.ReturnedModel = "synthetic", "synthetic-model"
		}
		result.RunID = startRun(key+"-extract", "extraction", extractCatalog.ConfigurationVersionID, v, created.Add(time.Second), manifest)
		if putErr := extraction.NewStore(pool).Put(ctx, result); putErr != nil {
			t.Fatal(putErr)
		}
		return result.RunID
	}
	project := func(run int64, seconds, want int) {
		t.Helper()
		if n, projectErr := domainStore.ProjectMunicipalExtraction(ctx, run, at(seconds)); projectErr != nil || n != want {
			t.Fatalf("projection of run %d: %d %v", run, n, projectErr)
		}
	}
	project(interpret("xa", xa, 10, "extracted", true, "ponte sintetico"), 15, 1)
	project(interpret("xb", xb, 30, "extracted", true, "ponte"), 40, 1)
	project(interpret("xb-later", xb, 50, "extracted", true, "ponte di Via Sintetica"), 55, 1)
	project(interpret("wa", wa, 10, "extracted", true, "ponte sintetico"), 15, 1)
	project(interpret("wb", wb, 30, "extracted", false, "ponte"), 40, 1)
	startRun("wb-fingerprint", "classification", classCatalog.ConfigurationVersionID, wb, at(60), true)
	project(interpret("ya", ya, 10, "extracted", true, "ponte sintetico", "sottopasso sintetico"), 15, 2)
	project(interpret("yb", yb, 30, "extracted", true, "sottopasso"), 40, 1)
	project(interpret("za", za, 10, "extracted", true, "ponte sintetico"), 15, 1)
	project(interpret("zb", zb, 30, "not_applicable", true), 40, 0)
	interpret("zc", zc, 30, "extracted", true, "ponte")
	project(interpret("zd", zd, 30, "uninterpreted", true), 40, 0)

	var own []string
	rows, err := pool.Query(ctx, `SELECT id FROM domain_local_measures WHERE source_id=$1`, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		own = append(own, id)
	}
	if rows.Close(); rows.Err() != nil {
		t.Fatal(rows.Err())
	}
	// current returns subject@version per document, and checks that the
	// municipality situation selects the same measures as search.
	current := func(t *testing.T, seconds int) map[string][]string {
		t.Helper()
		qt := QueryTime{EvaluationTime: at(seconds), KnownAt: at(seconds)}
		found, searchErr := store.Search(ctx, SearchQuery{Kind: "measure", SourceID: sourceID, MunicipalityISTAT: "050004", QueryTime: qt})
		if searchErr != nil {
			t.Fatal(searchErr)
		}
		situation, situationErr := store.MunicipalitySituation(ctx, SituationQuery{MunicipalityISTAT: "050004", QueryTime: qt})
		if situationErr != nil {
			t.Fatal(situationErr)
		}
		result, searched, situated := map[string][]string{}, []string{}, []string{}
		for _, m := range found.Measures {
			if len(m.Evidence) == 0 {
				t.Fatalf("measure %s without evidence", m.ID)
			}
			result[m.Evidence[0].DocumentID] = append(result[m.Evidence[0].DocumentID], m.Subject+"@"+m.Evidence[0].VersionID)
			searched = append(searched, m.ID)
		}
		for _, m := range situation.LocalMeasures {
			for _, id := range own {
				if m.ID == id {
					situated = append(situated, m.ID)
				}
			}
		}
		sort.Strings(searched)
		sort.Strings(situated)
		if !reflect.DeepEqual(searched, situated) {
			t.Fatalf("search and situation disagree at %ds: %v %v", seconds, searched, situated)
		}
		for _, list := range result {
			sort.Strings(list)
		}
		return result
	}
	doc := func(v documents.Version) string { return stringID(v.DocumentID) }
	of := func(subject string, v documents.Version) string { return subject + "@" + stringID(v.ID) }
	expect := func(t *testing.T, seconds int, want map[string][]string) {
		t.Helper()
		got := current(t, seconds)
		for document, measures := range want {
			sort.Strings(measures)
			if !reflect.DeepEqual(got[document], measures) {
				t.Fatalf("document %s at %ds: got %v, want %v", document, seconds, got[document], measures)
			}
		}
	}
	var before [6]int
	counts := `SELECT (SELECT count(*) FROM domain_local_measures),(SELECT count(*) FROM domain_measure_bindings),(SELECT count(*) FROM extraction_results),(SELECT count(*) FROM interpretation_input_equivalences),(SELECT count(*) FROM domain_interpretation_events),(SELECT count(*) FROM processing_runs)`
	if err = pool.QueryRow(ctx, counts).Scan(&before[0], &before[1], &before[2], &before[3], &before[4], &before[5]); err != nil {
		t.Fatal(err)
	}
	zMeasure := []string{of("ponte sintetico", za)}
	yAll := []string{of("ponte sintetico", ya), of("sottopasso sintetico", ya)}

	t.Run("boundary before the replacing projection", func(t *testing.T) {
		for _, seconds := range []int{20, 35} {
			expect(t, seconds, map[string][]string{doc(xa): {of("ponte sintetico", xa)}, doc(wa): {of("ponte sintetico", wa)}, doc(ya): yAll, doc(za): zMeasure})
		}
	})
	t.Run("template-only version reinterpreted", func(t *testing.T) {
		expect(t, 45, map[string][]string{doc(xa): {of("ponte", xb)}})
	})
	t.Run("boundary before a fingerprint", func(t *testing.T) {
		expect(t, 45, map[string][]string{doc(wa): {of("ponte sintetico", wa), of("ponte", wb)}})
		expect(t, 65, map[string][]string{doc(wa): {of("ponte", wb)}})
	})
	t.Run("genuine revision accumulates", func(t *testing.T) {
		expect(t, 65, map[string][]string{doc(ya): append([]string{of("sottopasso", yb)}, yAll...)})
	})
	t.Run("not relevant, unprojected or failed equivalents replace nothing", func(t *testing.T) {
		expect(t, 65, map[string][]string{doc(za): zMeasure})
	})
	t.Run("same-version replacement preserved", func(t *testing.T) {
		expect(t, 57, map[string][]string{doc(xa): {of("ponte di Via Sintetica", xb)}})
		expect(t, 45, map[string][]string{doc(xa): {of("ponte", xb)}})
	})
	var after [6]int
	if err = pool.QueryRow(ctx, counts).Scan(&after[0], &after[1], &after[2], &after[3], &after[4], &after[5]); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("public reads wrote retained data: %v -> %v", before, after)
	}

	// Bounded selection for many template versions. The fixture helpers fail
	// the enclosing test, so this block is not a subtest.
	{
		var last documents.Version
		for i := 0; i < 40; i++ {
			last = retain("bulk", fmt.Sprintf("template-%d", i), notice)
			project(interpret(fmt.Sprintf("bulk-%d", i), last, 100+3*i, "extracted", true, fmt.Sprintf("ponte sintetico %d", i)), 102+3*i, 1)
		}
		counted := &allPublicReads{Pool: pool, calls: map[string]int{}}
		shared := New(pool)
		shared.pool = counted
		qt := QueryTime{EvaluationTime: at(300), KnownAt: at(300)}
		found, searchErr := shared.Search(ctx, SearchQuery{Kind: "measure", SourceID: sourceID, MunicipalityISTAT: "050004", QueryTime: qt})
		if searchErr != nil {
			t.Fatal(searchErr)
		}
		bulk := []string{}
		for _, m := range found.Measures {
			if m.Evidence[0].DocumentID == doc(last) {
				bulk = append(bulk, m.Subject+"@"+m.Evidence[0].VersionID)
			}
		}
		selections := 0
		for sql, count := range counted.calls {
			if strings.HasPrefix(sql, "SELECT m.id FROM domain_local_measures m") {
				selections += count
			}
		}
		if !reflect.DeepEqual(bulk, []string{of("ponte sintetico 39", last)}) || selections != 1 {
			t.Fatalf("unbounded or wrong selection: measures=%v selection queries=%d", bulk, selections)
		}
	}
}
