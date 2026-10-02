package classification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

const ManifestVersion = "interpretation-input-v2"

type InputManifest struct {
	IssuerID                         *string
	Version, SourceID, Configuration string
	DocumentID                       int64
	Metadata                         json.RawMessage
	Complete                         bool
	Sections                         []ContentSection
	Resources                        []ManifestResource
	Hash                             string `json:"-"`
	Reusable                         bool   `json:"-"`
}
type ManifestResource struct {
	URL, Role, MediaType, Missing, Identity, Policy string
	Required, Eligible                              bool
	Pages                                           []ManifestPage
}
type ManifestPage struct {
	Number                       int
	Input, Output, Model, Status string
}

var municipalNoise = regexp.MustCompile("<!-- js-view-dom-id-[0-9a-f]{64} -->")
var municipalViewClass = regexp.MustCompile(`<div class="js-view-dom-id-[0-9a-f]{64}">`)

const MunicipalHTMLPolicy = "municipal-html-v2"

const CascinaHTMLPolicy = "cascina-html-v1"

var cascinaCSRFMeta = regexp.MustCompile(`<meta name="csrf-token" content="[A-Za-z0-9]{40}">`)
var cascinaCSRFInput = regexp.MustCompile(`<input type="hidden" name="_token" value="[A-Za-z0-9]{40}">`)

// HTMLPolicy identifies only source-specific, reviewed presentation noise.
func HTMLPolicy(source string) string {
	switch source {
	case "calcinaia-municipal":
		return MunicipalHTMLPolicy
	case "cascina-municipal":
		return CascinaHTMLPolicy
	default:
		return ""
	}
}

// CanonicalHTML removes exact reviewed transport fields for the named source.
// Text, dates, links, other attributes and card order remain byte-significant.
func CanonicalHTML(source string, body []byte) []byte {
	if source == "cascina-municipal" {
		body = cascinaCSRFMeta.ReplaceAll(body, []byte(`<meta name="csrf-token" content="">`))
		return cascinaCSRFInput.ReplaceAll(body, []byte(`<input type="hidden" name="_token" value="">`))
	}
	if source != "calcinaia-municipal" {
		return body
	}
	body = municipalNoise.ReplaceAll(body, []byte("<!-- js-view-dom-id -->"))
	return municipalViewClass.ReplaceAll(body, []byte(`<div class="js-view-dom-id">`))
}

func manifestHash(body []byte) string { sum := sha256.Sum256(body); return hex.EncodeToString(sum[:]) }

// BuildManifest preserves all text, metadata and graphical inputs. It removes
// only exact source-specific transport fields proven irrelevant by paired evidence.
// Missing or unverifiable graphical provenance is represented but never reusable.
func BuildManifest(ctx context.Context, retained documentStore, extracted ocrStore, version documents.Version, content Content, configuration string) (InputManifest, error) {
	manifest := InputManifest{IssuerID: version.IssuerID, Version: ManifestVersion, Configuration: configuration, DocumentID: version.DocumentID, Complete: content.Complete, Sections: append([]ContentSection(nil), content.Sections...), Reusable: content.Complete && len(content.Sections) > 0}
	if configuration == "" || version.DocumentID < 1 {
		return manifest, ErrInvalid
	}
	for _, ref := range version.Resources {
		if ref.Role == "original" {
			manifest.SourceID = ref.SourceID
		}
	}
	if manifest.SourceID == "" {
		return manifest, ErrInvalid
	}
	var metadata any
	decoder := json.NewDecoder(bytes.NewReader(version.Metadata))
	decoder.UseNumber()
	if decoder.Decode(&metadata) != nil {
		return manifest, ErrInvalid
	}
	manifest.Metadata, _ = json.Marshal(metadata)
	pages, err := extracted.PagesForVersion(ctx, version.ID)
	if err != nil {
		return manifest, err
	}
	results, err := extracted.ResourcesForVersion(ctx, version.ID)
	if err != nil {
		return manifest, err
	}
	byResource := map[string][]ocr.PageResult{}
	complete := map[string]int{}
	for _, page := range pages {
		byResource[page.ResourceURL] = append(byResource[page.ResourceURL], page)
	}
	for _, result := range results {
		if result.Status == "complete" {
			complete[result.ResourceURL] = result.PageCount
		}
	}
	refs := append([]documents.Reference(nil), version.Resources...)
	sort.Slice(refs, func(i, j int) bool { return refs[i].URL < refs[j].URL })
	for _, ref := range refs {
		media := strings.ToLower(strings.TrimSpace(strings.Split(ref.MediaType, ";")[0]))
		item := ManifestResource{URL: ref.URL, Role: ref.Role, MediaType: media, Missing: ref.Missing, Required: ref.Required, Eligible: ref.InferenceEligible(), Policy: registry.EligibilityVersion + ":default"}
		if ref.Inference != nil {
			item.Policy = ref.Inference.Policy
		}
		if policy := ref.LocalProcessing.Identity(); policy != "" {
			item.Policy += ":" + policy
		}
		if !item.Eligible {
			manifest.Resources = append(manifest.Resources, item)
			continue
		}
		if ref.Missing != "" {
			manifest.Reusable = false
			manifest.Resources = append(manifest.Resources, item)
			continue
		}
		switch media {
		case "application/pdf", "image/png", "image/jpeg":
			pp := byResource[ref.URL]
			sort.Slice(pp, func(i, j int) bool { return pp[i].PageNumber < pp[j].PageNumber })
			if len(pp) == 0 || complete[ref.URL] != len(pp) {
				manifest.Reusable = false
				item.Identity = ref.Hash
			}
			for index, page := range pp {
				if page.PageNumber != index+1 || page.Status != "complete" || len(page.InputSHA256) != 64 || len(page.OutputSHA256) != 64 {
					manifest.Reusable = false
					item.Identity = ref.Hash
				}
				item.Pages = append(item.Pages, ManifestPage{Number: page.PageNumber, Input: page.InputSHA256, Output: page.OutputSHA256, Model: page.ReturnedModel, Status: page.Status})
			}
		default:
			body, err := retained.Read(ctx, version.ID, ref.URL)
			if err != nil {
				return manifest, err
			}
			if (media == "text/html" || media == "application/xhtml+xml") && HTMLPolicy(manifest.SourceID) != "" {
				body = CanonicalHTML(manifest.SourceID, body)
				if manifest.SourceID == "cascina-municipal" {
					item.Policy += ":" + CascinaHTMLPolicy
				}
			}
			item.Identity = manifestHash(body)
		}
		manifest.Resources = append(manifest.Resources, item)
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return manifest, err
	}
	manifest.Hash = manifestHash(encoded)
	return manifest, nil
}

func (s *Store) PutManifest(ctx context.Context, runID int64, m InputManifest) error {
	body, err := json.Marshal(m)
	if err != nil || runID < 1 || m.Hash != manifestHash(body) {
		return ErrInvalid
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO interpretation_input_manifests(run_id,document_id,source_id,configuration,manifest_sha256,body,reusable) SELECT $1,$2,$3,$4,$5,$6,$7 WHERE EXISTS(SELECT 1 FROM processing_runs r JOIN retained_versions v ON v.id=r.document_version_id WHERE r.id=$1 AND v.document_id=$2 AND r.configuration_version_id=$4) ON CONFLICT DO NOTHING`, runID, m.DocumentID, m.SourceID, m.Configuration, m.Hash, body, m.Reusable)
	if err != nil {
		return err
	}
	var hash string
	if err = s.pool.QueryRow(ctx, `SELECT manifest_sha256 FROM interpretation_input_manifests WHERE run_id=$1`, runID).Scan(&hash); err != nil {
		return err
	}
	if hash != m.Hash {
		return ErrConflict
	}
	return nil
}
