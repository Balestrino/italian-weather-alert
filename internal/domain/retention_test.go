package domain

import (
	"testing"
	"time"
)

func TestCalendarExpiryClampsMonthEndInUTC(t *testing.T) {
	tests := []struct {
		first, want time.Time
	}{
		{time.Date(2026, 1, 31, 9, 17, 0, 123, time.FixedZone("source", 3600)), time.Date(2026, 4, 30, 8, 17, 0, 123, time.UTC)},
		{time.Date(2026, 11, 30, 23, 0, 0, 0, time.UTC), time.Date(2027, 2, 28, 23, 0, 0, 0, time.UTC)},
		{time.Date(2027, 11, 30, 23, 0, 0, 0, time.UTC), time.Date(2028, 2, 29, 23, 0, 0, 0, time.UTC)},
	}
	for _, test := range tests {
		got, err := CalendarExpiry(test.first, 3)
		if err != nil || !got.Equal(test.want) || got.Location() != time.UTC {
			t.Fatalf("CalendarExpiry(%s): got %s (%v), want %s", test.first, got, err, test.want)
		}
	}
}

func TestCalendarExpiryRejectsInvalidInput(t *testing.T) {
	if _, err := CalendarExpiry(time.Time{}, 3); err == nil {
		t.Fatal("zero acquisition time accepted")
	}
	if _, err := CalendarExpiry(time.Now(), 0); err == nil {
		t.Fatal("zero retention accepted")
	}
}
