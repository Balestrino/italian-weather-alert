//go:build integration

package publicquery

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/domain"
	"github.com/Balestrino/italian-weather-alert/internal/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/linking"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/publiccopy"
	"github.com/Balestrino/italian-weather-alert/internal/publicview"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
	"github.com/Balestrino/italian-weather-alert/internal/territory"
)

type memoryObjects struct {
	mu    sync.Mutex
	items map[string][]byte
}

func (m *memoryObjects) Ensure(_ context.Context, hash string, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.items == nil {
		m.items = map[string][]byte{}
	}
	m.items[hash] = append([]byte(nil), body...)
	return nil
}

func (m *memoryObjects) Read(_ context.Context, hash string, _ int64) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]byte(nil), m.items[hash]...), nil
}

func publicQueryTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		t.Fatal(err)
	}
	name, password := "iwa-public-query-test-"+hex.EncodeToString(token[:6]), hex.EncodeToString(token)
	passwordFile := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("isolated PostgreSQL operation failed: %v: %s", err, strings.ReplaceAll(string(out), password, "[redacted]"))
		}
		return strings.TrimSpace(string(out))
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	binding := listener.Addr().String() + ":5432"
	_ = listener.Close()
	run("run", "--pull=never", "--detach", "--name", name, "--publish", binding, "--mount", "type=bind,src="+passwordFile+",dst=/run/secrets/password,readonly", "--env", "POSTGRES_PASSWORD_FILE=/run/secrets/password", "--env", "POSTGRES_USER=iwa", "--env", "POSTGRES_DB=iwa", "postgres:16-alpine")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", name).Run() })
	address := run("port", name, "5432/tcp")
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatal(err)
	}
	portNumber, _ := strconv.Atoi(port)
	cfg, err := pgxpool.ParseConfig("postgres://iwa@localhost/iwa?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.Host, cfg.ConnConfig.Port, cfg.ConnConfig.Password = host, uint16(portNumber), password
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	deadline := time.Now().Add(25 * time.Second)
	for pool.Ping(context.Background()) != nil {
		if time.Now().After(deadline) {
			t.Fatal("test PostgreSQL did not become ready")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return pool
}

func TestFiveSharedPublicQueryGroups(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := publicQueryTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, acquisition.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, domain.Migrate, domain.MigrateGeography, domain.MigrateTemporal, domain.MigrateQuality} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("idempotent public-query migration: %v", err)
	}
	if err := publicview.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := publicview.Migrate(ctx, pool); err != nil {
		t.Fatalf("idempotent public-view migration: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	reg := registry.New(pool)
	for _, authority := range []registry.Authority{
		{ID: "regione-toscana", Name: "Regione Toscana", OfficialURL: "https://regione.example"},
		{ID: "comune-calcinaia", Name: "Comune di Calcinaia", OfficialURL: "https://calcinaia.example"},
		{ID: "dpc", Name: "Dipartimento della Protezione Civile", OfficialURL: "https://dpc.example"},
	} {
		if err := reg.CreateAuthority(ctx, authority); err != nil {
			t.Fatal(err)
		}
	}
	for _, channel := range []registry.Channel{
		{ID: "cfr", PublisherID: "regione-toscana", Platform: "CFR", URL: "https://regione.example/cfr"},
		{ID: "calcinaia", PublisherID: "comune-calcinaia", Platform: "municipal-site", URL: "https://calcinaia.example"},
		{ID: "dpc", PublisherID: "dpc", Platform: "internal-comparison", URL: "https://dpc.example/data"},
	} {
		if err := reg.CreateChannel(ctx, channel); err != nil {
			t.Fatal(err)
		}
	}
	createSource := func(source registry.Source, section string, copiesPermitted bool) {
		t.Helper()
		evidence := registry.Evidence{URL: section, Locator: "reviewed fixture evidence", ObservedAt: now}
		configuration := registry.Configuration{URL: section, Sections: []string{section}, AccessMethod: "fixture", Attribution: source.AuthorityID, Provenance: &evidence, Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true, PublicationPermitted: true, CopiesPermitted: copiesPermitted}, Limitations: []string{"fixture scope only"}}
		acceptance := registry.Acceptance{Report: registry.Evidence{URL: section, Locator: "fixture acceptance", ObservedAt: now}, PeriodStart: now.Add(-9 * 24 * time.Hour), PeriodEnd: now.Add(-time.Hour), Sections: []string{section}, ExtractionVerified: true, UpdatesVerified: true, AttachmentsVerified: true, ScannedAttachmentsVerified: true, HistoryVerified: true, FailureBehaviorVerified: true, InterfacesEquivalent: true, CoverageStatus: "accepted_with_limitations", CoverageLimitations: append([]string(nil), configuration.Limitations...)}
		if source.ProductID == "vigilance" || source.ProductID == "criticality" || source.ProductID == "monitoring" {
			acceptance.RiskCoverage = append([]string(nil), registry.ToscanaRisks...)
		}
		for _, err := range []error{
			reg.CreateSource(ctx, source, configuration, "test"),
			reg.RecordPreview(ctx, source.ID, 1, "test", evidence),
			reg.EnableCollection(ctx, source.ID, 1, "test"),
			reg.Accept(ctx, source.ID, 1, "test", acceptance),
		} {
			if err != nil {
				t.Fatal(err)
			}
		}
		review, err := reg.ReviewAcceptance(ctx, source.ID, 1, "release-reviewer", registry.Evidence{URL: section, Locator: "final fixture acceptance review", ObservedAt: now})
		if err != nil || review.Status != "accepted" {
			t.Fatalf("release review: %#v %v", review, err)
		}
		if err = reg.EnablePublic(ctx, source.ID, 1, "test"); err != nil {
			t.Fatal(err)
		}
	}
	createSource(registry.Source{ID: "calcinaia-public", AuthorityID: "comune-calcinaia", ChannelID: "calcinaia", ProductID: "municipal", Territory: "050004"}, "https://calcinaia.example/notices", true)
	createSource(registry.Source{ID: "criticality-public", AuthorityID: "regione-toscana", ChannelID: "cfr", ProductID: "criticality", Territory: "Toscana"}, "https://regione.example/cfr/criticality", false)
	dpcEvidence := registry.Evidence{URL: "https://dpc.example/data", Locator: "internal comparison fixture", ObservedAt: now}
	dpcConfiguration := registry.Configuration{URL: dpcEvidence.URL, Sections: []string{dpcEvidence.URL}, AccessMethod: "fixture", Attribution: "dpc", Provenance: &dpcEvidence, Policy: registry.Policy{Evidence: &dpcEvidence, CollectionPermitted: true, RetentionPermitted: true, PublicationPermitted: true}, Limitations: []string{"internal comparison only"}}
	for _, err := range []error{
		reg.CreateSource(ctx, registry.Source{ID: "dpc-internal", AuthorityID: "dpc", ChannelID: "dpc", ProductID: "dpc_comparison", Territory: "Toscana"}, dpcConfiguration, "test"),
		reg.RecordPreview(ctx, "dpc-internal", 1, "test", dpcEvidence),
		reg.EnableCollection(ctx, "dpc-internal", 1, "test"),
		reg.Accept(ctx, "dpc-internal", 1, "test", registry.Acceptance{Report: dpcEvidence, PeriodStart: now.Add(-9 * 24 * time.Hour), PeriodEnd: now.Add(-time.Hour), Sections: []string{dpcEvidence.URL}, ExtractionVerified: true, UpdatesVerified: true, AttachmentsVerified: true, ScannedAttachmentsVerified: true, HistoryVerified: true, FailureBehaviorVerified: true, InterfacesEquivalent: true, CoverageStatus: "accepted_with_limitations", CoverageLimitations: append([]string(nil), dpcConfiguration.Limitations...)}),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := reg.EnablePublic(ctx, "dpc-internal", 1, "test"); !errors.Is(err, registry.ErrPrerequisite) {
		t.Fatalf("internal DPC source became publicly eligible: %v", err)
	}

	domainStore := domain.New(pool)
	municipalities := domain.Dataset{ID: "municipalities-v1", Kind: domain.DatasetMunicipalities, Authority: "ISTAT", Publisher: "ISTAT", OfficialURL: "https://istat.example/municipalities", VersionLabel: "2026", SourceSHA256: strings.Repeat("a", 64), VerifiedAt: now, Applicability: "verified", Limitations: []string{}, Usable: true}
	zones := domain.Dataset{ID: "zones-v1", Kind: domain.DatasetZones, MunicipalityDatasetID: municipalities.ID, Authority: "Regione Toscana", Publisher: "Regione Toscana", OfficialURL: "https://regione.example/zones", VersionLabel: "2026", SourceSHA256: strings.Repeat("b", 64), VerifiedAt: now, Applicability: "verified", Limitations: []string{}, Usable: true}
	postal := domain.Dataset{ID: "postal-v1", Kind: domain.DatasetPostal, MunicipalityDatasetID: municipalities.ID, Authority: "Dataset CAP", Publisher: "Dataset CAP", OfficialURL: "https://cap.example/data", VersionLabel: "2026", SourceSHA256: strings.Repeat("c", 64), VerifiedAt: now, Applicability: "not_applicable", Limitations: []string{}, Usable: true}
	if err := domainStore.RegisterMunicipalities(ctx, municipalities, []domain.Municipality{{ISTAT: "050004", Name: "Calcinaia", Province: "Pisa", Supported: true}, {ISTAT: "050026", Name: "Pisa", Province: "Pisa", Supported: true}}); err != nil {
		t.Fatal(err)
	}
	if err := domainStore.RegisterZones(ctx, zones, []domain.ZoneMapping{{MunicipalityISTAT: "050004", Zone: "A4", SourceName: "Calcinaia", TerritorialScope: "whole_municipality", EvidenceLocator: "row 1"}, {MunicipalityISTAT: "050026", Zone: "A3", SourceName: "Pisa", TerritorialScope: "whole_municipality", EvidenceLocator: "row 2"}}); err != nil {
		t.Fatal(err)
	}
	if err := domainStore.RegisterPostal(ctx, postal, []domain.PostalMapping{{PostalCode: "56012", MunicipalityISTAT: "050004"}}); err != nil {
		t.Fatal(err)
	}
	for _, selection := range []struct{ kind, id string }{{domain.DatasetMunicipalities, municipalities.ID}, {domain.DatasetZones, zones.ID}, {domain.DatasetPostal, postal.ID}} {
		if err := domainStore.SelectDataset(ctx, selection.kind, selection.id, "test", now); err != nil {
			t.Fatal(err)
		}
	}
	zonesV2 := domain.Dataset{ID: "zones-v2", Kind: domain.DatasetZones, MunicipalityDatasetID: municipalities.ID, Authority: "Regione Toscana", Publisher: "Regione Toscana", OfficialURL: "https://regione.example/zones-v2", VersionLabel: "2026.2", SourceSHA256: strings.Repeat("d", 64), VerifiedAt: now.Add(time.Minute), Applicability: "verified", Limitations: []string{}, Usable: true}
	if err := domainStore.RegisterZones(ctx, zonesV2, []domain.ZoneMapping{{MunicipalityISTAT: "050004", Zone: "A4", SourceName: "Calcinaia", TerritorialScope: "whole_municipality", EvidenceLocator: "row 1"}, {MunicipalityISTAT: "050026", Zone: "A3", SourceName: "Pisa", TerritorialScope: "whole_municipality", EvidenceLocator: "row 2"}}); err != nil {
		t.Fatal(err)
	}
	if err := domainStore.SelectDataset(ctx, domain.DatasetZones, zonesV2.ID, "test", now.Add(4*time.Minute)); err != nil {
		t.Fatal(err)
	}

	objects := &memoryObjects{}
	documentStore := documents.New(pool, objects)
	issuer := "comune-calcinaia"
	municipalURL := "https://calcinaia.example/notices/closure"
	metadata, _ := json.Marshal(map[string]any{"kind": "notice", "publication": map[string]any{"original": "17 settembre 2026 ore 10:00", "precision": "instant", "instant": now.Add(-2 * time.Hour).Format(time.RFC3339)}})
	retain := func(id, source, url, body string, issuerID *string, raw json.RawMessage) documents.Version {
		t.Helper()
		version, err := documentStore.Retain(ctx, documents.Acquisition{ID: id, SourceID: source, Configuration: 1, URL: url, IssuerID: issuerID, Metadata: raw, Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: source, Configuration: 1, MediaType: "text/html", Bytes: []byte(body)}}})
		if err != nil {
			t.Fatal(err)
		}
		return version
	}
	municipalV1 := retain("municipal-v1", "calcinaia-public", municipalURL, "chiusura del sottopasso fino a nuova comunicazione", &issuer, metadata)
	time.Sleep(time.Millisecond)
	municipalV2 := retain("municipal-v2", "calcinaia-public", municipalURL, "successiva comunicazione non interpretabile", &issuer, metadata)
	regionalURL := "https://regione.example/cfr/criticality/bulletin"
	regionalV1 := retain("regional-v1", "criticality-public", regionalURL, "criticità gialla futura A4", nil, metadata)
	_ = retain("dpc-v1", "dpc-internal", "https://dpc.example/data/bulletin", "record DPC esclusivamente interno", nil, metadata)
	factsBoundary := time.Now().UTC()
	time.Sleep(time.Millisecond)

	measure := domain.LocalMeasure{ID: "closure-maremmana", DocumentVersionID: municipalV1.ID, SourceID: "calcinaia-public", MunicipalityISTAT: "050004", Kind: "closure", Subject: "sottopasso", Place: stringPointer("via Maremmana")}
	phase := domain.OperationalPhase{ID: "coc-active", DocumentVersionID: municipalV1.ID, SourceID: "calcinaia-public", MunicipalityISTAT: "050004", Phase: "COC attivo"}
	yellow := "yellow"
	regional := domain.RegionalRecord{ID: "criticality-future", DocumentVersionID: regionalV1.ID, SourceID: "criticality-public", Product: domain.ProductCriticality, OriginatingAuthorityID: "regione-toscana", Facts: []domain.RegionalFact{{Ordinal: 1, Risk: domain.RiskThunderstorms, OfficialRiskLabel: "Temporali forti", Zone: "A4", Level: &yellow}}}
	for _, err := range []error{domainStore.PutLocalMeasure(ctx, measure), domainStore.PutOperationalPhase(ctx, phase), domainStore.PutRegional(ctx, regional)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	start, end := now.Add(-time.Hour), now.Add(time.Hour)
	futureStart, futureEnd := now.Add(2*time.Hour), now.Add(8*time.Hour)
	condition := "fino a revoca"
	for _, temporal := range []domain.TemporalValue{
		{EntityKind: "local_measure", EntityID: measure.ID, Meaning: "validity", Original: "dalle 11 alle 13", Precision: "interval", Instant: &start, EndInstant: &end, EvidenceDocumentVersionID: municipalV1.ID, CreatedAt: now},
		{EntityKind: "operational_phase", EntityID: phase.ID, Meaning: "validity", Original: "fino a revoca", Precision: "conditional", Condition: &condition, EvidenceDocumentVersionID: municipalV1.ID, CreatedAt: now},
		{EntityKind: "regional_record", EntityID: regional.ID, Meaning: "validity", Original: "dalle 14 alle 20", Precision: "interval", Instant: &futureStart, EndInstant: &futureEnd, EvidenceDocumentVersionID: regionalV1.ID, CreatedAt: now},
	} {
		if _, err := domainStore.PutTemporal(ctx, temporal); err != nil {
			t.Fatal(err)
		}
	}
	for _, source := range []string{"calcinaia-public", "criticality-public"} {
		if err := domainStore.RecordProvenance(ctx, domain.ProvenanceAssessment{SourceID: source, Configuration: 1, State: "verified", EvidenceURL: "https://" + strings.Split(source, "-")[0] + ".example/evidence", EvidenceLocator: "fixture", Limitations: []string{}, AssessedAt: now}); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO acquisition_checks(source_id,configuration,worker_id,started_at,finished_at,reachable,content_recognized,complete,listing_count,document_count,check_state,publication_state) VALUES($1,1,'fixture',$2,$3,true,true,true,1,1,'complete_changed','observed')`, source, now.Add(-time.Minute), now); err != nil {
			t.Fatal(err)
		}
	}
	if err := domainStore.RecordInterpretation(ctx, domain.InterpretationAssessment{MeasureID: measure.ID, State: "supported", EvidenceDocumentVersionID: municipalV1.ID, Reason: "reviewed passage", Limitations: []string{}, Actor: "fixture", RecordedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := domainStore.RecordPrecedingStateWarning(ctx, measure.ID, municipalV2.ID, "newer document not interpreted", now); err != nil {
		t.Fatal(err)
	}

	queryTime := QueryTime{EvaluationTime: now.Add(5 * time.Minute), KnownAt: now.Add(5 * time.Minute)}
	store := New(pool)
	var versionsBefore, checksBefore, runsBefore int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM retained_versions),(SELECT count(*) FROM acquisition_checks),(SELECT count(*) FROM processing_runs)`).Scan(&versionsBefore, &checksBefore, &runsBefore); err != nil {
		t.Fatal(err)
	}
	discovery, err := store.DiscoverMunicipalities(ctx, DiscoveryQuery{QueryTime: queryTime, PostalCode: "56012"})
	if err != nil || len(discovery.Municipalities) != 1 || discovery.Municipalities[0].ISTAT != "050004" || discovery.Municipalities[0].LocalCoverage != "enabled" || len(discovery.Municipalities[0].Zones) != 1 || discovery.Municipalities[0].MappingVersion == nil || *discovery.Municipalities[0].MappingVersion != zonesV2.ID {
		t.Fatalf("discovery: %#v %v", discovery, err)
	}
	situation, err := store.MunicipalitySituation(ctx, SituationQuery{QueryTime: queryTime, MunicipalityISTAT: "050004"})
	if err != nil {
		t.Fatal(err)
	}
	if len(situation.LocalMeasures) != 1 || situation.LocalMeasures[0].Status != "current" || len(situation.RegionalProducts) != 1 || situation.RegionalProducts[0].Status != "future" || len(situation.DocumentsRequiringAttention) != 1 {
		t.Fatalf("situation did not prioritize local/current and separate future regional data: %#v", situation)
	}
	if len(situation.Coverage) != 2 || situation.Coverage[0].Product != "municipal" || situation.Coverage[1].Product != "criticality" {
		t.Fatalf("regional/local coverage was not separate: %#v", situation.Coverage)
	}
	from := now.Add(-365 * 24 * time.Hour)
	search, err := store.Search(ctx, SearchQuery{QueryTime: queryTime, Kind: "measure", MunicipalityISTAT: "050004", Status: "current", From: &from})
	if err != nil || len(search.Measures) != 1 || len(search.History.Gaps) == 0 {
		t.Fatalf("measure/history search: %#v %v", search, err)
	}
	documentSearch, err := store.Search(ctx, SearchQuery{QueryTime: queryTime, Kind: "document"})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range documentSearch.Documents {
		if item.SourceID == "dpc-internal" || strings.Contains(item.OfficialURL, "dpc.example") {
			t.Fatalf("internal DPC record escaped through public search: %#v", item)
		}
	}
	document, err := store.Document(ctx, DocumentQuery{QueryTime: queryTime, DocumentID: stringID(municipalV1.DocumentID), IncludeVersions: true})
	if err != nil || len(document.Versions) != 2 || document.Document.VersionID != stringID(municipalV2.ID) || document.Document.OfficialURL != municipalURL || document.Document.InterpretedAt != nil || document.Document.Kind != "notice" || document.Document.CopyURL != nil {
		t.Fatalf("document versions: %#v %v", document, err)
	}
	copyAccess, err := publiccopy.New(pool, documentStore, true, "https://alerts.example")
	if err != nil {
		t.Fatal(err)
	}
	copyLink, err := copyAccess.Link(ctx, stringID(municipalV2.DocumentID), stringID(municipalV2.ID))
	if err != nil || copyLink == nil || *copyLink != "https://alerts.example/v1/documents/"+stringID(municipalV2.DocumentID)+"/versions/"+stringID(municipalV2.ID)+"/content" {
		t.Fatalf("exact public copy link: %v %v", copyLink, err)
	}
	copyContent, err := copyAccess.Read(ctx, stringID(municipalV2.DocumentID), stringID(municipalV2.ID))
	if err != nil || string(copyContent.Bytes) != "successiva comunicazione non interpretabile" || copyContent.MediaType != "text/html" {
		t.Fatalf("exact uninterpreted copy: %#v %v", copyContent, err)
	}
	previousCopy, err := copyAccess.Read(ctx, stringID(municipalV1.DocumentID), stringID(municipalV1.ID))
	if err != nil || string(previousCopy.Bytes) != "chiusura del sottopasso fino a nuova comunicazione" || previousCopy.SHA256 == copyContent.SHA256 {
		t.Fatalf("stable URL versions were not exact: %#v %v", previousCopy, err)
	}
	if restricted, err := copyAccess.Link(ctx, stringID(regionalV1.DocumentID), stringID(regionalV1.ID)); err != nil || restricted != nil {
		t.Fatalf("restricted source published a copy link: %v %v", restricted, err)
	}
	if _, err := copyAccess.Read(ctx, stringID(regionalV1.DocumentID), stringID(regionalV1.ID)); !errors.Is(err, publiccopy.ErrRestricted) {
		t.Fatalf("restricted source copy read: %v", err)
	}
	disabledCopies, err := publiccopy.New(pool, nil, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if disabled, err := disabledCopies.Link(ctx, stringID(municipalV2.DocumentID), stringID(municipalV2.ID)); err != nil || disabled != nil {
		t.Fatalf("default-off service published a copy link: %v %v", disabled, err)
	}
	if _, err := disabledCopies.Read(ctx, stringID(municipalV2.DocumentID), stringID(municipalV2.ID)); !errors.Is(err, publiccopy.ErrDisabled) {
		t.Fatalf("default-off copy read: %v", err)
	}
	coverage, err := store.Coverage(ctx, CoverageQuery{QueryTime: queryTime, MunicipalityISTAT: "050004"})
	if err != nil || len(coverage.Sources) != 2 || coverage.History.Start == nil {
		t.Fatalf("coverage/history: %#v %v", coverage, err)
	}
	for _, source := range coverage.Sources {
		if source.CoverageStatus != "accepted_with_limitations" || len(source.CoverageLimitations) == 0 {
			t.Fatalf("bounded acceptance scope missing: %#v", source)
		}
	}
	viewStore := publicview.New(pool)
	view, err := viewStore.Create(ctx, publicview.Create{Operation: "get_source_coverage", RequestHash: strings.Repeat("a", 64), Request: json.RawMessage(`{"municipality_istat":"050004","page_size":1}`), Data: json.RawMessage(`[{"source_id":"calcinaia-public"}]`), History: json.RawMessage(`{"gaps":[],"limitations":[]}`), EvaluationTime: queryTime.EvaluationTime, KnownAt: queryTime.KnownAt, CreatedAt: now, ExpiresAt: now.Add(30 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	loadedView, err := viewStore.Get(ctx, view.ID)
	if err != nil || loadedView.RequestHash != view.RequestHash || loadedView.Operation != view.Operation || !loadedView.ExpiresAt.Equal(view.ExpiresAt) {
		t.Fatalf("persistent public view did not round-trip: %#v %v", loadedView, err)
	}
	historicalSituation, err := store.MunicipalitySituation(ctx, SituationQuery{QueryTime: QueryTime{EvaluationTime: queryTime.EvaluationTime, KnownAt: factsBoundary}, MunicipalityISTAT: "050004"})
	if err != nil || len(historicalSituation.LocalMeasures) != 0 || len(historicalSituation.RegionalProducts) != 0 || historicalSituation.Municipality.MappingVersion == nil || *historicalSituation.Municipality.MappingVersion != zones.ID {
		t.Fatalf("known_at leaked facts interpreted later: %#v %v", historicalSituation, err)
	}

	earlierKnownAt := municipalV2.FirstAcquiredAt.Add(-time.Nanosecond)
	knownBeforeRevision, err := store.Document(ctx, DocumentQuery{QueryTime: QueryTime{EvaluationTime: queryTime.EvaluationTime, KnownAt: earlierKnownAt}, DocumentID: stringID(municipalV1.DocumentID), IncludeVersions: true})
	if err != nil || len(knownBeforeRevision.Versions) != 1 || knownBeforeRevision.Document.VersionID != stringID(municipalV1.ID) {
		t.Fatalf("known_at leaked a later revision: %#v %v", knownBeforeRevision, err)
	}
	var versionsAfter, checksAfter, runsAfter int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM retained_versions),(SELECT count(*) FROM acquisition_checks),(SELECT count(*) FROM processing_runs)`).Scan(&versionsAfter, &checksAfter, &runsAfter); err != nil {
		t.Fatal(err)
	}
	if versionsAfter != versionsBefore || checksAfter != checksBefore || runsAfter != runsBefore {
		t.Fatalf("public queries changed acquisition or inference state: before=%d/%d/%d after=%d/%d/%d", versionsBefore, checksBefore, runsBefore, versionsAfter, checksAfter, runsAfter)
	}
	// Previewing a draft of an already public source must not publish its new
	// bytes, replace the latest accepted version, or expose a retained-copy link.
	active, err := reg.Version(ctx, "calcinaia-public", 1)
	if err != nil {
		t.Fatal(err)
	}
	draft := active.Configuration
	draft.Sections = []string{"https://calcinaia.example/new-section"}
	if _, err = reg.AppendConfiguration(ctx, "calcinaia-public", 1, draft, "draft-editor"); err != nil {
		t.Fatal(err)
	}
	draftVersion, err := documentStore.Retain(ctx, documents.Acquisition{
		ID: "draft-preview", SourceID: "calcinaia-public", Configuration: 2, URL: municipalURL, IssuerID: &issuer, Metadata: metadata,
		Resources: []documents.Resource{{URL: municipalURL, Role: "original", Required: true, SourceID: "calcinaia-public", Configuration: 2, MediaType: "text/html", Bytes: []byte("draft preview unpublished original")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	previewTime := time.Now().UTC()
	qt := QueryTime{EvaluationTime: previewTime, KnownAt: previewTime}
	visible, err := store.Document(ctx, DocumentQuery{QueryTime: qt, DocumentID: stringID(municipalV1.DocumentID), IncludeVersions: true})
	if err != nil || len(visible.Versions) != 2 || visible.Document.VersionID != stringID(municipalV2.ID) {
		t.Fatalf("draft replaced public version: %#v %v", visible, err)
	}
	if _, err = store.Document(ctx, DocumentQuery{QueryTime: qt, DocumentID: stringID(draftVersion.DocumentID), VersionID: stringID(draftVersion.ID)}); !errors.Is(err, ErrUnknownIdentifier) {
		t.Fatal("draft details exposed", err)
	}
	visibleSearch, err := store.Search(ctx, SearchQuery{QueryTime: qt, Kind: "document"})
	if err != nil {
		t.Fatal(err)
	}
	for _, document := range visibleSearch.Documents {
		if document.VersionID == stringID(draftVersion.ID) {
			t.Fatal("draft in public search")
		}
	}
	if link, err := copyAccess.Link(ctx, stringID(draftVersion.DocumentID), stringID(draftVersion.ID)); err != nil || link != nil {
		t.Fatal("draft copy link", link, err)
	}
	if _, err = copyAccess.Read(ctx, stringID(draftVersion.DocumentID), stringID(draftVersion.ID)); !errors.Is(err, publiccopy.ErrRestricted) {
		t.Fatal("draft copy exposed", err)
	}
	if err = reg.SetIntervals(ctx, "calcinaia-public", 0, 30, 90, "interval-editor"); err != nil {
		t.Fatal(err)
	}
	updatedCoverage, err := store.Coverage(ctx, CoverageQuery{QueryTime: qt, SourceID: "calcinaia-public"})
	if err != nil || len(updatedCoverage.Sources) != 1 || updatedCoverage.Sources[0].Quality.Updating.DelayThresholdSeconds != 90 || updatedCoverage.Sources[0].PublicState != "enabled" {
		t.Fatalf("immediate interval update changed acceptance or missed public threshold: %#v %v", updatedCoverage, err)
	}

	// A confirmed source defect freezes interpreted facts and validity, including
	// work that was already running, without hiding newly acquired documents.
	defect := registry.Evidence{URL: "https://calcinaia.example/defect", Locator: "fixture confirmed defect", ObservedAt: time.Now().UTC()}
	if err = reg.SuspendInterpretation(ctx, "calcinaia-public", 1, "test", defect); err != nil {
		t.Fatal(err)
	}
	if _, e := viewStore.Get(ctx, view.ID); !errors.Is(e, publicview.ErrNotFound) {
		t.Fatalf("cached view survived suspension: %v", e)
	}
	lateMeasure := measure
	lateMeasure.ID = "during-suspension"
	if err = domainStore.PutLocalMeasure(ctx, lateMeasure); err != nil {
		t.Fatal(err)
	}
	if err = domainStore.RecordInterpretation(ctx, domain.InterpretationAssessment{MeasureID: lateMeasure.ID, Actor: "test", Limitations: []string{}, State: "supported", EvidenceDocumentVersionID: municipalV1.ID, Reason: "fixture", RecordedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	replacementStart, replacementEnd := now.Add(-2*time.Hour), now.Add(-time.Hour)
	if _, err = domainStore.PutTemporal(ctx, domain.TemporalValue{EntityKind: "local_measure", EntityID: measure.ID, Meaning: "validity", Original: "unvalidated expiry", Precision: "interval", Instant: &replacementStart, EndInstant: &replacementEnd, EvidenceDocumentVersionID: municipalV1.ID, CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	suspensionQueryAt := time.Now()
	qt = QueryTime{EvaluationTime: suspensionQueryAt, KnownAt: suspensionQueryAt}
	frozen, err := store.Search(ctx, SearchQuery{QueryTime: qt, Kind: "measure", MunicipalityISTAT: "050004"})
	if err != nil || len(frozen.Measures) != 1 || frozen.Measures[0].ID != measure.ID || frozen.Measures[0].Status != "current" || frozen.Measures[0].Quality.Interpretation.State != "unreliable" {
		t.Fatalf("suspension leaked new interpretation or expiry: %#v %v", frozen, err)
	}
	visible, err = store.Document(ctx, DocumentQuery{QueryTime: qt, DocumentID: stringID(municipalV2.DocumentID)})
	if err != nil || visible.Document.OfficialURL != municipalURL || visible.Document.Quality.Interpretation.State != "unreliable" {
		t.Fatalf("document hidden during suspension: %#v %v", visible, err)
	}

	// Completed-interval fixture: the HTTP test separately exercises the validated
	// recovery operation. Old facts keep their warning; a historical view still
	// respects the suspension boundary after the source resumes.
	if _, err = pool.Exec(ctx, `WITH ended AS (UPDATE registry_interpretation_suspensions SET resumed_at=clock_timestamp(),resumed_by='fixture',recovery='{}' WHERE source_id='calcinaia-public' AND resumed_at IS NULL RETURNING source_id) UPDATE registry_sources SET interpretation_suspended_at=NULL FROM ended WHERE id=ended.source_id`); err != nil {
		t.Fatal(err)
	}
	resumedAt := time.Now()
	resumedQuery := QueryTime{EvaluationTime: resumedAt, KnownAt: resumedAt}
	resumed, err := store.Search(ctx, SearchQuery{QueryTime: resumedQuery, Kind: "measure", MunicipalityISTAT: "050004"})
	if err != nil || len(resumed.Measures) != 2 {
		t.Fatalf("resumed facts: %#v %v", resumed, err)
	}
	for _, m := range resumed.Measures {
		if m.ID == measure.ID && m.Quality.Interpretation.State != "unreliable" {
			t.Fatal("resume rehabilitated an old defective interpretation")
		}
	}
	historical, err := store.Search(ctx, SearchQuery{QueryTime: qt, Kind: "measure", MunicipalityISTAT: "050004"})
	if err != nil || len(historical.Measures) != 1 || historical.Measures[0].Status != "current" {
		t.Fatalf("historical suspension boundary lost: %#v %v", historical, err)
	}

	t.Run("territorial-public-perimeter", func(t *testing.T) {
		if err := domain.MigrateTerritories(ctx, pool); err != nil {
			t.Fatal(err)
		}
		if err := domainStore.BackfillToscana(ctx); err != nil {
			t.Fatal(err)
		}
		before, err := store.DiscoverMunicipalities(ctx, DiscoveryQuery{})
		if err != nil || len(before.Municipalities) != 2 {
			t.Fatal(before, err)
		}
		v, err := reg.Version(ctx, "criticality-public", 1)
		if err != nil {
			t.Fatal(err)
		}
		if err = reg.CreateSource(ctx, registry.Source{ID: "foreign-criticality", AuthorityID: "regione-toscana", ChannelID: "cfr", ProductID: "criticality", Territory: "03"}, v.Configuration, "test"); err != nil {
			t.Fatal(err)
		}
		if err = reg.EnablePublic(ctx, "foreign-criticality", 1, "test"); !errors.Is(err, territory.ErrPublicScope) {
			t.Fatal("foreign publication enabled", err)
		}
		// Deliberately bypass source gates to prove readers independently constrain scope.
		if _, err = pool.Exec(ctx, `UPDATE registry_sources SET active_revision=1,public_enabled=true WHERE id='foreign-criticality'`); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO registry_events(source_id,revision,kind,actor,evidence) VALUES('foreign-criticality',1,'acceptance','test','{}')`); err != nil {
			t.Fatal(err)
		}
		foreignVersion := retain("foreign-version", "foreign-criticality", regionalURL, "Foreign bulletin with overlapping A4 zone", nil, metadata)
		foreignRecord := regional
		foreignRecord.ID = "foreign-fact"
		foreignRecord.SourceID = "foreign-criticality"
		foreignRecord.DocumentVersionID = foreignVersion.ID
		if err = domainStore.PutRegional(ctx, foreignRecord); err != nil {
			t.Fatal(err)
		}
		results, err := store.Search(ctx, SearchQuery{Kind: "regional"})
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range results.Regional {
			if r.ID == "foreign-fact" {
				t.Fatal("foreign warning exposed")
			}
		}
		docs, err := store.Search(ctx, SearchQuery{Kind: "document", SourceID: "foreign-criticality"})
		if err != nil || len(docs.Documents) != 0 {
			t.Fatal("foreign search exposed", err)
		}
		if _, err = store.Document(ctx, DocumentQuery{DocumentID: stringID(foreignVersion.DocumentID), VersionID: stringID(foreignVersion.ID)}); err == nil {
			t.Fatal("foreign document exposed")
		}
		if _, err = copyAccess.Read(ctx, stringID(foreignVersion.DocumentID), stringID(foreignVersion.ID)); !errors.Is(err, publiccopy.ErrUnknown) {
			t.Fatal("foreign copy exposed", err)
		}
		coverage, err := store.Coverage(ctx, CoverageQuery{SourceID: "foreign-criticality"})
		if err != nil || len(coverage.Sources) != 0 {
			t.Fatal("foreign coverage exposed", err)
		}
		after, err := store.DiscoverMunicipalities(ctx, DiscoveryQuery{})
		if err != nil || len(after.Municipalities) != len(before.Municipalities) {
			t.Fatal("discovery changed", err)
		}
	})

}

func stringPointer(value string) *string { return &value }
