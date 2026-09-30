// Package config loads private runtime configuration without exposing its values.
package config

import (
	"errors"
	"net"
	"net/netip"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AdminTailscaleOrigin                                                   string
	Role, Listen, PostgresHost, PostgresUser, PostgresDB, PostgresPassword string
	RustFSURL, CrawlURL, CrawlToken                                        string
	PublicAllowance, PublicMaxPageSize                                     int
	PublicWindow                                                           time.Duration
	PublicTrustedProxies                                                   []netip.Prefix
	PublicViewLifetime                                                     time.Duration
	PublicCursorKey                                                        string
	PublicCopyAccess                                                       bool
	PublicBaseURL                                                          string
}

// Load accepts secrets only through files; errors never contain secret contents.
func Load() (Config, error) {
	c := Config{Role: env("IWA_ROLE", "public"), PostgresHost: env("IWA_POSTGRES_HOST", "postgres:5432"), PostgresUser: "iwa", PostgresDB: "iwa", RustFSURL: env("IWA_RUSTFS_URL", "http://rustfs:9000"), CrawlURL: env("IWA_CRAWL_URL", "http://crawl4ai:11235")}
	if c.Role != "public" && c.Role != "admin" && c.Role != "worker" {
		return Config{}, errors.New("invalid service role")
	}
	c.Listen = env("IWA_LISTEN", "127.0.0.1:8080")
	if c.Role == "admin" {
		c.Listen = env("IWA_LISTEN", "127.0.0.1:8081")
		c.AdminTailscaleOrigin = os.Getenv("IWA_ADMIN_TAILSCALE_ORIGIN")
		if c.AdminTailscaleOrigin != "" && !ValidAdminTailscaleOrigin(c.AdminTailscaleOrigin) {
			return Config{}, errors.New("IWA_ADMIN_TAILSCALE_ORIGIN must be an exact HTTPS ts.net origin")
		}
		host, _, err := net.SplitHostPort(c.Listen)
		ip := net.ParseIP(host)
		if err != nil || (env("IWA_CONTAINER_ADMIN", "false") != "true" && (ip == nil || !ip.IsLoopback())) {
			return Config{}, errors.New("administration requires a loopback listener or explicit container isolation")
		}
	}
	for _, raw := range []string{c.RustFSURL, c.CrawlURL} {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
			return Config{}, errors.New("invalid dependency URL")
		}
	}
	var err error
	c.PublicAllowance, err = configInt("IWA_PUBLIC_RATE_ALLOWANCE", 120, 1, 1_000_000)
	if err != nil {
		return Config{}, err
	}
	windowSeconds, err := configInt("IWA_PUBLIC_RATE_WINDOW_SECONDS", 60, 1, 86_400)
	if err != nil {
		return Config{}, err
	}
	c.PublicWindow = time.Duration(windowSeconds) * time.Second
	c.PublicMaxPageSize, err = configInt("IWA_PUBLIC_MAX_PAGE_SIZE", 100, 1, 1_000)
	if err != nil {
		return Config{}, err
	}
	c.PublicTrustedProxies, err = trustedProxies(os.Getenv("IWA_PUBLIC_TRUSTED_PROXIES"))
	if err != nil {
		return Config{}, err
	}
	viewSeconds, err := configInt("IWA_PUBLIC_VIEW_TTL_SECONDS", 1_800, 60, 86_400)
	if err != nil {
		return Config{}, err
	}
	c.PublicViewLifetime = time.Duration(viewSeconds) * time.Second
	if c.Role == "public" {
		c.PublicCursorKey, err = secret("IWA_PUBLIC_CURSOR_KEY_FILE")
		if err != nil {
			return Config{}, err
		}
		if len(c.PublicCursorKey) < 32 {
			return Config{}, errors.New("IWA_PUBLIC_CURSOR_KEY_FILE must contain at least 32 bytes")
		}
	}
	c.PublicCopyAccess, err = configBool("IWA_PUBLIC_COPY_ACCESS_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	c.PublicBaseURL = strings.TrimRight(os.Getenv("IWA_PUBLIC_BASE_URL"), "/")
	if c.PublicCopyAccess {
		u, parseErr := url.Parse(c.PublicBaseURL)
		if parseErr != nil || u.Scheme == "" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || (u.Scheme != "http" && u.Scheme != "https") {
			return Config{}, errors.New("IWA_PUBLIC_BASE_URL is invalid")
		}
	}
	c.PostgresPassword, err = secret("IWA_POSTGRES_PASSWORD_FILE")
	if err != nil {
		return Config{}, err
	}
	c.CrawlToken, err = secret("IWA_CRAWL_TOKEN_FILE")
	if err != nil {
		return Config{}, err
	}
	return c, nil
}

func configBool(key string, fallback bool) (bool, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil || (raw != "true" && raw != "false") {
		return false, errors.New(key + " is invalid")
	}
	return value, nil
}

func configInt(key string, fallback, minimum, maximum int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < minimum || value > maximum {
		return 0, errors.New(key + " is invalid")
	}
	return value, nil
}

func trustedProxies(raw string) ([]netip.Prefix, error) {
	if strings.TrimSpace(raw) == "" {
		return []netip.Prefix{}, nil
	}
	items := strings.Split(raw, ",")
	result := make([]netip.Prefix, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		prefix, err := netip.ParsePrefix(item)
		if err != nil {
			address, addressErr := netip.ParseAddr(item)
			if addressErr != nil {
				return nil, errors.New("IWA_PUBLIC_TRUSTED_PROXIES is invalid")
			}
			prefix = netip.PrefixFrom(address, address.BitLen())
		}
		result = append(result, prefix.Masked())
	}
	return result, nil
}
func env(key, fallback string) string {
	if s := os.Getenv(key); s != "" {
		return s
	}
	return fallback
}
func secret(key string) (string, error) {
	path := os.Getenv(key)
	if path == "" {
		return "", errors.New(key + " is required")
	}
	return secretFile(key, path)
}

func secretFile(label, path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New(label + " cannot be read")
	}
	s := strings.TrimSpace(string(b))
	if s == "" || strings.ContainsAny(s, "\r\n") {
		return "", errors.New(label + " must contain one nonempty line")
	}
	return s, nil
}

// ValidAdminTailscaleOrigin permits one canonical Serve HTTPS origin, never a wildcard.
func ValidAdminTailscaleOrigin(origin string) bool {
	return regexp.MustCompile(`^https://[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.ts\.net$`).MatchString(origin)
}
