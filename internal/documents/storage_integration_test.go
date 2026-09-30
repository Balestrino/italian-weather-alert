//go:build integration

package documents

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

// Own an isolated PostgreSQL container. No fixture is written to the Compose DB.
func testDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	name := "iwa-documents-test-" + hex.EncodeToString(b[:6])
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

func testObjects(t *testing.T) (*S3, func()) {
	t.Helper()
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		t.Fatal(err)
	}
	name := "iwa-objects-test-" + hex.EncodeToString(token[:6])
	key := hex.EncodeToString(token)
	dir := t.TempDir()
	for _, file := range []string{"access", "secret"} {
		if err := os.WriteFile(filepath.Join(dir, file), []byte(key), 0644); err != nil {
			t.Fatal(err)
		}
	}
	run := func(args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		b, err := exec.CommandContext(ctx, "docker", args...).Output()
		if err != nil {
			t.Fatalf("isolated RustFS operation failed: %v", err)
		}
		return strings.TrimSpace(string(b))
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	binding := listener.Addr().String() + ":9000"
	if err = listener.Close(); err != nil {
		t.Fatal(err)
	}
	run("run", "--pull=never", "-d", "--name", name, "-p", binding, "--volume", name+"-data:/data", "--mount", "type=bind,src="+filepath.Join(dir, "access")+",dst=/run/secrets/access,readonly", "--mount", "type=bind,src="+filepath.Join(dir, "secret")+",dst=/run/secrets/secret,readonly", "-e", "RUSTFS_ACCESS_KEY_FILE=/run/secrets/access", "-e", "RUSTFS_SECRET_KEY_FILE=/run/secrets/secret", "-e", "RUSTFS_CONSOLE_ENABLE=false", "rustfs/rustfs:1.0.0")
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-fv", name).Run()
		_ = exec.Command("docker", "volume", "rm", name+"-data").Run()
	})
	endpoint := "http://" + run("port", name, "9000/tcp")
	objects, err := NewS3(endpoint, "iwa-test-originals", key, key)
	if err != nil {
		t.Fatal(err)
	}
	wait := func() {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		for {
			if err := objects.Initialize(ctx); err == nil {
				return
			}
			if ctx.Err() != nil {
				b, _ := exec.Command("docker", "logs", "--tail", "20", name).CombinedOutput()
				t.Fatalf("RustFS did not become ready: %s", strings.ReplaceAll(string(b), key, "[redacted]"))
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	wait()
	return objects, func() { run("restart", "--time", "5", name); wait() }
}

func fixture(id string) Acquisition {
	return Acquisition{ID: id, SourceID: "municipal", Configuration: 1, URL: "https://comune.example/notice", Metadata: json.RawMessage(`{"published":"2026-09-01","updated":"2026-09-02"}`), Resources: []Resource{
		{URL: "https://comune.example/notice", Role: "original", Required: true, SourceID: "municipal", Configuration: 1, MediaType: "text/html", Bytes: []byte("<html>notice with ordinance and map</html>")},
		{URL: "https://comune.example/order.pdf", Role: "attachment", Required: true, SourceID: "municipal", Configuration: 1, MediaType: "application/pdf", Bytes: []byte("%PDF-synthetic ordinance")},
		{URL: "https://comune.example/map.png", Role: "resource", Required: true, SourceID: "municipal", Configuration: 1, MediaType: "image/png", Bytes: []byte("synthetic graphical resource")},
	}}
}

type faultObjects struct {
	Objects
	calls  int
	failAt int
	after  func()
}

func (f *faultObjects) Ensure(ctx context.Context, hash string, b []byte) error {
	f.calls++
	if f.calls == f.failAt {
		return ErrStorage
	}
	if err := f.Objects.Ensure(ctx, hash, b); err != nil {
		return err
	}
	if f.after != nil {
		fn := f.after
		f.after = nil
		fn()
	}
	return nil
}

func TestRetentionPostgresRustFS(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	pool, restartDB := testDB(t)
	objects, restartObjects := testObjects(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(registry.Migrate(ctx, pool))
	must(Migrate(ctx, pool))
	must(Migrate(ctx, pool))
	reg := registry.New(pool)
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "municipality", Name: "Fixture municipality", OfficialURL: "https://comune.example"}))
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "publisher", Name: "Fixture publisher", OfficialURL: "https://publisher.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "channel", PublisherID: "publisher", Platform: "fixture", URL: "https://comune.example"}))
	evidence := registry.Evidence{URL: "https://comune.example/policy", Locator: "synthetic permission", ObservedAt: time.Now().UTC()}
	c := registry.Configuration{URL: "https://comune.example", Sections: []string{"https://comune.example/notices"}, AccessMethod: "html", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}
	must(reg.CreateSource(ctx, registry.Source{ID: "municipal", AuthorityID: "municipality", ChannelID: "channel", ProductID: "municipal", Territory: "050004"}, c, "test"))
	store := New(pool, objects)
	a := fixture("first")
	issuer := "municipality"
	a.IssuerID = &issuer
	v, err := store.Retain(ctx, a)
	must(err)
	if !v.Complete || len(v.Resources) != 3 {
		t.Fatal("required evidence incomplete")
	}
	for _, r := range a.Resources {
		got, err := store.Read(ctx, v.ID, r.URL)
		must(err)
		if !bytes.Equal(got, r.Bytes) {
			t.Fatal("original bytes differ")
		}
	}
	// A retained object is inaccessible anonymously.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, objects.client.EndpointURL().String()+"/"+objects.bucket+"/"+objectKey(digest(a.Resources[0].Bytes)), nil)
	must(err)
	resp, err := http.DefaultClient.Do(req)
	must(err)
	resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("anonymous object response: %d", resp.StatusCode)
	}
	again, err := store.Retain(ctx, a)
	must(err)
	if again.ID != v.ID {
		t.Fatal("same request duplicated version")
	}
	a.ID = "unchanged"
	a.Resources[0], a.Resources[2] = a.Resources[2], a.Resources[0]
	again, err = store.Retain(ctx, a)
	must(err)
	if again.ID != v.ID || !again.FirstAcquiredAt.Equal(v.FirstAcquiredAt) || !bytes.Equal(again.Metadata, v.Metadata) {
		t.Fatal("unchanged acquisition changed history")
	}
	a.Resources[0].Bytes = []byte("revised map at same URL")
	if _, err = store.Retain(ctx, a); !errors.Is(err, ErrConflict) {
		t.Fatal("idempotency conflict not detected")
	}
	a.ID = "changed-resource"
	changed, err := store.Retain(ctx, a)
	must(err)
	if changed.ID == v.ID {
		t.Fatal("changed attachment lost")
	}
	a.ID = "changed-page"
	a.Resources[2].Bytes = []byte("revised page")
	revised, err := store.Retain(ctx, a)
	must(err)
	if revised.ID == changed.ID {
		t.Fatal("stable URL revision lost")
	}
	old, err := store.Read(ctx, v.ID, a.URL)
	must(err)
	if string(old) != "<html>notice with ordinance and map</html>" {
		t.Fatal("earlier version overwritten")
	}
	view := fixture("view-first")
	view.URL = "https://comune.example/view"
	view.Resources = view.Resources[:1]
	view.Resources[0].URL = view.URL
	view.Resources[0].Bytes = []byte("<html>Notice<!-- js-view-dom-id-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --> unchanged</html>")
	firstView, err := store.Retain(ctx, view)
	must(err)
	view.ID = "view-equivalent"
	view.Resources[0].Bytes = []byte("<html>Notice<!-- js-view-dom-id-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb --> unchanged</html>")
	noTransfer := &faultObjects{Objects: objects, failAt: 1}
	equivalentView, err := New(pool, noTransfer).Retain(ctx, view)
	must(err)
	if equivalentView.ID != firstView.ID || noTransfer.calls != 0 {
		t.Fatal("transport-only Drupal marker created a new version or object")
	}
	retainedView, err := store.Read(ctx, firstView.ID, view.URL)
	must(err)
	if !bytes.Contains(retainedView, []byte("js-view-dom-id-aaaaaaaa")) {
		t.Fatal("retained first original was modified")
	}
	view.ID = "view-updated"
	view.Resources[0].Bytes = []byte("<html>Changed notice<!-- js-view-dom-id-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb --> unchanged</html>")
	updatedView, err := store.Retain(ctx, view)
	must(err)
	if updatedView.ID == firstView.ID {
		t.Fatal("substantive HTML change was deduplicated")
	}
	missing := fixture("missing")
	missing.Resources[1].Bytes = nil
	missing.Resources[1].Missing = "unavailable"
	incomplete, err := store.Retain(ctx, missing)
	must(err)
	if incomplete.Complete {
		t.Fatal("missing required attachment reported complete")
	}
	repaired := fixture("now-available")
	repairedVersion, err := store.Retain(ctx, repaired)
	must(err)
	if repairedVersion.ID == incomplete.ID || !repairedVersion.Complete {
		t.Fatal("recovered resource not versioned")
	}
	// Retention permission is checked independently for every resource.
	c.Policy.RetentionPermitted = false
	rev, err := reg.AppendConfiguration(ctx, "municipal", 1, c, "test")
	must(err)
	denied := fixture("denied")
	denied.Resources[1].Configuration = rev
	if _, err = store.Retain(ctx, denied); !errors.Is(err, ErrPolicy) {
		t.Fatal("attachment inherited forbidden retention")
	}
	// Database staging failure must not start any object transfer.
	execSQL(ctx, pool, t, `CREATE FUNCTION fail_retention_stage() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected staging failure'; END; $$`)
	execSQL(ctx, pool, t, `CREATE TRIGGER injected_stage_failure BEFORE INSERT ON retained_acquisitions FOR EACH ROW EXECUTE FUNCTION fail_retention_stage()`)
	beforeStage := &faultObjects{Objects: objects}
	if _, err = New(pool, beforeStage).Retain(ctx, fixture("stage-failure")); err == nil || beforeStage.calls != 0 {
		t.Fatal("object effects before staging commit")
	}
	execSQL(ctx, pool, t, "DROP TRIGGER injected_stage_failure ON retained_acquisitions")
	// Partial object write: no visible version; staged bytes survive DB restart.
	partial := fixture("partial")
	partial.URL = "https://comune.example/partial"
	partial.Resources[0].URL = partial.URL
	faults := &faultObjects{Objects: objects, failAt: 2}
	if _, err = New(pool, faults).Retain(ctx, partial); !errors.Is(err, ErrStorage) {
		t.Fatal("injected write failure lost")
	}
	var count int
	must(pool.QueryRow(ctx, "SELECT count(*) FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id WHERE d.official_url=$1", partial.URL).Scan(&count))
	if count != 0 {
		t.Fatal("partial version visible")
	}
	pending, err := store.Pending(ctx, 100)
	must(err)
	if len(pending) != 1 || pending[0] != partial.ID {
		t.Fatal("partial acquisition not recoverable")
	}
	restartDB()
	restartObjects()
	recovered, err := New(pool, objects).Recover(ctx, partial.ID)
	must(err)
	if !recovered.Complete {
		t.Fatal("recovery incomplete")
	}
	// Lost acknowledgement after an S3 write: retry reuses the existing object.
	lost := fixture("lost-ack")
	lost.Resources[2].Bytes = []byte("new unique map bytes written before cancellation")
	callCtx, callCancel := context.WithCancel(ctx)
	faults = &faultObjects{Objects: objects, after: callCancel}
	if _, err = New(pool, faults).Retain(callCtx, lost); err == nil {
		t.Fatal("cancelled transfer succeeded")
	}
	_, err = New(pool, objects).Recover(ctx, lost.ID)
	must(err)
	// A PostgreSQL finalization failure after all S3 writes retains its journal.
	execSQL(ctx, pool, t, `CREATE FUNCTION fail_retention_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected finalization failure'; END; $$`)
	execSQL(ctx, pool, t, `CREATE TRIGGER injected_failure BEFORE INSERT ON retained_versions FOR EACH ROW EXECUTE FUNCTION fail_retention_commit()`)
	failedDB := fixture("db-failure")
	failedDB.Resources[0].Bytes = []byte("database failure after object write")
	if _, err = store.Retain(ctx, failedDB); err == nil {
		t.Fatal("database failure not injected")
	}
	execSQL(ctx, pool, t, "DROP TRIGGER injected_failure ON retained_versions")
	_, err = New(pool, objects).Recover(ctx, failedDB.ID)
	must(err)
	// Concurrency: independent checks and duplicate replay finalize only one version.
	var wg sync.WaitGroup
	ids := make(chan int64, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			in := fixture(fmt.Sprintf("concurrent-%d", i%4))
			in.Resources[0].Bytes = []byte("concurrent version")
			got, e := New(pool, objects).Retain(ctx, in)
			ids <- got.ID
			errs <- e
		}(i)
	}
	wg.Wait()
	close(ids)
	close(errs)
	for e := range errs {
		must(e)
	}
	var first int64
	for id := range ids {
		if first == 0 {
			first = id
		}
		if id != first {
			t.Fatal("concurrent duplicate versions")
		}
	}
	pending, err = store.Pending(ctx, 100)
	must(err)
	if len(pending) != 0 {
		t.Fatal("completed staging not cleared")
	}
	must(pool.QueryRow(ctx, "SELECT count(*) FROM retained_acquisitions WHERE payload IS NOT NULL").Scan(&count))
	if count != 0 {
		t.Fatal("staged bytes retained after commit")
	}
	// Corruption is reported, never accepted on size or ETag alone.
	original := fixture("unused").Resources[0].Bytes
	_, err = objects.client.PutObject(ctx, objects.bucket, objectKey(digest(original)), bytes.NewReader(bytes.Repeat([]byte("x"), len(original))), int64(len(original)), minio.PutObjectOptions{DisableMultipart: true})
	must(err)
	if _, err = store.Read(ctx, v.ID, "https://comune.example/notice"); !errors.Is(err, ErrStorage) {
		t.Fatal("object corruption accepted")
	}
	// Retention cleanup uses an idempotent RustFS delete after its PostgreSQL
	// transaction has durably removed references and queued the object hash.
	cleanupBody := []byte("unreferenced retention cleanup fixture")
	cleanupHash := digest(cleanupBody)
	must(objects.Ensure(ctx, cleanupHash, cleanupBody))
	must(objects.Delete(ctx, cleanupHash))
	must(objects.Delete(ctx, cleanupHash))
	if _, err = objects.Read(ctx, cleanupHash, int64(len(cleanupBody))); !errors.Is(err, errObjectMissing) {
		t.Fatalf("deleted RustFS object remained readable: %v", err)
	}
}

func execSQL(ctx context.Context, pool *pgxpool.Pool, t *testing.T, sql string) {
	t.Helper()
	if _, err := pool.Exec(ctx, sql); err != nil {
		t.Fatal(err)
	}
}
