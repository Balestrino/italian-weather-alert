package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"
)

const VerificationLogic = "multi-source-v1"

var ErrVerificationEvidence = errors.New("invalid verification evidence")

// EvidenceSelection addresses literal text in a retained resource. JSONPointer
// selects a scalar before applying byte offsets; OCRRunID/Page select an already
// completed retained OCR page. Values are derived locally, never supplied as facts.
type EvidenceSelection struct {
	ResourceURL string `json:"resource_url"`
	JSONPointer string `json:"json_pointer,omitempty"`
	OCRRunID    int64  `json:"ocr_run_id,omitempty"`
	Page        int    `json:"page,omitempty"`
	StartByte   int    `json:"start_byte"`
	EndByte     int    `json:"end_byte"`
	Locator     string `json:"locator"`
}

type VerificationEvidence struct {
	SourceID  string                       `json:"source_id"`
	VersionID int64                        `json:"version_id"`
	Fields    map[string]EvidenceSelection `json:"fields"`
}

// Checks are results of bounded acquisition/discovery work, not votes. Missing
// and unavailable checks must name their source and retain a stable reason.
type VerificationCheck struct {
	Role      string    `json:"role"`
	State     string    `json:"state"` // available, missing, unavailable, not_applicable
	Reason    string    `json:"reason"`
	CheckedAt time.Time `json:"checked_at"`
	VerificationEvidence
}

type VerificationRequest struct {
	ComparisonOnly    bool                 `json:"comparison_only,omitempty"`
	RequestID         string               `json:"request_id"`
	CandidateKey      string               `json:"candidate_key"`
	Kind              string               `json:"kind"` // local_measure, operational_phase, regional_record
	MunicipalityISTAT string               `json:"municipality_istat"`
	Candidate         VerificationEvidence `json:"candidate"`
	Checks            []VerificationCheck  `json:"checks"`
}

type VerifiedField struct {
	Value   string `json:"value"`
	Passage string `json:"passage"`
	EvidenceSelection
}

type ChannelEvidence struct {
	Role          string                   `json:"role"`
	SourceID      string                   `json:"source_id,omitempty"`
	VersionID     int64                    `json:"version_id,omitempty"`
	Configuration int                      `json:"configuration,omitempty"`
	AuthorityID   string                   `json:"authority_id,omitempty"`
	PublisherID   string                   `json:"publisher_id,omitempty"`
	Platform      string                   `json:"platform,omitempty"`
	URL           string                   `json:"url,omitempty"`
	Hash          string                   `json:"sha256,omitempty"`
	Complete      bool                     `json:"complete"`
	Fields        map[string]VerifiedField `json:"fields"`
}

type FieldVerification struct {
	Field     string         `json:"field"`
	Outcome   string         `json:"outcome"`
	Reason    string         `json:"reason"`
	Candidate *VerifiedField `json:"candidate,omitempty"`
	Compared  *VerifiedField `json:"compared,omitempty"`
}

type ChannelVerification struct {
	ChannelEvidence
	CheckedAt               time.Time           `json:"checked_at"`
	Outcome                 string              `json:"outcome"`
	Reason                  string              `json:"reason"`
	ComparedFields          []FieldVerification `json:"compared_fields"`
	ComparisonFromSourceID  string              `json:"comparison_from_source_id,omitempty"`
	ComparisonFromVersionID int64               `json:"comparison_from_version_id,omitempty"`
	ComparisonToSourceID    string              `json:"comparison_to_source_id,omitempty"`
	ComparisonToVersionID   int64               `json:"comparison_to_version_id,omitempty"`
}

type VerificationReceipt struct {
	ID                int64                 `json:"id,omitempty"`
	RequestID         string                `json:"request_id"`
	CandidateKey      string                `json:"candidate_key"`
	Kind              string                `json:"kind"`
	MunicipalityISTAT string                `json:"municipality_istat"`
	LogicVersion      string                `json:"logic_version"`
	VerifiedAt        time.Time             `json:"verified_at"`
	Candidate         ChannelEvidence       `json:"candidate"`
	Checks            []ChannelVerification `json:"checks"`
	Outcome           string                `json:"outcome"`
	Admitted          bool                  `json:"admitted"`
	DomainRecordID    string                `json:"domain_record_id,omitempty"`
	PrimaryRecordID   string                `json:"primary_record_id,omitempty"`
	ProjectionState   string                `json:"projection_state"`
}

var verificationRoles = []string{"regional", "municipal", "platform"}
var localIdentity = []string{"reference", "edition"}
var regionalIdentity = []string{"product", "risk", "zone", "issuance", "validity"}
var verificationFields = []string{"reference", "edition", "kind", "subject", "place", "phase", "validity", "product", "risk", "zone", "issuance", "level"}

func primaryRole(kind string) string {
	if kind == "regional_record" {
		return "regional"
	}
	return "municipal"
}

func identityFields(kind string) []string {
	if kind == "regional_record" {
		return regionalIdentity
	}
	return localIdentity
}

func requiredFacts(kind string, fields map[string]VerifiedField) []string {
	switch kind {
	case "local_measure":
		return []string{"kind", "subject"}
	case "operational_phase":
		return []string{"phase"}
	default:
		if fields["product"].Value == ProductMonitoring {
			return []string{"product", "risk", "zone"}
		}
		return []string{"product", "risk", "zone", "level"}
	}
}

func compareEvidence(kind string, candidate, compared ChannelEvidence) (string, string, []FieldVerification) {
	identityOK := candidate.SourceID == compared.SourceID && candidate.VersionID > 0 && candidate.VersionID == compared.VersionID && candidate.Role == primaryRole(kind)
	selfPrimary := identityOK
	if !selfPrimary {
		identityOK = true
	}
	for _, name := range identityFields(kind) {
		if selfPrimary {
			break
		}
		a, b := candidate.Fields[name].Value, compared.Fields[name].Value
		if a == "" || b == "" || a != b {
			identityOK = false
		}
	}
	if kind == "regional_record" && !selfPrimary {
		identityOK = identityOK && verificationAbsoluteTime(candidate.Fields["issuance"].Passage) && verificationAbsoluteTime(candidate.Fields["validity"].Passage)
	}
	names := map[string]bool{}
	for name := range candidate.Fields {
		names[name] = true
	}
	for _, name := range identityFields(kind) {
		names[name] = true
	}
	for _, name := range requiredFacts(kind, candidate.Fields) {
		names[name] = true
	}
	ordered := make([]string, 0, len(names))
	for name := range names {
		ordered = append(ordered, name)
	}
	slices.Sort(ordered)
	results := make([]FieldVerification, 0, len(ordered))
	outcome, reason := "corroborated", "supported_fields_only"
	for _, name := range ordered {
		result := FieldVerification{Field: name, Outcome: "corroborated", Reason: "same_supported_value"}
		a, hasA := candidate.Fields[name]
		b, hasB := compared.Fields[name]
		if hasA {
			result.Candidate = &a
		}
		if hasB {
			result.Compared = &b
		}
		switch {
		case !identityOK:
			result.Outcome, result.Reason = "non_comparable", "identity_or_edition_or_validity_not_established"
		case !hasA || a.Value == "":
			result.Outcome, result.Reason = "non_comparable", "candidate_field_unknown"
		case !hasB || b.Value == "":
			result.Outcome, result.Reason = "missing_evidence", "compared_field_not_supported"
		case a.Value == "unknown" || b.Value == "unknown":
			result.Outcome, result.Reason = "non_comparable", "unknown_is_not_an_alert_level"
		case a.Value != b.Value:
			result.Outcome, result.Reason = "conflict", "comparable_values_disagree"
		}
		if result.Outcome == "conflict" {
			outcome, reason = "conflict", result.Reason
		}
		if outcome != "conflict" && slices.Contains(requiredFacts(kind, candidate.Fields), name) && result.Outcome != "corroborated" {
			outcome, reason = result.Outcome, result.Reason
		}
		if !identityOK {
			outcome, reason = "non_comparable", result.Reason
		}
		results = append(results, result)
	}
	return outcome, reason, results
}

func verificationAbsoluteTime(raw string) bool {
	if _, _, ok := verificationInterval(raw); ok {
		return true
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02", "02/01/2006"} {
		if _, err := time.Parse(layout, cleanLiteral(raw)); err == nil {
			return true
		}
	}
	return false
}

func admitsFields(kind string, result ChannelVerification, candidate ChannelEvidence) bool {
	if result.Outcome != "corroborated" {
		return false
	}
	for _, name := range requiredFacts(kind, candidate.Fields) {
		if !slices.ContainsFunc(result.ComparedFields, func(f FieldVerification) bool { return f.Field == name && f.Outcome == "corroborated" }) {
			return false
		}
	}
	return true
}

func verificationHash(v any) string {
	body, _ := json.Marshal(v)
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

func cleanLiteral(raw string) string { return strings.Join(strings.Fields(raw), " ") }

func normalizedField(name, raw string) string {
	value := strings.ToLower(cleanLiteral(raw))
	aliases := map[string]map[string]string{
		"kind":    {"chiusura": "closure", "riapertura": "reopening", "restrizione": "restriction", "divieto": "prohibition", "sospensione": "suspension", "attivazione": "activation", "disattivazione": "deactivation", "aggiornamento operativo": "operational_update", "osservazione": "observation"},
		"level":   {"verde": "green", "giallo": "yellow", "arancione": "orange", "rosso": "red", "sconosciuto": "unknown", "non applicabile": "not_applicable"},
		"risk":    {"vento": "wind", "temporali": "thunderstorms", "temporali forti": "thunderstorms", "mare": "coastal_waves", "mareggiate": "coastal_waves", "neve": "snow", "ghiaccio": "ice", "idrogeologico": "minor_network_hydro", "idraulico": "main_network_hydraulic", "rischio idrogeologico-idraulico del reticolo minore": "minor_network_hydro", "rischio idraulico del reticolo principale": "main_network_hydraulic"},
		"product": {"vigilanza": "vigilance", "criticità": "criticality", "monitoraggio": "monitoring"},
	}
	if canonical, ok := aliases[name][value]; ok {
		value = canonical
	}
	if name == "edition" || name == "issuance" || name == "validity" {
		if t, err := time.Parse(time.RFC3339, cleanLiteral(raw)); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
		for _, layout := range []string{"2006-01-02", "02/01/2006"} {
			if t, err := time.Parse(layout, cleanLiteral(raw)); err == nil {
				return t.Format("2006-01-02")
			}
		}
		if name == "validity" {
			if start, end, ok := verificationInterval(raw); ok {
				return start.Format(time.RFC3339) + "/" + end.Format(time.RFC3339)
			}
		}
	}
	return value
}

func verificationInterval(raw string) (time.Time, time.Time, bool) {
	a, b, ok := strings.Cut(cleanLiteral(raw), "/")
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	start, e1 := time.Parse(time.RFC3339, a)
	end, e2 := time.Parse(time.RFC3339, b)
	return start.UTC(), end.UTC(), e1 == nil && e2 == nil && end.After(start)
}
