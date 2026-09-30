package domain

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

type SourceTerritory struct {
	SourceID, RegionCode, Profile          string
	MunicipalityDataset, MunicipalityISTAT *string
	RecordedAt                             time.Time
}

// TerritoryAt retains the association known at the requested observation boundary.
func (s *Store) TerritoryAt(ctx context.Context, source string, at time.Time) (SourceTerritory, error) {
	var v SourceTerritory
	err := s.pool.QueryRow(ctx, `SELECT source_id,region_code,municipality_dataset_id,municipality_istat,profile,recorded_at FROM territorial_source_associations WHERE source_id=$1 AND recorded_at<=$2 ORDER BY recorded_at DESC,id DESC LIMIT 1`, source, at).Scan(&v.SourceID, &v.RegionCode, &v.MunicipalityDataset, &v.MunicipalityISTAT, &v.Profile, &v.RecordedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrTerritoryNotFound
	}
	return v, err
}
func (s *Store) TerritorialDatasetAt(ctx context.Context, region, kind string, at time.Time) (string, error) {
	var cfg RegionConfiguration
	e := s.pool.QueryRow(ctx, `SELECT configuration FROM territorial_region_versions WHERE region_code=$1 AND recorded_at<=$2 ORDER BY revision DESC LIMIT 1`, region, at).Scan(&cfg)
	if e == nil {
		id := ""
		switch kind {
		case DatasetMunicipalities:
			id = cfg.MunicipalityDataset
		case DatasetZones:
			id = cfg.ZoneDataset
		case DatasetPostal:
			id = cfg.PostalDataset
		default:
			return "", ErrInvalid
		}
		if id == "" {
			return "", ErrTerritoryNotFound
		}
		return id, nil
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return "", e
	}
	var id string
	err := s.pool.QueryRow(ctx, `SELECT dataset_id FROM territorial_dataset_selections WHERE region_code=$1 AND kind=$2 AND selected_at<=$3 ORDER BY selected_at DESC,id DESC LIMIT 1`, region, kind, at).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrTerritoryNotFound
	}
	return id, err
}
