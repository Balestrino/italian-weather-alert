package domain

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Store) RegisterMunicipalities(ctx context.Context, dataset Dataset, values []Municipality) error {
	if !validDataset(dataset, DatasetMunicipalities) || len(values) == 0 {
		return ErrInvalid
	}
	return s.registerDataset(ctx, dataset, func(tx pgx.Tx) error {
		for _, value := range values {
			if len(value.ISTAT) != 6 || strings.TrimSpace(value.Name) == "" || strings.TrimSpace(value.Province) == "" {
				return ErrInvalid
			}
			if _, err := tx.Exec(ctx, `INSERT INTO geography_municipalities(dataset_id,istat,name,province,supported) VALUES($1,$2,$3,$4,$5)`, dataset.ID, value.ISTAT, value.Name, value.Province, value.Supported); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) RegisterZones(ctx context.Context, dataset Dataset, values []ZoneMapping) error {
	if !validDataset(dataset, DatasetZones) || len(values) == 0 {
		return ErrInvalid
	}
	return s.registerDataset(ctx, dataset, func(tx pgx.Tx) error {
		for index, value := range values {
			if len(value.MunicipalityISTAT) != 6 || strings.TrimSpace(value.Zone) == "" || strings.TrimSpace(value.SourceName) == "" || !slicesContains([]string{"whole_municipality", "partial_municipality"}, value.TerritorialScope) || strings.TrimSpace(value.EvidenceLocator) == "" {
				return ErrInvalid
			}
			if _, err := tx.Exec(ctx, `INSERT INTO geography_zone_mappings(dataset_id,municipality_dataset_id,municipality_istat,ordinal,zone,source_name,territorial_scope,evidence_locator) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, dataset.ID, dataset.MunicipalityDatasetID, value.MunicipalityISTAT, index+1, value.Zone, value.SourceName, value.TerritorialScope, value.EvidenceLocator); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) RegisterPostal(ctx context.Context, dataset Dataset, values []PostalMapping) error {
	if !validDataset(dataset, DatasetPostal) || len(values) == 0 {
		return ErrInvalid
	}
	return s.registerDataset(ctx, dataset, func(tx pgx.Tx) error {
		for _, value := range values {
			if len(value.PostalCode) != 5 || len(value.MunicipalityISTAT) != 6 {
				return ErrInvalid
			}
			if _, err := tx.Exec(ctx, `INSERT INTO geography_postal_mappings(dataset_id,municipality_dataset_id,postal_code,municipality_istat) VALUES($1,$2,$3,$4)`, dataset.ID, dataset.MunicipalityDatasetID, value.PostalCode, value.MunicipalityISTAT); err != nil {
				return err
			}
		}
		return nil
	})
}

func slicesContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (s *Store) registerDataset(ctx context.Context, dataset Dataset, rows func(pgx.Tx) error) error {
	limitations, err := json.Marshal(dataset.Limitations)
	if err != nil {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if dataset.Kind != DatasetMunicipalities {
		var kind string
		if err = tx.QueryRow(ctx, "SELECT kind FROM geography_datasets WHERE id=$1", dataset.MunicipalityDatasetID).Scan(&kind); err != nil || kind != DatasetMunicipalities {
			return ErrInvalid
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO geography_datasets(id,kind,municipality_dataset_id,authority,publisher,official_url,version_label,source_sha256,verified_at,applicable_from,applicable_to,applicability,limitations,usable)
 VALUES($1,$2,NULLIF($3,''),$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, dataset.ID, dataset.Kind, dataset.MunicipalityDatasetID, dataset.Authority, dataset.Publisher, dataset.OfficialURL, dataset.VersionLabel, dataset.SourceSHA256, dataset.VerifiedAt.UTC(), dateValue(dataset.ApplicableFrom), dateValue(dataset.ApplicableTo), dataset.Applicability, limitations, dataset.Usable)
	if err != nil {
		return err
	}
	if err = rows(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func dateValue(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format("2006-01-02")
}

func (s *Store) SelectDataset(ctx context.Context, kind, datasetID, actor string, at time.Time) error {
	if !slicesContains([]string{DatasetMunicipalities, DatasetZones, DatasetPostal}, kind) || strings.TrimSpace(datasetID) == "" || strings.TrimSpace(actor) == "" || at.IsZero() {
		return ErrInvalid
	}
	var actual string
	var usable bool
	if err := s.pool.QueryRow(ctx, "SELECT kind,usable FROM geography_datasets WHERE id=$1", datasetID).Scan(&actual, &usable); err != nil {
		return err
	}
	if actual != kind || !usable {
		return ErrInvalid
	}
	_, err := s.pool.Exec(ctx, "INSERT INTO geography_dataset_selections(kind,dataset_id,actor,selected_at) VALUES($1,$2,$3,$4)", kind, datasetID, actor, at.UTC())
	return err
}

func (s *Store) selectedDataset(ctx context.Context, kind string) (string, bool, error) {
	var id string
	err := s.pool.QueryRow(ctx, "SELECT dataset_id FROM geography_dataset_selections WHERE kind=$1 ORDER BY id DESC LIMIT 1", kind).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return id, err == nil, err
}

func (s *Store) LookupMunicipality(ctx context.Context, input Lookup) (MunicipalityLookup, error) {
	if (strings.TrimSpace(input.Name) == "") == (strings.TrimSpace(input.ISTAT) == "") {
		return MunicipalityLookup{}, ErrInvalid
	}
	municipalityDataset := input.MunicipalityDatasetID
	if municipalityDataset == "" {
		var ok bool
		var err error
		municipalityDataset, ok, err = s.selectedDataset(ctx, DatasetMunicipalities)
		if err != nil {
			return MunicipalityLookup{}, err
		}
		if !ok {
			return MunicipalityLookup{Status: "mapping_unavailable"}, nil
		}
	}
	query, argument := "m.istat=$2", input.ISTAT
	if input.Name != "" {
		query, argument = "lower(m.name)=lower($2)", input.Name
	}
	rows, err := s.pool.Query(ctx, `SELECT m.istat,m.name,m.province,m.supported FROM geography_municipalities m WHERE m.dataset_id=$1 AND `+query+` ORDER BY m.istat`, municipalityDataset, argument)
	if err != nil {
		return MunicipalityLookup{}, err
	}
	defer rows.Close()
	var candidates []MunicipalityCandidate
	for rows.Next() {
		var candidate MunicipalityCandidate
		if err = rows.Scan(&candidate.ISTAT, &candidate.Name, &candidate.Province, &candidate.Supported); err != nil {
			return MunicipalityLookup{}, err
		}
		candidate.MunicipalityDatasetID = municipalityDataset
		candidates = append(candidates, candidate)
	}
	if err = rows.Err(); err != nil {
		return MunicipalityLookup{}, err
	}
	if len(candidates) == 0 {
		return MunicipalityLookup{Status: "not_found"}, nil
	}
	if len(candidates) == 1 && !candidates[0].Supported {
		return MunicipalityLookup{Status: "unsupported_area", Candidates: candidates}, nil
	}
	zoneDataset := input.ZoneDatasetID
	if zoneDataset == "" && !input.ZoneDatasetResolved {
		zoneDataset, _, err = s.selectedDataset(ctx, DatasetZones)
		if err != nil {
			return MunicipalityLookup{}, err
		}
	}
	for index := range candidates {
		if candidates[index].Supported {
			if err = s.attachZones(ctx, &candidates[index], zoneDataset); err != nil {
				return MunicipalityLookup{}, err
			}
		}
	}
	status := "ok"
	if len(candidates) > 1 {
		status = "ambiguous_municipality"
	}
	return MunicipalityLookup{Status: status, Candidates: candidates}, nil
}

func (s *Store) attachZones(ctx context.Context, candidate *MunicipalityCandidate, datasetID string) error {
	if datasetID == "" {
		candidate.MappingApplicability = "unavailable"
		candidate.MappingLimitations = []string{"municipality-zone mapping unavailable"}
		return nil
	}
	var municipalityDataset, applicability string
	var from, to *time.Time
	var raw []byte
	if err := s.pool.QueryRow(ctx, `SELECT municipality_dataset_id,applicable_from,applicable_to,applicability,limitations FROM geography_datasets WHERE id=$1 AND kind='zone_mapping' AND usable`, datasetID).Scan(&municipalityDataset, &from, &to, &applicability, &raw); err != nil {
		return err
	}
	if municipalityDataset != candidate.MunicipalityDatasetID {
		candidate.MappingApplicability = "unavailable"
		candidate.MappingLimitations = []string{"selected municipality-zone mapping uses another municipality registry version"}
		return nil
	}
	rows, err := s.pool.Query(ctx, `SELECT zone,CASE WHEN bool_or(territorial_scope='partial_municipality') THEN 'partial_municipality' ELSE 'whole_municipality' END FROM geography_zone_mappings WHERE dataset_id=$1 AND municipality_istat=$2 GROUP BY zone ORDER BY zone`, datasetID, candidate.ISTAT)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var zone ZoneApplicability
		if err = rows.Scan(&zone.Zone, &zone.TerritorialScope); err != nil {
			return err
		}
		candidate.Zones = append(candidate.Zones, zone)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	candidate.MappingVersionID, candidate.MappingApplicableFrom, candidate.MappingApplicableTo, candidate.MappingApplicability = &datasetID, from, to, applicability
	if json.Unmarshal(raw, &candidate.MappingLimitations) != nil {
		return ErrInvalid
	}
	return nil
}

func (s *Store) LookupPostal(ctx context.Context, postalCode string) (PostalLookup, error) {
	if len(postalCode) != 5 {
		return PostalLookup{}, ErrInvalid
	}
	datasetID, ok, err := s.selectedDataset(ctx, DatasetPostal)
	if err != nil {
		return PostalLookup{}, err
	}
	if !ok {
		return PostalLookup{Status: "mapping_unavailable"}, nil
	}
	var municipalityDataset string
	if err = s.pool.QueryRow(ctx, "SELECT municipality_dataset_id FROM geography_datasets WHERE id=$1 AND kind='postal_candidates' AND usable", datasetID).Scan(&municipalityDataset); err != nil {
		return PostalLookup{}, err
	}
	rows, err := s.pool.Query(ctx, `SELECT m.istat,m.name,m.province,m.supported FROM geography_postal_mappings p JOIN geography_municipalities m ON m.dataset_id=p.municipality_dataset_id AND m.istat=p.municipality_istat WHERE p.dataset_id=$1 AND p.postal_code=$2 ORDER BY m.istat`, datasetID, postalCode)
	if err != nil {
		return PostalLookup{}, err
	}
	defer rows.Close()
	result := PostalLookup{DatasetID: datasetID}
	for rows.Next() {
		var candidate MunicipalityCandidate
		if err = rows.Scan(&candidate.ISTAT, &candidate.Name, &candidate.Province, &candidate.Supported); err != nil {
			return PostalLookup{}, err
		}
		candidate.MunicipalityDatasetID = municipalityDataset
		if candidate.Supported {
			zoneDataset, _, selectedErr := s.selectedDataset(ctx, DatasetZones)
			if selectedErr != nil {
				return PostalLookup{}, selectedErr
			}
			if zoneDataset != "" {
				if err = s.attachZones(ctx, &candidate, zoneDataset); err != nil {
					return PostalLookup{}, err
				}
			}
		}
		result.Candidates = append(result.Candidates, candidate)
	}
	if err = rows.Err(); err != nil {
		return PostalLookup{}, err
	}
	if len(result.Candidates) == 0 {
		result.Status = "not_found"
		return result, nil
	}
	result.Status = "candidates"
	result.RequiresSelection = true
	result.Ambiguous = len(result.Candidates) > 1
	return result, nil
}
