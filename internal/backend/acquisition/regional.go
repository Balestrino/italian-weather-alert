package acquisition

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"golang.org/x/net/html"
)

var (
	issuancePattern           = regexp.MustCompile(`(?i)Emissione(?: interna)? di\s+(.+?),\s*ore\s+([0-9]{1,2}[.:][0-9]{2})`)
	monitoringIssuancePattern = regexp.MustCompile(`(?i)Emesso il\s+([0-9]{2}/[0-9]{2}/[0-9]{4})\s+([0-9]{1,2}[.:][0-9]{2})`)
	regionalDatePattern       = regexp.MustCompile(`(?i)\b(?:Luned[iìí]|Marted[iìí]|Mercoled[iìí]|Gioved[iìí]|Venerd[iìí]|Sabato|Domenica),?\s+[0-9]{1,2}\s+[A-Za-z]+\s+[0-9]{4}\b`)
	zonePattern               = regexp.MustCompile(`\b(?:A[1-6]|B|C|E[1-3]|F[1-2]|I|L|M|O[1-3]|R[1-2]|S[1-3]|T|V)\b`)
)

var criticalityRiskLabels = []string{
	"RISCHIO IDROGEOLOGICO-IDRAULICO RETICOLO MINORE",
	"RISCHIO IDRAULICO RETICOLO PRINCIPALE",
	"TEMPORALI FORTI CON RISCHIO IDROGEOLOGICO-IDRAULICO RETICOLO MINORE",
	"RISCHIO VENTO",
	"RISCHIO MAREGGIATE",
	"RISCHIO NEVE",
	"RISCHIO GHIACCIO",
}

// RegionalObservation contains only facts that can be read from the retained
// HTML structure. Graphical colors are deliberately not inferred from OCR or
// surrounding prose; map bytes remain evidence for later visual extraction.
type RegionalObservation struct {
	Product               string   `json:"product"`
	IssuanceExpression    string   `json:"issuance_expression,omitempty"`
	ValidityExpressions   []string `json:"validity_expressions,omitempty"`
	AffectedZones         []string `json:"affected_zones,omitempty"`
	RisksOrPhenomena      []string `json:"risks_or_phenomena,omitempty"`
	ExplicitLevels        []string `json:"explicit_levels,omitempty"`
	GraphicalResources    []string `json:"graphical_resources,omitempty"`
	TextSaysNoCriticality bool     `json:"text_says_no_criticality,omitempty"`
	MonitoringNoEvent     bool     `json:"monitoring_no_event,omitempty"`
	LevelSource           string   `json:"level_source"`
}

// ObserveRegionalHTML applies the same bounded regional-product parser used by
// acquisition to a retained evaluation input. It deliberately exposes no
// transport or activation behavior: callers must supply already retained bytes
// and the exact graphical-resource identities established by acquisition.
func ObserveRegionalHTML(product string, body []byte, graphics []string) (RegionalObservation, error) {
	return observeRegionalHTML(product, body, graphics)
}

func validateRegionalConfiguration(cfg registry.Configuration) error {
	contract := cfg.RegionalProduct
	if contract == nil || len(cfg.Sections) != 1 || cfg.Sections[0] != cfg.URL || !cfg.Policy.CollectionPermitted || !cfg.Policy.RetentionPermitted {
		return ErrInvalidConfiguration
	}
	if contract.Kind != "vigilance" && contract.Kind != "criticality" && contract.Kind != "monitoring" {
		return ErrInvalidConfiguration
	}
	if contract.Kind == "monitoring" && cfg.ExpectedPublication != nil {
		return ErrInvalidConfiguration
	}
	if len(contract.ContentMarkers) == 0 || contract.MinimumGraphicalAssets < 0 {
		return ErrInvalidConfiguration
	}
	if (contract.Kind == "vigilance" || contract.Kind == "criticality") && contract.MinimumGraphicalAssets == 0 && !contract.PrintContainsGraphics {
		return ErrInvalidConfiguration
	}
	base, err := url.Parse(cfg.URL)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
		return ErrInvalidConfiguration
	}
	if contract.PrintURL != "" {
		printURL, parseErr := url.Parse(contract.PrintURL)
		if parseErr != nil || printURL.Scheme != base.Scheme || !strings.EqualFold(printURL.Host, base.Host) || printURL.User != nil || printURL.Fragment != "" {
			return ErrInvalidConfiguration
		}
	}
	for _, marker := range contract.ContentMarkers {
		if strings.TrimSpace(marker) == "" {
			return ErrInvalidConfiguration
		}
	}
	for _, prefix := range contract.ResourcePathPrefixes {
		if !strings.HasPrefix(prefix, "/") || strings.ContainsAny(prefix, "?#") {
			return ErrInvalidConfiguration
		}
	}
	return nil
}

func (e *Engine) acquireRegional(ctx context.Context, sourceID string, revision int, cfg registry.Configuration) (Preview, acquisitionState, error) {
	now := time.Now
	if e.Now != nil {
		now = e.Now
	}
	report := Preview{SourceID: sourceID, Configuration: revision, ObservedAt: now().UTC()}
	state := acquisitionState{contentRecognized: true}
	page, err := e.Crawler.Crawl(ctx, cfg.URL)
	observeCrawl(&state, page, err)
	if err != nil {
		return report, state, fmt.Errorf("crawl regional product: %w", err)
	}
	if !recognizesListing(page.HTML, cfg.RegionalProduct.ContentMarkers) {
		state.contentRecognized = false
		state.errorCode = "unrecognized_content"
		return report, state, ErrUnrecognizedContent
	}

	resourceURLs, err := regionalResourceURLs(cfg.URL, page.HTML, *cfg.RegionalProduct)
	if err != nil {
		state.errorCode = "invalid_configuration"
		return report, state, err
	}
	resources := []documents.Resource{{
		URL: cfg.URL, Role: "original", Required: true, SourceID: sourceID, Configuration: revision,
		MediaType: mediaTypeOr(page.MediaType, "text/html; charset=utf-8"), Bytes: page.HTML,
	}}
	graphics := make([]string, 0, len(resourceURLs))
	for _, target := range resourceURLs {
		resourcePage, crawlErr := e.Resources.Crawl(ctx, target.URL)
		observeCrawl(&state, resourcePage, crawlErr)
		if crawlErr != nil {
			return report, state, fmt.Errorf("crawl required regional %s: %w", target.Role, crawlErr)
		}
		resources = append(resources, documents.Resource{
			URL: target.URL, Role: "resource", Required: true, SourceID: sourceID, Configuration: revision,
			MediaType: mediaTypeOr(resourcePage.MediaType, target.DefaultMediaType), Bytes: resourcePage.HTML,
		})
		if target.Role == "graphic" {
			graphics = append(graphics, target.URL)
		}
	}
	observation, err := observeRegionalHTML(cfg.RegionalProduct.Kind, page.HTML, graphics)
	if err != nil {
		state.contentRecognized = false
		state.errorCode = "regional_extraction_failed"
		return report, state, err
	}
	metadata, _ := json.Marshal(observation)
	state.publicationObservedAt = regionalIssuanceInstant(observation.IssuanceExpression)
	hash := sha256.New()
	_, _ = hash.Write([]byte(cfg.URL))
	for _, resource := range resources {
		_, _ = hash.Write([]byte("\x00" + resource.URL + "\x00"))
		_, _ = hash.Write(resource.Bytes)
	}
	id := "regional-" + sourceID + "-" + fmt.Sprint(revision) + "-" + hex.EncodeToString(hash.Sum(nil))
	version, err := e.Retained.Retain(ctx, documents.Acquisition{
		ID: id, SourceID: sourceID, Configuration: revision, URL: cfg.URL, Metadata: metadata, Resources: resources,
	})
	if err != nil {
		state.errorCode = "retention_failed"
		return report, state, err
	}
	changed := false
	if e.Tracking != nil {
		// Regional endpoints have no listing traversal to register their target.
		// Register the stable product URL before recording its first version.
		discovery, rememberErr := e.Tracking.Remember(ctx, sourceID, revision, []DiscoveredDocument{{URL: cfg.URL}}, report.ObservedAt)
		if rememberErr != nil {
			state.errorCode = "tracking_failed"
			return report, state, rememberErr
		}
		state.newDocuments = discovery.NewDocuments
		changed, err = e.Tracking.RecordVersion(ctx, sourceID, cfg.URL, version.ID, report.ObservedAt)
		if err != nil {
			state.errorCode = "tracking_failed"
			return report, state, err
		}
	}
	report.Sections = []SectionPreview{{SectionURL: cfg.URL}}
	report.Documents = []RetainedPage{{InferenceResources: version.Resources, URL: cfg.URL, VersionID: version.ID, Hash: version.Hash, Changed: changed, Workload: "ordinary"}}
	return report, state, nil
}

// A product's edition is a publication observation, including on unchanged
// checks. An undated no-event state has no publication instant.
func regionalIssuanceInstant(expression string) *time.Time {
	parts := strings.Split(expression, ", ore ")
	if len(parts) != 2 {
		return nil
	}
	day, err := parseRegionalDay(parts[0])
	if err != nil {
		parsed, parseErr := time.Parse("02/01/2006", parts[0])
		if parseErr != nil {
			return nil
		}
		day = &parsed
	}
	clock, err := time.Parse("15:04", strings.ReplaceAll(parts[1], ".", ":"))
	if err != nil {
		return nil
	}
	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		return nil
	}
	instant := time.Date(day.Year(), day.Month(), day.Day(), clock.Hour(), clock.Minute(), 0, 0, loc).UTC()
	return &instant
}

type regionalResource struct {
	URL              string
	Role             string
	DefaultMediaType string
}

func regionalResourceURLs(pageURL string, body []byte, contract registry.RegionalProductContract) ([]regionalResource, error) {
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil, ErrInvalidConfiguration
	}
	seen := map[string]bool{}
	result := []regionalResource{}
	if contract.PrintURL != "" {
		seen[contract.PrintURL] = true
		result = append(result, regionalResource{URL: contract.PrintURL, Role: "print", DefaultMediaType: "application/pdf"})
	}
	root, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, ErrUnrecognizedContent
	}
	walkElements(root, func(node *html.Node) {
		if node.Data != "img" {
			return
		}
		for _, attribute := range node.Attr {
			if attribute.Key != "src" {
				continue
			}
			reference, parseErr := url.Parse(strings.TrimSpace(attribute.Val))
			if parseErr != nil {
				continue
			}
			resolved := base.ResolveReference(reference)
			if resolved.Scheme != base.Scheme || !strings.EqualFold(resolved.Host, base.Host) || resolved.User != nil || resolved.Fragment != "" || !hasAllowedPathPrefix(resolved.Path, contract.ResourcePathPrefixes) {
				continue
			}
			canonical := resolved.String()
			if !seen[canonical] {
				seen[canonical] = true
				result = append(result, regionalResource{URL: canonical, Role: "graphic", DefaultMediaType: "image/png"})
			}
		}
	})
	graphics := 0
	for _, item := range result {
		if item.Role == "graphic" {
			graphics++
		}
	}
	if graphics < contract.MinimumGraphicalAssets {
		return nil, fmt.Errorf("%w: expected at least %d graphical resources, found %d", ErrUnrecognizedContent, contract.MinimumGraphicalAssets, graphics)
	}
	slices.SortFunc(result, func(a, b regionalResource) int { return strings.Compare(a.URL, b.URL) })
	return result, nil
}

func hasAllowedPathPrefix(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func mediaTypeOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func observeRegionalHTML(product string, body []byte, graphics []string) (RegionalObservation, error) {
	root, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return RegionalObservation{}, ErrUnrecognizedContent
	}
	text := strings.Join(strings.Fields(nodeText(root)), " ")
	observation := RegionalObservation{Product: product, GraphicalResources: slices.Clone(graphics), LevelSource: "explicit HTML table only"}
	if match := issuancePattern.FindStringSubmatch(text); len(match) == 3 {
		observation.IssuanceExpression = strings.TrimSpace(match[1]) + ", ore " + match[2]
	}
	if product == "monitoring" && observation.IssuanceExpression == "" {
		if match := monitoringIssuancePattern.FindStringSubmatch(text); len(match) == 3 {
			observation.IssuanceExpression = match[1] + ", ore " + match[2]
		}
	}
	rows := htmlTableRows(root)
	switch product {
	case "vigilance":
		observeVigilanceTables(rows, &observation)
	case "criticality":
		observeCriticalityTables(rows, &observation)
		observeCriticalityText(text, &observation)
		observation.TextSaysNoCriticality = strings.Contains(strings.ToUpper(text), "CRITICITÀ PREVISTE: NESSUNA") || strings.Contains(strings.ToUpper(text), "CRITICITA PREVISTE: NESSUNA")
	case "monitoring":
		// Monitoring narrative may quote alert colors. It is never a new level
		// unless a separate criticality product explicitly issues that level.
		observation.LevelSource = "none: monitoring cannot issue an inferred level"
		// The event-driven endpoint explicitly publishes this undated state when
		// there is no current bulletin. A missing/empty page is not equivalent.
		if observation.IssuanceExpression == "" {
			walkElements(root, func(node *html.Node) {
				if node.Data == "b" && strings.Join(strings.Fields(nodeText(node)), " ") == "NESSUN AVVISO IN CORSO DI VALIDITÀ O EVENTO IN CORSO" {
					observation.MonitoringNoEvent = true
				}
			})
		}
	default:
		return RegionalObservation{}, ErrInvalidConfiguration
	}
	observation.AffectedZones = uniqueSorted(observation.AffectedZones)
	observation.RisksOrPhenomena = uniqueSorted(observation.RisksOrPhenomena)
	observation.ExplicitLevels = uniqueSorted(observation.ExplicitLevels)
	observation.ValidityExpressions = uniqueSorted(observation.ValidityExpressions)
	if observation.IssuanceExpression == "" && !observation.MonitoringNoEvent {
		return RegionalObservation{}, fmt.Errorf("%w: missing issuance expression", ErrUnrecognizedContent)
	}
	return observation, nil
}

func observeCriticalityText(text string, observation *RegionalObservation) {
	if len(observation.RisksOrPhenomena) == 0 {
		upper := strings.ToUpper(text)
		for _, label := range criticalityRiskLabels {
			if strings.Contains(upper, label) {
				observation.RisksOrPhenomena = append(observation.RisksOrPhenomena, label)
			}
		}
	}
	// A green bulletin has no explicit criticality rows, but its dated map pairs
	// still define the product's day precision. Precise table intervals win when
	// present; map dates never become level or affected-zone evidence.
	if len(observation.ValidityExpressions) == 0 {
		observation.ValidityExpressions = append(observation.ValidityExpressions, regionalDatePattern.FindAllString(text, -1)...)
	}
}

func htmlTableRows(root *html.Node) [][][]string {
	var tables [][][]string
	walkElements(root, func(node *html.Node) {
		if node.Data != "table" {
			return
		}
		var rows [][]string
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			collectDirectRows(child, &rows)
		}
		if len(rows) > 0 {
			tables = append(tables, rows)
		}
	})
	return tables
}

func collectDirectRows(node *html.Node, rows *[][]string) {
	if node.Type == html.ElementNode && node.Data == "table" {
		return
	}
	if node.Type == html.ElementNode && node.Data == "tr" {
		var cells []string
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && (child.Data == "td" || child.Data == "th") {
				cells = append(cells, strings.Join(strings.Fields(nodeText(child)), " "))
			}
		}
		if len(cells) > 0 {
			*rows = append(*rows, cells)
		}
		return
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		collectDirectRows(child, rows)
	}
}

func observeVigilanceTables(tables [][][]string, observation *RegionalObservation) {
	phenomena := map[string]bool{"pioggia": true, "temporali": true, "vento": true, "mare": true, "neve": true, "ghiaccio": true}
	for _, rows := range tables {
		matched := 0
		for _, row := range rows {
			if len(row) > 0 && phenomena[strings.ToLower(row[0])] {
				matched++
			}
		}
		if matched < 4 {
			continue
		}
		if len(rows[0]) > 1 {
			observation.ValidityExpressions = append(observation.ValidityExpressions, rows[0][1:]...)
		}
		for _, row := range rows[1:] {
			if len(row) < 2 || !phenomena[strings.ToLower(row[0])] {
				continue
			}
			observation.RisksOrPhenomena = append(observation.RisksOrPhenomena, row[0])
			for _, cell := range row[1:] {
				observation.AffectedZones = append(observation.AffectedZones, zonePattern.FindAllString(cell, -1)...)
			}
		}
		return
	}
}

func observeCriticalityTables(tables [][][]string, observation *RegionalObservation) {
	for _, rows := range tables {
		if len(rows) < 2 || len(rows[0]) < 4 {
			continue
		}
		header := strings.ToUpper(strings.Join(rows[0], " "))
		if !strings.Contains(header, "ZONE") || !strings.Contains(header, "RISCHIO") || !strings.Contains(header, "CRITIC") {
			continue
		}
		for _, row := range rows[1:] {
			if len(row) < 4 {
				continue
			}
			observation.AffectedZones = append(observation.AffectedZones, zonePattern.FindAllString(row[0], -1)...)
			observation.RisksOrPhenomena = append(observation.RisksOrPhenomena, row[1])
			observation.ValidityExpressions = append(observation.ValidityExpressions, row[2])
			level := strings.ToLower(strings.TrimSpace(row[3]))
			if level == "verde" || level == "giallo" || level == "arancione" || level == "rosso" {
				observation.ExplicitLevels = append(observation.ExplicitLevels, level)
			}
		}
		return
	}
}

func uniqueSorted(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	slices.Sort(result)
	return result
}
