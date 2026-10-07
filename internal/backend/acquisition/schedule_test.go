package acquisition

import (
	"testing"
	"time"
)

func TestExpectedPublicationAcceptsEarlyIssuanceWithoutCarryingOlderCycles(t *testing.T) {
	anchor := time.Date(2026, 10, 7, 11, 0, 0, 0, time.UTC)
	interval, tolerance := 86400, 7200
	checked := anchor.Add(5 * time.Hour)
	for _, test := range []struct {
		name      string
		issued    time.Time
		checked   time.Time
		tolerance int
		want      string
	}{
		{"vigilance early issuance", anchor.Add(-81 * time.Minute), checked, tolerance, "observed"},
		{"early boundary", anchor.Add(-2 * time.Hour), checked, tolerance, "observed"},
		{"outside early tolerance", anchor.Add(-2*time.Hour - time.Second), checked, tolerance, "missing"},
		{"previous day", anchor.Add(-24 * time.Hour), checked, tolerance, "missing"},
		{"next day needs new issuance", anchor.Add(-81 * time.Minute), checked.Add(24 * time.Hour), tolerance, "missing"},
		{"zero tolerance", anchor.Add(-time.Second), checked, 0, "missing"},
		{"oversized tolerance cannot accept old cycle", anchor.Add(-24 * time.Hour), checked.Add(72 * time.Hour), 86400 * 3, "awaiting"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := publicationStateAt(&anchor, &interval, &test.tolerance, &test.issued, test.checked); got != test.want {
				t.Fatalf("publication state %q, want %q", got, test.want)
			}
		})
	}
	if got := publicationStateAt(&anchor, &interval, &tolerance, nil, anchor.Add(time.Hour)); got != "awaiting" {
		t.Fatalf("missing issuance before deadline: %q", got)
	}
}
