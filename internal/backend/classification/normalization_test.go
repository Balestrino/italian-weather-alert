package classification

import (
	"strings"
	"testing"
)

func TestNormalizeOCRPreservesReversibleRetainedOffsets(t *testing.T) {
	raw := "##  **ORDINA**\r\n\tla chiusura del sottopasso `Verde`  fino a revoca"
	view := NormalizeOCR(raw)
	want := "ORDINA la chiusura del sottopasso Verde fino a revoca"
	if view.Text != want {
		t.Fatalf("unexpected normalized OCR view: %q", view.Text)
	}
	start := strings.Index(view.Text, "ORDINA")
	end := strings.Index(view.Text, " fino")
	sourceStart, sourceEnd, ok := view.SourceRange(start, end)
	if !ok || raw[sourceStart:sourceEnd] != "ORDINA**\r\n\tla chiusura del sottopasso `Verde" {
		t.Fatalf("normalized evidence did not resolve to retained OCR: %d:%d %q", sourceStart, sourceEnd, raw[sourceStart:sourceEnd])
	}
	for offset := 0; offset < len(view.Text); offset++ {
		covered := false
		for _, span := range view.OffsetMap {
			if offset >= span.NormalizedStart && offset < span.NormalizedEnd {
				covered = true
				break
			}
		}
		if !covered {
			t.Fatalf("normalized byte %d has no retained-source mapping", offset)
		}
	}
}

func TestNormalizeOCRCanonicalizesMappedApostrophes(t *testing.T) {
	raw := "all’evolversi dellʼallerta e dell‘evento"
	view := NormalizeOCR(raw)
	if view.Text != "all'evolversi dell'allerta e dell'evento" || !view.Valid() {
		t.Fatalf("apostrophes were not normalized reversibly: %#v", view)
	}
	start := strings.Index(view.Text, "dell'allerta")
	sourceStart, sourceEnd, ok := view.SourceRange(start, start+len("dell'allerta"))
	if !ok || raw[sourceStart:sourceEnd] != "dellʼallerta" {
		t.Fatalf("normalized apostrophe did not map to retained OCR: %d:%d %v", sourceStart, sourceEnd, ok)
	}
}
