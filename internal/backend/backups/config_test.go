package backups

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrivateConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.json")
	valid := `{"enabled":true,"endpoint":"https://backup.example.test","bucket":"private","access_key":"secret-id","secret_key":"secret-value"}`
	if err := os.WriteFile(path, []byte(valid), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := LoadConfig(path)
	if err != nil || c.IntervalSeconds != 86400 || c.RetentionDays != 30 || c.TimeoutSeconds != 3600 {
		t.Fatalf("defaults: %+v", err)
	}
	for _, raw := range []string{valid + ` {}`, `{"enabled":false,"unknown":true}`, `{"enabled":true}`, `{"enabled":true,"endpoint":"http://backup.example.test","bucket":"b","access_key":"a","secret_key":"s"}`} {
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadConfig(path); err != ErrConfig {
			t.Fatal("invalid configuration accepted")
		}
	}
	for _, change := range []func(*Config){func(c *Config) { c.Prefix = "../foreign" }, func(c *Config) { c.RetentionDays = 0 }, func(c *Config) { c.IntervalSeconds = 0 }, func(c *Config) { c.Endpoint = "https://user:password@host" }, func(c *Config) { c.Endpoint = "http://localhost:9000"; c.LoopbackTest = true }} {
		bad := c
		change(&bad)
		if bad.Validate() != ErrConfig {
			t.Fatal("invalid option accepted")
		}
	}
	c.Endpoint = "http://127.0.0.1:9000"
	c.LoopbackTest = true
	if c.Validate() != nil {
		t.Fatal("explicit loopback fixture rejected")
	}
	if c, err := LoadConfig(""); err != nil || c.Enabled {
		t.Fatal("backup should default disabled")
	}
}
