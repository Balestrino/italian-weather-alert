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

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publiccopy"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publicview"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"github.com/Balestrino/italian-weather-alert/internal/testfixtures/cfrgraphics"
	"github.com/jackc/pgx/v5/pgxpool"
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

	t.Run("development-municipality-publication", func(t *testing.T) {
		publication := domain.NewDevelopmentPublication(pool, "development")
		state, e := publication.State(ctx, "09", "050004")
		if e != nil || state.Enabled || state.Revision != 0 {
			t.Fatal(state, e)
		}
		if e = domain.MigrateTerritories(ctx, pool); e != nil {
			t.Fatal(e)
		}
		for _, mode := range []string{"production", "staging", ""} {
			if _, e = domain.NewDevelopmentPublication(pool, mode).Set(ctx, "09", "050004", 0, true, "test"); !errors.Is(e, domain.ErrDevelopmentOnly) {
				t.Fatal("non-development mutation allowed", mode, e)
			}
		}
		configuration, e := reg.Version(ctx, "criticality-public", 1)
		if e != nil {
			t.Fatal(e)
		}
		makePending := func(id, product, territoryID string) {
			t.Helper()
			cfg := configuration.Configuration
			source := registry.Source{ID: id, AuthorityID: "regione-toscana", ChannelID: "cfr", ProductID: product, Territory: territoryID}
			if e := reg.CreateSource(ctx, source, cfg, "fixture"); e != nil {
				t.Fatal(e)
			}
			if e := reg.RecordPreview(ctx, id, 1, "fixture", *cfg.Provenance); e != nil {
				t.Fatal(e)
			}
			if e := reg.EnableCollection(ctx, id, 1, "fixture"); e != nil {
				t.Fatal(e)
			}
		}
		makePending("pending-criticality", "criticality", "Toscana")
		makePending("pending-calcinaia", "municipal", "050004")
		otherState, e := domainStore.MunicipalityState(ctx, "09", "050026")
		if e != nil {
			t.Fatal(e)
		}
		if _, e = domainStore.SetMunicipalityEnabled(ctx, "09", "050026", otherState.Revision, true, "fixture"); e != nil {
			t.Fatal(e)
		}
		makePending("pending-pisa", "municipal", "050026")
		vr := retain("pending-regional-v1", "pending-criticality", regionalURL, "synthetic bulletin", nil, metadata)
		record := regional
		record.ID = "pending-regional"
		record.SourceID = "pending-criticality"
		record.DocumentVersionID = vr.ID
		record.Facts = []domain.RegionalFact{{Ordinal: 1, Risk: domain.RiskThunderstorms, OfficialRiskLabel: "Temporali", Zone: "A4", Level: &yellow}, {Ordinal: 2, Risk: domain.RiskThunderstorms, OfficialRiskLabel: "Temporali", Zone: "A3", Level: &yellow}}
		if e = domainStore.PutRegional(ctx, record); e != nil {
			t.Fatal(e)
		}
		vm := retain("pending-municipal-v1", "pending-calcinaia", municipalURL, "synthetic closure", nil, metadata)
		vo := retain("pending-other-v1", "pending-pisa", "https://regione.example/notices/other", "synthetic other closure", nil, metadata)
		local := measure
		local.ID = "pending-local"
		local.SourceID = "pending-calcinaia"
		local.DocumentVersionID = vm.ID
		if e = domainStore.PutLocalMeasure(ctx, local); e != nil {
			t.Fatal(e)
		}
		if e = domainStore.RecordInterpretation(ctx, domain.InterpretationAssessment{MeasureID: local.ID, State: "supported", EvidenceDocumentVersionID: vm.ID, Reason: "synthetic passage", Limitations: []string{}, Actor: "fixture", RecordedAt: time.Now()}); e != nil {
			t.Fatal(e)
		}
		if _, e = pool.Exec(ctx, `INSERT INTO geography_zone_mappings(dataset_id,municipality_dataset_id,municipality_istat,ordinal,zone,source_name,territorial_scope,evidence_locator) VALUES('zones-v2','municipalities-v1','050026',3,'A4','Pisa','partial_municipality','synthetic shared zone'),('zones-v1','municipalities-v1','050026',3,'A4','Pisa','partial_municipality','synthetic shared zone')`); e != nil {
			t.Fatal(e)
		}

		dev := NewForEnvironment(pool, "development")
		scopeBefore, e := dev.PublicationScope(ctx)
		if e != nil {
			t.Fatal(e)
		}
		assertRegional := func(store *Store, istat string, want int) {
			t.Helper()
			found, e := store.Search(ctx, SearchQuery{Kind: "regional", SourceID: "pending-criticality", MunicipalityISTAT: istat})
			if e != nil || len(found.Regional) != want {
				t.Fatalf("regional scope %s: %#v %v", istat, found, e)
			}
			for _, f := range found.Regional {
				if f.Zone != "A4" {
					t.Fatal("unselected zone exposed", f)
				}
			}
		}
		assertRegional(dev, "", 0)
		if _, e = publication.Set(ctx, "09", "050004", 0, true, "reviewer"); e != nil {
			t.Fatal(e)
		}
		if _, e = publication.Set(ctx, "09", "050004", 0, false, "stale"); !errors.Is(e, domain.ErrConflict) {
			t.Fatal("missing CAS", e)
		}
		if _, e = publication.Set(ctx, "09", "999999", 0, true, "reviewer"); !errors.Is(e, domain.ErrTerritoryNotFound) {
			t.Fatal("unknown municipality", e)
		}
		assertRegional(dev, "050004", 1)
		assertRegional(dev, "050026", 0)
		assertRegional(dev, "", 1)
		for _, mode := range []string{"production", "staging", ""} {
			assertRegional(NewForEnvironment(pool, mode), "050004", 0)
		}
		situation, e := dev.MunicipalitySituation(ctx, SituationQuery{MunicipalityISTAT: "050004"})
		if e != nil || !situation.Municipality.DevelopmentPublication {
			t.Fatal(situation, e)
		}
		foundLocal := false
		for _, m := range situation.LocalMeasures {
			if m.ID == "pending-local" {
				foundLocal = true
				if !strings.Contains(strings.Join(m.Quality.Interpretation.Limitations, " "), "Development publication") {
					t.Fatal("missing warning")
				}
			}
		}
		if !foundLocal {
			t.Fatal("municipal measure hidden")
		}
		coverage, e := dev.Coverage(ctx, CoverageQuery{MunicipalityISTAT: "050004", SourceID: "pending-criticality"})
		if e != nil || len(coverage.Sources) != 1 || !coverage.Sources[0].DevelopmentPublication || coverage.Sources[0].PublicState != "pending" || coverage.Sources[0].CoverageStatus != "pending" {
			t.Fatal("acceptance misrepresented", coverage, e)
		}
		if _, e = dev.Document(ctx, DocumentQuery{DocumentID: stringID(vm.DocumentID), VersionID: stringID(vm.ID)}); e != nil {
			t.Fatal("chosen document hidden", e)
		}
		if _, e = dev.Document(ctx, DocumentQuery{DocumentID: stringID(vo.DocumentID), VersionID: stringID(vo.ID)}); !errors.Is(e, ErrUnknownIdentifier) {
			t.Fatal("other document leaked", e)
		}
		if _, e = store.Document(ctx, DocumentQuery{DocumentID: stringID(vm.DocumentID)}); !errors.Is(e, ErrUnknownIdentifier) {
			t.Fatal("strict document leaked", e)
		}
		rows, e := dev.Search(ctx, SearchQuery{Kind: "document", SourceID: "dpc-internal"})
		if e != nil || len(rows.Documents) != 0 {
			t.Fatal("DPC leaked", rows, e)
		}
		if _, e = copyAccess.Read(ctx, stringID(vm.DocumentID), stringID(vm.ID)); !errors.Is(e, publiccopy.ErrRestricted) {
			t.Fatal("unaccepted copy exposed", e)
		}
		st, e := reg.State(ctx, "pending-criticality")
		if e != nil || st.PublicEnabled {
			t.Fatal("global publication changed", st, e)
		}
		var acceptances int
		if e = pool.QueryRow(ctx, `SELECT count(*) FROM registry_events WHERE source_id='pending-criticality' AND kind='acceptance'`).Scan(&acceptances); e != nil || acceptances != 0 {
			t.Fatal("fabricated acceptance", acceptances, e)
		}

		// An earlier query mapping can differ from today's regional configuration.
		if _, e = pool.Exec(ctx, `INSERT INTO geography_zone_mappings(dataset_id,municipality_dataset_id,municipality_istat,ordinal,zone,source_name,territorial_scope,evidence_locator) VALUES('zones-v1','municipalities-v1','050004',4,'A2','Calcinaia','partial_municipality','synthetic historical mapping')`); e != nil {
			t.Fatal(e)
		}
		historic := record
		historic.ID = "pending-historic-zone"
		historic.Facts = []domain.RegionalFact{{Ordinal: 1, Risk: domain.RiskThunderstorms, OfficialRiskLabel: "Temporali", Zone: "A2", Level: &yellow}}
		if e = domainStore.PutRegional(ctx, historic); e != nil {
			t.Fatal(e)
		}
		historical, e := dev.Search(ctx, SearchQuery{Kind: "regional", SourceID: "pending-criticality", MunicipalityISTAT: "050004"})
		if e != nil || len(historical.Regional) != 2 {
			t.Fatal("query mapping lost historical applicability", historical, e)
		}
		scopeEnabled, e := dev.PublicationScope(ctx)
		if e != nil || scopeEnabled == scopeBefore {
			t.Fatal("view scope unchanged", e)
		}
		if _, e = publication.Set(ctx, "09", "050004", 1, false, "reviewer"); e != nil {
			t.Fatal(e)
		}
		assertRegional(dev, "050004", 0)
		scopeRevoked, e := dev.PublicationScope(ctx)
		if e != nil || scopeRevoked == scopeEnabled || scopeRevoked == scopeBefore {
			t.Fatal("revocation did not invalidate saved views", e)
		}
		if _, e = pool.Exec(ctx, `DELETE FROM development_publication_events`); e == nil {
			t.Fatal("audit mutable")
		}
	})

	t.Run("verification projections preserve knowledge and current deduplication", func(t *testing.T) {
		proof := registry.Evidence{URL: "https://calcinaia.example/notices", Locator: "synthetic referral", ObservedAt: time.Now().UTC()}
		contract := &registry.CittadinoInformatoContract{MunicipalityISTAT: "050004", MunicipalitySlug: "calcinaia", Publisher: "comune_calcinaia", Updates: true, PageSize: 1}
		sections := []string{contract.BaseURL() + "aggiornamenti/"}
		cfg := registry.Configuration{URL: contract.BaseURL(), Sections: sections, AccessMethod: registry.CittadinoInformatoAccess, Attribution: "synthetic", Policy: registry.Policy{Evidence: &proof, CollectionPermitted: true, RetentionPermitted: true}, CittadinoInformato: contract, Discovery: registry.Discovery{MaxPagesPerSection: 1, MaxDocuments: 5, BootstrapDays: 30}, Referral: &registry.Referral{Evidence: proof, Destination: contract.BaseURL(), Territory: "050004", ProductID: "municipal", Sections: sections, Context: "synthetic scoped referral"}}
		cfg.Provenance = &proof
		cfg.Policy.PublicationPermitted = true
		if e := reg.CreateChannel(ctx, registry.Channel{ID: "verification-platform", PublisherID: "comune-calcinaia", Platform: "cittadino-informato", URL: contract.BaseURL(), External: true}); e != nil {
			t.Fatal(e)
		}
		if e := reg.CreateSource(ctx, registry.Source{ID: "verification-platform", AuthorityID: "comune-calcinaia", ChannelID: "verification-platform", ProductID: "municipal", Territory: "050004"}, cfg, "fixture"); e != nil {
			t.Fatal(e)
		}
		makeVerified := func(requestID, action string) domain.VerificationReceipt {
			t.Helper()
			body := "Ordinanza sintetica 88 | edizione 1 | " + action + " | Ponte Sintetico | fino a nuovo ordine"
			url := "https://calcinaia.example/notices/verified-primary"
			version := retain(requestID, "calcinaia-public", url, body, &issuer, metadata)
			fields := map[string]domain.EvidenceSelection{}
			for name, value := range map[string]string{"reference": "Ordinanza sintetica 88", "edition": "edizione 1", "kind": action, "subject": "Ponte Sintetico", "validity": "fino a nuovo ordine"} {
				start := strings.Index(body, value)
				fields[name] = domain.EvidenceSelection{ResourceURL: url, StartByte: start, EndByte: start + len(value), Locator: "synthetic primary " + name}
			}
			at := time.Now().UTC()
			request := domain.VerificationRequest{RequestID: requestID, CandidateKey: "primary-only-act", Kind: "local_measure", MunicipalityISTAT: "050004", Candidate: domain.VerificationEvidence{SourceID: "calcinaia-public", VersionID: version.ID, Fields: fields}, Checks: []domain.VerificationCheck{{Role: "regional", State: "not_applicable", Reason: "local_only", CheckedAt: at}, {Role: "platform", State: "missing", Reason: "not_republished", CheckedAt: at, VerificationEvidence: domain.VerificationEvidence{SourceID: "verification-platform"}}}}
			receipt, e := domainStore.VerifyMultiSource(ctx, documentStore, request, at)
			if e != nil || !receipt.Admitted {
				t.Fatal("primary-only projection failed", e, receipt.Outcome)
			}
			return receipt
		}
		first := makeVerified("verified-primary-first", "chiusura")
		second := makeVerified("verified-primary-revision", "riapertura")
		for _, receipt := range []domain.VerificationReceipt{first, second} {
			measures, e := store.measures(ctx, "050004", QueryTime{KnownAt: receipt.VerifiedAt, EvaluationTime: receipt.VerifiedAt})
			if e != nil {
				t.Fatal(e)
			}
			var relevant []Measure
			for _, measure := range measures {
				if measure.ID == first.DomainRecordID || measure.ID == second.DomainRecordID {
					relevant = append(relevant, measure)
				}
			}
			if len(relevant) != 1 || relevant[0].ID != receipt.DomainRecordID {
				t.Fatal("public current/history projection missing or duplicated", receipt.ID, relevant)
			}
			verification := relevant[0].Verifications
			if len(verification) != 1 || verification[0].ID != stringID(receipt.ID) || len(verification[0].Checks) != 3 {
				t.Fatal("public verification crossed the knowledge boundary or lost channel roles", verification)
			}
			if verification[0].Checks[0].Outcome != "not_applicable" || verification[0].Checks[2].Outcome != "missing_evidence" {
				t.Fatal("primary-only checks changed meaning", verification[0])
			}
			if !verification[0].Checks[1].EvidenceVisible || len(verification[0].Checks[1].Fields) == 0 {
				t.Fatal("public primary field evidence missing", verification[0])
			}
			encoded, e := json.Marshal(verification)
			if e != nil || strings.Contains(string(encoded), "request_id") || strings.Contains(string(encoded), "candidate_key") || strings.Contains(string(encoded), "not_republished") || strings.Contains(string(encoded), "verification-platform") {
				t.Fatal("private request/check data exposed", string(encoded), e)
			}
			if hidden, e := store.forMunicipality("050026").verifications(ctx, "local_measure", receipt.DomainRecordID, 0, QueryTime{KnownAt: receipt.VerifiedAt}); e != nil || len(hidden) != 0 {
				t.Fatal("unselected municipality received candidate evidence", hidden, e)
			}
			document, e := store.documentVersion(ctx, receipt.Candidate.VersionID, QueryTime{KnownAt: receipt.VerifiedAt, EvaluationTime: receipt.VerifiedAt})
			if e != nil || len(document.Verifications) != 1 || document.Verifications[0].ID != stringID(receipt.ID) {
				t.Fatal("document verification attribution missing", document, e)
			}
		}
		platformURL := contract.APIBase() + "aggiornamenti/123?comune=calcinaia"
		platformFields := map[string]string{"reference": "Ordinanza sintetica 88", "edition": "edizione 1", "kind": "riapertura", "subject": "Ponte Sintetico", "validity": "fino a nuovo ordine"}
		platformBody, e := json.Marshal(platformFields)
		if e != nil {
			t.Fatal(e)
		}
		platformVersion, e := documentStore.Retain(ctx, documents.Acquisition{ID: "verified-platform-conflict", SourceID: "verification-platform", Configuration: 1, URL: platformURL, Resources: []documents.Resource{{URL: platformURL, Role: "original", Required: true, SourceID: "verification-platform", Configuration: 1, MediaType: "application/json", Bytes: platformBody}}})
		if e != nil {
			t.Fatal(e)
		}
		selections := map[string]domain.EvidenceSelection{}
		for field, text := range platformFields {
			selections[field] = domain.EvidenceSelection{ResourceURL: platformURL, JSONPointer: "/" + field, StartByte: 0, EndByte: len(text), Locator: "synthetic platform " + field}
		}
		primaryFields := map[string]domain.EvidenceSelection{}
		for field, selected := range first.Candidate.Fields {
			primaryFields[field] = selected.EvidenceSelection
		}
		at := time.Now().UTC()
		conflict, e := domainStore.VerifyMultiSource(ctx, documentStore, domain.VerificationRequest{RequestID: "private-platform-conflict-request", CandidateKey: "private-platform-candidate", Kind: "local_measure", MunicipalityISTAT: "050004", Candidate: domain.VerificationEvidence{SourceID: "verification-platform", VersionID: platformVersion.ID, Fields: selections}, Checks: []domain.VerificationCheck{{Role: "regional", State: "not_applicable", Reason: "private local-only reason", CheckedAt: at}, {Role: "municipal", State: "available", Reason: "private counterpart lookup", CheckedAt: at, VerificationEvidence: domain.VerificationEvidence{SourceID: "calcinaia-public", VersionID: first.Candidate.VersionID, Fields: primaryFields}}}}, at)
		if e != nil || conflict.Admitted || conflict.Outcome != "conflict" || conflict.PrimaryRecordID != first.DomainRecordID {
			t.Fatal("conflicting republication overrode the primary", conflict, e)
		}
		if hidden, e := store.verifications(ctx, "", "", first.Candidate.VersionID, QueryTime{KnownAt: at}); e != nil || len(hidden) != 1 {
			t.Fatal("unpublished candidate leaked through primary document", hidden, e)
		}
		// Even when its candidate is published, an unpublished counterpart must
		// not disclose its field values, passages, hashes or operational reasons.
		redacted, e := store.publicVerification(ctx, conflict, QueryTime{KnownAt: at})
		if e != nil || redacted.Checks[2].EvidenceVisible || len(redacted.Checks[2].Fields) != 0 || redacted.Checks[2].SourceID != "" || redacted.Checks[2].Reason != "channel_evidence_not_public" {
			t.Fatal("private counterpart evidence exposed", redacted, e)
		}
		acceptance := registry.Acceptance{Report: proof, PeriodStart: at.Add(-8 * 24 * time.Hour), PeriodEnd: at, Sections: sections, ExtractionVerified: true, UpdatesVerified: true, AttachmentsVerified: true, ScannedAttachmentsVerified: true, HistoryVerified: true, FailureBehaviorVerified: true, InterfacesEquivalent: true, CoverageStatus: "accepted_with_limitations", CoverageLimitations: []string{"synthetic fixture only"}}
		acceptance.Report.ObservedAt = at
		for _, e := range []error{reg.RecordPreview(ctx, "verification-platform", 1, "fixture", proof), reg.EnableCollection(ctx, "verification-platform", 1, "fixture"), reg.Accept(ctx, "verification-platform", 1, "fixture", acceptance)} {
			if e != nil {
				t.Fatal(e)
			}
		}
		if _, e = reg.ReviewAcceptance(ctx, "verification-platform", 1, "fixture", acceptance.Report); e != nil {
			t.Fatal(e)
		}
		if e = reg.EnablePublic(ctx, "verification-platform", 1, "fixture"); e != nil {
			t.Fatal(e)
		}
		published, e := store.verifications(ctx, "", "", platformVersion.ID, QueryTime{KnownAt: at})
		if e != nil || len(published) != 1 || published[0].Checks[2].Outcome != "conflict" || !published[0].Checks[2].EvidenceVisible || published[0].Checks[2].FromVersionID != stringID(platformVersion.ID) || published[0].Checks[2].ToVersionID != stringID(first.Candidate.VersionID) {
			t.Fatal("published conflict lost primary direction/attribution", published, e)
		}
		platformDoc, e := store.documentVersion(ctx, platformVersion.ID, QueryTime{KnownAt: at, EvaluationTime: at})
		if e != nil || platformDoc.CopyURL != nil || len(platformDoc.Verifications) != 1 {
			t.Fatal("platform document lost link-only verification", platformDoc, e)
		}
	})

	t.Run("primary-extraction-projection", func(t *testing.T) {
		createSource(registry.Source{ID: "projection-primary", AuthorityID: "comune-calcinaia", ChannelID: "calcinaia", ProductID: "municipal", Territory: "050004"}, "https://calcinaia.example/projection", true)
		version := retain("projection-closure", "projection-primary", "https://calcinaia.example/projection/closure", "Chiusura del ponte sintetico in Via Sintetica fino a revoca", &issuer, metadata)
		provenanceBoundary := time.Now().UTC().Truncate(time.Microsecond)
		time.Sleep(time.Millisecond)
		for i := 0; i < 2; i++ {
			if e := domainStore.RecordConfiguredProvenance(ctx, "projection-primary", time.Now()); e != nil {
				t.Fatal(e)
			}
		}
		var provenanceCount int
		if e := pool.QueryRow(ctx, `SELECT count(*) FROM domain_provenance_events WHERE source_id='projection-primary'`).Scan(&provenanceCount); e != nil || provenanceCount != 1 {
			t.Fatal("provenance repeated", provenanceCount, e)
		}
		currentQuality, e := store.quality(ctx, "projection-primary", QueryTime{KnownAt: time.Now(), EvaluationTime: time.Now()}, Dimension{State: "supported"})
		if e != nil || currentQuality.Provenance.State != "verified" {
			t.Fatal("registered evidence lost", currentQuality, e)
		}
		pastQuality, e := store.quality(ctx, "projection-primary", QueryTime{KnownAt: provenanceBoundary, EvaluationTime: provenanceBoundary}, Dimension{State: "supported"})
		if e != nil || pastQuality.Provenance.State != "unresolved" {
			t.Fatal("future provenance leaked", pastQuality, e)
		}
		processingStore := processing.New(pool)
		classCatalog, e := classification.RegisterCatalog(ctx, processingStore, "openai-chat", "synthetic-model", time.Now())
		if e != nil {
			t.Fatal(e)
		}
		extractCatalog, e := extraction.RegisterCatalog(ctx, processingStore, "openai-chat", "synthetic-model", time.Now())
		if e != nil {
			t.Fatal(e)
		}
		sourceID := "projection-primary"
		classRun, e := processingStore.StartRun(ctx, processing.RunRequest{IdempotencyKey: "projection-class", Workload: "ordinary", Stage: "classification", ConfigurationVersionID: classCatalog.ConfigurationVersionID, SourceID: &sourceID, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{}`), CreatedAt: time.Now()})
		if e != nil {
			t.Fatal(e)
		}
		relevant := true
		if e = classification.NewStore(pool).Put(ctx, classification.Result{RunID: classRun.ID, DocumentVersionID: version.ID, Status: "classified", Relevant: &relevant, ReasonCode: "local_weather_measure", EvidenceQuote: "Chiusura del ponte sintetico", ContentSHA256: strings.Repeat("a", 64), ContentComplete: true, ProviderResponseID: "synthetic", ReturnedModel: "synthetic-model", CreatedAt: time.Now()}); e != nil {
			t.Fatal(e)
		}
		put := func(key, workload string) int64 {
			t.Helper()
			run, e := processingStore.StartRun(ctx, processing.RunRequest{IdempotencyKey: key, Workload: workload, Stage: "extraction", ConfigurationVersionID: extractCatalog.ConfigurationVersionID, SourceID: &sourceID, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{}`), CreatedAt: time.Now()})
			if e != nil {
				t.Fatal(e)
			}
			m := extraction.Measure{Ordinal: 1, Kind: "closure", Subject: "ponte sintetico", Place: stringPointer("Via Sintetica"), ValidUntil: stringPointer("fino a revoca"), IndeterminateFields: []string{"valid_from"}, Evidence: []extraction.Evidence{{Field: "kind", ResourceURL: "https://calcinaia.example/projection/closure", Quote: "Chiusura"}, {Field: "subject", ResourceURL: "https://calcinaia.example/projection/closure", Quote: "ponte sintetico"}, {Field: "place", ResourceURL: "https://calcinaia.example/projection/closure", Quote: "Via Sintetica"}, {Field: "valid_until", ResourceURL: "https://calcinaia.example/projection/closure", Quote: "fino a revoca"}}}
			if e = extraction.NewStore(pool).Put(ctx, extraction.Result{RunID: run.ID, DocumentVersionID: version.ID, ClassificationRunID: classRun.ID, Status: "extracted", ReasonCode: "measures_extracted", ContentSHA256: strings.Repeat("b", 64), ContentComplete: true, ProviderResponseID: "synthetic", ReturnedModel: "synthetic-model", CreatedAt: time.Now(), Measures: []extraction.Measure{m}}); e != nil {
				t.Fatal(e)
			}
			return run.ID
		}
		listing := retain("projection-listing", "projection-primary", "https://calcinaia.example/projection?page=1", "Synthetic section container", &issuer, metadata)
		situation, e := store.MunicipalitySituation(ctx, SituationQuery{MunicipalityISTAT: "050004"})
		if e != nil {
			t.Fatal(e)
		}
		firstVisible := false
		for _, d := range situation.DocumentsRequiringAttention {
			if d.ID == stringID(listing.DocumentID) {
				t.Fatal("listing reported as a pending notice")
			}
			if d.ID == stringID(version.DocumentID) {
				firstVisible = true
			}
		}
		if !firstVisible {
			t.Fatal("first notice disappeared without a previous measure")
		}
		runID := put("projection-extract", "ordinary")
		progress, e := store.Coverage(ctx, CoverageQuery{SourceID: sourceID, MunicipalityISTAT: "050004"})
		if e != nil || len(progress.Sources) != 1 || progress.Sources[0].Quality.Interpretation.State != "partial" {
			t.Fatal("projection gap not reflected in source interpretation", progress, e)
		}
		before := time.Now().UTC()
		if d, e := store.Document(ctx, DocumentQuery{DocumentID: stringID(version.DocumentID), QueryTime: QueryTime{KnownAt: before, EvaluationTime: before}}); e != nil || d.Document.Quality.Interpretation.State != "partial" {
			t.Fatal("unprojected extraction claimed complete", d, e)
		}
		for i := 0; i < 2; i++ {
			if n, e := domainStore.ProjectMunicipalExtraction(ctx, runID, time.Now()); e != nil || n != 1 {
				t.Fatal("idempotent primary projection", n, e)
			}
		}
		q := SearchQuery{Kind: "measure", SourceID: sourceID, MunicipalityISTAT: "050004"}
		found, e := store.Search(ctx, q)
		if e != nil || len(found.Measures) != 1 || len(found.Measures[0].Evidence) != 4 || found.Measures[0].Validity.Precision != "conditional" || found.Measures[0].Status != "undetermined" {
			t.Fatal("primary evidence/unknown validity lost", found, e)
		}
		q.QueryTime = QueryTime{KnownAt: before, EvaluationTime: before}
		if found, e = store.Search(ctx, q); e != nil || len(found.Measures) != 0 {
			t.Fatal("future projection leaked into history", found, e)
		}
		evaluationID := put("projection-evaluation", "evaluation")
		if n, e := domainStore.ProjectMunicipalExtraction(ctx, evaluationID, time.Now()); e != nil || n != 0 {
			t.Fatal("evaluation published facts", n, e)
		}
		if e = reg.SuspendInterpretation(ctx, sourceID, 1, "fixture", registry.Evidence{URL: "https://calcinaia.example/projection", Locator: "synthetic defect", ObservedAt: time.Now()}); e != nil {
			t.Fatal(e)
		}
		if n, e := domainStore.ProjectMunicipalExtraction(ctx, runID, time.Now()); e != nil || n != 0 {
			t.Fatal("suspended source projected facts", n, e)
		}
	})

	t.Run("CFR graphics preserve weather and prior knowledge", func(t *testing.T) {
		if e := domain.MigrateTerritories(ctx, pool); e != nil {
			t.Fatal(e)
		}
		if e := domainStore.BackfillToscana(ctx); e != nil {
			t.Fatal(e)
		}
		sourceID, url := "vigilance-graphics", "https://regione.example/cfr/vigilance-graphics"
		createSource(registry.Source{ID: sourceID, AuthorityID: "regione-toscana", ChannelID: "cfr", ProductID: "vigilance", Territory: "Toscana"}, url, false)
		html, bbox, pages := cfrgraphics.Vigilance()
		reader := graphicsFixtureReader{acquisition.VectorEvidence{BBox: bbox, Pages: pages}}
		version, e := documentStore.Retain(ctx, documents.Acquisition{ID: "vigilance-graphics-version", SourceID: sourceID, Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{
			{URL: url, Role: "original", Required: true, SourceID: sourceID, Configuration: 1, MediaType: "text/html", Bytes: html},
			{URL: url + ".pdf", Role: "resource", Required: true, SourceID: sourceID, Configuration: 1, MediaType: "application/pdf", Bytes: []byte("%PDF-synthetic-reader-fixture")},
		}})
		if e != nil {
			t.Fatal(e)
		}
		// Preserve one legacy table-only projection to verify the knowledge boundary.
		if _, e = pool.Exec(ctx, `INSERT INTO domain_regional_projections(document_version_id,logic_version,source_id,product,status,statement,limitations) VALUES($1,'cfr-vector-v3',$2,'vigilance','partial','legacy table only','[]')`, version.ID, sourceID); e != nil {
			t.Fatal(e)
		}
		if _, e = pool.Exec(ctx, `INSERT INTO domain_regional_records(id,document_version_id,source_id,product,originating_authority_id,publisher_id,platform,municipal_republication,projection_logic,evidence_locator) VALUES('legacy-vigilance',$1,$2,'vigilance','regione-toscana','regione-toscana','CFR',false,'cfr-vector-v3','legacy HTML')`, version.ID, sourceID); e != nil {
			t.Fatal(e)
		}
		if _, e = pool.Exec(ctx, `INSERT INTO domain_regional_facts(regional_record_id,product,ordinal,risk,official_risk_label,zone,level) VALUES('legacy-vigilance','vigilance',1,'wind','Vento','A4','not_applicable')`); e != nil {
			t.Fatal(e)
		}
		before := time.Now().UTC()
		for i := 0; i < 2; i++ {
			report, e := domainStore.ProjectCFRWithReader(ctx, documentStore, version.ID, reader)
			if e != nil || report.Facts != 364 || len(report.Limitations) != 0 {
				t.Fatal("graphics projection", report, e)
			}
		}
		query := SearchQuery{Kind: "regional", SourceID: sourceID, Zone: "A4"}
		found, e := store.Search(ctx, query)
		if e != nil || len(found.Regional) != 14 {
			t.Fatal("replacement projection missing or duplicated", len(found.Regional), e)
		}
		for _, f := range found.Regional {
			if f.Weather == nil || f.Level != "not_applicable" || len(f.Evidence) != 1 || f.Evidence[0].SourceURL != url+".pdf" || f.Evidence[0].Page == nil || *f.Evidence[0].Page != 1 {
				t.Fatal("persisted weather/evidence lost", f)
			}
			if f.Weather.Phenomenon == "rainfall" && f.Weather.TotalRainfallBand != "40 - 60" {
				t.Fatal("rainfall cumulative meaning lost", f.Weather)
			}
		}
		query.QueryTime = QueryTime{KnownAt: before, EvaluationTime: before}
		past, e := store.Search(ctx, query)
		if e != nil || len(past.Regional) != 1 || past.Regional[0].Weather != nil || !strings.HasPrefix(past.Regional[0].ID, "legacy-vigilance") {
			t.Fatal("past knowledge rewritten", past, e)
		}
		var count int
		if e = pool.QueryRow(ctx, `SELECT count(*) FROM domain_regional_records WHERE document_version_id=$1`, version.ID).Scan(&count); e != nil || count != 365 {
			t.Fatal("immutable history deleted or duplicated", count, e)
		}
		coverage, e := store.Coverage(ctx, CoverageQuery{SourceID: sourceID})
		if e != nil || len(coverage.Sources) != 1 || coverage.Sources[0].Quality.Interpretation.State != "supported" {
			t.Fatal("complete graphical interpretation reported partial", coverage, e)
		}
		// Incomplete required evidence cannot enter the vector reader/projection.
		incomplete, e := documentStore.Retain(ctx, documents.Acquisition{ID: "vigilance-graphics-incomplete", SourceID: sourceID, Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{
			{URL: url, Role: "original", Required: true, SourceID: sourceID, Configuration: 1, MediaType: "text/html", Bytes: html},
			{URL: url + ".pdf", Role: "resource", Required: true, SourceID: sourceID, Configuration: 1, Missing: "unavailable"},
		}})
		if e != nil {
			t.Fatal(e)
		}
		report, e := domainStore.ProjectCFRWithReader(ctx, documentStore, incomplete.ID, reader)
		if e != nil || report.Status != "unsupported" || report.Facts != 0 {
			t.Fatal("incomplete PDF evidence projected", report, e)
		}
		query.QueryTime = QueryTime{}
		found, e = store.Search(ctx, query)
		if e != nil || len(found.Regional) != 14 || !strings.Contains(strings.Join(found.Regional[0].Quality.Interpretation.Limitations, " "), "newer retained bulletin") {
			t.Fatal("preceding state or failure warning lost", len(found.Regional), e)
		}
	})
}

func stringPointer(value string) *string { return &value }

type graphicsFixtureReader struct{ evidence acquisition.VectorEvidence }

func (r graphicsFixtureReader) Read(context.Context, []byte) (acquisition.VectorEvidence, error) {
	return r.evidence, nil
}
