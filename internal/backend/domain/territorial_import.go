package domain

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const MaxMunicipalityImportBytes = 4 << 20

var municipalityCodePattern = regexp.MustCompile(`^[0-9]{6}$`)

type MunicipalityImport struct {
	Region               string    `json:"region"`
	OfficialURL          string    `json:"official_url"`
	Version              string    `json:"version"`
	VerifiedAt           time.Time `json:"verified_at"`
	ExpectedCount        int       `json:"expected_count"`
	Complete             bool      `json:"complete"`
	CompletenessEvidence string    `json:"completeness_evidence"`
}
type MunicipalityImportPreview struct {
	Metadata       MunicipalityImport
	SHA256         string
	Municipalities []Municipality
	Complete       bool
}

// PreviewMunicipalities validates all rows. It does not write or select a dataset.
func PreviewMunicipalities(body []byte, meta MunicipalityImport) (MunicipalityImportPreview, error) {
	p := MunicipalityImportPreview{Metadata: meta, Municipalities: []Municipality{}}
	u, e := url.Parse(meta.OfficialURL)
	if len(body) == 0 || len(body) > MaxMunicipalityImportBytes || e != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || strings.TrimSpace(meta.Version) == "" || meta.VerifiedAt.IsZero() || meta.ExpectedCount < 1 || strings.TrimSpace(meta.CompletenessEvidence) == "" {
		return p, ErrInvalid
	}
	ref, err := csv.NewReader(strings.NewReader(municipalityReference)).ReadAll()
	if err != nil {
		return p, err
	}
	known := map[string]string{}
	validRegion := false
	for _, r := range ref[1:] {
		known[r[0]] = r[2]
		validRegion = validRegion || r[2] == meta.Region
	}
	if !validRegion {
		return p, ErrInvalid
	}
	reader := csv.NewReader(bytes.NewReader(body))
	header, err := reader.Read()
	if err != nil {
		return p, ErrInvalid
	}
	columns := map[string]int{}
	for i, name := range header {
		if _, ok := columns[name]; ok {
			return p, ErrInvalid
		}
		columns[name] = i
	}
	for _, key := range []string{"istat", "comune", "codice_regione", "sigla"} {
		if _, ok := columns[key]; !ok {
			return p, ErrInvalid
		}
	}
	seen := map[string]bool{}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return p, ErrInvalid
		}
		istat := row[columns["istat"]]
		region := row[columns["codice_regione"]]
		name, province := strings.TrimSpace(row[columns["comune"]]), strings.TrimSpace(row[columns["sigla"]])
		if !municipalityCodePattern.MatchString(istat) || seen[istat] || region != meta.Region || name == "" || province == "" || len(name) > 300 || len(province) > 100 {
			return p, ErrInvalid
		}
		if expected, ok := known[istat]; ok && expected != region {
			return p, ErrInvalid
		}
		seen[istat] = true
		p.Municipalities = append(p.Municipalities, Municipality{ISTAT: istat, Name: name, Province: province, Supported: true})
	}
	if len(p.Municipalities) == 0 || len(p.Municipalities) > 10000 || len(p.Municipalities) > meta.ExpectedCount {
		return p, ErrInvalid
	}
	hash := sha256.Sum256(body)
	p.SHA256 = hex.EncodeToString(hash[:])
	p.Complete = meta.Complete && len(p.Municipalities) == meta.ExpectedCount
	return p, nil
}
func (s *Store) AdoptMunicipalities(ctx context.Context, body []byte, meta MunicipalityImport, previewHash string, expectedRevision int, actor string) (string, error) {
	preview, err := PreviewMunicipalities(body, meta)
	if err != nil {
		return "", err
	}
	if !preview.Complete || preview.SHA256 != previewHash || strings.TrimSpace(actor) == "" {
		return "", ErrInvalid
	}
	raw, _ := json.Marshal(meta)
	h := sha256.Sum256(append(raw, body...))
	id := "territory-" + meta.Region + "-" + hex.EncodeToString(h[:])
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	var revision int
	var enabled bool
	if err = tx.QueryRow(ctx, `SELECT revision,enabled FROM territorial_regions WHERE code=$1 FOR UPDATE`, meta.Region).Scan(&revision, &enabled); err != nil {
		return "", err
	}
	if revision != expectedRevision {
		return "", ErrConflict
	}
	cfg := RegionConfiguration{Profiles: []string{}}
	if revision > 0 {
		if err = tx.QueryRow(ctx, `SELECT configuration FROM territorial_region_versions WHERE region_code=$1 AND revision=$2`, meta.Region, revision).Scan(&cfg); err != nil {
			return "", err
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO geography_datasets(id,kind,authority,publisher,official_url,version_label,source_sha256,verified_at,applicability,limitations,usable) VALUES($1,'municipality_registry','ISTAT','ISTAT',$2,$3,$4,$5,'verified','[]',true)`, id, meta.OfficialURL, meta.Version, preview.SHA256, meta.VerifiedAt); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO territorial_dataset_regions VALUES($1,$2,true,$3)`, id, meta.Region, meta.CompletenessEvidence); err != nil {
		return "", err
	}
	for _, m := range preview.Municipalities {
		if _, err = tx.Exec(ctx, `INSERT INTO geography_municipalities VALUES($1,$2,$3,$4,true)`, id, m.ISTAT, m.Name, m.Province); err != nil {
			return "", err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO territorial_municipalities VALUES($1,$2,$3)`, meta.Region, id, m.ISTAT); err != nil {
			return "", err
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO territorial_dataset_selections(region_code,kind,dataset_id,actor) VALUES($1,'municipality_registry',$2,$3)`, meta.Region, id, actor); err != nil {
		return "", err
	}
	// Old mappings refer to a different immutable municipality snapshot. Keep them in history.
	cfg.MunicipalityDataset = id
	cfg.ZoneDataset = ""
	cfg.PostalDataset = ""
	if _, err = writeRegion(ctx, tx, meta.Region, revision, cfg, enabled, actor, "dataset_adopted"); err != nil {
		return "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", err
	}
	return id, nil
}
