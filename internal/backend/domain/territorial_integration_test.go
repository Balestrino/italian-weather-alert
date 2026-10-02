//go:build integration

package domain

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"github.com/jackc/pgx/v5/pgxpool"
)

func territorialTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	p := domainTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, Migrate, MigrateGeography, MigrateTerritories} {
		if err := m(context.Background(), p); err != nil {
			t.Fatal(err)
		}
	}
	return p
}
func TestTerritorialCatalogAndRevisions(t *testing.T) {
	ctx := context.Background()
	p := territorialTestDB(t)
	s := New(p)
	if err := MigrateTerritories(ctx, p); err != nil {
		t.Fatal(err)
	}
	regions, err := s.Regions(ctx)
	if err != nil || len(regions) != 20 {
		t.Fatalf("catalog: %d %v", len(regions), err)
	}
	for _, r := range regions {
		if r.Enabled || r.Revision != 0 {
			t.Fatalf("new region active: %+v", r)
		}
	}
	r, err := s.Region(ctx, "09")
	if err != nil || r.Name != "Toscana" {
		t.Fatalf("Toscana: %+v %v", r, err)
	}
	if _, err = p.Exec(ctx, `INSERT INTO territorial_regions(code,name,reference) VALUES('09','Duplicate','{}')`); err == nil {
		t.Fatal("duplicate accepted")
	}
	tx, err := p.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rev, err := writeRegion(ctx, tx, "09", 0, RegionConfiguration{Profiles: []string{"toscana-cfr"}}, false, "operator", "configuration")
	if err != nil || rev != 1 {
		t.Fatal(rev, err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	tx, err = p.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = writeRegion(ctx, tx, "09", 0, RegionConfiguration{}, false, "operator", "configuration"); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale change: %v", err)
	}
	tx.Rollback(ctx)
	events, err := s.RegionEvents(ctx, "09")
	if err != nil || len(events) != 1 || events[0].Actor != "operator" || events[0].Revision != 1 {
		t.Fatalf("events: %+v %v", events, err)
	}
	for _, q := range []string{`UPDATE territorial_region_versions SET actor='changed'`, `DELETE FROM territorial_region_events`} {
		if _, err = p.Exec(ctx, q); err == nil {
			t.Fatal("mutable audit", q)
		}
	}
}

func TestTerritorialScopeAndHistory(t *testing.T) {
	ctx := context.Background()
	p := territorialTestDB(t)
	s := New(p)
	must := func(q string, args ...any) {
		t.Helper()
		if _, err := p.Exec(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, v := range []struct{ region, id, istat string }{{"09", "tuscany", "050004"}, {"03", "lombardy", "015146"}} {
		must(`INSERT INTO geography_datasets(id,kind,authority,publisher,official_url,version_label,source_sha256,verified_at,applicability,limitations,usable) VALUES($1,'municipality_registry','ISTAT','ISTAT','https://example.test','v1',repeat('a',64),now(),'verified','[]',true)`, v.id)
		must(`INSERT INTO geography_municipalities VALUES($1,$2,'Municipality','Province',true)`, v.id, v.istat)
		must(`INSERT INTO territorial_dataset_regions VALUES($1,$2,true,'official complete register')`, v.id, v.region)
		must(`INSERT INTO territorial_municipalities VALUES($1,$2,$3)`, v.region, v.id, v.istat)
		must(`INSERT INTO territorial_dataset_selections(region_code,kind,dataset_id,actor,selected_at) VALUES($1,'municipality_registry',$2,'test','2026-01-01')`, v.region, v.id)
		must(`INSERT INTO geography_datasets(id,kind,municipality_dataset_id,authority,publisher,official_url,version_label,source_sha256,verified_at,applicability,limitations,usable) VALUES($1,'zone_mapping',$2,'Region','Region','https://example.test','v1',repeat('b',64),now(),'verified','[]',true)`, v.id+"-zones", v.id)
		must(`INSERT INTO territorial_dataset_regions VALUES($1,$2,true,'official mapping')`, v.id+"-zones", v.region)
		must(`INSERT INTO geography_zone_mappings VALUES($1,$2,$3,1,'A1','Official','whole_municipality','row 1')`, v.id+"-zones", v.id, v.istat)
	}
	if _, err := p.Exec(ctx, `INSERT INTO territorial_municipalities VALUES('09','lombardy','015146')`); err == nil {
		t.Fatal("foreign membership accepted")
	}
	if _, err := p.Exec(ctx, `INSERT INTO territorial_dataset_selections(region_code,kind,dataset_id,actor) VALUES('09','municipality_registry','lombardy','test')`); err == nil {
		t.Fatal("foreign selection accepted")
	}
	if _, err := p.Exec(ctx, `INSERT INTO territorial_dataset_selections(region_code,kind,dataset_id,actor) VALUES('09','zone_mapping','tuscany','test')`); err == nil {
		t.Fatal("wrong kind accepted")
	}
	at := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	for region, want := range map[string]string{"09": "tuscany", "03": "lombardy"} {
		got, err := s.TerritorialDatasetAt(ctx, region, DatasetMunicipalities, at)
		if err != nil || got != want {
			t.Fatal(region, got, err)
		}
	}
	must(`INSERT INTO registry_authorities VALUES('authority','Authority','https://example.test')`)
	must(`INSERT INTO registry_channels VALUES('channel','authority','web','https://example.test',false)`)
	must(`INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES('source','authority','channel','criticality','Toscana')`)
	must(`INSERT INTO territorial_source_associations(source_id,region_code,profile,actor,recorded_at) VALUES('source','09','toscana-cfr','test','2026-01-01')`)
	must(`INSERT INTO territorial_source_associations(source_id,region_code,profile,actor,recorded_at) VALUES('source','03','unsupported','test','2026-03-01')`)
	old, err := s.TerritoryAt(ctx, "source", at)
	if err != nil || old.RegionCode != "09" {
		t.Fatal(old, err)
	}
	latest, err := s.TerritoryAt(ctx, "source", at.AddDate(0, 2, 0))
	if err != nil || latest.RegionCode != "03" {
		t.Fatal(latest, err)
	}
	if _, err := p.Exec(ctx, `INSERT INTO territorial_source_associations(source_id,region_code,municipality_dataset_id,municipality_istat,profile,actor) VALUES('source','09','lombardy','015146','municipal-html','test')`); err == nil {
		t.Fatal("foreign source accepted")
	}
	if _, err := p.Exec(ctx, `DELETE FROM territorial_source_associations`); err == nil {
		t.Fatal("history deleted")
	}
}

func TestToscanaBackfillPreservesLegacyState(t *testing.T) {
	ctx := context.Background()
	p := territorialTestDB(t)
	s := New(p)
	dataset := Dataset{ID: "legacy-toscana", Kind: DatasetMunicipalities, Authority: "ISTAT", Publisher: "ISTAT", OfficialURL: "https://example.test/istat", VersionLabel: "2026", SourceSHA256: strings.Repeat("a", 64), VerifiedAt: time.Now(), Applicability: "verified", Limitations: []string{}, Usable: true}
	if err := s.RegisterMunicipalities(ctx, dataset, readMunicipalities(t)); err != nil {
		t.Fatal(err)
	}
	if err := s.SelectDataset(ctx, DatasetMunicipalities, dataset.ID, "test", time.Now()); err != nil {
		t.Fatal(err)
	}
	must := func(q string) {
		t.Helper()
		if _, err := p.Exec(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	must(`INSERT INTO registry_authorities VALUES('legacy','Legacy','https://example.test')`)
	must(`INSERT INTO registry_channels VALUES('legacy','legacy','web','https://example.test',false)`)
	must(`INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES('regional','legacy','legacy','criticality','Toscana'),('municipal','legacy','legacy','municipal','050004'),('unresolved','legacy','legacy','municipal','999999')`)
	must(`INSERT INTO registry_configurations(source_id,revision,body,actor,created_at) VALUES('regional',1,'{}','original','2026-01-01')`)
	must(`UPDATE registry_sources SET latest_revision=1,active_revision=1,collection_enabled=true,public_enabled=true WHERE id='regional'`)
	var before, after string
	if err := p.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(s) ORDER BY id)::text FROM registry_sources s`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := s.BackfillToscana(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(s) ORDER BY id)::text FROM registry_sources s`).Scan(&after); err != nil || before != after {
		t.Fatal("source changed", err)
	}
	r, err := s.Region(ctx, "09")
	if err != nil || !r.Enabled || r.Revision != 1 || r.Configuration.MunicipalityDataset != dataset.ID {
		t.Fatalf("region %+v %v", r, err)
	}
	var count int
	if err = p.QueryRow(ctx, `SELECT count(*) FROM territorial_dataset_selections`).Scan(&count); err != nil || count != 1 {
		t.Fatal(count, err)
	}
	var selected time.Time
	if err = p.QueryRow(ctx, `SELECT selected_at FROM geography_dataset_selections`).Scan(&selected); err != nil {
		t.Fatal(err)
	}
	got, err := s.TerritorialDatasetAt(ctx, "09", DatasetMunicipalities, selected)
	if err != nil || got != dataset.ID {
		t.Fatal(got, err)
	}
	association, err := s.TerritoryAt(ctx, "regional", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	if err != nil || association.RegionCode != "09" {
		t.Fatal(association, err)
	}
	report, err := s.TerritoryMigrationReport(ctx)
	if err != nil || len(report.UnassociatedSources) != 1 || report.UnassociatedSources[0] != "unresolved" {
		t.Fatal(report, err)
	}
}

func TestTerritorialImportAtomicityAndRetirement(t *testing.T) {
	ctx := context.Background()
	p := territorialTestDB(t)
	s := New(p)
	meta, body := importFixture()
	preview, err := PreviewMunicipalities(body, meta)
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 0, "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 0, "test"); !errors.Is(err, ErrConflict) {
		t.Fatal("stale adoption", err)
	}
	meta.ExpectedCount = 3
	if _, err = s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 1, "test"); !errors.Is(err, ErrInvalid) {
		t.Fatal("partial adopted", err)
	}
	meta.ExpectedCount = 1
	meta.Version = "retirement"
	body = []byte("istat,comune,codice_regione,sigla\n048017,Firenze,09,FI\n")
	preview, err = PreviewMunicipalities(body, meta)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 1, "test")
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err = p.QueryRow(ctx, `SELECT count(*) FROM territorial_municipalities WHERE dataset_id=$1 AND istat='050004'`, id).Scan(&count); err != nil || count != 1 {
		t.Fatal("history lost", err)
	}
	if err = p.QueryRow(ctx, `SELECT count(*) FROM territorial_municipalities WHERE dataset_id=$1 AND istat='050004'`, second).Scan(&count); err != nil || count != 0 {
		t.Fatal("retired still current", err)
	}
	// A constraint failure after entering the transaction must not change the current revision or selection.
	if _, err = s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 2, "test"); err == nil {
		t.Fatal("duplicate dataset adopted")
	}
	region, err := s.Region(ctx, "09")
	if err != nil || region.Revision != 2 || region.Configuration.MunicipalityDataset != second {
		t.Fatal(region, err)
	}
	if err = p.QueryRow(ctx, `SELECT count(*) FROM territorial_dataset_selections WHERE region_code='09'`).Scan(&count); err != nil || count != 2 {
		t.Fatal(count, err)
	}
}

func TestRegionConfigurationLifecycle(t *testing.T) {
	ctx := context.Background()
	p := territorialTestDB(t)
	s := New(p)
	if _, err := s.SetRegionEnabled(ctx, "03", 0, true, "operator"); !errors.Is(err, ErrRegionIncomplete) {
		t.Fatal("incomplete enabled", err)
	}
	meta, body := importFixture()
	preview, _ := PreviewMunicipalities(body, meta)
	dataset, err := s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 0, "importer")
	if err != nil {
		t.Fatal(err)
	}
	rev, err := s.ConfigureRegion(ctx, "09", 1, RegionConfiguration{MunicipalityDataset: dataset, Profiles: []string{"toscana-cfr", "municipal-html"}}, "configurator")
	if err != nil || rev != 2 {
		t.Fatal(rev, err)
	}
	if _, err = s.ConfigureRegion(ctx, "03", 0, RegionConfiguration{MunicipalityDataset: dataset}, "operator"); !errors.Is(err, ErrInvalid) {
		t.Fatal("foreign dataset allowed", err)
	}
	if _, err = s.SetRegionEnabled(ctx, "09", 2, true, "operator"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SetRegionEnabled(ctx, "09", 2, false, "operator"); !errors.Is(err, ErrConflict) {
		t.Fatal("stale mutation accepted", err)
	}
	if _, err = s.SetRegionEnabled(ctx, "09", 3, false, "operator"); err != nil {
		t.Fatal(err)
	}
	region, err := s.Region(ctx, "09")
	if err != nil || region.Enabled || region.Configuration.MunicipalityDataset != dataset || region.Revision != 4 {
		t.Fatal(region, err)
	}
	var globalSelections int
	if err = p.QueryRow(ctx, `SELECT count(*) FROM geography_dataset_selections`).Scan(&globalSelections); err != nil || globalSelections != 0 {
		t.Fatal("global selection modified", err)
	}
	events, err := s.RegionEvents(ctx, "09")
	if err != nil || len(events) != 4 || events[0].Kind != "disabled" || events[1].Kind != "enabled" {
		t.Fatal(events, err)
	}
}

func TestSourceTerritoryProfiles(t *testing.T) {
	ctx := context.Background()
	p := territorialTestDB(t)
	s := New(p)
	meta, _ := importFixture()
	meta.Region = "03"
	meta.ExpectedCount = 1
	body := []byte("istat,comune,codice_regione,sigla\n015146,Milano,03,MI\n")
	preview, _ := PreviewMunicipalities(body, meta)
	id, err := s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 0, "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ConfigureRegion(ctx, "03", 1, RegionConfiguration{MunicipalityDataset: id, Profiles: []string{"municipal-html"}}, "test"); err != nil {
		t.Fatal(err)
	}
	reg := registry.New(p)
	if err = reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "Authority", OfficialURL: "https://example.test"}); err != nil {
		t.Fatal(err)
	}
	if err = reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "web", URL: "https://example.test"}); err != nil {
		t.Fatal(err)
	}
	cfg := registry.Configuration{URL: "https://example.test", Sections: []string{"https://example.test"}, AccessMethod: "html", Attribution: "test"}
	if err = reg.CreateSource(ctx, registry.Source{ID: "milan", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "015146"}, cfg, "test"); err != nil {
		t.Fatal(err)
	}
	if err = territory.CheckProfile(ctx, p, "milan", ""); err != nil {
		t.Fatal(err)
	}
	if err = reg.CreateSource(ctx, registry.Source{ID: "unknown", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "999999"}, cfg, "test"); !errors.Is(err, territory.ErrAssociation) {
		t.Fatal("unresolved accepted", err)
	}
	if err = reg.CreateSource(ctx, registry.Source{ID: "regional-new", AuthorityID: "a", ChannelID: "c", ProductID: "criticality", Territory: "03"}, cfg, "test"); err != nil {
		t.Fatal(err)
	}
	if err = reg.EnableCollection(ctx, "regional-new", 1, "test"); !errors.Is(err, territory.ErrUnsupported) {
		t.Fatal("foreign regional profile allowed", err)
	}
	cfg.RegionalProduct = &registry.RegionalProductContract{Kind: "monitoring", ContentMarkers: []string{"fixture"}}
	if _, err = reg.AppendConfiguration(ctx, "milan", 1, cfg, "test"); err != nil {
		t.Fatal(err)
	}
	if err = reg.CheckTerritorialExecution(ctx, "milan", 2); !errors.Is(err, territory.ErrUnsupported) {
		t.Fatal("municipal source executed regional adapter", err)
	}
	if err = reg.EnableCollection(ctx, "milan", 2, "test"); !errors.Is(err, territory.ErrUnsupported) {
		t.Fatal("municipal source activated regional adapter", err)
	}
	if err = territory.CheckProfile(ctx, p, "milan", "toscana-cfr"); !errors.Is(err, territory.ErrUnsupported) {
		t.Fatal("wrong profile allowed", err)
	}
}

func TestDisabledRegionGatesClaimsAndManualExecution(t *testing.T) {
	ctx := context.Background()
	p := territorialTestDB(t)
	s := New(p)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{jobs.Migrate, acquisition.Migrate} {
		if err := m(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	for _, region := range []string{"09", "03"} {
		meta, _ := importFixture()
		meta.Region = region
		meta.ExpectedCount = 1
		istat := "050004"
		if region == "03" {
			istat = "015146"
		}
		body := []byte("istat,comune,codice_regione,sigla\n" + istat + ",Municipality," + region + ",PR\n")
		preview, _ := PreviewMunicipalities(body, meta)
		id, err := s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 0, "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.ConfigureRegion(ctx, region, 1, RegionConfiguration{MunicipalityDataset: id, Profiles: []string{"municipal-html"}}, "test"); err != nil {
			t.Fatal(err)
		}
		if _, err = s.SetRegionEnabled(ctx, region, 2, true, "test"); err != nil {
			t.Fatal(err)
		}
		if _, err = s.SetMunicipalityEnabled(ctx, region, istat, 0, true, "test"); err != nil {
			t.Fatal(err)
		}
	}
	reg := registry.New(p)
	if err := reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "a", OfficialURL: "https://example.test"}); err != nil {
		t.Fatal(err)
	}
	if err := reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "web", URL: "https://example.test"}); err != nil {
		t.Fatal(err)
	}
	for _, istat := range []string{"050004", "015146"} {
		cfg := registry.Configuration{URL: "https://example.test", Sections: []string{"https://example.test"}, AccessMethod: "html", Attribution: "fixture"}
		if err := reg.CreateSource(ctx, registry.Source{ID: istat, AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: istat}, cfg, "test"); err != nil {
			t.Fatal(err)
		}
		if _, err := p.Exec(ctx, `UPDATE registry_sources SET collection_enabled=true,active_revision=1 WHERE id=$1`, istat); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := p.Exec(ctx, `INSERT INTO retained_documents(source_id,official_url) VALUES('050004','https://example.test/document'); INSERT INTO retained_acquisitions(id,document_id,source_id,configuration,request_hash,content_hash,payload) SELECT 'staged',id,'050004',1,'fixture','fixture','{}' FROM retained_documents WHERE source_id='050004'`); err != nil {
		t.Fatal(err)
	}
	checkStaged := func(want bool) {
		t.Helper()
		var allowed bool
		if err := p.QueryRow(ctx, `SELECT territorial_job_allowed('{"acquisition_id":"staged"}')`).Scan(&allowed); err != nil || allowed != want {
			t.Fatalf("staged acquisition gate=%v want=%v err=%v", allowed, want, err)
		}
	}
	checkStaged(true)
	queue := jobs.New(p)
	now := time.Now().UTC()
	enqueue := func(key, source string) {
		t.Helper()
		_, err := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "test", Kind: "test", IdempotencyKey: key, Payload: json.RawMessage(`{"source_id":"` + source + `"}`), MaxAttempts: 1, RetryBase: time.Second, AvailableAt: now})
		if err != nil {
			t.Fatal(err)
		}
	}
	enqueue("running", "050004")
	running, err := queue.Claim(ctx, "test", "worker", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	enqueue("pending-disabled", "050004")
	enqueue("pending-enabled", "015146")
	if _, err = s.SetRegionEnabled(ctx, "09", 3, false, "test"); err != nil {
		t.Fatal(err)
	}
	checkStaged(false)
	if err = queue.Complete(ctx, running, jobs.Result{Payload: json.RawMessage(`{}`)}, time.Now()); err != nil {
		t.Fatal("running job cannot finish", err)
	}
	claim, err := queue.Claim(ctx, "test", "worker", time.Now(), time.Minute)
	if err != nil || claim.IdempotencyKey != "pending-enabled" {
		t.Fatal(claim, err)
	}
	if _, err = queue.Claim(ctx, "test", "worker", time.Now(), time.Minute); !errors.Is(err, jobs.ErrNoJob) {
		t.Fatal("disabled claimed", err)
	}
	if err = reg.CheckTerritorialExecution(ctx, "050004", 1); !errors.Is(err, territory.ErrDisabled) {
		t.Fatal("manual execution allowed", err)
	}
	schedule := acquisition.NewScheduleStore(p)
	if err = schedule.SyncEnabled(ctx, now); err != nil {
		t.Fatal(err)
	}
	due, err := schedule.ClaimDue(ctx, "worker", time.Now(), time.Minute)
	if err != nil || due.SourceID != "015146" {
		t.Fatal(due, err)
	}
	if _, err = schedule.ClaimDue(ctx, "worker", time.Now(), time.Minute); !errors.Is(err, acquisition.ErrNoDueCheck) {
		t.Fatal("disabled acquisition claimed", err)
	}
	if _, err = s.SetRegionEnabled(ctx, "09", 4, true, "test"); err != nil {
		t.Fatal(err)
	}
	checkStaged(true)
	claim, err = queue.Claim(ctx, "test", "worker", time.Now(), time.Minute)
	if err != nil || claim.IdempotencyKey != "pending-disabled" {
		t.Fatal(claim, err)
	}
	var count int
	if err = p.QueryRow(ctx, `SELECT count(*) FROM processing_jobs`).Scan(&count); err != nil || count != 3 {
		t.Fatal("reenable created work", count, err)
	}
}
