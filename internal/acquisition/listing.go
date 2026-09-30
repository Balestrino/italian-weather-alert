package acquisition

import (
	"slices"
	"strings"
	"time"

	"golang.org/x/net/html"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

type DiscoveredDocument struct {
	URL             string
	PublicationDate *time.Time
}

func discoverDocuments(baseRaw string, body []byte, cfg registry.Configuration, links []Link) []DiscoveredDocument {
	known := map[string]*time.Time{}
	for _, link := range links {
		if canonical, ok := boundedURL(baseRaw, link.URL); ok && isDocument(cfg, canonical) {
			known[canonical] = nil
		}
	}
	root, err := html.Parse(strings.NewReader(string(body)))
	if err == nil && listingParserConfigured(cfg.Discovery) {
		walkElements(root, func(node *html.Node) {
			if !hasClass(node, cfg.Discovery.ListingItemClass) {
				return
			}
			href := descendantAttribute(node, "a", "href", func(value string) bool {
				canonical, ok := boundedURL(baseRaw, value)
				return ok && isDocument(cfg, canonical)
			})
			if href == "" {
				return
			}
			canonical, _ := boundedURL(baseRaw, href)
			dateText := descendantTextByClass(node, cfg.Discovery.ListingDateClass)
			if cfg.Discovery.ListingDateLabel != "" {
				var found bool
				_, dateText, found = strings.Cut(dateText, cfg.Discovery.ListingDateLabel)
				if !found {
					return
				}
			}
			if parsed, ok := parseListingDate(dateText, cfg.Discovery.ListingDateLayout, cfg.Discovery.ListingDateLocale); ok {
				known[canonical] = &parsed
			}
		})
	}
	result := make([]DiscoveredDocument, 0, len(known))
	for raw, published := range known {
		result = append(result, DiscoveredDocument{URL: raw, PublicationDate: published})
	}
	slicesSortDiscovered(result)
	return result
}

func listingParserConfigured(d registry.Discovery) bool {
	return d.ListingItemClass != "" && d.ListingDateClass != "" && d.ListingDateLayout != "" && d.ListingDateLocale != ""
}

func parseListingDate(raw, layout, locale string) (time.Time, bool) {
	value := strings.Join(strings.Fields(raw), " ")
	if locale == "it" {
		months := map[string]string{
			"gen": "Jan", "gennaio": "Jan", "feb": "Feb", "febbraio": "Feb",
			"mar": "Mar", "marzo": "Mar", "apr": "Apr", "aprile": "Apr",
			"mag": "May", "maggio": "May", "giu": "Jun", "giugno": "Jun",
			"lug": "Jul", "luglio": "Jul", "ago": "Aug", "agosto": "Aug",
			"set": "Sep", "settembre": "Sep", "ott": "Oct", "ottobre": "Oct",
			"nov": "Nov", "novembre": "Nov", "dic": "Dec", "dicembre": "Dec",
		}
		parts := strings.Fields(value)
		for i, part := range parts {
			if month, ok := months[strings.ToLower(part)]; ok {
				parts[i] = month
			}
		}
		value = strings.Join(parts, " ")
	} else if locale != "en" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(layout, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func walkElements(node *html.Node, visit func(*html.Node)) {
	if node.Type == html.ElementNode {
		visit(node)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walkElements(child, visit)
	}
}

func hasClass(node *html.Node, class string) bool {
	for _, attribute := range node.Attr {
		if attribute.Key == "class" {
			for _, value := range strings.Fields(attribute.Val) {
				if value == class {
					return true
				}
			}
		}
	}
	return false
}

func descendantAttribute(node *html.Node, tag, attribute string, accept func(string) bool) string {
	if node.Type == html.ElementNode && node.Data == tag {
		for _, item := range node.Attr {
			if item.Key == attribute && accept(item.Val) {
				return item.Val
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if value := descendantAttribute(child, tag, attribute, accept); value != "" {
			return value
		}
	}
	return ""
}

func descendantTextByClass(node *html.Node, class string) string {
	if node.Type == html.ElementNode && hasClass(node, class) {
		return nodeText(node)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if value := descendantTextByClass(child, class); value != "" {
			return value
		}
	}
	return ""
}

func nodeText(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}
	var builder strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		builder.WriteString(nodeText(child))
		builder.WriteByte(' ')
	}
	return strings.TrimSpace(builder.String())
}

func slicesSortDiscovered(items []DiscoveredDocument) {
	slices.SortFunc(items, func(a, b DiscoveredDocument) int { return strings.Compare(a.URL, b.URL) })
}
