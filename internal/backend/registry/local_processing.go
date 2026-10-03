package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
)

const LocalProcessingVersion = "local-processing-v1"

// LocalProcessing declares reviewed interpretation boundaries, never acquisition
// permissions. Text-only PDF scopes must exclude maps and meaningful graphics.
type LocalProcessing struct {
	Version             string   `json:"version"`
	Evidence            Evidence `json:"evidence"`
	HTMLBodySelectors   []string `json:"html_body_selectors,omitempty"`
	TextPDFPathPrefixes []string `json:"text_pdf_path_prefixes,omitempty"`
	StructuredFormat    string   `json:"structured_format,omitempty"`
}

var bodySelector = regexp.MustCompile(`^(?:[a-z][a-z0-9-]*)?(?:[.#][A-Za-z_][A-Za-z0-9_-]*)?$`)

func ValidBodySelector(s string) bool { return s != "" && len(s) <= 128 && bodySelector.MatchString(s) }

func (p *LocalProcessing) valid(regional *RegionalProductContract) bool {
	if p == nil {
		return true
	}
	if p.Version != LocalProcessingVersion || !validEvidence(p.Evidence) || len(p.HTMLBodySelectors) > 16 || len(p.TextPDFPathPrefixes) > 32 {
		return false
	}
	if len(p.HTMLBodySelectors)+len(p.TextPDFPathPrefixes) == 0 && p.StructuredFormat == "" {
		return false
	}
	seen := map[string]bool{}
	for _, s := range p.HTMLBodySelectors {
		if !ValidBodySelector(s) || seen[s] {
			return false
		}
		seen[s] = true
	}
	for _, prefix := range p.TextPDFPathPrefixes {
		// A canonical directory, rather than a broad root or filename substring.
		if prefix == "/" || len(prefix) > 1024 || !strings.HasPrefix(prefix, "/") || !strings.HasSuffix(prefix, "/") || strings.ContainsAny(prefix, "%?#\\* \t\r\n") || strings.Contains(prefix, "//") || strings.Contains(prefix, "/../") || strings.Contains(prefix, "/./") || seen[prefix] {
			return false
		}
		seen[prefix] = true
	}
	if p.StructuredFormat != "" {
		if regional == nil || p.StructuredFormat != "cfr-"+regional.Kind+"-v1" || !validRegionalKind(regional.Kind) || len(p.HTMLBodySelectors) > 0 {
			return false
		}
	}
	// Graphical regional products are never declared text-only.
	return regional == nil || len(p.TextPDFPathPrefixes) == 0
}

func (p *LocalProcessing) Identity() string {
	if p == nil {
		return ""
	}
	body, _ := json.Marshal(p)
	sum := sha256.Sum256(body)
	return LocalProcessingVersion + ":" + hex.EncodeToString(sum[:])
}

func (p *LocalProcessing) TextPDF(raw string) bool {
	if p == nil || p.Version != LocalProcessingVersion {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || !validURL(raw) || u.RawPath != "" || strings.ContainsAny(u.Path, "\\") || strings.Contains(u.Path, "//") || strings.Contains(u.Path, "/../") || strings.Contains(u.Path, "/./") {
		return false
	}
	for _, prefix := range p.TextPDFPathPrefixes {
		if strings.HasPrefix(u.Path, prefix) {
			return true
		}
	}
	return false
}
