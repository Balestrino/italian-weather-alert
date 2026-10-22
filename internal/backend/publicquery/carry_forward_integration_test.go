//go:build integration

package publicquery

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
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// allPublicReads counts every round trip, including multi-row queries.
type allPublicReads struct {
	*pgxpool.Pool
	calls map[string]int
}

func (p *allPublicReads) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	p.calls[sql]++
	return p.Pool.Query(ctx, sql, args...)
}
func (p *allPublicReads) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	p.calls[sql]++
	return p.Pool.QueryRow(ctx, sql, args...)
}

func TestCarryForwardEquivalentInterpretations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := publicQueryTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, acquisition.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, domain.Migrate, domain.MigrateGeography, domain.MigrateTemporal, domain.MigrateQuality, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Truncate(time.Second)
	reg := registry.New(pool)
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "comune-calcinaia", Name: "Comune di Calcinaia", OfficialURL: "https://calcinaia.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "calcinaia", PublisherID: "comune-calcinaia", Platform: "municipal-site", URL: "https://calcinaia.example"}),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"carry-public", "carry-defective"} {
		section := "https://calcinaia.example/" + id
		evidence := registry.Evidence{URL: section, Locator: "reviewed fixture evidence", ObservedAt: now}
		configuration := registry.Configuration{URL: section, Sections: []string{section}, AccessMethod: "fixture", Attribution: "comune-calcinaia", Provenance: &evidence, Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true, PublicationPermitted: true}, Limitations: []string{"fixture scope only"}}
		acceptance := registry.Acceptance{Report: evidence, PeriodStart: now.Add(-9 * 24 * time.Hour), PeriodEnd: now.Add(-time.Hour), Sections: []string{section}, ExtractionVerified: true, UpdatesVerified: true, AttachmentsVerified: true, ScannedAttachmentsVerified: true, HistoryVerified: true, FailureBehaviorVerified: true, InterfacesEquivalent: true, CoverageStatus: "accepted_with_limitations", CoverageLimitations: []string{"fixture scope only"}}
		for _, err := range []error{
			reg.CreateSource(ctx, registry.Source{ID: id, AuthorityID: "comune-calcinaia", ChannelID: "calcinaia", ProductID: "municipal", Territory: "050004"}, configuration, "test"),
			reg.RecordPreview(ctx, id, 1, "test", evidence),
			reg.EnableCollection(ctx, id, 1, "test"),
			reg.Accept(ctx, id, 1, "test", acceptance),
		} {
			if err != nil {
				t.Fatal(err)
			}
		}
		if review, err := reg.ReviewAcceptance(ctx, id, 1, "release-reviewer", registry.Evidence{URL: section, Locator: "fixture review", ObservedAt: now}); err != nil || review.Status != "accepted" {
			t.Fatal("review", review, err)
		}
		if err := reg.EnablePublic(ctx, id, 1, "test"); err != nil {
			t.Fatal(err)
		}
	}
	retained := documents.New(pool, &memoryObjects{})
	const notice = `<main class="%s"><h1>Allerta meteo</h1><p>Il Comune dispone la chiusura del sottopasso dalle ore 18.</p></main>`
	retain := func(source, url, class string) documents.Version {
		t.Helper()
		v, err := retained.Retain(ctx, documents.Acquisition{ID: source + url + class, SourceID: source, Configuration: 1, URL: url, Metadata: json.RawMessage(`{"source_publication_date":"2026-10-07"}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: source, Configuration: 1, MediaType: "text/html", Bytes: []byte(fmt.Sprintf(notice, class))}}})
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	const urlA, urlB, urlC, urlD = "https://calcinaia.example/avviso-a", "https://calcinaia.example/avviso-b", "https://calcinaia.example/avviso-c", "https://calcinaia.example/avviso-d"
	a1 := retain("carry-public", urlA, "v1")
	a2 := retain("carry-public", urlA, "template-2")
	a3 := retain("carry-public", urlA, "template-3")
	b1 := retain("carry-public", urlB, "v1")
	d1 := retain("carry-public", urlD, "v1")
	c1 := retain("carry-defective", urlC, "v1")
	c2 := retain("carry-defective", urlC, "template-2")
	base := time.Now().UTC().Add(time.Second)
	at := func(seconds int) time.Time { return base.Add(time.Duration(seconds) * time.Second) }

	process := processing.New(pool)
	oldClass, err := classification.RegisterLegacyCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	newClass, err := classification.RegisterCatalog(ctx, process, "local-openai-chat", "qwen3.5-9b", now)
	if err != nil {
		t.Fatal(err)
	}
	oldExtract, err := extraction.RegisterCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	newExtract, err := extraction.RegisterCatalog(ctx, process, "local-openai-chat", "qwen3.5-9b", now)
	if err != nil || oldExtract.ConfigurationVersionID == newExtract.ConfigurationVersionID || oldClass.ConfigurationVersionID == newClass.ConfigurationVersionID {
		t.Fatal("fixture needs distinct configurations", err)
	}
	manifests := classification.NewStore(pool)
	hash := strings.Repeat("e", 64)
	startRun := func(key, stage, configuration string, v documents.Version, created time.Time) int64 {
		t.Helper()
		source := v.Resources[0].SourceID
		run, runErr := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: key, Workload: "ordinary", Stage: stage, ConfigurationVersionID: configuration, SourceID: &source, DocumentVersionID: &v.ID, Subject: json.RawMessage(`{"fixture":"carry-forward"}`), CreatedAt: created})
		if runErr != nil {
			t.Fatal(runErr)
		}
		content, gatherErr := classification.GatherContent(ctx, retained, ocr.NewStore(pool), v)
		if gatherErr != nil {
			t.Fatal(gatherErr)
		}
		m, buildErr := classification.BuildManifest(ctx, retained, ocr.NewStore(pool), v, content, configuration)
		if buildErr != nil {
			t.Fatal(buildErr)
		}
		if putErr := manifests.PutManifest(ctx, run.ID, m); putErr != nil {
			t.Fatal(putErr)
		}
		return run.ID
	}
	classify := func(key, configuration string, v documents.Version, created time.Time, relevant *bool) int64 {
		t.Helper()
		run := startRun(key, "classification", configuration, v, created)
		if relevant != nil {
			reason := "local_weather_measure"
			if !*relevant {
				reason = "not_relevant"
			}
			if _, execErr := pool.Exec(ctx, `INSERT INTO classification_results(run_id,document_version_id,status,relevant,reason_code,evidence_quote,content_sha256,content_complete,provider_response_id,returned_model,created_at) VALUES($1,$2,'classified',$3,$4,'chiusura del sottopasso',$5,true,'fixture','fixture-model',$6)`, run, v.ID, *relevant, reason, hash, created.Add(5*time.Second)); execErr != nil {
				t.Fatal(execErr)
			}
		}
		return run
	}
	extract := func(key, configuration string, v documents.Version, classRun int64, created time.Time, status, reason string) int64 {
		t.Helper()
		run := startRun(key, "extraction", configuration, v, created)
		var provider, model *string
		if status == "extracted" {
			value, name := "fixture", "fixture-model"
			provider, model = &value, &name
		}
		if _, execErr := pool.Exec(ctx, `INSERT INTO extraction_results(run_id,document_version_id,classification_run_id,status,reason_code,content_sha256,content_complete,provider_response_id,returned_model,created_at) VALUES($1,$2,$3,$4,$5,$6,true,$7,$8,$9)`, run, v.ID, classRun, status, reason, hash, provider, model, created.Add(10*time.Second)); execErr != nil {
			t.Fatal(execErr)
		}
		return run
	}
	relevant, irrelevant := true, false
	// a1: validated at 60s. a2: pending from 40s, failed at 80s, own not-relevant result at 200s.
	a1Class := classify("a1-class", oldClass.ConfigurationVersionID, a1, at(10), &relevant)
	a1Run := extract("a1-extract", oldExtract.ConfigurationVersionID, a1, a1Class, at(50), "extracted", "measures_extracted")
	a2Class := classify("a2-class", newClass.ConfigurationVersionID, a2, at(40), &relevant)
	extract("a2-extract", newExtract.ConfigurationVersionID, a2, a2Class, at(70), "uninterpreted", "extraction_evidence_invalid")
	a2Later := classify("a2-class-later", newClass.ConfigurationVersionID, a2, at(180), &irrelevant)
	extract("a2-extract-later", newExtract.ConfigurationVersionID, a2, a2Later, at(190), "not_applicable", "not_relevant")
	// a3: classification rejected, so only its manifest exists.
	classify("a3-class", newClass.ConfigurationVersionID, a3, at(85), nil)
	// b1: failed without any equivalent. d1: another document with a1's text.
	b1Class := classify("b1-class", newClass.ConfigurationVersionID, b1, at(40), &relevant)
	extract("b1-extract", newExtract.ConfigurationVersionID, b1, b1Class, at(70), "uninterpreted", "extraction_evidence_invalid")
	d1Class := classify("d1-class", newClass.ConfigurationVersionID, d1, at(40), &relevant)
	extract("d1-extract", newExtract.ConfigurationVersionID, d1, d1Class, at(70), "uninterpreted", "extraction_evidence_invalid")
	// c1 is validated before a confirmed defect (100–110s); c2 fails afterwards.
	c1Class := classify("c1-class", oldClass.ConfigurationVersionID, c1, at(10), &relevant)
	extract("c1-extract", oldExtract.ConfigurationVersionID, c1, c1Class, at(50), "extracted", "measures_extracted")
	if _, err = pool.Exec(ctx, `INSERT INTO registry_interpretation_suspensions(source_id,revision,suspended_at,actor,defect,resumed_at,resumed_by,recovery) VALUES('carry-defective',1,$1,'fixture','{"url":"https://calcinaia.example/defect"}',$2,'fixture','{"url":"https://calcinaia.example/recovery"}')`, at(100), at(110)); err != nil {
		t.Fatal(err)
	}
	c2Class := classify("c2-class", newClass.ConfigurationVersionID, c2, at(120), &relevant)
	extract("c2-extract", newExtract.ConfigurationVersionID, c2, c2Class, at(130), "uninterpreted", "extraction_evidence_invalid")

	store := New(pool)
	read := func(seconds int, v documents.Version) Document {
		t.Helper()
		qt := QueryTime{EvaluationTime: at(seconds), KnownAt: at(seconds)}
		result, readErr := store.Document(ctx, DocumentQuery{QueryTime: qt, DocumentID: stringID(v.DocumentID), VersionID: stringID(v.ID)})
		if readErr != nil {
			t.Fatal(readErr)
		}
		return result.Document
	}
	carried := func(d Document, own string) bool {
		i := d.Quality.Interpretation
		return i.State == "supported" && len(i.Limitations) >= 2 &&
			i.Limitations[0] == fmt.Sprintf("carried_forward_from_equivalent_version:%d:run:%d:configuration:%s", a1.ID, a1Run, oldExtract.ConfigurationVersionID) &&
			i.Limitations[1] == own && (d.VersionID == stringID(a2.ID) || d.VersionID == stringID(a3.ID)) &&
			d.InterpretationRunID != nil && *d.InterpretationRunID == stringID(a1Run) && len(i.Evidence) == 1 && i.Evidence[0].VersionID == stringID(a1.ID) && *i.Evidence[0].Locator == carriedForwardLocator
	}
	var before [7]int
	counts := `SELECT (SELECT count(*) FROM classification_results),(SELECT count(*) FROM extraction_results),(SELECT count(*) FROM interpretation_reuse),(SELECT count(*) FROM domain_local_measures),(SELECT count(*) FROM domain_interpretation_events),(SELECT count(*) FROM processing_runs),(SELECT count(*) FROM interpretation_input_equivalences)`
	if err = pool.QueryRow(ctx, counts).Scan(&before[0], &before[1], &before[2], &before[3], &before[4], &before[5], &before[6]); err != nil {
		t.Fatal(err)
	}

	t.Run("boundary before the earlier result", func(t *testing.T) {
		if d := read(55, a2); d.Quality.Interpretation.State != "not_processed" || strings.HasPrefix(strings.Join(d.Quality.Interpretation.Limitations, ","), "carried_forward") {
			t.Fatalf("carried before the earlier result existed: %#v", d.Quality.Interpretation)
		}
	})
	t.Run("pending version with a manifest", func(t *testing.T) {
		if d := read(65, a2); !carried(d, "own_interpretation:not_available") {
			t.Fatalf("pending version not carried forward: %#v", d.Quality.Interpretation)
		}
	})
	t.Run("failed version", func(t *testing.T) {
		d := read(90, a2)
		if !carried(d, "own_interpretation:failed:extraction_evidence_invalid") || d.SHA256 != a2.Hash {
			t.Fatalf("failed version not carried forward: %#v", d)
		}
	})
	t.Run("rejected classification", func(t *testing.T) {
		if d := read(95, a3); !carried(d, "own_interpretation:not_available") {
			t.Fatalf("rejected classification not carried forward: %#v", d.Quality.Interpretation)
		}
	})
	t.Run("own later result takes precedence", func(t *testing.T) {
		d := read(300, a2)
		if d.Quality.Interpretation.State != "supported" || strings.HasPrefix(d.Quality.Interpretation.Limitations[0], "carried_forward") || d.InterpretationRunID == nil || *d.InterpretationRunID == stringID(a1Run) {
			t.Fatalf("own result hidden by carry-forward: %#v", d)
		}
	})
	t.Run("no equivalent, other document or unreliable earlier", func(t *testing.T) {
		for _, v := range []documents.Version{b1, d1, c2} {
			if d := read(300, v); d.Quality.Interpretation.State != "failed" || len(d.Quality.Interpretation.Evidence) != 0 {
				t.Fatalf("version %d unexpectedly carried: %#v", v.ID, d.Quality.Interpretation)
			}
		}
		if d := read(300, c1); d.Quality.Interpretation.State != "unreliable" {
			t.Fatalf("fixture earlier interpretation is not unreliable: %#v", d.Quality.Interpretation)
		}
	})
	t.Run("search, versions and detail agree", func(t *testing.T) {
		qt := QueryTime{EvaluationTime: at(90), KnownAt: at(90)}
		detail, readErr := store.Document(ctx, DocumentQuery{QueryTime: qt, DocumentID: stringID(a1.DocumentID), IncludeVersions: true})
		if readErr != nil {
			t.Fatal(readErr)
		}
		search, searchErr := store.Search(ctx, SearchQuery{QueryTime: qt, Kind: "document", SourceID: "carry-public"})
		if searchErr != nil {
			t.Fatal(searchErr)
		}
		bySearch := map[string]Document{}
		for _, d := range search.Documents {
			bySearch[d.VersionID] = d
		}
		found := 0
		for _, d := range append(detail.Versions, detail.Document) {
			if d.VersionID != stringID(a2.ID) {
				continue
			}
			found++
			if !carried(d, "own_interpretation:failed:extraction_evidence_invalid") {
				t.Fatalf("versions list differs: %#v", d.Quality.Interpretation)
			}
		}
		if found != 1 {
			t.Fatal("a2 missing from the versions list")
		}
		fromSearch, _ := json.Marshal(bySearch[stringID(a3.ID)].Quality.Interpretation)
		fromDetail, _ := json.Marshal(detail.Document.Quality.Interpretation)
		if s, ok := bySearch[stringID(a3.ID)]; !ok || string(fromSearch) != string(fromDetail) {
			t.Fatalf("search and detail disagree for the latest version: %#v %#v", s.Quality.Interpretation, detail.Document.Quality.Interpretation)
		}
	})
	t.Run("bounded reads for many template versions", func(t *testing.T) {
		for i := 0; i < 40; i++ {
			v := retain("carry-public", urlA, fmt.Sprintf("template-bulk-%d", i))
			classify(fmt.Sprintf("bulk-class-%d", i), newClass.ConfigurationVersionID, v, at(150), nil)
		}
		counted := &allPublicReads{Pool: pool, calls: map[string]int{}}
		shared := New(pool)
		shared.pool = counted
		qt := QueryTime{EvaluationTime: at(160), KnownAt: at(160)}
		detail, readErr := shared.Document(ctx, DocumentQuery{QueryTime: qt, DocumentID: stringID(a1.DocumentID), IncludeVersions: true})
		if readErr != nil || len(detail.Versions) != 43 {
			t.Fatal("bulk versions unavailable", len(detail.Versions), readErr)
		}
		carriedVersions := 0
		for _, d := range detail.Versions {
			if strings.HasPrefix(strings.Join(d.Quality.Interpretation.Limitations, ","), "carried_forward_from_equivalent_version:") {
				carriedVersions++
			}
		}
		assessments, candidates := 0, 0
		for sql, count := range counted.calls {
			if strings.HasPrefix(sql, "WITH versions AS MATERIALIZED") {
				assessments += count
			}
			if sql == equivalentCandidatesSQL {
				candidates += count
			}
		}
		if carriedVersions != 42 || assessments > 4 || candidates > 2 {
			t.Fatalf("unbounded carry-forward: carried=%d assessments=%d candidate queries=%d", carriedVersions, assessments, candidates)
		}
	})
	var after [7]int
	if err = pool.QueryRow(ctx, counts).Scan(&after[0], &after[1], &after[2], &after[3], &after[4], &after[5], &after[6]); err != nil {
		t.Fatal(err)
	}
	// The bulk subtest adds 40 classification runs and manifests of its own.
	after[5] -= 40
	after[6] -= 40
	if after != before {
		t.Fatalf("public reads wrote retained data: %v -> %v", before, after)
	}
	t.Run("active suspension cutoff", func(t *testing.T) {
		// a1's validated result (60s) lies after this still-open suspension.
		if _, execErr := pool.Exec(ctx, `INSERT INTO registry_interpretation_suspensions(source_id,revision,suspended_at,actor,defect) VALUES('carry-public',1,$1,'fixture','{"url":"https://calcinaia.example/open-defect"}')`, at(58)); execErr != nil {
			t.Fatal(execErr)
		}
		if d := read(90, a2); d.Quality.Interpretation.State == "supported" || strings.HasPrefix(strings.Join(d.Quality.Interpretation.Limitations, ","), "carried_forward") {
			t.Fatalf("carried a result beyond the source interpretation cutoff: %#v", d.Quality.Interpretation)
		}
	})
}
