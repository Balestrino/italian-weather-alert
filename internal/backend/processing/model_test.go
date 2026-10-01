package processing

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
)

func integer(value int64) *int64 { return &value }

func TestUsageValidationPreservesUnknown(t *testing.T) {
	quantities, other, err := validateUsage(Usage{Status: "reported", InputTokens: integer(0), OutputTokens: integer(12), OtherUnits: map[string]int64{"pages": 2}})
	if err != nil || quantities["input_tokens"] != 0 || quantities["pages"] != 2 || string(other) != `{"pages":2}` {
		t.Fatalf("reported usage rejected: %v %#v %s", err, quantities, other)
	}
	if _, _, err = validateUsage(Usage{Status: "unavailable", InputTokens: integer(0)}); !errors.Is(err, ErrInvalid) {
		t.Fatal("unknown usage was silently converted to zero")
	}
	if _, _, err = validateUsage(Usage{Status: "partial"}); !errors.Is(err, ErrInvalid) {
		t.Fatal("empty partial usage accepted")
	}
}

func TestCostRequiresCompletePricing(t *testing.T) {
	quantities := map[string]int64{"input_tokens": 1500, "output_tokens": 1}
	rates := map[string]Rate{
		"input_tokens":  {Metric: "input_tokens", UnitSize: 1000, PriceMicrounits: 100},
		"output_tokens": {Metric: "output_tokens", UnitSize: 1000, PriceMicrounits: 200},
	}
	if cost, ok := calculateCost(quantities, rates); !ok || cost != 151 {
		t.Fatalf("unexpected rounded cost: %d %v", cost, ok)
	}
	delete(rates, "output_tokens")
	if _, ok := calculateCost(quantities, rates); ok {
		t.Fatal("partial pricing produced an estimated total")
	}
	if _, ok := calculateCost(map[string]int64{"huge": math.MaxInt64}, map[string]Rate{"huge": {Metric: "huge", UnitSize: 1, PriceMicrounits: math.MaxInt64}}); ok {
		t.Fatal("overflowed cost accepted")
	}
}

func TestCanonicalObject(t *testing.T) {
	a, ah, err := canonicalObject(json.RawMessage(`{"b":2,"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	b, bh, err := canonicalObject(json.RawMessage(` {"a":1,"b":2} `))
	if err != nil || string(a) != string(b) || ah != bh {
		t.Fatal("equivalent configuration objects differ")
	}
}
