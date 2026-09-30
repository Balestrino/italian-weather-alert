package linking

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
)

var responseFormat = json.RawMessage(`{"type":"json_schema","json_schema":{"name":"measure_update_link","strict":true,"schema":{"type":"object","properties":{"status":{"type":"string","enum":["linked","unresolved","no_relation"]},"reason_code":{"type":"string","enum":["partial_reopening","cross_document_evidence","explicit_reference","ambiguous_relation","insufficient_evidence"]},"relation":{"type":["string","null"],"enum":["reopens","cancels","extends","corrects","updates",null]},"candidate_index":{"type":["integer","null"],"minimum":1},"current_evidence_indices":{"type":"array","items":{"type":"integer","minimum":1}},"candidate_evidence_indices":{"type":"array","items":{"type":"integer","minimum":1}}},"required":["status","reason_code","relation","candidate_index","current_evidence_indices","candidate_evidence_indices"],"additionalProperties":false}}}`)

type decision struct {
	Status                   string  `json:"status"`
	ReasonCode               string  `json:"reason_code"`
	Relation                 *string `json:"relation"`
	CandidateIndex           *int    `json:"candidate_index"`
	CurrentEvidenceIndices   []int   `json:"current_evidence_indices"`
	CandidateEvidenceIndices []int   `json:"candidate_evidence_indices"`
}

func Request(model string, current MeasureContext, candidates []Candidate) (inference.Request, error) {
	if model == "" || current.RunID < 1 || current.Ordinal < 1 || len(candidates) == 0 || len(candidates) > 100 {
		return inference.Request{}, ErrInvalid
	}
	body, err := json.Marshal(struct {
		Current    MeasureContext `json:"current"`
		Candidates []Candidate    `json:"candidates"`
	}{current, candidates})
	if err != nil {
		return inference.Request{}, ErrInvalid
	}
	input, _ := json.Marshal(string(body))
	system, _ := json.Marshal(PromptBody)
	return inference.Request{Model: model, Messages: []inference.Message{{Role: "system", Content: system}, {Role: "user", Content: input}}, ResponseFormat: responseFormat, MaxCompletionTokens: 4096}, nil
}

func Parse(raw string, current MeasureContext, candidates []Candidate) (Result, error) {
	var value decision
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || !errors.Is(decoder.Decode(new(any)), io.EOF) {
		return Result{}, ErrInvalid
	}
	validReason := value.ReasonCode == "partial_reopening" || value.ReasonCode == "cross_document_evidence" || value.ReasonCode == "explicit_reference" || value.ReasonCode == "ambiguous_relation" || value.ReasonCode == "insufficient_evidence"
	if !validReason || (value.Status != "linked" && value.Status != "unresolved" && value.Status != "no_relation") {
		return Result{}, ErrInvalid
	}
	result := Result{CurrentRunID: current.RunID, CurrentOrdinal: current.Ordinal, Status: value.Status, ReasonCode: value.ReasonCode}
	if value.Status == "linked" {
		if value.Relation == nil || value.CandidateIndex == nil || *value.CandidateIndex < 1 || *value.CandidateIndex > len(candidates) || len(value.CurrentEvidenceIndices) == 0 || len(value.CandidateEvidenceIndices) == 0 {
			return Result{}, ErrInvalid
		}
		if ambiguousMissingPlace(current, candidates) {
			result.Status, result.ReasonCode = "unresolved", "ambiguous_relation"
			var err error
			result.CurrentEvidence, err = selectEvidence(current.Evidence, value.CurrentEvidenceIndices)
			return result, err
		}
		candidate := candidates[*value.CandidateIndex-1]
		result.Relation = value.Relation
		result.CandidateRunID, result.CandidateOrdinal = &candidate.RunID, &candidate.Ordinal
		var err error
		result.CurrentEvidence, err = selectEvidence(current.Evidence, value.CurrentEvidenceIndices)
		if err != nil {
			return Result{}, err
		}
		result.CandidateEvidence, err = selectEvidence(candidate.Evidence, value.CandidateEvidenceIndices)
		if err != nil {
			return Result{}, err
		}
		return result, nil
	}
	if value.Relation != nil || value.CandidateIndex != nil || len(value.CandidateEvidenceIndices) != 0 {
		return Result{}, ErrInvalid
	}
	if value.Status == "unresolved" && value.ReasonCode != "ambiguous_relation" && value.ReasonCode != "insufficient_evidence" {
		return Result{}, ErrInvalid
	}
	if value.Status == "no_relation" && value.ReasonCode != "insufficient_evidence" {
		return Result{}, ErrInvalid
	}
	var err error
	result.CurrentEvidence, err = selectEvidence(current.Evidence, value.CurrentEvidenceIndices)
	if err != nil && value.Status == "unresolved" {
		// An unresolved decision asserts no relationship or candidate. A malformed
		// optional locator must not turn that safe outcome into a model failure;
		// retain the raw response and persist the ambiguity without evidence.
		result.CurrentEvidence = nil
		return result, nil
	}
	return result, err
}

func ambiguousMissingPlace(current MeasureContext, candidates []Candidate) bool {
	if strings.TrimSpace(current.Place) != "" {
		return false
	}
	subject := strings.ToLower(strings.TrimSpace(current.Subject))
	places := map[string]bool{}
	for _, candidate := range candidates {
		candidateSubject := strings.ToLower(strings.TrimSpace(candidate.Subject))
		place := strings.ToLower(strings.TrimSpace(candidate.Place))
		if place == "" || subject == "" || (candidateSubject != subject && !strings.Contains(candidateSubject, subject) && !strings.Contains(subject, candidateSubject)) {
			continue
		}
		places[place] = true
	}
	return len(places) > 1
}

func selectEvidence(source []extraction.Evidence, indices []int) ([]extraction.Evidence, error) {
	selected := make([]extraction.Evidence, 0, len(indices))
	seen := map[int]bool{}
	for _, index := range indices {
		if index < 1 || index > len(source) || seen[index] {
			return nil, ErrInvalid
		}
		seen[index] = true
		selected = append(selected, source[index-1])
	}
	return selected, nil
}
