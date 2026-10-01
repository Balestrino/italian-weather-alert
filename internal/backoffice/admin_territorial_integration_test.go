//go:build integration

package backoffice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publicquery"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func territorialAdminDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	p := adminTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, acquisition.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, domain.Migrate, domain.MigrateGeography, domain.MigrateTerritories, domain.MigrateTemporal, domain.MigrateQuality, publicquery.Migrate} {
		if err := m(context.Background(), p); err != nil {
			t.Fatal(err)
		}
	}
	return p
}
func adoptTestMunicipalities(t *testing.T, p *pgxpool.Pool, region string, count int) string {
	t.Helper()
	body := "istat,comune,codice_regione,sigla\n"
	for i := 0; i < count; i++ {
		body += fmt.Sprintf("9%s%03d,Comune %03d,%s,PR\n", region, i, i, region)
	}
	meta := domain.MunicipalityImport{Region: region, OfficialURL: "https://example.test/official", Version: "fixture", VerifiedAt: time.Now(), ExpectedCount: count, Complete: true, CompletenessEvidence: "Complete test fixture"}
	preview, err := domain.PreviewMunicipalities([]byte(body), meta)
	if err != nil {
		t.Fatal(err)
	}
	id, err := domain.New(p).AdoptMunicipalities(context.Background(), []byte(body), meta, preview.SHA256, 0, "test")
	if err != nil {
		t.Fatal(err)
	}
	return id
}
func TestTerritorialMunicipalityReaders(t *testing.T) {
	ctx := context.Background()
	p := territorialAdminDB(t)
	adoptTestMunicipalities(t, p, "09", 125)
	adoptTestMunicipalities(t, p, "03", 2)
	s := operations.New(p)
	page, err := s.Municipalities(ctx, "09", operations.MunicipalityFilter{Limit: 50}, time.Now())
	if err != nil || page.Total != 125 || len(page.Items) != 50 || page.Next == "" {
		t.Fatal(page, err)
	}
	next, err := s.Municipalities(ctx, "09", operations.MunicipalityFilter{Limit: 50, After: page.Next}, time.Now())
	if err != nil || next.Total != 125 || len(next.Items) != 50 || next.Items[0].ISTAT == page.Items[0].ISTAT {
		t.Fatal(next, err)
	}
	filtered, err := s.Municipalities(ctx, "09", operations.MunicipalityFilter{Search: "Comune 12", Province: "PR", Coverage: "none"}, time.Now())
	if err != nil || filtered.Total != 5 {
		t.Fatal(filtered, err)
	}
	if _, err = s.Municipality(ctx, "03", "909000"); !errors.Is(err, domain.ErrTerritoryNotFound) {
		t.Fatal("foreign municipality shown", err)
	}
	detail, err := s.Municipality(ctx, "09", "909000")
	if err != nil || detail.Historical || detail.Sources != 0 {
		t.Fatal(detail, err)
	}
	reg := registry.New(p)
	for _, err := range []error{reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "a", OfficialURL: "https://example.test"}), reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "web", URL: "https://example.test"}), reg.CreateSource(ctx, registry.Source{ID: "s", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "909000"}, registry.Configuration{URL: "https://example.test", Sections: []string{"https://example.test"}, AccessMethod: "html", Attribution: "test"}, "test")} {
		if err != nil {
			t.Fatal(err)
		}
	}
	configured, err := s.Municipalities(ctx, "09", operations.MunicipalityFilter{Coverage: "configured"}, time.Now())
	if err != nil || configured.Total != 1 || configured.Items[0].ISTAT != "909000" {
		t.Fatal(configured, err)
	}
	sources, err := s.TerritorialSources(ctx, "09", "909000")
	if err != nil || len(sources) != 1 || sources[0].EffectiveCheckSeconds != 600 {
		t.Fatal(sources, err)
	}
	overview, err := s.TerritorialOverview(ctx, time.Now())
	if err != nil || len(overview.Regions) != 20 {
		t.Fatal(overview, err)
	}
	for _, r := range overview.Regions {
		if r.Code == "09" && (r.Municipalities == nil || *r.Municipalities != 125 || r.Sources == nil || *r.Sources != 1) {
			t.Fatal(r)
		}
	}
	// The query count is independent of the requested page size.
	counter := &territorialQueryCounter{}
	config := p.Config()
	config.ConnConfig.Tracer = counter
	measured, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer measured.Close()
	if _, err = operations.New(measured).Municipalities(ctx, "09", operations.MunicipalityFilter{Limit: 1}, time.Now()); err != nil {
		t.Fatal(err)
	}
	small := counter.n
	counter.n = 0
	if _, err = operations.New(measured).Municipalities(ctx, "09", operations.MunicipalityFilter{Limit: 100}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if counter.n != small || counter.n > 6 {
		t.Fatal("page incurs per-municipality queries", small, counter.n)
	}
	// One missing operational dependency must preserve the complete municipality summary.
	if _, err = p.Exec(ctx, `ALTER TABLE acquisition_source_status RENAME TO unavailable_status`); err != nil {
		t.Fatal(err)
	}
	overview, err = s.TerritorialOverview(ctx, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range overview.Regions {
		if r.Code == "09" && (r.Municipalities == nil || *r.Municipalities != 125 || r.Sources != nil) {
			t.Fatal("partial failure hidden", r)
		}
	}
}

type territorialQueryCounter struct{ n int }

func (c *territorialQueryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	c.n++
	return ctx
}
func (c *territorialQueryCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func TestTerritorialPrivateResultsAndMappings(t *testing.T) {
	ctx := context.Background()
	p := territorialAdminDB(t)
	municipalities := adoptTestMunicipalities(t, p, "09", 2)
	adoptTestMunicipalities(t, p, "03", 2)
	geo := domain.New(p)
	reg := registry.New(p)
	ops := operations.New(p)
	for _, err := range []error{reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "Authority", OfficialURL: "https://example.test"}), reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "web", URL: "https://example.test"})} {
		if err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	zones := domain.Dataset{ID: "partial-zones", Kind: domain.DatasetZones, MunicipalityDatasetID: municipalities, Authority: "Region", Publisher: "Region", OfficialURL: "https://example.test/zones", VersionLabel: "test", SourceSHA256: strings.Repeat("a", 64), VerifiedAt: now, Applicability: "verified", Limitations: []string{}, Usable: true}
	if err := geo.RegisterZones(ctx, zones, []domain.ZoneMapping{{MunicipalityISTAT: "909000", Zone: "A1", SourceName: "Official", TerritorialScope: "partial_municipality", EvidenceLocator: "one"}, {MunicipalityISTAT: "909000", Zone: "A2", SourceName: "Official", TerritorialScope: "partial_municipality", EvidenceLocator: "two"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Exec(ctx, `INSERT INTO territorial_dataset_regions VALUES('partial-zones','09',true,'official mapping')`); err != nil {
		t.Fatal(err)
	}
	if _, err := geo.ConfigureRegion(ctx, "09", 1, domain.RegionConfiguration{MunicipalityDataset: municipalities, ZoneDataset: zones.ID, Profiles: []string{"municipal-html", "toscana-cfr"}}, "test"); err != nil {
		t.Fatal(err)
	}
	retained := documents.New(p, adminObjects{})
	for _, region := range []string{"09", "03"} {
		cfg := registry.Configuration{URL: "https://example.test", Sections: []string{"https://example.test"}, AccessMethod: "html", Attribution: "test", Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true, Evidence: &registry.Evidence{URL: "https://example.test", Locator: "fixture", ObservedAt: now}}}
		source := "regional-" + region
		if err := reg.CreateSource(ctx, registry.Source{ID: source, AuthorityID: "a", ChannelID: "c", ProductID: "criticality", Territory: region}, cfg, "test"); err != nil {
			t.Fatal(err)
		}
		version, err := retained.Retain(ctx, documents.Acquisition{ID: source, SourceID: source, Configuration: 1, URL: "https://example.test/" + source, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: "https://example.test/" + source, Role: "original", Required: true, SourceID: source, Configuration: 1, MediaType: "text/html", Bytes: []byte("Regional bulletin " + region)}}})
		if err != nil {
			t.Fatal(err)
		}
		yellow := "yellow"
		record := domain.RegionalRecord{ID: source, SourceID: source, DocumentVersionID: version.ID, Product: "criticality", OriginatingAuthorityID: "a", Facts: []domain.RegionalFact{{Ordinal: 1, Risk: domain.RiskWind, OfficialRiskLabel: "Vento", Zone: "A1", Level: &yellow}}}
		if err = geo.PutRegional(ctx, record); err != nil {
			t.Fatal(err)
		}
	}
	at := time.Now().Add(time.Second)
	results, err := ops.TerritorialResults(ctx, "09", "909000", at)
	if err != nil || results.DocumentsUnavailable || results.Alerts.FactsUnavailable || len(results.Documents) != 1 || results.Documents[0].Public || results.Documents[0].Classification != "not_processed" || len(results.Alerts.UndatedRegional) != 1 || results.MappingUnavailable || len(results.Zones) != 2 {
		t.Fatalf("private results: %+v %v", results, err)
	}
	if !strings.HasPrefix(results.Alerts.UndatedRegional[0].ID, "regional-09:") {
		t.Fatal("cross-region fact")
	}
	for _, z := range results.Zones {
		if z.TerritorialScope != "partial_municipality" {
			t.Fatal("partial lost")
		}
	}
	other, err := ops.TerritorialResults(ctx, "03", "", at)
	if err != nil || len(other.Documents) != 1 || len(other.Alerts.UndatedRegional) != 1 || !strings.HasPrefix(other.Alerts.UndatedRegional[0].ID, "regional-03:") {
		t.Fatalf("second region: %+v %v", other, err)
	}
	missing, err := ops.TerritorialResults(ctx, "09", "909001", at)
	if err != nil || !missing.MappingUnavailable || len(missing.Alerts.UndatedRegional) != 0 {
		t.Fatal("unmapped municipality received warning", missing, err)
	}
	if _, err = ops.TerritorialResults(ctx, "03", "909000", at); !errors.Is(err, domain.ErrTerritoryNotFound) {
		t.Fatal("foreign detail accepted", err)
	}
	if len(results.Alerts.Days[0].Regional) != 0 || len(results.Alerts.Days[1].Regional) != 0 {
		t.Fatal("uncertain validity assigned")
	}
}

func TestTerritorialHistoryScopeAndPagination(t *testing.T) {
	ctx := context.Background()
	p := territorialAdminDB(t)
	adoptTestMunicipalities(t, p, "09", 2)
	adoptTestMunicipalities(t, p, "03", 2)
	reg := registry.New(p)
	ops := operations.New(p)
	for _, err := range []error{reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "a", OfficialURL: "https://example.test"}), reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "web", URL: "https://example.test"}), reg.CreateSource(ctx, registry.Source{ID: "s", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "909000"}, registry.Configuration{URL: "https://example.test", Sections: []string{"https://example.test"}, AccessMethod: "html", Attribution: "fixture"}, "test")} {
		if err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	eventAt := now.Add(time.Second)
	moveAt := now.Add(2 * time.Second)
	later := now.Add(3 * time.Second)
	if _, err := p.Exec(ctx, `INSERT INTO registry_events(source_id,revision,kind,actor,evidence,created_at) VALUES('s',1,'preview','first','{}',$1),('s',1,'preview','second','{}',$1)`, eventAt); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Exec(ctx, `INSERT INTO territorial_source_associations(source_id,region_code,profile,actor,recorded_at) VALUES('s','03','unsupported','mover',$1)`, moveAt); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Exec(ctx, `INSERT INTO registry_events(source_id,revision,kind,actor,evidence,created_at) VALUES('s',1,'preview','later','{}',$1)`, later); err != nil {
		t.Fatal(err)
	}
	through := now.Add(4 * time.Second)
	f := operations.TerritorialHistoryFilter{Source: "s", Kind: "source", From: &now, Through: &through, Limit: 1}
	page, err := ops.TerritorialHistory(ctx, "09", "909000", f, through)
	if err != nil || len(page.Entries) != 1 || page.Next == "" || page.RetentionKnown || len(page.Limitations) < 2 {
		t.Fatal(page, err)
	}
	f.After = page.Next
	next, err := ops.TerritorialHistory(ctx, "09", "909000", f, through)
	if err != nil || len(next.Entries) != 1 || next.Entries[0].ID == page.Entries[0].ID || next.Next != "" {
		t.Fatal(next, err)
	}
	f.After = ""
	foreign, err := ops.TerritorialHistory(ctx, "03", "", f, through)
	if err != nil || len(foreign.Entries) != 1 || foreign.Entries[0].Actor != "later" {
		t.Fatal(foreign, err)
	}
	if _, err = domain.New(p).SetRegionEnabled(ctx, "09", 1, false, "operator"); err != nil {
		t.Fatal(err)
	}
	disabled, err := ops.TerritorialHistory(ctx, "09", "909000", f, through)
	if err != nil || len(disabled.Entries) != 1 {
		t.Fatal("disabled history missing", err)
	}
	// Retiring a municipality preserves its original source events and detail link.
	body := []byte("istat,comune,codice_regione,sigla\n909001,Remaining,09,PR\n")
	meta := domain.MunicipalityImport{Region: "09", OfficialURL: "https://example.test/new", Version: "retired", VerifiedAt: time.Now(), ExpectedCount: 1, Complete: true, CompletenessEvidence: "Fixture retirement"}
	preview, _ := domain.PreviewMunicipalities(body, meta)
	if _, err = domain.New(p).AdoptMunicipalities(ctx, body, meta, preview.SHA256, 2, "operator"); err != nil {
		t.Fatal(err)
	}
	m, err := ops.Municipality(ctx, "09", "909000")
	if err != nil || !m.Historical {
		t.Fatal(m, err)
	}
	retired, err := ops.TerritorialHistory(ctx, "09", "909000", f, through)
	if err != nil || len(retired.Entries) != 1 {
		t.Fatal(retired, err)
	}
}

func TestTerritorialHistorySeparatesListingSnapshots(t *testing.T) {
	ctx := context.Background()
	p := territorialAdminDB(t)
	adoptTestMunicipalities(t, p, "09", 1)
	reg := registry.New(p)
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "a", OfficialURL: "https://example.test"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "web", URL: "https://example.test"}),
		reg.CreateSource(ctx, registry.Source{ID: "s", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "909000"}, registry.Configuration{URL: "https://example.test", Sections: []string{"https://example.test/notizie"}, AccessMethod: "crawl4ai", Discovery: registry.Discovery{PaginationParameter: "page"}, Attribution: "fixture"}, "test"),
		reg.CreateSource(ctx, registry.Source{ID: "direct", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "909000"}, registry.Configuration{URL: "https://example.test", Sections: []string{"https://example.test/direct"}, AccessMethod: "html", Attribution: "fixture"}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	from := time.Now().UTC().Add(time.Second)
	for i, item := range []struct{ source, raw string }{
		{"s", "https://example.test/notizie"},
		{"s", "https://example.test/notizie?page=1"},
		{"s", "https://example.test/novita/avviso"},
		{"direct", "https://example.test/direct"},
	} {
		var documentID, versionID int64
		hash := strings.Repeat(string(rune('a'+i)), 64)
		if err := p.QueryRow(ctx, "INSERT INTO retained_documents(source_id,official_url) VALUES($1,$2) RETURNING id", item.source, item.raw).Scan(&documentID); err != nil {
			t.Fatal(err)
		}
		if _, err := p.Exec(ctx, "INSERT INTO retained_objects(hash,object_key,byte_size) VALUES($1,$2,1)", hash, "fixture/"+hash); err != nil {
			t.Fatal(err)
		}
		if err := p.QueryRow(ctx, "INSERT INTO retained_versions(document_id,content_hash,first_acquired_at,complete,metadata) VALUES($1,$2,$3,true,'{}') RETURNING id", documentID, hash, from.Add(time.Duration(i)*time.Second)).Scan(&versionID); err != nil {
			t.Fatal(err)
		}
		if _, err := p.Exec(ctx, "INSERT INTO retained_resources(version_id,url,role,required,source_id,configuration,media_type,object_hash,missing) VALUES($1,$2,'original',true,$3,1,'text/html',$4,'')", versionID, item.raw, item.source, hash); err != nil {
			t.Fatal(err)
		}
	}
	reader := operations.New(p)
	for kind, want := range map[string]int{"": 1, "document": 1, "listing": 2, "all": 3} {
		page, err := reader.TerritorialHistory(ctx, "09", "909000", operations.TerritorialHistoryFilter{Source: "s", Kind: kind, From: &from}, from.Add(10*time.Second))
		if err != nil || len(page.Entries) != want {
			t.Fatalf("kind %q: %d entries, want %d: %v", kind, len(page.Entries), want, err)
		}
	}
	direct, err := reader.TerritorialHistory(ctx, "09", "909000", operations.TerritorialHistoryFilter{Source: "direct", Kind: "document", From: &from}, from.Add(10*time.Second))
	if err != nil || len(direct.Entries) != 1 {
		t.Fatal("direct source was classified as a listing", direct, err)
	}
	listingFilter := operations.TerritorialHistoryFilter{Source: "s", Kind: "listing", From: &from, Limit: 1}
	first, err := reader.TerritorialHistory(ctx, "09", "909000", listingFilter, from.Add(10*time.Second))
	if err != nil || len(first.Entries) != 1 || first.Next == "" {
		t.Fatal("listing history did not paginate", first, err)
	}
	listingFilter.After = first.Next
	second, err := reader.TerritorialHistory(ctx, "09", "909000", listingFilter, from.Add(10*time.Second))
	if err != nil || len(second.Entries) != 1 || second.Entries[0].ID == first.Entries[0].ID || second.Next != "" {
		t.Fatal("listing history pagination did not advance", second, err)
	}
}

func TestRegionalDetailRoutes(t *testing.T) {
	p := territorialAdminDB(t)
	adoptTestMunicipalities(t, p, "09", 125)
	h := HandlerWithAdministration(nil, AdminRuntime{Territories: operations.New(p), TerritoryConfig: domain.New(p)})
	for _, path := range []string{"/admin/regions/09", "/admin/regions/09?tab=municipalities", "/admin/regions/09?tab=configuration", "/admin/regions/09?tab=history", "/admin/regions/03?tab=municipalities"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1"+path, nil))
		if w.Code != 200 {
			t.Fatal(path, w.Code, w.Body)
		}
		if strings.Contains(path, "09?tab=municipalities") && (!strings.Contains(w.Body.String(), "125 comuni") || !strings.Contains(w.Body.String(), "/municipalities/909000") || !strings.Contains(w.Body.String(), "Comuni successivi")) {
			t.Fatal(w.Body)
		}
		if strings.Contains(path, "03?") && !strings.Contains(w.Body.String(), "anagrafica dei comuni non è ancora") {
			t.Fatal(w.Body)
		}
	}
	for path, status := range map[string]int{"/admin/regions/99": 404, "/admin/regions/09?tab=invalid": 400, "/admin/regions/09?tab=history&from=bad": 400} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1"+path, nil))
		if w.Code != status {
			t.Fatal(path, w.Code, w.Body)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1/admin/regions/09?tab=municipalities&province=PR&coverage=none", nil))
	if !strings.Contains(w.Body.String(), "coverage=none") || !strings.Contains(w.Body.String(), "province=PR") {
		t.Fatal("pagination lost filters", w.Body)
	}
}

func TestMunicipalityWorkspaceRoutes(t *testing.T) {
	ctx := context.Background()
	p := territorialAdminDB(t)
	adoptTestMunicipalities(t, p, "09", 2)
	adoptTestMunicipalities(t, p, "03", 2)
	reg := registry.New(p)
	for _, err := range []error{reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "a", OfficialURL: "https://example.test"}), reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "web", URL: "https://example.test"})} {
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"one", "two"} {
		if err := reg.CreateSource(ctx, registry.Source{ID: id, AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "909000"}, registry.Configuration{URL: "https://example.test", Sections: []string{"https://example.test"}, AccessMethod: "html", Attribution: "test"}, "operator"); err != nil {
			t.Fatal(err)
		}
	}
	h := HandlerWithAdministration(nil, AdminRuntime{Territories: operations.New(p), TerritoryConfig: domain.New(p), Registry: reg})
	for _, tab := range []string{"results", "configuration", "history"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1/admin/regions/09/municipalities/909000?tab="+tab, nil))
		if w.Code != 200 || !strings.Contains(w.Body.String(), "Comune 000") || !strings.Contains(w.Body.String(), ">Dati</a>") {
			t.Fatal(tab, w.Code, w.Body)
		}
		if tab == "configuration" && (!strings.Contains(w.Body.String(), "/admin/sources/one") || !strings.Contains(w.Body.String(), "/admin/sources/two") || !strings.Contains(w.Body.String(), "municipality=909000")) {
			t.Fatal("local sources missing", w.Body)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1/admin/regions/09/municipalities/909001", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Nessuna fonte locale configurata") || !strings.Contains(w.Body.String(), "Mappatura delle zone non disponibile") {
		t.Fatal(w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1/admin/regions/03/municipalities/909000", nil))
	if w.Code != 404 || strings.Contains(w.Body.String(), "Comune 000") {
		t.Fatal("foreign municipality", w.Code, w.Body)
	}
}

func TestTerritoryOnboardingAndContextualActions(t *testing.T) {
	ctx := context.Background()
	p := territorialAdminDB(t)
	reg := registry.New(p)
	geo := domain.New(p)
	h := HandlerWithAdministration(nil, AdminRuntime{Territories: operations.New(p), TerritoryConfig: geo, Registry: reg})
	request := func(method, path string, form url.Values) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "http://127.0.0.1"+path, strings.NewReader(form.Encode()))
		r.Header.Set("Accept", "text/html")
		if method == "POST" {
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	body := "istat,comune,codice_regione,sigla\n015146,Milano,03,MI\n"
	form := url.Values{"actor": {"operator"}, "expected_revision": {"0"}, "official_url": {"https://example.test/istat"}, "version": {"fixture"}, "verified_at": {"2026-09-15"}, "expected_count": {"1"}, "complete": {"true"}, "completeness_evidence": {"Complete attributed fixture"}, "csv": {body}}
	if w := request("GET", "/admin/regions/03/setup", nil); w.Code != 200 {
		t.Fatal(w.Code, w.Body)
	}
	if w := request("POST", "/admin/regions/03/import/preview", form); w.Code != 200 || !strings.Contains(w.Body.String(), "Adotta anagrafica verificata") {
		t.Fatal(w.Code, w.Body)
	}
	region, err := geo.Region(ctx, "03")
	if err != nil || region.Revision != 0 {
		t.Fatal("preview mutated", region, err)
	}
	preview, err := domain.PreviewMunicipalities([]byte(body), domain.MunicipalityImport{Region: "03", OfficialURL: form.Get("official_url"), Version: "fixture", VerifiedAt: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC), ExpectedCount: 1, Complete: true, CompletenessEvidence: form.Get("completeness_evidence")})
	if err != nil {
		t.Fatal(err)
	}
	form.Set("sha256", preview.SHA256)
	if w := request("POST", "/admin/regions/03/import/adopt", form); w.Code != 303 {
		t.Fatal(w.Code, w.Body)
	}
	if w := request("POST", "/admin/regions/03/import/adopt", form); w.Code != 409 {
		t.Fatal("stale adoption", w.Code, w.Body)
	}
	region, _ = geo.Region(ctx, "03")
	config := url.Values{"actor": {"operator"}, "expected_revision": {"1"}, "municipality_dataset": {region.Configuration.MunicipalityDataset}, "zone_dataset": {""}, "postal_dataset": {""}, "municipal_profile": {"true"}}
	if w := request("POST", "/admin/regions/03/configuration", config); w.Code != 303 {
		t.Fatal(w.Code, w.Body)
	}
	enabled := url.Values{"actor": {"operator"}, "expected_revision": {"2"}, "enabled": {"true"}}
	if w := request("POST", "/admin/regions/03/enabled", enabled); w.Code != 303 {
		t.Fatal(w.Code, w.Body)
	}
	source := url.Values{"actor": {"operator"}, "source_id": {"milan-news"}, "authority_name": {"Comune di Milano"}, "authority_url": {"https://example.test"}, "url": {"https://example.test/news"}, "sections": {"https://example.test/news"}, "attribution": {"Comune di Milano"}, "product": {"municipal"}}
	if w := request("POST", "/admin/regions/03/sources/new?municipality=015146", source); w.Code != 303 {
		t.Fatal(w.Code, w.Body)
	}
	back := "/admin/regions/03/municipalities/015146?tab=configuration"
	w := request("GET", "/admin/sources/milan-news", nil)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Torna al territorio") || !strings.Contains(w.Body.String(), "return_to=") {
		t.Fatal(w.Code, w.Body)
	}
	intervals := url.Values{"actor": {"operator"}, "expected_revision": {"0"}, "check_seconds": {"900"}, "delay_seconds": {"1800"}}
	w = request("POST", "/admin/sources/milan-news/intervals?return_to="+url.QueryEscape(back), intervals)
	if w.Code != 200 || !strings.Contains(w.Body.String(), back) {
		t.Fatal("context lost", w.Code, w.Body)
	}
	for _, bad := range []string{"https://evil.example", "/admin/regions/09?tab=configuration", "/admin/regions/03/municipalities/999999?tab=configuration"} {
		w = request("POST", "/admin/sources/milan-news/intervals?return_to="+url.QueryEscape(bad), intervals)
		if w.Code != 400 {
			t.Fatal("unsafe return accepted", w.Code)
		}
	}
	intervals.Set("expected_revision", "1")
	intervals.Add("actor", "duplicate")
	if w = request("POST", "/admin/sources/milan-news/intervals", intervals); w.Code != 400 {
		t.Fatal("duplicate accepted", w.Code)
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1/admin/regions/03/enabled", strings.NewReader(enabled.Encode()))
	r.Header.Set("Origin", "https://evil.example")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("foreign origin", w.Code)
	}
	for _, path := range []string{"/admin/regions/03", "/admin/regions/03?tab=history", "/admin/regions/03/municipalities/015146"} {
		if w = request("GET", path, nil); w.Code != 200 {
			t.Fatal(w.Code, w.Body)
		}
	}
	var count int
	if err = p.QueryRow(ctx, `SELECT count(*) FROM processing_jobs`).Scan(&count); err != nil || count != 0 {
		t.Fatal("navigation scheduled jobs", count, err)
	}
	st, err := reg.State(ctx, "milan-news")
	if err != nil || st.CollectionEnabled || st.PublicEnabled {
		t.Fatal("draft activated", st, err)
	}
}

func TestTerritoryDashboardBrowser(t *testing.T) {
	node := os.Getenv("IWA_BROWSER_NODE")
	if node == "" {
		t.Skip("set IWA_BROWSER_NODE and IWA_PLAYWRIGHT_MODULE")
	}
	p := territorialAdminDB(t)
	adoptTestMunicipalities(t, p, "09", 125)
	h := httptest.NewServer(HandlerWithAdministration(nil, AdminRuntime{Territories: operations.New(p), TerritoryConfig: domain.New(p), Registry: registry.New(p)}))
	defer h.Close()
	script, err := filepath.Abs("../../scripts/check-territories.cjs")
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(node, script, h.URL).CombinedOutput()
	if err != nil {
		t.Fatalf("browser: %v\n%s", err, out)
	}
	t.Log(string(out))
	var count int
	if err = p.QueryRow(context.Background(), `SELECT count(*) FROM processing_jobs`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("navigation/setup enqueued work: %d %v", count, err)
	}
}
