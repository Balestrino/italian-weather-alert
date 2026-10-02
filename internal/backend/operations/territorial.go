package operations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5"
	"strings"
	"time"
)

type RegionSummary struct {
	domain.Region
	Municipalities              *int64
	EnabledMunicipalities       *int64
	Sources, Collecting, Issues *int64
	LastResult                  *time.Time
	ObservedAt                  time.Time
}
type TerritorialOverview struct {
	Regions    []RegionSummary
	ObservedAt time.Time
}

func (s *Store) TerritorialOverview(ctx context.Context, at time.Time) (TerritorialOverview, error) {
	regions, err := domain.New(s.pool).Regions(ctx)
	if err != nil {
		return TerritorialOverview{}, err
	}
	result := TerritorialOverview{Regions: []RegionSummary{}, ObservedAt: at}
	byCode := map[string]int{}
	for _, r := range regions {
		byCode[r.Code] = len(result.Regions)
		result.Regions = append(result.Regions, RegionSummary{Region: r, ObservedAt: at})
	}
	// All-region aggregates are independent: a missing operational subsection must not hide geography.
	rows, err := s.pool.Query(ctx, `SELECT r.code,count(m.istat),count(m.istat) FILTER(WHERE st.enabled) FROM territorial_regions r JOIN territorial_region_versions v ON v.region_code=r.code AND v.revision=r.revision LEFT JOIN territorial_municipalities m ON m.region_code=r.code AND m.dataset_id=v.configuration->>'municipality_dataset' LEFT JOIN territorial_municipality_state st ON st.region_code=m.region_code AND st.istat=m.istat WHERE COALESCE(v.configuration->>'municipality_dataset','')<>'' GROUP BY r.code`)
	if err == nil {
		for rows.Next() {
			var code string
			var n, enabled int64
			if e := rows.Scan(&code, &n, &enabled); e != nil {
				rows.Close()
				return result, e
			}
			result.Regions[byCode[code]].Municipalities = &n
			result.Regions[byCode[code]].EnabledMunicipalities = &enabled
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			for i := range result.Regions {
				result.Regions[i].Municipalities = nil
				result.Regions[i].EnabledMunicipalities = nil
			}
		}
	}
	rows, err = s.pool.Query(ctx, `SELECT r.code,count(s.id),count(s.id) FILTER(WHERE s.collection_enabled),count(s.id) FILTER(WHERE st.last_error_code IS NOT NULL OR (st.last_complete_at IS NOT NULL AND st.last_complete_at+st.delay_seconds*interval '1 second'<$1)),max(d.last_result)
 FROM territorial_regions r LEFT JOIN territorial_current_sources a ON a.region_code=r.code LEFT JOIN registry_sources s ON s.id=a.source_id LEFT JOIN acquisition_source_status st ON st.source_id=s.id LEFT JOIN LATERAL(SELECT max(v.first_acquired_at) last_result FROM retained_documents d JOIN retained_versions v ON v.document_id=d.id WHERE d.source_id=s.id)d ON true GROUP BY r.code`, at)
	if err == nil {
		for rows.Next() {
			var code string
			var sources, collecting, issues int64
			var last *time.Time
			if e := rows.Scan(&code, &sources, &collecting, &issues, &last); e != nil {
				rows.Close()
				return result, e
			}
			r := &result.Regions[byCode[code]]
			r.Sources = &sources
			r.Collecting = &collecting
			r.Issues = &issues
			r.LastResult = last
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			for i := range result.Regions {
				r := &result.Regions[i]
				r.Sources = nil
				r.Collecting = nil
				r.Issues = nil
				r.LastResult = nil
			}
		}
	}
	return result, nil
}

type MunicipalityFilter struct {
	Search, Province, Coverage, After string
	Limit                             int
}
type TerritorialMunicipality struct {
	RegionCode, DatasetID, ISTAT, Name, Province string
	Activation                                   domain.MunicipalityState
	Historical                                   bool
	Sources, Collecting                          int64
}
type MunicipalityPage struct {
	Region     domain.Region
	Items      []TerritorialMunicipality
	Total      int64
	Next       string
	Filter     MunicipalityFilter
	ObservedAt time.Time
}

func municipalityCursor(name, istat string) string {
	b, _ := json.Marshal([]string{name, istat})
	return base64.RawURLEncoding.EncodeToString(b)
}
func parseMunicipalityCursor(raw string) (string, string, error) {
	if raw == "" {
		return "", "", nil
	}
	if len(raw) > 2048 {
		return "", "", ErrInvalid
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	var v []string
	if err != nil || json.Unmarshal(b, &v) != nil || len(v) != 2 {
		return "", "", ErrInvalid
	}
	return v[0], v[1], nil
}

const municipalityListSQL = `WITH source_counts AS(SELECT a.region_code,a.municipality_istat,count(*) sources,count(*) FILTER(WHERE s.collection_enabled) collecting FROM territorial_current_sources a JOIN registry_sources s ON s.id=a.source_id GROUP BY a.region_code,a.municipality_istat), matching AS(
 SELECT m.region_code,m.dataset_id,m.istat,g.name,g.province,COALESCE(c.sources,0) sources,COALESCE(c.collecting,0) collecting,COALESCE(st.revision,0) revision,COALESCE(st.enabled,false) enabled FROM territorial_municipalities m LEFT JOIN territorial_municipality_state st ON st.region_code=m.region_code AND st.istat=m.istat JOIN geography_municipalities g ON g.dataset_id=m.dataset_id AND g.istat=m.istat LEFT JOIN source_counts c ON c.region_code=m.region_code AND c.municipality_istat=m.istat
 WHERE m.region_code=$1 AND m.dataset_id=$2 AND ($3='' OR strpos(lower(g.name),lower($3))>0 OR strpos(g.istat,$3)>0) AND ($4='' OR g.province=$4)
 AND ($5='' OR ($5='none' AND COALESCE(c.sources,0)=0) OR ($5='configured' AND c.sources>0) OR ($5='collecting' AND c.collecting>0))) `

func (s *Store) Municipalities(ctx context.Context, region string, f MunicipalityFilter, at time.Time) (MunicipalityPage, error) {
	result := MunicipalityPage{Items: []TerritorialMunicipality{}, Filter: f, ObservedAt: at}
	if f.Limit == 0 {
		f.Limit = 50
	}
	if f.Limit < 1 || f.Limit > 100 || len(f.Search) > 200 || len(f.Province) > 100 || (f.Coverage != "" && f.Coverage != "none" && f.Coverage != "configured" && f.Coverage != "collecting") {
		return result, ErrInvalid
	}
	name, istat, err := parseMunicipalityCursor(f.After)
	if err != nil {
		return result, err
	}
	result.Region, err = domain.New(s.pool).Region(ctx, region)
	if err != nil {
		return result, err
	}
	result.Filter = f
	if result.Region.Configuration.MunicipalityDataset == "" {
		return result, nil
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	args := []any{region, result.Region.Configuration.MunicipalityDataset, strings.TrimSpace(f.Search), f.Province, f.Coverage}
	if err = tx.QueryRow(ctx, municipalityListSQL+`SELECT count(*) FROM matching`, args...).Scan(&result.Total); err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, municipalityListSQL+`SELECT region_code,dataset_id,istat,name,province,sources,collecting,revision,enabled FROM matching WHERE ($6='' OR (name,istat)>($6,$7)) ORDER BY name,istat LIMIT $8`, append(args, name, istat, f.Limit+1)...)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var m TerritorialMunicipality
		if err = rows.Scan(&m.RegionCode, &m.DatasetID, &m.ISTAT, &m.Name, &m.Province, &m.Sources, &m.Collecting, &m.Activation.Revision, &m.Activation.Enabled); err != nil {
			rows.Close()
			return result, err
		}
		m.Activation.RegionCode, m.Activation.ISTAT, m.Activation.RegionEnabled = m.RegionCode, m.ISTAT, result.Region.Enabled
		m.Activation.ResolveEligibility()
		result.Items = append(result.Items, m)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	if len(result.Items) > f.Limit {
		result.Items = result.Items[:f.Limit]
		last := result.Items[len(result.Items)-1]
		result.Next = municipalityCursor(last.Name, last.ISTAT)
	}
	return result, tx.Commit(ctx)
}
func (s *Store) Municipality(ctx context.Context, region, istat string) (TerritorialMunicipality, error) {
	var m TerritorialMunicipality
	err := s.pool.QueryRow(ctx, `SELECT m.region_code,m.dataset_id,m.istat,g.name,g.province,(m.dataset_id IS DISTINCT FROM v.configuration->>'municipality_dataset'),(SELECT count(*) FROM territorial_current_sources a WHERE a.region_code=m.region_code AND a.municipality_istat=m.istat),(SELECT count(*) FROM territorial_current_sources a JOIN registry_sources s ON s.id=a.source_id WHERE a.region_code=m.region_code AND a.municipality_istat=m.istat AND s.collection_enabled),COALESCE(st.revision,0),COALESCE(st.enabled,false),r.enabled
 FROM territorial_municipalities m JOIN geography_municipalities g ON g.dataset_id=m.dataset_id AND g.istat=m.istat JOIN geography_datasets d ON d.id=m.dataset_id LEFT JOIN territorial_municipality_state st ON st.region_code=m.region_code AND st.istat=m.istat JOIN territorial_regions r ON r.code=m.region_code LEFT JOIN territorial_region_versions v ON v.region_code=r.code AND v.revision=r.revision
 WHERE m.region_code=$1 AND m.istat=$2 ORDER BY (m.dataset_id=v.configuration->>'municipality_dataset') DESC NULLS LAST,d.verified_at DESC,d.id DESC LIMIT 1`, region, istat).Scan(&m.RegionCode, &m.DatasetID, &m.ISTAT, &m.Name, &m.Province, &m.Historical, &m.Sources, &m.Collecting, &m.Activation.Revision, &m.Activation.Enabled, &m.Activation.RegionEnabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, domain.ErrTerritoryNotFound
	}
	m.Activation.RegionCode, m.Activation.ISTAT, m.Activation.Historical = m.RegionCode, m.ISTAT, m.Historical
	m.Activation.ResolveEligibility()
	return m, err
}

type TerritorialSource struct {
	ID, Product, RegionCode, Profile             string
	MunicipalityISTAT                            *string
	Revision                                     int
	ActiveRevision                               *int
	Collecting, Public                           bool
	Configuration                                registry.Configuration
	EffectiveCheckSeconds, EffectiveDelaySeconds int
	LastChecked                                  *time.Time
	ErrorCode                                    *string
}

func (s *Store) TerritorialSources(ctx context.Context, region, istat string) ([]TerritorialSource, error) {
	if _, err := domain.New(s.pool).Region(ctx, region); err != nil {
		return nil, err
	}
	if istat != "" {
		if _, err := s.Municipality(ctx, region, istat); err != nil {
			return nil, err
		}
	}
	rows, err := s.pool.Query(ctx, `SELECT s.id,s.product_id,a.region_code,a.profile,a.municipality_istat,s.latest_revision,s.active_revision,s.collection_enabled,s.public_enabled,c.body,COALESCE(s.check_seconds_override,(c.body->>'check_seconds')::int,600),COALESCE(s.delay_seconds_override,(c.body->>'delay_seconds')::int,1800),st.last_complete_at,st.last_error_code FROM territorial_current_sources a JOIN registry_sources s ON s.id=a.source_id JOIN registry_configurations c ON c.source_id=s.id AND c.revision=COALESCE(s.active_revision,s.latest_revision) LEFT JOIN acquisition_source_status st ON st.source_id=s.id WHERE a.region_code=$1 AND ($2='' OR a.municipality_istat=$2) ORDER BY s.id`, region, istat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []TerritorialSource{}
	for rows.Next() {
		var v TerritorialSource
		if err = rows.Scan(&v.ID, &v.Product, &v.RegionCode, &v.Profile, &v.MunicipalityISTAT, &v.Revision, &v.ActiveRevision, &v.Collecting, &v.Public, &v.Configuration, &v.EffectiveCheckSeconds, &v.EffectiveDelaySeconds, &v.LastChecked, &v.ErrorCode); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
