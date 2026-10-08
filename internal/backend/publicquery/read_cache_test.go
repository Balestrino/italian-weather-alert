package publicquery

import (
	"testing"
	"time"
)

func TestQueryReadsTimeScopeAndIndependentLimitations(t *testing.T) {
	qt := QueryTime{EvaluationTime: time.Now().UTC(), KnownAt: time.Now().UTC()}
	shared := &Store{}
	first := shared.withQueryReads(qt)
	if first.reads == shared.reads || first.forMunicipality("050004").withQueryReads(qt).reads != first.reads {
		t.Fatal("nested request ownership")
	}
	if shared.withQueryReads(qt).reads == first.reads || first.withQueryReads(QueryTime{EvaluationTime: qt.EvaluationTime, KnownAt: qt.KnownAt.Add(-time.Second)}).reads == first.reads {
		t.Fatal("cache escaped its query boundary")
	}
	limits := make([]string, 1, 8)
	limits[0] = "source limitation"
	first.reads.quality[sourceReadKey{"source", ""}] = Quality{Provenance: Dimension{Limitations: limits}, Interpretation: Dimension{Limitations: []string{"source warning"}}}
	a, err := first.quality(t.Context(), "source", qt, Dimension{State: "supported", Limitations: []string{"fact A"}})
	if err != nil {
		t.Fatal(err)
	}
	a.Provenance.Limitations[0] = "changed"
	a.Interpretation.Limitations[0] = "changed"
	b, err := first.quality(t.Context(), "source", qt, Dimension{State: "partial", Limitations: []string{"fact B"}})
	if err != nil {
		t.Fatal(err)
	}
	if b.Provenance.Limitations[0] != "source limitation" || b.Interpretation.State != "partial" || len(b.Interpretation.Limitations) != 2 || b.Interpretation.Limitations[0] != "fact B" {
		t.Fatal("fact interpretations shared mutable state", b)
	}
}

func TestVerificationCutoffUsesPostgresTimestampPrecision(t *testing.T) {
	cutoff := time.Date(2026, 10, 8, 7, 0, 0, 123456000, time.UTC)
	qt := QueryTime{EvaluationTime: cutoff, KnownAt: cutoff}
	query := (&Store{}).withQueryReads(qt)
	query.reads.evidenceCutoffs[evidenceReadKey{42, "source", ""}] = &cutoff
	for _, test := range []struct {
		at   time.Time
		want bool
	}{
		{cutoff.Add(999 * time.Nanosecond), true},
		{cutoff.Add(time.Microsecond), false},
	} {
		got, err := query.verificationEvidenceVisible(t.Context(), 42, "source", qt.KnownAt, test.at)
		if err != nil || got != test.want {
			t.Fatalf("precision boundary at %s: visible=%v err=%v", test.at, got, err)
		}
	}
}
