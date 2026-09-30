// Package diagnostics derives internal source-quality snapshots and retains
// scoped DPC comparisons. It never writes public warning/domain records.
package diagnostics

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid diagnostic record")
	ErrNotFound = errors.New("diagnostic record not found")
	ErrConflict = errors.New("diagnostic record conflict")
)

type Signal struct {
	State       string          `json:"state"`
	Evidence    json.RawMessage `json:"evidence"`
	Limitations []string        `json:"limitations"`
}

type Snapshot struct {
	SourceID       string    `json:"source_id"`
	ObservedAt     time.Time `json:"observed_at"`
	Availability   Signal    `json:"availability"`
	Timeliness     Signal    `json:"timeliness"`
	Interpretation Signal    `json:"interpretation"`
}

type Evidence struct {
	ResourceURL string `json:"resource_url"`
	Page        *int   `json:"page"`
	Locator     string `json:"locator"`
}

type Dimension struct {
	Name          string   `json:"name"`
	State         string   `json:"state"`
	RegionalValue *string  `json:"regional_value"`
	DPCValue      *string  `json:"dpc_value"`
	Reason        string   `json:"reason"`
	Regional      Evidence `json:"regional_evidence"`
	DPC           Evidence `json:"dpc_evidence"`
}

type Comparison struct {
	ID                string      `json:"id"`
	RegionalSourceID  string      `json:"regional_source_id"`
	DPCSourceID       string      `json:"dpc_source_id"`
	RegionalVersionID int64       `json:"regional_version_id"`
	DPCVersionID      int64       `json:"dpc_version_id"`
	Scope             string      `json:"scope"`
	Actor             string      `json:"actor"`
	Dimensions        []Dimension `json:"dimensions"`
	ComparedAt        time.Time   `json:"compared_at"`
}
