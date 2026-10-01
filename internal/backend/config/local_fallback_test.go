package config

import (
	"testing"
	"time"
)

func TestLocalFallbackIsExplicitAndCredentialIsolated(t *testing.T) {
	t.Setenv("IWA_LOCAL_FALLBACK_ENABLED", "false")
	t.Setenv("IWA_LOCAL_FALLBACK_ENDPOINT", "invalid")
	if cfg, err := LoadLocalFallback(); err != nil || cfg.Enabled {
		t.Fatal("disabled fallback reads endpoint", err)
	}
	t.Setenv("IWA_LOCAL_FALLBACK_ENABLED", "true")
	t.Setenv("IWA_REGOLO_API_KEY_FILE", "/missing/remote-key")
	t.Setenv("IWA_LOCAL_FALLBACK_ENDPOINT", "http://127.0.0.1:8080/v1/chat/completions")
	t.Setenv("IWA_LOCAL_FALLBACK_MODEL", "qwen3.5-test")
	t.Setenv("IWA_LOCAL_FALLBACK_SOURCES", "fixture,fixture")
	cfg, err := LoadLocalFallback()
	if err != nil || cfg.Endpoint.APIKey != "local-no-key" || cfg.Endpoint.Adapter != "local-openai-chat" || len(cfg.Sources) != 1 || cfg.Timeout != 180*time.Second || cfg.OCR {
		t.Fatalf("independent local configuration: %+v %v", cfg, err)
	}
	for _, test := range []struct{ key, value string }{
		{"IWA_LOCAL_FALLBACK_ENDPOINT", "http://user:secret@localhost/v1/chat/completions"},
		{"IWA_LOCAL_FALLBACK_ENDPOINT", "http://localhost/v1/chat/completions?key=secret"},
		{"IWA_LOCAL_FALLBACK_ENDPOINT", "http://localhost/"},
		{"IWA_LOCAL_FALLBACK_MODEL", "model\nsecret"},
		{"IWA_LOCAL_FALLBACK_SOURCES", ""},
		{"IWA_LOCAL_FALLBACK_SOURCES", "*"},
		{"IWA_LOCAL_FALLBACK_TIMEOUT_SECONDS", "601"},
		{"IWA_LOCAL_FALLBACK_OCR_ENABLED", "perhaps"},
	} {
		t.Run(test.key+test.value, func(t *testing.T) {
			t.Setenv(test.key, test.value)
			if _, err := LoadLocalFallback(); err == nil {
				t.Fatal("unsafe or ambiguous setting accepted")
			}
		})
	}
}
