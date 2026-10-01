package classification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"golang.org/x/net/html"
)

const maxClassificationContentBytes = 6 << 20

type documentStore interface {
	Version(context.Context, int64) (documents.Version, error)
	Read(context.Context, int64, string) ([]byte, error)
}

type ocrStore interface {
	PagesForVersion(context.Context, int64) ([]ocr.PageResult, error)
	ResourcesForVersion(context.Context, int64) ([]ocr.ResourceResult, error)
}

type ContentSection struct {
	ResourceURL string `json:"resource_url"`
	Role        string `json:"role"`
	Page        int    `json:"page,omitempty"`
	Text        string `json:"text"`
	// OffsetMap is populated for retained OCR pages and deliberately omitted
	// from provider payloads. It maps normalized evidence back to verbatim OCR.
	OffsetMap []OffsetSpan `json:"-"`
}

type Content struct {
	LegacyLiteral bool             `json:"-"` // Internal compatibility mode for the pre-canary extraction parser.
	Sections      []ContentSection `json:"sections"`
	Complete      bool             `json:"content_complete"`
	Hash          string           `json:"-"`
	Text          string           `json:"-"`
}

// Private aliases preserve the compact vocabulary used by the classifier while
// allowing later interpretation stages to consume the exact same evidence set.
type contentSection = ContentSection
type fullContent = Content

func GatherContent(ctx context.Context, retained documentStore, extracted ocrStore, version documents.Version) (Content, error) {
	pages, err := extracted.PagesForVersion(ctx, version.ID)
	if err != nil {
		return Content{}, err
	}
	resourceResults, err := extracted.ResourcesForVersion(ctx, version.ID)
	if err != nil {
		return Content{}, err
	}
	byResource := map[string][]ocr.PageResult{}
	for _, page := range pages {
		byResource[page.ResourceURL] = append(byResource[page.ResourceURL], page)
	}
	ocrStatus := map[string]string{}
	for _, result := range resourceResults {
		ocrStatus[result.ResourceURL] = result.Status
	}
	content := Content{Complete: version.Complete}
	for _, reference := range version.Resources {
		if !reference.InferenceEligible() {
			continue
		}
		if reference.Missing != "" {
			if reference.Required {
				content.Complete = false
			}
			continue
		}
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(reference.MediaType, ";")[0]))
		if mediaType == "application/pdf" || mediaType == "image/png" || mediaType == "image/jpeg" {
			status, ok := ocrStatus[reference.URL]
			if !ok || status == "missing" || status == "unreadable" || status == "partial_unreadable" {
				if reference.Required {
					content.Complete = false
				}
			}
			for _, page := range byResource[reference.URL] {
				if page.Status == "complete" && strings.TrimSpace(page.ExtractedText) != "" {
					view := NormalizeOCR(page.ExtractedText)
					if view.Text != "" {
						content.Sections = append(content.Sections, ContentSection{ResourceURL: reference.URL, Role: reference.Role, Page: page.PageNumber, Text: view.Text, OffsetMap: view.OffsetMap})
					}
				}
			}
			continue
		}
		if !textual(mediaType) {
			if reference.Required {
				content.Complete = false
			}
			continue
		}
		body, readErr := retained.Read(ctx, version.ID, reference.URL)
		if readErr != nil {
			return fullContent{}, readErr
		}
		text := normalize(string(body))
		if mediaType == "text/html" || mediaType == "application/xhtml+xml" {
			text = visibleHTML(body)
		}
		if text != "" {
			content.Sections = append(content.Sections, ContentSection{ResourceURL: reference.URL, Role: reference.Role, Text: text})
		}
	}
	wire, err := json.Marshal(struct {
		DocumentVersionID int64            `json:"document_version_id"`
		Sections          []ContentSection `json:"sections"`
	}{version.ID, content.Sections})
	if err != nil || len(wire) > maxClassificationContentBytes {
		return Content{}, ErrInvalid
	}
	hash := sha256.Sum256(wire)
	content.Hash = hex.EncodeToString(hash[:])
	var joined strings.Builder
	for _, section := range content.Sections {
		joined.WriteString(section.Text)
		joined.WriteByte('\n')
	}
	content.Text = joined.String()
	return content, nil
}

func gatherContent(ctx context.Context, retained documentStore, extracted ocrStore, version documents.Version) (fullContent, error) {
	return GatherContent(ctx, retained, extracted, version)
}

func textual(mediaType string) bool {
	return strings.HasPrefix(mediaType, "text/") || mediaType == "application/json" || mediaType == "application/xml" || strings.HasSuffix(mediaType, "+xml")
}

func normalize(value string) string { return strings.Join(strings.Fields(value), " ") }

func Normalize(value string) string { return normalize(value) }

func visibleHTML(body []byte) string {
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return normalize(string(body))
	}
	var values []string
	var walk func(*html.Node, bool)
	walk = func(node *html.Node, hidden bool) {
		if node.Type == html.ElementNode && (node.Data == "script" || node.Data == "style" || node.Data == "noscript") {
			hidden = true
		}
		if node.Type == html.TextNode && !hidden && strings.TrimSpace(node.Data) != "" {
			values = append(values, node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child, hidden)
		}
	}
	walk(root, false)
	return normalize(strings.Join(values, " "))
}

func contentMessage(versionID int64, content fullContent) (json.RawMessage, error) {
	body, err := json.Marshal(struct {
		DocumentVersionID int64            `json:"document_version_id"`
		ContentComplete   bool             `json:"content_complete"`
		Sections          []ContentSection `json:"sections"`
	}{versionID, content.Complete, content.Sections})
	if err != nil || len(body) > maxClassificationContentBytes {
		return nil, fmt.Errorf("%w", ErrInvalid)
	}
	return body, nil
}
