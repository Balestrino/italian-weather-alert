package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func inferenceSecret(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "regolo-key")
	if err := os.WriteFile(path, []byte("private-inference-fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestInferenceConfigurationUsesPrivateFilesAndSeparateEndpoints(t *testing.T) {
	t.Setenv("IWA_REGOLO_API_KEY_FILE", inferenceSecret(t))
	t.Setenv("IWA_QWEN_ENDPOINT", "https://qwen.example/v1/chat/completions")
	t.Setenv("IWA_OCR_ENDPOINT", "http://local-ocr:8080/v1/chat/completions")
	t.Setenv("IWA_OCR_ADAPTER", "local-chat")
	t.Setenv("IWA_INFERENCE_MAX_ATTEMPTS", "2")
	t.Setenv("IWA_INFERENCE_RETRY_BASE_SECONDS", "7")

	cfg, err := LoadInference()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Qwen.Model != "qwen3.8-27b" || cfg.OCR.Model != "deepseek-ocr-2" || cfg.OCR.Adapter != "local-chat" || cfg.Qwen.APIKey != "private-inference-fixture" || cfg.MaxAttempts != 2 || cfg.RetryBase != 7*time.Second {
		t.Fatalf("unexpected configuration: %#v", cfg)
	}
}

func TestInferenceConfigurationRejectsLeaksAndUnboundedRetries(t *testing.T) {
	secret := inferenceSecret(t)
	t.Setenv("IWA_REGOLO_API_KEY_FILE", secret)
	for _, attempts := range []string{"0", "4", "not-a-number"} {
		t.Setenv("IWA_INFERENCE_MAX_ATTEMPTS", attempts)
		if _, err := LoadInference(); err == nil {
			t.Fatalf("retry limit %q accepted", attempts)
		}
	}
	t.Setenv("IWA_INFERENCE_MAX_ATTEMPTS", "3")
	t.Setenv("IWA_QWEN_ENDPOINT", "https://user:private-inference-fixture@provider.example/v1")
	if _, err := LoadInference(); err == nil || strings.Contains(err.Error(), "private-inference-fixture") {
		t.Fatal("credential URL accepted or leaked")
	}
	t.Setenv("IWA_QWEN_ENDPOINT", "https://provider.example/v1?key=private-inference-fixture")
	if _, err := LoadInference(); err == nil || strings.Contains(err.Error(), "private-inference-fixture") {
		t.Fatal("credential query accepted or leaked")
	}
	t.Setenv("IWA_QWEN_ENDPOINT", "")
	t.Setenv("IWA_REGOLO_API_KEY_FILE", "/missing/private-inference-fixture")
	if _, err := LoadInference(); err == nil || strings.Contains(err.Error(), "private-inference-fixture") {
		t.Fatal("missing secret accepted or path leaked")
	}
}

func TestInferenceEndpointSpecificSecretOverridesSharedSecret(t *testing.T) {
	shared := inferenceSecret(t)
	ocr := filepath.Join(t.TempDir(), "ocr-key")
	if err := os.WriteFile(ocr, []byte("private-ocr-fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("IWA_REGOLO_API_KEY_FILE", shared)
	t.Setenv("IWA_OCR_API_KEY_FILE", ocr)
	cfg, err := LoadInference()
	if err != nil || cfg.Qwen.APIKey != "private-inference-fixture" || cfg.OCR.APIKey != "private-ocr-fixture" {
		t.Fatal("endpoint-specific credential was not isolated")
	}
}

func TestOCRWorkerDoesNotRequireQwenCredential(t *testing.T) {
	ocr := inferenceSecret(t)
	t.Setenv("IWA_REGOLO_API_KEY_FILE", "")
	t.Setenv("IWA_QWEN_API_KEY_FILE", "")
	t.Setenv("IWA_OCR_API_KEY_FILE", ocr)
	endpoint, err := LoadOCR()
	if err != nil || endpoint.Model != "deepseek-ocr-2" || endpoint.APIKey != "private-inference-fixture" {
		t.Fatalf("independent OCR configuration rejected: %#v %v", endpoint, err)
	}
}

func TestQwenWorkerDoesNotRequireOCRCredential(t *testing.T) {
	qwen := inferenceSecret(t)
	t.Setenv("IWA_REGOLO_API_KEY_FILE", "")
	t.Setenv("IWA_OCR_API_KEY_FILE", "")
	t.Setenv("IWA_QWEN_API_KEY_FILE", qwen)
	endpoint, err := LoadQwen()
	if err != nil || endpoint.Model != "qwen3.8-27b" || endpoint.APIKey != "private-inference-fixture" {
		t.Fatalf("independent Qwen configuration rejected: %#v %v", endpoint, err)
	}
}
