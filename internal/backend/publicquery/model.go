// Package publicquery exposes the shared, read-only application queries used by
// both the public JSON API and MCP transports. It never schedules acquisition
// or interpretation work.
package publicquery

import (
	"errors"
	"time"
)

var (
	ErrInvalidParameters     = errors.New("invalid public query parameters")
	ErrAmbiguousMunicipality = errors.New("ambiguous municipality")
	ErrUnknownIdentifier     = errors.New("unknown identifier")
	ErrUnsupportedArea       = errors.New("unsupported area")
	ErrMappingUnavailable    = errors.New("mapping unavailable")
)

type QueryTime struct {
	EvaluationTime time.Time
	KnownAt        time.Time
}

type History struct {
	PublicationScope string `json:"publication_scope,omitempty"`
	Start            *time.Time
	Gaps             []string
	Limitations      []string
}

type Temporal struct {
	Original   string     `json:"original"`
	Precision  string     `json:"precision"`
	Instant    *time.Time `json:"instant"`
	Date       *string    `json:"date"`
	EndInstant *time.Time `json:"end_instant"`
	Timezone   *string    `json:"timezone"`
	Assumption *string    `json:"assumption"`
	Condition  *string    `json:"-"`
}

type Evidence struct {
	DocumentID string  `json:"document_id"`
	VersionID  string  `json:"version_id"`
	SourceURL  string  `json:"source_url"`
	Page       *int    `json:"page"`
	Locator    *string `json:"locator"`
	Passage    *string `json:"passage"`
}

type Dimension struct {
	State       string     `json:"state"`
	Limitations []string   `json:"limitations"`
	Evidence    []Evidence `json:"evidence"`
}

type Updating struct {
	State                 string     `json:"state"`
	LastAttemptAt         *time.Time `json:"last_attempt_at"`
	LastCompleteAt        *time.Time `json:"last_complete_check_at"`
	DelayThresholdSeconds int        `json:"delay_threshold_seconds"`
	Limitations           []string   `json:"limitations"`
}

type Quality struct {
	Provenance     Dimension `json:"provenance"`
	Interpretation Dimension `json:"interpretation"`
	Updating       Updating  `json:"updating"`
}

type Municipality struct {
	DevelopmentPublication bool     `json:"development_publication"`
	ISTAT                  string   `json:"istat"`
	Name                   string   `json:"name"`
	Zones                  []string `json:"zones"`
	MappingVersion         *string  `json:"mapping_version"`
	MappingLimitations     []string `json:"mapping_limitations"`
	LocalCoverage          string   `json:"local_coverage"`
}

type DiscoveryQuery struct {
	QueryTime
	Name, ISTAT, PostalCode string
}

type DiscoveryResult struct {
	Municipalities []Municipality
	History        History
}

type Attachment struct {
	DocumentID  *string `json:"document_id"`
	OfficialURL string  `json:"official_url"`
	Status      string  `json:"status"`
}

type Document struct {
	Verifications       []Verification `json:"verifications,omitempty"`
	ID                  string         `json:"id"`
	VersionID           string         `json:"version_id"`
	SourceID            string         `json:"source_id"`
	Issuer              *string        `json:"issuer"`
	Publisher           string         `json:"publisher"`
	Kind                string         `json:"kind"`
	OfficialURL         string         `json:"official_url"`
	CopyURL             *string        `json:"copy_url"`
	SHA256              string         `json:"sha256"`
	Publication         Temporal       `json:"publication"`
	Modification        Temporal       `json:"modification"`
	AcquiredAt          time.Time      `json:"acquired_at"`
	InterpretedAt       *time.Time     `json:"interpreted_at"`
	InterpretationRunID *string        `json:"interpretation_run_id"`
	Quality             Quality        `json:"quality"`
	Attachments         []Attachment   `json:"attachments"`
}

type Measure struct {
	Verifications                 []Verification `json:"verifications,omitempty"`
	ID                            string         `json:"id"`
	RevisionID                    string         `json:"revision_id"`
	MunicipalityISTAT             string         `json:"municipality_istat"`
	Issuer                        *string        `json:"issuer"`
	Action                        string         `json:"action"`
	Place                         *string        `json:"place"`
	TerritorialScope              string         `json:"territorial_scope"`
	Validity                      Temporal       `json:"validity"`
	Status                        string         `json:"status"`
	Evidence                      []Evidence     `json:"evidence"`
	Quality                       Quality        `json:"quality"`
	RelatedMeasureIDs             []string       `json:"related_measure_ids"`
	RelationshipStatus            string         `json:"relationship_status"`
	NewerUninterpretedDocumentIDs []string       `json:"newer_uninterpreted_document_ids"`
}

type RegionalWarning struct {
	Verifications     []Verification `json:"verifications,omitempty"`
	ID                string         `json:"id"`
	Product           string         `json:"product"`
	Risk              string         `json:"risk"`
	OfficialRiskLabel string         `json:"official_risk_label"`
	Zone              string         `json:"zone"`
	MappingVersion    string         `json:"mapping_version"`
	Level             string         `json:"level"`
	Validity          Temporal       `json:"validity"`
	Status            string         `json:"status"`
	Evidence          []Evidence     `json:"evidence"`
	Quality           Quality        `json:"quality"`
}

type OperationalPhase struct {
	Verifications     []Verification `json:"verifications,omitempty"`
	ID                string         `json:"id"`
	MunicipalityISTAT string         `json:"municipality_istat"`
	Authority         string         `json:"authority"`
	Phase             string         `json:"phase"`
	Validity          Temporal       `json:"validity"`
	Evidence          []Evidence     `json:"evidence"`
	Quality           Quality        `json:"quality"`
}

type Coverage struct {
	DevelopmentPublication bool     `json:"development_publication"`
	SourceID               string   `json:"source_id"`
	Product                string   `json:"product"`
	Territory              string   `json:"territory"`
	DeclaredSections       []string `json:"declared_sections"`
	PublicState            string   `json:"public_state"`
	CoverageStatus         string   `json:"coverage_status"`
	CoverageLimitations    []string `json:"coverage_limitations"`
	Quality                Quality  `json:"quality"`
}

type SituationQuery struct {
	QueryTime
	MunicipalityISTAT string
}

type Situation struct {
	Municipality                Municipality       `json:"municipality"`
	LocalMeasures               []Measure          `json:"local_measures"`
	OperationalPhases           []OperationalPhase `json:"operational_phases"`
	RegionalProducts            []RegionalWarning  `json:"regional_products"`
	DocumentsRequiringAttention []Document         `json:"documents_requiring_attention"`
	Coverage                    []Coverage         `json:"coverage"`
	History                     History            `json:"-"`
}

type SearchQuery struct {
	QueryTime
	Kind, MunicipalityISTAT, Zone, Product, Risk, Status, SourceID string
	From, To                                                       *time.Time
}

type SearchResult struct {
	Kind      string
	Documents []Document
	Measures  []Measure
	Regional  []RegionalWarning
	History   History
}

type DocumentQuery struct {
	QueryTime
	DocumentID      string
	VersionID       string
	IncludeVersions bool
}

type DocumentResult struct {
	Document Document
	Versions []Document
	History  History
}

type CoverageQuery struct {
	QueryTime
	MunicipalityISTAT, SourceID, Product string
}

type CoverageResult struct {
	Sources []Coverage
	History History
}

func normalizeTime(value QueryTime) (QueryTime, error) {
	if value.EvaluationTime.IsZero() {
		value.EvaluationTime = time.Now().UTC()
	} else {
		value.EvaluationTime = value.EvaluationTime.UTC()
	}
	if value.KnownAt.IsZero() {
		value.KnownAt = value.EvaluationTime
	} else {
		value.KnownAt = value.KnownAt.UTC()
	}
	if value.KnownAt.After(value.EvaluationTime) {
		return QueryTime{}, ErrInvalidParameters
	}
	return value, nil
}

func unknownTemporal() Temporal { return Temporal{Precision: "unknown"} }
