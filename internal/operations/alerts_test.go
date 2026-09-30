package operations

import (
	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
	"testing"
	"time"
)

func TestAlertDates(t *testing.T) {
	for _, s := range []string{"Mercoledí, 23 Settembre 2026", "23 settembre 2026"} {
		if d, ok := bulletinDate(s); !ok || d != "2026-09-23" {
			t.Fatal(s, d, ok)
		}
	}
	for _, s := range []string{"31 febbraio 2026", "domani", "23/09", "dal 23 settembre 2026"} {
		if _, ok := bulletinDate(s); ok {
			t.Fatal(s)
		}
	}
	for _, tc := range []struct {
		at, date string
		hours    int
	}{{"2026-09-23T22:15:00Z", "24/09/2026", 24}, {"2026-03-29T08:00:00Z", "29/03/2026", 23}, {"2026-10-25T08:00:00Z", "25/10/2026", 25}} {
		at, _ := time.Parse(time.RFC3339, tc.at)
		r := newAlerts(at)
		d := r.Days[0]
		if d.Date != tc.date || d.End.Sub(d.Start) != time.Duration(tc.hours)*time.Hour || r.Days[1].Start != d.End {
			t.Fatal(r)
		}
		end := d.End
		start := d.Start.Add(-time.Hour)
		if m, k := temporalDay(publicquery.Temporal{Precision: "instant", Instant: &end}, d); m || !k {
			t.Fatal("exclusive end")
		}
		if m, k := temporalDay(publicquery.Temporal{Precision: "interval", Instant: &start, EndInstant: &end}, d); !m || !k {
			t.Fatal("overlap")
		}
		if m, k := temporalDay(publicquery.Temporal{Precision: "conditional"}, d); m || k {
			t.Fatal("unknown")
		}
		date := d.Start.Format(time.DateOnly)
		if m, k := temporalDay(publicquery.Temporal{Precision: "date", Date: &date}, d); !m || !k {
			t.Fatal("date")
		}
	}
}
