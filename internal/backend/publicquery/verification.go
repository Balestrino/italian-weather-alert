package publicquery

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
)

// Verification is the public view of a receipt. Request identities and private
// selection/OCR coordinates belong to the administrative evidence store.
type Verification struct {
	ID                string              `json:"id"`
	Kind              string              `json:"kind"`
	MunicipalityISTAT string              `json:"municipality_istat"`
	LogicVersion      string              `json:"logic_version"`
	VerifiedAt        time.Time           `json:"verified_at"`
	CandidateRole     string              `json:"candidate_role"`
	Outcome           string              `json:"outcome"`
	Admitted          bool                `json:"admitted"`
	ProjectionState   string              `json:"projection_state"`
	Checks            []VerificationCheck `json:"checks"`
	Limitations       []string            `json:"limitations"`
}

type VerificationCheck struct {
	Role            string              `json:"role"`
	SourceID        string              `json:"source_id,omitempty"`
	VersionID       string              `json:"version_id,omitempty"`
	Configuration   int                 `json:"configuration,omitempty"`
	AuthorityID     string              `json:"authority_id,omitempty"`
	PublisherID     string              `json:"publisher_id,omitempty"`
	Platform        string              `json:"platform,omitempty"`
	SourceURL       string              `json:"source_url,omitempty"`
	SHA256          string              `json:"sha256,omitempty"`
	Complete        bool                `json:"complete"`
	EvidenceVisible bool                `json:"evidence_visible"`
	CheckedAt       time.Time           `json:"checked_at"`
	Outcome         string              `json:"outcome"`
	Reason          string              `json:"reason"`
	FromVersionID   string              `json:"from_version_id,omitempty"`
	ToVersionID     string              `json:"to_version_id,omitempty"`
	Fields          []VerificationField `json:"fields"`
}

type VerificationField struct {
	Field          string     `json:"field"`
	Outcome        string     `json:"outcome"`
	Reason         string     `json:"reason"`
	CandidateValue *string    `json:"candidate_value,omitempty"`
	ComparedValue  *string    `json:"compared_value,omitempty"`
	Evidence       []Evidence `json:"evidence"`
}

func (s *Store) verifications(ctx context.Context, kind, recordID string, versionID int64, qt QueryTime) ([]Verification, error) {
	municipalVisibility := "true"
	if s.development && s.region == "" {
		municipalVisibility = "EXISTS(SELECT 1 FROM development_publication_current_municipalities dm WHERE dm.istat=r.municipality_istat)"
	}
	// The candidate itself must be publishable. A shared primary bulletin must
	// never reveal private candidates for another municipality through its receipts.
	rows, err := s.pool.Query(ctx, `SELECT r.id,r.body FROM domain_verification_receipts r
 JOIN retained_versions v ON v.id=r.candidate_version_id
 JOIN `+s.sourcesSQL()+` s ON s.id=r.candidate_source_id
 WHERE r.verified_at<=registry_interpretation_cutoff(s.id,$4)
 AND v.first_acquired_at<=$4 AND (`+s.visibilitySQL()+`)
 AND ($1='' OR r.kind=$1)
 AND (($2<>'' AND (r.body->>'primary_record_id'=$2 OR r.body->>'domain_record_id'=$2))
 OR ($3::bigint>0 AND (r.candidate_version_id=$3 OR EXISTS(
 SELECT 1 FROM domain_verification_dependencies d WHERE d.receipt_id=r.id AND d.evidence_version_id=$3))))
 AND ($5='' OR r.municipality_istat=$5)
 AND (`+municipalVisibility+`)
 ORDER BY r.verified_at DESC,r.id DESC LIMIT 101`, kind, recordID, versionID, qt.KnownAt, s.municipalityScope)
	if err != nil {
		return nil, err
	}
	var receipts []domain.VerificationReceipt
	for rows.Next() {
		var id int64
		var body []byte
		if err = rows.Scan(&id, &body); err != nil {
			rows.Close()
			return nil, err
		}
		var receipt domain.VerificationReceipt
		if err = json.Unmarshal(body, &receipt); err != nil {
			rows.Close()
			return nil, err
		}
		receipt.ID = id
		receipts = append(receipts, receipt)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	truncated := len(receipts) > 100
	if truncated {
		receipts = receipts[:100]
	}
	result := make([]Verification, 0, len(receipts))
	for _, receipt := range receipts {
		value, loadErr := s.publicVerification(ctx, receipt, qt)
		if loadErr != nil {
			return nil, loadErr
		}
		if truncated {
			value.Limitations = append(value.Limitations, "Only the latest 100 publishable receipts at this knowledge boundary are included.")
		}
		result = append(result, value)
	}
	return result, nil
}

func (s *Store) publicVerification(ctx context.Context, receipt domain.VerificationReceipt, qt QueryTime) (Verification, error) {
	value := Verification{ID: stringID(receipt.ID), Kind: receipt.Kind, MunicipalityISTAT: receipt.MunicipalityISTAT, LogicVersion: receipt.LogicVersion, VerifiedAt: receipt.VerifiedAt, CandidateRole: receipt.Candidate.Role, Outcome: receipt.Outcome, Admitted: receipt.Admitted, ProjectionState: receipt.ProjectionState, Checks: []VerificationCheck{}, Limitations: []string{"Verification compares retained field evidence; it does not establish semantic source acceptance or exhaustive coverage. Admission refers to the candidate; supported primary facts may remain available independently."}}
	visible := map[int64]bool{}
	for _, channel := range receipt.Checks {
		if channel.VersionID == 0 {
			continue
		}
		var allowed bool
		err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 JOIN `+s.sourcesSQL()+` s ON s.id=d.source_id WHERE v.id=$1 AND s.id=$2 AND v.first_acquired_at<=$3
 AND $4::timestamptz<=registry_interpretation_cutoff(s.id,$3) AND (`+s.visibilitySQL()+`))`, channel.VersionID, channel.SourceID, qt.KnownAt, receipt.VerifiedAt).Scan(&allowed)
		if err != nil {
			return value, err
		}
		visible[channel.VersionID] = allowed
	}
	for _, channel := range receipt.Checks {
		check := VerificationCheck{Role: channel.Role, CheckedAt: channel.CheckedAt, Outcome: channel.Outcome, Reason: channel.Outcome, Fields: []VerificationField{}}
		if visible[channel.VersionID] {
			check.SourceID, check.VersionID, check.Configuration = channel.SourceID, stringID(channel.VersionID), channel.Configuration
			check.AuthorityID, check.PublisherID, check.Platform = channel.AuthorityID, channel.PublisherID, channel.Platform
			check.SourceURL, check.SHA256, check.Complete, check.EvidenceVisible = channel.URL, channel.Hash, channel.Complete, true
			check.Reason = channel.Reason
			if visible[channel.ComparisonFromVersionID] {
				check.FromVersionID = stringID(channel.ComparisonFromVersionID)
			}
			if visible[channel.ComparisonToVersionID] {
				check.ToVersionID = stringID(channel.ComparisonToVersionID)
			}
			for _, field := range channel.ComparedFields {
				out := VerificationField{Field: field.Field, Outcome: field.Outcome, Reason: field.Reason, Evidence: []Evidence{}}
				for _, side := range []struct {
					version int64
					field   *domain.VerifiedField
					value   **string
				}{{channel.ComparisonFromVersionID, field.Candidate, &out.CandidateValue}, {channel.ComparisonToVersionID, field.Compared, &out.ComparedValue}} {
					if side.field == nil || !visible[side.version] {
						continue
					}
					*side.value = &side.field.Value
					evidence, err := s.versionEvidence(ctx, side.version, side.field.Locator)
					if err != nil {
						return value, err
					}
					evidence.SourceURL, evidence.Passage = side.field.ResourceURL, &side.field.Passage
					if side.field.Page > 0 {
						evidence.Page = &side.field.Page
					}
					out.Evidence = append(out.Evidence, evidence)
				}
				check.Fields = append(check.Fields, out)
			}
		} else if channel.VersionID > 0 {
			check.Reason = "channel_evidence_not_public"
			value.Limitations = append(value.Limitations, "Evidence for the "+channel.Role+" channel is not publicly enabled in this scope.")
		}
		value.Checks = append(value.Checks, check)
	}
	return value, nil
}
