package classification

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// OffsetSpan maps a half-open byte interval in normalized text to the
// corresponding half-open byte interval in the retained source text.
// Spans are ordered, non-overlapping and cover every normalized byte.
type OffsetSpan struct {
	NormalizedStart int `json:"normalized_start"`
	NormalizedEnd   int `json:"normalized_end"`
	SourceStart     int `json:"source_start"`
	SourceEnd       int `json:"source_end"`
}

// NormalizedText is the deterministic inference view of retained OCR text.
// Text never replaces the retained OCR output; OffsetMap resolves accepted
// normalized evidence back to that output's byte offsets.
type NormalizedText struct {
	Text      string
	OffsetMap []OffsetSpan
}

func (n NormalizedText) Valid() bool {
	if n.Text == "" || !utf8.ValidString(n.Text) || len(n.OffsetMap) == 0 {
		return false
	}
	normalizedCursor, sourceCursor := 0, 0
	for _, span := range n.OffsetMap {
		if span.NormalizedStart != normalizedCursor || span.NormalizedEnd <= span.NormalizedStart || span.SourceStart < sourceCursor || span.SourceEnd <= span.SourceStart {
			return false
		}
		normalizedCursor, sourceCursor = span.NormalizedEnd, span.SourceEnd
	}
	return normalizedCursor == len(n.Text)
}

// NormalizeOCR removes presentation-only Markdown markers commonly emitted by
// OCR and collapses whitespace while preserving a byte-level source mapping.
func NormalizeOCR(value string) NormalizedText {
	var out strings.Builder
	spans := make([]OffsetSpan, 0, len(value)/4)
	pendingSpaceStart, pendingSpaceEnd := -1, -1
	atLineStart := true
	activeMarker := map[string]bool{}

	appendMapped := func(text string, sourceStart, sourceEnd int) {
		start := out.Len()
		out.WriteString(text)
		span := OffsetSpan{NormalizedStart: start, NormalizedEnd: out.Len(), SourceStart: sourceStart, SourceEnd: sourceEnd}
		if len(spans) > 0 {
			last := &spans[len(spans)-1]
			lastNormalizedLength := last.NormalizedEnd - last.NormalizedStart
			lastSourceLength := last.SourceEnd - last.SourceStart
			spanNormalizedLength := span.NormalizedEnd - span.NormalizedStart
			spanSourceLength := span.SourceEnd - span.SourceStart
			if last.NormalizedEnd == span.NormalizedStart && last.SourceEnd == span.SourceStart && lastNormalizedLength == lastSourceLength && spanNormalizedLength == spanSourceLength {
				last.NormalizedEnd = span.NormalizedEnd
				last.SourceEnd = span.SourceEnd
				return
			}
		}
		spans = append(spans, span)
	}
	flushSpace := func() {
		if pendingSpaceStart >= 0 && out.Len() > 0 {
			appendMapped(" ", pendingSpaceStart, pendingSpaceEnd)
		}
		pendingSpaceStart, pendingSpaceEnd = -1, -1
	}

	for offset := 0; offset < len(value); {
		r, size := utf8.DecodeRuneInString(value[offset:])
		if r == utf8.RuneError && size == 1 {
			r, size = rune(value[offset]), 1
		}
		if r == '\r' || r == '\n' {
			if pendingSpaceStart < 0 {
				pendingSpaceStart = offset
			}
			pendingSpaceEnd = offset + size
			offset += size
			atLineStart = true
			continue
		}
		if unicode.IsSpace(r) {
			if pendingSpaceStart < 0 {
				pendingSpaceStart = offset
			}
			pendingSpaceEnd = offset + size
			offset += size
			continue
		}
		if atLineStart && r == '#' {
			end := offset
			for end < len(value) && end-offset < 6 && value[end] == '#' {
				end++
			}
			if end < len(value) {
				next, _ := utf8.DecodeRuneInString(value[end:])
				if unicode.IsSpace(next) {
					offset = end
					continue
				}
			}
		}
		atLineStart = false
		if offset+1 < len(value) {
			pair := value[offset : offset+2]
			if pair == "**" || pair == "__" || pair == "~~" {
				if activeMarker[pair] || strings.Contains(value[offset+2:], pair) {
					activeMarker[pair] = !activeMarker[pair]
					offset += 2
					continue
				}
			}
		}
		if r == '`' {
			if activeMarker["`"] || strings.Contains(value[offset+size:], "`") {
				activeMarker["`"] = !activeMarker["`"]
				offset += size
				continue
			}
		}
		flushSpace()
		if r == '\u2018' || r == '\u2019' || r == '\u02bc' {
			appendMapped("'", offset, offset+size)
		} else {
			appendMapped(value[offset:offset+size], offset, offset+size)
		}
		offset += size
	}
	return NormalizedText{Text: out.String(), OffsetMap: spans}
}

// SourceRange resolves a normalized half-open byte range to retained-source
// offsets. Callers must provide UTF-8/rune-aligned normalized boundaries.
func (n NormalizedText) SourceRange(start, end int) (int, int, bool) {
	if start < 0 || end <= start || end > len(n.Text) || len(n.OffsetMap) == 0 {
		return 0, 0, false
	}
	first := sort.Search(len(n.OffsetMap), func(index int) bool { return n.OffsetMap[index].NormalizedEnd > start })
	lastAfter := sort.Search(len(n.OffsetMap), func(index int) bool { return n.OffsetMap[index].NormalizedStart >= end })
	last := lastAfter - 1
	if first >= len(n.OffsetMap) || last < first {
		return 0, 0, false
	}
	left, right := n.OffsetMap[first], n.OffsetMap[last]
	sourceStart, sourceEnd := left.SourceStart, right.SourceEnd
	if left.NormalizedEnd-left.NormalizedStart == left.SourceEnd-left.SourceStart {
		sourceStart += start - left.NormalizedStart
	}
	if right.NormalizedEnd-right.NormalizedStart == right.SourceEnd-right.SourceStart {
		sourceEnd -= right.NormalizedEnd - end
	}
	if sourceStart < 0 || sourceEnd <= sourceStart {
		return 0, 0, false
	}
	return sourceStart, sourceEnd, true
}
