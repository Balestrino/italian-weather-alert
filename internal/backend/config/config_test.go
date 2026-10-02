package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setup(t *testing.T) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(p, []byte("private-fixture-value\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("IWA_POSTGRES_PASSWORD_FILE", p)
	t.Setenv("IWA_CRAWL_TOKEN_FILE", p)
	cursorKey := filepath.Join(t.TempDir(), "cursor-key")
	if err := os.WriteFile(cursorKey, []byte(strings.Repeat("c", 64)), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("IWA_ROLE", "public")
	t.Setenv("IWA_ENVIRONMENT", "")
	t.Setenv("IWA_LISTEN", "")
	t.Setenv("IWA_CONTAINER_ADMIN", "")
	t.Setenv("IWA_RUSTFS_URL", "")
	t.Setenv("IWA_CRAWL_URL", "")
	t.Setenv("IWA_PUBLIC_RATE_ALLOWANCE", "")
	t.Setenv("IWA_PUBLIC_RATE_WINDOW_SECONDS", "")
	t.Setenv("IWA_PUBLIC_MAX_PAGE_SIZE", "")
	t.Setenv("IWA_PUBLIC_TRUSTED_PROXIES", "")
	t.Setenv("IWA_PUBLIC_VIEW_TTL_SECONDS", "")
	t.Setenv("IWA_PUBLIC_CURSOR_KEY_FILE", cursorKey)
	t.Setenv("IWA_PUBLIC_COPY_ACCESS_ENABLED", "")
	t.Setenv("IWA_PUBLIC_BASE_URL", "")
}

func TestPublicAccessConfiguration(t *testing.T) {
	setup(t)
	c, err := Load()
	if err != nil || c.PublicAllowance != 120 || c.PublicWindow.Seconds() != 60 || c.PublicMaxPageSize != 100 || len(c.PublicTrustedProxies) != 0 || c.PublicViewLifetime != 30*time.Minute || c.PublicCopyAccess {
		t.Fatalf("unexpected defaults: %#v %v", c, err)
	}
	t.Setenv("IWA_PUBLIC_RATE_ALLOWANCE", "17")
	t.Setenv("IWA_PUBLIC_RATE_WINDOW_SECONDS", "31")
	t.Setenv("IWA_PUBLIC_MAX_PAGE_SIZE", "23")
	t.Setenv("IWA_PUBLIC_TRUSTED_PROXIES", "10.0.0.0/8, 2001:db8::1")
	t.Setenv("IWA_PUBLIC_VIEW_TTL_SECONDS", "900")
	c, err = Load()
	if err != nil || c.PublicAllowance != 17 || c.PublicWindow.Seconds() != 31 || c.PublicMaxPageSize != 23 || len(c.PublicTrustedProxies) != 2 || c.PublicViewLifetime != 15*time.Minute {
		t.Fatalf("valid public limits rejected: %#v %v", c, err)
	}
	for key, value := range map[string]string{
		"IWA_PUBLIC_RATE_ALLOWANCE":      "0",
		"IWA_PUBLIC_RATE_WINDOW_SECONDS": "bad",
		"IWA_PUBLIC_MAX_PAGE_SIZE":       "1001",
		"IWA_PUBLIC_TRUSTED_PROXIES":     "not-an-address",
		"IWA_PUBLIC_VIEW_TTL_SECONDS":    "59",
	} {
		setup(t)
		t.Setenv(key, value)
		if _, err := Load(); err == nil {
			t.Errorf("accepted %s=%q", key, value)
		}
	}
	setup(t)
	short := filepath.Join(t.TempDir(), "short-cursor-key")
	if err := os.WriteFile(short, []byte("short"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("IWA_PUBLIC_CURSOR_KEY_FILE", short)
	if _, err := Load(); err == nil {
		t.Fatal("short public cursor key accepted")
	}
	setup(t)
	t.Setenv("IWA_PUBLIC_COPY_ACCESS_ENABLED", "true")
	t.Setenv("IWA_PUBLIC_BASE_URL", "https://alerts.example")
	c, err = Load()
	if err != nil || !c.PublicCopyAccess || c.PublicBaseURL != "https://alerts.example" {
		t.Fatalf("valid copy access config rejected: enabled=%v base=%q err=%v", c.PublicCopyAccess, c.PublicBaseURL, err)
	}
	for _, base := range []string{"", "ftp://alerts.example", "https://user:secret@alerts.example", "https://alerts.example/path"} {
		setup(t)
		t.Setenv("IWA_PUBLIC_COPY_ACCESS_ENABLED", "true")
		t.Setenv("IWA_PUBLIC_BASE_URL", base)
		if _, err := Load(); err == nil {
			t.Errorf("invalid public base URL accepted: %q", base)
		}
	}
	setup(t)
	t.Setenv("IWA_PUBLIC_COPY_ACCESS_ENABLED", "yes")
	if _, err := Load(); err == nil {
		t.Fatal("invalid copy access boolean accepted")
	}
}
func TestPrivateConfiguration(t *testing.T) {
	setup(t)
	c, err := Load()
	if err != nil || c.PostgresPassword != "private-fixture-value" {
		t.Fatal("valid file configuration rejected")
	}
	t.Setenv("IWA_CRAWL_TOKEN_FILE", "/missing/private-fixture-value")
	_, err = Load()
	if err == nil || strings.Contains(err.Error(), "private-fixture-value") {
		t.Fatal("secret path exposed or missing file accepted")
	}
}
func TestAdminBoundary(t *testing.T) {
	setup(t)
	t.Setenv("IWA_ROLE", "admin")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	for _, listen := range []string{"127.0.0.1:8081", "[::1]:8081"} {
		t.Setenv("IWA_LISTEN", listen)
		if _, err := Load(); err != nil {
			t.Fatalf("loopback admin rejected: %s", listen)
		}
	}
	for _, listen := range []string{"0.0.0.0:8081", "[::]:8081", ":8081", "192.0.2.10:8081", "example.test:8081", "localhost:8081", "invalid"} {
		t.Setenv("IWA_LISTEN", listen)
		if _, err := Load(); err == nil {
			t.Fatalf("nonliteral/nonlocal native admin accepted: %s", listen)
		}
	}
	t.Setenv("IWA_LISTEN", "0.0.0.0:8081")
	t.Setenv("IWA_CONTAINER_ADMIN", "true")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerRole(t *testing.T) {
	setup(t)
	t.Setenv("IWA_ROLE", "worker")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}
func TestRejectCredentialURLsAndEmptySecrets(t *testing.T) {
	setup(t)
	t.Setenv("IWA_CRAWL_URL", "http://user:private-fixture-value@host")
	_, err := Load()
	if err == nil || strings.Contains(err.Error(), "private-fixture-value") {
		t.Fatal("credential URL accepted or exposed")
	}
	t.Setenv("IWA_CRAWL_URL", "")
	p := filepath.Join(t.TempDir(), "empty")
	_ = os.WriteFile(p, []byte("\n"), 0600)
	t.Setenv("IWA_CRAWL_TOKEN_FILE", p)
	if _, err := Load(); err == nil {
		t.Fatal("empty token accepted")
	}
}

func TestStorageSecretsAreSeparate(t *testing.T) {
	setup(t)
	t.Setenv("IWA_RUSTFS_ACCESS_KEY_FILE", "")
	t.Setenv("IWA_RUSTFS_SECRET_KEY_FILE", "")
	if _, err := Load(); err != nil {
		t.Fatal("public runtime requires storage secrets")
	}
	if _, err := LoadStorage(); err == nil {
		t.Fatal("storage accepted missing credentials")
	}
	path := filepath.Join(t.TempDir(), "key")
	if err := os.WriteFile(path, []byte("private-storage-fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("IWA_RUSTFS_ACCESS_KEY_FILE", path)
	t.Setenv("IWA_RUSTFS_SECRET_KEY_FILE", path)
	t.Setenv("IWA_RUSTFS_BUCKET", "")
	c, err := LoadStorage()
	if err != nil || c.Bucket != "iwa-originals" || c.SecretKey != "private-storage-fixture" {
		t.Fatal("storage configuration rejected")
	}
	t.Setenv("IWA_RUSTFS_SECRET_KEY_FILE", "/private-storage-fixture/missing")
	if _, err = LoadStorage(); err == nil || strings.Contains(err.Error(), "private-storage-fixture") {
		t.Fatal("storage secret path exposed")
	}
}

func TestAdminTailscaleOrigin(t *testing.T) {
	setup(t)
	t.Setenv("IWA_ROLE", "admin")
	for _, origin := range []string{"", "https://iwa.tail123.ts.net"} {
		t.Setenv("IWA_ADMIN_TAILSCALE_ORIGIN", origin)
		c, err := Load()
		if err != nil || c.AdminTailscaleOrigin != origin {
			t.Fatalf("valid origin rejected: %v", err)
		}
	}
	for _, origin := range []string{"http://iwa.tail123.ts.net", "https://*.tail123.ts.net", "https://iwa.tail123.ts.net/", "https://user@iwa.tail123.ts.net", "https://iwa.tail123.ts.net?", "https://iwa.tail123.ts.net#", "https://iwa.tail123.ts.net.evil.example", "https://iwa.tail123.ts.net:443", "https://localhost", "https://iwa.tail123.ts.net\n"} {
		t.Setenv("IWA_ADMIN_TAILSCALE_ORIGIN", origin)
		if _, err := Load(); err == nil {
			t.Errorf("invalid origin accepted: %q", origin)
		}
	}
}

func TestEnvironmentDefaultsStrict(t *testing.T) {
	setup(t)
	c, err := Load()
	if err != nil || c.Environment != "production" {
		t.Fatal("strict default", err)
	}
	for _, mode := range []string{"development", "staging", "production"} {
		t.Setenv("IWA_ENVIRONMENT", mode)
		c, err = Load()
		if err != nil || c.Environment != mode {
			t.Fatal(mode, err)
		}
		expected := "127.0.0.1:8080"
		if mode == "development" {
			expected = "0.0.0.0:8080"
		}
		if c.Listen != expected {
			t.Fatalf("%s public listener: got %q want %q", mode, c.Listen, expected)
		}
		t.Setenv("IWA_ROLE", "admin")
		c, err = Load()
		if err != nil || c.Listen != "127.0.0.1:8081" {
			t.Fatalf("%s admin listener: %q %v", mode, c.Listen, err)
		}
		t.Setenv("IWA_ROLE", "public")
	}
	t.Setenv("IWA_ENVIRONMENT", "dev")
	if _, err = Load(); err == nil {
		t.Fatal("invalid environment accepted")
	}
}
