package notifications

import (
	"testing"
	"time"
)

func TestDailyReportSchedule(t *testing.T) {
	c := DailyReportConfig{Enabled: true, LocalTime: "21:00", Timezone: "Europe/Rome"}
	for _, day := range []string{"2026-03-29", "2026-10-25", "2026-09-26"} {
		loc, _ := time.LoadLocation(c.Timezone)
		at, _ := time.ParseInLocation("2006-01-02 15:04", day+" 21:00", loc)
		if date, due := c.Due(at); date != day || !due {
			t.Fatal("local schedule", date, due)
		}
		if _, due := c.Due(at.Add(-time.Second)); due {
			t.Fatal("early delivery")
		}
		if date, due := c.Due(at.Add(time.Hour)); date != day || !due {
			t.Fatal("restart recovery")
		}
	}
	for _, bad := range []DailyReportConfig{{true, "02:30", "Europe/Rome"}, {true, "21:00", "Local"}, {true, "21:00", "Bad/Zone"}, {true, "24:00", "Europe/Rome"}, {true, "21:0", "Europe/Rome"}} {
		if bad.Validate() == nil {
			t.Fatal("invalid schedule accepted")
		}
	}
	if _, due := (DailyReportConfig{}).Due(time.Now()); due {
		t.Fatal("default enabled")
	}
	if (Config{DailyReport: &c}).Validate() == nil {
		t.Fatal("report enabled without SMTP")
	}
}
