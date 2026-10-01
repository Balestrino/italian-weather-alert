package publicquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	collectedOnly    bool
	snapshotDatasets map[string]string
	region           string
	scopeTime        time.Time
	pool             interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
		QueryRow(context.Context, string, ...any) pgx.Row
	}
	geography *domain.Store
}

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool, geography: domain.New(pool)} }

func (s *Store) DiscoverMunicipalities(ctx context.Context, query DiscoveryQuery) (DiscoveryResult, error) {
	qt, err := normalizeTime(query.QueryTime)
	if err != nil || countSet(query.Name, query.ISTAT, query.PostalCode) > 1 || (query.ISTAT != "" && len(query.ISTAT) != 6) || (query.PostalCode != "" && len(query.PostalCode) != 5) {
		return DiscoveryResult{}, ErrInvalidParameters
	}
	query.QueryTime = qt
	var candidates []domain.MunicipalityCandidate
	switch {
	case query.PostalCode != "":
		candidates, err = s.postalCandidates(ctx, query.PostalCode, qt.KnownAt)
		if err != nil {
			return DiscoveryResult{}, err
		}
	case query.Name != "" || query.ISTAT != "":
		municipalityDataset, lookupErr := s.selectedDatasetAt(ctx, domain.DatasetMunicipalities, qt.KnownAt)
		if lookupErr != nil {
			return DiscoveryResult{}, lookupErr
		}
		zoneDataset, lookupErr := s.selectedDatasetAt(ctx, domain.DatasetZones, qt.KnownAt)
		if errors.Is(lookupErr, ErrMappingUnavailable) {
			zoneDataset, lookupErr = "", nil
		}
		if lookupErr != nil {
			return DiscoveryResult{}, lookupErr
		}
		lookup, lookupErr := s.geography.LookupMunicipality(ctx, domain.Lookup{Name: query.Name, ISTAT: query.ISTAT, MunicipalityDatasetID: municipalityDataset, ZoneDatasetID: zoneDataset, ZoneDatasetResolved: true})
		if lookupErr != nil {
			return DiscoveryResult{}, lookupErr
		}
		switch lookup.Status {
		case "mapping_unavailable":
			return DiscoveryResult{}, ErrMappingUnavailable
		case "not_found":
			return DiscoveryResult{}, ErrUnknownIdentifier
		case "unsupported_area":
			return DiscoveryResult{}, ErrUnsupportedArea
		case "ambiguous_municipality":
			candidates = lookup.Candidates
			result, convertErr := s.municipalities(ctx, candidates)
			return DiscoveryResult{Municipalities: result}, errors.Join(ErrAmbiguousMunicipality, convertErr)
		}
		candidates = lookup.Candidates
	default:
		candidates, err = s.listMunicipalities(ctx, qt.KnownAt)
		if err != nil {
			return DiscoveryResult{}, err
		}
	}
	values, err := s.municipalities(ctx, candidates)
	if err != nil {
		return DiscoveryResult{}, err
	}
	history, err := s.history(ctx, qt.KnownAt, nil, historyScope{})
	return DiscoveryResult{Municipalities: values, History: history}, err
}

func countSet(values ...string) int {
	count := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			count++
		}
	}
	return count
}

func (s *Store) selectedDatasetAt(ctx context.Context, kind string, knownAt time.Time) (string, error) {
	if s.snapshotDatasets != nil {
		if id := s.snapshotDatasets[kind]; id != "" {
			return id, nil
		}
		return "", ErrMappingUnavailable
	}
	if s.region != "" {
		id, err := s.geography.TerritorialDatasetAt(ctx, s.region, kind, knownAt)
		if errors.Is(err, domain.ErrTerritoryNotFound) {
			return "", ErrMappingUnavailable
		}
		return id, err
	}

	var id string
	err := s.pool.QueryRow(ctx, `SELECT dataset_id FROM geography_dataset_selections WHERE kind=$1 AND registry_dataset_in_public_scope(dataset_id) AND selected_at<=$2 ORDER BY id DESC LIMIT 1`, kind, knownAt).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrMappingUnavailable
	}
	return id, err
}

func (s *Store) listMunicipalities(ctx context.Context, knownAt time.Time) ([]domain.MunicipalityCandidate, error) {
	municipalityDataset, err := s.selectedDatasetAt(ctx, domain.DatasetMunicipalities, knownAt)
	if err != nil {
		return nil, err
	}
	zoneDataset, err := s.selectedDatasetAt(ctx, domain.DatasetZones, knownAt)
	if errors.Is(err, ErrMappingUnavailable) {
		zoneDataset, err = "", nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT istat,name,province,supported FROM geography_municipalities WHERE dataset_id=$1 ORDER BY istat`, municipalityDataset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.MunicipalityCandidate
	for rows.Next() {
		var value domain.MunicipalityCandidate
		if err = rows.Scan(&value.ISTAT, &value.Name, &value.Province, &value.Supported); err != nil {
			return nil, err
		}
		value.MunicipalityDatasetID = municipalityDataset
		lookup, lookupErr := s.geography.LookupMunicipality(ctx, domain.Lookup{ISTAT: value.ISTAT, MunicipalityDatasetID: municipalityDataset, ZoneDatasetID: zoneDataset, ZoneDatasetResolved: true})
		if lookupErr != nil {
			return nil, lookupErr
		}
		if len(lookup.Candidates) == 1 {
			value = lookup.Candidates[0]
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (s *Store) postalCandidates(ctx context.Context, postalCode string, knownAt time.Time) ([]domain.MunicipalityCandidate, error) {
	postalDataset, err := s.selectedDatasetAt(ctx, domain.DatasetPostal, knownAt)
	if err != nil {
		return nil, err
	}
	var municipalityDataset string
	if err = s.pool.QueryRow(ctx, `SELECT municipality_dataset_id FROM geography_datasets WHERE id=$1 AND kind='postal_candidates' AND usable`, postalDataset).Scan(&municipalityDataset); errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMappingUnavailable
	} else if err != nil {
		return nil, err
	}
	zoneDataset, err := s.selectedDatasetAt(ctx, domain.DatasetZones, knownAt)
	if errors.Is(err, ErrMappingUnavailable) {
		zoneDataset, err = "", nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT municipality_istat FROM geography_postal_mappings WHERE dataset_id=$1 AND postal_code=$2 ORDER BY municipality_istat`, postalDataset, postalCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var istatCodes []string
	for rows.Next() {
		var istat string
		if err = rows.Scan(&istat); err != nil {
			return nil, err
		}
		istatCodes = append(istatCodes, istat)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(istatCodes) == 0 {
		return nil, ErrUnknownIdentifier
	}
	result := make([]domain.MunicipalityCandidate, 0, len(istatCodes))
	for _, istat := range istatCodes {
		lookup, lookupErr := s.geography.LookupMunicipality(ctx, domain.Lookup{ISTAT: istat, MunicipalityDatasetID: municipalityDataset, ZoneDatasetID: zoneDataset, ZoneDatasetResolved: true})
		if lookupErr != nil {
			return nil, lookupErr
		}
		result = append(result, lookup.Candidates...)
	}
	return result, nil
}

func (s *Store) municipalities(ctx context.Context, candidates []domain.MunicipalityCandidate) ([]Municipality, error) {
	result := make([]Municipality, 0, len(candidates))
	for _, candidate := range candidates {
		coverage, err := s.localCoverage(ctx, candidate.ISTAT)
		if err != nil {
			return nil, err
		}
		value := Municipality{ISTAT: candidate.ISTAT, Name: candidate.Name, MappingVersion: candidate.MappingVersionID, MappingLimitations: nonNil(candidate.MappingLimitations), LocalCoverage: coverage, Zones: []string{}}
		for _, zone := range candidate.Zones {
			value.Zones = append(value.Zones, zone.Zone)
		}
		result = append(result, value)
	}
	return result, nil
}

func (s *Store) municipality(ctx context.Context, istat string, knownAt time.Time) (Municipality, error) {
	if len(istat) != 6 {
		return Municipality{}, ErrInvalidParameters
	}
	municipalityDataset, err := s.selectedDatasetAt(ctx, domain.DatasetMunicipalities, knownAt)
	if err != nil {
		return Municipality{}, err
	}
	zoneDataset, err := s.selectedDatasetAt(ctx, domain.DatasetZones, knownAt)
	if errors.Is(err, ErrMappingUnavailable) {
		zoneDataset, err = "", nil
	}
	if err != nil {
		return Municipality{}, err
	}
	lookup, err := s.geography.LookupMunicipality(ctx, domain.Lookup{ISTAT: istat, MunicipalityDatasetID: municipalityDataset, ZoneDatasetID: zoneDataset, ZoneDatasetResolved: true})
	if err != nil {
		return Municipality{}, err
	}
	switch lookup.Status {
	case "mapping_unavailable":
		return Municipality{}, ErrMappingUnavailable
	case "unsupported_area":
		return Municipality{}, ErrUnsupportedArea
	case "not_found":
		return Municipality{}, ErrUnknownIdentifier
	}
	values, err := s.municipalities(ctx, lookup.Candidates)
	if err != nil {
		return Municipality{}, err
	}
	if len(values) != 1 {
		return Municipality{}, ErrUnknownIdentifier
	}
	return values[0], nil
}

func (s *Store) localCoverage(ctx context.Context, istat string) (string, error) {
	var total, enabled, accepted int
	err := s.pool.QueryRow(ctx, `SELECT count(*),count(*) FILTER (WHERE s.public_enabled),count(*) FILTER (WHERE EXISTS(
 SELECT 1 FROM registry_events e JOIN registry_acceptance_reviews rr ON rr.acceptance_event_id=e.id
 WHERE e.source_id=s.id AND e.revision=s.active_revision AND e.kind='acceptance' AND rr.status='accepted'
 AND e.id=(SELECT max(latest.id) FROM registry_events latest WHERE latest.source_id=e.source_id AND latest.revision=e.revision AND latest.kind='acceptance')
 AND NOT EXISTS(SELECT 1 FROM registry_regressions r WHERE r.source_id=e.source_id AND r.revision=e.revision AND NOT r.passed AND r.recorded_at>=e.created_at)))
 FROM `+s.sourcesSQL()+` s WHERE s.product_id='municipal' AND s.territory=$1`, istat).Scan(&total, &enabled, &accepted)
	if err != nil {
		return "", err
	}
	if enabled > 0 {
		return "enabled", nil
	}
	if accepted > 0 {
		return "suspended", nil
	}
	if total > 0 {
		return "pending", nil
	}
	return "outside_scope", nil
}

type historyScope struct {
	DocumentID   int64
	Municipality string
	SourceID     string
	Product      string
}

func (s *Store) history(ctx context.Context, knownAt time.Time, requestedFrom *time.Time, scope historyScope) (History, error) {
	var start *time.Time
	err := s.pool.QueryRow(ctx, `SELECT min(v.first_acquired_at)
 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 JOIN `+s.sourcesSQL()+` s ON s.id=d.source_id JOIN registry_products p ON p.id=s.product_id
	 WHERE p.public_eligible AND `+s.visibilitySQL()+` AND v.first_acquired_at<=$1
	   AND ($2::bigint=0 OR d.id=$2) AND ($3='' OR s.territory=$3 OR s.product_id<>'municipal')
	   AND ($4='' OR s.id=$4) AND ($5='' OR s.product_id=$5)`, knownAt, scope.DocumentID, scope.Municipality, scope.SourceID, scope.Product).Scan(&start)
	if err != nil {
		return History{}, err
	}
	result := History{Start: utcPointer(start), Gaps: []string{}, Limitations: []string{"history reflects retained service knowledge and is not a complete pre-startup archive"}}
	if requestedFrom != nil && (start == nil || requestedFrom.Before(*start)) {
		result.Gaps = append(result.Gaps, "requested interval begins before retained service history")
	}
	return result, nil
}

func utcPointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	result := value.UTC()
	return &result
}

func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func stringID(value int64) string { return strconv.FormatInt(value, 10) }

func (s *Store) Coverage(ctx context.Context, query CoverageQuery) (CoverageResult, error) {
	qt, err := normalizeTime(query.QueryTime)
	if err != nil || (query.MunicipalityISTAT != "" && len(query.MunicipalityISTAT) != 6) {
		return CoverageResult{}, ErrInvalidParameters
	}
	rows, err := s.pool.Query(ctx, `SELECT s.id,s.product_id,s.territory,s.public_enabled,s.collection_enabled,
 COALESCE(EXISTS(SELECT 1 FROM registry_events e JOIN registry_acceptance_reviews rr ON rr.acceptance_event_id=e.id
  WHERE e.source_id=s.id AND e.revision=s.active_revision AND e.kind='acceptance' AND rr.status='accepted'
  AND e.id=(SELECT max(latest.id) FROM registry_events latest WHERE latest.source_id=e.source_id AND latest.revision=e.revision AND latest.kind='acceptance')
  AND NOT EXISTS(SELECT 1 FROM registry_regressions r WHERE r.source_id=e.source_id AND r.revision=e.revision AND NOT r.passed AND r.recorded_at>=e.created_at)),false),
	 c.body,
	 COALESCE((SELECT rr.report FROM registry_acceptance_reviews rr JOIN registry_events e ON e.id=rr.acceptance_event_id
	  WHERE rr.source_id=s.id AND rr.revision=s.active_revision AND rr.status='accepted'
	  AND e.id=(SELECT max(latest.id) FROM registry_events latest WHERE latest.source_id=s.id AND latest.revision=s.active_revision AND latest.kind='acceptance')
	  AND NOT EXISTS(SELECT 1 FROM registry_regressions r WHERE r.source_id=s.id AND r.revision=s.active_revision AND NOT r.passed AND r.recorded_at>=e.created_at)
	  ORDER BY rr.id DESC LIMIT 1),'{}'::jsonb)
 FROM `+s.sourcesSQL()+` s JOIN registry_products p ON p.id=s.product_id
 JOIN registry_configurations c ON c.source_id=s.id AND c.revision=COALESCE(s.active_revision,s.latest_revision)
 WHERE p.public_eligible AND ($1='' OR s.id=$1) AND ($2='' OR s.product_id=$2)
   AND ($3='' OR s.territory=$3 OR s.product_id<>'municipal')
 ORDER BY CASE WHEN s.product_id='municipal' THEN 0 ELSE 1 END,s.id`, query.SourceID, query.Product, query.MunicipalityISTAT)
	if err != nil {
		return CoverageResult{}, err
	}
	defer rows.Close()
	result := CoverageResult{Sources: []Coverage{}}
	for rows.Next() {
		var value Coverage
		var publicEnabled, collectionEnabled, accepted bool
		var raw, releaseRaw []byte
		if err = rows.Scan(&value.SourceID, &value.Product, &value.Territory, &publicEnabled, &collectionEnabled, &accepted, &raw, &releaseRaw); err != nil {
			return CoverageResult{}, err
		}
		var configuration struct {
			Sections    []string `json:"sections"`
			Limitations []string `json:"limitations"`
		}
		if json.Unmarshal(raw, &configuration) != nil {
			return CoverageResult{}, fmt.Errorf("decode source configuration: %w", ErrInvalidParameters)
		}
		value.DeclaredSections = nonNil(configuration.Sections)
		var review struct {
			CoverageStatus      string   `json:"coverage_status"`
			CoverageLimitations []string `json:"coverage_limitations"`
		}
		if json.Unmarshal(releaseRaw, &review) != nil {
			return CoverageResult{}, fmt.Errorf("decode release review: %w", ErrInvalidParameters)
		}
		value.CoverageStatus = "pending"
		value.CoverageLimitations = nonNil(configuration.Limitations)
		if accepted {
			value.CoverageStatus = review.CoverageStatus
			for _, limitation := range review.CoverageLimitations {
				if !slices.Contains(value.CoverageLimitations, limitation) {
					value.CoverageLimitations = append(value.CoverageLimitations, limitation)
				}
			}
		}
		switch {
		case publicEnabled:
			value.PublicState = "enabled"
		case accepted:
			value.PublicState = "suspended"
		default:
			value.PublicState = "pending"
		}
		interpretation := Dimension{State: "not_processed", Limitations: []string{"source acceptance is pending"}, Evidence: []Evidence{}}
		if accepted {
			interpretation = Dimension{State: "supported", Limitations: []string{"source acceptance is recorded separately from individual facts"}, Evidence: []Evidence{}}
			if value.Product == "criticality" || value.Product == "vigilance" || value.Product == "monitoring" {
				var status, statement string
				var versionID int64
				var limits []string
				projectionErr := s.pool.QueryRow(ctx, `SELECT p.status,p.statement,p.limitations,p.document_version_id FROM domain_regional_projections p JOIN retained_versions v ON v.id=p.document_version_id WHERE p.source_id=$1 AND p.projected_at<=registry_interpretation_cutoff($1,$2) AND v.first_acquired_at<=$2 ORDER BY v.first_acquired_at DESC,v.id DESC,p.projected_at DESC LIMIT 1`, value.SourceID, qt.KnownAt).Scan(&status, &statement, &limits, &versionID)
				if projectionErr == nil {
					interpretation.State = "partial"
					if status == "unsupported" {
						interpretation.State = "failed"
					}
					if status == "no_event" {
						interpretation.State = "supported"
					}
					interpretation.Limitations = append(interpretation.Limitations, limits...)
					if statement != "" {
						interpretation.Limitations = append(interpretation.Limitations, "Retained bulletin statement: "+statement)
					}
					evidence, e := s.versionEvidence(ctx, versionID, "regional projection: "+status)
					if e != nil {
						return CoverageResult{}, e
					}
					interpretation.Evidence = []Evidence{evidence}
				} else if !errors.Is(projectionErr, pgx.ErrNoRows) {
					return CoverageResult{}, projectionErr
				}
			}
		}
		value.Quality, err = s.quality(ctx, value.SourceID, qt, interpretation)
		if err != nil {
			return CoverageResult{}, err
		}
		value.Quality.Provenance.Limitations = append(value.Quality.Provenance.Limitations, configuration.Limitations...)
		if !collectionEnabled {
			value.Quality.Updating.State = "suspended"
			value.Quality.Updating.Limitations = append(value.Quality.Updating.Limitations, "collection is suspended")
		}
		result.Sources = append(result.Sources, value)
	}
	if err = rows.Err(); err != nil {
		return CoverageResult{}, err
	}
	result.History, err = s.history(ctx, qt.KnownAt, nil, historyScope{Municipality: query.MunicipalityISTAT, SourceID: query.SourceID, Product: query.Product})
	return result, err
}

func (s *Store) quality(ctx context.Context, sourceID string, qt QueryTime, interpretation Dimension) (Quality, error) {
	result := Quality{Interpretation: interpretation}
	var suspended bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM registry_interpretation_suspensions WHERE source_id=$1 AND (resumed_at IS NULL OR (suspended_at<=$2 AND resumed_at>$2)))`, sourceID, qt.KnownAt).Scan(&suspended); err != nil {
		return Quality{}, err
	}
	if suspended {
		result.Interpretation.State = "unreliable"
		result.Interpretation.Limitations = append(result.Interpretation.Limitations, "confirmed source interpretation defect; new interpretations are suspended pending correction, validation and explicit reprocessing")
	}
	result.Interpretation.Limitations = nonNil(result.Interpretation.Limitations)
	result.Interpretation.Evidence = nonNilEvidence(result.Interpretation.Evidence)
	var raw []byte
	var url, locator string
	err := s.pool.QueryRow(ctx, `SELECT state,evidence_url,evidence_locator,limitations
 FROM domain_provenance_events WHERE source_id=$1 AND assessed_at<=$2 ORDER BY assessed_at DESC,id DESC LIMIT 1`, sourceID, qt.KnownAt).Scan(&result.Provenance.State, &url, &locator, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		result.Provenance = Dimension{State: "unresolved", Limitations: []string{"no provenance assessment was available at the knowledge boundary"}, Evidence: []Evidence{}}
	} else if err != nil {
		return Quality{}, err
	} else {
		if json.Unmarshal(raw, &result.Provenance.Limitations) != nil {
			return Quality{}, ErrInvalidParameters
		}
		result.Provenance.Evidence = []Evidence{}
		result.Provenance.Limitations = append(result.Provenance.Limitations, "provenance evidence: "+url+" ("+locator+")")
	}
	var lastAttempt, lastComplete *time.Time
	var reachable, recognized, complete bool
	var errorCode *string
	err = s.pool.QueryRow(ctx, `WITH latest AS (
	 SELECT started_at,reachable,content_recognized,complete,error_code
	 FROM acquisition_checks WHERE source_id=$1 AND finished_at<=$2 ORDER BY finished_at DESC,id DESC LIMIT 1)
	 SELECT latest.started_at,
	   (SELECT max(finished_at) FROM acquisition_checks WHERE source_id=$1 AND finished_at<=$2 AND complete),
	   latest.reachable,latest.content_recognized,latest.complete,latest.error_code FROM latest`, sourceID, qt.KnownAt).Scan(&lastAttempt, &lastComplete, &reachable, &recognized, &complete, &errorCode)
	var delaySeconds int
	var collectionEnabled bool
	if scanErr := s.pool.QueryRow(ctx, `SELECT s.collection_enabled,COALESCE(s.delay_seconds_override,(c.body->>'delay_seconds')::integer,1800)
 FROM `+s.sourcesSQL()+` s JOIN registry_configurations c ON c.source_id=s.id AND c.revision=COALESCE(s.active_revision,s.latest_revision) WHERE s.id=$1`, sourceID).Scan(&collectionEnabled, &delaySeconds); scanErr != nil {
		return Quality{}, scanErr
	}
	result.Updating = Updating{LastAttemptAt: utcPointer(lastAttempt), LastCompleteAt: utcPointer(lastComplete), DelayThresholdSeconds: delaySeconds, Limitations: []string{}}
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		result.Updating.State = "not_verified"
		result.Updating.Limitations = []string{"no source check was available at the knowledge boundary"}
	case err != nil:
		return Quality{}, err
	case !collectionEnabled:
		result.Updating.State = "suspended"
		result.Updating.Limitations = []string{"collection is suspended; retained validity is unchanged"}
	case !complete || !reachable || !recognized || errorCode != nil:
		result.Updating.State = "failed"
		result.Updating.Limitations = []string{"latest source check was incomplete; fact validity is evaluated separately"}
	case lastComplete != nil && !qt.EvaluationTime.Before(lastComplete.Add(time.Duration(delaySeconds)*time.Second)):
		result.Updating.State = "delayed"
		result.Updating.Limitations = []string{"complete-check delay threshold reached; fact validity is evaluated separately"}
	default:
		result.Updating.State = "ok"
	}
	return result, nil
}

func nonNilEvidence(values []Evidence) []Evidence {
	if values == nil {
		return []Evidence{}
	}
	return values
}

func statusForTemporal(value Temporal, evaluatedAt time.Time) string {
	switch value.Precision {
	case "interval":
		if value.Instant != nil && evaluatedAt.Before(*value.Instant) {
			return "future"
		}
		if value.EndInstant != nil && !evaluatedAt.Before(*value.EndInstant) {
			return "expired"
		}
		return "current"
	case "instant":
		if value.Instant != nil && evaluatedAt.Before(*value.Instant) {
			return "future"
		}
	}
	return "undetermined"
}

func sortStatus(status string) int {
	return slices.Index([]string{"current", "undetermined", "future", "expired", "cancelled", "superseded"}, status)
}

// Resuming a source never silently rehabilitates an older defective result.
// Corrected processing writes new versioned results/assessments instead.
func (s *Store) preserveDefectWarning(ctx context.Context, versionID int64, interpretedAt time.Time, result Dimension) (Dimension, error) {
	var defective bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM registry_interpretation_suspensions x JOIN retained_documents d ON d.source_id=x.source_id JOIN retained_versions v ON v.document_id=d.id WHERE v.id=$1 AND x.suspended_at>=$2)`, versionID, interpretedAt).Scan(&defective)
	if defective {
		result.State = "unreliable"
		result.Limitations = append(result.Limitations, "this interpretation predates a confirmed source defect; validated resumption does not rewrite earlier results")
	}
	return result, err
}
