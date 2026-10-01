package config

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"time"
)

type LocalFallback struct {
	Enabled  bool
	OCR      bool
	Endpoint InferenceEndpoint
	Sources  []string
	Timeout  time.Duration
}

// Local credentials are independent of the remote account, including the
// non-secret placeholder required by unauthenticated local chat transports.
func LoadLocalFallback() (LocalFallback, error) {
	var result LocalFallback
	switch os.Getenv("IWA_LOCAL_FALLBACK_ENABLED") {
	case "", "false":
		return result, nil
	case "true":
		result.Enabled = true
	default:
		return result, errors.New("IWA_LOCAL_FALLBACK_ENABLED is invalid")
	}
	result.Endpoint = InferenceEndpoint{Adapter: "local-openai-chat", URL: os.Getenv("IWA_LOCAL_FALLBACK_ENDPOINT"), Model: os.Getenv("IWA_LOCAL_FALLBACK_MODEL"), APIKey: "local-no-key"}
	u, err := url.Parse(result.Endpoint.URL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") || u.Path != "/v1/chat/completions" {
		return LocalFallback{}, errors.New("IWA_LOCAL_FALLBACK_ENDPOINT is invalid")
	}
	if model := result.Endpoint.Model; strings.TrimSpace(model) != model || model == "" || len(model) > 200 || strings.ContainsAny(model, "\r\n") {
		return LocalFallback{}, errors.New("IWA_LOCAL_FALLBACK_MODEL is invalid")
	}
	seen := map[string]bool{}
	for _, source := range strings.Split(os.Getenv("IWA_LOCAL_FALLBACK_SOURCES"), ",") {
		source = strings.TrimSpace(source)
		if !rolloutSource.MatchString(source) {
			return LocalFallback{}, errors.New("IWA_LOCAL_FALLBACK_SOURCES is required and must be explicit")
		}
		if !seen[source] {
			result.Sources = append(result.Sources, source)
			seen[source] = true
		}
	}
	seconds, err := boundedInt("IWA_LOCAL_FALLBACK_TIMEOUT_SECONDS", 180, 1, 600)
	if err != nil {
		return LocalFallback{}, err
	}
	result.Timeout = time.Duration(seconds) * time.Second
	switch os.Getenv("IWA_LOCAL_FALLBACK_OCR_ENABLED") {
	case "", "false":
	case "true":
		result.OCR = true
	default:
		return LocalFallback{}, errors.New("IWA_LOCAL_FALLBACK_OCR_ENABLED is invalid")
	}
	if path := os.Getenv("IWA_LOCAL_FALLBACK_API_KEY_FILE"); path != "" {
		result.Endpoint.APIKey, err = secretFile("IWA_LOCAL_FALLBACK_API_KEY_FILE", path)
		if err != nil {
			return LocalFallback{}, err
		}
	}
	return result, nil
}
