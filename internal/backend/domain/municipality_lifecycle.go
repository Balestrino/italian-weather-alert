package domain

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

type MunicipalityState struct {
	RegionCode    string `json:"region_code"`
	ISTAT         string `json:"istat"`
	Revision      int    `json:"revision"`
	Enabled       bool   `json:"enabled"`
	RegionEnabled bool   `json:"region_enabled"`
	Historical    bool   `json:"historical"`
	Eligible      bool   `json:"territorially_eligible"`
	BlockedBy     string `json:"blocked_by,omitempty"`
}

func (s *Store) MunicipalityState(ctx context.Context, region, istat string) (MunicipalityState, error) {
	v := MunicipalityState{RegionCode: region, ISTAT: istat}
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(st.revision,0),COALESCE(st.enabled,false),r.enabled,
 NOT EXISTS(SELECT 1 FROM territorial_municipalities current WHERE current.region_code=r.code AND current.istat=$2 AND current.dataset_id=rv.configuration->>'municipality_dataset')
 FROM territorial_regions r LEFT JOIN territorial_region_versions rv ON rv.region_code=r.code AND rv.revision=r.revision
 LEFT JOIN territorial_municipality_state st ON st.region_code=r.code AND st.istat=$2
 WHERE r.code=$1 AND EXISTS(SELECT 1 FROM territorial_municipalities m WHERE m.region_code=r.code AND m.istat=$2)`, region, istat).
		Scan(&v.Revision, &v.Enabled, &v.RegionEnabled, &v.Historical)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrTerritoryNotFound
	}
	if err != nil {
		return v, err
	}
	v.ResolveEligibility()
	return v, nil
}

// ResolveEligibility describes territorial admission, independently of source flags and worker execution.
func (v *MunicipalityState) ResolveEligibility() {
	v.Eligible = v.Enabled && v.RegionEnabled && !v.Historical
	v.BlockedBy = ""
	switch {
	case !v.RegionEnabled:
		v.BlockedBy = "region_disabled"
	case !v.Enabled:
		v.BlockedBy = "municipality_disabled"
	case v.Historical:
		v.BlockedBy = "municipality_retired"
	}
}

func (s *Store) SetMunicipalityEnabled(ctx context.Context, region, istat string, expected int, enabled bool, actor string) (int, error) {
	if expected < 0 || strings.TrimSpace(actor) == "" {
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
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM territorial_municipalities WHERE region_code=$1 AND istat=$2 AND (NOT $3 OR dataset_id=$4))`, region, istat, enabled, dataset).Scan(&member)
	if err != nil {
		return 0, err
	}
	if !member {
		return 0, ErrTerritoryNotFound
	}
	if _, err = tx.Exec(ctx, `INSERT INTO territorial_municipality_state(region_code,istat) VALUES($1,$2) ON CONFLICT DO NOTHING`, region, istat); err != nil {
		return 0, err
	}
	var revision int
	if err = tx.QueryRow(ctx, `SELECT revision FROM territorial_municipality_state WHERE region_code=$1 AND istat=$2 FOR UPDATE`, region, istat).Scan(&revision); err != nil {
		return 0, err
	}
	if revision != expected {
		return 0, ErrConflict
	}
	revision++
	if _, err = tx.Exec(ctx, `UPDATE territorial_municipality_state SET revision=$3,enabled=$4 WHERE region_code=$1 AND istat=$2`, region, istat, revision, enabled); err != nil {
		return 0, err
	}
	kind := "disabled"
	if enabled {
		kind = "enabled"
	}
	if _, err = tx.Exec(ctx, `INSERT INTO territorial_municipality_events(region_code,istat,revision,enabled,kind,actor) VALUES($1,$2,$3,$4,$5,$6)`, region, istat, revision, enabled, kind, actor); err != nil {
		return 0, err
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return revision, nil
}
