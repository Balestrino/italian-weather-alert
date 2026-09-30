package interpretation

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
)

const PreflightVersion = "preflight-v1"

type PreflightDocuments interface {
	Version(context.Context, int64) (documents.Version, error)
	Read(context.Context, int64, string) ([]byte, error)
}
type Preflight struct {
	Documents                       PreflightDocuments
	Renderer                        ocr.Renderer
	Configuration, RendererIdentity string
	Sources                         map[string]bool
}
type PreflightManifest struct {
	IssuerID                                *string
	Policy, Configuration, Renderer, Source string
	DocumentID                              int64
	Metadata                                json.RawMessage
	Complete                                bool
	Resources                               []PreflightResource
}
type PreflightResource struct {
	URL, Source, Role, Media, Missing, Policy string
	Required, Eligible                        bool
	Identity                                  string
	Pages                                     []string
}
type PreflightDecision struct {
	VersionID        int64  `json:"version_id"`
	RepresentativeID int64  `json:"representative_id"`
	Fingerprint      string `json:"fingerprint"`
	Complete         bool   `json:"complete"`
}

func preflightHash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// Build never uses OCR or an inference adapter. Unknown resources retain their
// exact byte identity; only fixture-proven presentation noise is normalized.
func (p *Preflight) Build(ctx context.Context, v documents.Version) (PreflightManifest, error) {
	m := PreflightManifest{IssuerID: v.IssuerID, Policy: PreflightVersion, Configuration: p.Configuration, Renderer: p.RendererIdentity, DocumentID: v.DocumentID, Complete: v.Complete}
	if p.Configuration == "" || p.RendererIdentity == "" || v.DocumentID < 1 || p.Documents == nil {
		return m, ErrInvalid
	}
	var metadata any
	dec := json.NewDecoder(bytes.NewReader(v.Metadata))
	dec.UseNumber()
	if dec.Decode(&metadata) != nil {
		return m, ErrInvalid
	}
	m.Metadata, _ = json.Marshal(metadata)
	refs := append([]documents.Reference(nil), v.Resources...)
	sort.Slice(refs, func(i, j int) bool { return refs[i].URL < refs[j].URL })
	for _, r := range refs {
		if r.Role == "original" {
			m.Source = r.SourceID
		}
	}
	if m.Source == "" {
		return m, ErrInvalid
	}
	for _, r := range refs {
		item := PreflightResource{URL: r.URL, Source: r.SourceID, Role: r.Role, Media: strings.ToLower(strings.TrimSpace(strings.Split(r.MediaType, ";")[0])), Missing: r.Missing, Required: r.Required, Eligible: r.InferenceEligible()}
		if r.Inference != nil {
			item.Policy = r.Inference.Policy
		}
		if r.Missing != "" {
			m.Complete = false
		} else if item.Eligible {
			body, err := p.Documents.Read(ctx, v.ID, r.URL)
			if err != nil {
				return m, err
			}
			if preflightHash(body) != r.Hash {
				return m, documents.ErrStorage
			}
			switch {
			case item.Media == "application/pdf" && (m.Source == "cfr-criticality" || m.Source == "cfr-vigilance"):
				if p.Renderer == nil {
					return m, ErrInvalid
				}
				renderCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
				pages, err := p.Renderer.Render(renderCtx, body)
				cancel()
				if err != nil {
					return m, err
				}
				if len(pages) == 0 {
					return m, ErrInvalid
				}
				for i, page := range pages {
					if page.Number != i+1 || len(page.Bytes) == 0 {
						return m, ErrInvalid
					}
					item.Pages = append(item.Pages, preflightHash(page.Bytes))
				}
			default:
				if (item.Media == "text/html" || item.Media == "application/xhtml+xml") && m.Source == "calcinaia-municipal" {
					body = classification.CanonicalHTML(m.Source, body)
					item.Policy += ":" + classification.MunicipalHTMLPolicy
				}
				item.Identity = preflightHash(body)
			}
		}
		m.Resources = append(m.Resources, item)
	}
	return m, nil
}

// Prepare persists a contiguous equivalence dependency, never an interpretation.
// A transaction serializes decisions for the document; absent prior preparation,
// an archived boundary or uncertainty conservatively starts a new group.
func (s *Scheduler) Prepare(ctx context.Context, p *Preflight, id int64) (PreflightDecision, error) {
	d := PreflightDecision{VersionID: id, RepresentativeID: id}
	v, err := p.Documents.Version(ctx, id)
	if err != nil {
		return d, err
	}
	source := ""
	for _, r := range v.Resources {
		if r.Role == "original" {
			source = r.SourceID
		}
	}
	if !p.Sources[source] {
		return d, nil
	}
	m, err := p.Build(ctx, v)
	if err != nil {
		return d, err
	}
	wire, err := json.Marshal(m)
	if err != nil {
		return d, err
	}
	d.Fingerprint = preflightHash(wire)
	d.Complete = m.Complete
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return d, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", -v.DocumentID); err != nil {
		return d, err
	}
	var archived bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM interpretation_archives WHERE document_version_id=$1)", id).Scan(&archived); err != nil {
		return d, err
	}
	if archived {
		return d, tx.Commit(ctx)
	}
	var old PreflightDecision
	var cfg string
	err = tx.QueryRow(ctx, "SELECT representative_version_id,fingerprint,complete,configuration FROM interpretation_preflight WHERE document_version_id=$1", id).Scan(&old.RepresentativeID, &old.Fingerprint, &old.Complete, &cfg)
	if err == nil {
		// A catalog rotation changes the candidate preflight configuration,
		// but an already prepared version keeps its immutable original decision.
		// Only a mismatch under the same configuration indicates corrupted input.
		if cfg == p.Configuration && old.Fingerprint != d.Fingerprint {
			return d, ErrInvalid
		}
		old.VersionID = id
		return old, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return d, err
	}
	if m.Complete {
		var prior int64
		err = tx.QueryRow(ctx, `SELECT p.representative_version_id FROM interpretation_preflight p
 WHERE p.document_version_id=(SELECT max(id) FROM retained_versions WHERE document_id=$1 AND id<$2)
 AND p.complete AND p.fingerprint=$3 AND p.configuration=$4
 AND NOT EXISTS(SELECT 1 FROM interpretation_archives a WHERE a.document_version_id IN(p.document_version_id,p.representative_version_id))`, v.DocumentID, id, d.Fingerprint, p.Configuration).Scan(&prior)
		if err == nil {
			d.RepresentativeID = prior
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return d, err
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO interpretation_preflight(document_version_id,representative_version_id,configuration,fingerprint,body,complete) VALUES($1,$2,$3,$4,$5,$6)`, id, d.RepresentativeID, p.Configuration, d.Fingerprint, wire, m.Complete)
	if err != nil {
		return d, err
	}
	return d, tx.Commit(ctx)
}
