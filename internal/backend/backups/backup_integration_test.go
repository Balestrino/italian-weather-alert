//go:build integration

package backups

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/evidenceguard"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/notifications"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func testDB(t *testing.T) (*pgxpool.Pool, string) {
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
	return pool, name
}

func testObjects(t *testing.T) (*documents.S3, Config) {
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
	objects, err := documents.NewS3(endpoint, "iwa-test-originals", key, key)
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
	return objects, Config{Enabled: true, Endpoint: endpoint, Bucket: "iwa-backup-test", AccessKey: key, SecretKey: key, Region: "us-east-1", Prefix: "test", LoopbackTest: true, IntervalSeconds: 86400, RetentionDays: 30, TimeoutSeconds: 60, ScratchDirectory: t.TempDir()}
}

type dockerDump struct {
	name  string
	after func()
}

func (d dockerDump) Dump(ctx context.Context, snapshot, path string) error {
	if d.after != nil {
		d.after()
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	cmd := exec.CommandContext(ctx, "docker", "exec", d.name, "pg_dump", "-U", "iwa", "-d", "iwa", "--format=custom", "--no-owner", "--no-acl", "--snapshot="+snapshot)
	cmd.Stdout = f
	return cmd.Run()
}

type brokenDestination struct {
	Destination
	failPublish, failPrune bool
}

func (d brokenDestination) Publish(ctx context.Context, b Bundle, at time.Time) error {
	if d.failPublish {
		return ErrBackup
	}
	return d.Destination.Publish(ctx, b, at)
}
func (d brokenDestination) Prune(ctx context.Context, cut time.Time, keep string) error {
	if d.failPrune {
		return ErrBackup
	}
	return d.Destination.Prune(ctx, cut, keep)
}
func TestConsistentBundleScheduleTransferAndFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	pool, name := testDB(t)
	objects, c := testObjects(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, notifications.Migrate, Migrate, Migrate} {
		must(m(ctx, pool))
	}
	reg := registry.New(pool)
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "municipality", Name: "Fixture", OfficialURL: "https://comune.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "channel", PublisherID: "municipality", Platform: "fixture", URL: "https://comune.example"}))
	evidence := registry.Evidence{URL: "https://comune.example/policy", Locator: "fixture", ObservedAt: time.Now()}
	config := registry.Configuration{URL: "https://comune.example", Sections: []string{"https://comune.example/notices"}, AccessMethod: "html", Attribution: "fixture", Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}}
	must(reg.CreateSource(ctx, registry.Source{ID: "municipal", AuthorityID: "municipality", ChannelID: "channel", ProductID: "municipal", Territory: "050004"}, config, "test"))
	acquisition := documents.Acquisition{ID: "first", SourceID: "municipal", Configuration: 1, URL: "https://comune.example/notice", Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{
		{URL: "https://comune.example/notice", Role: "original", Required: true, SourceID: "municipal", Configuration: 1, MediaType: "text/html", Bytes: []byte("retained notice")},
		{URL: "https://comune.example/order.pdf", Role: "attachment", Required: true, SourceID: "municipal", Configuration: 1, MediaType: "application/pdf", Bytes: []byte("synthetic ordinance")},
	}}
	docs := documents.New(pool, objects)
	v, err := docs.Retain(ctx, acquisition)
	must(err)
	acquisition.ID = "second"
	acquisition.Resources[0].Bytes = []byte("changed notice")
	_, err = docs.Retain(ctx, acquisition)
	must(err)
	acquisition.ID = "staged"
	acquisition.Resources[0].Bytes = []byte("staged bytes survive in database")
	must(docs.Stage(ctx, acquisition))
	_, err = pool.Exec(ctx, `INSERT INTO processing_configuration_versions(id,name,stage,revision,logic_version,settings,content_hash,created_at) VALUES('fixture','fixture','extraction','1','1','{}',repeat('a',64),now());`)
	must(err)
	_, err = pool.Exec(ctx, `INSERT INTO processing_runs(idempotency_key,request_hash,workload,stage,configuration_version_id,source_id,document_version_id,subject,created_at) VALUES('fixture',repeat('b',64),'evaluation','extraction','fixture','municipal',$1,'{}',now())`, v.ID)
	must(err)
	dest, err := NewS3Destination(c)
	must(err)
	must(dest.client.MakeBucket(ctx, c.Bucket, minio.MakeBucketOptions{Region: c.Region}))
	// Add a new version after export: both inventory and real pg_dump must remain
	// on the earlier snapshot. Cleanup cannot acquire its exclusive guard.
	dump := dockerDump{name: name, after: func() {
		lockCtx, stop := context.WithTimeout(ctx, 150*time.Millisecond)
		defer stop()
		unlock, e := evidenceguard.Lock(lockCtx, pool, false)
		if e == nil {
			unlock()
			t.Fatal("cleanup not blocked during snapshot")
		}
		acquisition.ID = "later"
		acquisition.Resources[0].Bytes = []byte("post snapshot")
		_, e = docs.Retain(ctx, acquisition)
		must(e)
	}}
	id := strings.Repeat("1", 32)
	b, cleanup, err := Build(ctx, pool, objects, dump, c.ScratchDirectory, id)
	must(err)
	defer cleanup()
	if len(b.Manifest.Objects) != 3 {
		t.Fatalf("snapshot inventory: %d", len(b.Manifest.Objects))
	}
	// Lock released after local capture, before potentially slow network transfer.
	unlock, err := evidenceguard.Lock(ctx, pool, false)
	must(err)
	unlock()
	must(dest.Publish(ctx, b, time.Now()))
	remote, err := dest.client.GetObject(ctx, c.Bucket, dest.key(id, ".tar"), minio.GetObjectOptions{})
	must(err)
	tr := tar.NewReader(remote)
	var db []byte
	count := 0
	for {
		h, e := tr.Next()
		if e == io.EOF {
			break
		}
		must(e)
		raw, e := io.ReadAll(tr)
		must(e)
		switch {
		case h.Name == "postgres.dump":
			db = raw
		case strings.HasPrefix(h.Name, "objects/"):
			count++
		case h.Name == "manifest.json":
			var manifest Manifest
			must(json.Unmarshal(raw, &manifest))
			if manifest.ID != id {
				t.Fatal("manifest mismatch")
			}
		}
	}
	must(remote.Close())
	if count != 3 || len(db) == 0 {
		t.Fatal("remote archive incomplete")
	}
	// Structural recovery check for the generated artifact. The operational
	// restoration rehearsal, including job semantics, remains separate task 8.3.
	must(exec.CommandContext(ctx, "docker", "exec", name, "createdb", "-U", "iwa", "restored").Run())
	restore := exec.CommandContext(ctx, "docker", "exec", "-i", name, "pg_restore", "-U", "iwa", "-d", "restored", "--exit-on-error", "--no-owner", "--no-acl")
	restore.Stdin = bytes.NewReader(db)
	must(restore.Run())
	check := exec.CommandContext(ctx, "docker", "exec", name, "psql", "-U", "iwa", "-d", "restored", "-Atc", `SELECT (SELECT count(*) FROM retained_versions),(SELECT count(*) FROM retained_objects),(SELECT count(*) FROM retained_acquisitions WHERE payload IS NOT NULL),(SELECT count(*) FROM processing_runs p JOIN retained_versions v ON v.id=p.document_version_id JOIN processing_configuration_versions c ON c.id=p.configuration_version_id),(SELECT count(*) FROM retained_resources r JOIN registry_configurations c ON (c.source_id,c.revision)=(r.source_id,r.configuration));`)
	out, err := check.Output()
	must(err)
	if strings.TrimSpace(string(out)) != "2|3|1|1|4" {
		t.Fatalf("recovered graph: %s", out)
	}
	// Missing or corrupt originals cannot produce a successful bundle.
	must(objects.Delete(ctx, v.Resources[0].Hash))
	_, cleanBad, err := Build(ctx, pool, objects, dockerDump{name: name}, c.ScratchDirectory, strings.Repeat("2", 32))
	cleanBad()
	if err == nil {
		t.Fatal("missing original accepted")
	}
	must(objects.Ensure(ctx, v.Resources[0].Hash, []byte("retained notice")))
	w := &Worker{Store: New(pool), Config: c, Objects: objects, Dump: dockerDump{name: name}, Destination: brokenDestination{Destination: dest, failPublish: true}}
	now := time.Now().UTC()
	worked, err := w.Tick(ctx, now)
	if !worked || err == nil {
		t.Fatal("transfer failure not reported")
	}
	report, err := w.Store.Report(ctx)
	must(err)
	if report.Total != 1 || report.Runs[0].ErrorCode != "transfer_failed" {
		t.Fatal("failure not persisted")
	}
	notificationReport, err := notifications.New(pool).Report(ctx)
	must(err)
	if notificationReport.OpenIncidents != 1 || notificationReport.TotalMessages != 1 {
		t.Fatal("backup failure did not reach incident outbox")
	}
	w.Destination = dest
	// Re-deliver an unacknowledged result after a real database restart. The
	// notification hook must keep the same incident/message identity.
	_, err = pool.Exec(ctx, `UPDATE backup_runs SET notified=false`)
	must(err)
	must(exec.CommandContext(ctx, "docker", "restart", "--time", "5", name).Run())
	deadline := time.Now().Add(20 * time.Second)
	for pool.Ping(ctx) != nil {
		if time.Now().After(deadline) {
			t.Fatal("database restart timed out")
		}
		time.Sleep(100 * time.Millisecond)
	}
	worked, err = w.Tick(ctx, now.Add(time.Hour))
	must(err)
	if worked {
		t.Fatal("restart lost schedule")
	}
	// A changed interval applies immediately; concurrent schedulers run only once.
	w.Config.IntervalSeconds = 60
	var wg sync.WaitGroup
	var mu sync.Mutex
	workedCount := 0
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worked, e := w.Tick(ctx, now.Add(2*time.Minute))
			if e != nil {
				t.Error(e)
			}
			if worked {
				mu.Lock()
				workedCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if workedCount != 1 {
		t.Fatal("concurrent backup duplicated")
	}
	notificationReport, err = notifications.New(pool).Report(ctx)
	must(err)
	if notificationReport.OpenIncidents != 0 || notificationReport.TotalMessages != 2 {
		t.Fatal("successful backup did not recover incident")
	}
	// Interrupted state becomes visible and emits a failure even with backup disabled.
	_, err = pool.Exec(ctx, `INSERT INTO backup_runs(id,started_at,next_due_at,state) VALUES('interrupted',now(),now()+interval '1 day','running')`)
	must(err)
	w.Config.Enabled = false
	_, err = w.Tick(ctx, now.Add(3*time.Minute))
	must(err)
	var state, code string
	must(pool.QueryRow(ctx, `SELECT state,error_code FROM backup_runs WHERE id='interrupted'`).Scan(&state, &code))
	if state != "failed" || code != "interrupted" {
		t.Fatal("crash not recovered")
	}
	// A completed transfer with failed retention is still an operational failure.
	w.Config.Enabled = true
	w.Destination = brokenDestination{Destination: dest, failPrune: true}
	worked, err = w.Tick(ctx, now.Add(4*time.Minute))
	if !worked || err == nil {
		t.Fatal("retention failure hidden")
	}
	report, err = w.Store.Report(ctx)
	must(err)
	if report.Runs[0].ErrorCode != "retention_failed" {
		t.Fatal("retention failure not persisted")
	}
	// Delete only valid expired bundles in our namespace, preserving the current
	// bundle and unrelated keys; failure paths never prune old backups.
	old, cleanOld, err := Build(ctx, pool, objects, dockerDump{name: name}, c.ScratchDirectory, strings.Repeat("3", 32))
	must(err)
	defer cleanOld()
	must(dest.Publish(ctx, old, now.Add(-40*24*time.Hour)))
	_, err = dest.client.PutObject(ctx, c.Bucket, "unrelated/keep", strings.NewReader("keep"), 4, minio.PutObjectOptions{})
	must(err)
	must(dest.Prune(ctx, now.Add(-30*24*time.Hour), id))
	if _, err = dest.client.StatObject(ctx, c.Bucket, dest.key(old.Manifest.ID, ".tar"), minio.StatObjectOptions{}); err == nil {
		t.Fatal("expired bundle retained")
	}
	_, err = dest.client.StatObject(ctx, c.Bucket, dest.key(id, ".tar"), minio.StatObjectOptions{})
	must(err)
	_, err = dest.client.StatObject(ctx, c.Bucket, "unrelated/keep", minio.StatObjectOptions{})
	must(err)
}

// Exercise the shipped executable and real pg_dump client when its image has
// been built explicitly. Never pulls images or touches the Compose volumes.
func TestPackagedBackupWorker(t *testing.T) {
	image := os.Getenv("IWA_BACKUP_TEST_IMAGE")
	if image == "" {
		t.Skip("set IWA_BACKUP_TEST_IMAGE after building the application image")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool, _ := testDB(t)
	objects, source := testObjects(t)
	_, c := testObjects(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, notifications.Migrate, Migrate} {
		must(m(ctx, pool))
	}
	dest, err := NewS3Destination(c)
	must(err)
	must(dest.client.MakeBucket(ctx, c.Bucket, minio.MakeBucketOptions{Region: c.Region}))
	_, err = pool.Exec(ctx, `INSERT INTO registry_authorities VALUES('fixture','Fixture','https://example.test');
 INSERT INTO registry_channels VALUES('fixture','fixture','Fixture','https://example.test',false);
 INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES('fixture','fixture','fixture','municipal','050004');
 INSERT INTO registry_configurations(source_id,revision,body,actor) VALUES('fixture',1,'{"policy":{"collection_permitted":true,"retention_permitted":true,"evidence":{"url":"https://example.test"}}}','test');`)
	must(err)
	// Directly retain one immutable resource plus its references in this fixture.
	raw := []byte("packaged worker original")
	hash := sha256.Sum256(raw)
	digest := hex.EncodeToString(hash[:])
	must(objects.Ensure(ctx, digest, raw))
	_, err = pool.Exec(ctx, `INSERT INTO retained_objects VALUES($1,$2,$3)`, digest, "sha256/"+digest[:2]+"/"+digest, len(raw))
	must(err)
	var doc, version int64
	must(pool.QueryRow(ctx, `INSERT INTO retained_documents(source_id,official_url) VALUES('fixture','https://example.test/notice') RETURNING id`).Scan(&doc))
	must(pool.QueryRow(ctx, `INSERT INTO retained_versions(document_id,content_hash,first_acquired_at,complete,metadata) VALUES($1,$2,now(),true,'{}') RETURNING id`, doc, digest).Scan(&version))
	_, err = pool.Exec(ctx, `INSERT INTO retained_resources VALUES($1,'https://example.test/notice','original',true,'fixture',1,'text/plain',$2,'')`, version, digest)
	must(err)
	dir := t.TempDir()
	must(os.Chmod(dir, 0755))
	c.ScratchDirectory = "/tmp"
	cfg, err := json.Marshal(c)
	must(err)
	pc := pool.Config().ConnConfig
	for filename, raw := range map[string][]byte{"db": []byte(pc.Password), "crawl": []byte("unused-fixture-token"), "access": []byte(source.AccessKey), "secret": []byte(source.SecretKey), "backup.json": cfg} {
		must(os.WriteFile(filepath.Join(dir, filename), raw, 0644))
	}
	name := "iwa-packaged-backup-" + strings.TrimPrefix(filepath.Base(dir), "0") + strconv.FormatInt(time.Now().UnixNano(), 10)
	cmd := exec.CommandContext(ctx, "docker", "run", "--pull=never", "-d", "--name", name, "--network", "host", "--read-only", "--tmpfs", "/tmp:rw,noexec,nosuid,nodev,size=128m", "--mount", "type=bind,src="+dir+",dst=/run/secrets,readonly",
		"-e", "IWA_ROLE=worker", "-e", "IWA_POSTGRES_HOST="+net.JoinHostPort(pc.Host, strconv.Itoa(int(pc.Port))), "-e", "IWA_POSTGRES_PASSWORD_FILE=/run/secrets/db", "-e", "IWA_CRAWL_TOKEN_FILE=/run/secrets/crawl", "-e", "IWA_RUSTFS_URL="+source.Endpoint, "-e", "IWA_RUSTFS_BUCKET=iwa-test-originals", "-e", "IWA_RUSTFS_ACCESS_KEY_FILE=/run/secrets/access", "-e", "IWA_RUSTFS_SECRET_KEY_FILE=/run/secrets/secret", "-e", "IWA_BACKUPS_CONFIG_FILE=/run/secrets/backup.json", image, "backup-worker")
	must(cmd.Run())
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", name).Run() })
	deadline := time.Now().Add(30 * time.Second)
	for {
		r, e := New(pool).Report(ctx)
		must(e)
		if len(r.Runs) > 0 && r.Runs[0].State != "running" {
			if r.Runs[0].State != "succeeded" || r.Runs[0].Objects != 1 {
				t.Fatalf("packaged worker failed: %s", r.Runs[0].ErrorCode)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("packaged backup worker did not finish")
		}
		time.Sleep(100 * time.Millisecond)
	}
	must(exec.CommandContext(ctx, "docker", "stop", "--time", "5", name).Run())
	logs, err := exec.CommandContext(ctx, "docker", "logs", name).CombinedOutput()
	must(err)
	for _, secret := range []string{pc.Password, source.SecretKey, c.SecretKey, c.Endpoint} {
		if bytes.Contains(logs, []byte(secret)) {
			t.Fatal("private backup configuration in logs")
		}
	}
}
