//go:build integration

package registry

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
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Own an isolated PostgreSQL container. No fixture is written to the Compose DB.
func testDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	name := "iwa-registry-test-" + hex.EncodeToString(b[:6])
	password := hex.EncodeToString(b)
	file := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(file, []byte(password), 0644); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", args...).Output()
		if err != nil {
			t.Fatalf("isolated Docker operation failed: %v", err)
		}
		return strings.TrimSpace(string(out))
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	binding := listener.Addr().String() + ":5432"
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	run("run", "--pull=never", "--detach", "--name", name, "--publish", binding, "--mount", "type=bind,src="+file+",dst=/run/secrets/password,readonly", "--env", "POSTGRES_PASSWORD_FILE=/run/secrets/password", "--env", "POSTGRES_USER=iwa", "--env", "POSTGRES_DB=iwa", "postgres:16-alpine")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", name).Run() })
	addr := run("port", name, "5432/tcp")
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatal(err)
	}
	n, _ := strconv.Atoi(port)
	cfg, err := pgxpool.ParseConfig("postgres://iwa@localhost/iwa?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.Host = host
	cfg.ConnConfig.Port = uint16(n)
	cfg.ConnConfig.Password = password
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	wait := func() {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		for {
			pingErr := pool.Ping(ctx)
			if pingErr == nil {
				return
			}
			if ctx.Err() != nil {
				out, _ := exec.Command("docker", "logs", "--tail", "20", name).CombinedOutput()
				t.Fatalf("test PostgreSQL did not become ready: %s; %s", strings.ReplaceAll(pingErr.Error(), password, "[redacted]"), strings.ReplaceAll(string(out), password, "[redacted]"))
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	wait()
	return pool, func() { run("restart", "--time", "5", name); wait() }
}
func fixtureConfig() Configuration {
	now := time.Now().UTC()
	e := Evidence{URL: "https://comune.example/notizie", Locator: "paragraph 1", ObservedAt: now}
	return Configuration{URL: "https://platform.example/calcinaia", Sections: []string{"https://platform.example/calcinaia/notizie"}, AccessMethod: "html", Attribution: "Comune, published by associated service", Provenance: &e, Policy: Policy{Evidence: &e, CollectionPermitted: true, RetentionPermitted: true, PublicationPermitted: true}, Limitations: []string{"Only configured notices; not all municipal publications"}}
}
func referral(c Configuration) *Referral {
	return &Referral{Evidence: Evidence{URL: "https://comune.example/protezione-civile", Locator: "official outbound link", ObservedAt: time.Now().UTC()}, Destination: c.URL, Context: "Municipal notices", Sections: c.Sections, ProductID: "municipal", Territory: "050004"}
}
func acceptance(c Configuration) Acceptance {
	now := time.Now().UTC()
	return Acceptance{Report: Evidence{URL: "https://evidence.example/trial", Locator: "reviewed report", ObservedAt: now}, PeriodStart: now.Add(-8 * 24 * time.Hour), PeriodEnd: now.Add(-time.Hour), Sections: c.Sections, ExtractionVerified: true, UpdatesVerified: true, AttachmentsVerified: true, ScannedAttachmentsVerified: true, HistoryVerified: true, FailureBehaviorVerified: true, InterfacesEquivalent: true, CoverageStatus: "accepted_with_limitations", CoverageLimitations: append([]string(nil), c.Limitations...)}
}

func reviewAcceptance(t *testing.T, s *Store, ctx context.Context, id string, revision int) AcceptanceReview {
	t.Helper()
	review, err := s.ReviewAcceptance(ctx, id, revision, "release-reviewer", Evidence{URL: "https://evidence.example/release", Locator: "final source-scoped acceptance review", ObservedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	if review.Status != "accepted" {
		t.Fatalf("release review pending: %#v", review)
	}
	return review
}
func TestRegistryPostgres(t *testing.T) {
	ctx := context.Background()
	pool, restart := testDB(t)
	s := New(pool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	wantErr := func(err, want error) {
		t.Helper()
		if !errors.Is(err, want) {
			t.Fatalf("got %v, want %v", err, want)
		}
	}
	must(Migrate(ctx, pool))
	must(Migrate(ctx, pool))
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- Migrate(ctx, pool) }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		must(err)
	}
	must(s.CreateAuthority(ctx, Authority{"municipality", "Comune", "https://comune.example"}))
	must(s.CreateAuthority(ctx, Authority{"union", "Associated publisher", "https://union.example"}))
	must(s.CreateChannel(ctx, Channel{"external", "union", "Cittadino platform", "https://platform.example", true}))
	source := Source{"calcinaia", "municipality", "external", "municipal", "050004"}
	c := fixtureConfig()
	must(s.CreateSource(ctx, source, c, "operator"))
	st, err := s.State(ctx, source.ID)
	must(err)
	if st.LatestRevision != 1 || st.ActiveRevision != nil || st.CollectionEnabled || st.PublicEnabled || st.Accepted {
		t.Fatal("new source not pending")
	}
	ch, err := s.Channel(ctx, st.Source.ChannelID)
	must(err)
	if ch.PublisherID == st.Source.AuthorityID || ch.Platform != "Cittadino platform" {
		t.Fatal("issuer/publisher/platform conflated")
	}
	a, err := s.Authority(ctx, st.Source.AuthorityID)
	must(err)
	if a.Name != "Comune" {
		t.Fatal("issuer lost")
	}
	v, err := s.Version(ctx, source.ID, 1)
	must(err)
	if v.Configuration.CheckSeconds != 600 || v.Configuration.DelaySeconds != 1800 {
		t.Fatal("defaults missing")
	}
	wantErr(s.EnableCollection(ctx, source.ID, 1, "operator"), ErrPrerequisite)
	wantErr(s.EnablePublic(ctx, source.ID, 1, "operator"), ErrConflict)
	preview := Evidence{URL: "https://evidence.example/preview", Locator: "successful preview", ObservedAt: time.Now().UTC()}
	must(s.RecordPreview(ctx, source.ID, 1, "operator", preview))
	must(s.EnableCollection(ctx, source.ID, 1, "operator"))
	st, err = s.State(ctx, source.ID)
	must(err)
	if !st.CollectionEnabled || st.PublicEnabled {
		t.Fatal("collection enabled public access")
	}
	wantErr(s.Accept(ctx, source.ID, 1, "reviewer", acceptance(c)), ErrPrerequisite)
	// A plan-maps referral cannot recognize municipal notices.
	c.Referral = referral(c)
	c.Referral.ProductID = "plan_maps"
	rev, err := s.AppendConfiguration(ctx, source.ID, 1, c, "operator")
	must(err)
	st, err = s.State(ctx, source.ID)
	must(err)
	if *st.ActiveRevision != 1 || rev != 2 {
		t.Fatal("draft replaced active revision")
	}
	must(s.RecordPreview(ctx, source.ID, 2, "operator", preview))
	must(s.EnableCollection(ctx, source.ID, 2, "operator"))
	wantErr(s.Accept(ctx, source.ID, 2, "reviewer", acceptance(c)), ErrPrerequisite)
	c.Referral = referral(c)
	rev, err = s.AppendConfiguration(ctx, source.ID, 2, c, "operator")
	must(err)
	must(s.RecordPreview(ctx, source.ID, rev, "operator", preview))
	must(s.EnableCollection(ctx, source.ID, rev, "operator"))
	bad := acceptance(c)
	bad.KnownOmissions = 1
	wantErr(s.Accept(ctx, source.ID, rev, "reviewer", bad), ErrInvalid)
	bad = acceptance(c)
	bad.PeriodStart = bad.PeriodEnd.Add(-time.Hour)
	wantErr(s.Accept(ctx, source.ID, rev, "reviewer", bad), ErrInvalid)
	bad = acceptance(c)
	bad.Sections = []string{"https://platform.example/other"}
	wantErr(s.Accept(ctx, source.ID, rev, "reviewer", bad), ErrPrerequisite)
	must(s.Accept(ctx, source.ID, rev, "reviewer", acceptance(c)))
	st, err = s.State(ctx, source.ID)
	must(err)
	if !st.Accepted || st.PublicEnabled {
		t.Fatal("acceptance enabled publication")
	}
	wantErr(s.EnablePublic(ctx, source.ID, rev, "operator"), ErrPrerequisite)
	review := reviewAcceptance(t, s, ctx, source.ID, rev)
	readiness, err := s.ReleaseReadiness(ctx)
	must(err)
	if readiness.Status != "partial" || len(readiness.Scopes) != 4 || readiness.Scopes[3].Status != "accepted_with_limitations" {
		t.Fatalf("partial readiness overstated: %#v", readiness)
	}
	must(s.EnablePublic(ctx, source.ID, rev, "operator"))
	if _, err = pool.Exec(ctx, "UPDATE registry_acceptance_reviews SET status='pending' WHERE id=$1", review.ID); err == nil {
		t.Fatal("acceptance review history mutable")
	}
	must(s.Disable(ctx, source.ID, rev, "operator", false))
	st, err = s.State(ctx, source.ID)
	must(err)
	if st.CollectionEnabled || !st.PublicEnabled {
		t.Fatal("collection stop erased public status")
	}
	must(s.EnableCollection(ctx, source.ID, rev, "operator"))
	// New revisions cannot inherit preview/acceptance. Concurrent writes use CAS.
	original := c
	c.URL = "https://platform.example/changed"
	c.Referral = nil
	concurrent := make(chan error, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := s.AppendConfiguration(ctx, source.ID, rev, c, "editor")
			concurrent <- e
		}()
	}
	wg.Wait()
	close(concurrent)
	successes := 0
	for e := range concurrent {
		if e == nil {
			successes++
		} else {
			wantErr(e, ErrConflict)
		}
	}
	if successes != 1 {
		t.Fatal("concurrent revision overwrite")
	}
	st, err = s.State(ctx, source.ID)
	must(err)
	if *st.ActiveRevision != rev || !st.PublicEnabled {
		t.Fatal("draft changed published config")
	}
	wantErr(s.EnableCollection(ctx, source.ID, rev+1, "operator"), ErrPrerequisite)
	must(s.RecordPreview(ctx, source.ID, rev+1, "operator", preview))
	must(s.EnableCollection(ctx, source.ID, rev+1, "operator"))
	st, err = s.State(ctx, source.ID)
	must(err)
	if st.Accepted || st.PublicEnabled {
		t.Fatal("old acceptance leaked into new configuration")
	}
	wantErr(s.EnablePublic(ctx, source.ID, rev+1, "operator"), ErrPrerequisite)
	old, err := s.Version(ctx, source.ID, rev)
	must(err)
	if old.Configuration.URL != original.URL || old.Configuration.Referral == nil {
		t.Fatal("old configuration overwritten")
	}
	if _, err = pool.Exec(ctx, "UPDATE registry_configurations SET actor='changed'"); err == nil {
		t.Fatal("history mutable")
	}
	if _, err = pool.Exec(ctx, "DELETE FROM registry_events"); err == nil {
		t.Fatal("events mutable")
	}
	// Other products sharing a channel do not inherit permissions or acceptance.
	monitoring := Source{"monitoring", "municipality", "external", "monitoring", "050004"}
	pending := fixtureConfig()
	pending.Policy = Policy{}
	must(s.CreateSource(ctx, monitoring, pending, "operator"))
	wantErr(s.RecordPreview(ctx, "monitoring", 1, "operator", preview), ErrPrerequisite)
	st, err = s.State(ctx, "monitoring")
	must(err)
	if st.Accepted || st.PublicEnabled || st.CollectionEnabled {
		t.Fatal("product state inherited")
	}
	// DPC comparison stays internal even after a valid trial record.
	must(s.CreateChannel(ctx, Channel{"direct", "municipality", "official site", "https://comune.example", false}))
	direct := fixtureConfig()
	direct.URL = "https://comune.example/bollettini"
	direct.Sections = []string{direct.URL}
	direct.Referral = nil
	must(s.CreateSource(ctx, Source{"dpc", "municipality", "direct", "dpc_comparison", "Toscana"}, direct, "operator"))
	must(s.RecordPreview(ctx, "dpc", 1, "operator", preview))
	must(s.EnableCollection(ctx, "dpc", 1, "operator"))
	must(s.Accept(ctx, "dpc", 1, "reviewer", acceptance(direct)))
	wantErr(s.EnablePublic(ctx, "dpc", 1, "operator"), ErrPrerequisite)

	// External recognition must cover the exact territory, sections and referral origin.
	for _, mismatch := range []string{"territory", "sections", "origin", "unresolved"} {
		cfg := fixtureConfig()
		cfg.Referral = referral(cfg)
		switch mismatch {
		case "territory":
			cfg.Referral.Territory = "other"
		case "sections":
			cfg.Referral.Sections = []string{"https://platform.example/maps"}
		case "origin":
			cfg.Referral.Evidence.URL = "https://unrecognized.example/link"
		case "unresolved":
			cfg.Unresolved = []string{"reuse applicability unresolved"}
		}
		id := "scope-" + mismatch
		must(s.CreateSource(ctx, Source{id, "municipality", "external", "municipal", "050004"}, cfg, "operator"))
		must(s.RecordPreview(ctx, id, 1, "operator", preview))
		must(s.EnableCollection(ctx, id, 1, "operator"))
		wantErr(s.Accept(ctx, id, 1, "reviewer", acceptance(cfg)), ErrPrerequisite)
	}
	// Failed writes cannot leave an identity without its first configuration.
	if err := s.CreateSource(ctx, Source{"broken", "missing-authority", "external", "municipal", "050004"}, fixtureConfig(), "operator"); err == nil {
		t.Fatal("missing authority accepted")
	}
	_, err = s.State(ctx, "broken")
	wantErr(err, ErrNotFound)
	var checksum string
	must(pool.QueryRow(ctx, "SELECT checksum FROM iwa_migrations WHERE name='001_registry'").Scan(&checksum))
	_, err = pool.Exec(ctx, "UPDATE iwa_migrations SET checksum='changed' WHERE name='001_registry'")
	must(err)
	if err = Migrate(ctx, pool); err == nil {
		t.Fatal("changed migration silently accepted")
	}
	_, err = pool.Exec(ctx, "UPDATE iwa_migrations SET checksum=$1 WHERE name='001_registry'", checksum)
	must(err)
	events, err := s.Events(ctx, source.ID)
	must(err)
	if len(events) < 10 {
		t.Fatal("audit history lost")
	}
	must(s.SetIntervals(ctx, source.ID, 0, 120, 900, "interval-operator"))
	wantErr(s.SetIntervals(ctx, source.ID, 0, 20, 60, "stale-operator"), ErrConflict)
	wantErr(s.SetIntervals(ctx, source.ID, 1, 0, 60, "operator"), ErrInvalid)
	events, err = s.Events(ctx, source.ID)
	must(err)
	restart()
	s = New(pool)
	must(Migrate(ctx, pool))
	st, err = s.State(ctx, source.ID)
	must(err)
	if *st.ActiveRevision != rev+1 || !st.CollectionEnabled || st.PublicEnabled {
		t.Fatal("restart lost state")
	}
	after, err := s.Events(ctx, source.ID)
	must(err)
	if len(after) != len(events) {
		t.Fatal("restart lost events")
	}
	intervals, err := s.Intervals(ctx, source.ID)
	must(err)
	if intervals.Revision != 1 || *intervals.CheckSeconds != 120 || *intervals.DelaySeconds != 900 {
		t.Fatal("database restart lost immediate interval overrides")
	}
	t.Log("real PostgreSQL: migration replay/concurrency, pending scopes, referral bounds, immutable versions, separate activation, CAS and restart persistence passed")
}

// Synthetic release controls: no acceptance or activation of a live source.
func TestFourMVPScopesEnableIndependentlyAndKeepPartialCoverage(t *testing.T) {
	ctx := context.Background()
	pool, _ := testDB(t)
	s := New(pool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(Migrate(ctx, pool))
	must(s.CreateAuthority(ctx, Authority{"synthetic-mvp", "Synthetic authority", "https://comune.example"}))
	must(s.CreateChannel(ctx, Channel{"synthetic-direct", "synthetic-mvp", "direct", "https://comune.example", false}))
	products := []string{"vigilance", "criticality", "monitoring", "municipal"}
	sources := []string{"synthetic-vigilance", "synthetic-criticality", "synthetic-monitoring", "synthetic-calcinaia"}
	configs := make([]Configuration, 4)
	for i, product := range products {
		c := fixtureConfig()
		c.URL = "https://comune.example/" + product
		c.Sections = []string{c.URL}
		c.Limitations = nil
		configs[i] = c
		territory := "Toscana"
		if product == "municipal" {
			territory = "050004"
		}
		must(s.CreateSource(ctx, Source{sources[i], "synthetic-mvp", "synthetic-direct", product, territory}, c, "synthetic-operator"))
		must(s.RecordPreview(ctx, sources[i], 1, "synthetic-operator", Evidence{URL: c.URL, Locator: "synthetic successful preview", ObservedAt: time.Now().UTC()}))
		must(s.EnableCollection(ctx, sources[i], 1, "synthetic-operator"))
		if err := s.EnablePublic(ctx, sources[i], 1, "synthetic-operator"); !errors.Is(err, ErrPrerequisite) {
			t.Fatalf("unaccepted scope enabled: %s %v", product, err)
		}
	}
	readiness, err := s.ReleaseReadiness(ctx)
	must(err)
	if readiness.Status != "pending" {
		t.Fatal(readiness)
	}
	for i, id := range sources {
		a := acceptance(configs[i])
		a.CoverageStatus = "accepted_declared_scope"
		a.CoverageLimitations = nil
		if products[i] != "municipal" {
			a.RiskCoverage = append([]string(nil), ToscanaRisks...)
		}
		if i == 0 {
			// Missing risk/scanning/history/failure evidence cannot be called accepted.
			for _, field := range []string{"risk", "scan", "history", "failure", "interfaces"} {
				bad := a
				switch field {
				case "risk":
					bad.RiskCoverage = bad.RiskCoverage[:6]
				case "scan":
					bad.ScannedAttachmentsVerified = false
				case "history":
					bad.HistoryVerified = false
				case "failure":
					bad.FailureBehaviorVerified = false
				case "interfaces":
					bad.InterfacesEquivalent = false
				}
				if err := s.Accept(ctx, id, 1, "synthetic-reviewer", bad); !errors.Is(err, ErrInvalid) {
					t.Fatalf("missing %s accepted: %v", field, err)
				}
			}
		}
		if i == 3 {
			a.CoverageStatus = "accepted_with_limitations"
			a.CoverageLimitations = []string{"Only the declared synthetic municipal sections"}
		}
		must(s.Accept(ctx, id, 1, "synthetic-reviewer", a))
		if err := s.EnablePublic(ctx, id, 1, "synthetic-operator"); !errors.Is(err, ErrPrerequisite) {
			t.Fatalf("unreviewed acceptance enabled: %v", err)
		}
		reviewAcceptance(t, s, ctx, id, 1)
		st, err := s.State(ctx, id)
		must(err)
		if st.PublicEnabled {
			t.Fatal("review enabled source implicitly", id)
		}
		must(s.EnablePublic(ctx, id, 1, "synthetic-operator"))
		for j, other := range sources {
			st, err := s.State(ctx, other)
			must(err)
			if st.PublicEnabled != (j <= i) {
				t.Fatalf("independent activation changed %s: %#v", other, st)
			}
		}
		readiness, err = s.ReleaseReadiness(ctx)
		must(err)
		if readiness.Status != "partial" || len(readiness.Scopes) != 4 {
			t.Fatal("incomplete or limited MVP reported ready", readiness)
		}
	}
	// All four accepted does not erase the municipal coverage limitation.
	if readiness.Scopes[3].Status != "accepted_with_limitations" {
		t.Fatal(readiness)
	}
	full := acceptance(configs[3])
	full.CoverageStatus = "accepted_declared_scope"
	full.CoverageLimitations = nil
	must(s.Accept(ctx, sources[3], 1, "synthetic-reviewer", full))
	reviewAcceptance(t, s, ctx, sources[3], 1)
	readiness, err = s.ReleaseReadiness(ctx)
	must(err)
	if readiness.Status != "ready" {
		t.Fatal("four fully reviewed scopes not ready", readiness)
	}
	// Suspension is source-scoped and cannot erase collection or other publications.
	must(s.Disable(ctx, sources[1], 1, "synthetic-operator", true))
	for i, id := range sources {
		st, err := s.State(ctx, id)
		must(err)
		if !st.CollectionEnabled || st.PublicEnabled != (i != 1) {
			t.Fatalf("source-scoped suspension changed unrelated state: %#v", st)
		}
	}
}
