package domain

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/evidenceguard"
	"github.com/jackc/pgx/v5"
)

var verificationISTAT = regexp.MustCompile(`^[0-9]{6}$`)

// VerifyMultiSource executes deterministic comparisons on interpreted, retained
// evidence. There is no approval step and no majority-vote resolution. Primary
// projection and its immutable receipt are committed in one transaction.
func (s *Store) VerifyMultiSource(ctx context.Context, docs RegionalDocuments, request VerificationRequest, at time.Time) (VerificationReceipt, error) {
	out := VerificationReceipt{RequestID: request.RequestID, CandidateKey: request.CandidateKey, Kind: request.Kind, MunicipalityISTAT: request.MunicipalityISTAT, LogicVersion: VerificationLogic, VerifiedAt: at.UTC(), ProjectionState: "diagnostic"}
	if s == nil || s.pool == nil || docs == nil || at.IsZero() || strings.TrimSpace(request.RequestID) == "" || len(request.RequestID) > 200 || strings.TrimSpace(request.CandidateKey) == "" || len(request.CandidateKey) > 200 ||
		!verificationISTAT.MatchString(request.MunicipalityISTAT) || !slices.Contains([]string{"local_measure", "operational_phase", "regional_record"}, request.Kind) || len(request.Checks) != 2 {
		return out, ErrInvalid
	}
	hash := verificationHash(struct {
		Logic   string
		Request VerificationRequest
	}{VerificationLogic, request})
	if old, err := s.existingVerification(ctx, s.pool, request.RequestID, hash); err != pgx.ErrNoRows {
		return old, err
	}
	unlock, err := evidenceguard.Lock(ctx, s.pool, true)
	if err != nil {
		return out, err
	}
	defer unlock()
	candidate, err := s.loadVerificationEvidence(ctx, docs, request.Candidate, request.MunicipalityISTAT, at)
	if err != nil {
		return out, err
	}
	if request.Kind != "regional_record" && candidate.Role == "regional" {
		return out, ErrVerificationEvidence
	}
	out.Candidate = candidate
	checks := map[string]ChannelVerification{}
	checks[candidate.Role] = ChannelVerification{ChannelEvidence: candidate, CheckedAt: at.UTC(), Outcome: "corroborated", Reason: "candidate_channel_evidence"}
	for _, input := range request.Checks {
		if !slices.Contains(verificationRoles, input.Role) || input.Role == candidate.Role || !slices.Contains([]string{"available", "missing", "unavailable", "not_applicable"}, input.State) ||
			input.CheckedAt.IsZero() || input.CheckedAt.After(at) || len(input.Reason) > 200 || strings.TrimSpace(input.Reason) == "" {
			return out, ErrInvalid
		}
		if _, exists := checks[input.Role]; exists {
			return out, ErrInvalid
		}
		if input.State == "not_applicable" && (input.Role == primaryRole(request.Kind) || input.Role == "platform") {
			return out, ErrInvalid
		}
		if input.State != "available" && (input.VersionID != 0 || len(input.Fields) != 0) || input.State != "not_applicable" && input.SourceID == "" {
			return out, ErrInvalid
		}
		check := ChannelVerification{ChannelEvidence: ChannelEvidence{Role: input.Role, Fields: map[string]VerifiedField{}}, CheckedAt: input.CheckedAt.UTC(), Reason: input.Reason}
		if input.SourceID != "" {
			check.ChannelEvidence, err = s.verificationSource(ctx, input.SourceID, request.MunicipalityISTAT, 0)
			if err != nil {
				return out, err
			}
			if check.Role != input.Role {
				return out, ErrVerificationEvidence
			}
		}
		switch input.State {
		case "missing":
			check.Outcome = "missing_evidence"
		case "unavailable":
			check.Outcome = "unavailable"
		case "not_applicable":
			check.Outcome = "not_applicable"
		case "available":
			check.ChannelEvidence, err = s.loadVerificationEvidence(ctx, docs, input.VerificationEvidence, request.MunicipalityISTAT, at)
			if err != nil {
				if !errors.Is(err, documents.ErrStorage) && !errors.Is(err, pgx.ErrNoRows) {
					return out, err
				}
				if check.SourceID == "" {
					check.ChannelEvidence, err = s.verificationSource(ctx, input.SourceID, request.MunicipalityISTAT, 0)
					if err != nil {
						return out, err
					}
				}
				check.VersionID = input.VersionID
				check.Outcome = "unavailable"
				check.Reason = "retained_evidence_unavailable"
			} else {
				if check.Role != input.Role {
					return out, ErrVerificationEvidence
				}
				if !check.Complete {
					check.Outcome = "unavailable"
					check.Reason = "required_evidence_incomplete"
				}
			}
		}
		checks[input.Role] = check
	}
	if !candidate.Complete {
		own := checks[candidate.Role]
		own.Outcome = "unavailable"
		own.Reason = "candidate_required_evidence_incomplete"
		checks[candidate.Role] = own
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,730068))", request.RequestID); err != nil {
		return out, err
	}
	if old, err := s.existingVerification(ctx, tx, request.RequestID, hash); err != pgx.ErrNoRows {
		if err != nil {
			return out, err
		}
		return old, tx.Commit(ctx)
	}
	var territorial bool
	if err = tx.QueryRow(ctx, "SELECT to_regprocedure('territorial_source_allowed(text)') IS NOT NULL").Scan(&territorial); err != nil {
		return out, err
	}
	ids := []string{}
	for _, c := range checks {
		if c.SourceID != "" && !slices.Contains(ids, c.SourceID) {
			ids = append(ids, c.SourceID)
		}
	}
	slices.Sort(ids)
	for _, id := range ids {
		var suspended *time.Time
		allowed := true
		query := "SELECT interpretation_suspended_at,true FROM registry_sources WHERE id=$1 FOR SHARE"
		if territorial {
			query = "SELECT interpretation_suspended_at,territorial_source_allowed(id) FROM registry_sources WHERE id=$1 FOR SHARE"
		}
		if err = tx.QueryRow(ctx, query, id).Scan(&suspended, &allowed); err != nil {
			return out, err
		}
		if suspended != nil || !allowed {
			for role, c := range checks {
				if c.SourceID == id && c.VersionID > 0 {
					c.Outcome = "unavailable"
					c.Reason = "interpretation_or_territorial_scope_suspended"
					checks[role] = c
				}
			}
		}
	}
	for _, check := range checks {
		if check.Hash != "" {
			var first time.Time
			if err = tx.QueryRow(ctx, "SELECT first_acquired_at FROM retained_versions WHERE id=$1", check.VersionID).Scan(&first); err != nil {
				return out, err
			}
			if first.After(at) {
				return out, ErrVerificationEvidence
			}
		}
	}
	primary := checks[primaryRole(request.Kind)]
	for _, role := range verificationRoles {
		check := checks[role]
		if check.Outcome == "" || role == candidate.Role && check.Outcome == "corroborated" {
			check.ComparisonFromSourceID, check.ComparisonFromVersionID = candidate.SourceID, candidate.VersionID
			check.ComparisonToSourceID, check.ComparisonToVersionID = check.SourceID, check.VersionID
			check.Outcome, check.Reason, check.ComparedFields = compareEvidence(request.Kind, candidate, check.ChannelEvidence)
		}
		// Every republication is compared to the originating primary too. This
		// records disagreement even when two secondary channels agree with each other.
		if (request.Kind == "regional_record" && role != "regional" || role == candidate.Role && role != primaryRole(request.Kind)) && check.VersionID > 0 && check.Complete && check.Outcome != "unavailable" {
			check.ComparisonFromSourceID, check.ComparisonFromVersionID = check.SourceID, check.VersionID
			check.ComparisonToSourceID, check.ComparisonToVersionID = primary.SourceID, primary.VersionID
			if primary.VersionID > 0 && primary.Complete && primary.Outcome != "unavailable" {
				check.Outcome, check.Reason, check.ComparedFields = compareEvidence(request.Kind, check.ChannelEvidence, primary.ChannelEvidence)
			} else {
				check.Outcome, check.Reason, check.ComparedFields = primary.Outcome, "required_primary_not_available_for_comparison", nil
				for _, name := range requiredFacts(request.Kind, check.Fields) {
					result := FieldVerification{Field: name, Outcome: check.Outcome, Reason: check.Reason}
					if field, ok := check.Fields[name]; ok {
						result.Candidate = &field
					}
					check.ComparedFields = append(check.ComparedFields, result)
				}
			}
		}
		checks[role] = check
		out.Checks = append(out.Checks, check)
	}
	primary = checks[primaryRole(request.Kind)]
	out.Outcome = primary.Outcome
	out.Admitted = candidate.Complete && checks[candidate.Role].Outcome != "unavailable" && admitsFields(request.Kind, primary, candidate)
	// A primary-only publication needs no platform republication. Its own
	// supported fields remain admitted even when the other channels are missing.
	if candidate.Role == primaryRole(request.Kind) {
		out.Admitted = candidate.Complete && admitsFields(request.Kind, primary, candidate)
	}
	if primary.VersionID > 0 && primary.Complete && primary.Outcome != "unavailable" && !request.ComparisonOnly {
		out.PrimaryRecordID, out.ProjectionState, err = s.projectVerificationPrimary(ctx, tx, request.Kind, request.MunicipalityISTAT, primary.ChannelEvidence, at.UTC())
		if err != nil {
			return out, err
		}
	}
	if out.Admitted && out.PrimaryRecordID != "" {
		out.DomainRecordID = out.PrimaryRecordID
	} else {
		out.Admitted = false
	}
	if !out.Admitted && out.PrimaryRecordID != "" {
		out.ProjectionState = "primary_only"
	}
	body, err := json.Marshal(out)
	if err != nil {
		return out, ErrInvalid
	}
	if err = tx.QueryRow(ctx, `INSERT INTO domain_verification_receipts(request_id,request_sha256,candidate_key,candidate_source_id,candidate_version_id,kind,municipality_istat,verified_at,body)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, request.RequestID, hash, request.CandidateKey, candidate.SourceID, candidate.VersionID, request.Kind, request.MunicipalityISTAT, at.UTC(), body).Scan(&out.ID); err != nil {
		return out, err
	}
	for _, check := range out.Checks {
		if check.VersionID > 0 && check.Hash != "" && check.VersionID != candidate.VersionID {
			if _, err = tx.Exec(ctx, `INSERT INTO domain_verification_dependencies(receipt_id,supported_version_id,evidence_version_id) VALUES($1,$2,$3),($1,$3,$2) ON CONFLICT DO NOTHING`, out.ID, candidate.VersionID, check.VersionID); err != nil {
				return out, err
			}
		}
	}
	return out, tx.Commit(ctx)
}

func (s *Store) existingVerification(ctx context.Context, q querierVerification, id, hash string) (VerificationReceipt, error) {
	var receipt VerificationReceipt
	var body []byte
	var stored string
	err := q.QueryRow(ctx, "SELECT id,request_sha256,body FROM domain_verification_receipts WHERE request_id=$1", id).Scan(&receipt.ID, &stored, &body)
	if err != nil {
		return receipt, err
	}
	if stored != hash {
		return receipt, ErrConflict
	}
	number := receipt.ID
	if json.Unmarshal(body, &receipt) != nil {
		return receipt, ErrInvalid
	}
	receipt.ID = number
	return receipt, nil
}

type querierVerification interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (s *Store) VerificationReceipt(ctx context.Context, id int64) (VerificationReceipt, error) {
	var value VerificationReceipt
	var body []byte
	if id < 1 {
		return value, ErrInvalid
	}
	if err := s.pool.QueryRow(ctx, "SELECT body FROM domain_verification_receipts WHERE id=$1", id).Scan(&body); err != nil {
		return value, err
	}
	if json.Unmarshal(body, &value) != nil {
		return value, ErrInvalid
	}
	value.ID = id
	return value, nil
}

func (s *Store) VerificationHistory(ctx context.Context, source, candidate string, after int64, limit int) ([]VerificationReceipt, error) {
	if source == "" || candidate == "" || after < 0 || limit < 1 || limit > 100 {
		return nil, ErrInvalid
	}
	rows, err := s.pool.Query(ctx, "SELECT id,body FROM domain_verification_receipts WHERE candidate_source_id=$1 AND candidate_key=$2 AND id>$3 ORDER BY id LIMIT $4", source, candidate, after, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []VerificationReceipt{}
	for rows.Next() {
		var id int64
		var body []byte
		if err = rows.Scan(&id, &body); err != nil {
			return nil, err
		}
		var value VerificationReceipt
		if json.Unmarshal(body, &value) != nil {
			return nil, ErrInvalid
		}
		value.ID = id
		values = append(values, value)
	}
	return values, rows.Err()
}
