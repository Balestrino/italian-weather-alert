package domain

import (
	"context"
	_ "embed"
	"encoding/csv"
	"github.com/jackc/pgx/v5"
	"strings"
)

// Official reference and checksum are attributed in territorial_catalog.json.
//
//go:embed territorial_municipalities.csv
var municipalityReference string

type TerritoryMigrationReport struct {
	UnassociatedSources  []string
	UnassociatedDatasets []string
}

func (s *Store) TerritoryMigrationReport(ctx context.Context) (TerritoryMigrationReport, error) {
	r := TerritoryMigrationReport{UnassociatedSources: []string{}, UnassociatedDatasets: []string{}}
	for _, item := range []struct {
		q   string
		out *[]string
	}{
		{`SELECT id FROM registry_sources WHERE NOT EXISTS(SELECT 1 FROM territorial_source_associations WHERE source_id=registry_sources.id) ORDER BY id`, &r.UnassociatedSources},
		{`SELECT id FROM geography_datasets WHERE NOT EXISTS(SELECT 1 FROM territorial_dataset_regions WHERE dataset_id=geography_datasets.id) ORDER BY id`, &r.UnassociatedDatasets},
	} {
		rows, err := s.pool.Query(ctx, item.q)
		if err != nil {
			return r, err
		}
		for rows.Next() {
			var id string
			if err = rows.Scan(&id); err != nil {
				rows.Close()
				return r, err
			}
			*item.out = append(*item.out, id)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

// BackfillToscana adds explicit associations without rewriting legacy source or geography state.
func (s *Store) BackfillToscana(ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(730021)`); err != nil {
		return err
	}
	var done bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM territorial_region_events WHERE kind='migration' AND region_code='09')`).Scan(&done); err != nil {
		return err
	}
	if done {
		return tx.Commit(ctx)
	}
	if _, err = tx.Exec(ctx, `CREATE TEMP TABLE territorial_reference(istat text PRIMARY KEY,region text) ON COMMIT DROP`); err != nil {
		return err
	}
	reference, err := csv.NewReader(strings.NewReader(municipalityReference)).ReadAll()
	if err != nil {
		return err
	}
	records := [][]any{}
	for _, r := range reference[1:] {
		records = append(records, []any{r[0], r[2]})
	}
	if _, err = tx.CopyFrom(ctx, pgx.Identifier{"territorial_reference"}, []string{"istat", "region"}, pgx.CopyFromRows(records)); err != nil {
		return err
	}
	for _, q := range []string{
		`INSERT INTO territorial_dataset_regions(dataset_id,region_code,complete,completeness_evidence)
 SELECT d.id,'09',(SELECT count(*) FROM geography_municipalities m WHERE m.dataset_id=d.id)=(SELECT count(*) FROM territorial_reference WHERE region='09'),'Backfill against retained ISTAT 2026-09-15 reference'
 FROM geography_datasets d WHERE d.kind='municipality_registry' AND d.usable AND EXISTS(SELECT 1 FROM geography_municipalities WHERE dataset_id=d.id)
 AND NOT EXISTS(SELECT 1 FROM geography_municipalities m LEFT JOIN territorial_reference r ON r.istat=m.istat WHERE m.dataset_id=d.id AND (r.region IS NULL OR r.region<>'09')) ON CONFLICT DO NOTHING`,
		`INSERT INTO territorial_municipalities SELECT d.region_code,m.dataset_id,m.istat FROM geography_municipalities m JOIN territorial_dataset_regions d ON d.dataset_id=m.dataset_id WHERE d.region_code='09' ON CONFLICT DO NOTHING`,
		`INSERT INTO territorial_dataset_regions(dataset_id,region_code,complete,completeness_evidence) SELECT d.id,'09',true,'Legacy attributed mapping' FROM geography_datasets d JOIN territorial_dataset_regions parent ON parent.dataset_id=d.municipality_dataset_id WHERE parent.region_code='09' ON CONFLICT DO NOTHING`,
		`INSERT INTO territorial_dataset_selections(region_code,kind,dataset_id,actor,selected_at) SELECT '09',s.kind,s.dataset_id,s.actor,s.selected_at FROM geography_dataset_selections s JOIN territorial_dataset_regions d ON d.dataset_id=s.dataset_id WHERE d.region_code='09' ORDER BY s.id`,
		`INSERT INTO territorial_source_associations(source_id,region_code,municipality_dataset_id,municipality_istat,profile,actor,recorded_at)
 SELECT s.id,'09',m.dataset_id,CASE WHEN s.product_id='municipal' THEN s.territory END,
 CASE WHEN s.product_id='municipal' THEN 'municipal-html' WHEN s.product_id='dpc_comparison' THEN 'dpc-comparison' ELSE 'toscana-cfr' END,'territorial-migration',COALESCE((SELECT min(created_at) FROM registry_configurations WHERE source_id=s.id),clock_timestamp())
 FROM registry_sources s LEFT JOIN LATERAL(SELECT tm.dataset_id FROM territorial_municipalities tm JOIN geography_datasets d ON d.id=tm.dataset_id WHERE tm.region_code='09' AND tm.istat=s.territory AND s.product_id='municipal' ORDER BY d.verified_at DESC,d.id LIMIT 1)m ON true
 WHERE ((s.territory='Toscana' AND s.product_id<>'municipal') OR (s.product_id='municipal' AND m.dataset_id IS NOT NULL)) AND NOT EXISTS(SELECT 1 FROM territorial_source_associations WHERE source_id=s.id)`,
	} {
		if _, err = tx.Exec(ctx, q); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(ctx, municipalityLifecycleBackfillSQL); err != nil {
		return err
	}
	var revision int
	var enabled bool
	if err = tx.QueryRow(ctx, `SELECT revision,enabled FROM territorial_regions WHERE code='09' FOR UPDATE`).Scan(&revision, &enabled); err != nil {
		return err
	}
	if revision == 0 {
		var cfg RegionConfiguration
		cfg.Profiles = []string{"toscana-cfr", "municipal-html", "dpc-comparison"}
		for kind, dest := range map[string]*string{DatasetMunicipalities: &cfg.MunicipalityDataset, DatasetZones: &cfg.ZoneDataset, DatasetPostal: &cfg.PostalDataset} {
			e := tx.QueryRow(ctx, `SELECT dataset_id FROM territorial_dataset_selections WHERE region_code='09' AND kind=$1 ORDER BY selected_at DESC,id DESC LIMIT 1`, kind).Scan(dest)
			if e != nil && e != pgx.ErrNoRows {
				return e
			}
		}
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM registry_sources s JOIN territorial_current_sources a ON a.source_id=s.id WHERE a.region_code='09' AND s.collection_enabled)`).Scan(&enabled); err != nil {
			return err
		}
		if _, err = writeRegion(ctx, tx, "09", 0, cfg, enabled, "territorial-migration", "migration"); err != nil {
			return err
		}
	}
	if revision != 0 {
		var cfg RegionConfiguration
		if err = tx.QueryRow(ctx, `SELECT configuration FROM territorial_region_versions WHERE region_code='09' AND revision=$1`, revision).Scan(&cfg); err != nil {
			return err
		}
		if _, err = writeRegion(ctx, tx, "09", revision, cfg, enabled, "territorial-migration", "migration"); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
