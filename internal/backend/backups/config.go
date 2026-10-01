// Package backups creates consistent PostgreSQL and retained-object bundles.
package backups

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

var ErrConfig = errors.New("invalid private backup configuration")
var ErrBackup = errors.New("backup operation failed")
var prefixPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,79}$`)

type Config struct {
	Enabled          bool   `json:"enabled"`
	IntervalSeconds  int    `json:"interval_seconds"`
	RetentionDays    int    `json:"retention_days"`
	TimeoutSeconds   int    `json:"timeout_seconds"`
	ScratchDirectory string `json:"scratch_directory"`
	Endpoint         string `json:"endpoint"`
	Bucket           string `json:"bucket"`
	Prefix           string `json:"prefix"`
	Region           string `json:"region"`
	AccessKey        string `json:"access_key"`
	SecretKey        string `json:"secret_key"`
	// Only explicit loopback integration fixtures can use cleartext transport.
	LoopbackTest bool `json:"loopback_test"`
}

func LoadConfig(path string) (Config, error) {
	if path == "" {
		return Config{}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return Config{}, ErrConfig
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, 65537))
	if err != nil || len(raw) > 65536 {
		return Config{}, ErrConfig
	}
	c := Config{IntervalSeconds: 86400, RetentionDays: 30, TimeoutSeconds: 3600, ScratchDirectory: "/tmp", Prefix: "iwa-backups", Region: "us-east-1"}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&c) != nil || d.Decode(new(any)) != io.EOF || c.Validate() != nil {
		return Config{}, ErrConfig
	}
	return c, nil
}
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	u, err := url.Parse(c.Endpoint)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return ErrConfig
	}
	if c.LoopbackTest {
		ip := net.ParseIP(u.Hostname())
		if ip == nil || !ip.IsLoopback() || u.Scheme != "http" {
			return ErrConfig
		}
	} else if u.Scheme != "https" {
		return ErrConfig
	}
	if c.IntervalSeconds < 60 || c.IntervalSeconds > 2592000 || c.RetentionDays < 1 || c.RetentionDays > 3650 || c.TimeoutSeconds < 10 || c.TimeoutSeconds > 86400 || !filepath.IsAbs(c.ScratchDirectory) || !prefixPattern.MatchString(c.Prefix) || c.Bucket == "" || c.AccessKey == "" || c.SecretKey == "" || c.Region == "" {
		return ErrConfig
	}
	return nil
}
