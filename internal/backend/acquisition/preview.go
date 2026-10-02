package acquisition

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"golang.org/x/net/html"
)

type registryStore interface {
	Version(context.Context, string, int) (registry.Version, error)
	RecordPreview(context.Context, string, int, string, registry.Evidence) error
}

type retentionStore interface {
	Retain(context.Context, documents.Acquisition) (documents.Version, error)
}

type Engine struct {
	Registry    registryStore
	Retained    retentionStore
	Crawler     Crawler
	Resources   Crawler
	Tracking    revisionTracker
	Now         func() time.Time
	validatePDF func(context.Context, Page) error
}

type RetainedPage struct {
	InferenceResources []documents.Reference `json:"inference_resources,omitempty"`
	URL                string                `json:"url"`
	VersionID          int64                 `json:"version_id"`
	Hash               string                `json:"content_sha256,omitempty"`
	Changed            bool                  `json:"changed,omitempty"`
	Workload           string                `json:"workload,omitempty"`
}

type SectionPreview struct {
	SectionURL string         `json:"section_url"`
	Listings   []RetainedPage `json:"listings"`
}

type Preview struct {
	SourceID      string           `json:"source_id"`
	Configuration int              `json:"configuration"`
	ObservedAt    time.Time        `json:"observed_at"`
	Sections      []SectionPreview `json:"sections"`
	Documents     []RetainedPage   `json:"documents"`
}

type CheckOutcome struct {
	SourceID              string
	Configuration         int
	StartedAt             time.Time
	FinishedAt            time.Time
	Reachable             bool
	ContentRecognized     bool
	Complete              bool
	ErrorCode             string
	RetryAfter            *time.Time
	Listings              int
	Documents             int
	Revisions             int
	NewDocuments          int
	PublicationObservedAt *time.Time
	Retained              []RetainedPage
}

type acquisitionState struct {
	reachable             bool
	contentRecognized     bool
	errorCode             string
	retryAfter            *time.Time
	newDocuments          int
	publicationObservedAt *time.Time
}

func (e *Engine) Preview(ctx context.Context, sourceID string, revision int, actor string) (Preview, error) {
	if e.Registry == nil || e.Retained == nil || e.Crawler == nil || strings.TrimSpace(sourceID) == "" || revision < 1 || strings.TrimSpace(actor) == "" {
		return Preview{}, ErrInvalidConfiguration
	}
	if guard, ok := e.Registry.(interface {
		CheckTerritorialExecution(context.Context, string, int) error
	}); ok {
		if err := guard.CheckTerritorialExecution(ctx, sourceID, revision); err != nil {
			return Preview{}, err
		}
	}
	v, err := e.Registry.Version(ctx, sourceID, revision)
	if err != nil {
		return Preview{}, err
	}
	cfg := v.Configuration
	if err = validateAcquisitionConfiguration(cfg); err != nil {
		return Preview{}, err
	}
	if cfg.RegionalProduct != nil && e.Resources == nil {
		return Preview{}, ErrInvalidConfiguration
	}
	report, _, err := e.acquire(ctx, sourceID, revision, cfg)
	if err != nil {
		return Preview{}, err
	}
	locator := fmt.Sprintf("acquisition preview: %d sections, %d listings, %d documents", len(report.Sections), listingCount(report.Sections), len(report.Documents))
	// Preview the candidate's policy, not the currently active source revision.
	previewEligibility := func(pages []RetainedPage) {
		for i := range pages {
			for j := range pages[i].InferenceResources {
				ref := &pages[i].InferenceResources[j]
				if ref.SourceID == sourceID {
					decision := cfg.InferenceEligibility.Decide(ref.URL, ref.Hash, ref.Role, ref.Missing)
					ref.Inference = &decision
				}
			}
		}
	}
	previewEligibility(report.Documents)
	for i := range report.Sections {
		previewEligibility(report.Sections[i].Listings)
	}
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return Preview{}, err
	}
	if err = e.Registry.RecordPreview(ctx, sourceID, revision, actor, registry.Evidence{URL: cfg.Sections[0], Locator: locator, ObservedAt: report.ObservedAt, Report: reportJSON}); err != nil {
		return Preview{}, err
	}
	return report, nil
}

func (e *Engine) Check(ctx context.Context, sourceID string, revision int, startedAt time.Time) CheckOutcome {
	now := time.Now
	if e.Now != nil {
		now = e.Now
	}
	outcome := CheckOutcome{SourceID: sourceID, Configuration: revision, StartedAt: startedAt.UTC()}
	if startedAt.IsZero() || e.Registry == nil || e.Retained == nil || e.Crawler == nil {
		outcome.ErrorCode = "invalid_configuration"
		outcome.FinishedAt = now().UTC()
		return outcome
	}
	if guard, ok := e.Registry.(interface {
		CheckTerritorialExecution(context.Context, string, int) error
	}); ok {
		if err := guard.CheckTerritorialExecution(ctx, sourceID, revision); err != nil {
			outcome.ErrorCode = "territory_unavailable"
			outcome.FinishedAt = now().UTC()
			return outcome
		}
	}
	v, err := e.Registry.Version(ctx, sourceID, revision)
	if err != nil || validateAcquisitionConfiguration(v.Configuration) != nil || (v.Configuration.RegionalProduct != nil && e.Resources == nil) {
		outcome.ErrorCode = "invalid_configuration"
		outcome.FinishedAt = now().UTC()
		return outcome
	}
	var report Preview
	var state acquisitionState
	if v.Configuration.RegionalProduct != nil {
		report, state, err = e.acquireRegional(ctx, sourceID, revision, v.Configuration)
	} else if e.Tracking != nil {
		report, state, err = e.scheduledAcquire(ctx, sourceID, revision, v.Configuration)
	} else {
		report, state, err = e.acquire(ctx, sourceID, revision, v.Configuration)
	}
	outcome.FinishedAt = now().UTC()
	outcome.Reachable = state.reachable
	outcome.ContentRecognized = state.contentRecognized
	outcome.ErrorCode = state.errorCode
	outcome.RetryAfter = state.retryAfter
	outcome.Listings = listingCount(report.Sections)
	outcome.Documents = len(report.Documents)
	outcome.NewDocuments = state.newDocuments
	outcome.PublicationObservedAt = state.publicationObservedAt
	outcome.Retained = append([]RetainedPage(nil), report.Documents...)
	for _, document := range report.Documents {
		if document.Changed {
			outcome.Revisions++
		}
	}
	if err == nil {
		outcome.Complete = true
		outcome.ErrorCode = ""
	}
	if err != nil && outcome.ErrorCode == "" {
		outcome.ErrorCode = "acquisition_failed"
	}
	return outcome
}

func (e *Engine) acquire(ctx context.Context, sourceID string, revision int, cfg registry.Configuration) (Preview, acquisitionState, error) {
	if cfg.RegionalProduct != nil {
		return e.acquireRegional(ctx, sourceID, revision, cfg)
	}
	now := time.Now
	if e.Now != nil {
		now = e.Now
	}
	report := Preview{SourceID: sourceID, Configuration: revision, ObservedAt: now().UTC()}
	state := acquisitionState{contentRecognized: true}
	discovered := map[string]*time.Time{}
	for _, section := range cfg.Sections {
		part, found, err := e.traverseSection(ctx, sourceID, revision, cfg, section, &state)
		if err != nil {
			return report, state, err
		}
		report.Sections = append(report.Sections, part)
		for _, item := range found {
			if previous, exists := discovered[item.URL]; !exists || previous == nil {
				discovered[item.URL] = item.PublicationDate
			}
			if len(discovered) > cfg.Discovery.MaxDocuments {
				state.errorCode = "document_limit_reached"
				return report, state, fmt.Errorf("%w: document limit reached", ErrInvalidConfiguration)
			}
		}
	}
	documentURLs := make([]string, 0, len(discovered))
	for item := range discovered {
		documentURLs = append(documentURLs, item)
	}
	slices.Sort(documentURLs)
	for _, item := range documentURLs {
		page, err := e.Crawler.Crawl(ctx, item)
		observeCrawl(&state, page, err)
		if err != nil {
			return report, state, fmt.Errorf("crawl discovered document: %w", err)
		}
		var retained RetainedPage
		if cfg.Attachments == nil {
			retained, err = e.retain(ctx, sourceID, revision, item, page.HTML)
		} else {
			retained, err = e.retainDocument(ctx, sourceID, revision, PlannedDocument{URL: item, PublicationDate: discovered[item]}, page, report.ObservedAt, &state, cfg.Attachments, false)
		}
		if err != nil {
			if state.errorCode == "" {
				state.errorCode = "retention_failed"
			}
			return report, state, fmt.Errorf("retain discovered document: %w", err)
		}
		report.Documents = append(report.Documents, retained)
	}
	return report, state, nil
}

func (e *Engine) scheduledAcquire(ctx context.Context, sourceID string, revision int, cfg registry.Configuration) (Preview, acquisitionState, error) {
	now := time.Now
	if e.Now != nil {
		now = e.Now
	}
	observedAt := now().UTC()
	report := Preview{SourceID: sourceID, Configuration: revision, ObservedAt: observedAt}
	state := acquisitionState{contentRecognized: true}
	if !listingParserConfigured(cfg.Discovery) {
		state.errorCode = "listing_date_parser_missing"
		return report, state, ErrInvalidConfiguration
	}
	byURL := map[string]DiscoveredDocument{}
	for _, section := range cfg.Sections {
		part, found, err := e.traverseSection(ctx, sourceID, revision, cfg, section, &state)
		if err != nil {
			return report, state, err
		}
		report.Sections = append(report.Sections, part)
		for _, document := range found {
			previous, exists := byURL[document.URL]
			if !exists || previous.PublicationDate == nil {
				byURL[document.URL] = document
			}
		}
	}
	discovered := make([]DiscoveredDocument, 0, len(byURL))
	for _, document := range byURL {
		discovered = append(discovered, document)
	}
	slicesSortDiscovered(discovered)
	if len(discovered) > cfg.Discovery.MaxDocuments {
		state.errorCode = "document_limit_reached"
		return report, state, ErrInvalidConfiguration
	}
	discovery, err := e.Tracking.Remember(ctx, sourceID, revision, discovered, observedAt)
	if err != nil {
		state.errorCode = "tracking_failed"
		return report, state, err
	}
	state.newDocuments = discovery.NewDocuments
	state.publicationObservedAt = discovery.NewestPublication
	plan, err := e.Tracking.Plan(ctx, sourceID, revision, observedAt, cfg.Discovery.BootstrapDays)
	if err != nil {
		state.errorCode = "tracking_failed"
		return report, state, err
	}
	if len(plan.Documents) > cfg.Discovery.MaxDocuments {
		state.errorCode = "document_limit_reached"
		return report, state, ErrInvalidConfiguration
	}
	var unavailableErr error
	for _, target := range plan.Documents {
		if err := ctx.Err(); err != nil {
			return report, state, err
		}
		if target.NextAttemptAt != nil && observedAt.Before(*target.NextAttemptAt) {
			state.errorCode = "source_http_error"
			state.contentRecognized = false
			unavailableErr = ErrCrawlUnavailable
			continue
		}
		retained, err := e.retainPlanned(ctx, sourceID, revision, target, observedAt, &state, cfg.Attachments)
		if retained.VersionID > 0 {
			retained.Workload = "ordinary"
			if plan.Bootstrap {
				retained.Workload = "bootstrap"
			}
			report.Documents = append(report.Documents, retained)
		}
		if err != nil {
			var failure *CrawlFailure
			if errors.As(err, &failure) && failure.Code == "source_http_error" && (failure.StatusCode == 404 || failure.StatusCode == 410) {
				if recordErr := e.Tracking.RecordUnavailable(ctx, sourceID, revision, target.URL, failure.StatusCode, now().UTC()); recordErr != nil {
					state.errorCode = "tracking_failed"
					return report, state, recordErr
				}
				unavailableErr = err
				continue
			}
			// Rate limits, service failures and persistence errors still stop
			// the source: they are not evidence of a single removed document.
			return report, state, err
		}
		if target.UnavailableStatus != 0 {
			if err := e.Tracking.ClearUnavailable(ctx, sourceID, revision, target.URL, now().UTC()); err != nil {
				state.errorCode = "tracking_failed"
				return report, state, err
			}
		}
	}
	if unavailableErr != nil {
		return report, state, unavailableErr
	}
	if plan.Bootstrap {
		if err := e.Tracking.CompleteBootstrap(ctx, sourceID, revision, observedAt); err != nil {
			state.errorCode = "tracking_failed"
			return report, state, err
		}
	}
	return report, state, nil
}

func (e *Engine) retainPlanned(ctx context.Context, source string, revision int, target PlannedDocument, checkedAt time.Time, state *acquisitionState, policy *registry.AttachmentPolicy) (RetainedPage, error) {
	page, err := e.Crawler.Crawl(ctx, target.URL)
	observeCrawl(state, page, err)
	if err != nil {
		return RetainedPage{}, fmt.Errorf("crawl tracked document: %w", err)
	}
	return e.retainDocument(ctx, source, revision, target, page, checkedAt, state, policy, true)
}

func (e *Engine) retainDocument(ctx context.Context, source string, revision int, target PlannedDocument, page Page, checkedAt time.Time, state *acquisitionState, policy *registry.AttachmentPolicy, track bool) (RetainedPage, error) {
	mediaType := page.MediaType
	if mediaType == "" {
		mediaType = "text/html; charset=utf-8"
	}
	resources := []documents.Resource{{URL: target.URL, Role: "original", Required: true, SourceID: source, Configuration: revision, MediaType: mediaType, Bytes: page.HTML}}
	requiredAttachmentMissing := false
	requiredInvalidPDF := false
	// Discover document-linked PDFs on every check so newly added attachments
	// participate in the version hash even when the parent URL is unchanged.
	seenResources := map[string]bool{target.URL: true}
	for _, dependency := range target.Resources {
		seenResources[dependency.URL] = true
	}
	contentClass := ""
	if policy != nil {
		contentClass = policy.ContentClass
	}
	links, err := documentPDFLinks(target.URL, page.HTML, contentClass)
	if err != nil {
		state.errorCode = "attachment_content_unrecognized"
		state.contentRecognized = false
		return RetainedPage{}, err
	}
	for _, attachment := range links {
		if !seenResources[attachment] {
			target.Resources = append(target.Resources, PlannedResource{URL: attachment, Required: true})
			seenResources[attachment] = true
		}
	}
	if len(target.Resources) > 255 {
		state.errorCode = "attachment_limit_reached"
		return RetainedPage{}, ErrInvalidConfiguration
	}
	for _, dependency := range target.Resources {
		if !policy.Allows(target.URL, dependency.URL) {
			resources = append(resources, documents.Resource{URL: dependency.URL, Role: "attachment", Required: dependency.Required, SourceID: source, Configuration: revision, Missing: "forbidden"})
			requiredAttachmentMissing = requiredAttachmentMissing || dependency.Required
			continue
		}
		fetcher := e.Resources
		if fetcher == nil {
			fetcher = e.Crawler
		}
		var resourcePage Page
		var crawlErr error
		parentURL, _ := url.Parse(target.URL)
		resourceURL, _ := url.Parse(dependency.URL)
		if !sameHTTPOrigin(parentURL, resourceURL) {
			bounded, ok := fetcher.(interface {
				CrawlBounded(context.Context, string, func(string) bool) (Page, error)
			})
			if !ok {
				crawlErr = ErrInvalidConfiguration
			} else {
				resourcePage, crawlErr = bounded.CrawlBounded(ctx, dependency.URL, func(raw string) bool { return policy.Allows(target.URL, raw) })
			}
		} else {
			resourcePage, crawlErr = fetcher.Crawl(ctx, dependency.URL)
		}
		if crawlErr != nil {
			if dependency.Required {
				observeCrawl(state, resourcePage, crawlErr)
				requiredAttachmentMissing = true
			}
			resources = append(resources, documents.Resource{URL: dependency.URL, Role: "attachment", Required: dependency.Required, SourceID: source, Configuration: revision, Missing: "unavailable"})
			continue
		}
		if policy != nil && policy.ValidatePDF {
			validate := e.validatePDF
			if validate == nil {
				validate = validateAttachmentPDF
			}
			if validate(ctx, resourcePage) != nil {
				resources = append(resources, documents.Resource{URL: dependency.URL, Role: "attachment", Required: dependency.Required, SourceID: source, Configuration: revision, Missing: "unavailable"})
				requiredAttachmentMissing = requiredAttachmentMissing || dependency.Required
				requiredInvalidPDF = requiredInvalidPDF || dependency.Required
				continue
			}
		}
		observeCrawl(state, resourcePage, nil)
		resourceType := resourcePage.MediaType
		if resourceType == "" {
			resourceType = "application/octet-stream"
		}
		resources = append(resources, documents.Resource{URL: dependency.URL, Role: "attachment", Required: dependency.Required, SourceID: source, Configuration: revision, MediaType: resourceType, Bytes: resourcePage.HTML})
	}
	metadata := map[string]any{}
	if target.PublicationDate != nil {
		metadata["source_publication_date"] = target.PublicationDate.UTC().Format("2006-01-02")
	}
	metadataJSON, _ := json.Marshal(metadata)
	hash := sha256.New()
	_, _ = hash.Write([]byte(target.URL + "\x00"))
	_, _ = hash.Write(metadataJSON)
	for _, resource := range resources {
		_, _ = hash.Write([]byte("\x00" + resource.URL + "\x00" + resource.Missing + "\x00"))
		_, _ = hash.Write(resource.Bytes)
	}
	id := "check-" + source + "-" + fmt.Sprint(revision) + "-" + hex.EncodeToString(hash.Sum(nil))
	version, err := e.Retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: source, Configuration: revision, URL: target.URL, Metadata: metadataJSON, Resources: resources})
	if err != nil {
		state.errorCode = "retention_failed"
		return RetainedPage{}, err
	}
	changed := false
	if track {
		changed, err = e.Tracking.RecordVersion(ctx, source, target.URL, version.ID, checkedAt)
		if err != nil {
			state.errorCode = "tracking_failed"
			return RetainedPage{}, err
		}
	}
	retained := RetainedPage{InferenceResources: version.Resources, URL: target.URL, VersionID: version.ID, Hash: version.Hash, Changed: changed}
	if requiredAttachmentMissing {
		state.errorCode = "required_attachment_unavailable"
		if requiredInvalidPDF {
			state.errorCode = "invalid_attachment_pdf"
		}
		return retained, ErrRequiredAttachment
	}
	return retained, nil
}

// Only explicit PDF links are attachments; images, navigation and arbitrary
// external pages are not newly discovered collection channels. Out-of-origin
// PDF links are returned for a visible missing reference, never fetched.
func linkedPDFs(parent string, body []byte) []string {
	links, _ := documentPDFLinks(parent, body, "")
	return links
}

func documentPDFLinks(parent string, body []byte, contentClass string) ([]string, error) {
	base, err := url.Parse(parent)
	if err != nil {
		return nil, ErrInvalidConfiguration
	}
	root, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, ErrUnrecognizedContent
	}
	seen := map[string]bool{}
	contentFound := contentClass == ""
	var result []string
	walkElements(root, func(node *html.Node) {
		if contentClass != "" && hasClass(node, contentClass) {
			contentFound = true
		}
		if node.Data != "a" {
			return
		}
		if contentClass != "" {
			within := false
			for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
				if ancestor.Data == "footer" || ancestor.Data == "nav" || ancestor.Data == "header" {
					return
				}
				within = within || hasClass(ancestor, contentClass)
			}
			if !within {
				return
			}
		}
		for _, attribute := range node.Attr {
			if attribute.Key != "href" {
				continue
			}
			ref, err := url.Parse(strings.TrimSpace(attribute.Val))
			if err != nil {
				continue
			}
			resolved := base.ResolveReference(ref)
			if (resolved.Scheme != "http" && resolved.Scheme != "https") || resolved.Host == "" || resolved.User != nil || !strings.HasSuffix(strings.ToLower(resolved.Path), ".pdf") {
				continue
			}
			resolved.Fragment = ""
			value := resolved.String()
			if !seen[value] {
				seen[value] = true
				result = append(result, value)
			}
		}
	})
	slices.Sort(result)
	if !contentFound {
		return nil, ErrUnrecognizedContent
	}
	return result, nil
}

func (e *Engine) traverseSection(ctx context.Context, sourceID string, revision int, cfg registry.Configuration, section string, state *acquisitionState) (SectionPreview, []DiscoveredDocument, error) {
	part := SectionPreview{SectionURL: section}
	queue := []string{section}
	seen := map[string]bool{}
	documentsFound := map[string]*time.Time{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if seen[current] {
			continue
		}
		if len(seen) >= cfg.Discovery.MaxPagesPerSection {
			state.errorCode = "page_limit_reached"
			return part, nil, fmt.Errorf("%w: section page limit reached", ErrInvalidConfiguration)
		}
		seen[current] = true
		page, err := e.Crawler.Crawl(ctx, current)
		observeCrawl(state, page, err)
		if err != nil {
			return part, nil, fmt.Errorf("crawl listing: %w", err)
		}
		if !recognizesListing(page.HTML, cfg.Discovery.ListingContentMarkers) {
			state.contentRecognized = false
			state.errorCode = "unrecognized_content"
			return part, nil, ErrUnrecognizedContent
		}
		retained, err := e.retain(ctx, sourceID, revision, current, page.HTML)
		if err != nil {
			state.errorCode = "retention_failed"
			return part, nil, fmt.Errorf("retain listing: %w", err)
		}
		part.Listings = append(part.Listings, retained)
		for _, document := range discoverDocuments(current, page.HTML, cfg, page.Links) {
			if previous, exists := documentsFound[document.URL]; !exists || previous == nil {
				documentsFound[document.URL] = document.PublicationDate
			}
		}
		for _, link := range page.Links {
			canonical, ok := boundedURL(current, link.URL)
			if !ok {
				continue
			}
			if pagination, ok := paginationURL(section, canonical, cfg.Discovery.PaginationParameter); ok && !seen[pagination] {
				queue = append(queue, pagination)
			}
		}
	}
	items := make([]DiscoveredDocument, 0, len(documentsFound))
	for item, published := range documentsFound {
		items = append(items, DiscoveredDocument{URL: item, PublicationDate: published})
	}
	slicesSortDiscovered(items)
	return part, items, nil
}

func (e *Engine) retain(ctx context.Context, source string, revision int, target string, original []byte) (RetainedPage, error) {
	h := sha256.Sum256(append([]byte(target+"\x00"), original...))
	id := "preview-" + source + "-" + fmt.Sprint(revision) + "-" + hex.EncodeToString(h[:])
	v, err := e.Retained.Retain(ctx, documents.Acquisition{
		ID: id, SourceID: source, Configuration: revision, URL: target, Metadata: json.RawMessage(`{}`),
		Resources: []documents.Resource{{URL: target, Role: "original", Required: true, SourceID: source, Configuration: revision, MediaType: "text/html; charset=utf-8", Bytes: original}},
	})
	return RetainedPage{InferenceResources: v.Resources, URL: target, VersionID: v.ID, Hash: v.Hash, Workload: "ordinary"}, err
}

func validateDiscovery(cfg registry.Configuration) error {
	if !cfg.Policy.CollectionPermitted || !cfg.Policy.RetentionPermitted || len(cfg.Sections) == 0 || cfg.Discovery.MaxPagesPerSection < 1 || cfg.Discovery.MaxPagesPerSection > 1000 || cfg.Discovery.MaxDocuments < 1 || cfg.Discovery.MaxDocuments > 10000 || strings.TrimSpace(cfg.Discovery.PaginationParameter) == "" || len(cfg.Discovery.DocumentPathPrefixes) == 0 || len(cfg.Discovery.ListingContentMarkers) == 0 {
		return ErrInvalidConfiguration
	}
	for _, prefix := range cfg.Discovery.DocumentPathPrefixes {
		if !strings.HasPrefix(prefix, "/") || strings.ContainsAny(prefix, "?#") {
			return ErrInvalidConfiguration
		}
	}
	for _, marker := range cfg.Discovery.ListingContentMarkers {
		if strings.TrimSpace(marker) == "" {
			return ErrInvalidConfiguration
		}
	}
	return nil
}

func validateAcquisitionConfiguration(cfg registry.Configuration) error {
	if !cfg.Attachments.Valid() || cfg.Attachments != nil && cfg.RegionalProduct != nil {
		return ErrInvalidConfiguration
	}
	if cfg.RegionalProduct != nil {
		if cfg.AccessMethod != "crawl4ai-html-pdf" {
			return ErrInvalidConfiguration
		}
		return validateRegionalConfiguration(cfg)
	}
	if cfg.AccessMethod != "crawl4ai" {
		return ErrInvalidConfiguration
	}
	return validateDiscovery(cfg)
}

func observeCrawl(state *acquisitionState, page Page, err error) {
	if page.StatusCode > 0 {
		state.reachable = true
	}
	if err == nil {
		return
	}
	state.contentRecognized = false
	var failure *CrawlFailure
	if errors.As(err, &failure) {
		state.reachable = state.reachable || failure.Reachable
		state.errorCode = failure.Code
		state.retryAfter = failure.RetryAfter
		return
	}
	state.errorCode = "crawl_failed"
}

func recognizesListing(body []byte, markers []string) bool {
	if len(body) == 0 {
		return false
	}
	text := string(body)
	for _, marker := range markers {
		if !strings.Contains(text, marker) {
			return false
		}
	}
	return true
}

func boundedURL(baseRaw, linkRaw string) (string, bool) {
	base, err := url.Parse(baseRaw)
	if err != nil {
		return "", false
	}
	link, err := url.Parse(strings.TrimSpace(linkRaw))
	if err != nil {
		return "", false
	}
	resolved := base.ResolveReference(link)
	if resolved.Scheme != base.Scheme || !strings.EqualFold(resolved.Host, base.Host) || resolved.User != nil || resolved.Fragment != "" {
		return "", false
	}
	resolved.Fragment = ""
	return resolved.String(), true
}

func isDocument(cfg registry.Configuration, raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.RawQuery != "" {
		return false
	}
	for _, prefix := range cfg.Discovery.DocumentPathPrefixes {
		if strings.HasPrefix(u.EscapedPath(), prefix) && len(u.EscapedPath()) > len(prefix) {
			return true
		}
	}
	return false
}

func paginationURL(sectionRaw, candidateRaw, parameter string) (string, bool) {
	section, a := url.Parse(sectionRaw)
	candidate, b := url.Parse(candidateRaw)
	if a != nil || b != nil || section.Scheme != candidate.Scheme || !strings.EqualFold(section.Host, candidate.Host) || section.EscapedPath() != candidate.EscapedPath() {
		return "", false
	}
	query := candidate.Query()
	if len(query) != 1 || len(query[parameter]) != 1 || strings.TrimSpace(query.Get(parameter)) == "" {
		return "", false
	}
	// Drupal exposes the first configured listing both without a query and as
	// page=0. Treat them as one traversal node and preserve the configured URL.
	if section.RawQuery == "" && query.Get(parameter) == "0" {
		return section.String(), true
	}
	return candidate.String(), true
}

func listingCount(sections []SectionPreview) int {
	total := 0
	for _, section := range sections {
		total += len(section.Listings)
	}
	return total
}
