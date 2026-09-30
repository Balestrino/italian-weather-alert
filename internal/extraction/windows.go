package extraction

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/Balestrino/italian-weather-alert/internal/classification"
)

const (
	MaxWindowCoreBytes    = 4 << 10
	MaxWindowContextBytes = 1 << 10
	maxOperationalWindows = 1536
)

// WindowOffset resolves a normalized window interval to the retained OCR page.
type WindowOffset struct {
	NormalizedStart int `json:"normalized_start"`
	NormalizedEnd   int `json:"normalized_end"`
	SourceStart     int `json:"source_start"`
	SourceEnd       int `json:"source_end"`
}

// Window has a non-overlapping core and bounded context in one retained
// resource/page. Only operative predicates beginning in the core belong to it.
type Window struct {
	legacyLiteral     bool
	Ordinal           int            `json:"ordinal"`
	Total             int            `json:"total"`
	DocumentVersionID int64          `json:"document_version_id"`
	ResourceURL       string         `json:"resource_url"`
	Role              string         `json:"role"`
	Page              int            `json:"page,omitempty"`
	ContextStartByte  int            `json:"context_start_byte"`
	ContextEndByte    int            `json:"context_end_byte"`
	CoreStartByte     int            `json:"core_start_byte"`
	CoreEndByte       int            `json:"core_end_byte"`
	SourceStartByte   int            `json:"source_start_byte,omitempty"`
	SourceEndByte     int            `json:"source_end_byte,omitempty"`
	Text              string         `json:"text"`
	Hash              string         `json:"sha256"`
	OffsetMap         []WindowOffset `json:"normalization_map,omitempty"`
}

func BuildOperationalWindows(documentVersionID int64, content classification.Content) ([]Window, error) {
	return buildOperationalWindows(documentVersionID, content, MaxWindowCoreBytes, MaxWindowContextBytes)
}

func buildOperationalWindows(documentVersionID int64, content classification.Content, coreLimit, contextLimit int) ([]Window, error) {
	if documentVersionID < 1 || !content.Complete || len(content.Sections) == 0 || coreLimit < 32 || contextLimit < 0 || coreLimit+2*contextLimit > classification.MaxSegmentTextBytes {
		return nil, ErrInvalid
	}
	var windows []Window
	for _, section := range content.Sections {
		if section.ResourceURL == "" || section.Text == "" || !utf8.ValidString(section.Text) {
			return nil, ErrInvalid
		}
		if section.Page > 0 && !(classification.NormalizedText{Text: section.Text, OffsetMap: section.OffsetMap}).Valid() {
			return nil, ErrInvalid
		}
		for coreStart := 0; coreStart < len(section.Text); {
			for coreStart < len(section.Text) && section.Text[coreStart] == ' ' {
				coreStart++
			}
			if coreStart == len(section.Text) {
				break
			}
			coreEnd := stableCoreEnd(section.Text, coreStart, coreLimit)
			contextStart := stableContextStart(section.Text, coreStart, contextLimit)
			contextEnd := stableContextEnd(section.Text, coreEnd, contextLimit)
			window := Window{
				Ordinal: len(windows) + 1, DocumentVersionID: documentVersionID,
				ResourceURL: section.ResourceURL, Role: section.Role, Page: section.Page,
				ContextStartByte: contextStart, ContextEndByte: contextEnd,
				CoreStartByte: coreStart, CoreEndByte: coreEnd,
				Text: section.Text[contextStart:contextEnd],
			}
			window.OffsetMap = windowOffsets(section.Text, section.OffsetMap, contextStart, contextEnd)
			if section.Page > 0 {
				view := classification.NormalizedText{Text: section.Text, OffsetMap: section.OffsetMap}
				var ok bool
				window.SourceStartByte, window.SourceEndByte, ok = view.SourceRange(contextStart, contextEnd)
				if !ok {
					return nil, ErrInvalid
				}
			}
			if !validWindowOffsetMap(window.Page, window.ContextStartByte, window.ContextEndByte, window.SourceStartByte, window.SourceEndByte, window.OffsetMap) {
				return nil, ErrInvalid
			}
			wire, err := json.Marshal(struct {
				DocumentVersionID                                            int64
				Ordinal, Page                                                int
				ResourceURL, Role                                            string
				ContextStartByte, ContextEndByte, CoreStartByte, CoreEndByte int
				SourceStartByte, SourceEndByte                               int
				Text                                                         string
			}{window.DocumentVersionID, window.Ordinal, window.Page, window.ResourceURL, window.Role, window.ContextStartByte, window.ContextEndByte, window.CoreStartByte, window.CoreEndByte, window.SourceStartByte, window.SourceEndByte, window.Text})
			if err != nil || len(windows) >= maxOperationalWindows || len(window.Text) > classification.MaxSegmentTextBytes {
				return nil, ErrInvalid
			}
			digest := sha256.Sum256(wire)
			window.Hash = hex.EncodeToString(digest[:])
			windows = append(windows, window)
			coreStart = coreEnd
		}
	}
	if len(windows) == 0 {
		return nil, ErrInvalid
	}
	for index := range windows {
		windows[index].Total = len(windows)
	}
	return windows, nil
}

func stableCoreEnd(text string, start, limit int) int {
	if len(text)-start <= limit {
		return len(text)
	}
	maximum := runeBoundaryAtOrBefore(text, start+limit)
	floor := start + limit/2
	for cursor := maximum; cursor > floor; cursor-- {
		if isSentenceBoundary(text, cursor) {
			return cursor
		}
	}
	for cursor := maximum; cursor > floor; cursor-- {
		if text[cursor-1] == ' ' {
			return cursor
		}
	}
	return maximum
}

func stableContextStart(text string, coreStart, allowance int) int {
	if coreStart == 0 || allowance == 0 {
		return coreStart
	}
	minimum := runeBoundaryAtOrAfter(text, max(0, coreStart-allowance))
	for cursor := minimum; cursor < coreStart; cursor++ {
		if cursor == 0 || isSentenceBoundary(text, cursor) {
			return cursor
		}
	}
	for cursor := minimum; cursor < coreStart; cursor++ {
		if cursor == 0 || text[cursor-1] == ' ' {
			return cursor
		}
	}
	return minimum
}

func stableContextEnd(text string, coreEnd, allowance int) int {
	if coreEnd == len(text) || allowance == 0 {
		return coreEnd
	}
	maximum := runeBoundaryAtOrBefore(text, min(len(text), coreEnd+allowance))
	for cursor := coreEnd + 1; cursor <= maximum; cursor++ {
		if isSentenceBoundary(text, cursor) {
			return cursor
		}
	}
	for cursor := maximum; cursor > coreEnd; cursor-- {
		if text[cursor-1] == ' ' {
			return cursor - 1
		}
	}
	return maximum
}

func isSentenceBoundary(text string, end int) bool {
	if end <= 0 || end > len(text) || end < len(text) && text[end] != ' ' {
		return false
	}
	switch text[end-1] {
	case '.', '!', '?', ';', ':':
		return true
	default:
		return false
	}
}

func runeBoundaryAtOrBefore(text string, offset int) int {
	if offset >= len(text) {
		return len(text)
	}
	for offset > 0 && !utf8.RuneStart(text[offset]) {
		offset--
	}
	return offset
}

func runeBoundaryAtOrAfter(text string, offset int) int {
	for offset < len(text) && !utf8.RuneStart(text[offset]) {
		offset++
	}
	return offset
}

func windowOffsets(text string, offsets []classification.OffsetSpan, start, end int) []WindowOffset {
	var result []WindowOffset
	view := classification.NormalizedText{Text: text, OffsetMap: offsets}
	for _, span := range offsets {
		if span.NormalizedEnd <= start || span.NormalizedStart >= end {
			continue
		}
		normalizedStart, normalizedEnd := max(start, span.NormalizedStart), min(end, span.NormalizedEnd)
		sourceStart, sourceEnd, ok := view.SourceRange(normalizedStart, normalizedEnd)
		if !ok {
			return nil
		}
		result = append(result, WindowOffset{normalizedStart, normalizedEnd, sourceStart, sourceEnd})
	}
	return result
}

func validWindowOffsetMap(page, normalizedStart, normalizedEnd, sourceStart, sourceEnd int, offsets []WindowOffset) bool {
	if page == 0 {
		return sourceStart == 0 && sourceEnd == 0 && len(offsets) == 0
	}
	if normalizedStart < 0 || normalizedEnd <= normalizedStart || sourceStart < 0 || sourceEnd <= sourceStart || len(offsets) == 0 || offsets[0].NormalizedStart != normalizedStart || offsets[0].SourceStart != sourceStart || offsets[len(offsets)-1].NormalizedEnd != normalizedEnd || offsets[len(offsets)-1].SourceEnd != sourceEnd {
		return false
	}
	normalizedCursor, sourceCursor := normalizedStart, sourceStart
	for _, span := range offsets {
		if span.NormalizedStart != normalizedCursor || span.NormalizedEnd <= span.NormalizedStart || span.SourceStart < sourceCursor || span.SourceEnd <= span.SourceStart {
			return false
		}
		normalizedCursor, sourceCursor = span.NormalizedEnd, span.SourceEnd
	}
	return normalizedCursor == normalizedEnd && sourceCursor == sourceEnd
}

func (w Window) Content() classification.Content {
	section := classification.ContentSection{ResourceURL: w.ResourceURL, Role: w.Role, Page: w.Page, Text: w.Text}
	return classification.Content{LegacyLiteral: w.legacyLiteral, Sections: []classification.ContentSection{section}, Complete: true, Hash: w.Hash, Text: w.Text}
}

func (w Window) ownsOperativeQuote(quote string) bool {
	quote = classification.Normalize(quote)
	if quote == "" {
		return false
	}
	text, needle := w.Text, quote
	for local := 0; local <= len(text)-len(needle); {
		index := strings.Index(text[local:], needle)
		if index < 0 {
			return false
		}
		absolute := w.ContextStartByte + local + index
		if absolute >= w.CoreStartByte && absolute < w.CoreEndByte {
			return true
		}
		local += index + 1
	}
	return false
}
