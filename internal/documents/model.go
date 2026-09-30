// Package documents retains original evidence independently of interpretation.
package documents

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid acquisition")
	ErrConflict = errors.New("acquisition ID reused with different input")
	ErrPolicy   = errors.New("source does not permit retention")
	ErrStorage  = errors.New("object storage unavailable or integrity check failed")
)

// Resource declares every necessary dependency, including inaccessible ones.
// SourceID/Configuration identify the policy that permits retaining its bytes;
// external attachments must use their own verified source configuration.
type Resource struct {
	URL           string `json:"url"`
	Role          string `json:"role"` // original, attachment, or resource
	Required      bool   `json:"required"`
	SourceID      string `json:"source_id"`
	Configuration int    `json:"configuration"`
	MediaType     string `json:"media_type"`
	Bytes         []byte `json:"bytes"`
	Missing       string `json:"missing,omitempty"` // stable reason code, never raw upstream errors
}

type Acquisition struct {
	ID            string          `json:"id"` // caller-generated idempotency key; preserve across retries
	SourceID      string          `json:"source_id"`
	Configuration int             `json:"configuration"`
	URL           string          `json:"url"`
	IssuerID      *string         `json:"issuer_id,omitempty"` // unknown stays unknown, not inferred from publisher
	Metadata      json.RawMessage `json:"metadata"`            // original publication/update expressions, not acquisition time
	Resources     []Resource      `json:"resources"`
}

type Reference struct {
	Inference     *registry.EligibilityDecision `json:"inference,omitempty"`
	URL           string                        `json:"url"`
	Role          string                        `json:"role"`
	Required      bool                          `json:"required"`
	SourceID      string                        `json:"source_id"`
	Configuration int                           `json:"configuration"`
	MediaType     string                        `json:"media_type"`
	Hash          string                        `json:"sha256,omitempty"`
	Size          int64                         `json:"size"`
	Missing       string                        `json:"missing,omitempty"`
}

type Version struct {
	IssuerID        *string
	ID              int64
	DocumentID      int64
	Hash            string
	FirstAcquiredAt time.Time
	Complete        bool
	Metadata        json.RawMessage
	Resources       []Reference
}

const MaxAcquisitionBytes = 32 << 20

// Drupal Views generates a new opaque DOM marker on every response. It can
// appear in comments and CSS classes; neither changes publication content.
// Retain the first response's original bytes unchanged.
var drupalViewDOMID = regexp.MustCompile(`js-view-dom-id-[0-9a-f]{64}`)

func digest(b []byte) string       { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func objectKey(hash string) string { return "sha256/" + hash[:2] + "/" + hash }
func validURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Host != "" && (u.Scheme == "https" || u.Scheme == "http") && u.User == nil && u.Fragment == ""
}

func prepare(a Acquisition) (Acquisition, []Reference, string, bool, error) {
	if strings.TrimSpace(a.ID) == "" || a.SourceID == "" || a.Configuration < 1 || !validURL(a.URL) || len(a.Resources) == 0 || len(a.Resources) > 256 {
		return a, nil, "", false, ErrInvalid
	}
	if len(a.Metadata) == 0 {
		a.Metadata = json.RawMessage(`{}`)
	}
	var metadata map[string]any
	if json.Unmarshal(a.Metadata, &metadata) != nil || metadata == nil {
		return a, nil, "", false, ErrInvalid
	}
	// Canonicalize metadata and resource ordering so transport order is irrelevant.
	var canonical any
	dec := json.NewDecoder(strings.NewReader(string(a.Metadata)))
	dec.UseNumber()
	if dec.Decode(&canonical) != nil {
		return a, nil, "", false, ErrInvalid
	}
	a.Metadata, _ = json.Marshal(canonical)
	a.Resources = append([]Resource(nil), a.Resources...)
	sort.Slice(a.Resources, func(i, j int) bool { return a.Resources[i].URL < a.Resources[j].URL })
	refs := make([]Reference, 0, len(a.Resources))
	seen := map[string]bool{}
	original := 0
	total := len(a.Metadata)
	complete := true
	for _, r := range a.Resources {
		if !validURL(r.URL) || seen[r.URL] || r.SourceID == "" || r.Configuration < 1 || (r.Role != "original" && r.Role != "attachment" && r.Role != "resource") {
			return a, nil, "", false, ErrInvalid
		}
		seen[r.URL] = true
		if r.Role == "original" {
			original++
			if r.URL != a.URL || !r.Required || r.Missing != "" || r.SourceID != a.SourceID || r.Configuration != a.Configuration {
				return a, nil, "", false, ErrInvalid
			}
		}
		ref := Reference{URL: r.URL, Role: r.Role, Required: r.Required, SourceID: r.SourceID, Configuration: r.Configuration, MediaType: r.MediaType, Missing: r.Missing}
		if r.Missing != "" {
			if r.Bytes != nil || (r.Missing != "unavailable" && r.Missing != "forbidden" && r.Missing != "not_acquired") {
				return a, nil, "", false, ErrInvalid
			}
			if r.Required {
				complete = false
			}
		} else {
			if r.Bytes == nil || r.MediaType == "" {
				return a, nil, "", false, ErrInvalid
			}
			total += len(r.Bytes)
			ref.Hash = digest(r.Bytes)
			ref.Size = int64(len(r.Bytes))
		}
		refs = append(refs, ref)
	}
	if original != 1 || total > MaxAcquisitionBytes {
		return a, nil, "", false, ErrInvalid
	}
	// Policy revisions do not turn unchanged evidence into a new content version.
	contentRefs := append([]Reference(nil), refs...)
	for i := range contentRefs {
		contentRefs[i].Configuration = 0
		if contentRefs[i].Missing == "" && strings.HasPrefix(strings.ToLower(contentRefs[i].MediaType), "text/html") {
			contentRefs[i].Hash = digest(drupalViewDOMID.ReplaceAll(a.Resources[i].Bytes, []byte("js-view-dom-id")))
		} else if contentRefs[i].Missing == "" && contentRefs[i].Role == "resource" && (a.SourceID == "cfr-criticality" || a.SourceID == "cfr-vigilance") {
			// CFR regenerates its print PDF and PNG graphics on every request.
			// Only known transport metadata is ignored; retained object hashes
			// above still address the exact original bytes.
			canonical := cfrContentIdentity(a.Resources[i].Bytes, contentRefs[i].MediaType)
			contentRefs[i].Hash = digest(canonical)
			contentRefs[i].Size = int64(len(canonical))
		}
	}
	b, _ := json.Marshal(struct {
		Issuer    *string
		Metadata  json.RawMessage
		Resources []Reference
	}{a.IssuerID, a.Metadata, contentRefs})
	return a, refs, digest(b), complete, nil
}

// InferenceEligible defaults conservatively for legacy and in-memory evidence.
func (r Reference) InferenceEligible() bool {
	return r.Inference == nil || r.Inference.Eligible || r.Missing != "" || r.Role != "resource"
}
