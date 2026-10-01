package classification

import (
	"fmt"
	"strings"
	"testing"
)

func TestClassificationRunKeyPreservesExistingAndBoundsReprocessing(t *testing.T) {
	content := strings.Repeat("a", 64)
	configuration := "classification-config-9e3c22b8a9aa1171"
	manifest := ":" + strings.Repeat("b", 64)
	old := fmt.Sprintf("classification:%d:%s:%s:%s", 1234, content, configuration, "") + manifest
	if len(old) > 200 {
		t.Fatal("ordinary fixture already overflows")
	}
	if got := classificationRunKey(1234, content, configuration, "", manifest); got != old {
		t.Fatal("existing run identity changed")
	}
	selection := "reprocess-aaaaaaaaaaaaaaaaaaaa"
	long := fmt.Sprintf("classification:%d:%s:%s:%s", 1234, content, configuration, selection) + manifest
	if len(long) <= 200 {
		t.Fatal("fixture does not reproduce reprocessing overflow")
	}
	key := classificationRunKey(1234, content, configuration, selection, manifest)
	if len(key) > 200 || key == old || key != classificationRunKey(1234, content, configuration, selection, manifest) {
		t.Fatal("unstable or unbounded reprocessing identity")
	}
	variants := []string{
		classificationRunKey(1235, content, configuration, selection, manifest),
		classificationRunKey(1234, strings.Repeat("c", 64), configuration, selection, manifest),
		classificationRunKey(1234, content, configuration+"2", selection, manifest),
		classificationRunKey(1234, content, configuration, selection+"2", manifest),
		classificationRunKey(1234, content, configuration, selection, ":"+strings.Repeat("c", 64)),
	}
	for _, other := range variants {
		if key == other {
			t.Fatal("distinct input/configuration/selection collapsed")
		}
	}
}
