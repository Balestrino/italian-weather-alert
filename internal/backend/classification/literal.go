package classification

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// CanonicalLiteral accepts only a contiguous match under the already reviewed
// OCR presentation normalization. It returns the exact source substring, not a
// repaired model assertion. Byte mapping also survives Unicode case-width changes.
func CanonicalLiteral(source, quote string) (string, bool) {
	return canonicalLiteral(source, quote, false)
}

const LocalLiteralVersion = "local-contiguous-double-quotes-v1"

// CanonicalLocalLiteral also accepts typographic double-quote presentation,
// while returning the exact contiguous source substring with original offsets.
func CanonicalLocalLiteral(source, quote string) (string, bool) {
	return canonicalLiteral(source, quote, true)
}

func canonicalLiteral(source, quote string, doubleQuotes bool) (string, bool) {
	view := NormalizeOCR(source)
	needle := strings.ToLower(NormalizeOCR(quote).Text)
	if doubleQuotes {
		needle = strings.NewReplacer("“", "\"", "”", "\"").Replace(needle)
	}
	if needle == "" {
		return "", false
	}
	var folded strings.Builder
	var starts, ends []int
	for offset, r := range view.Text {
		lower := string(unicode.ToLower(r))
		if doubleQuotes && (r == '“' || r == '”') {
			lower = "\""
		}
		folded.WriteString(lower)
		for range []byte(lower) {
			starts = append(starts, offset)
			ends = append(ends, offset+utf8.RuneLen(r))
		}
	}
	index := strings.Index(folded.String(), needle)
	if index < 0 {
		return "", false
	}
	start, end, ok := view.SourceRange(starts[index], ends[index+len(needle)-1])
	if !ok || end > len(source) {
		return "", false
	}
	return source[start:end], true
}
