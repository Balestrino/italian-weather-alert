package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"mime"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"golang.org/x/net/html"
)

func (s *Store) verificationSource(ctx context.Context, sourceID, municipality string, revision int) (ChannelEvidence, error) {
	e := ChannelEvidence{SourceID: sourceID, Fields: map[string]VerifiedField{}}
	var product, territory string
	var external bool
	var body []byte
	err := s.pool.QueryRow(ctx, `SELECT s.authority_id,c.publisher_id,c.platform,c.url,s.product_id,s.territory,c.external,config.revision,config.body
 FROM registry_sources s JOIN registry_channels c ON c.id=s.channel_id
 JOIN registry_configurations config ON config.source_id=s.id AND config.revision=CASE WHEN $2::int>0 THEN $2 ELSE COALESCE(s.active_revision,s.latest_revision) END
 WHERE s.id=$1`, sourceID, revision).Scan(&e.AuthorityID, &e.PublisherID, &e.Platform, &e.URL, &product, &territory, &external, &e.Configuration, &body)
	if err != nil {
		return e, err
	}
	var cfg registry.Configuration
	if json.Unmarshal(body, &cfg) != nil {
		return e, ErrVerificationEvidence
	}
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return e, ErrVerificationEvidence
	}
	switch {
	case cfg.CittadinoInformato != nil:
		if product != "municipal" || territory != municipality || !external || e.Platform != "cittadino-informato" || !cfg.CittadinoInformato.Valid(cfg) {
			return e, ErrVerificationEvidence
		}
		e.Role = "platform"
	case strings.EqualFold(u.Hostname(), "cittadinoinformato.it") || strings.HasSuffix(strings.ToLower(u.Hostname()), ".cittadinoinformato.it") || e.Platform == "cittadino-informato":
		return e, ErrVerificationEvidence
	case product == "municipal" && territory == municipality:
		e.Role = "municipal"
		if external && (cfg.Referral == nil || cfg.Referral.Territory != municipality || cfg.Referral.ProductID != "municipal") {
			return e, ErrVerificationEvidence
		}
	case slices.Contains(Products, product) && slices.Contains([]string{"09", "toscana", "regione toscana"}, strings.ToLower(cleanLiteral(territory))) && e.AuthorityID == e.PublisherID && !external:
		e.Role = "regional"
	default:
		return e, ErrVerificationEvidence
	}
	return e, nil
}

func (s *Store) loadVerificationEvidence(ctx context.Context, docs RegionalDocuments, input VerificationEvidence, municipality string, at time.Time) (ChannelEvidence, error) {
	e := ChannelEvidence{}
	if input.SourceID == "" || input.VersionID < 1 || len(input.Fields) > len(verificationFields) {
		return e, ErrInvalid
	}
	v, err := docs.Version(ctx, input.VersionID)
	if err != nil {
		return e, err
	}
	var actualSource string
	if err = s.pool.QueryRow(ctx, `SELECT d.source_id FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id WHERE v.id=$1`, input.VersionID).Scan(&actualSource); err != nil {
		return e, err
	}
	if actualSource != input.SourceID {
		return e, ErrVerificationEvidence
	}
	var original documents.Reference
	for _, r := range v.Resources {
		if r.Role == "original" {
			original = r
		}
	}
	if original.SourceID != input.SourceID || original.Missing != "" {
		return e, ErrVerificationEvidence
	}
	e, err = s.verificationSource(ctx, input.SourceID, municipality, original.Configuration)
	if err != nil {
		return e, err
	}
	e.VersionID, e.Hash, e.Complete, e.URL = input.VersionID, v.Hash, v.Complete, original.URL
	// Original API/public URLs stay distinct, as recorded by the acquisition adapter.
	var metadata map[string]any
	if json.Unmarshal(v.Metadata, &metadata) == nil && e.Role == "platform" {
		if public, ok := metadata["official_link"].(string); ok {
			e.URL = public
		}
	}
	resourceBytes := map[string][]byte{}
	for name, selection := range input.Fields {
		if !slices.Contains(verificationFields, name) || selection.Locator == "" || len(selection.Locator) > 500 || selection.StartByte < 0 || selection.EndByte <= selection.StartByte || selection.EndByte-selection.StartByte > 2000 {
			return e, ErrVerificationEvidence
		}
		var resource documents.Reference
		for _, r := range v.Resources {
			if r.URL == selection.ResourceURL {
				resource = r
			}
		}
		if resource.URL == "" || resource.Missing != "" {
			return e, ErrVerificationEvidence
		}
		body, cached := resourceBytes[resource.URL]
		if !cached {
			var readErr error
			body, readErr = docs.Read(ctx, input.VersionID, resource.URL)
			if readErr != nil {
				return e, readErr
			}
			resourceBytes[resource.URL] = body
		}
		var text string
		if selection.OCRRunID > 0 || selection.Page > 0 {
			if selection.OCRRunID < 1 || selection.Page < 1 || selection.JSONPointer != "" || (resource.MediaType != "application/pdf" && !strings.HasPrefix(resource.MediaType, "image/")) {
				return e, ErrVerificationEvidence
			}
			err = s.pool.QueryRow(ctx, `SELECT p.extracted_text FROM ocr_page_results p JOIN ocr_resource_results r USING(run_id)
 WHERE p.run_id=$1 AND p.page_number=$2 AND p.document_version_id=$3 AND p.resource_url=$4 AND p.status='complete'
 AND r.document_version_id=p.document_version_id AND r.resource_url=p.resource_url AND r.status='complete' AND r.page_count>=p.page_number AND p.created_at<=$5 AND r.created_at<=$5`, selection.OCRRunID, selection.Page, input.VersionID, resource.URL, at).Scan(&text)
			if err != nil {
				return e, ErrVerificationEvidence
			}
			text = cleanLiteral(text)
		} else {
			text, err = verificationText(resource.MediaType, body, selection.JSONPointer)
			if err != nil {
				return e, err
			}
		}
		if selection.EndByte > len(text) || !utf8.ValidString(text[selection.StartByte:selection.EndByte]) {
			return e, ErrVerificationEvidence
		}
		if slices.Contains([]string{"kind", "product", "risk", "level"}, name) && !verificationWordBounds(text, selection.StartByte, selection.EndByte) {
			return e, ErrVerificationEvidence
		}
		passage := text[selection.StartByte:selection.EndByte]
		value := normalizedField(name, passage)
		if value == "" {
			return e, ErrVerificationEvidence
		}
		if name == "kind" && !slices.Contains([]string{"closure", "reopening", "restriction", "prohibition", "suspension", "activation", "deactivation", "operational_update", "observation"}, value) ||
			name == "risk" && !slices.Contains(Risks, value) || name == "level" && !slices.Contains(Levels, value) || name == "product" && !slices.Contains(Products, value) {
			return e, ErrVerificationEvidence
		}
		e.Fields[name] = VerifiedField{Value: value, Passage: passage, EvidenceSelection: selection}
	}
	return e, nil
}

func verificationText(media string, body []byte, pointer string) (string, error) {
	media, _, err := mime.ParseMediaType(media)
	if err != nil {
		return "", ErrVerificationEvidence
	}
	switch media {
	case "application/json":
		if pointer == "" || !strings.HasPrefix(pointer, "/") {
			return "", ErrVerificationEvidence
		}
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.UseNumber()
		var value any
		if decoder.Decode(&value) != nil {
			return "", ErrVerificationEvidence
		}
		for _, token := range strings.Split(pointer[1:], "/") {
			for i := 0; i < len(token); i++ {
				if token[i] == '~' {
					if i+1 >= len(token) || token[i+1] != '0' && token[i+1] != '1' {
						return "", ErrVerificationEvidence
					}
					i++
				}
			}
			token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
			switch node := value.(type) {
			case map[string]any:
				value = node[token]
			case []any:
				index, err := strconv.Atoi(token)
				if err != nil || index < 0 || index >= len(node) || strconv.Itoa(index) != token {
					return "", ErrVerificationEvidence
				}
				value = node[index]
			default:
				return "", ErrVerificationEvidence
			}
		}
		switch node := value.(type) {
		case string:
			return verificationHTMLText(node)
		case json.Number:
			return node.String(), nil
		default:
			return "", ErrVerificationEvidence
		}
	case "text/html":
		if pointer != "" {
			return "", ErrVerificationEvidence
		}
		return verificationHTMLText(string(body))
	case "text/plain":
		if pointer != "" || !utf8.Valid(body) {
			return "", ErrVerificationEvidence
		}
		return cleanLiteral(string(body)), nil
	default:
		return "", ErrVerificationEvidence
	}
}

func verificationHTMLText(body string) (string, error) {
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return "", ErrVerificationEvidence
	}
	var text []string
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && slices.Contains([]string{"script", "style", "template", "noscript"}, n.Data) {
			return
		}
		if n.Type == html.TextNode {
			text = append(text, n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(root)
	return cleanLiteral(strings.Join(text, " ")), nil
}

func verificationWordBounds(text string, start, end int) bool {
	word := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }
	if start > 0 {
		before, _ := utf8.DecodeLastRuneInString(text[:start])
		first, _ := utf8.DecodeRuneInString(text[start:end])
		if word(before) && word(first) {
			return false
		}
	}
	if end < len(text) {
		after, _ := utf8.DecodeRuneInString(text[end:])
		last, _ := utf8.DecodeLastRuneInString(text[start:end])
		if word(last) && word(after) {
			return false
		}
	}
	return true
}
