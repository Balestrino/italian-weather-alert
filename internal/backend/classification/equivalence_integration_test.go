//go:build integration

package classification

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestEquivalenceTablePersistsAndBackfillsFingerprints(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := classificationTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, ocr.Migrate, Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	var applied int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM iwa_migrations WHERE name='100_interpretation_input_equivalences'").Scan(&applied); err != nil || applied != 1 {
		t.Fatal("equivalence migration not applied exactly once", applied, err)
	}
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://comune.example/policy", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "eq-authority", Name: "Equivalence municipality", OfficialURL: "https://comune.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "eq-channel", PublisherID: "eq-authority", Platform: "fixture", URL: "https://comune.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "calcinaia-municipal", AuthorityID: "eq-authority", ChannelID: "eq-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://comune.example", Sections: []string{"https://comune.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	retained := documents.New(pool, &integrationObjects{})
	const url = "https://comune.example/avviso"
	retain := func(id, html string) documents.Version {
		t.Helper()
		v, err := retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: "calcinaia-municipal", Configuration: 1, URL: url, Metadata: json.RawMessage(`{"source_publication_date":"2026-10-07"}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "calcinaia-municipal", Configuration: 1, MediaType: "text/html", Bytes: []byte(html)}}})
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	const page = `<main class="%s"><h1>Allerta meteo</h1><p>Il Comune dispone la chiusura del sottopasso dalle ore %d.</p></main>`
	first := retain("eq-1", fmt.Sprintf(page, "v1", 18))
	template := retain("eq-2", fmt.Sprintf(page, "template-2026-10-21", 18))
	changed := retain("eq-3", fmt.Sprintf(page, "v1", 19))
	if first.ID == template.ID || first.ID == changed.ID {
		t.Fatal("fixture versions collapsed")
	}
	process := processing.New(pool)
	legacy, err := RegisterLegacyCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil {
		t.Fatal(err)
	}
	current, err := RegisterCatalog(ctx, process, "openai-chat", "qwen3.8-27b", now)
	if err != nil || current.ConfigurationVersionID == legacy.ConfigurationVersionID {
		t.Fatal("fixture needs two configurations", err)
	}
	store := NewStore(pool)
	put := func(key string, v documents.Version, configuration string, reusable bool) int64 {
		t.Helper()
		source := "calcinaia-municipal"
		run, runErr := process.StartRun(ctx, processing.RunRequest{IdempotencyKey: key, Workload: "ordinary", Stage: "classification", ConfigurationVersionID: configuration, SourceID: &source, DocumentVersionID: &v.ID, Subject: json.RawMessage(`{"fixture":"equivalence"}`), CreatedAt: now})
		if runErr != nil {
			t.Fatal(runErr)
		}
		content, gatherErr := GatherContent(ctx, retained, ocr.NewStore(pool), v)
		if gatherErr != nil {
			t.Fatal(gatherErr)
		}
		m, buildErr := BuildManifest(ctx, retained, ocr.NewStore(pool), v, content, configuration)
		if buildErr != nil {
			t.Fatal(buildErr)
		}
		m.Reusable = m.Reusable && reusable
		for i := 0; i < 2; i++ {
			if putErr := store.PutManifest(ctx, run.ID, m); putErr != nil {
				t.Fatal(putErr)
			}
		}
		return run.ID
	}
	runFirst := put("eq-run-1", first, legacy.ConfigurationVersionID, true)
	runTemplate := put("eq-run-2", template, current.ConfigurationVersionID, true)
	runChanged := put("eq-run-3", changed, current.ConfigurationVersionID, true)
	runOpaque := put("eq-run-4", first, current.ConfigurationVersionID, false)
	fingerprints := func() map[int64]string {
		t.Helper()
		rows, queryErr := pool.Query(ctx, "SELECT e.run_id,e.fingerprint FROM interpretation_input_equivalences e JOIN processing_runs r ON r.id=e.run_id WHERE e.version_id=r.document_version_id AND e.fingerprint_version=$1", EquivalenceVersion)
		if queryErr != nil {
			t.Fatal(queryErr)
		}
		defer rows.Close()
		values := map[int64]string{}
		for rows.Next() {
			var run int64
			var value string
			if scanErr := rows.Scan(&run, &value); scanErr != nil {
				t.Fatal(scanErr)
			}
			values[run] = value
		}
		return values
	}
	stored := fingerprints()
	if len(stored) != 3 || stored[runFirst] == "" || stored[runFirst] != stored[runTemplate] || stored[runChanged] == stored[runFirst] || stored[runOpaque] != "" {
		t.Fatal("unexpected stored fingerprints", stored)
	}
	for _, statement := range []string{
		"UPDATE interpretation_input_equivalences SET fingerprint=repeat('0',64) WHERE run_id=" + strconv.FormatInt(runFirst, 10),
		"DELETE FROM interpretation_input_equivalences WHERE run_id=" + strconv.FormatInt(runFirst, 10),
		"INSERT INTO interpretation_input_equivalences(run_id,document_id,version_id,fingerprint,fingerprint_version,created_at) VALUES (999999," + strconv.FormatInt(first.DocumentID, 10) + "," + strconv.FormatInt(first.ID, 10) + ",repeat('0',64),'x',now())",
	} {
		if _, execErr := pool.Exec(ctx, statement); execErr == nil {
			t.Fatal("equivalence row accepted an unsafe write", statement)
		}
	}

	// Simulate manifests recorded before the migration, then rebuild their rows.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	for _, statement := range []string{
		"ALTER TABLE interpretation_input_equivalences DISABLE TRIGGER interpretation_input_equivalence_immutable",
		"DELETE FROM interpretation_input_equivalences",
		"ALTER TABLE interpretation_input_equivalences ENABLE TRIGGER interpretation_input_equivalence_immutable",
	} {
		if _, err = tx.Exec(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	total := EquivalenceBackfill{}
	for after, pages := int64(0), 0; ; pages++ {
		report, backfillErr := store.BackfillEquivalences(ctx, after, 2)
		if backfillErr != nil || pages > 4 {
			t.Fatal("backfill did not finish", backfillErr)
		}
		total.Scanned += report.Scanned
		total.Written += report.Written
		total.NotReusable += report.NotReusable
		after = report.NextAfterRun
		if report.Complete {
			break
		}
	}
	if total.Scanned != 4 || total.Written != 3 || total.NotReusable != 1 {
		t.Fatalf("unexpected backfill totals: %#v", total)
	}
	rebuilt := fingerprints()
	for run, value := range stored {
		if rebuilt[run] != value {
			t.Fatal("backfill changed a fingerprint", run)
		}
	}
	again, err := store.BackfillEquivalences(ctx, 0, 1000)
	if err != nil || again.Written != 0 || again.AlreadyStored != 3 || again.NotReusable != 1 || !again.Complete {
		t.Fatalf("backfill rerun wrote rows: %#v %v", again, err)
	}
	if _, err = store.BackfillEquivalences(ctx, -1, 10); err == nil {
		t.Fatal("negative cursor accepted")
	}
}
