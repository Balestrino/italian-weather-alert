package notifications

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/domain"
	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
	"github.com/Balestrino/italian-weather-alert/internal/territory"
	"sort"
	"time"
)

type ReportSource struct {
	ID, Product, Profile              string
	Municipality                      *string
	Public, Supported                 bool
	LastChecked                       *time.Time
	DelaySeconds                      int
	Error                             *string
	Documents, Pending, Uninterpreted int
	Bulletins                         []ReportBulletin
}
type ReportBulletin struct {
	URL         string
	Version     int64
	AcquiredAt  time.Time
	Observation acquisition.RegionalObservation
}
type ReportMunicipality struct {
	ISTAT, Name    string
	Zones          []string
	PartialMapping bool
}
type ReportRegion struct {
	Code, Name     string
	Sources        []ReportSource
	Municipalities []ReportMunicipality
	Warnings       []publicquery.RegionalWarning
	Measures       []publicquery.Measure
	Phases         []publicquery.OperationalPhase
}
type ReportSnapshot struct {
	ObservedAt time.Time
	Regions    []ReportRegion
}
type factReader func(context.Context, pgx.Tx, string, map[string]string, time.Time) ([]publicquery.RegionalWarning, []publicquery.Measure, []publicquery.OperationalPhase, error)
type ReportReader struct {
	Pool  *pgxpool.Pool
	facts factReader
}

func (r ReportReader) Read(ctx context.Context, at time.Time) (ReportSnapshot, error) {
	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return ReportSnapshot{}, err
	}
	defer tx.Rollback(ctx)
	out := ReportSnapshot{ObservedAt: at, Regions: []ReportRegion{}}
	rows, err := tx.Query(ctx, `SELECT r.code,r.name,COALESCE(v.configuration,'{}') FROM territorial_regions r LEFT JOIN territorial_region_versions v ON v.region_code=r.code AND v.revision=r.revision WHERE r.enabled ORDER BY r.code`)
	if err != nil {
		return out, err
	}
	configs := map[string]domain.RegionConfiguration{}
	for rows.Next() {
		var region ReportRegion
		var cfg domain.RegionConfiguration
		if err = rows.Scan(&region.Code, &region.Name, &cfg); err != nil {
			rows.Close()
			return out, err
		}
		out.Regions = append(out.Regions, region)
		configs[region.Code] = cfg
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	for i := range out.Regions {
		region := &out.Regions[i]
		cfg := configs[region.Code]
		rows, err = tx.Query(ctx, `SELECT s.id,s.product_id,a.profile,a.municipality_istat,s.public_enabled,st.last_complete_at,COALESCE(s.delay_seconds_override,(c.body->>'delay_seconds')::int,1800),st.last_error_code FROM territorial_current_sources a JOIN registry_sources s ON s.id=a.source_id JOIN registry_configurations c ON c.source_id=s.id AND c.revision=COALESCE(s.active_revision,s.latest_revision) LEFT JOIN acquisition_source_status st ON st.source_id=s.id WHERE a.region_code=$1 AND s.collection_enabled ORDER BY s.id`, region.Code)
		if err != nil {
			return out, err
		}
		for rows.Next() {
			var s ReportSource
			if err = rows.Scan(&s.ID, &s.Product, &s.Profile, &s.Municipality, &s.Public, &s.LastChecked, &s.DelaySeconds, &s.Error); err != nil {
				rows.Close()
				return out, err
			}
			s.Supported = territory.Compatible(region.Code, s.Product, s.Profile)
			listed := false
			for _, p := range cfg.Profiles {
				if p == s.Profile {
					listed = true
				}
			}
			s.Supported = s.Supported && listed
			region.Sources = append(region.Sources, s)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return out, err
		}
		municipalities := map[string]bool{}
		for j := range region.Sources {
			source := &region.Sources[j]
			// Counts cover all latest document versions, not just a preview page.
			err = tx.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE c.status IS NULL),count(*) FILTER(WHERE c.status IS NOT NULL AND (e.status IS NULL OR e.status='uninterpreted')) FROM retained_documents d JOIN LATERAL(SELECT id FROM retained_versions WHERE document_id=d.id AND first_acquired_at<=$2 ORDER BY first_acquired_at DESC,id DESC LIMIT 1)v ON true LEFT JOIN LATERAL(SELECT cr.status FROM classification_results cr JOIN processing_runs p ON p.id=cr.run_id WHERE p.document_version_id=v.id AND p.created_at<=$2 ORDER BY p.created_at DESC,p.id DESC LIMIT 1)c ON true LEFT JOIN LATERAL(SELECT er.status FROM extraction_results er JOIN processing_runs p ON p.id=er.run_id WHERE p.document_version_id=v.id AND p.created_at<=$2 ORDER BY p.created_at DESC,p.id DESC LIMIT 1)e ON true WHERE d.source_id=$1`, source.ID, at).Scan(&source.Documents, &source.Pending, &source.Uninterpreted)
			if err != nil {
				return out, err
			}
			if source.Municipality != nil {
				municipalities[*source.Municipality] = true
			}
			if source.Product == "municipal" {
				continue
			}
			rows, err = tx.Query(ctx, `SELECT d.official_url,v.id,v.first_acquired_at,v.metadata FROM retained_documents d JOIN LATERAL(SELECT id,first_acquired_at,metadata FROM retained_versions WHERE document_id=d.id AND first_acquired_at<=$2 ORDER BY first_acquired_at DESC,id DESC LIMIT 1)v ON true WHERE d.source_id=$1 ORDER BY v.first_acquired_at DESC,v.id DESC LIMIT 10`, source.ID, at)
			if err != nil {
				return out, err
			}
			for rows.Next() {
				var b ReportBulletin
				var raw []byte
				if err = rows.Scan(&b.URL, &b.Version, &b.AcquiredAt, &raw); err != nil {
					rows.Close()
					return out, err
				}
				if err = json.Unmarshal(raw, &b.Observation); err != nil {
					rows.Close()
					return out, err
				}
				source.Bulletins = append(source.Bulletins, b)
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				return out, err
			}
		}
		ids := []string{}
		for id := range municipalities {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			m := ReportMunicipality{ISTAT: id, Name: id}
			err = tx.QueryRow(ctx, `SELECT name FROM geography_municipalities WHERE dataset_id=$1 AND istat=$2`, cfg.MunicipalityDataset, id).Scan(&m.Name)
			if err != nil && err != pgx.ErrNoRows {
				return out, err
			}
			rows, err = tx.Query(ctx, `SELECT zone,bool_or(territorial_scope='partial_municipality') FROM geography_zone_mappings WHERE dataset_id=$1 AND municipality_dataset_id=$2 AND municipality_istat=$3 GROUP BY zone ORDER BY zone`, cfg.ZoneDataset, cfg.MunicipalityDataset, id)
			if err != nil {
				return out, err
			}
			for rows.Next() {
				var zone string
				var partial bool
				if err = rows.Scan(&zone, &partial); err != nil {
					rows.Close()
					return out, err
				}
				m.Zones = append(m.Zones, zone)
				m.PartialMapping = m.PartialMapping || partial
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				return out, err
			}
			region.Municipalities = append(region.Municipalities, m)
		}
		if len(region.Sources) > 0 {
			read := r.facts
			if read == nil {
				read = publicquery.ReadReportFacts
			}
			region.Warnings, region.Measures, region.Phases, err = read(ctx, tx, region.Code, map[string]string{domain.DatasetMunicipalities: cfg.MunicipalityDataset, domain.DatasetZones: cfg.ZoneDataset}, at)
			if err != nil {
				return out, err
			}
		}
	}
	return out, tx.Commit(ctx)
}
