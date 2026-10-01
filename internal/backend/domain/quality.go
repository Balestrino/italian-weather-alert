package domain

import "time"

type QualityEvidence struct {
	DocumentVersionID int64
	URL, Locator      string
	ObservedAt        time.Time
}

type Dimension struct {
	State       string
	Limitations []string
	Evidence    []QualityEvidence
}

type UpdatingDimension struct {
	State                         string
	LastAttemptAt, LastCompleteAt *time.Time
	Limitations                   []string
}

type Quality struct {
	Provenance     Dimension
	Interpretation Dimension
	Updating       UpdatingDimension
}

type ProvenanceAssessment struct {
	SourceID, State, EvidenceURL, EvidenceLocator string
	Configuration                                 int
	Limitations                                   []string
	AssessedAt                                    time.Time
}

type InterpretationAssessment struct {
	MeasureID, State, Reason, Actor string
	EvidenceDocumentVersionID       int64
	Limitations                     []string
	RecordedAt                      time.Time
}

type AttentionDocument struct {
	DocumentVersionID           int64
	OfficialURL, Status, Reason string
	FirstAcquiredAt             time.Time
}

type MeasureView struct {
	Measure  LocalMeasure
	State    MeasureState
	Quality  Quality
	Warnings []AttentionDocument
}
