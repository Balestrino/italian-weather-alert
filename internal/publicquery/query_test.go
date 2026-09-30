package publicquery

import (
	"errors"
	"testing"
	"time"
)

func TestNormalizeTimeAndSearchValidation(t *testing.T) {
	evaluation := time.Date(2026, 9, 17, 12, 0, 0, 0, time.FixedZone("test", 2*60*60))
	known := evaluation.Add(-time.Minute)
	value, err := normalizeTime(QueryTime{EvaluationTime: evaluation, KnownAt: known})
	if err != nil || value.EvaluationTime.Location() != time.UTC || value.KnownAt.Location() != time.UTC {
		t.Fatalf("normalize query time: %#v %v", value, err)
	}
	if _, err = normalizeTime(QueryTime{EvaluationTime: evaluation, KnownAt: evaluation.Add(time.Second)}); !errors.Is(err, ErrInvalidParameters) {
		t.Fatalf("future knowledge boundary accepted: %v", err)
	}
	store := &Store{}
	if _, err = store.Search(t.Context(), SearchQuery{Kind: "measure", Product: "criticality", QueryTime: QueryTime{EvaluationTime: evaluation}}); !errors.Is(err, ErrInvalidParameters) {
		t.Fatalf("incompatible filter accepted: %v", err)
	}
}

func TestTemporalStatusAndIntersection(t *testing.T) {
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	start, end := now.Add(time.Hour), now.Add(2*time.Hour)
	value := Temporal{Precision: "interval", Instant: &start, EndInstant: &end}
	if status := statusForTemporal(value, now); status != "future" {
		t.Fatalf("status=%s", status)
	}
	from, to := now.Add(90*time.Minute), now.Add(3*time.Hour)
	if !intersects(value, &from, &to) {
		t.Fatal("overlapping interval was excluded")
	}
	from = now.Add(3 * time.Hour)
	if intersects(value, &from, nil) {
		t.Fatal("non-overlapping interval was included")
	}
	if !intersects(unknownTemporal(), &from, nil) {
		t.Fatal("undetermined validity was silently excluded")
	}
}
