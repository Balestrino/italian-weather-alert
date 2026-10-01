package trialcost

import (
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func TestInputValidationAndOverflowSafeCosts(t *testing.T) {
	now := time.Now().UTC()
	quantity, unit, price := int64(2), int64(1), int64(3)
	input := Input{Category: "hosting", Workload: "ordinary", Component: "vm", Status: "reported", Basis: "rate", Quantity: &quantity, Unit: "hour", UnitSize: &unit, PriceMicrounits: &price, Currency: "EUR", EffectiveFrom: pointerTime(now.Add(-time.Hour)), EffectiveThrough: pointerTime(now.Add(time.Hour)), Evidence: registry.Evidence{URL: "https://evidence.example/rate", Locator: "retained quote", ObservedAt: now}, Notes: "fixture"}
	if !validInput(input) {
		t.Fatal("valid measured input rejected")
	}
	input.EffectiveThrough = input.EffectiveFrom
	if validInput(input) {
		t.Fatal("empty pricing interval accepted")
	}
	if _, ok := pricedCost(2, 3, 4); !ok {
		t.Fatal("bounded cost rejected")
	}
	if _, ok := pricedCost(1<<62, 1, 4); ok {
		t.Fatal("overflowing cost accepted")
	}
}

func pointerTime(value time.Time) *time.Time { return &value }
