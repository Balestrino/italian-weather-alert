//go:build integration

package diagnostics

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/domain"
	"github.com/Balestrino/italian-weather-alert/internal/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/operations"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

func TestDiagnosticsRetainScopedEvidenceWithoutChangingRegionalWarnings(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := diagnosticsTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{
		registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate,
		acquisition.Migrate, classification.Migrate, extraction.Migrate, Migrate, domain.Migrate,
	} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Truncate(time.Second)
	reg := registry.New(pool)
	if err := reg.CreateAuthority(ctx, registry.Authority{ID: "region", Name: "Regione fixture", OfficialURL: "https://region.example"}); err != nil {
		t.Fatal(err)
	}
	if err := reg.CreateAuthority(ctx, registry.Authority{ID: "dpc", Name: "DPC fixture", OfficialURL: "https://dpc.example"}); err != nil {
		t.Fatal(err)
	}
	if err := reg.CreateChannel(ctx, registry.Channel{ID: "regional-channel", PublisherID: "region", Platform: "regional", URL: "https://region.example/alerts"}); err != nil {
		t.Fatal(err)
	}
	if err := reg.CreateChannel(ctx, registry.Channel{ID: "dpc-channel", PublisherID: "dpc", Platform: "repository", URL: "https://dpc.example/data"}); err != nil {
		t.Fatal(err)
	}
	evidence := registry.Evidence{URL: "https://region.example/policy", Locator: "reviewed fixture policy", ObservedAt: now.Add(-time.Hour)}
	regionalConfig := registry.Configuration{
		URL: "https://region.example/alerts/criticality", Sections: []string{"https://region.example/alerts/criticality"},
		AccessMethod: "fixture", Attribution: "region", DelaySeconds: 1800,
		Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true},
	}
	if err := reg.CreateSource(ctx, registry.Source{ID: "regional", AuthorityID: "region", ChannelID: "regional-channel", ProductID: "criticality", Territory: "Toscana"}, regionalConfig, "test"); err != nil {
		t.Fatal(err)
	}
	if err := reg.RecordPreview(ctx, "regional", 1, "test", evidence); err != nil {
		t.Fatal(err)
	}
	if err := reg.EnableCollection(ctx, "regional", 1, "test"); err != nil {
		t.Fatal(err)
	}
	dpcEvidence := registry.Evidence{URL: "https://dpc.example/license", Locator: "reviewed fixture distribution", ObservedAt: now.Add(-time.Hour)}
	dpcConfig := registry.Configuration{
		URL: "https://dpc.example/data", Sections: []string{"https://dpc.example/data"}, AccessMethod: "fixture", Attribution: "dpc",
		Policy: registry.Policy{Evidence: &dpcEvidence, CollectionPermitted: true, RetentionPermitted: true},
	}
	if err := reg.CreateSource(ctx, registry.Source{ID: "dpc-internal", AuthorityID: "dpc", ChannelID: "dpc-channel", ProductID: "dpc_comparison", Territory: "Italia"}, dpcConfig, "test"); err != nil {
		t.Fatal(err)
	}
	regionalURL := "https://region.example/alerts/criticality/map.png"
	dpcURL := "https://dpc.example/data/today.json"
	regionalVersion := retainFixture(t, ctx, pool, "regional", regionalURL, strings.Repeat("a", 64), now.Add(-20*time.Minute))
	dpcVersion := retainFixture(t, ctx, pool, "dpc-internal", dpcURL, strings.Repeat("b", 64), now.Add(-20*time.Minute))
	if _, err := pool.Exec(ctx, `INSERT INTO domain_regional_records(id,document_version_id,source_id,product,originating_authority_id,publisher_id,platform,municipal_republication)
 VALUES('warning-before',$1,'regional','criticality','region','region','regional',false)`, regionalVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO domain_regional_facts(regional_record_id,product,ordinal,risk,official_risk_label,zone,level)
 VALUES('warning-before','criticality',1,'wind','Vento','A4','yellow')`); err != nil {
		t.Fatal(err)
	}
	store := New(pool)
	snapshot, err := store.Snapshot(ctx, "regional", now)
	if err != nil || snapshot.Availability.State != "unverified" || snapshot.Timeliness.State != "unverified" || snapshot.Interpretation.State != "degraded" {
		t.Fatalf("initial diagnostics: %#v %v", snapshot, err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO acquisition_checks(source_id,configuration,worker_id,started_at,finished_at,reachable,content_recognized,complete,listing_count,document_count,check_state,publication_state)
 VALUES('regional',1,'fixture',$1,$2,true,true,true,1,1,'complete_unchanged','not_expected')`, now.Add(-11*time.Minute), now.Add(-10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO processing_configuration_versions(id,name,stage,revision,logic_version,settings,content_hash,created_at)
 VALUES('class-fixture','classification fixture','classification','1','fixture','{}',$1,$2)`, strings.Repeat("c", 64), now.Add(-9*time.Minute)); err != nil {
		t.Fatal(err)
	}
	var runID int64
	if err = pool.QueryRow(ctx, `INSERT INTO processing_runs(idempotency_key,request_hash,workload,stage,configuration_version_id,source_id,document_version_id,subject,created_at)
 VALUES('class-run',$1,'evaluation','classification','class-fixture','regional',$2,'{}',$3) RETURNING id`, strings.Repeat("d", 64), regionalVersion, now.Add(-8*time.Minute)).Scan(&runID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO classification_results(run_id,document_version_id,status,relevant,reason_code,evidence_quote,content_sha256,content_complete,provider_response_id,returned_model,created_at)
 VALUES($1,$2,'classified',false,'not_relevant','fixture',$3,true,'response','model',$4)`, runID, regionalVersion, strings.Repeat("e", 64), now.Add(-7*time.Minute)); err != nil {
		t.Fatal(err)
	}
	snapshot, err = store.Snapshot(ctx, "regional", now)
	if err != nil || snapshot.Availability.State != "available" || snapshot.Timeliness.State != "current" || snapshot.Interpretation.State != "unverified" {
		t.Fatalf("model output was incorrectly treated as reviewed correctness: %#v %v", snapshot, err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO registry_regressions(id,source_id,revision,suite,contract_hash,passed,actor,report,recorded_at)
 VALUES('reviewed-suite','regional',1,'human-review',$1,true,'reviewer','{}',$2)`, strings.Repeat("f", 64), now.Add(-6*time.Minute)); err != nil {
		t.Fatal(err)
	}
	snapshot, err = store.Snapshot(ctx, "regional", now)
	if err != nil || snapshot.Interpretation.State != "supported" {
		t.Fatalf("reviewed diagnostics: %#v %v", snapshot, err)
	}
	before := warningState(t, ctx, pool)
	dimensions := []Dimension{
		{Name: "risk", State: "same", RegionalValue: text("wind"), DPCValue: text("wind"), Regional: Evidence{ResourceURL: regionalURL, Locator: "regional risk row"}, DPC: Evidence{ResourceURL: dpcURL, Locator: "DPC risk field"}},
		{Name: "territory", State: "different", RegionalValue: text("A4"), DPCValue: text("Toscana"), Regional: Evidence{ResourceURL: regionalURL, Locator: "regional zone"}, DPC: Evidence{ResourceURL: dpcURL, Locator: "DPC territory"}},
		{Name: "issuance", State: "not_comparable", Reason: "different issuance cycles", Regional: Evidence{ResourceURL: regionalURL, Locator: "regional issue time"}, DPC: Evidence{ResourceURL: dpcURL, Locator: "DPC snapshot time"}},
		{Name: "validity", State: "not_comparable", Reason: "no matching DPC validity interval", Regional: Evidence{ResourceURL: regionalURL, Locator: "regional interval"}, DPC: Evidence{ResourceURL: dpcURL, Locator: "DPC forecast day"}},
	}
	comparison, err := store.Compare(ctx, Comparison{RegionalVersionID: regionalVersion, DPCVersionID: dpcVersion, Scope: "criticality/Toscana/2026-09-17", Actor: "reviewer", Dimensions: dimensions, ComparedAt: now})
	if err != nil || comparison.RegionalSourceID != "regional" || comparison.DPCSourceID != "dpc-internal" || comparison.ID == "" {
		t.Fatalf("comparison: %#v %v", comparison, err)
	}
	if _, err = store.Compare(ctx, comparison); err != nil {
		t.Fatalf("idempotent comparison: %v", err)
	}
	changed := comparison
	changed.Dimensions = append([]Dimension(nil), comparison.Dimensions...)
	changed.Dimensions[0].RegionalValue = text("ice")
	if _, err = store.Compare(ctx, changed); !errors.Is(err, ErrConflict) {
		t.Fatalf("conflicting comparison identity accepted: %v", err)
	}
	invalid := comparison
	invalid.ID = "wrong-evidence"
	invalid.Dimensions = append([]Dimension(nil), comparison.Dimensions...)
	invalid.Dimensions[0].DPC.ResourceURL = regionalURL
	if _, err = store.Compare(ctx, invalid); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cross-side evidence accepted: %v", err)
	}
	var storedRegional, storedDPC string
	var storedDimensions string
	if err = pool.QueryRow(ctx, `SELECT regional_source_id,dpc_source_id,dimensions::text FROM diagnostic_dpc_comparisons WHERE id=$1`, comparison.ID).Scan(&storedRegional, &storedDPC, &storedDimensions); err != nil {
		t.Fatal(err)
	}
	if storedRegional != "regional" || storedDPC != "dpc-internal" || !strings.Contains(storedDimensions, regionalURL) || !strings.Contains(storedDimensions, dpcURL) || !strings.Contains(storedDimensions, "not_comparable") {
		t.Fatalf("comparison evidence was not retained on both sides: %s", storedDimensions)
	}
	report, err := operations.New(pool).Page(ctx, "dpc-comparisons", operations.Filter{SourceID: "regional", Page: 1}, now)
	if err != nil || len(report.Table.Rows) != 1 || report.Table.Rows[0]["dpc_source_id"] != "dpc-internal" {
		t.Fatalf("internal comparison history: %#v %v", report, err)
	}
	after := warningState(t, ctx, pool)
	if before != after {
		t.Fatalf("internal DPC comparison altered regional warnings: before=%s after=%s", before, after)
	}
	if _, err = pool.Exec(ctx, "UPDATE diagnostic_dpc_comparisons SET actor='tampered' WHERE id=$1", comparison.ID); err == nil {
		t.Fatal("append-only diagnostic history was mutable")
	}
}

func retainFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sourceID, resourceURL, hash string, acquiredAt time.Time) int64 {
	t.Helper()
	var documentID, versionID int64
	if err := pool.QueryRow(ctx, "INSERT INTO retained_documents(source_id,official_url) VALUES($1,$2) RETURNING id", sourceID, resourceURL).Scan(&documentID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO retained_versions(document_id,content_hash,first_acquired_at,complete,metadata) VALUES($1,$2,$3,true,'{}') RETURNING id`, documentID, hash, acquiredAt).Scan(&versionID); err != nil {
		t.Fatal(err)
	}
	objectHash := strings.Repeat(hash[:1], 64)
	if _, err := pool.Exec(ctx, "INSERT INTO retained_objects(hash,object_key,byte_size) VALUES($1,$2,10) ON CONFLICT DO NOTHING", objectHash, "objects/"+objectHash); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO retained_resources(version_id,url,role,required,source_id,configuration,media_type,object_hash,missing) VALUES($1,$2,'original',true,$3,1,'application/octet-stream',$4,'')`, versionID, resourceURL, sourceID, objectHash); err != nil {
		t.Fatal(err)
	}
	return versionID
}

func warningState(t *testing.T, ctx context.Context, pool *pgxpool.Pool) string {
	t.Helper()
	var state string
	if err := pool.QueryRow(ctx, `SELECT jsonb_build_object(
 'records',(SELECT jsonb_agg(r ORDER BY r.id) FROM domain_regional_records r),
 'facts',(SELECT jsonb_agg(f ORDER BY f.regional_record_id,f.ordinal) FROM domain_regional_facts f),
 'public',(SELECT jsonb_agg(jsonb_build_object('id',id,'enabled',public_enabled) ORDER BY id) FROM registry_sources)
)::text`).Scan(&state); err != nil {
		t.Fatal(err)
	}
	return state
}

func diagnosticsTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	name := "iwa-diagnostics-test-" + hex.EncodeToString(random[:6])
	password := hex.EncodeToString(random)
	passwordFile := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		t.Helper()
		commandCtx, commandCancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer commandCancel()
		output, err := exec.CommandContext(commandCtx, "docker", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("isolated PostgreSQL operation failed: %v: %s", err, strings.ReplaceAll(string(output), password, "[redacted]"))
		}
		return strings.TrimSpace(string(output))
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
	configuration, err := pgxpool.ParseConfig("postgres://iwa@localhost/iwa?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	configuration.ConnConfig.Host = host
	configuration.ConnConfig.Port = uint16(portNumber)
	configuration.ConnConfig.Password = password
	pool, err := pgxpool.NewWithConfig(context.Background(), configuration)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	readyCtx, readyCancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer readyCancel()
	for pool.Ping(readyCtx) != nil {
		if readyCtx.Err() != nil {
			t.Fatal("test PostgreSQL did not become ready")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return pool
}
