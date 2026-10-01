package domain

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"slices"
	"strings"
)

var ErrRegionIncomplete = errors.New("region requires a complete municipality register")
var ErrRegionDisabled = errors.New("region is disabled")
var ErrProfileUnsupported = errors.New("processing profile unsupported for this territory")

func (s *Store) ConfigureRegion(ctx context.Context, code string, expected int, cfg RegionConfiguration, actor string) (int, error) {
	return s.changeRegion(ctx, code, expected, &cfg, nil, actor)
}
func (s *Store) SetRegionEnabled(ctx context.Context, code string, expected int, enabled bool, actor string) (int, error) {
	return s.changeRegion(ctx, code, expected, nil, &enabled, actor)
}
func validateRegionConfiguration(ctx context.Context, tx pgx.Tx, code string, cfg RegionConfiguration, enabled bool) error {
	seen := map[string]bool{}
	for _, p := range cfg.Profiles {
		if seen[p] || !slices.Contains([]string{"municipal-html", "toscana-cfr", "dpc-comparison"}, p) || (p != "municipal-html" && code != "09") {
			return ErrProfileUnsupported
		}
		seen[p] = true
	}
	if enabled && cfg.MunicipalityDataset == "" {
		return ErrRegionIncomplete
	}
	for kind, id := range map[string]string{DatasetMunicipalities: cfg.MunicipalityDataset, DatasetZones: cfg.ZoneDataset, DatasetPostal: cfg.PostalDataset} {
		if id == "" {
			continue
		}
		var complete, usable bool
		var parent *string
		err := tx.QueryRow(ctx, `SELECT r.complete,d.usable,d.municipality_dataset_id FROM territorial_dataset_regions r JOIN geography_datasets d ON d.id=r.dataset_id WHERE r.region_code=$1 AND d.kind=$2 AND d.id=$3`, code, kind, id).Scan(&complete, &usable, &parent)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalid
		}
		if err != nil {
			return err
		}
		if !usable || (kind != DatasetMunicipalities && (parent == nil || *parent != cfg.MunicipalityDataset)) {
			return ErrInvalid
		}
		if enabled && kind == DatasetMunicipalities && !complete {
			return ErrRegionIncomplete
		}
	}
	return nil
}
func (s *Store) changeRegion(ctx context.Context, code string, expected int, configuration *RegionConfiguration, enabled *bool, actor string) (int, error) {
	if strings.TrimSpace(actor) == "" || expected < 0 {
		return 0, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var revision int
	var active bool
	err = tx.QueryRow(ctx, `SELECT revision,enabled FROM territorial_regions WHERE code=$1 FOR UPDATE`, code).Scan(&revision, &active)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrTerritoryNotFound
	}
	if err != nil {
		return 0, err
	}
	if revision != expected {
		return 0, ErrConflict
	}
	cfg := RegionConfiguration{Profiles: []string{}}
	if revision > 0 {
		if err = tx.QueryRow(ctx, `SELECT configuration FROM territorial_region_versions WHERE region_code=$1 AND revision=$2`, code, revision).Scan(&cfg); err != nil {
			return 0, err
		}
	}
	old := cfg
	kind := "configuration"
	if configuration != nil {
		cfg = *configuration
		if cfg.Profiles == nil {
			cfg.Profiles = []string{}
		}
	}
	if enabled != nil {
		active = *enabled
		kind = "disabled"
		if active {
			kind = "enabled"
		}
	}
	// Disablement must remain possible even when an old setup has become incomplete.
	if enabled == nil || *enabled {
		if err = validateRegionConfiguration(ctx, tx, code, cfg, active); err != nil {
			return 0, err
		}
	}
	for datasetKind, pair := range map[string][2]string{DatasetMunicipalities: {old.MunicipalityDataset, cfg.MunicipalityDataset}, DatasetZones: {old.ZoneDataset, cfg.ZoneDataset}, DatasetPostal: {old.PostalDataset, cfg.PostalDataset}} {
		if pair[1] != "" && pair[0] != pair[1] {
			if _, err = tx.Exec(ctx, `INSERT INTO territorial_dataset_selections(region_code,kind,dataset_id,actor) VALUES($1,$2,$3,$4)`, code, datasetKind, pair[1], actor); err != nil {
				return 0, err
			}
		}
	}
	rev, err := writeRegion(ctx, tx, code, expected, cfg, active, actor, kind)
	if err != nil {
		return 0, err
	}
	return rev, tx.Commit(ctx)
}
