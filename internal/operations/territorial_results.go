package operations

import (
	"context"
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/domain"
	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
	"time"
)

type TerritorialDocument struct {
	SourceID, Product, URL     string
	VersionID                  int64
	AcquiredAt                 time.Time
	Public                     bool
	Classification, Extraction string
	ErrorCode                  *string
}
type TerritorialResults struct {
	Alerts               Alerts
	Documents            []TerritorialDocument
	MoreDocuments        bool
	MappingUnavailable   bool
	Zones                []domain.ZoneApplicability
	DocumentsUnavailable bool
}

func (s *Store) TerritorialResults(ctx context.Context, region, istat string, at time.Time) (TerritorialResults, error) {
	result := TerritorialResults{Alerts: newAlerts(at), Documents: []TerritorialDocument{}, Zones: []domain.ZoneApplicability{}}
	if _, err := domain.New(s.pool).Region(ctx, region); err != nil {
		return result, err
	}
	if istat != "" {
		if _, err := s.Municipality(ctx, region, istat); err != nil {
			return result, err
		}
		geo := domain.New(s.pool)
		municipalityDataset, e := geo.TerritorialDatasetAt(ctx, region, domain.DatasetMunicipalities, at)
		zoneDataset, zerr := geo.TerritorialDatasetAt(ctx, region, domain.DatasetZones, at)
		result.MappingUnavailable = e != nil || zerr != nil
		if !result.MappingUnavailable {
			lookup, err := geo.LookupMunicipality(ctx, domain.Lookup{ISTAT: istat, MunicipalityDatasetID: municipalityDataset, ZoneDatasetID: zoneDataset, ZoneDatasetResolved: true})
			if err != nil {
				return result, err
			}
			for _, candidate := range lookup.Candidates {
				for _, z := range candidate.Zones {
					result.Zones = append(result.Zones, z)
				}
			}
			if len(result.Zones) == 0 {
				result.MappingUnavailable = true
			}
		}
	}
	// Limit the document preview only; counts and history readers cover the full territory.
	rows, err := s.pool.Query(ctx, `SELECT s.id,s.product_id,d.official_url,v.id,v.first_acquired_at,s.public_enabled,COALESCE(c.status,'not_processed'),COALESCE(e.status,'not_processed'),st.last_error_code,v.metadata
 FROM retained_documents d JOIN registry_sources s ON s.id=d.source_id
 JOIN LATERAL(SELECT * FROM territorial_source_associations a WHERE a.source_id=s.id AND a.recorded_at<=$3 ORDER BY a.recorded_at DESC,a.id DESC LIMIT 1)a ON a.region_code=$1
 JOIN LATERAL(SELECT * FROM retained_versions WHERE document_id=d.id AND first_acquired_at<=$3 ORDER BY first_acquired_at DESC,id DESC LIMIT 1)v ON true
 LEFT JOIN acquisition_source_status st ON st.source_id=s.id
 LEFT JOIN LATERAL(SELECT cr.status FROM classification_results cr JOIN processing_runs pr ON pr.id=cr.run_id WHERE pr.document_version_id=v.id AND pr.created_at<=$3 ORDER BY pr.created_at DESC,pr.id DESC LIMIT 1)c ON true
 LEFT JOIN LATERAL(SELECT er.status FROM extraction_results er JOIN processing_runs pr ON pr.id=er.run_id WHERE pr.document_version_id=v.id AND pr.created_at<=$3 ORDER BY pr.created_at DESC,pr.id DESC LIMIT 1)e ON true
 WHERE ($2='' OR a.municipality_istat=$2 OR a.municipality_istat IS NULL) ORDER BY v.first_acquired_at DESC,v.id DESC LIMIT 101`, region, istat, at)
	if err == nil {
		for rows.Next() {
			var d TerritorialDocument
			var raw json.RawMessage
			if e := rows.Scan(&d.SourceID, &d.Product, &d.URL, &d.VersionID, &d.AcquiredAt, &d.Public, &d.Classification, &d.Extraction, &d.ErrorCode, &raw); e != nil {
				rows.Close()
				return result, e
			}
			if len(result.Documents) == 100 {
				result.MoreDocuments = true
				continue
			}
			result.Documents = append(result.Documents, d)
			if d.Product == "municipal" {
				continue
			}
			var metadata acquisition.RegionalObservation
			if json.Unmarshal(raw, &metadata) != nil {
				continue
			}
			bulletin := AlertBulletin{Source: d.SourceID, Product: d.Product, URL: d.URL, Version: d.VersionID, AcquiredAt: d.AcquiredAt, Issued: metadata.IssuanceExpression, Validity: metadata.ValidityExpressions, NoCriticalityText: metadata.TextSaysNoCriticality}
			assigned := false
			for i := range result.Alerts.Days {
				for _, expression := range bulletin.Validity {
					date, ok := bulletinDate(expression)
					if ok && date == result.Alerts.Days[i].Start.Format(time.DateOnly) {
						result.Alerts.Days[i].Bulletins = append(result.Alerts.Days[i].Bulletins, bulletin)
						assigned = true
						break
					}
				}
			}
			if !assigned {
				result.Alerts.UndatedBulletins = append(result.Alerts.UndatedBulletins, bulletin)
			}
		}
		err = rows.Err()
		rows.Close()
	}
	if err != nil {
		result.DocumentsUnavailable = true
		result.Documents = []TerritorialDocument{}
		for i := range result.Alerts.Days {
			result.Alerts.Days[i].Bulletins = nil
		}
		result.Alerts.UndatedBulletins = nil
	}
	if len(result.Documents) > 100 {
		result.MoreDocuments = true
		result.Documents = result.Documents[:100]
	}
	warnings, measures, err := publicquery.ReadTerritorialFacts(ctx, s.pool, region, istat, at)
	if err != nil {
		result.Alerts.FactsUnavailable = true
		return result, nil
	}
	for _, w := range warnings {
		if w.Status == "cancelled" || w.Status == "superseded" {
			continue
		}
		known := false
		for i := range result.Alerts.Days {
			match, k := temporalDay(w.Validity, result.Alerts.Days[i])
			known = known || k
			if match {
				result.Alerts.Days[i].Regional = append(result.Alerts.Days[i].Regional, w)
			}
		}
		if !known {
			result.Alerts.UndatedRegional = append(result.Alerts.UndatedRegional, w)
		}
	}
	for _, m := range measures {
		if m.Status == "cancelled" || m.Status == "superseded" {
			continue
		}
		known := false
		for i := range result.Alerts.Days {
			match, k := temporalDay(m.Validity, result.Alerts.Days[i])
			known = known || k
			if match {
				result.Alerts.Days[i].Measures = append(result.Alerts.Days[i].Measures, m)
			}
		}
		if !known {
			result.Alerts.UndatedMeasures = append(result.Alerts.UndatedMeasures, m)
		}
	}
	return result, nil
}
