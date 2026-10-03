package domain

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

var ErrDevelopmentOnly = errors.New("manual publication requires the development environment")

type DevelopmentPublicationState struct {
	RegionCode string `json:"region_code"`
	ISTAT      string `json:"istat"`
	Revision   int    `json:"revision"`
	Enabled    bool   `json:"enabled"`
}
type DevelopmentPublication struct {
	*Store
	development bool
}

func NewDevelopmentPublication(pool *pgxpool.Pool, environment string) *DevelopmentPublication {
	return &DevelopmentPublication{Store: New(pool), development: environment == "development"}
}
func (s *DevelopmentPublication) State(ctx context.Context, region, istat string) (DevelopmentPublicationState, error) {
	v := DevelopmentPublicationState{RegionCode: region, ISTAT: istat}
	if !s.development {
		return v, ErrDevelopmentOnly
	}
	if _, err := s.MunicipalityState(ctx, region, istat); err != nil {
		return v, err
	}
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(st.revision,0),COALESCE(st.enabled,false) FROM territorial_regions r LEFT JOIN development_publication_municipalities st ON st.region_code=r.code AND st.istat=$2 WHERE r.code=$1`, region, istat).Scan(&v.Revision, &v.Enabled)
	return v, err
}
func (s *DevelopmentPublication) Set(ctx context.Context, region, istat string, expected int, enabled bool, actor string) (int, error) {
	if !s.development {
		return 0, ErrDevelopmentOnly
	}
	if enabled && region != "09" {
		return 0, ErrProfileUnsupported
	}
	if expected < 0 || len(actor) > 200 || strings.TrimSpace(actor) == "" {
		return 0, ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var dataset string
	err = tx.QueryRow(ctx, `SELECT COALESCE(v.configuration->>'municipality_dataset','') FROM territorial_regions r
 LEFT JOIN territorial_region_versions v ON v.region_code=r.code AND v.revision=r.revision WHERE r.code=$1 FOR SHARE OF r`, region).Scan(&dataset)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrTerritoryNotFound
	}
	if err != nil {
		return 0, err
	}
	var member bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM territorial_municipalities m JOIN geography_municipalities g ON g.dataset_id=m.dataset_id AND g.istat=m.istat WHERE m.region_code=$1 AND m.istat=$2 AND (NOT $3 OR (m.dataset_id=$4 AND g.supported)))`, region, istat, enabled, dataset).Scan(&member)
	if err != nil {
		return 0, err
	}
	if !member {
		return 0, ErrTerritoryNotFound
	}
	if _, err = tx.Exec(ctx, `INSERT INTO development_publication_municipalities(region_code,istat) VALUES($1,$2) ON CONFLICT DO NOTHING`, region, istat); err != nil {
		return 0, err
	}
	var revision int
	if err = tx.QueryRow(ctx, `SELECT revision FROM development_publication_municipalities WHERE region_code=$1 AND istat=$2 FOR UPDATE`, region, istat).Scan(&revision); err != nil {
		return 0, err
	}
	if revision != expected {
		return 0, ErrConflict
	}
	revision++
	if _, err = tx.Exec(ctx, `UPDATE development_publication_municipalities SET revision=$3,enabled=$4 WHERE region_code=$1 AND istat=$2`, region, istat, revision, enabled); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO development_publication_events(region_code,istat,revision,enabled,actor) VALUES($1,$2,$3,$4,$5)`, region, istat, revision, enabled, actor); err != nil {
		return 0, err
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return revision, nil
}
