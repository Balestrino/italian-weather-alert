//go:build integration

package domain

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/embedding"
	"github.com/Balestrino/italian-weather-alert/internal/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/linking"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

type retentionObjects struct {
	*domainObjects
	failures int
}

func (o *retentionObjects) Delete(_ context.Context, hash string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.failures > 0 {
		o.failures--
		return errors.New("injected object deletion failure")
	}
	delete(o.items, hash)
	return nil
}

func TestRetentionCalendarPolicyProtectionCessationAndRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := domainTestDB(t)
	migrations := []func(context.Context, *pgxpool.Pool) error{
		registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate,
		acquisition.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate,
		linking.Migrate, embedding.Migrate, interpretation.Migrate, Migrate,
		MigrateGeography, MigrateTemporal, MigrateQuality, MigrateRetention,
	}
	for _, migrate := range migrations {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if err := MigrateRetention(ctx, pool); err != nil {
		t.Fatalf("idempotent retention migration: %v", err)
	}

	now := time.Date(2026, 1, 31, 10, 30, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://calcinaia.example/terms", Locator: "retention fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "retention-authority", Name: "Comune", OfficialURL: "https://calcinaia.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "retention-channel", PublisherID: "retention-authority", Platform: "municipal-site", URL: "https://calcinaia.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "retention-source", AuthorityID: "retention-authority", ChannelID: "retention-channel", ProductID: "municipal", Territory: "050004"}, registry.Configuration{URL: "https://calcinaia.example", Sections: []string{"https://calcinaia.example/notices"}, AccessMethod: "fixture", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	objects := &retentionObjects{domainObjects: &domainObjects{items: map[string][]byte{}}}
	documentsStore := documents.New(pool, objects)
	retain := func(id, path, body string) documents.Version {
		t.Helper()
		url := "https://calcinaia.example/notices/" + path
		version, err := documentsStore.Retain(ctx, documents.Acquisition{ID: id, SourceID: "retention-source", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "retention-source", Configuration: 1, MediaType: "text/html", Bytes: []byte(body)}}})
		if err != nil {
			t.Fatal(err)
		}
		return version
	}

	ordinary := retain("ordinary-first", "ordinary", "ordinary old notice")
	unchanged := retain("ordinary-unchanged", "ordinary", "ordinary old notice")
	if unchanged.ID != ordinary.ID || !unchanged.FirstAcquiredAt.Equal(ordinary.FirstAcquiredAt) {
		t.Fatal("unchanged check reset first acquisition")
	}
	active := retain("active-first", "active", "closure until further notice")
	view := retain("view-first", "view", "version held by a public query view")
	for _, version := range []documents.Version{ordinary, active, view} {
		if _, err := pool.Exec(ctx, `UPDATE retained_versions SET first_acquired_at=$1 WHERE id=$2`, now, version.ID); err != nil {
			t.Fatal(err)
		}
	}
	store := New(pool)
	measure := LocalMeasure{ID: "retention-active-closure", DocumentVersionID: active.ID, SourceID: "retention-source", MunicipalityISTAT: "050004", Kind: "closure", Subject: "sottopasso"}
	if err := store.PutLocalMeasure(ctx, measure); err != nil {
		t.Fatal(err)
	}
	seedRetentionDependencyGraph(t, ctx, pool, active.ID, "https://calcinaia.example/notices/active", now, measure.ID)
	retention := NewRetention(pool, objects)
	policy, err := retention.CurrentPolicy(ctx)
	if err != nil || policy.Months != 3 || policy.Actor != "migration-default" {
		t.Fatalf("initial retention policy: %#v %v", policy, err)
	}
	viewThrough := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	if err := retention.Protect(ctx, view.ID, "query_view", "view-42", &viewThrough, now); err != nil {
		t.Fatal(err)
	}

	beforeBoundary, err := retention.Cleanup(ctx, time.Date(2026, 4, 30, 10, 29, 59, 0, time.UTC))
	if err != nil || beforeBoundary.DeletedVersions != 0 {
		t.Fatalf("deleted before calendar boundary: %#v %v", beforeBoundary, err)
	}
	if _, err = retention.SetPolicy(ctx, 6, "operator", time.Date(2026, 4, 30, 10, 29, 59, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	longer, err := retention.Cleanup(ctx, time.Date(2026, 4, 30, 10, 30, 0, 0, time.UTC))
	if err != nil || longer.DeletedVersions != 0 {
		t.Fatalf("new duration did not apply to existing versions: %#v %v", longer, err)
	}
	if _, err = retention.SetPolicy(ctx, 3, "operator", time.Date(2026, 4, 30, 10, 30, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	objects.failures = 1
	atBoundary, err := retention.Cleanup(ctx, time.Date(2026, 4, 30, 10, 30, 0, 0, time.UTC))
	if err != nil || atBoundary.DeletedVersions != 1 || atBoundary.FailedObjects != 1 || atBoundary.PendingObjects != 1 {
		t.Fatalf("boundary cleanup or durable retry queue: %#v %v", atBoundary, err)
	}
	var exists bool
	if err = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM retained_versions WHERE id=$1)`, ordinary.ID).Scan(&exists); err != nil || exists {
		t.Fatalf("ordinary expired version retained: exists=%v err=%v", exists, err)
	}
	for _, id := range []int64{active.ID, view.ID} {
		if err = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM retained_versions WHERE id=$1)`, id).Scan(&exists); err != nil || !exists {
			t.Fatalf("protected version %d removed: exists=%v err=%v", id, exists, err)
		}
	}
	retry, err := retention.Cleanup(ctx, time.Date(2026, 4, 30, 10, 31, 0, 0, time.UTC))
	if err != nil || retry.DeletedObjects != 1 || retry.PendingObjects != 0 {
		t.Fatalf("object deletion was not retried: %#v %v", retry, err)
	}

	start, end := now, time.Date(2026, 4, 29, 23, 59, 0, 0, time.UTC)
	if _, err = store.PutTemporal(ctx, TemporalValue{EntityKind: "local_measure", EntityID: measure.ID, Meaning: "validity", Original: "through 29 April", Precision: "interval", Instant: &start, EndInstant: &end, EvidenceDocumentVersionID: active.ID, CreatedAt: end}); err != nil {
		t.Fatal(err)
	}
	afterCessation, err := retention.Cleanup(ctx, time.Date(2026, 4, 30, 10, 32, 0, 0, time.UTC))
	if err != nil || afterCessation.DeletedVersions != 1 || afterCessation.DeletedObjects != 1 {
		t.Fatalf("ceased measure evidence remained protected: %#v %v", afterCessation, err)
	}
	var measures, temporal int
	if err = pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM domain_local_measures WHERE id=$1),(SELECT count(*) FROM domain_temporal_values WHERE entity_id=$1)`, measure.ID).Scan(&measures, &temporal); err != nil || measures != 0 || temporal != 0 {
		t.Fatalf("dependent graph survived cleanup: measures=%d temporal=%d err=%v", measures, temporal, err)
	}
	var derived int
	if err = pool.QueryRow(ctx, `SELECT
	 (SELECT count(*) FROM processing_runs WHERE document_version_id=$1)+
	 (SELECT count(*) FROM classification_results WHERE document_version_id=$1)+
	 (SELECT count(*) FROM classification_segments WHERE document_version_id=$1)+
	 (SELECT count(*) FROM extraction_results WHERE document_version_id=$1)+
	 (SELECT count(*) FROM extraction_segments WHERE document_version_id=$1)+
	 (SELECT count(*) FROM ocr_resource_results WHERE document_version_id=$1)+
	 (SELECT count(*) FROM interpretation_triggers WHERE document_version_id=$1)`, active.ID).Scan(&derived); err != nil || derived != 0 {
		t.Fatalf("processing dependency graph survived cleanup: rows=%d err=%v", derived, err)
	}
	var extractionTemporal int
	if err = pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM extracted_temporal_candidates)+(SELECT count(*) FROM extraction_temporal_candidate_evidence)`).Scan(&extractionTemporal); err != nil || extractionTemporal != 0 {
		t.Fatalf("temporal candidate dependency graph survived cleanup: rows=%d err=%v", extractionTemporal, err)
	}

	afterViewExpiry, err := retention.Cleanup(ctx, time.Date(2026, 5, 1, 0, 0, 1, 0, time.UTC))
	if err != nil || afterViewExpiry.DeletedVersions != 1 {
		t.Fatalf("expired query-view protection remained: %#v %v", afterViewExpiry, err)
	}
}

func seedRetentionDependencyGraph(t *testing.T, ctx context.Context, pool *pgxpool.Pool, versionID int64, resourceURL string, at time.Time, measureID string) {
	t.Helper()
	hash := "1111111111111111111111111111111111111111111111111111111111111111"
	if _, err := pool.Exec(ctx, `INSERT INTO processing_configuration_versions(id,name,stage,revision,logic_version,settings,content_hash,created_at)
	 VALUES('retention-config','retention-fixture','classification','v1','fixture-v1','{}',$1,$2)`, hash, at); err != nil {
		t.Fatal(err)
	}
	newRun := func(key, stage string) int64 {
		t.Helper()
		var id int64
		if err := pool.QueryRow(ctx, `INSERT INTO processing_runs(idempotency_key,request_hash,workload,stage,configuration_version_id,document_version_id,subject,created_at)
		 VALUES($1,$2,'evaluation',$3,'retention-config',$4,'{}',$5) RETURNING id`, key, hash, stage, versionID, at).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	classificationRun := newRun("retention-classification", "classification")
	extractionRun := newRun("retention-extraction", "extraction")
	linkingRun := newRun("retention-linking", "linking")
	embeddingRun := newRun("retention-embedding", "embedding")
	ocrRun := newRun("retention-ocr", "ocr")
	statements := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO classification_results(run_id,document_version_id,status,relevant,reason_code,evidence_quote,content_sha256,content_complete,provider_response_id,returned_model,created_at) VALUES($1,$2,'classified',true,'local_weather_measure','closure',$3,true,'fixture','fixture',$4)`, []any{classificationRun, versionID, hash, at}},
		{`INSERT INTO classification_segments(run_id,ordinal,total,document_version_id,resource_url,role,start_byte,end_byte,content_sha256,response_sha256,relevant,reason_code,evidence_quote,provider_response_id,returned_model) VALUES($1,1,1,$2,$3,'original',0,7,$4,$4,true,'local_weather_measure','closure','fixture','fixture')`, []any{classificationRun, versionID, resourceURL, hash}},
		{`INSERT INTO extraction_results(run_id,document_version_id,classification_run_id,status,reason_code,content_sha256,content_complete,provider_response_id,returned_model,created_at) VALUES($1,$2,$3,'extracted','measures_extracted',$4,true,'fixture','fixture',$5)`, []any{extractionRun, versionID, classificationRun, hash, at}},
		{`INSERT INTO extraction_segments(run_id,ordinal,total,document_version_id,resource_url,role,start_byte,end_byte,content_sha256,response_sha256,provider_response_id,returned_model,measure_count) VALUES($1,1,1,$2,$3,'original',0,7,$4,$4,'fixture','fixture',1)`, []any{extractionRun, versionID, resourceURL, hash}},
		{`INSERT INTO extracted_measures(run_id,ordinal,kind,subject,indeterminate_fields) VALUES($1,1,'closure','sottopasso','["place","valid_from","valid_until"]')`, []any{extractionRun}},
		{`INSERT INTO extraction_evidence(run_id,measure_ordinal,ordinal,field_name,resource_url,quote,segment_ordinal) VALUES($1,1,1,'kind',$2,'closure',1)`, []any{extractionRun, resourceURL}},
		{`INSERT INTO extraction_evidence(run_id,measure_ordinal,ordinal,field_name,resource_url,quote,segment_ordinal) VALUES($1,1,2,'valid_from',$2,'10 settembre',1),($1,1,3,'valid_from',$2,'10 agosto',1)`, []any{extractionRun, resourceURL}},
		{`INSERT INTO extracted_temporal_candidates(run_id,measure_ordinal,ordinal,field_name,original_expression,conflict_identity) VALUES($1,1,1,'valid_from','10 settembre','retention-conflict'),($1,1,2,'valid_from','10 agosto','retention-conflict')`, []any{extractionRun}},
		{`INSERT INTO extraction_temporal_candidate_evidence(run_id,measure_ordinal,candidate_ordinal,ordinal,evidence_ordinal) VALUES($1,1,1,1,2),($1,1,2,1,3)`, []any{extractionRun}},
		{`INSERT INTO linking_results(run_id,current_extraction_run_id,current_measure_ordinal,status,reason_code,created_at) VALUES($1,$2,1,'no_relation','no_candidates',$3)`, []any{linkingRun, extractionRun, at}},
		{`INSERT INTO linking_evidence(run_id,side,ordinal,field_name,resource_url,quote) VALUES($1,'current',1,'kind',$2,'closure')`, []any{linkingRun, resourceURL}},
		{`INSERT INTO measure_embeddings(run_id,extraction_run_id,measure_ordinal,configuration_version_id,dimensions,vector,returned_model,created_at) VALUES($1,$2,1,'retention-config',1,ARRAY[1.0]::real[],'fixture',$3)`, []any{embeddingRun, extractionRun, at}},
		{`INSERT INTO ocr_resource_results(run_id,document_version_id,resource_url,status,page_count,created_at) VALUES($1,$2,$3,'complete',0,$4)`, []any{ocrRun, versionID, resourceURL, at}},
		{`INSERT INTO processing_run_attempts(run_id,number,started_at,finished_at,duration_ms,outcome,usage_status,other_units,cost_status) VALUES($1,1,$2,$2,0,'succeeded','not_applicable','{}','not_applicable')`, []any{linkingRun, at}},
		{`INSERT INTO interpretation_triggers(document_version_id,evidence_hash,reason,workload,created_at) VALUES($1,$2,'content_changed','evaluation',$3)`, []any{versionID, hash, at}},
	}
	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement.sql, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	if err := New(pool).BindExtraction(ctx, measureID, extractionRun, 1); err != nil {
		t.Fatal(err)
	}
}
