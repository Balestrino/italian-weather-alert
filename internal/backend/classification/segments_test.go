package classification

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSegmentContentIsDeterministicBoundedAndPositioned(t *testing.T) {
	text := Normalize(strings.Repeat("pioggia è prevista sul territorio comunale ", 900))
	content := Content{Complete: true, Sections: []ContentSection{{ResourceURL: "https://comune.example/a", Role: "attachment", Page: 2, Text: text}}}
	first, err := SegmentContent(91, content)
	if err != nil {
		t.Fatal(err)
	}
	second, err := SegmentContent(91, content)
	if err != nil || len(first) < 2 || len(first) != len(second) {
		t.Fatalf("deterministic segmentation unavailable: %d %d %v", len(first), len(second), err)
	}
	var rebuilt strings.Builder
	for index, segment := range first {
		if !reflect.DeepEqual(segment, second[index]) || segment.Ordinal != index+1 || segment.Total != len(first) || segment.DocumentVersionID != 91 || segment.ResourceURL != "https://comune.example/a" || segment.Page != 2 || len(segment.Text) > MaxSegmentTextBytes || !utf8.ValidString(segment.Text) || segment.Hash == "" {
			t.Fatalf("invalid segment %d: %#v", index, segment)
		}
		if text[segment.StartByte:segment.EndByte] != segment.Text {
			t.Fatalf("segment %d lost its original byte interval", index)
		}
		if index > 0 {
			rebuilt.WriteByte(' ')
		}
		rebuilt.WriteString(segment.Text)
	}
	if rebuilt.String() != text {
		t.Fatal("segmentation omitted or duplicated normalized content")
	}
}
