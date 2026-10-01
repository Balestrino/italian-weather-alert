package diagnostics

import (
	"errors"
	"testing"
)

func text(value string) *string { return &value }

func validDimensions() []Dimension {
	return []Dimension{
		{Name: "risk", State: "same", RegionalValue: text("wind"), DPCValue: text("wind")},
		{Name: "territory", State: "different", RegionalValue: text("A4"), DPCValue: text("Toscana")},
		{Name: "issuance", State: "not_comparable", Reason: "the products have different issuance cycles"},
		{Name: "validity", State: "not_comparable", Reason: "the DPC snapshot has no matching interval"},
	}
}

func TestDimensionsRequireExactScopeAndExplicitNonComparability(t *testing.T) {
	if err := validateDimensions(validDimensions()); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func([]Dimension){
		"duplicate":                func(values []Dimension) { values[3].Name = "risk" },
		"missing reason":           func(values []Dimension) { values[2].Reason = "" },
		"values on incomparable":   func(values []Dimension) { values[2].RegionalValue = text("invented") },
		"missing comparable value": func(values []Dimension) { values[0].DPCValue = nil },
		"reason on comparable":     func(values []Dimension) { values[0].Reason = "unexpected" },
		"unknown state":            func(values []Dimension) { values[0].State = "similar" },
	} {
		t.Run(name, func(t *testing.T) {
			values := validDimensions()
			mutate(values)
			if !errors.Is(validateDimensions(values), ErrInvalid) {
				t.Fatal("invalid comparison dimension was accepted")
			}
		})
	}
}
