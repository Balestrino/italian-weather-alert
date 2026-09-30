package domain

import (
	"errors"
	"testing"
	"time"
)

func TestClosedRiskAndProductCatalogs(t *testing.T) {
	if len(Risks) != 7 || len(Products) != 3 || len(Levels) != 6 {
		t.Fatalf("unexpected catalogs: risks=%v products=%v levels=%v", Risks, Products, Levels)
	}
	orange := "orange"
	monitoring := RegionalRecord{ID: "m", DocumentVersionID: 1, SourceID: "s", Product: ProductMonitoring, OriginatingAuthorityID: "a", Facts: []RegionalFact{{Ordinal: 1, Risk: RiskWind, OfficialRiskLabel: "Vento", Level: &orange}}}
	if validRegional(monitoring) {
		t.Fatal("monitoring must not issue a warning level")
	}
	monitoring.Facts[0].Level = nil
	if !validRegional(monitoring) {
		t.Fatal("level-free monitoring should be valid")
	}
}

func TestEuropeRomeInstantRequiresAnUnambiguousJustifiedWallTime(t *testing.T) {
	value, err := EuropeRomeInstant("20 agosto 2026 alle 18:00", 2026, time.August, 20, 18, 0, "local municipal act in Calcinaia")
	if err != nil || value.Instant == nil || value.Instant.Format(time.RFC3339) != "2026-08-20T16:00:00Z" || value.Timezone == nil || *value.Timezone != "Europe/Rome" {
		t.Fatalf("Europe/Rome conversion failed: %#v %v", value, err)
	}
	if _, err = EuropeRomeInstant("25 ottobre 2026 alle 02:30", 2026, time.October, 25, 2, 30, "local municipal act"); !errors.Is(err, ErrTemporalAmbiguous) {
		t.Fatalf("DST overlap was resolved without evidence: %v", err)
	}
	if _, err = EuropeRomeInstant("29 marzo 2026 alle 02:30", 2026, time.March, 29, 2, 30, "local municipal act"); !errors.Is(err, ErrTemporalAmbiguous) {
		t.Fatalf("nonexistent local time was resolved: %v", err)
	}
}
