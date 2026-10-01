package extraction

import (
	"slices"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
)

func RemapResult(old Result, windows []Window, content classification.Content, base Result) (Result, bool) {
	if old.Status != "extracted" || !old.ContentComplete || !content.Complete || len(windows) == 0 || len(old.Segments) != len(windows) {
		return Result{}, false
	}
	result := old
	result.RunID = base.RunID
	result.DocumentVersionID = base.DocumentVersionID
	result.ClassificationRunID = base.ClassificationRunID
	result.ContentSHA256 = content.Hash
	result.CreatedAt = base.CreatedAt
	result.InputTokens = nil
	result.OutputTokens = nil
	result.CacheReadTokens = nil
	result.Segments = append([]SegmentResult(nil), old.Segments...)
	for index, window := range windows {
		prior := &result.Segments[index]
		if prior.Ordinal != window.Ordinal || prior.Total != window.Total || prior.ResourceURL != window.ResourceURL || prior.Role != window.Role || prior.Page != window.Page || prior.StartByte != window.ContextStartByte || prior.EndByte != window.ContextEndByte || prior.CoreStartByte != window.CoreStartByte || prior.CoreEndByte != window.CoreEndByte || prior.SourceStartByte != window.SourceStartByte || prior.SourceEndByte != window.SourceEndByte || !slices.Equal(prior.NormalizationMap, window.OffsetMap) {
			return Result{}, false
		}
		prior.DocumentVersionID = base.DocumentVersionID
		prior.ContentSHA256 = window.Hash
		prior.NormalizationMap = window.OffsetMap
		prior.InputTokens = nil
		prior.OutputTokens = nil
		prior.CacheReadTokens = nil
	}
	for _, measure := range old.Measures {
		for _, evidence := range measure.Evidence {
			if evidence.SegmentOrdinal < 1 || evidence.SegmentOrdinal > len(windows) {
				return Result{}, false
			}
			window := windows[evidence.SegmentOrdinal-1]
			if !literalReference(window.Content(), evidence.ResourceURL, evidence.Page, evidence.Quote) {
				return Result{}, false
			}
		}
	}
	measures, err := Merge([][]Measure{old.Measures}, content, len(windows))
	if err != nil || len(measures) != len(old.Measures) {
		return Result{}, false
	}
	result.Measures = measures
	return result, true
}

// Keep the mapping's acquisition/publication expressions verbatim; only the
// time at which this new evidence association was prepared is new.
