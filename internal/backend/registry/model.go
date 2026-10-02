// Package registry persists source identity, immutable configurations and
// operator decisions. It does not acquire sources or expose public mutations.
package registry

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInvalid      = errors.New("invalid registry input")
	ErrConflict     = errors.New("registry revision conflict")
	ErrPrerequisite = errors.New("registry prerequisite missing")
	ErrNotFound     = errors.New("registry source not found")
)

type Authority struct{ ID, Name, OfficialURL string }
type Channel struct {
	ID, PublisherID, Platform, URL string
	External                       bool
}
type Source struct{ ID, AuthorityID, ChannelID, ProductID, Territory string }
type Evidence struct {
	URL        string          `json:"url"`
	Locator    string          `json:"locator"`
	ObservedAt time.Time       `json:"observed_at"`
	Report     json.RawMessage `json:"report,omitempty"`
}
type Referral struct {
	Evidence    Evidence `json:"evidence"`
	Destination string   `json:"destination"`
	Context     string   `json:"context"`
	Sections    []string `json:"sections"`
	ProductID   string   `json:"product_id"`
	Territory   string   `json:"territory"`
}
type Policy struct {
	Evidence             *Evidence `json:"evidence,omitempty"`
	CollectionPermitted  bool      `json:"collection_permitted"`
	RetentionPermitted   bool      `json:"retention_permitted"`
	PublicationPermitted bool      `json:"publication_permitted"`
	CopiesPermitted      bool      `json:"copies_permitted"`
	Conditions           string    `json:"conditions"`
}

type ExpectedPublication struct {
	Evidence         Evidence  `json:"evidence"`
	Anchor           time.Time `json:"anchor"`
	IntervalSeconds  int       `json:"interval_seconds"`
	ToleranceSeconds int       `json:"tolerance_seconds"`
}

// RegionalProductContract describes one operator-reviewed CFR product endpoint.
// Resource prefixes are URL paths on the same origin and deliberately constrain
// which graphical dependencies may be acquired from otherwise dynamic HTML.
type RegionalProductContract struct {
	Kind                   string   `json:"kind"`
	ContentMarkers         []string `json:"content_markers"`
	PrintURL               string   `json:"print_url,omitempty"`
	ResourcePathPrefixes   []string `json:"resource_path_prefixes,omitempty"`
	MinimumGraphicalAssets int      `json:"minimum_graphical_assets,omitempty"`
	PrintContainsGraphics  bool     `json:"print_contains_graphics,omitempty"`
}

// Discovery confines acquisition to operator-declared listing boundaries.
// Document prefixes are URL paths on the configured source origin; pagination
// may only vary the declared query parameter on an exact section path.
type Discovery struct {
	DocumentPathPrefixes  []string `json:"document_path_prefixes,omitempty"`
	PaginationParameter   string   `json:"pagination_parameter,omitempty"`
	MaxPagesPerSection    int      `json:"max_pages_per_section,omitempty"`
	MaxDocuments          int      `json:"max_documents,omitempty"`
	ListingContentMarkers []string `json:"listing_content_markers,omitempty"`
	ListingItemClass      string   `json:"listing_item_class,omitempty"`
	ListingDateClass      string   `json:"listing_date_class,omitempty"`
	ListingDateLabel      string   `json:"listing_date_label,omitempty"`
	ListingDateLayout     string   `json:"listing_date_layout,omitempty"`
	ListingDateLocale     string   `json:"listing_date_locale,omitempty"`
	BootstrapDays         int      `json:"bootstrap_days,omitempty"`
}

// Every configuration is a complete snapshot, never a patch. Missing evidence
// remains explicit; no source/product inherits another source's permissions.
type Configuration struct {
	ProcessingProfile    string                   `json:"processing_profile,omitempty"`
	InferenceEligibility *InferenceEligibility    `json:"inference_eligibility,omitempty"`
	Attachments          *AttachmentPolicy        `json:"attachments,omitempty"`
	Unresolved           []string                 `json:"unresolved"`
	URL                  string                   `json:"url"`
	Sections             []string                 `json:"sections"`
	AccessMethod         string                   `json:"access_method"`
	Attribution          string                   `json:"attribution"`
	Provenance           *Evidence                `json:"provenance,omitempty"`
	Referral             *Referral                `json:"referral,omitempty"`
	Policy               Policy                   `json:"policy"`
	CheckSeconds         int                      `json:"check_seconds"`
	DelaySeconds         int                      `json:"delay_seconds"`
	BackoffBaseSeconds   int                      `json:"backoff_base_seconds"`
	BackoffMaxSeconds    int                      `json:"backoff_max_seconds"`
	Limitations          []string                 `json:"limitations"`
	Discovery            Discovery                `json:"discovery,omitempty"`
	ExpectedPublication  *ExpectedPublication     `json:"expected_publication,omitempty"`
	RegionalProduct      *RegionalProductContract `json:"regional_product,omitempty"`
}
type Version struct {
	SourceID      string
	Revision      int
	Configuration Configuration
	Actor         string
	CreatedAt     time.Time
}
type State struct {
	Source                           Source
	LatestRevision                   int
	ActiveRevision                   *int
	CollectionEnabled, PublicEnabled bool
	Accepted                         bool
	InterpretationSuspendedAt        *time.Time
}
type Acceptance struct {
	Report                     Evidence  `json:"report"`
	PeriodStart                time.Time `json:"period_start"`
	PeriodEnd                  time.Time `json:"period_end"`
	Sections                   []string  `json:"sections"`
	RiskCoverage               []string  `json:"risk_coverage"`
	ExtractionVerified         bool      `json:"extraction_verified"`
	UpdatesVerified            bool      `json:"updates_verified"`
	AttachmentsVerified        bool      `json:"attachments_verified"`
	ScannedAttachmentsVerified bool      `json:"scanned_attachments_verified"`
	HistoryVerified            bool      `json:"history_verified"`
	FailureBehaviorVerified    bool      `json:"failure_behavior_verified"`
	InterfacesEquivalent       bool      `json:"interfaces_equivalent"`
	CoverageStatus             string    `json:"coverage_status"`
	CoverageLimitations        []string  `json:"coverage_limitations"`
	KnownOmissions             int       `json:"known_omissions"`
	UnsupportedAssertions      int       `json:"unsupported_assertions"`
}
type Event struct {
	Revision    int
	Kind, Actor string
	CreatedAt   time.Time
	Evidence    json.RawMessage
}

func validURL(raw string) bool {
	u, e := url.Parse(raw)
	return e == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && u.User == nil && u.Fragment == ""
}
func validEvidence(e Evidence) bool {
	return validURL(e.URL) && strings.TrimSpace(e.Locator) != "" && !e.ObservedAt.IsZero()
}
func (c *Configuration) defaults() {
	if c.CheckSeconds == 0 {
		c.CheckSeconds = 600
	}
	if c.DelaySeconds == 0 {
		c.DelaySeconds = 1800
	}
	if c.BackoffBaseSeconds == 0 {
		c.BackoffBaseSeconds = 60
	}
	if c.BackoffMaxSeconds == 0 {
		c.BackoffMaxSeconds = 3600
	}
	if c.Discovery.BootstrapDays == 0 {
		c.Discovery.BootstrapDays = 30
	}
}
func (c Configuration) valid() bool {
	if !c.Attachments.Valid() || c.Attachments != nil && (c.AccessMethod != "crawl4ai" || c.RegionalProduct != nil) {
		return false
	}
	if !c.InferenceEligibility.valid(c.URL) {
		return false
	}
	if !validURL(c.URL) || len(c.Sections) == 0 || strings.TrimSpace(c.AccessMethod) == "" || strings.TrimSpace(c.Attribution) == "" || c.CheckSeconds <= 0 || c.DelaySeconds <= 0 || c.BackoffBaseSeconds <= 0 || c.BackoffMaxSeconds < c.BackoffBaseSeconds {
		return false
	}
	seen := map[string]bool{}
	for _, s := range c.Sections {
		if !validURL(s) || seen[s] {
			return false
		}
		seen[s] = true
	}
	if c.Provenance != nil && !validEvidence(*c.Provenance) {
		return false
	}
	if c.Policy.Evidence != nil && !validEvidence(*c.Policy.Evidence) {
		return false
	}
	if (c.Policy.CollectionPermitted || c.Policy.RetentionPermitted || c.Policy.PublicationPermitted || c.Policy.CopiesPermitted) && c.Policy.Evidence == nil {
		return false
	}
	if c.Policy.CopiesPermitted && (!c.Policy.PublicationPermitted || !c.Policy.RetentionPermitted) {
		return false
	}
	if c.Referral != nil {
		r := c.Referral
		if !validEvidence(r.Evidence) || !validURL(r.Destination) || strings.TrimSpace(r.Context) == "" || r.ProductID == "" || r.Territory == "" {
			return false
		}
	}
	d := c.Discovery
	parserFields := []string{d.ListingItemClass, d.ListingDateClass, d.ListingDateLayout, d.ListingDateLocale}
	parserSet := 0
	for _, value := range parserFields {
		if strings.TrimSpace(value) != "" {
			parserSet++
		}
	}
	if parserSet != 0 && parserSet != len(parserFields) {
		return false
	}
	if d.ListingDateLabel != "" && parserSet != len(parserFields) {
		return false
	}
	if d.BootstrapDays < 1 || d.BootstrapDays > 366 {
		return false
	}
	if expected := c.ExpectedPublication; expected != nil {
		if !validEvidence(expected.Evidence) || expected.Anchor.IsZero() || expected.IntervalSeconds < 60 || expected.ToleranceSeconds < 0 || expected.ToleranceSeconds > expected.IntervalSeconds {
			return false
		}
	}
	if regional := c.RegionalProduct; regional != nil {
		if len(c.Sections) != 1 || c.Sections[0] != c.URL || !validRegionalKind(regional.Kind) || len(regional.ContentMarkers) == 0 || regional.MinimumGraphicalAssets < 0 {
			return false
		}
		if regional.Kind == "monitoring" && c.ExpectedPublication != nil {
			return false
		}
		seenMarkers := map[string]bool{}
		for _, marker := range regional.ContentMarkers {
			marker = strings.TrimSpace(marker)
			if marker == "" || seenMarkers[marker] {
				return false
			}
			seenMarkers[marker] = true
		}
		if regional.PrintURL != "" && (!validURL(regional.PrintURL) || !sameOrigin(c.URL, regional.PrintURL)) {
			return false
		}
		if regional.PrintContainsGraphics && regional.PrintURL == "" {
			return false
		}
		if (regional.Kind == "vigilance" || regional.Kind == "criticality") && regional.MinimumGraphicalAssets == 0 && !regional.PrintContainsGraphics {
			return false
		}
		seenPrefixes := map[string]bool{}
		for _, prefix := range regional.ResourcePathPrefixes {
			if !strings.HasPrefix(prefix, "/") || strings.ContainsAny(prefix, "?#") || seenPrefixes[prefix] {
				return false
			}
			seenPrefixes[prefix] = true
		}
	}
	return true
}

func validRegionalKind(kind string) bool {
	return kind == "vigilance" || kind == "criticality" || kind == "monitoring"
}

func sameOrigin(left, right string) bool {
	a, errA := url.Parse(left)
	b, errB := url.Parse(right)
	return errA == nil && errB == nil && a.Scheme == b.Scheme && strings.EqualFold(a.Host, b.Host)
}

func (c Configuration) regionalKind() string {
	if c.RegionalProduct == nil {
		return ""
	}
	return c.RegionalProduct.Kind
}
