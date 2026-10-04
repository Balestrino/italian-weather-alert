// Package observation persists and evaluates the paced source-observation
// campaign required before Toscana/Calcinaia source acceptance.
package observation

import (
	"errors"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

var (
	ErrInvalid  = errors.New("invalid observation campaign input")
	ErrNotFound = errors.New("observation campaign not found")
	ErrConflict = errors.New("observation campaign conflict")
)

const MinimumDuration = 7 * 24 * time.Hour

type StartRequest struct {
	ID        string    `json:"id"`
	Actor     string    `json:"actor"`
	Scope     string    `json:"scope,omitempty"`
	StartedAt time.Time `json:"started_at"`
	SourceIDs []string  `json:"source_ids"`
}

type Source struct {
	SourceID      string   `json:"source_id"`
	Configuration int      `json:"configuration"`
	Product       string   `json:"product"`
	Territory     string   `json:"territory"`
	Sections      []string `json:"sections"`
	CheckSeconds  int      `json:"check_seconds"`
	DelaySeconds  int      `json:"delay_seconds"`
}

type Campaign struct {
	ID           string      `json:"id"`
	Actor        string      `json:"actor"`
	Scope        string      `json:"scope"`
	StartedAt    time.Time   `json:"started_at"`
	MinimumEndAt time.Time   `json:"minimum_end_at"`
	CreatedAt    time.Time   `json:"created_at"`
	Sources      []Source    `json:"sources"`
	Latest       *Assessment `json:"latest_assessment,omitempty"`
}

// Review is attributable evidence, including explicitly delegated document
// review, associated with retained service evidence. Actual
// observations must identify a persisted version or failed check. Retained
// cases are allowed only for an event/failure absent during the live period.
type Review struct {
	ID           string            `json:"id"`
	SourceID     string            `json:"source_id"`
	Kind         string            `json:"kind"`
	Status       string            `json:"status"`
	CheckID      *int64            `json:"check_id,omitempty"`
	VersionID    *int64            `json:"version_id,omitempty"`
	CaseID       string            `json:"case_id,omitempty"`
	Evidence     registry.Evidence `json:"evidence"`
	Notes        string            `json:"notes,omitempty"`
	Actor        string            `json:"actor,omitempty"`
	RecordedAt   time.Time         `json:"recorded_at,omitempty"`
	Supersedes   string            `json:"supersedes,omitempty"`
	Correction   string            `json:"correction,omitempty"`
	SupersededBy string            `json:"superseded_by,omitempty"`
}

type SourceReport struct {
	Source
	Checks                   int        `json:"checks"`
	CompleteChecks           int        `json:"complete_checks"`
	FailedChecks             int        `json:"failed_checks"`
	ObservedLocalDays        int        `json:"observed_local_days"`
	FirstCheckAt             *time.Time `json:"first_check_at,omitempty"`
	LastCheckAt              *time.Time `json:"last_check_at,omitempty"`
	MaximumGapSeconds        int64      `json:"maximum_gap_seconds"`
	VersionsAcquired         int        `json:"versions_acquired"`
	RequiredResources        int        `json:"required_resources"`
	MissingRequiredResources int        `json:"missing_required_resources"`
	OriginalComparisons      int        `json:"original_comparisons"`
	AttachmentComparisons    int        `json:"attachment_comparisons"`
	FailureExercises         int        `json:"failure_exercises"`
	RetainedEventCases       int        `json:"retained_event_cases"`
	Reviews                  []Review   `json:"reviews"`
	Issues                   []string   `json:"issues"`
}

type Report struct {
	CampaignID      string         `json:"campaign_id"`
	Scope           string         `json:"scope"`
	StartedAt       time.Time      `json:"started_at"`
	Through         time.Time      `json:"through"`
	MinimumEndAt    time.Time      `json:"minimum_end_at"`
	Status          string         `json:"status"`
	MissingProducts []string       `json:"missing_products,omitempty"`
	Sources         []SourceReport `json:"sources"`
	Issues          []string       `json:"issues"`
}

type Assessment struct {
	ID         int64             `json:"id"`
	CampaignID string            `json:"campaign_id"`
	Actor      string            `json:"actor"`
	Through    time.Time         `json:"through"`
	Status     string            `json:"status"`
	Evidence   registry.Evidence `json:"evidence"`
	Report     Report            `json:"report"`
	RecordedAt time.Time         `json:"recorded_at"`
}
