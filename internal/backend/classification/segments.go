package classification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"unicode/utf8"
)

const (
	// MaxSegmentTextBytes keeps each Qwen request bounded independently of the
	// retained document size. Content collection still enforces its separate
	// six-MiB safety limit.
	MaxSegmentTextBytes = 12 << 10
	maxContentSegments  = 512
)

// Segment identifies an exact byte interval in one normalized retained
// resource/page. StartByte is inclusive and EndByte is exclusive.
type Segment struct {
	Ordinal, Total     int
	DocumentVersionID  int64
	ResourceURL, Role  string
	Page               int
	StartByte, EndByte int
	Text, Hash         string
	JSONScalars        []JSONScalar
}

func SegmentContent(documentVersionID int64, content Content) ([]Segment, error) {
	if documentVersionID < 1 || !content.Complete || len(content.Sections) == 0 {
		return nil, ErrInvalid
	}
	var segments []Segment
	for _, section := range content.Sections {
		for start := 0; start < len(section.Text); {
			for start < len(section.Text) && section.Text[start] == ' ' {
				start++
			}
			if start == len(section.Text) {
				break
			}
			end := segmentEnd(section.Text, start)
			segment := Segment{
				Ordinal: len(segments) + 1, DocumentVersionID: documentVersionID,
				ResourceURL: section.ResourceURL, Role: section.Role, Page: section.Page,
				StartByte: start, EndByte: end, Text: section.Text[start:end],
			}
			for _, scalar := range section.JSONScalars {
				if scalar.StartByte < end && scalar.EndByte > start {
					scalar.StartByte = max(scalar.StartByte, start) - start
					scalar.EndByte = min(scalar.EndByte, end) - start
					segment.JSONScalars = append(segment.JSONScalars, scalar)
				}
			}
			wire, err := json.Marshal(struct {
				DocumentVersionID                 int64
				Ordinal, Page, StartByte, EndByte int
				ResourceURL, Role, Text           string
				JSONScalars                       []JSONScalar `json:",omitempty"`
			}{segment.DocumentVersionID, segment.Ordinal, segment.Page, segment.StartByte, segment.EndByte, segment.ResourceURL, segment.Role, segment.Text, segment.JSONScalars})
			if err != nil || len(segments) >= maxContentSegments {
				return nil, ErrInvalid
			}
			hash := sha256.Sum256(wire)
			segment.Hash = hex.EncodeToString(hash[:])
			segments = append(segments, segment)
			start = end
		}
	}
	if len(segments) == 0 {
		return nil, ErrInvalid
	}
	for index := range segments {
		segments[index].Total = len(segments)
	}
	return segments, nil
}

func segmentEnd(text string, start int) int {
	if remaining := len(text) - start; remaining <= MaxSegmentTextBytes {
		return len(text)
	}
	end := start + MaxSegmentTextBytes
	for end > start && !utf8.RuneStart(text[end]) {
		end--
	}
	// Prefer a word boundary in the final quarter without allowing a tiny
	// segment merely because the normalized text contains an earlier space.
	floor := start + 3*MaxSegmentTextBytes/4
	for cursor := end; cursor > floor; cursor-- {
		if text[cursor-1] == ' ' {
			return cursor - 1
		}
	}
	return end
}

func (s Segment) Content() Content {
	section := ContentSection{ResourceURL: s.ResourceURL, Role: s.Role, Page: s.Page, Text: s.Text, JSONScalars: s.JSONScalars}
	return Content{Sections: []ContentSection{section}, Complete: true, Hash: s.Hash, Text: s.Text}
}
