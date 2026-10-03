package processing

import (
	"context"
	"encoding/json"
	"errors"
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
	err := s.RegisterConfiguration(ctx, c)
	if !errors.Is(err, ErrConflict) {
		return c.ID, err
	}
	// Multiple fallback revisions can share a catalog name. The legacy local
	// revision then collides with UNIQUE(name,stage,revision), even though its
	// fallback and immutable ID differ. Keep existing rows and namespace the
	// new registration by the exact fallback rather than replacing history.
	c.ID = "local-processing-" + digest([]byte(fallback+":"+registry.LocalProcessingVersion+":fallback-scoped-name-v2"))
	c.Name = "local-processing-" + digest([]byte(fallback))[:32]
	return c.ID, s.RegisterConfiguration(ctx, c)
}
