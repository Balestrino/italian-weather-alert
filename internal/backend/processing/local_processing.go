package processing

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

// RegisterLocalProcessing derives an immutable local-first configuration from
// an existing fallback, leaving old catalog identities and results untouched.
func (s *Store) RegisterLocalProcessing(ctx context.Context, fallback string, at time.Time) (string, error) {
	var c ConfigurationVersion
	if err := s.pool.QueryRow(ctx, `SELECT name,stage,revision,model_version_id,prompt_version_id,logic_version,settings FROM processing_configuration_versions WHERE id=$1`, fallback).Scan(&c.Name, &c.Stage, &c.Revision, &c.ModelVersionID, &c.PromptVersionID, &c.LogicVersion, &c.Settings); err != nil {
		return "", err
	}
	if c.Stage != "ocr" && c.Stage != "classification" && c.Stage != "extraction" {
		return "", ErrInvalid
	}
	c.ID = "local-processing-" + digest([]byte(fallback+":"+registry.LocalProcessingVersion))
	c.Revision = registry.LocalProcessingVersion
	c.LogicVersion = registry.LocalProcessingVersion
	c.CreatedAt = at
	c.Settings, _ = json.Marshal(map[string]any{"local_processing": registry.LocalProcessingVersion, "fallback_configuration": fallback, "fallback_settings": c.Settings})
	return c.ID, s.RegisterConfiguration(ctx, c)
}
