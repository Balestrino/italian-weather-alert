package embedding

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/processing"
)

type Catalog struct{ ConfigurationVersionID string }
type catalogStore interface {
	RegisterModel(context.Context, processing.ModelVersion) error
	RegisterConfiguration(context.Context, processing.ConfigurationVersion) error
}

func RegisterCatalog(ctx context.Context, s catalogStore, adapter, model string, dimensions int, at time.Time) (Catalog, error) {
	if adapter == "" || model == "" || dimensions != 1024 || at.IsZero() {
		return Catalog{}, ErrInvalid
	}
	modelID := stable("embedding-model", adapter, model)
	configID := stable("embedding-config", modelID, strconv.Itoa(dimensions))
	if e := s.RegisterModel(ctx, processing.ModelVersion{ID: modelID, Provider: adapter, Model: model, Revision: "runtime-contract-v1", Capabilities: json.RawMessage(`{"family":"embedding","mrl_dimensions":1024}`), CreatedAt: at}); e != nil {
		return Catalog{}, e
	}
	if e := s.RegisterConfiguration(ctx, processing.ConfigurationVersion{ID: configID, Name: "measure-semantic-embedding", Stage: "embedding", Revision: "v1", ModelVersionID: &modelID, LogicVersion: "normalized-cosine-v1", Settings: json.RawMessage(`{"dimensions":1024,"input":"kind_subject_place"}`), CreatedAt: at}); e != nil {
		return Catalog{}, e
	}
	return Catalog{configID}, nil
}
func stable(prefix string, parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return prefix + "-" + hex.EncodeToString(h.Sum(nil))[:16]
}
