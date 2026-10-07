//go:build integration

package acquisition

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

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func acquisitionTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	name := "iwa-acquisition-test-" + hex.EncodeToString(random[:6])
	password := hex.EncodeToString(random)
	passwordFile := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		output, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
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
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	for pool.Ping(ctx) != nil {
		if ctx.Err() != nil {
			t.Fatal("test PostgreSQL did not become ready")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return pool
}

func TestScheduleStatusBackoffRetryAfterAndCompleteTime(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := acquisitionTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	reg := registry.New(pool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	base := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "municipality", Name: "Municipality", OfficialURL: "https://source.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "channel", PublisherID: "municipality", Platform: "fixture", URL: "https://source.example"}))
	evidence := registry.Evidence{URL: "https://source.example/reuse", Locator: "fixture permission", ObservedAt: base}
	configuration := registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/notices"}, AccessMethod: "crawl4ai", Attribution: "Municipality", CheckSeconds: 600, DelaySeconds: 1800, BackoffBaseSeconds: 60, BackoffMaxSeconds: 300, Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}
	must(reg.CreateSource(ctx, registry.Source{ID: "source", AuthorityID: "municipality", ChannelID: "channel", ProductID: "municipal", Territory: "050004"}, configuration, "test"))
	must(reg.RecordPreview(ctx, "source", 1, "test", evidence))
	must(reg.EnableCollection(ctx, "source", 1, "test"))
	store := NewScheduleStore(pool)
	must(store.SyncEnabled(ctx, base))

	claim1, err := store.ClaimDue(ctx, "worker-1", base, time.Minute)
	must(err)
	invalidFinished := base.Add(time.Second)
	must(store.Finish(ctx, claim1, CheckOutcome{SourceID: "source", Configuration: 1, StartedAt: claim1.StartedAt, FinishedAt: invalidFinished, Reachable: true, ErrorCode: "unrecognized_content"}))
	status, err := store.Status(ctx, "source", invalidFinished)
	must(err)
	if status.UpdatingState != "not_yet_verified" || status.ConsecutiveFailures != 1 || status.LastReachableAt == nil || status.LastContentAt != nil || status.LastCompleteAt != nil || status.FirstErrorAt == nil || status.LastErrorCode != "unrecognized_content" || !status.NextCheckAt.Equal(invalidFinished.Add(time.Minute)) {
		t.Fatalf("first error not visible or timestamps conflated: %#v", status)
	}
	if _, err = store.ClaimDue(ctx, "worker-early", invalidFinished.Add(30*time.Second), time.Minute); !errors.Is(err, ErrNoDueCheck) {
		t.Fatalf("source backoff not respected: %v", err)
	}

	claim2, err := store.ClaimDue(ctx, "worker-2", status.NextCheckAt, time.Minute)
	must(err)
	rateFinished := status.NextCheckAt.Add(time.Second)
	retryAfter := rateFinished.Add(20 * time.Minute)
	must(store.Finish(ctx, claim2, CheckOutcome{SourceID: "source", Configuration: 1, StartedAt: claim2.StartedAt, FinishedAt: rateFinished, Reachable: true, ErrorCode: "source_rate_limited", RetryAfter: &retryAfter}))
	status, _ = store.Status(ctx, "source", rateFinished)
	if status.ConsecutiveFailures != 2 || !status.NextCheckAt.Equal(retryAfter) || status.RetryAfter == nil || !status.RetryAfter.Equal(retryAfter) {
		t.Fatalf("Retry-After did not override bounded backoff: %#v", status)
	}

	claim3, err := store.ClaimDue(ctx, "worker-3", retryAfter, time.Minute)
	must(err)
	completeAt := retryAfter.Add(time.Second)
	must(store.Finish(ctx, claim3, CheckOutcome{SourceID: "source", Configuration: 1, StartedAt: claim3.StartedAt, FinishedAt: completeAt, Reachable: true, ContentRecognized: true, Complete: true, Listings: 3, Documents: 12}))
	status, _ = store.Status(ctx, "source", completeAt.Add(29*time.Minute))
	if status.UpdatingState != "current" || status.ConsecutiveFailures != 0 || status.LastCompleteAt == nil || !status.LastCompleteAt.Equal(completeAt) || status.FirstErrorAt != nil || status.LastErrorCode != "" || !status.NextCheckAt.Equal(completeAt.Add(10*time.Minute)) {
		t.Fatalf("successful complete check state incorrect: %#v", status)
	}
	status, _ = store.Status(ctx, "source", completeAt.Add(30*time.Minute))
	if status.UpdatingState != "delayed" {
		t.Fatalf("independent delay threshold not applied: %#v", status)
	}

	claim4, err := store.ClaimDue(ctx, "worker-4", completeAt.Add(10*time.Minute), time.Minute)
	must(err)
	incompleteAt := completeAt.Add(10*time.Minute + time.Second)
	must(store.Finish(ctx, claim4, CheckOutcome{SourceID: "source", Configuration: 1, StartedAt: claim4.StartedAt, FinishedAt: incompleteAt, Reachable: true, ContentRecognized: true, ErrorCode: "incomplete_scope"}))
	status, _ = store.Status(ctx, "source", incompleteAt)
	if status.LastCompleteAt == nil || !status.LastCompleteAt.Equal(completeAt) || status.LastContentAt == nil || !status.LastContentAt.Equal(incompleteAt) || status.LastErrorCode != "incomplete_scope" {
		t.Fatalf("incomplete check advanced complete time or lost content status: %#v", status)
	}
	checks, err := store.Checks(ctx, "source")
	must(err)
	if len(checks) != 4 || checks[0].Complete || checks[1].ErrorCode != "source_rate_limited" || !checks[2].Complete || checks[3].Complete {
		t.Fatalf("check history incomplete: %#v", checks)
	}
}

func TestExpiredLeaseBecomesVisibleFailure(t *testing.T) {
	ctx := context.Background()
	pool := acquisitionTestDB(t)
	if err := registry.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	reg := registry.New(pool)
	base := time.Date(2026, 9, 16, 14, 0, 0, 0, time.UTC)
	evidence := registry.Evidence{URL: "https://source.example/reuse", Locator: "fixture", ObservedAt: base}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "A", OfficialURL: "https://source.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "fixture", URL: "https://source.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "s", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "x"}, registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/list"}, AccessMethod: "html", Attribution: "A", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true}}, "test"),
		reg.RecordPreview(ctx, "s", 1, "test", evidence), reg.EnableCollection(ctx, "s", 1, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	store := NewScheduleStore(pool)
	if err := store.SyncEnabled(ctx, base); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ClaimDue(ctx, "crashed-worker", base, 3*time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ClaimDue(ctx, "recovery-worker", base.Add(4*time.Second), time.Minute); !errors.Is(err, ErrNoDueCheck) {
		t.Fatalf("expired check should enter backoff, got %v", err)
	}
	checks, err := store.Checks(ctx, "s")
	if err != nil || len(checks) != 1 || checks[0].ErrorCode != "check_lease_expired" {
		t.Fatalf("expired lease not persisted: checks=%#v err=%v", checks, err)
	}
	// Replay the observed saturated municipal schedule: a lost 20-minute
	// lease plus its 60-minute backoff explains the 80-minute cadence.
	if _, err := pool.Exec(ctx, "UPDATE acquisition_source_status SET consecutive_failures=76,next_check_at=$1,backoff_max_seconds=3600 WHERE source_id='s'", base); err != nil {
		t.Fatal(err)
	}
	claim, err := store.ClaimDue(ctx, "cycle-worker", base, 20*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.RecoverExpired(ctx, base.Add(20*time.Minute)); err != nil {
		t.Fatal(err)
	}
	status, err := store.Status(ctx, "s", base.Add(20*time.Minute))
	if err != nil || !status.NextCheckAt.Equal(base.Add(80*time.Minute)) {
		t.Fatalf("80-minute cycle not reproduced: %#v %v", status, err)
	}
	if err = store.Heartbeat(ctx, claim, base.Add(20*time.Minute), 20*time.Minute); !errors.Is(err, ErrStaleCheck) {
		t.Fatalf("lost ownership not fenced: %v", err)
	}

}

func TestExpectedPublicationStatesStayDistinct(t *testing.T) {
	ctx := context.Background()
	pool := acquisitionTestDB(t)
	if err := registry.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	reg := registry.New(pool)
	base := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	evidence := registry.Evidence{URL: "https://source.example/calendar", Locator: "documented daily publication", ObservedAt: base.Add(-24 * time.Hour)}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "publication-a", Name: "A", OfficialURL: "https://source.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "publication-c", PublisherID: "publication-a", Platform: "fixture", URL: "https://source.example"}),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	configuration := func(expected bool) registry.Configuration {
		cfg := registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/list"}, AccessMethod: "html", Attribution: "A", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true}}
		if expected {
			cfg.ExpectedPublication = &registry.ExpectedPublication{Evidence: evidence, Anchor: base.Add(-10 * time.Minute), IntervalSeconds: 86400, ToleranceSeconds: 60}
		}
		return cfg
	}
	for _, item := range []struct {
		id       string
		expected bool
	}{{"expected", true}, {"occasional", false}} {
		if err := reg.CreateSource(ctx, registry.Source{ID: item.id, AuthorityID: "publication-a", ChannelID: "publication-c", ProductID: "municipal", Territory: "x"}, configuration(item.expected), "test"); err != nil {
			t.Fatal(err)
		}
		if err := reg.RecordPreview(ctx, item.id, 1, "test", evidence); err != nil {
			t.Fatal(err)
		}
		if err := reg.EnableCollection(ctx, item.id, 1, "test"); err != nil {
			t.Fatal(err)
		}
	}
	store := NewScheduleStore(pool)
	if err := store.SyncEnabled(ctx, base); err != nil {
		t.Fatal(err)
	}

	expectedClaim, err := store.ClaimDue(ctx, "expected-worker", base, time.Minute)
	if err != nil || expectedClaim.SourceID != "expected" {
		t.Fatalf("claim expected source: claim=%#v err=%v", expectedClaim, err)
	}
	firstFinished := base.Add(time.Second)
	if err = store.Finish(ctx, expectedClaim, CheckOutcome{SourceID: "expected", Configuration: 1, StartedAt: expectedClaim.StartedAt, FinishedAt: firstFinished, Reachable: true, ContentRecognized: true, Complete: true}); err != nil {
		t.Fatal(err)
	}
	status, err := store.Status(ctx, "expected", firstFinished)
	if err != nil || status.LastCheckState != "complete_unchanged" || status.PublicationState != "missing" || status.LastErrorCode != "" {
		t.Fatalf("successful unchanged expected check conflated: status=%#v err=%v", status, err)
	}

	occasionalClaim, err := store.ClaimDue(ctx, "occasional-worker", base.Add(2*time.Second), time.Minute)
	if err != nil || occasionalClaim.SourceID != "occasional" {
		t.Fatalf("claim occasional source: claim=%#v err=%v", occasionalClaim, err)
	}
	if err = store.Finish(ctx, occasionalClaim, CheckOutcome{SourceID: "occasional", Configuration: 1, StartedAt: occasionalClaim.StartedAt, FinishedAt: base.Add(3 * time.Second), Reachable: true, ContentRecognized: true, Complete: true}); err != nil {
		t.Fatal(err)
	}
	status, err = store.Status(ctx, "occasional", base.Add(3*time.Second))
	if err != nil || status.LastCheckState != "complete_unchanged" || status.PublicationState != "not_expected" {
		t.Fatalf("occasional silence treated as missing publication: status=%#v err=%v", status, err)
	}
	if err = reg.Disable(ctx, "occasional", 1, "test", false); err != nil {
		t.Fatal(err)
	}

	changedClaim, err := store.ClaimDue(ctx, "changed-worker", firstFinished.Add(10*time.Minute), time.Minute)
	if err != nil || changedClaim.SourceID != "expected" {
		t.Fatalf("claim changed source: claim=%#v err=%v", changedClaim, err)
	}
	// An edition issued shortly before the expected instant still belongs to
	// this cycle, and its actual timestamp must remain unchanged in storage.
	publication := base.Add(-10*time.Minute - 30*time.Second)
	changedFinished := changedClaim.StartedAt.Add(time.Second)
	if err = store.Finish(ctx, changedClaim, CheckOutcome{SourceID: "expected", Configuration: 1, StartedAt: changedClaim.StartedAt, FinishedAt: changedFinished, Reachable: true, ContentRecognized: true, Complete: true, NewDocuments: 1, PublicationObservedAt: &publication}); err != nil {
		t.Fatal(err)
	}
	status, _ = store.Status(ctx, "expected", changedFinished)
	if status.LastCheckState != "complete_changed" || status.PublicationState != "observed" || status.LastPublicationAt == nil || !status.LastPublicationAt.Equal(publication) {
		t.Fatalf("observed expected publication not recorded: %#v", status)
	}

	failureClaim, err := store.ClaimDue(ctx, "failure-worker", changedFinished.Add(10*time.Minute), time.Minute)
	if err != nil || failureClaim.SourceID != "expected" {
		t.Fatalf("claim failing source: claim=%#v err=%v", failureClaim, err)
	}
	failureFinished := failureClaim.StartedAt.Add(time.Second)
	if err = store.Finish(ctx, failureClaim, CheckOutcome{SourceID: "expected", Configuration: 1, StartedAt: failureClaim.StartedAt, FinishedAt: failureFinished, ErrorCode: "collection_failed"}); err != nil {
		t.Fatal(err)
	}
	status, _ = store.Status(ctx, "expected", failureFinished)
	if status.LastCheckState != "collection_failure" || status.PublicationState != "observed" || status.LastErrorCode != "collection_failed" {
		t.Fatalf("collection failure conflated with publication state: %#v", status)
	}
	checks, err := store.Checks(ctx, "expected")
	if err != nil || len(checks) != 3 || checks[0].CheckState != "complete_unchanged" || checks[0].PublicationState != "missing" || checks[1].CheckState != "complete_changed" || checks[1].PublicationState != "observed" || checks[2].CheckState != "collection_failure" || checks[2].PublicationState != "observed" {
		t.Fatalf("distinct check history not preserved: checks=%#v err=%v", checks, err)
	}
}

func TestRegionalFirstCheckRegistersTrackingTarget(t *testing.T) {
	ctx := context.Background()
	pool := acquisitionTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	reg := registry.New(pool)
	now := time.Now().UTC()
	const target = "https://source.example/monitoring"
	evidence := registry.Evidence{URL: target, Locator: "fixture", ObservedAt: now}
	cfg := registry.Configuration{URL: target, Sections: []string{target}, AccessMethod: "crawl4ai-html-pdf", Attribution: "Fixture",
		Policy:          registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true},
		RegionalProduct: &registry.RegionalProductContract{Kind: "monitoring", ContentMarkers: []string{"Monitoraggio evento"}}}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "A", OfficialURL: target}),
		reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "fixture", URL: target}),
		reg.CreateSource(ctx, registry.Source{ID: "monitor", AuthorityID: "a", ChannelID: "c", ProductID: "monitoring", Territory: "Toscana"}, cfg, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	crawler := &fixtureCrawler{pages: map[string]Page{target: {URL: target, HTML: []byte(`<h2>Monitoraggio evento</h2><b>NESSUN AVVISO IN CORSO DI VALIDITÀ O EVENTO IN CORSO</b>`), StatusCode: 200}}}
	engine := Engine{Registry: reg, Retained: &stableRetention{}, Crawler: crawler, Resources: crawler, Tracking: NewTrackingStore(pool)}
	first := engine.Check(ctx, "monitor", 1, now)
	if !first.Complete || first.NewDocuments != 1 || first.Revisions != 0 {
		t.Fatalf("initial regional tracking failed: %#v", first)
	}
	second := engine.Check(ctx, "monitor", 1, time.Now().UTC())
	if !second.Complete || second.NewDocuments != 0 || second.Revisions != 0 {
		t.Fatalf("unchanged regional check failed: %#v", second)
	}
	crawler.pages[target] = Page{URL: target, HTML: []byte(`<h2>Monitoraggio evento</h2><p>Emesso il 18/09/2026 12.00</p>`), StatusCode: 200}
	third := engine.Check(ctx, "monitor", 1, time.Now().UTC())
	if !third.Complete || third.NewDocuments != 0 || third.Revisions != 1 {
		t.Fatalf("regional transition lost: %#v", third)
	}
}

func TestRevisionTrackingBootstrapAndRecurringSelection(t *testing.T) {
	ctx := context.Background()
	pool := acquisitionTestDB(t)
	if err := registry.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	reg := registry.New(pool)
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	evidence := registry.Evidence{URL: "https://source.example/reuse", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "track-a", Name: "A", OfficialURL: "https://source.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "track-c", PublisherID: "track-a", Platform: "fixture", URL: "https://source.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "track-s", AuthorityID: "track-a", ChannelID: "track-c", ProductID: "municipal", Territory: "x"}, registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/list"}, AccessMethod: "html", Attribution: "A", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true}}, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	date := func(year int, month time.Month, day int) *time.Time {
		value := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		return &value
	}
	tracker := NewTrackingStore(pool)
	documents := []DiscoveredDocument{
		{URL: "https://source.example/recent-irrelevant", PublicationDate: date(2026, 9, 1)},
		{URL: "https://source.example/old-ongoing", PublicationDate: date(2026, 6, 1)},
		{URL: "https://source.example/old-irrelevant", PublicationDate: date(2026, 6, 2)},
		{URL: "https://source.example/unknown-date"},
	}
	discovery, err := tracker.Remember(ctx, "track-s", 1, documents, now)
	if err != nil {
		t.Fatal(err)
	}
	if discovery.NewDocuments != 4 || discovery.NewestPublication == nil || !discovery.NewestPublication.Equal(*documents[0].PublicationDate) {
		t.Fatalf("newly discovered publications not reported: %#v", discovery)
	}
	discovery, err = tracker.Remember(ctx, "track-s", 1, documents, now.Add(time.Minute))
	if err != nil || discovery.NewDocuments != 0 || discovery.NewestPublication != nil {
		t.Fatalf("unchanged discovery mislabeled: result=%#v err=%v", discovery, err)
	}
	if err := tracker.SetReviewState(ctx, "track-s", documents[0].URL, "irrelevant", "none"); err != nil {
		t.Fatal(err)
	}
	if err := tracker.SetReviewState(ctx, "track-s", documents[1].URL, "relevant", "ongoing"); err != nil {
		t.Fatal(err)
	}
	if err := tracker.SetReviewState(ctx, "track-s", documents[2].URL, "irrelevant", "none"); err != nil {
		t.Fatal(err)
	}
	explicit := "https://source.example/explicit-old"
	if err := tracker.AddExplicitReference(ctx, "track-s", 1, explicit, date(2025, 1, 1), now); err != nil {
		t.Fatal(err)
	}
	attachment := "https://source.example/old-ongoing.pdf"
	if err := tracker.AddDependency(ctx, "track-s", documents[1].URL, attachment, true, now); err != nil {
		t.Fatal(err)
	}
	bootstrap, err := tracker.Plan(ctx, "track-s", 1, now, 30)
	if err != nil {
		t.Fatal(err)
	}
	if !bootstrap.Bootstrap || plannedURLs(bootstrap) != explicit+","+documents[1].URL+","+documents[0].URL+","+documents[3].URL || len(bootstrap.Documents[1].Resources) != 1 {
		t.Fatalf("unexpected bootstrap plan: %#v urls=%s", bootstrap, plannedURLs(bootstrap))
	}
	if err := tracker.CompleteBootstrap(ctx, "track-s", 1, now); err != nil {
		t.Fatal(err)
	}
	recurring, err := tracker.Plan(ctx, "track-s", 1, now.Add(10*time.Minute), 30)
	if err != nil {
		t.Fatal(err)
	}
	if recurring.Bootstrap || plannedURLs(recurring) != documents[1].URL+","+documents[0].URL+","+documents[3].URL {
		t.Fatalf("unexpected recurring plan: %#v urls=%s", recurring, plannedURLs(recurring))
	}
	changed, err := tracker.RecordVersion(ctx, "track-s", documents[1].URL, 10, now)
	if err != nil || changed {
		t.Fatalf("first acquisition mislabeled as revision: changed=%v err=%v", changed, err)
	}
	changed, _ = tracker.RecordVersion(ctx, "track-s", documents[1].URL, 10, now.Add(time.Minute))
	if changed {
		t.Fatal("unchanged acquisition mislabeled as revision")
	}
	changed, _ = tracker.RecordVersion(ctx, "track-s", documents[1].URL, 11, now.Add(2*time.Minute))
	if !changed {
		t.Fatal("new version not detected")
	}
	// A reviewed unpadded layout repairs a previously undated old target, while
	// recent and ongoing targets retain their existing selection behavior.
	cfg := registry.Configuration{Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/"}, ListingItemClass: "card-body", ListingDateClass: "h5", ListingDateLayout: "2 Jan 2006", ListingDateLocale: "it"}}
	body := []byte(`<div class="card-body"><span class="h5">8 ottobre 2024</span><a href="/unknown-date">Old notice</a></div>`)
	corrected := discoverDocuments("https://source.example/list", body, cfg, []Link{{URL: documents[3].URL}})
	if len(corrected) != 1 || corrected[0].PublicationDate == nil {
		t.Fatal("unpadded date not discovered")
	}
	if _, err := tracker.Remember(ctx, "track-s", 1, corrected, now.Add(20*time.Minute)); err != nil {
		t.Fatal(err)
	}
	repaired, err := tracker.Plan(ctx, "track-s", 1, now.Add(21*time.Minute), 30)
	if err != nil || repaired.Bootstrap || plannedURLs(repaired) != documents[1].URL+","+documents[0].URL {
		t.Fatalf("old repaired target still planned: %#v %v", repaired, err)
	}
	var persisted time.Time
	if err := pool.QueryRow(ctx, "SELECT source_publication_date FROM acquisition_targets WHERE source_id='track-s' AND url=$1", documents[3].URL).Scan(&persisted); err != nil || !persisted.Equal(*corrected[0].PublicationDate) {
		t.Fatalf("corrected date not persisted: %v %v", persisted, err)
	}
}

func plannedURLs(plan RevisionPlan) string {
	values := make([]string, 0, len(plan.Documents))
	for _, document := range plan.Documents {
		values = append(values, document.URL)
	}
	return strings.Join(values, ",")
}
