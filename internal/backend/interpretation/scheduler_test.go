package interpretation

import (
	"testing"
	"time"
)

func TestUnchangedContentDoesNotScheduleAndSelectionsAreScoped(t *testing.T) {
	s := &Scheduler{}
	ok, err := s.Automatic(nil, AcquisitionEvent{DocumentVersionID: 1, EvidenceHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Workload: "ordinary", At: time.Now()})
	if err != nil || ok {
		t.Fatalf("unchanged content scheduled: %v %v", ok, err)
	}
	at := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)
	a := selectionID(Selection{SourceID: "calcinaia", Actor: "operator", At: at})
	b := selectionID(Selection{SourceID: "pisa", Actor: "operator", At: at})
	if a == b || a == "" {
		t.Fatal("selection scope is not preserved")
	}
}
