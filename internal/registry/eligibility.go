package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

const EligibilityVersion = "resource-eligibility-v1"

// DecorationRule is an exact, reviewed exception, never a filename/size guess.
// A different image at the same URL immediately becomes eligible again.
type DecorationRule struct {
	URL      string   `json:"url"`
	SHA256   string   `json:"sha256"`
	Reason   string   `json:"reason"`
	Evidence Evidence `json:"evidence"`
}
type InferenceEligibility struct {
	Version     string           `json:"version"`
	Decorations []DecorationRule `json:"decorations"`
}
type EligibilityDecision struct {
	Eligible bool   `json:"eligible"`
	Policy   string `json:"policy"`
	Reason   string `json:"reason"`
}

var eligibilityHash = regexp.MustCompile(`^[0-9a-f]{64}$`)

func (p *InferenceEligibility) valid(origin string) bool {
	if p == nil {
		return true
	}
	if p.Version != EligibilityVersion || len(p.Decorations) > 64 {
		return false
	}
	seen := map[string]bool{}
	for _, rule := range p.Decorations {
		if !validURL(rule.URL) || !sameOrigin(origin, rule.URL) || !eligibilityHash.MatchString(rule.SHA256) || rule.Reason != "reviewed_decoration" || !validEvidence(rule.Evidence) || seen[rule.URL] {
			return false
		}
		seen[rule.URL] = true
	}
	return true
}

func (p *InferenceEligibility) Decide(url, hash, role, missing string) EligibilityDecision {
	decision := EligibilityDecision{Eligible: true, Policy: EligibilityVersion + ":default", Reason: "meaningful_or_unknown"}
	if p == nil {
		return decision
	}
	encoded, _ := json.Marshal(p)
	sum := sha256.Sum256(encoded)
	decision.Policy = EligibilityVersion + ":" + hex.EncodeToString(sum[:])
	// Missing required evidence is never hidden by a decoration exception.
	if p.Version != EligibilityVersion || role != "resource" || missing != "" || !eligibilityHash.MatchString(hash) {
		return decision
	}
	for _, rule := range p.Decorations {
		if rule.URL == url && rule.SHA256 == hash && rule.Reason == "reviewed_decoration" && validEvidence(rule.Evidence) {
			decision.Eligible = false
			decision.Reason = rule.Reason
			break
		}
	}
	return decision
}
