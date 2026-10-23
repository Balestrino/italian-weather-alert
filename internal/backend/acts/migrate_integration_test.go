//go:build integration

package acts

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/embedding"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// migrateAll applies every schema the act tables and retention depend on.
func migrateAll(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{
		registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate,
		acquisition.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate,
		linking.Migrate, embedding.Migrate, interpretation.Migrate, domain.Migrate,
		domain.MigrateGeography, domain.MigrateTemporal, domain.MigrateQuality, domain.MigrateRetention, Migrate,
	} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReferencedActTablesAreAppendOnlyAndRetained(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := actsTestDB(t)
	migrateAll(t, ctx, pool)
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("idempotent referenced acts migration: %v", err)
	}
	now := time.Date(2026, 1, 31, 10, 30, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://calcinaia.example/terms", Locator: "acts fixture", ObservedAt: now}
	configuration := registry.Configuration{URL: "https://calcinaia.example", Sections: []string{"https://calcinaia.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "acts-authority", Name: "Comune", OfficialURL: "https://calcinaia.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "acts-channel", PublisherID: "acts-authority", Platform: "municipal-site", URL: "https://calcinaia.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "acts-source", AuthorityID: "acts-authority", ChannelID: "acts-channel", ProductID: "municipal", Territory: "050004"}, configuration, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	objects := &actsObjects{}
	store := documents.New(pool, objects)
	retain := func(path, body string) documents.Version {
		t.Helper()
		url := "https://calcinaia.example/notices/" + path
		v, err := store.Retain(ctx, documents.Acquisition{ID: path, SourceID: "acts-source", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "acts-source", Configuration: 1, MediaType: "text/html", Bytes: []byte(body)}}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `UPDATE retained_versions SET first_acquired_at=$1 WHERE id=$2`, now, v.ID); err != nil {
			t.Fatal(err)
		}
		return v
	}
	// Each pair is a notice with an observation measure and the act clarifying it.
	type pair struct {
		notice, act documents.Version
		measure     string
	}
	seed := func(name string) pair {
		t.Helper()
		p := pair{notice: retain(name+"-notice", "avviso "+name), act: retain(name+"-act", "ordinanza "+name), measure: "acts-" + name}
		if err := domain.New(pool).PutLocalMeasure(ctx, domain.LocalMeasure{ID: p.measure, DocumentVersionID: p.notice.ID, SourceID: "acts-source", MunicipalityISTAT: "050004", Kind: "observation", Subject: "cimiteri"}); err != nil {
			t.Fatal(err)
		}
		var reference, resolution int64
		if err := pool.QueryRow(ctx, `INSERT INTO act_references(document_version_id,ordinal,kind,passage,resource_url,detection_version,detected_at)
 VALUES($1,1,'ordinanza','mediante apposita ordinanza',$2,'test',$3) RETURNING id`, p.notice.ID, p.notice.Resources[0].URL, now).Scan(&reference); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO act_albo_records(source_id,record_id,category,registry_year,registry_number,object,detail_url,document_version_id,recorded_at)
 VALUES('acts-source',$1,'ORDINANZE',2026,1,'EMERGENZA','https://albo.example/display/1',$2,$3)`, name[len(name)-1:], p.act.ID, now); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO act_resolutions(reference_id,outcome,albo_source_id,albo_record_id,recorded_at)
 VALUES($1,'unique_candidate','acts-source',$2,$3) RETURNING id`, reference, name[len(name)-1:], now).Scan(&resolution); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO act_clarifications(resolution_id,act_document_version_id,local_measure_id,outcome,disposition_ordinal,act_resource_url,act_passage,parser_version,recorded_at)
 VALUES($1,$2,$3,'clarified',1,$4,'a partire dalle ore 00.00','test',$5)`, resolution, p.act.ID, p.measure, p.act.Resources[0].URL, now); err != nil {
			t.Fatal(err)
		}
		return p
	}
	expired, protected := seed("expired-1"), seed("protected-2")
	for table, column := range map[string]string{"act_references": "detected_at", "act_albo_records": "recorded_at", "act_resolutions": "recorded_at", "act_clarifications": "recorded_at"} {
		if _, err := pool.Exec(ctx, "UPDATE "+table+" SET "+column+"="+column); err == nil {
			t.Fatalf("%s accepted an update", table)
		}
		if _, err := pool.Exec(ctx, "DELETE FROM "+table); err == nil {
			t.Fatalf("%s accepted a delete", table)
		}
	}
	if _, err := pool.Exec(ctx, `INSERT INTO act_resolutions(reference_id,outcome,recorded_at) SELECT id,'ambiguous',now() FROM act_references LIMIT 1`); err == nil {
		t.Fatal("ambiguous resolution accepted without a reason")
	}

	retention := domain.NewRetention(pool, objects)
	if err := retention.Protect(ctx, protected.notice.ID, "query_view", "acts-protected", nil, now); err != nil {
		t.Fatal(err)
	}
	if _, err := retention.Cleanup(ctx, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	exists := func(id int64) bool {
		var found bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM retained_versions WHERE id=$1)`, id).Scan(&found); err != nil {
			t.Fatal(err)
		}
		return found
	}
	if exists(expired.notice.ID) || exists(expired.act.ID) {
		t.Fatal("expired notice or its act retained")
	}
	if !exists(protected.notice.ID) || !exists(protected.act.ID) {
		t.Fatal("act of a retained notice removed")
	}
	var references, clarifications int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM act_references),(SELECT count(*) FROM act_clarifications)`).Scan(&references, &clarifications); err != nil || references != 1 || clarifications != 1 {
		t.Fatalf("retention left act rows: references=%d clarifications=%d err=%v", references, clarifications, err)
	}
}
