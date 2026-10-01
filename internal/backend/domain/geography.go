package domain

import (
	"slices"
	"strings"
	"time"
)

const (
	DatasetMunicipalities = "municipality_registry"
	DatasetZones          = "zone_mapping"
	DatasetPostal         = "postal_candidates"
)

type Dataset struct {
	ID, Kind, MunicipalityDatasetID   string
	Authority, Publisher, OfficialURL string
	VersionLabel, SourceSHA256        string
	VerifiedAt                        time.Time
	ApplicableFrom, ApplicableTo      *time.Time
	Applicability                     string
	Limitations                       []string
	Usable                            bool
}

type Municipality struct {
	ISTAT, Name, Province string
	Supported             bool
}

type ZoneMapping struct {
	MunicipalityISTAT string
	Zone              string
	SourceName        string
	TerritorialScope  string
	EvidenceLocator   string
}

type PostalMapping struct{ PostalCode, MunicipalityISTAT string }

type Lookup struct {
	Name, ISTAT                          string
	MunicipalityDatasetID, ZoneDatasetID string
	// ZoneDatasetResolved distinguishes an intentionally unavailable mapping at
	// a historical knowledge boundary from a request to use the current selection.
	ZoneDatasetResolved bool
}

type ZoneApplicability struct {
	Zone, TerritorialScope string
}

type MunicipalityCandidate struct {
	Municipality
	Zones                                      []ZoneApplicability
	MunicipalityDatasetID                      string
	MappingVersionID                           *string
	MappingApplicableFrom, MappingApplicableTo *time.Time
	MappingApplicability                       string
	MappingLimitations                         []string
}

type MunicipalityLookup struct {
	Status     string
	Candidates []MunicipalityCandidate
}

type PostalLookup struct {
	Status, DatasetID string
	Candidates        []MunicipalityCandidate
	RequiresSelection bool
	Ambiguous         bool
}

type LevelSummary struct {
	Level, AppliesTo string
	Zones            []string
	PartialTerritory bool
}

func validDataset(value Dataset, kind string) bool {
	if value.Kind != kind || strings.TrimSpace(value.ID) == "" || strings.TrimSpace(value.Authority) == "" || strings.TrimSpace(value.Publisher) == "" || strings.TrimSpace(value.OfficialURL) == "" || strings.TrimSpace(value.VersionLabel) == "" || len(value.SourceSHA256) != 64 || value.VerifiedAt.IsZero() || !slices.Contains([]string{"verified", "unresolved", "not_applicable"}, value.Applicability) {
		return false
	}
	if kind == DatasetMunicipalities {
		return value.MunicipalityDatasetID == ""
	}
	return value.MunicipalityDatasetID != ""
}

func SummarizeMunicipalityLevel(candidate MunicipalityCandidate, levels map[string]string) (LevelSummary, bool) {
	priority := map[string]int{"not_applicable": 0, "unknown": 1, "green": 2, "yellow": 3, "orange": 4, "red": 5}
	best, bestPriority := "", -1
	for _, zone := range candidate.Zones {
		level, ok := levels[zone.Zone]
		if !ok {
			continue
		}
		value, valid := priority[level]
		if !valid {
			continue
		}
		if value > bestPriority {
			best, bestPriority = level, value
		}
	}
	if best == "" {
		return LevelSummary{}, false
	}
	result := LevelSummary{Level: best, AppliesTo: "municipality", Zones: []string{}}
	for _, zone := range candidate.Zones {
		if levels[zone.Zone] == best {
			result.Zones = append(result.Zones, zone.Zone)
			if zone.TerritorialScope == "partial_municipality" {
				result.PartialTerritory = true
			}
		}
	}
	result.PartialTerritory = result.PartialTerritory || len(result.Zones) < len(candidate.Zones)
	if result.PartialTerritory {
		result.AppliesTo = "part_of_municipality"
	}
	return result, true
}
