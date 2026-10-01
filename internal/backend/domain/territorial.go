package domain

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

//go:embed territorial_catalog.json
var territorialCatalog []byte

var ErrTerritoryNotFound = errors.New("territory not found")

type RegionConfiguration struct {
	MunicipalityDataset string   `json:"municipality_dataset"`
	ZoneDataset         string   `json:"zone_dataset"`
	PostalDataset       string   `json:"postal_dataset"`
	Profiles            []string `json:"profiles"`
}
type Region struct {
	Code          string              `json:"code"`
	Name          string              `json:"name"`
	Revision      int                 `json:"revision"`
	Enabled       bool                `json:"enabled"`
	Configuration RegionConfiguration `json:"configuration"`
}
type RegionEvent struct {
	ID            int64
	RegionCode    string
	Revision      int
	Kind, Actor   string
	RecordedAt    time.Time
	Configuration RegionConfiguration
	Enabled       bool
}

func seedRegions(ctx context.Context, tx pgx.Tx) error {
	var ref struct {
		Regions []struct{ Code, Name string } `json:"regions"`
	}
	if err := json.Unmarshal(territorialCatalog, &ref); err != nil {
		return err
	}
	for _, r := range ref.Regions {
		if _, err := tx.Exec(ctx, `INSERT INTO territorial_regions(code,name,reference) VALUES($1,$2,$3)`, r.Code, r.Name, territorialCatalog); err != nil {
			return err
		}
	}
	return nil
}
func (s *Store) Regions(ctx context.Context) ([]Region, error) {
	rows, err := s.pool.Query(ctx, `SELECT r.code,r.name,r.revision,r.enabled,COALESCE(v.configuration,'{}') FROM territorial_regions r LEFT JOIN territorial_region_versions v ON v.region_code=r.code AND v.revision=r.revision ORDER BY r.code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Region{}
	for rows.Next() {
		var r Region
		if err = rows.Scan(&r.Code, &r.Name, &r.Revision, &r.Enabled, &r.Configuration); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
func (s *Store) Region(ctx context.Context, code string) (Region, error) {
	var r Region
	err := s.pool.QueryRow(ctx, `SELECT r.code,r.name,r.revision,r.enabled,COALESCE(v.configuration,'{}') FROM territorial_regions r LEFT JOIN territorial_region_versions v ON v.region_code=r.code AND v.revision=r.revision WHERE r.code=$1`, code).Scan(&r.Code, &r.Name, &r.Revision, &r.Enabled, &r.Configuration)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, ErrTerritoryNotFound
	}
	return r, err
}

// writeRegion must run under the region row lock. All state and evidence commit together.
func writeRegion(ctx context.Context, tx pgx.Tx, code string, expected int, cfg RegionConfiguration, enabled bool, actor, kind string) (int, error) {
	if strings.TrimSpace(actor) == "" || expected < 0 {
		return 0, ErrInvalid
	}
	tag, err := tx.Exec(ctx, `UPDATE territorial_regions SET revision=revision+1,enabled=$3 WHERE code=$1 AND revision=$2`, code, expected, enabled)
	if err != nil {
		return 0, err
	}
	if tag.RowsAffected() != 1 {
		return 0, ErrConflict
	}
	revision := expected + 1
	if _, err = tx.Exec(ctx, `INSERT INTO territorial_region_versions(region_code,revision,configuration,enabled,actor) VALUES($1,$2,$3,$4,$5)`, code, revision, cfg, enabled, actor); err != nil {
		return 0, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO territorial_region_events(region_code,revision,kind,actor) VALUES($1,$2,$3,$4)`, code, revision, kind, actor)
	return revision, err
}
func (s *Store) RegionEvents(ctx context.Context, code string) ([]RegionEvent, error) {
	rows, err := s.pool.Query(ctx, `SELECT e.id,e.region_code,e.revision,e.kind,e.actor,e.recorded_at,v.configuration,v.enabled FROM territorial_region_events e JOIN territorial_region_versions v ON v.region_code=e.region_code AND v.revision=e.revision WHERE e.region_code=$1 ORDER BY e.recorded_at DESC,e.id DESC LIMIT 100`, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []RegionEvent{}
	for rows.Next() {
		var e RegionEvent
		if err = rows.Scan(&e.ID, &e.RegionCode, &e.Revision, &e.Kind, &e.Actor, &e.RecordedAt, &e.Configuration, &e.Enabled); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
