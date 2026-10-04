//go:build integration

package observation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCampaignDerivesSevenDayEvidenceAndRefusesPrematureCompletion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := observationTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, acquisition.Migrate, Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Truncate(time.Second)
	started := now.Add(-8 * 24 * time.Hour)
	through := now.Add(-time.Hour)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://evidence.example/source-contract", Locator: "reviewed fixture source contract", ObservedAt: started.Add(-time.Hour)}
	for _, authority := range []registry.Authority{
		{ID: "regione", Name: "Regione Toscana", OfficialURL: "https://regione.example"},
		{ID: "calcinaia", Name: "Comune di Calcinaia", OfficialURL: "https://calcinaia.example"},
	} {
		if err := reg.CreateAuthority(ctx, authority); err != nil {
			t.Fatal(err)
		}
	}
	for _, channel := range []registry.Channel{
		{ID: "cfr", PublisherID: "regione", Platform: "CFR", URL: "https://regione.example/cfr"},
		{ID: "calcinaia", PublisherID: "calcinaia", Platform: "municipal", URL: "https://calcinaia.example"},
	} {
		if err := reg.CreateChannel(ctx, channel); err != nil {
			t.Fatal(err)
		}
	}
	type sourceFixture struct {
		id, product, territory, authority, channel string
	}
	fixtures := []sourceFixture{
		{"vigilance", "vigilance", "Toscana", "regione", "cfr"},
		{"criticality", "criticality", "Toscana", "regione", "cfr"},
		{"monitoring", "monitoring", "Toscana", "regione", "cfr"},
		{"calcinaia", "municipal", "050004", "calcinaia", "calcinaia"},
	}
	versions := map[string]int64{}
	for index, fixture := range fixtures {
		section := "https://regione.example/cfr/" + fixture.id
		if fixture.product == "municipal" {
			section = "https://calcinaia.example/notices"
		}
		configuration := registry.Configuration{
			URL: section, Sections: []string{section}, AccessMethod: "fixture", Attribution: fixture.authority,
			Provenance: &evidence, Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true},
			CheckSeconds: 600, DelaySeconds: 1800, Limitations: []string{"fixture only"},
		}
		for _, err := range []error{
			reg.CreateSource(ctx, registry.Source{ID: fixture.id, AuthorityID: fixture.authority, ChannelID: fixture.channel, ProductID: fixture.product, Territory: fixture.territory}, configuration, "test"),
			reg.RecordPreview(ctx, fixture.id, 1, "test", evidence),
			reg.EnableCollection(ctx, fixture.id, 1, "test"),
		} {
			if err != nil {
				t.Fatal(err)
			}
		}
		var documentID, versionID int64
		if err := pool.QueryRow(ctx, "INSERT INTO retained_documents(source_id,official_url) VALUES($1,$2) RETURNING id", fixture.id, section).Scan(&documentID); err != nil {
			t.Fatal(err)
		}
		hash := strings.Repeat(string(rune('a'+index)), 64)
		if err := pool.QueryRow(ctx, `INSERT INTO retained_versions(document_id,content_hash,first_acquired_at,complete,metadata)
 VALUES($1,$2,$3,true,'{}') RETURNING id`, documentID, hash, started.Add(time.Hour)).Scan(&versionID); err != nil {
			t.Fatal(err)
		}
		versions[fixture.id] = versionID
		for resourceIndex, role := range []string{"original", "resource"} {
			resourceHash := fmt.Sprintf("%064x", index*2+resourceIndex+1)
			if _, err := pool.Exec(ctx, "INSERT INTO retained_objects(hash,object_key,byte_size) VALUES($1,$2,10)", resourceHash, "objects/"+resourceHash); err != nil {
				t.Fatal(err)
			}
			if _, err := pool.Exec(ctx, `INSERT INTO retained_resources(version_id,url,role,required,source_id,configuration,media_type,object_hash,missing)
 VALUES($1,$2,$3,true,$4,1,$5,$6,'')`, versionID, section+"/"+role, role, fixture.id, "application/octet-stream", resourceHash); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := pool.Exec(ctx, `INSERT INTO acquisition_checks(source_id,configuration,worker_id,started_at,finished_at,reachable,content_recognized,complete,listing_count,document_count,check_state,publication_state)
 SELECT $1,1,'paced-fixture',point,point+interval '1 second',true,true,true,1,1,'complete_unchanged','not_expected'
 FROM generate_series($2::timestamptz,$3::timestamptz-interval '1 second',interval '10 minutes') point`, fixture.id, started, through); err != nil {
			t.Fatal(err)
		}
	}
	store := New(pool)
	ids := []string{"vigilance", "criticality", "monitoring", "calcinaia"}
	campaign, err := store.Start(ctx, StartRequest{ID: "toscana-calcinaia-2026-09", Actor: "operator", StartedAt: started, SourceIDs: ids})
	if err != nil || len(campaign.Sources) != 4 || !campaign.MinimumEndAt.Equal(started.Add(MinimumDuration)) {
		t.Fatalf("start campaign: %#v %v", campaign, err)
	}
	if _, err = store.Start(ctx, StartRequest{ID: campaign.ID, Actor: "operator", StartedAt: started, SourceIDs: ids}); !errors.Is(err, ErrConflict) {
		t.Fatalf("campaign identity was reusable: %v", err)
	}
	reviewed := registry.Evidence{URL: "https://evidence.example/trial-review", Locator: "human comparison fixture", ObservedAt: through}
	for _, fixture := range fixtures {
		version := versions[fixture.id]
		for _, review := range []Review{
			{ID: fixture.id + "-original", SourceID: fixture.id, Kind: "original_comparison", Status: "pass", VersionID: &version, Evidence: reviewed},
			{ID: fixture.id + "-attachment", SourceID: fixture.id, Kind: "attachment_comparison", Status: "pass", VersionID: &version, Evidence: reviewed},
			{ID: fixture.id + "-failure", SourceID: fixture.id, Kind: "failure_retained", Status: "pass", CaseID: "retained-429", Evidence: reviewed},
		} {
			if _, err = store.RecordReview(ctx, campaign.ID, "reviewer", review); err != nil {
				t.Fatalf("record %s: %v", review.ID, err)
			}
		}
	}
	premature, err := store.Assess(ctx, campaign.ID, "reviewer", started.Add(6*24*time.Hour), registry.Evidence{URL: reviewed.URL, Locator: "premature assessment", ObservedAt: through})
	if err != nil || premature.Status != "running" || !slicesContains(premature.Report.Issues, "minimum_duration_not_reached") {
		t.Fatalf("premature assessment: %#v %v", premature, err)
	}
	assessment, err := store.Assess(ctx, campaign.ID, "reviewer", through, registry.Evidence{URL: reviewed.URL, Locator: "seven-day assessment", ObservedAt: now})
	if err != nil || assessment.Status != "complete" || len(assessment.Report.Issues) != 0 || len(assessment.Report.Sources) != 4 {
		t.Fatalf("completed assessment: %#v %v", assessment, err)
	}
	for _, source := range assessment.Report.Sources {
		if source.ObservedLocalDays < 7 || source.CompleteChecks < 1000 || source.MaximumGapSeconds > int64(source.DelaySeconds) || source.OriginalComparisons != 1 || source.AttachmentComparisons != 1 || source.FailureExercises != 1 {
			t.Fatalf("incomplete source evidence: %#v", source)
		}
	}
	if _, err = store.Assess(ctx, campaign.ID, "reviewer", through, registry.Evidence{URL: reviewed.URL, Locator: "duplicate completion", ObservedAt: now}); !errors.Is(err, ErrConflict) {
		t.Fatalf("completed campaign accepted another assessment: %v", err)
	}
	partial, err := store.Start(ctx, StartRequest{ID: "partial", Actor: "operator", StartedAt: started, SourceIDs: []string{"calcinaia"}})
	if err != nil || len(partial.Sources) != 1 {
		t.Fatalf("partial campaign start: %#v %v", partial, err)
	}
	report, err := store.Report(ctx, partial.ID, through)
	if err != nil || report.Status != "extended" || len(report.MissingProducts) != 3 {
		t.Fatalf("partial campaign was not extended: %#v %v", report, err)
	}
	if _, err = pool.Exec(ctx, "UPDATE observation_campaigns SET actor='tampered' WHERE id=$1", campaign.ID); err == nil {
		t.Fatal("append-only campaign was mutable")
	}
	t.Run("individual planned municipality campaigns retain the same gates", func(t *testing.T) {
		for _, istat := range []string{"049009", "050026", "050029", "050008", "048017"} {
			id := "municipality-" + istat
			configuration := registry.Configuration{
				URL: "https://calcinaia.example/" + id, Sections: []string{"https://calcinaia.example/" + id}, AccessMethod: "fixture", Attribution: "synthetic municipality",
				Provenance: &evidence, Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}, CheckSeconds: 600, DelaySeconds: 1800,
			}
			for _, e := range []error{
				reg.CreateSource(ctx, registry.Source{ID: id, AuthorityID: "calcinaia", ChannelID: "calcinaia", ProductID: "municipal", Territory: istat}, configuration, "fixture"),
				reg.RecordPreview(ctx, id, 1, "fixture", evidence),
				reg.EnableCollection(ctx, id, 1, "fixture"),
			} {
				if e != nil {
					t.Fatal(e)
				}
			}
			request := StartRequest{ID: id, Scope: "municipality", Actor: "operator", StartedAt: started, SourceIDs: []string{id}}
			c, e := store.Start(ctx, request)
			if istat == "048017" {
				if !errors.Is(e, ErrInvalid) {
					t.Fatal("unplanned municipality admitted", e)
				}
				continue
			}
			if e != nil || c.Scope != "municipality" || len(c.Sources) != 1 || c.Sources[0].Territory != istat {
				t.Fatal("municipality scope lost", c, e)
			}
			report, e := store.Report(ctx, c.ID, through)
			if e != nil || report.Status != "extended" || len(report.MissingProducts) != 0 || !slicesContains(report.Sources[0].Issues, "no_persisted_checks") || !slicesContains(report.Sources[0].Issues, "original_comparison_missing") {
				t.Fatal("missing municipal evidence passed or regional products required", report, e)
			}
			if _, e = pool.Exec(ctx, `INSERT INTO acquisition_checks(source_id,configuration,worker_id,started_at,finished_at,reachable,content_recognized,complete,listing_count,document_count,check_state,publication_state)
 SELECT $1,1,'paced-fixture',point,point+interval '1 second',true,true,true,1,1,'complete_unchanged','not_expected'
 FROM generate_series($2::timestamptz,$3::timestamptz-interval '1 second',interval '10 minutes') point`, id, started, through); e != nil {
				t.Fatal(e)
			}
			for _, review := range []Review{
				{ID: "attachment-scope", SourceID: id, Kind: "attachment_not_applicable", Status: "pass", Evidence: reviewed},
				{ID: "failure", SourceID: id, Kind: "failure_retained", Status: "pass", CaseID: "synthetic-http-429", Evidence: reviewed},
				{ID: "event", SourceID: id, Kind: "absent_event_retained", Status: "pass", CaseID: "synthetic-event", Evidence: reviewed},
			} {
				if _, e = store.RecordReview(ctx, c.ID, "fixture-reviewer", review); e != nil {
					t.Fatal(e)
				}
			}
			report, e = store.Report(ctx, c.ID, through)
			if e != nil || report.Status != "extended" || !slicesContains(report.Sources[0].Issues, "original_comparison_missing") {
				t.Fatal("elapsed time and retained controls replaced original comparison", report, e)
			}
			var document, version int64
			if e = pool.QueryRow(ctx, "INSERT INTO retained_documents(source_id,official_url) VALUES($1,$2) RETURNING id", id, configuration.URL).Scan(&document); e != nil {
				t.Fatal(e)
			}
			hash := fmt.Sprintf("%064s", istat)
			if e = pool.QueryRow(ctx, `INSERT INTO retained_versions(document_id,content_hash,first_acquired_at,complete,metadata)
 VALUES($1,$2,$3,true,'{}') RETURNING id`, document, hash, started.Add(time.Hour)).Scan(&version); e != nil {
				t.Fatal(e)
			}
			if _, e = pool.Exec(ctx, "INSERT INTO retained_objects(hash,object_key,byte_size) VALUES($1,$2,10)", hash, "objects/"+hash); e != nil {
				t.Fatal(e)
			}
			for _, role := range []string{"original", "attachment"} {
				if _, e = pool.Exec(ctx, `INSERT INTO retained_resources(version_id,url,role,required,source_id,configuration,media_type,object_hash,missing)
 VALUES($1,$2,$3,true,$4,1,'application/pdf',$5,'')`, version, configuration.URL+"/"+role, role, id, hash); e != nil {
					t.Fatal(e)
				}
			}
			if _, e = store.RecordReview(ctx, c.ID, "fixture-reviewer", Review{ID: "original", SourceID: id, Kind: "original_comparison", Status: "pass", VersionID: &version, Evidence: reviewed}); e != nil {
				t.Fatal(e)
			}
			report, e = store.Report(ctx, c.ID, through)
			if e != nil || report.Status != "extended" || !slicesContains(report.Sources[0].Issues, "attachment_comparison_missing") {
				t.Fatal("attachment-not-applicable bypassed a required scanned attachment", report, e)
			}
			if _, e = store.RecordReview(ctx, c.ID, "fixture-reviewer", Review{ID: "attachment", SourceID: id, Kind: "attachment_comparison", Status: "pass", VersionID: &version, Evidence: reviewed}); e != nil {
				t.Fatal(e)
			}
			report, e = store.Report(ctx, c.ID, started.Add(6*24*time.Hour))
			if e != nil || report.Status != "running" || !slicesContains(report.Issues, "minimum_duration_not_reached") {
				t.Fatal("municipal evidence bypassed seven complete days", report, e)
			}
			assessment, e := store.Assess(ctx, c.ID, "fixture-reviewer", through, reviewed)
			if e != nil || assessment.Status != "complete" || len(assessment.Report.Sources) != 1 || len(assessment.Report.MissingProducts) != 0 {
				t.Fatal("complete individual municipal evidence failed", assessment, e)
			}
			request.ID, request.Scope = id+"-implicit", ""
			if _, e = store.Start(ctx, request); !errors.Is(e, ErrInvalid) {
				t.Fatal("additional municipality weakened implicit MVP scope", e)
			}
			st, e := reg.State(ctx, id)
			if e != nil || st.Accepted || st.PublicEnabled {
				t.Fatal("campaign granted acceptance or public activation", st, e)
			}
		}
		for _, request := range []StartRequest{
			{ID: "regional-in-municipality", Scope: "municipality", Actor: "operator", StartedAt: started, SourceIDs: []string{"vigilance"}},
			{ID: "multiple-in-municipality", Scope: "municipality", Actor: "operator", StartedAt: started, SourceIDs: []string{"municipality-049009", "municipality-050026"}},
			{ID: "invalid-scope", Scope: "invented", Actor: "operator", StartedAt: started, SourceIDs: []string{"calcinaia"}},
		} {
			if _, e := store.Start(ctx, request); !errors.Is(e, ErrInvalid) {
				t.Fatal("unsupported campaign scope admitted", request.Scope, e)
			}
		}
	})
	t.Run("review unchanged original under a later acquisition configuration", func(t *testing.T) {
		v, err := reg.Version(ctx, "monitoring", 1)
		if err != nil {
			t.Fatal(err)
		}
		revision, err := reg.AppendConfiguration(ctx, "monitoring", 1, v.Configuration, "test")
		if err != nil {
			t.Fatal(err)
		}
		if err = reg.RecordPreview(ctx, "monitoring", revision, "test", evidence); err != nil {
			t.Fatal(err)
		}
		if err = reg.EnableCollection(ctx, "monitoring", revision, "test"); err != nil {
			t.Fatal(err)
		}
		c, err := store.Start(ctx, StartRequest{ID: "reused-original", Actor: "operator", StartedAt: started, SourceIDs: []string{"monitoring"}})
		if err != nil {
			t.Fatal(err)
		}
		version := versions["monitoring"]
		review := Review{ID: "reused", SourceID: "monitoring", Kind: "original_comparison", Status: "pass", VersionID: &version, Evidence: reviewed}
		if _, err = store.RecordReview(ctx, c.ID, "delegated-assistant", review); !errors.Is(err, ErrInvalid) {
			t.Fatalf("unproven configuration reuse accepted: %v", err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO retained_acquisitions(id,document_id,source_id,configuration,request_hash,content_hash,acquired_at,version_id)
 SELECT 'reused-fixture',document_id,'monitoring',$2,content_hash,content_hash,$3,id FROM retained_versions WHERE id=$1`, version, revision, reviewed.ObservedAt.Add(time.Minute)); err != nil {
			t.Fatal(err)
		}
		if _, err = store.RecordReview(ctx, c.ID, "delegated-assistant", review); !errors.Is(err, ErrInvalid) {
			t.Fatalf("future acquisition proved earlier review: %v", err)
		}
		if _, err = pool.Exec(ctx, "UPDATE retained_acquisitions SET acquired_at=$1 WHERE id='reused-fixture'", reviewed.ObservedAt.Add(-time.Minute)); err != nil {
			t.Fatal(err)
		}
		if _, err = store.RecordReview(ctx, c.ID, "delegated-assistant", review); err != nil {
			t.Fatalf("finalized matching acquisition was rejected: %v", err)
		}
	})
}

func TestScopeMigrationPreservesLegacyCampaigns(t *testing.T) {
	ctx := context.Background()
	pool := observationTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, acquisition.Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		t.Fatal(err)
	}
	checksum := sha256.Sum256([]byte(schema))
	if _, err := pool.Exec(ctx, "INSERT INTO iwa_migrations(name,checksum) VALUES('030_observational_trials',$1)", hex.EncodeToString(checksum[:])); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO observation_campaigns(id,actor,started_at) VALUES('legacy','operator',clock_timestamp())"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := Migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	campaign, err := New(pool).Get(ctx, "legacy")
	if err != nil || campaign.Scope != "mvp" {
		t.Fatal("legacy campaign reinterpreted or upgrade failed", campaign, err)
	}
	if _, err := pool.Exec(ctx, "UPDATE observation_campaigns SET scope='municipality' WHERE id='legacy'"); err == nil {
		t.Fatal("immutable legacy campaign scope could be changed")
	}
}

func slicesContains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func observationTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	name := "iwa-observation-test-" + hex.EncodeToString(random[:6])
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
