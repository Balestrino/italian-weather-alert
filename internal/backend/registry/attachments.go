package registry

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

// AttachmentScope authorizes only linked resources, not discovery of a channel.
// Its referral and reuse policy belong to the enclosing source revision.
type AttachmentScope struct {
	Origin     string   `json:"origin"`
	PathPrefix string   `json:"path_prefix"`
	Referral   Evidence `json:"referral"`
	Policy     Policy   `json:"policy"`
}

type AttachmentPolicy struct {
	ContentClass string            `json:"content_class,omitempty"`
	ValidatePDF  bool              `json:"validate_pdf"`
	External     []AttachmentScope `json:"external,omitempty"`
}

var attachmentClass = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

func (p *AttachmentPolicy) Valid() bool {
	if p == nil {
		return true
	}
	if p.ContentClass != "" && !attachmentClass.MatchString(p.ContentClass) || len(p.External) > 32 {
		return false
	}
	seen := map[string]bool{}
	for _, scope := range p.External {
		u, err := url.Parse(scope.Origin)
		prefix := scope.PathPrefix
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.ContainsAny(u.Host, "*%\\") {
			return false
		}
		if prefix == "/" || !strings.HasPrefix(prefix, "/") || !strings.HasSuffix(prefix, "/") || strings.ContainsAny(prefix, "?#%\\") || path.Clean(prefix)+"/" != prefix {
			return false
		}
		key := strings.ToLower(scope.Origin) + prefix
		if seen[key] || !validEvidence(scope.Referral) || scope.Policy.Evidence == nil || !validEvidence(*scope.Policy.Evidence) || !scope.Policy.CollectionPermitted || !scope.Policy.RetentionPermitted || strings.TrimSpace(scope.Policy.Conditions) == "" {
			return false
		}
		seen[key] = true
	}
	return true
}

func (p *AttachmentPolicy) Allows(parent, target string) bool {
	u, err := url.Parse(target)
	if err != nil || u.Host == "" || u.User != nil || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	if sameOrigin(parent, target) {
		return true
	}
	if p == nil || !p.Valid() || u.Scheme != "https" || u.RawQuery != "" || u.ForceQuery || u.RawPath != "" || strings.Contains(u.Path, "\\") || path.Clean(u.Path) != u.Path {
		return false
	}
	for _, scope := range p.External {
		if sameOrigin(scope.Origin, target) && strings.HasPrefix(u.Path, scope.PathPrefix) {
			return true
		}
	}
	return false
}
