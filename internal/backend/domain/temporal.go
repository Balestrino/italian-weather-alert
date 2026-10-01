package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrTemporalAmbiguous = errors.New("ambiguous or nonexistent Europe/Rome local time")

type TemporalValue struct {
	ID                              int64
	EntityKind, EntityID, Meaning   string
	Original, Precision             string
	Instant, Date, EndInstant       *time.Time
	Timezone, Assumption, Condition *string
	ConflictGroup                   *string
	EvidenceDocumentVersionID       int64
	CreatedAt                       time.Time
}

type MeasureUpdate struct {
	LinkingRunID                     int64
	UpdateMeasureID, TargetMeasureID string
	Relation                         string
	AppliedAt                        time.Time
}

type MeasureState struct {
	MeasureID       string
	Status          string
	UpdatedBy       []string
	UpdateRelations []string
}

func EuropeRomeInstant(original string, year int, month time.Month, day, hour, minute int, assumption string) (TemporalValue, error) {
	if strings.TrimSpace(original) == "" || strings.TrimSpace(assumption) == "" {
		return TemporalValue{}, ErrInvalid
	}
	location, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		return TemporalValue{}, fmt.Errorf("load Europe/Rome: %w", err)
	}
	candidate := time.Date(year, month, day, hour, minute, 0, 0, location)
	matches := map[int64]time.Time{}
	for _, value := range []time.Time{candidate.Add(-time.Hour), candidate, candidate.Add(time.Hour)} {
		local := value.In(location)
		if local.Year() == year && local.Month() == month && local.Day() == day && local.Hour() == hour && local.Minute() == minute {
			matches[value.Unix()] = value
		}
	}
	if len(matches) != 1 {
		return TemporalValue{}, ErrTemporalAmbiguous
	}
	var instant time.Time
	for _, value := range matches {
		instant = value.UTC()
	}
	timezone := "Europe/Rome"
	return TemporalValue{Original: original, Precision: "instant", Instant: &instant, Timezone: &timezone, Assumption: &assumption}, nil
}

func validTemporal(value TemporalValue) bool {
	if !slicesContains([]string{"local_measure", "operational_phase", "regional_record"}, value.EntityKind) || strings.TrimSpace(value.EntityID) == "" || !slicesContains([]string{"publication", "source_modification", "event", "validity", "page_expiry", "platform", "acquisition", "attestation"}, value.Meaning) || !slicesContains([]string{"instant", "date", "interval", "conditional", "unknown"}, value.Precision) || value.EvidenceDocumentVersionID < 1 || value.CreatedAt.IsZero() {
		return false
	}
	if (value.Timezone == nil) != (value.Assumption == nil) || (value.Assumption != nil && strings.TrimSpace(*value.Assumption) == "") {
		return false
	}
	switch value.Precision {
	case "instant":
		return value.Instant != nil && value.Date == nil && value.EndInstant == nil && value.Condition == nil
	case "date":
		return value.Instant == nil && value.Date != nil && value.EndInstant == nil && value.Timezone == nil && value.Condition == nil
	case "interval":
		return value.Instant != nil && value.Date == nil && value.EndInstant != nil && !value.EndInstant.Before(*value.Instant) && value.Condition == nil
	case "conditional":
		return value.Date == nil && value.EndInstant == nil && value.Condition != nil && strings.TrimSpace(*value.Condition) != ""
	case "unknown":
		return value.Instant == nil && value.Date == nil && value.EndInstant == nil && value.Timezone == nil
	default:
		return false
	}
}
