package domain

import (
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"testing"
	"time"
)

func TestMunicipalStartHasVerifiedYearWeekdayAndUnknownEnd(t *testing.T) {
	start := "dalle ore 17.00 di mercoledì 7 ottobre"
	m := extraction.Measure{Kind: "activation", ValidFrom: &start}
	v := municipalValidity(m, 1, "fixture", time.Now(), "2026-10-07")
	if v.Precision != "instant" || v.Instant == nil || v.Instant.Format(time.RFC3339) != "2026-10-07T15:00:00Z" || v.EndInstant != nil || v.Assumption == nil || v.Timezone == nil || !validTemporal(v) {
		t.Fatal(v)
	}
	for _, test := range []struct{ expression, publication string }{
		{start, "2025-10-07"}, {start, "2026-10-06"}, {"dalle ore 25.00 di mercoledì 7 ottobre", "2026-10-07"}, {"dalle ore 17.00 di giovedì 7 ottobre", "2026-10-07"}, {start, ""}, {"dalle ore 17.00 di domenica 31 febbraio 2026", "2026-02-28"},
	} {
		m.ValidFrom = &test.expression
		if v := municipalValidity(m, 1, "fixture", time.Now(), test.publication); v.Precision != "unknown" {
			t.Fatal("invented operative time", test, v)
		}
	}
}
