//go:build integration

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUnavailableTargetCommandsListAndRecordDecisions(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		t.Fatal(err)
	}
	name, password := "iwa-unavailable-cli-test-"+hex.EncodeToString(token[:6]), hex.EncodeToString(token)
	file := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(file, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	output, err := exec.CommandContext(ctx, "docker", "run", "--pull=never", "--detach", "--name", name, "--publish", address+":5432", "--mount", "type=bind,src="+file+",dst=/run/secrets/password,readonly", "--env", "POSTGRES_PASSWORD_FILE=/run/secrets/password", "--env", "POSTGRES_USER=iwa", "--env", "POSTGRES_DB=iwa", "postgres:16-alpine").CombinedOutput()
	if err != nil {
		t.Fatalf("isolated PostgreSQL failed: %v %s", err, strings.ReplaceAll(string(output), password, "[redacted]"))
	}
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", name).Run() })
	connection, err := pgxpool.ParseConfig("postgres://iwa@" + address + "/iwa?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	connection.ConnConfig.Password = password
	pool, err := pgxpool.NewWithConfig(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	for pool.Ping(ctx) != nil {
		if ctx.Err() != nil {
			t.Fatal("test database unavailable")
		}
		time.Sleep(100 * time.Millisecond)
	}
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, acquisition.Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	reg := registry.New(pool)
	evidence := registry.Evidence{URL: "https://source.example/reuse", Locator: "fixture", ObservedAt: now}
	for _, err := range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "cli-a", Name: "A", OfficialURL: "https://source.example"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "cli-c", PublisherID: "cli-a", Platform: "fixture", URL: "https://source.example"}),
		reg.CreateSource(ctx, registry.Source{ID: "cli-s", AuthorityID: "cli-a", ChannelID: "cli-c", ProductID: "municipal", Territory: "x"}, registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/list"}, AccessMethod: "html", Attribution: "A", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true}}, "test"),
		reg.RecordPreview(ctx, "cli-s", 1, "test", evidence),
		reg.EnableCollection(ctx, "cli-s", 1, "test"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	tracker := acquisition.NewTrackingStore(pool)
	stale := "https://source.example/notices/renamed"
	seenAt := now.Add(-time.Hour)
	if _, err := tracker.Remember(ctx, "cli-s", 1, []acquisition.DiscoveredDocument{{URL: stale}}, seenAt); err != nil {
		t.Fatal(err)
	}
	if err := tracker.RecordUnavailable(ctx, "cli-s", 1, stale, 404, now); err != nil {
		t.Fatal(err)
	}
	// Classification needs a later traversal that no longer listed the URL.
	if _, err := pool.Exec(ctx, `INSERT INTO acquisition_checks(source_id,configuration,worker_id,started_at,finished_at,reachable,content_recognized,complete,listing_count,document_count,check_state)
 VALUES('cli-s',1,'test',$1,$1,true,true,true,1,0,'complete_unchanged')`, now); err != nil {
		t.Fatal(err)
	}
	run := func(args []string, input string) (map[string]json.RawMessage, error) {
		t.Helper()
		c, ok, err := parseControlCommand(args)
		if !ok || err != nil {
			t.Fatal(args, err)
		}
		c.Input = strings.NewReader(input)
		var out bytes.Buffer
		if err := executeControlEnvironment(ctx, pool, "admin", "development", c, &out); err != nil {
			return nil, err
		}
		var result map[string]json.RawMessage
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatalf("output is not JSON: %s", out.String())
		}
		return result, nil
	}
	if err := tracker.SetReviewState(ctx, "cli-s", stale, "relevant", "unresolved"); err != nil {
		t.Fatal(err)
	}
	tracked, err := run([]string{"source-tracked-targets", "cli-s"}, "")
	if err != nil {
		t.Fatal(err)
	}
	var followed []acquisition.FollowedTarget
	if err := json.Unmarshal(tracked["followed_targets"], &followed); err != nil || len(followed) != 1 || followed[0].URL != stale || followed[0].MeasureState != "unresolved" || followed[0].Recheck != "window" || followed[0].LimitDays != 90 {
		t.Fatalf("followed targets: %s %v", tracked["followed_targets"], err)
	}
	if _, err := run([]string{"source-tracked-targets", "missing-source"}, ""); err == nil || controlErrorCode(err) != "control_unavailable" {
		t.Fatalf("unknown source: %v", err)
	}
	// Without a measure the stale notice is withdrawn and can be excluded below.
	if err := tracker.SetReviewState(ctx, "cli-s", stale, "unknown", "none"); err != nil {
		t.Fatal(err)
	}
	listed, err := run([]string{"source-unavailable-targets", "cli-s"}, "")
	if err != nil {
		t.Fatal(err)
	}
	var targets []acquisition.UnavailableTarget
	if err := json.Unmarshal(listed["unavailable_targets"], &targets); err != nil || len(targets) != 1 || targets[0].URL != stale || targets[0].Classification != "withdrawn" || !targets[0].LastSeenAt.Equal(seenAt) {
		t.Fatalf("listed targets: %s %v", listed["unavailable_targets"], err)
	}
	payload := `{"url":"` + stale + `","observed_last_seen_at":"` + seenAt.Format(time.RFC3339Nano) + `","reason":"absent from complete listings","evidence":{"pages":1}}`
	recorded, err := run([]string{"source-unavailable-decide", "cli-s", "1", "exclude_stale_unavailable", "operator"}, payload)
	if err != nil {
		t.Fatal(err)
	}
	var disposition acquisition.TargetDisposition
	if err := json.Unmarshal(recorded["disposition"], &disposition); err != nil || disposition.ID == 0 || disposition.Actor != "operator" {
		t.Fatalf("recorded disposition: %s %v", recorded["disposition"], err)
	}
	_, err = run([]string{"source-unavailable-decide", "cli-s", "1", "exclude_stale_unavailable", "operator"}, payload)
	if !errors.Is(err, acquisition.ErrTargetAlreadyExcluded) || controlErrorCode(err) != "target_already_excluded" {
		t.Fatalf("repeated exclusion: %v", err)
	}
	listed, err = run([]string{"source-unavailable-targets", "cli-s"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(listed["unavailable_targets"], &targets); err != nil || targets[0].Classification != "excluded" || len(targets[0].Dispositions) != 1 {
		t.Fatalf("listed after exclusion: %s", listed["unavailable_targets"])
	}
}
