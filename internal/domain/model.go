// Package domain persists the public-safety facts derived from retained
// evidence without merging regional warnings, local measures, and operational
// phases into one ambiguous alert record.
package domain

import (
	"errors"
	"slices"
	"strings"
)

var (
	ErrInvalid  = errors.New("invalid domain record")
	ErrConflict = errors.New("domain record conflict")
)

const (
	RiskMinorNetworkHydro    = "minor_network_hydro"
	RiskMainNetworkHydraulic = "main_network_hydraulic"
	RiskThunderstorms        = "thunderstorms"
	RiskWind                 = "wind"
	RiskCoastalWaves         = "coastal_waves"
	RiskSnow                 = "snow"
	RiskIce                  = "ice"

	ProductVigilance   = "vigilance"
	ProductCriticality = "criticality"
	ProductMonitoring  = "monitoring"
)

var Risks = []string{
	RiskMinorNetworkHydro,
	RiskMainNetworkHydraulic,
	RiskThunderstorms,
	RiskWind,
	RiskCoastalWaves,
	RiskSnow,
	RiskIce,
}

var Products = []string{ProductVigilance, ProductCriticality, ProductMonitoring}
var Levels = []string{"green", "yellow", "orange", "red", "unknown", "not_applicable"}

type RegionalFact struct {
	Ordinal           int
	Risk              string
	OfficialRiskLabel string
	Zone              string
	Level             *string
}

type RegionalRecord struct {
	ID                     string
	DocumentVersionID      int64
	SourceID               string
	Product                string
	OriginatingAuthorityID string
	PublisherID            string
	Platform               string
	MunicipalRepublication bool
	Facts                  []RegionalFact
}

type LocalMeasure struct {
	ID, SourceID, MunicipalityISTAT string
	DocumentVersionID               int64
	IssuingAuthorityID              string
	PublisherID, Platform           string
	Kind, Subject                   string
	Place                           *string
	RegionalRecordID                *string
}

type OperationalPhase struct {
	ID, SourceID, MunicipalityISTAT string
	DocumentVersionID               int64
	AuthorityID                     string
	PublisherID, Platform           string
	Phase                           string
	RegionalRecordID                *string
}

func validRegional(value RegionalRecord) bool {
	if strings.TrimSpace(value.ID) == "" || value.DocumentVersionID < 1 || strings.TrimSpace(value.SourceID) == "" || !slices.Contains(Products, value.Product) || strings.TrimSpace(value.OriginatingAuthorityID) == "" || len(value.Facts) == 0 {
		return false
	}
	seen := map[int]bool{}
	for _, fact := range value.Facts {
		if fact.Ordinal < 1 || seen[fact.Ordinal] || !slices.Contains(Risks, fact.Risk) || strings.TrimSpace(fact.OfficialRiskLabel) == "" {
			return false
		}
		seen[fact.Ordinal] = true
		if fact.Level != nil && !slices.Contains(Levels, *fact.Level) {
			return false
		}
		if value.Product == ProductMonitoring && fact.Level != nil {
			return false
		}
	}
	return true
}

func validLocal(value LocalMeasure) bool {
	return strings.TrimSpace(value.ID) != "" && value.DocumentVersionID > 0 && strings.TrimSpace(value.SourceID) != "" && strings.TrimSpace(value.MunicipalityISTAT) != "" && strings.TrimSpace(value.Kind) != "" && strings.TrimSpace(value.Subject) != ""
}

func validPhase(value OperationalPhase) bool {
	return strings.TrimSpace(value.ID) != "" && value.DocumentVersionID > 0 && strings.TrimSpace(value.SourceID) != "" && strings.TrimSpace(value.MunicipalityISTAT) != "" && strings.TrimSpace(value.Phase) != ""
}
