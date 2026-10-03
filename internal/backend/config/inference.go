package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"time"
)

const defaultRegoloChatEndpoint = "https://api.regolo.ai/v1/chat/completions"
const defaultRegoloEmbeddingEndpoint = "https://api.regolo.ai/v1/embeddings"

// InferenceEndpoint is private worker/probe configuration. API keys are loaded
// only on the code paths that call an inference provider.
type InferenceEndpoint struct {
	Adapter, URL, Model, APIKey string
}

type Inference struct {
	Qwen, OCR   InferenceEndpoint
	MaxAttempts int
	RetryBase   time.Duration
}

// LoadInference loads independently replaceable Qwen and OCR endpoints. A
// shared Regolo key file is the fallback, while per-endpoint files allow later
// providers to use distinct credentials without changing domain code.
func LoadInference() (Inference, error) {
	maxAttempts, err := boundedInt("IWA_INFERENCE_MAX_ATTEMPTS", 3, 1, 3)
	if err != nil {
		return Inference{}, err
	}
	retrySeconds, err := boundedInt("IWA_INFERENCE_RETRY_BASE_SECONDS", 2, 1, 3600)
	if err != nil {
		return Inference{}, err
	}
	sharedKeyFile := os.Getenv("IWA_REGOLO_API_KEY_FILE")
	qwen, err := loadQwen(sharedKeyFile)
	if err != nil {
		return Inference{}, err
	}
	ocr, err := loadOCR(sharedKeyFile)
	if err != nil {
		return Inference{}, err
	}
	return Inference{Qwen: qwen, OCR: ocr, MaxAttempts: maxAttempts, RetryBase: time.Duration(retrySeconds) * time.Second}, nil
}

// LoadQwen loads only the classification/extraction endpoint.
func LoadQwen() (InferenceEndpoint, error) {
	return loadQwen(os.Getenv("IWA_REGOLO_API_KEY_FILE"))
}

func loadQwen(sharedKeyFile string) (InferenceEndpoint, error) {
	return inferenceEndpoint("IWA_QWEN", "qwen3.8-27b", sharedKeyFile)
}

// LoadOCR loads only the OCR endpoint. A worker that has no Qwen handler must
// not require or read an unrelated Qwen credential.
func LoadOCR() (InferenceEndpoint, error) {
	return loadOCR(os.Getenv("IWA_REGOLO_API_KEY_FILE"))
}

func LoadEmbedding() (InferenceEndpoint, error) {
	shared := os.Getenv("IWA_REGOLO_API_KEY_FILE")
	c := InferenceEndpoint{Adapter: env("IWA_EMBEDDING_ADAPTER", "openai-embeddings"), URL: env("IWA_EMBEDDING_ENDPOINT", defaultRegoloEmbeddingEndpoint), Model: env("IWA_EMBEDDING_MODEL", "Qwen3-Embedding-8B")}
	u, err := url.Parse(c.URL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Scheme != "https" {
		return InferenceEndpoint{}, errors.New("IWA_EMBEDDING_ENDPOINT is invalid")
	}
	keyFile := os.Getenv("IWA_EMBEDDING_API_KEY_FILE")
	keyVar := "IWA_EMBEDDING_API_KEY_FILE"
	if keyFile == "" {
		keyFile = shared
		keyVar = "IWA_REGOLO_API_KEY_FILE"
	}
	if keyFile == "" {
		return InferenceEndpoint{}, errors.New(keyVar + " is required")
	}
	c.APIKey, err = secretFile(keyVar, keyFile)
	return c, err
}

func loadOCR(sharedKeyFile string) (InferenceEndpoint, error) {
	return inferenceEndpoint("IWA_OCR", "deepseek-ocr-2", sharedKeyFile)
}

func inferenceEndpoint(prefix, defaultModel, sharedKeyFile string) (InferenceEndpoint, error) {
	c := InferenceEndpoint{
		Adapter: env(prefix+"_ADAPTER", "openai-chat"),
		URL:     env(prefix+"_ENDPOINT", defaultRegoloChatEndpoint),
		Model:   env(prefix+"_MODEL", defaultModel),
	}
	u, err := url.Parse(c.URL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return InferenceEndpoint{}, errors.New(prefix + "_ENDPOINT is invalid")
	}
	if c.Adapter == "" || c.Model == "" {
		return InferenceEndpoint{}, errors.New(prefix + " adapter and model are required")
	}
	keyVariable := prefix + "_API_KEY_FILE"
	keyFile := os.Getenv(keyVariable)
	if keyFile == "" {
		keyVariable = "IWA_REGOLO_API_KEY_FILE"
		keyFile = sharedKeyFile
	}
	if keyFile == "" {
		return InferenceEndpoint{}, errors.New(keyVariable + " is required")
	}
	c.APIKey, err = secretFile(keyVariable, keyFile)
	if err != nil {
		return InferenceEndpoint{}, err
	}
	return c, nil
}

func boundedInt(key string, fallback, minimum, maximum int) (int, error) {
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
