package publicquery

import (
	"context"
	"slices"
	"time"
)

// A query owns this cache. Store itself remains safe to share between requests;
// normalized nested operations share reads only at the same time boundary.
type queryReads struct {
	time            QueryTime
	quality         map[sourceReadKey]Quality
	interpretations map[int64]documentAssessment
	documents       map[documentReadKey]Document
	attachments     map[int64][]Attachment
	versionReceipts map[int64]bool
	evidence        map[int64]Evidence
	evidenceCutoffs map[evidenceReadKey]*time.Time
}
type evidenceReadKey struct {
	version              int64
	source, municipality string
}
type documentReadKey struct {
	version      int64
	municipality string
}
type sourceReadKey struct{ source, municipality string }
type documentAssessment struct {
	dimension     Dimension
	interpretedAt *time.Time
	runID         *string
}

func (s *Store) withQueryReads(qt QueryTime) *Store {
	copy := *s
	if copy.reads == nil || copy.reads.time != qt {
		copy.reads = &queryReads{
			time: qt, quality: map[sourceReadKey]Quality{}, interpretations: map[int64]documentAssessment{},
			documents: map[documentReadKey]Document{}, attachments: map[int64][]Attachment{}, versionReceipts: map[int64]bool{},
			evidence: map[int64]Evidence{}, evidenceCutoffs: map[evidenceReadKey]*time.Time{},
		}
	}
	return &copy
}
func cloneDimension(value Dimension) Dimension {
	value.Limitations = slices.Clone(value.Limitations)
	value.Evidence = slices.Clone(value.Evidence)
	return value
}
func (s *Store) quality(ctx context.Context, sourceID string, qt QueryTime, interpretation Dimension) (Quality, error) {
	key := sourceReadKey{sourceID, s.municipalityScope}
	var base Quality
	var found bool
	if s.reads != nil {
		base, found = s.reads.quality[key]
	}
	if !found {
		var err error
		base, err = s.loadQuality(ctx, sourceID, qt, Dimension{Limitations: []string{}, Evidence: []Evidence{}})
		if err != nil {
			return Quality{}, err
		}
		if s.reads != nil {
			s.reads.quality[key] = base
		}
	}
	result := base
	result.Provenance = cloneDimension(base.Provenance)
	result.Updating.Limitations = slices.Clone(base.Updating.Limitations)
	result.Interpretation = cloneDimension(interpretation)
	result.Interpretation.Limitations = append(result.Interpretation.Limitations, base.Interpretation.Limitations...)
	if base.Interpretation.State == "unreliable" {
		result.Interpretation.State = "unreliable"
	}
	result.Interpretation.Limitations = nonNil(result.Interpretation.Limitations)
	result.Interpretation.Evidence = nonNilEvidence(result.Interpretation.Evidence)
	return result, nil
}
func (s *Store) documentInterpretation(ctx context.Context, versionID int64, knownAt time.Time) (Dimension, *time.Time, *string, error) {
	if s.reads != nil {
		if value, ok := s.reads.interpretations[versionID]; ok {
			return cloneDimension(value.dimension), value.interpretedAt, value.runID, nil
		}
	}
	dimension, at, run, err := s.loadDocumentInterpretation(ctx, versionID, knownAt)
	if err == nil && s.reads != nil {
		s.reads.interpretations[versionID] = documentAssessment{cloneDimension(dimension), at, run}
	}
	return dimension, at, run, err
}
