package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errControlArguments = errors.New("invalid control arguments")
var errControlRole = errors.New("controls require the admin role")

type controlCommand struct {
	Name, Source, Region, ISTAT, Actor string
	Revision                           int
	// Decision and Input carry a reviewed unavailable-target disposition; the
	// decision payload is read from Input because it holds free text and a
	// JSON evidence object.
	Decision string
	Input    io.Reader
}

// unavailableDecision is the stdin payload of source-unavailable-decide.
type unavailableDecision struct {
	URL                 string          `json:"url"`
	ObservedLastSeenAt  time.Time       `json:"observed_last_seen_at"`
	Reason              string          `json:"reason"`
	Evidence            json.RawMessage `json:"evidence"`
	AcknowledgeMeasures bool            `json:"acknowledge_measures"`
}

func parseControlCommand(args []string) (controlCommand, bool, error) {
	var c controlCommand
	if len(args) == 0 {
		return c, false, nil
	}
	c.Name = args[0]
	var count int
	switch c.Name {
	case "source-embedding-status":
		count = 2
	case "source-embedding-enable", "source-embedding-disable":
		count = 4
	case "embedding-status":
		count = 1
	case "embedding-enable", "embedding-disable":
		count = 3
	case "source-status", "region-status", "source-unavailable-targets", "source-tracked-targets":
		count = 2
	case "source-unavailable-decide":
		count = 5
	case "municipality-status", "municipality-development-publication-status":
		count = 3
	case "source-enable-collection", "source-suspend-collection", "region-enable", "region-disable":
		count = 4
	case "municipality-enable", "municipality-disable", "municipality-development-publication-enable", "municipality-development-publication-disable":
		count = 5
	default:
		return c, false, nil
	}
	if len(args) != count {
		return c, true, errControlArguments
	}
	for _, arg := range args[1:] {
		if strings.TrimSpace(arg) == "" {
			return c, true, errControlArguments
		}
	}
	switch {
	case strings.HasPrefix(c.Name, "embedding-"):
	case strings.HasPrefix(c.Name, "source-"):
		c.Source = args[1]
	case strings.HasPrefix(c.Name, "region-"):
		c.Region = args[1]
	default:
		c.Region, c.ISTAT = args[1], args[2]
	}
	if c.Region != "" && !controlDigits(c.Region, 2) {
		return c, true, errControlArguments
	}
	if c.ISTAT != "" && !controlDigits(c.ISTAT, 6) {
		return c, true, errControlArguments
	}
	if c.Name == "source-unavailable-decide" {
		configuration, err := strconv.Atoi(args[2])
		if err != nil || configuration < 1 || (args[3] != acquisition.DecisionExclude && args[3] != acquisition.DecisionResume) {
			return c, true, errControlArguments
		}
		c.Revision, c.Decision, c.Actor = configuration, args[3], args[4]
		return c, true, nil
	}
	if !strings.HasSuffix(c.Name, "-status") && c.Name != "source-unavailable-targets" && c.Name != "source-tracked-targets" {
		rev, err := strconv.Atoi(args[len(args)-2])
		if err != nil || rev < 0 || (c.Source != "" && rev == 0 && !strings.HasPrefix(c.Name, "source-embedding-")) {
			return c, true, errControlArguments
		}
		c.Revision, c.Actor = rev, args[len(args)-1]
	}
	return c, true, nil
}

func controlDigits(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// No worker, crawler, provider or container client is constructed by controls.
func executeControl(ctx context.Context, pool *pgxpool.Pool, role string, c controlCommand, out io.Writer) error {
	return executeControlEnvironment(ctx, pool, role, "production", c, out)
}
func executeControlEnvironment(ctx context.Context, pool *pgxpool.Pool, role, environment string, c controlCommand, out io.Writer) error {
	if role != "admin" {
		return errControlRole
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	reg, geo := registry.New(pool), domain.New(pool)
	var result any
	var err error
	switch c.Name {
	case "source-embedding-enable", "source-embedding-disable", "source-embedding-status":
		queue := jobs.New(pool)
		if c.Name != "source-embedding-status" {
			_, err = queue.SetSourceEmbeddingEnabled(ctx, c.Source, c.Revision, c.Name == "source-embedding-enable", c.Actor, time.Now())
		}
		if err == nil {
			result, err = queue.SourceEmbeddingState(ctx, c.Source)
		}
	case "embedding-enable", "embedding-disable", "embedding-status":
		queue := jobs.New(pool)
		if c.Name != "embedding-status" {
			_, err = queue.SetEmbeddingEnabled(ctx, c.Revision, c.Name == "embedding-enable", c.Actor, time.Now())
		}
		if err == nil {
			result, err = queue.EmbeddingState(ctx)
		}
	case "municipality-development-publication-enable", "municipality-development-publication-disable", "municipality-development-publication-status":
		publication := domain.NewDevelopmentPublication(pool, environment)
		if !strings.HasSuffix(c.Name, "-status") {
			_, err = publication.Set(ctx, c.Region, c.ISTAT, c.Revision, strings.HasSuffix(c.Name, "-enable"), c.Actor)
		}
		if err == nil {
			result, err = publication.State(ctx, c.Region, c.ISTAT)
		}
	case "source-enable-collection", "source-suspend-collection", "source-status":
		if c.Name == "source-enable-collection" {
			err = reg.EnableCollection(ctx, c.Source, c.Revision, c.Actor)
		}
		if c.Name == "source-suspend-collection" {
			err = reg.Disable(ctx, c.Source, c.Revision, c.Actor, false)
		}
		if err == nil {
			var state registry.State
			state, err = reg.State(ctx, c.Source)
			result = map[string]any{"source_id": state.Source.ID, "latest_revision": state.LatestRevision, "active_revision": state.ActiveRevision, "collection_enabled": state.CollectionEnabled, "public_enabled": state.PublicEnabled, "guidance": "Collection enablement is saved configuration. Automatic collection requires an active worker and enabled territories; publication is separate."}
		}
	case "source-unavailable-targets":
		var targets []acquisition.UnavailableTarget
		targets, err = acquisition.NewDispositionStore(pool).ListUnavailable(ctx, c.Source)
		if targets == nil {
			targets = []acquisition.UnavailableTarget{}
		}
		result = map[string]any{"source_id": c.Source, "unavailable_targets": targets, "guidance": "Withdrawn targets no longer fail checks. Blocking targets are still listed, referenced or support a measure; record a reviewed decision only after checking complete listings."}
	case "source-tracked-targets":
		var targets []acquisition.FollowedTarget
		targets, err = acquisition.NewTrackingStore(pool).FollowedTargets(ctx, c.Source, time.Now())
		result = map[string]any{"source_id": c.Source, "followed_targets": targets, "guidance": "States are derived at the start of each scheduled check from validated interpretation. Open-ended measures stop being rechecked after the source limit; listing only reads the last derived state."}
	case "source-unavailable-decide":
		var payload unavailableDecision
		if c.Input == nil {
			return errControlArguments
		}
		decoder := json.NewDecoder(io.LimitReader(c.Input, 64<<10))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&payload) != nil {
			return errControlArguments
		}
		var recorded acquisition.TargetDisposition
		recorded, err = acquisition.NewDispositionStore(pool).RecordDisposition(ctx, acquisition.DispositionRequest{SourceID: c.Source, Configuration: c.Revision, URL: payload.URL, Decision: c.Decision, ObservedLastSeenAt: payload.ObservedLastSeenAt, Actor: c.Actor, Reason: payload.Reason, Evidence: payload.Evidence, AcknowledgeMeasures: payload.AcknowledgeMeasures})
		result = map[string]any{"source_id": c.Source, "url": payload.URL, "disposition": recorded, "guidance": "The decision changes only the revision plan of this exact URL. It does not start a check, change publication or accept the source."}
	case "region-enable", "region-disable", "region-status":
		if c.Name != "region-status" {
			_, err = geo.SetRegionEnabled(ctx, c.Region, c.Revision, c.Name == "region-enable", c.Actor)
		}
		if err == nil {
			var state domain.Region
			state, err = geo.Region(ctx, c.Region)
			result = map[string]any{"region_code": state.Code, "name": state.Name, "revision": state.Revision, "enabled": state.Enabled, "guidance": "Region enablement preserves municipality and source choices and does not start a worker or grant publication."}
		}
	case "municipality-enable", "municipality-disable", "municipality-status":
		if c.Name != "municipality-status" {
			_, err = geo.SetMunicipalityEnabled(ctx, c.Region, c.ISTAT, c.Revision, c.Name == "municipality-enable", c.Actor)
		}
		if err == nil {
			result, err = geo.MunicipalityState(ctx, c.Region, c.ISTAT)
		}
	default:
		return errControlArguments
	}
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(result)
}

func controlErrorCode(err error) string {
	if code := acquisition.DispositionErrorCode(err); code != "disposition_failed" {
		return code
	}
	switch {
	case errors.Is(err, domain.ErrDevelopmentOnly):
		return "development_environment_required"
	case errors.Is(err, errControlArguments), errors.Is(err, registry.ErrInvalid), errors.Is(err, domain.ErrInvalid), errors.Is(err, jobs.ErrInvalid):
		return "invalid_arguments"
	case errors.Is(err, errControlRole):
		return "admin_role_required"
	case errors.Is(err, registry.ErrNotFound), errors.Is(err, domain.ErrTerritoryNotFound):
		return "not_found"
	case errors.Is(err, registry.ErrConflict), errors.Is(err, domain.ErrConflict), errors.Is(err, jobs.ErrConflict):
		return "revision_conflict"
	case errors.Is(err, registry.ErrPrerequisite):
		return "preview_or_policy_required"
	case errors.Is(err, domain.ErrRegionIncomplete):
		return "complete_municipality_register_required"
	case errors.Is(err, territory.ErrUnsupported), errors.Is(err, domain.ErrProfileUnsupported):
		return "unsupported_profile"
	case errors.Is(err, territory.ErrAssociation):
		return "unresolved_territory"
	case errors.Is(err, territory.ErrDisabled):
		return "territory_disabled"
	default:
		return "control_unavailable"
	}
}
