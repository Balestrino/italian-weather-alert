//go:build integration

package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

func adminTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	name := "iwa-admin-test-" + hex.EncodeToString(random[:6])
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

type adminObjects map[string][]byte

func (m adminObjects) Ensure(_ context.Context, hash string, b []byte) error {
	m[hash] = bytes.Clone(b)
	return nil
}
func (m adminObjects) Read(_ context.Context, hash string, _ int64) ([]byte, error) {
	return bytes.Clone(m[hash]), nil
}

type adminCrawler struct {
	calls int
	fail  bool
}

func (c *adminCrawler) Crawl(_ context.Context, target string) (acquisition.Page, error) {
	c.calls++
	if c.fail {
		return acquisition.Page{}, errors.New("upstream credential must never be exposed")
	}
	p := acquisition.Page{URL: target, HTML: []byte("configured listing"), StatusCode: 200}
	if strings.Contains(target, "listing") {
		p.Links = []acquisition.Link{{URL: "https://source.example/notices/one"}}
	} else {
		p.HTML = []byte("document original")
	}
	return p, nil
}

func TestAdminPersistedSourceLifecycle(t *testing.T) { testAdminPersistedSourceLifecycle(t, false) }
func TestAdminGuidedSourceLifecycle(t *testing.T)    { testAdminPersistedSourceLifecycle(t, true) }
func testAdminPersistedSourceLifecycle(t *testing.T, guided bool) {
	ctx := context.Background()
	pool := adminTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, acquisition.Migrate, registry.Migrate, acquisition.Migrate} {
		must(migrate(ctx, pool))
	}
	reg := registry.New(pool)
	objects := adminObjects{}
	crawler := &adminCrawler{}
	engine := &acquisition.Engine{Registry: reg, Retained: documents.New(pool, objects), Crawler: crawler}
	host := httptest.NewServer(HandlerWithAdministration(nil, AdminRuntime{Registry: reg, Preview: engine}))
	defer host.Close()
	request := func(method, path string, payload any, code int) []byte {
		t.Helper()
		var body io.Reader
		if payload != nil {
			b, e := json.Marshal(payload)
			must(e)
			body = bytes.NewReader(b)
		}
		req, e := http.NewRequest(method, host.URL+path, body)
		must(e)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		if guided && method == "POST" {
			action := path[strings.LastIndex(path, "/")+1:]
			if len(sourceFields(action, adminSourceView{})) > 0 {
				values := url.Values{}
				for k, v := range payload.(map[string]any) {
					values.Set(k, fmt.Sprint(v))
				}
				req.Body = io.NopCloser(strings.NewReader(values.Encode()))
				req.ContentLength = int64(len(values.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		}
		resp, e := http.DefaultClient.Do(req)
		must(e)
		defer resp.Body.Close()
		b, e := io.ReadAll(resp.Body)
		must(e)
		if resp.StatusCode != code {
			t.Fatalf("%s %s: %d wanted %d: %s", method, path, resp.StatusCode, code, b)
		}
		return b
	}
	request("POST", "/admin/authorities", registry.Authority{ID: "a", Name: "Comune", OfficialURL: "https://source.example"}, 200)
	request("POST", "/admin/channels", registry.Channel{ID: "c", PublisherID: "a", Platform: "official", URL: "https://source.example"}, 200)
	evidence := registry.Evidence{URL: "https://source.example/reuse", Locator: "synthetic permitted source", ObservedAt: time.Now().UTC()}
	cfg := registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/listing"}, AccessMethod: "crawl4ai", Attribution: "Comune", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}, Discovery: registry.Discovery{PaginationParameter: "page", DocumentPathPrefixes: []string{"/notices/"}, ListingContentMarkers: []string{"configured"}, MaxPagesPerSection: 2, MaxDocuments: 10}}
	request("POST", "/admin/sources", map[string]any{"actor": "creator", "source": registry.Source{ID: "s", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "050004"}, "configuration": cfg}, 200)
	action := func(rev int) any { return map[string]any{"revision": rev, "actor": "operator"} }
	state := func(active int, enabled bool) {
		t.Helper()
		st, e := reg.State(ctx, "s")
		must(e)
		if (active == 0 && st.ActiveRevision != nil) || (active > 0 && (st.ActiveRevision == nil || *st.ActiveRevision != active)) || st.CollectionEnabled != enabled || st.PublicEnabled {
			t.Fatalf("unexpected state: %#v", st)
		}
	}
	state(0, false)
	request("POST", "/admin/sources/s/enable-collection", action(1), 409)
	crawler.fail = true
	b := request("POST", "/admin/sources/s/preview", action(1), 503)
	if bytes.Contains(b, []byte("credential")) {
		t.Fatal("upstream error leaked")
	}
	request("POST", "/admin/sources/s/enable-collection", action(1), 409)
	crawler.fail = false
	b = request("POST", "/admin/sources/s/preview", action(1), 200)
	var preview acquisition.Preview
	must(json.Unmarshal(b, &preview))
	if len(preview.Documents) != 1 || len(preview.Sections) != 1 || len(objects) != 2 {
		t.Fatalf("incomplete real retained preview: %s", b)
	}
	state(0, false)
	request("POST", "/admin/sources/s/enable-collection", action(1), 200)
	state(1, true)
	schedule := acquisition.NewScheduleStore(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	must(schedule.SyncEnabled(ctx, now))
	claim, err := schedule.ClaimDue(ctx, "worker", now, time.Minute)
	must(err)
	complete := now.Add(time.Second)
	must(schedule.Finish(ctx, claim, acquisition.CheckOutcome{SourceID: "s", Configuration: 1, StartedAt: claim.StartedAt, FinishedAt: complete, Reachable: true, ContentRecognized: true, Complete: true}))
	// All three boundary changes remain drafts, including after failed previews.
	cfg.URL = "https://source.example/new"
	cfg.Sections = []string{"https://source.example/other-listing"}
	cfg.AccessMethod = "unsupported"
	request("POST", "/admin/sources/s/configuration", map[string]any{"expected_revision": 1, "actor": "editor", "configuration": cfg}, 200)
	request("POST", "/admin/sources/s/configuration", map[string]any{"expected_revision": 1, "actor": "stale", "configuration": cfg}, 409)
	state(1, true)
	calls := crawler.calls
	request("POST", "/admin/sources/s/preview", action(2), 400)
	if crawler.calls != calls {
		t.Fatal("unsupported acquisition method executed")
	}
	request("POST", "/admin/sources/s/enable-collection", action(2), 409)
	state(1, true)
	request("POST", "/admin/sources/s/intervals", map[string]any{"expected_revision": 0, "check_seconds": 30, "delay_seconds": 60, "actor": "scheduler"}, 200)
	status, err := schedule.Status(ctx, "s", complete.Add(time.Minute))
	must(err)
	if !status.NextCheckAt.Equal(complete.Add(30*time.Second)) || status.UpdatingState != "delayed" || status.LastCompleteAt == nil || !status.LastCompleteAt.Equal(complete) {
		t.Fatalf("interval change was not immediate: %#v", status)
	}
	request("POST", "/admin/sources/s/intervals", map[string]any{"expected_revision": 0, "check_seconds": 99, "delay_seconds": 99, "actor": "stale"}, 409)
	state(1, true)
	request("POST", "/admin/sources/s/suspend-collection", action(1), 200)
	state(1, false)
	if _, err = schedule.ClaimDue(ctx, "worker", complete.Add(time.Minute), time.Minute); !errors.Is(err, acquisition.ErrNoDueCheck) {
		t.Fatal("suspended source claimed", err)
	}
	request("POST", "/admin/sources/s/enable-collection", action(1), 200)
	// Retry-After and in-flight leases survive interval edits.
	claim, err = schedule.ClaimDue(ctx, "worker", complete.Add(time.Minute), time.Minute)
	must(err)
	request("POST", "/admin/sources/s/intervals", map[string]any{"expected_revision": 1, "check_seconds": 20, "delay_seconds": 60, "actor": "scheduler"}, 200)
	retry := claim.StartedAt.Add(time.Hour)
	must(schedule.Finish(ctx, claim, acquisition.CheckOutcome{SourceID: "s", Configuration: 1, StartedAt: claim.StartedAt, FinishedAt: claim.StartedAt.Add(time.Second), ErrorCode: "source_rate_limited", RetryAfter: &retry}))
	request("POST", "/admin/sources/s/intervals", map[string]any{"expected_revision": 2, "check_seconds": 10, "delay_seconds": 120, "actor": "scheduler"}, 200)
	status, err = schedule.Status(ctx, "s", complete)
	must(err)
	if !status.NextCheckAt.Equal(retry) || status.ConsecutiveFailures != 1 {
		t.Fatalf("interval reset backoff: %#v", status)
	}
	cfg.AccessMethod = "crawl4ai"
	request("POST", "/admin/sources/s/configuration", map[string]any{"expected_revision": 2, "actor": "editor", "configuration": cfg}, 200)
	request("POST", "/admin/sources/s/preview", action(3), 200)
	state(1, true)
	request("POST", "/admin/sources/s/enable-collection", action(2), 409)
	request("POST", "/admin/sources/s/enable-collection", action(3), 200)
	state(3, true)
	must(schedule.SyncEnabled(ctx, now))
	// A fresh store reconstructs state, original snapshots, full preview and actor audit.
	reg = registry.New(pool)
	versions, err := reg.Versions(ctx, "s")
	must(err)
	if len(versions) != 3 || versions[0].Configuration.URL != "https://source.example" || versions[0].Actor != "creator" || versions[1].Configuration.AccessMethod != "unsupported" {
		t.Fatal("immutable history lost")
	}
	intervals, err := reg.Intervals(ctx, "s")
	must(err)
	if intervals.Revision != 3 || *intervals.CheckSeconds != 10 || *intervals.DelaySeconds != 120 {
		t.Fatal("overrides lost at activation")
	}
	events, err := reg.Events(ctx, "s")
	must(err)
	found := 0
	for _, event := range events {
		if event.Kind == "preview" {
			var e registry.Evidence
			must(json.Unmarshal(event.Evidence, &e))
			if len(e.Report) == 0 {
				t.Fatal("preview report not persisted")
			}
			found++
		}
	}
	if found != 2 {
		t.Fatal("failed preview recorded as successful")
	}
	request("GET", "/admin/sources", nil, 200)
	request("GET", "/admin/sources/s", nil, 200)
	// The actual no-script HTML page offers escaped history and forms, including
	// form decoding through the same mutation path as the JSON client.
	resp, err := http.Get(host.URL + "/admin/sources/s")
	must(err)
	html, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	must(err)
	if resp.StatusCode != 200 || !bytes.Contains(html, []byte("Cronologia configurazioni")) || !bytes.Contains(html, []byte("/suspend-collection")) {
		t.Fatalf("missing admin page: %s", html)
	}
	resp, err = http.PostForm(host.URL+"/admin/sources/s/suspend-collection", url.Values{"payload": {`{"revision":3,"actor":"browser"}`}})
	must(err)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal("browser form failed")
	}
	state(3, false)
	// Reject browser cross-origin requests before touching the database.
	req, err := http.NewRequest("POST", host.URL+"/admin/sources/s/enable-collection", strings.NewReader(`{"revision":3,"actor":"attacker"}`))
	must(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://evil.example")
	resp, err = http.DefaultClient.Do(req)
	must(err)
	resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatal("cross-origin mutation allowed")
	}
	state(3, false)
	t.Log("real PostgreSQL + HTTP: drafts, retained previews, explicit activation, immutable history, immediate intervals/suspension, CAS, leases, backoff and browser forms passed")
}
