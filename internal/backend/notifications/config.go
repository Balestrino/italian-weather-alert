// Package notifications groups operational failures into durable email incidents.
package notifications

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

var ErrConfig = errors.New("invalid private notification configuration")

// Config is loaded only by the worker. Never log or expose this structure.
type Config struct {
	Enabled         bool               `json:"enabled"`
	Host            string             `json:"host"`
	Port            int                `json:"port"`
	TLS             string             `json:"tls"`
	Username        string             `json:"username"`
	Password        string             `json:"password"`
	From            string             `json:"from"`
	Recipients      []string           `json:"recipients"`
	ReminderSeconds int                `json:"reminder_seconds"`
	PollSeconds     int                `json:"poll_seconds"`
	TimeoutSeconds  int                `json:"timeout_seconds"`
	DailyReport     *DailyReportConfig `json:"daily_report,omitempty"`
}

type DailyReportConfig struct {
	Enabled   bool   `json:"enabled"`
	LocalTime string `json:"local_time"`
	Timezone  string `json:"timezone"`
}

func (c DailyReportConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	t, err := time.Parse("15:04", c.LocalTime)
	if err != nil || t.Format("15:04") != c.LocalTime || t.Hour() < 18 {
		return ErrConfig
	}
	if c.Timezone == "" || c.Timezone == "Local" {
		return ErrConfig
	}
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return ErrConfig
	}
	return nil
}

// Due uses the local calendar date, so restarts and offset changes do not
// change the evening report identity or backfill previous dates.
func (c DailyReportConfig) Due(at time.Time) (string, bool) {
	if !c.Enabled || c.Validate() != nil || at.IsZero() {
		return "", false
	}
	loc, _ := time.LoadLocation(c.Timezone)
	local := at.In(loc)
	t, _ := time.Parse("15:04", c.LocalTime)
	due := time.Date(local.Year(), local.Month(), local.Day(), t.Hour(), t.Minute(), 0, 0, loc)
	return local.Format(time.DateOnly), !at.Before(due)
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
	c := Config{Port: 587, TLS: "starttls", ReminderSeconds: 21600, PollSeconds: 30, TimeoutSeconds: 15}
	raw, readErr := io.ReadAll(io.LimitReader(f, 65537))
	if readErr != nil || len(raw) > 65536 {
		return Config{}, ErrConfig
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&c) != nil || d.Decode(new(any)) != io.EOF || c.Validate() != nil {
		return Config{}, ErrConfig
	}
	return c, nil
}

func (c Config) Validate() error {
	if c.DailyReport != nil && (c.DailyReport.Validate() != nil || c.DailyReport.Enabled && !c.Enabled) {
		return ErrConfig
	}
	if !c.Enabled {
		return nil
	}
	if c.Host == "" || strings.ContainsAny(c.Host, " /\\\r\n\t@") || c.Port < 1 || c.Port > 65535 || c.PollSeconds < 1 || c.PollSeconds > 3600 || c.TimeoutSeconds < 1 || c.TimeoutSeconds > 60 || c.ReminderSeconds < 0 || c.ReminderSeconds > 2592000 {
		return ErrConfig
	}
	if c.TLS != "starttls" && c.TLS != "implicit" && c.TLS != "loopback" {
		return ErrConfig
	}
	// Plain SMTP is available only for an explicitly addressed local test relay.
	if c.TLS == "loopback" && (net.ParseIP(c.Host) == nil || !net.ParseIP(c.Host).IsLoopback() || c.Username != "" || c.Password != "") {
		return ErrConfig
	}
	if (c.Username == "") != (c.Password == "") || strings.ContainsAny(c.Username+c.Password, "\r\n\x00") || !mailbox(c.From) || len(c.Recipients) == 0 || len(c.Recipients) > 50 {
		return ErrConfig
	}
	seen := map[string]bool{}
	for _, r := range c.Recipients {
		if !mailbox(r) || seen[r] {
			return ErrConfig
		}
		seen[r] = true
	}
	return nil
}

func mailbox(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s && !strings.ContainsAny(s, "\r\n") && len(s) < 255 && strings.Contains(s, "@")
}
func (c Config) address() string         { return net.JoinHostPort(c.Host, strconv.Itoa(c.Port)) }
func (c Config) reminder() time.Duration { return time.Duration(c.ReminderSeconds) * time.Second }
