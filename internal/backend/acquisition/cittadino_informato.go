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
	"strconv"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

type boundedCrawler interface {
	CrawlBounded(context.Context, string, func(string) bool) (Page, error)
}

type cittadinoNotice struct {
	ID          int64  `json:"id"`
	Publisher   string `json:"ente"`
	URL         string `json:"url"`
	Title       string `json:"titolo"`
	Content     string `json:"contenuto"`
	Published   string `json:"data_pubblicazione"`
	DisplayFrom string `json:"data_inizio"`
	DisplayTo   string `json:"data_fine"`
}

type cittadinoListing struct {
	Count *int              `json:"count"`
	Total *int              `json:"total_count"`
	Pages *int              `json:"pages"`
	Page  *int              `json:"page"`
	Posts []cittadinoNotice `json:"posts"`
}

func (e *Engine) acquireCittadinoInformato(ctx context.Context, source string, revision int, cfg registry.Configuration, track bool) (Preview, acquisitionState, error) {
	now := time.Now
	if e.Now != nil {
		now = e.Now
	}
	report := Preview{SourceID: source, Configuration: revision, ObservedAt: now().UTC()}
	state := acquisitionState{contentRecognized: true}
	contract := cfg.CittadinoInformato
	if contract == nil || !contract.Valid(cfg) {
		return report, state, ErrInvalidConfiguration
	}
	fetcher, ok := e.Resources.(boundedCrawler)
	if !ok {
		state.errorCode = "bounded_api_client_missing"
		return report, state, ErrInvalidConfiguration
	}
	fetch := func(raw string) (Page, error) {
		page, err := fetcher.CrawlBounded(ctx, raw, func(candidate string) bool { return candidate == raw })
		observeCrawl(&state, page, err)
		if err == nil && (page.URL != raw || page.MediaType != "application/json" || !json.Valid(page.HTML)) {
			state.errorCode = "platform_api_invalid_response"
			state.contentRecognized = false
			return page, ErrUnrecognizedContent
		}
		return page, err
	}
	invalid := func(code string) (Preview, acquisitionState, error) {
		state.contentRecognized = false
		state.errorCode = code
		return report, state, ErrUnrecognizedContent
	}
	identityURL := contract.APIBase() + "comune?nome=" + url.QueryEscape(contract.MunicipalitySlug)
	identityPage, err := fetch(identityURL)
	if err != nil {
		return report, state, err
	}
	var identity struct {
		ID   int64  `json:"id"`
		Slug string `json:"slug"`
	}
	if json.Unmarshal(identityPage.HTML, &identity) != nil || identity.ID < 1 || identity.Slug != contract.MunicipalitySlug {
		return invalid("platform_municipality_mismatch")
	}

	if contract.Updates {
		section := SectionPreview{SectionURL: contract.BaseURL() + "aggiornamenti/"}
		targets := []DiscoveredDocument{}
		seenIDs := map[int64]bool{}
		seenURLs := map[string]bool{}
		total, pages := -1, -1
		for number := 1; ; number++ {
			if number > cfg.Discovery.MaxPagesPerSection {
				return invalid("listing_limit_reached")
			}
			query := url.Values{"comune": {contract.MunicipalitySlug}, "ente": {contract.Publisher}, "per_page": {strconv.Itoa(contract.PageSize)}, "page": {strconv.Itoa(number)}}
			if contract.UpdatesSince != "" {
				date, _ := time.Parse("2006-01-02", contract.UpdatesSince)
				query.Set("data_inizio", date.Format("02/01/2006"))
			}
			listingURL := contract.APIBase() + "aggiornamenti?" + query.Encode()
			page, err := fetch(listingURL)
			if err != nil {
				return report, state, err
			}
			var listing cittadinoListing
			if json.Unmarshal(page.HTML, &listing) != nil || listing.Count == nil || listing.Total == nil || listing.Pages == nil || listing.Page == nil ||
				*listing.Count != len(listing.Posts) || *listing.Page != number || *listing.Total < 0 || *listing.Pages < 0 || len(listing.Posts) > contract.PageSize {
				return invalid("platform_listing_invalid")
			}
			if total == -1 {
				total, pages = *listing.Total, *listing.Pages
			}
			if total != *listing.Total || pages != *listing.Pages || pages != (total+contract.PageSize-1)/contract.PageSize {
				return invalid("platform_listing_changed")
			}
			if total > cfg.Discovery.MaxDocuments {
				return invalid("document_limit_reached")
			}
			if pages > cfg.Discovery.MaxPagesPerSection {
				return invalid("listing_limit_reached")
			}
			if total > 0 && len(listing.Posts) != min(contract.PageSize, total-(number-1)*contract.PageSize) {
				return invalid("platform_listing_incomplete")
			}
			for _, notice := range listing.Posts {
				if !validCittadinoNotice(contract, notice, false) || seenIDs[notice.ID] || seenURLs[notice.URL] {
					return invalid("platform_notice_scope_mismatch")
				}
				seenIDs[notice.ID], seenURLs[notice.URL] = true, true
				targets = append(targets, DiscoveredDocument{URL: cittadinoDetailURL(contract, notice.ID)})
			}
			retained, err := e.retainCittadino(ctx, source, revision, listingURL, page, map[string]any{"channel_role": "platform", "kind": "listing", "municipality_istat": contract.MunicipalityISTAT}, nil, &state, contract, false, report.ObservedAt)
			if err != nil {
				return report, state, err
			}
			section.Listings = append(section.Listings, retained)
			if number >= pages {
				break
			}
		}
		if len(targets) != total {
			return invalid("platform_listing_incomplete")
		}
		report.Sections = append(report.Sections, section)
		plan := RevisionPlan{}
		for _, target := range targets {
			plan.Documents = append(plan.Documents, PlannedDocument{URL: target.URL})
		}
		if track {
			discovery, err := e.Tracking.Remember(ctx, source, revision, targets, report.ObservedAt)
			if err != nil {
				state.errorCode = "tracking_failed"
				return report, state, err
			}
			state.newDocuments = discovery.NewDocuments
			plan, err = e.Tracking.Plan(ctx, source, revision, report.ObservedAt, cfg.Discovery.BootstrapDays)
			if err != nil {
				state.errorCode = "tracking_failed"
				return report, state, err
			}
		}
		if len(plan.Documents) > cfg.Discovery.MaxDocuments {
			return invalid("document_limit_reached")
		}
		var deferred error
		for _, target := range plan.Documents {
			if contract.Risks && (target.URL == contract.APIBase()+"rischi/oggi" || target.URL == contract.APIBase()+"rischi/domani") {
				continue // Rechecked below without a publication-date cutoff.
			}
			id, ok := cittadinoTargetID(contract, target.URL)
			if !ok {
				return invalid("platform_target_outside_scope")
			}
			if target.NextAttemptAt != nil && report.ObservedAt.Before(*target.NextAttemptAt) {
				state.errorCode = "source_http_error"
				state.contentRecognized = false
				deferred = ErrCrawlUnavailable
				continue
			}
			page, err := fetch(target.URL)
			if err != nil {
				var failure *CrawlFailure
				if track && errors.As(err, &failure) && (failure.StatusCode == 404 || failure.StatusCode == 410) {
					if err := e.Tracking.RecordUnavailable(ctx, source, revision, target.URL, failure.StatusCode, report.ObservedAt); err != nil {
						state.errorCode = "tracking_failed"
						return report, state, err
					}
					deferred = err
					continue
				}
				return report, state, err
			}
			var notice cittadinoNotice
			if json.Unmarshal(page.HTML, &notice) != nil || notice.ID != id || !validCittadinoNotice(contract, notice, true) {
				return invalid("platform_notice_scope_mismatch")
			}
			published := cittadinoDate(notice.Published)
			metadata := map[string]any{"channel_role": "platform", "kind": "notice", "verification_state": "pending", "municipality_istat": contract.MunicipalityISTAT,
				"platform_notice_id": id, "publisher": notice.Publisher, "official_link": notice.URL, "api_url": target.URL, "publication_expression": notice.Published,
				"platform_display_start": notice.DisplayFrom, "platform_display_end": notice.DisplayTo}
			if published != nil {
				metadata["source_publication_date"] = published.Format("2006-01-02")
			}
			resources := slices.Clone(target.Resources)
			for _, raw := range linkedPDFs(notice.URL, []byte(notice.Content)) {
				if !slices.ContainsFunc(resources, func(r PlannedResource) bool { return r.URL == raw }) {
					resources = append(resources, PlannedResource{URL: raw, Required: true})
				}
			}
			retained, err := e.retainCittadino(ctx, source, revision, notice.URL, page, metadata, resources, &state, contract, track, report.ObservedAt)
			if retained.VersionID > 0 {
				retained.Workload = "ordinary"
				if plan.Bootstrap {
					retained.Workload = "bootstrap"
				}
				report.Documents = append(report.Documents, retained)
			}
			if err != nil {
				return report, state, err
			}
			if track {
				if _, err := e.Tracking.Remember(ctx, source, revision, []DiscoveredDocument{{URL: target.URL, PublicationDate: published}}, report.ObservedAt); err != nil {
					state.errorCode = "tracking_failed"
					return report, state, err
				}
				if target.PublicationDate == nil && published != nil && (state.publicationObservedAt == nil || published.After(*state.publicationObservedAt)) {
					state.publicationObservedAt = published
				}
				if target.UnavailableStatus != 0 {
					if err := e.Tracking.ClearUnavailable(ctx, source, revision, target.URL, report.ObservedAt); err != nil {
						state.errorCode = "tracking_failed"
						return report, state, err
					}
				}
			}
		}
		if deferred != nil {
			return report, state, deferred
		}
	}
	if contract.Risks {
		report.Sections = append(report.Sections, SectionPreview{SectionURL: contract.BaseURL()})
		for _, day := range []string{"oggi", "domani"} {
			raw := contract.APIBase() + "rischi/" + day
			page, err := fetch(raw)
			if err != nil {
				return report, state, err
			}
			var response struct {
				Alerts []map[string]json.RawMessage `json:"stati_allerta"`
			}
			if json.Unmarshal(page.HTML, &response) != nil || len(response.Alerts) == 0 || len(response.Alerts) > 7 {
				return invalid("platform_risk_invalid")
			}
			for _, alert := range response.Alerts {
				var fields map[string]json.RawMessage
				var risk, level string
				if json.Unmarshal(alert["tipologia"], &risk) != nil || strings.TrimSpace(risk) == "" ||
					json.Unmarshal(alert[day], &fields) != nil || fields["allerta"] == nil || fields["data_bollettino"] == nil ||
					json.Unmarshal(fields["allerta"], &level) != nil || !slices.Contains([]string{"verde", "giallo", "arancione", "rosso"}, level) {
					return invalid("platform_risk_invalid")
				}
			}
			if len(report.Documents) >= cfg.Discovery.MaxDocuments {
				return invalid("document_limit_reached")
			}
			if track {
				discovery, err := e.Tracking.Remember(ctx, source, revision, []DiscoveredDocument{{URL: raw}}, report.ObservedAt)
				if err != nil {
					state.errorCode = "tracking_failed"
					return report, state, err
				}
				state.newDocuments += discovery.NewDocuments
			}
			metadata := map[string]any{"channel_role": "platform", "kind": "regional_republication", "verification_state": "pending",
				"municipality_istat": contract.MunicipalityISTAT, "official_link": contract.BaseURL(), "platform_day": day,
				"date_meaning": "platform_bulletin_expressions; no inferred CFR issuance or validity"}
			retained, err := e.retainCittadino(ctx, source, revision, raw, page, metadata, nil, &state, contract, track, report.ObservedAt)
			if err != nil {
				return report, state, err
			}
			retained.Workload = "ordinary"
			report.Documents = append(report.Documents, retained)
		}
	}
	if track {
		if err := e.Tracking.CompleteBootstrap(ctx, source, revision, report.ObservedAt); err != nil {
			state.errorCode = "tracking_failed"
			return report, state, err
		}
	}
	return report, state, nil
}

func cittadinoDetailURL(p *registry.CittadinoInformatoContract, id int64) string {
	return p.APIBase() + "aggiornamenti/" + strconv.FormatInt(id, 10) + "?comune=" + url.QueryEscape(p.MunicipalitySlug)
}

func cittadinoTargetID(p *registry.CittadinoInformatoContract, raw string) (int64, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return 0, false
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(u.Path, "/"+p.MunicipalitySlug+"/wp-json/cittadino/v2/aggiornamenti/"), 10, 64)
	return id, err == nil && id > 0 && raw == cittadinoDetailURL(p, id)
}

func validCittadinoNotice(p *registry.CittadinoInformatoContract, n cittadinoNotice, detail bool) bool {
	u, err := url.Parse(n.URL)
	return err == nil && n.ID > 0 && n.Publisher == p.Publisher && strings.TrimSpace(n.Title) != "" &&
		u.Scheme == "https" && u.Host == "cittadinoinformato.it" && u.User == nil && u.RawQuery == "" && u.Fragment == "" &&
		strings.HasPrefix(u.EscapedPath(), "/"+p.MunicipalitySlug+"/") && !strings.Contains(u.EscapedPath(), "wp-json") &&
		!strings.Contains(u.EscapedPath(), "..") && (!detail || strings.TrimSpace(n.Content) != "")
}

func cittadinoDate(raw string) *time.Time {
	date, err := time.Parse("02/01/2006", raw)
	if err != nil {
		return nil
	}
	return &date
}

func (e *Engine) retainCittadino(ctx context.Context, source string, revision int, officialURL string, page Page, metadata map[string]any, dependencies []PlannedResource, state *acquisitionState, contract *registry.CittadinoInformatoContract, track bool, at time.Time) (RetainedPage, error) {
	if len(dependencies) > 254 {
		state.errorCode = "attachment_limit_reached"
		return RetainedPage{}, ErrInvalidConfiguration
	}
	resources := []documents.Resource{{URL: page.URL, Role: "original", Required: true, SourceID: source, Configuration: revision, MediaType: "application/json", Bytes: page.HTML}}
	missing := false
	for _, dependency := range dependencies {
		resource := documents.Resource{URL: dependency.URL, Role: "attachment", Required: dependency.Required, SourceID: source, Configuration: revision}
		if !contract.AllowsAttachment(dependency.URL) {
			resource.Missing = "forbidden"
		} else {
			fetcher := e.Resources.(boundedCrawler)
			attachment, err := fetcher.CrawlBounded(ctx, dependency.URL, contract.AllowsAttachment)
			if err != nil {
				resource.Missing = "unavailable"
				if dependency.Required {
					observeCrawl(state, attachment, err)
				}
			} else {
				validate := e.validatePDF
				if validate == nil {
					validate = validateAttachmentPDF
				}
				if validate(ctx, attachment) != nil {
					resource.Missing = "unavailable"
					if dependency.Required {
						state.errorCode = "invalid_attachment_pdf"
					}
				} else {
					resource.MediaType, resource.Bytes = "application/pdf", attachment.HTML
				}
			}
		}
		missing = missing || dependency.Required && resource.Missing != ""
		resources = append(resources, resource)
	}
	encoded, _ := json.Marshal(metadata)
	hash := sha256.New()
	hash.Write([]byte(officialURL))
	hash.Write(encoded)
	for _, resource := range resources {
		hash.Write([]byte(resource.URL + "\x00" + resource.Missing + "\x00"))
		hash.Write(resource.Bytes)
	}
	id := fmt.Sprintf("cittadino-%s-%d-%s", source, revision, hex.EncodeToString(hash.Sum(nil)))
	version, err := e.Retained.Retain(ctx, documents.Acquisition{ID: id, SourceID: source, Configuration: revision, URL: page.URL, Metadata: encoded, Resources: resources})
	if err != nil {
		state.errorCode = "retention_failed"
		return RetainedPage{}, err
	}
	retained := RetainedPage{URL: officialURL, VersionID: version.ID, Hash: version.Hash, InferenceResources: version.Resources}
	if track {
		retained.Changed, err = e.Tracking.RecordVersion(ctx, source, page.URL, version.ID, at)
		if err != nil {
			state.errorCode = "tracking_failed"
			return retained, err
		}
	}
	if missing {
		if state.errorCode == "" {
			state.errorCode = "required_attachment_unavailable"
		}
		return retained, ErrRequiredAttachment
	}
	return retained, nil
}
