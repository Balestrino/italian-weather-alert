package backups

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDestinationRejectsCorruptReadback(t *testing.T) {
	var marked atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".complete.json") {
			marked.Store(true)
		}
		if r.Method == "PUT" {
			_, _ = io.Copy(io.Discard, r.Body)
			w.Header().Set("ETag", `"fixture"`)
			w.WriteHeader(200)
			return
		}
		w.Header().Set("Content-Length", "7")
		_, _ = w.Write([]byte("corrupt"))
	}))
	defer srv.Close()
	c := Config{Enabled: true, Endpoint: srv.URL, Bucket: "fixture", Prefix: "test", AccessKey: "a", SecretKey: "b", Region: "us-east-1", LoopbackTest: true, IntervalSeconds: 86400, RetentionDays: 30, TimeoutSeconds: 60, ScratchDirectory: t.TempDir()}
	dest, err := NewS3Destination(c)
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte("correct")
	h := sha256.Sum256(raw)
	path := filepath.Join(t.TempDir(), "archive")
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	b := Bundle{Path: path, SHA256: hex.EncodeToString(h[:]), Size: 7, Manifest: Manifest{ID: strings.Repeat("a", 32)}}
	if err = dest.Publish(context.Background(), b, time.Now()); err == nil || marked.Load() {
		t.Fatal("corrupt readback marked complete")
	}
}
