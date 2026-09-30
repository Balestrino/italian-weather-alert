package config

import (
	"os"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/processing"
)

func LoadProviderGatePolicy() (processing.GatePolicy, error) {
	p := processing.DefaultGatePolicy()
	threshold, err := boundedInt("IWA_PROVIDER_FAILURE_THRESHOLD", 3, 1, 100)
	if err != nil {
		return p, err
	}
	initial, err := boundedInt("IWA_PROVIDER_COOLDOWN_SECONDS", 60, 1, 86400)
	if err != nil {
		return p, err
	}
	maximum, err := boundedInt("IWA_PROVIDER_MAX_COOLDOWN_SECONDS", 900, initial, 86400)
	if err != nil {
		return p, err
	}
	p.FailureThreshold = threshold
	p.InitialCooldown = time.Duration(initial) * time.Second
	p.MaxCooldown = time.Duration(maximum) * time.Second
	return p, nil
}

// GateAlias names credential scope without deriving an identifier from key bytes.
// Explicit independent key files default to isolated scopes.
func GateAlias(prefix string) string {
	if alias := os.Getenv(prefix + "_GATE_SCOPE"); alias != "" {
		return alias
	}
	if file := os.Getenv(prefix + "_API_KEY_FILE"); file != "" && file != os.Getenv("IWA_REGOLO_API_KEY_FILE") {
		return prefix + "-private"
	}
	if alias := os.Getenv("IWA_REGOLO_GATE_SCOPE"); alias != "" {
		return alias
	}
	return "regolo-primary"
}
