package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
)

type EmbeddingResponse struct {
	Model       string
	Vectors     [][]float32
	InputTokens *int64
	Diagnostic  Diagnostic
}
type Embedder interface {
	Name() string
	Embed(context.Context, string, []string, int) (EmbeddingResponse, error)
}
type OpenAIEmbedder struct {
	name, endpoint, key string
	client              *http.Client
}

func NewOpenAIEmbedder(name, endpoint, key string, client *http.Client) (*OpenAIEmbedder, error) {
	u, err := url.Parse(endpoint)
	if name == "" || key == "" || client == nil || err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return nil, ErrInvalid
	}
	return &OpenAIEmbedder{name, endpoint, key, client}, nil
}
func (e *OpenAIEmbedder) Name() string { return e.name }
func (e *OpenAIEmbedder) Embed(ctx context.Context, model string, input []string, dimensions int) (EmbeddingResponse, error) {
	if model == "" || len(input) == 0 || len(input) > 100 || dimensions < 1 || dimensions > 8192 {
		return EmbeddingResponse{}, ErrInvalid
	}
	body, _ := json.Marshal(map[string]any{"model": model, "input": input, "dimensions": dimensions})
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint, bytes.NewReader(body))
	if err != nil {
		return EmbeddingResponse{}, ErrInvalid
	}
	request.Header.Set("Authorization", "Bearer "+e.key)
	request.Header.Set("Content-Type", "application/json")
	response, err := e.client.Do(request)
	if err != nil {
		return EmbeddingResponse{}, &CallError{Code: "provider_temporary", Temporary: true}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		failure := rejection(response, e.key)
		return EmbeddingResponse{InputTokens: failure.Usage.InputTokens, Diagnostic: failure.Diagnostic}, failure
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil {
		return EmbeddingResponse{}, &CallError{Code: "provider_temporary", Temporary: true}
	}
	var wire struct {
		Model string `json:"model"`
		Data  []struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Usage struct {
			PromptTokens *int64 `json:"prompt_tokens"`
		} `json:"usage"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decodeErr := decoder.Decode(&wire)
	result := EmbeddingResponse{InputTokens: nonnegative(wire.Usage.PromptTokens), Diagnostic: responseDiagnostic(response, e.key)}
	invalid := func() (EmbeddingResponse, error) {
		return result, &CallError{Code: "provider_response_invalid", StatusCode: response.StatusCode, Diagnostic: result.Diagnostic, Usage: Usage{InputTokens: result.InputTokens}}
	}
	if decodeErr != nil || wire.Model == "" || len(wire.Data) != len(input) {
		return invalid()
	}
	if err = decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return invalid()
	}
	vectors := make([][]float32, len(input))
	for _, item := range wire.Data {
		if item.Index < 0 || item.Index >= len(input) || len(item.Embedding) != dimensions || vectors[item.Index] != nil {
			return invalid()
		}
		var norm float64
		for _, v := range item.Embedding {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				return invalid()
			}
			norm += float64(v * v)
		}
		if norm == 0 {
			return invalid()
		}
		scale := float32(1 / math.Sqrt(norm))
		for i := range item.Embedding {
			item.Embedding[i] *= scale
		}
		vectors[item.Index] = item.Embedding
	}
	for _, v := range vectors {
		if v == nil {
			return EmbeddingResponse{}, errors.New("invalid embedding order")
		}
	}
	result.Model, result.Vectors = wire.Model, vectors
	return result, nil
}
