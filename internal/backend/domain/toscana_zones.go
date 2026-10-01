package domain

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

// AdoptToscanaZones adopts the hash-bound reviewed reconciliation, never a
// guessed mapping. Historical applicability remains unresolved.
func (s *Store) AdoptToscanaZones(ctx context.Context, body []byte, actor string) (string, error) {
	hash := sha256.Sum256(body)
	if hex.EncodeToString(hash[:]) != "152a3537e61d1609d4f12c83fb2ffe3da812a141d1d063cc12a69dafa70a6388" || strings.TrimSpace(actor) == "" {
		return "", ErrInvalid
	}
	rows, err := csv.NewReader(strings.NewReader(string(body))).ReadAll()
	if err != nil {
		return "", err
	}
	zones := map[string]map[string]bool{}
	pairs := map[string]int{}
	for _, r := range rows[1:] {
		if zones[r[3]] == nil {
			zones[r[3]] = map[string]bool{}
		}
		zones[r[3]][r[0]] = true
		pairs[r[3]+r[0]]++
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	var revision int
	var enabled bool
	var cfg RegionConfiguration
	if err = tx.QueryRow(ctx, "SELECT revision,enabled FROM territorial_regions WHERE code='09' FOR UPDATE").Scan(&revision, &enabled); err != nil {
		return "", err
	}
	if err = tx.QueryRow(ctx, "SELECT configuration FROM territorial_region_versions WHERE region_code='09' AND revision=$1", revision).Scan(&cfg); err != nil {
		return "", err
	}
	if cfg.MunicipalityDataset == "" {
		return "", ErrRegionIncomplete
	}
	id := "dgr1526-2025-a6-" + cfg.MunicipalityDataset
	var exists bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM geography_datasets WHERE id=$1)", id).Scan(&exists); err != nil {
		return "", err
	}
	if !exists {
		limitations, _ := json.Marshal([]string{"DGR 1526/2025 Annex 6 current-reference mapping; exact historical effective interval is unresolved. Zone membership alone does not establish a risk level.", "Reviewed reconciliation CSV SHA256: 152a3537e61d1609d4f12c83fb2ffe3da812a141d1d063cc12a69dafa70a6388; underlying applicability review is retained separately."})
		_, err = tx.Exec(ctx, `INSERT INTO geography_datasets(id,kind,municipality_dataset_id,authority,publisher,official_url,version_label,source_sha256,verified_at,applicability,limitations,usable) VALUES($1,'zone_mapping',$2,'Regione Toscana','Regione Toscana','https://www.regione.toscana.it/documents/d/guest/delibera-n-1526-del-20-10-2025-allegato-6-pdf','DGR 1526/2025 Annex 6','28bdd78f8b4c4e5a0eed5bf1f3d86bbf13f7d4212659eefb33a8202ed39d2f91','2026-09-16T00:00:00Z','unresolved',$3,true)`, id, cfg.MunicipalityDataset, limitations)
		if err != nil {
			return "", err
		}
		for n, r := range rows[1:] {
			scope := "whole_municipality"
			if len(zones[r[3]]) > 1 || pairs[r[3]+r[0]] > 1 || r[1] != r[4] {
				scope = "partial_municipality"
			}
			_, err = tx.Exec(ctx, `INSERT INTO geography_zone_mappings(dataset_id,municipality_dataset_id,municipality_istat,ordinal,zone,source_name,territorial_scope,evidence_locator) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, id, cfg.MunicipalityDataset, r[3], n+1, r[0], r[1], scope, "DGR 1526/2025 Annex 6 p."+r[6])
			if err != nil {
				return "", err
			}
		}
		var missing int
		if err = tx.QueryRow(ctx, `SELECT count(*) FROM geography_municipalities m WHERE m.dataset_id=$1 AND NOT EXISTS(SELECT 1 FROM geography_zone_mappings z WHERE z.dataset_id=$2 AND z.municipality_istat=m.istat)`, cfg.MunicipalityDataset, id).Scan(&missing); err != nil {
			return "", err
		}
		if missing != 0 {
			return "", ErrRegionIncomplete
		}
		_, err = tx.Exec(ctx, `INSERT INTO territorial_dataset_regions VALUES($1,'09',true,'282 reviewed source rows; 273 reconciled municipalities; underlying research evidence is retained separately')`, id)
		if err != nil {
			return "", err
		}
	}
	if cfg.ZoneDataset != id {
		cfg.ZoneDataset = id
		if err = validateRegionConfiguration(ctx, tx, "09", cfg, enabled); err != nil {
			return "", err
		}
		_, err = tx.Exec(ctx, `INSERT INTO territorial_dataset_selections(region_code,kind,dataset_id,actor) VALUES('09','zone_mapping',$1,$2)`, id, actor)
		if err != nil {
			return "", err
		}
		if _, err = writeRegion(ctx, tx, "09", revision, cfg, enabled, actor, "configuration"); err != nil {
			return "", err
		}
	}
	var selected string
	if err = tx.QueryRow(ctx, `SELECT COALESCE((SELECT dataset_id FROM geography_dataset_selections WHERE kind='zone_mapping' ORDER BY id DESC LIMIT 1),'')`).Scan(&selected); err != nil {
		return "", err
	}
	if selected != id {
		_, err = tx.Exec(ctx, `INSERT INTO geography_dataset_selections(kind,dataset_id,actor,selected_at) VALUES('zone_mapping',$1,$2,$3)`, id, actor, time.Now().UTC())
		if err != nil {
			return "", err
		}
	}
	return id, tx.Commit(ctx)
}
