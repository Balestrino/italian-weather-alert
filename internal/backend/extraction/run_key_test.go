package extraction

import (
	"fmt"
	"strings"
	"testing"
)

func TestExtractionRunKeyPreservesExistingAndBoundsLocalProcessing(t *testing.T) {
	content := strings.Repeat("a", 64)
	configuration := "extraction-config-9e3c22b8a9aa1171"
	manifest := ":" + strings.Repeat("b", 64)
	old := fmt.Sprintf("extraction:%d:%d:%s:%s", 1234, 5678, content, configuration) + manifest
	if len(old) > 200 {
		t.Fatal("ordinary fixture already overflows")
	}
	if got := extractionRunKey(1234, 5678, content, configuration, manifest); got != old {
		t.Fatal("existing run identity changed")
	}
	local := "local-processing-" + strings.Repeat("d", 64)
	long := fmt.Sprintf("extraction:%d:%d:%s:%s", 1234, 5678, content, local) + manifest
	if len(long) <= 200 {
		t.Fatal("fixture does not reproduce local-processing overflow")
	}
	key := extractionRunKey(1234, 5678, content, local, manifest)
	if len(key) > 200 || key == old || key != extractionRunKey(1234, 5678, content, local, manifest) {
		t.Fatal("unstable or unbounded local-processing identity")
	}
	variants := []string{
		extractionRunKey(1235, 5678, content, local, manifest),
		extractionRunKey(1234, 5679, content, local, manifest),
		extractionRunKey(1234, 5678, strings.Repeat("c", 64), local, manifest),
		extractionRunKey(1234, 5678, content, local+"2", manifest),
		extractionRunKey(1234, 5678, content, local, ":"+strings.Repeat("c", 64)),
	}
	for _, other := range variants {
		if key == other {
			t.Fatal("distinct input/configuration collapsed")
		}
	}
}
