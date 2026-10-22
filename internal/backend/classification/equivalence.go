package classification

import (
	"bytes"
	"context"
	"encoding/json"
)

const EquivalenceVersion = "interpretation-equivalence-v1"

// EquivalenceFingerprint identifies an interpretation input independently of
// the processing configuration. Textual resources contribute through their
// interpreted sections, so markup outside that text does not matter; OCR and
// other non-text inputs keep their retained identity and page hashes.
func EquivalenceFingerprint(m InputManifest) (string, bool) {
	if !m.Reusable || !m.Complete || len(m.Sections) == 0 || m.Version != ManifestVersion || m.SourceID == "" || m.DocumentID < 1 {
		return "", false
	}
	// Metadata may come back from jsonb with different spacing.
	var metadata any
	decoder := json.NewDecoder(bytes.NewReader(m.Metadata))
	decoder.UseNumber()
	if decoder.Decode(&metadata) != nil {
		return "", false
	}
	type page struct {
		Number                int
		Input, Output, Status string
	}
	type resource struct {
		URL, Role, MediaType, Missing, Identity, Policy string
		Required, Eligible                              bool
		Pages                                           []page
	}
	resources := make([]resource, 0, len(m.Resources))
	for _, r := range m.Resources {
		item := resource{URL: r.URL, Role: r.Role, MediaType: r.MediaType, Missing: r.Missing, Policy: r.Policy, Required: r.Required, Eligible: r.Eligible}
		if !textual(r.MediaType) {
			item.Identity = r.Identity
			for _, p := range r.Pages {
				item.Pages = append(item.Pages, page{Number: p.Number, Input: p.Input, Output: p.Output, Status: p.Status})
			}
		}
		resources = append(resources, item)
	}
	wire, err := json.Marshal(struct {
		Version, SourceID string
		DocumentID        int64
		IssuerID          *string
		Metadata          any
		Complete          bool
		Sections          []ContentSection
		Resources         []resource
	}{EquivalenceVersion, m.SourceID, m.DocumentID, m.IssuerID, metadata, m.Complete, m.Sections, resources})
	if err != nil {
		return "", false
	}
	return manifestHash(wire), true
}

// recordEquivalence is idempotent: the fingerprint is a pure function of the
// immutable manifest, so a retry or backfill writes the same row once.
func (s *Store) recordEquivalence(ctx context.Context, runID int64, m InputManifest) error {
	fingerprint, ok := EquivalenceFingerprint(m)
	if !ok {
		return nil
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO interpretation_input_equivalences(run_id,document_id,version_id,fingerprint,fingerprint_version,created_at)
 SELECT m.run_id,m.document_id,r.document_version_id,$2,$3,clock_timestamp() FROM interpretation_input_manifests m JOIN processing_runs r ON r.id=m.run_id
 WHERE m.run_id=$1 AND m.reusable AND r.document_version_id IS NOT NULL ON CONFLICT DO NOTHING`, runID, fingerprint, EquivalenceVersion)
	return err
}

type EquivalenceBackfill struct {
	Scanned       int   `json:"scanned"`
	Written       int   `json:"written"`
	NotReusable   int   `json:"not_reusable"`
	AlreadyStored int   `json:"already_stored"`
	NextAfterRun  int64 `json:"next_after_run"`
	Complete      bool  `json:"complete"`
}

// BackfillEquivalences reads retained manifests only; it never accesses
// object storage or providers. Resume with NextAfterRun until Complete.
func (s *Store) BackfillEquivalences(ctx context.Context, afterRun int64, limit int) (EquivalenceBackfill, error) {
	report := EquivalenceBackfill{NextAfterRun: afterRun}
	if afterRun < 0 || limit < 1 || limit > 1000 {
		return report, ErrInvalid
	}
	rows, err := s.pool.Query(ctx, `SELECT m.run_id,m.body,m.reusable,e.run_id IS NOT NULL FROM interpretation_input_manifests m
 LEFT JOIN interpretation_input_equivalences e ON e.run_id=m.run_id WHERE m.run_id>$1 ORDER BY m.run_id LIMIT $2`, afterRun, limit)
	if err != nil {
		return report, err
	}
	type pending struct {
		run      int64
		manifest InputManifest
	}
	var todo []pending
	for rows.Next() {
		var run int64
		var body []byte
		var reusable, stored bool
		if err = rows.Scan(&run, &body, &reusable, &stored); err != nil {
			rows.Close()
			return report, err
		}
		report.Scanned++
		report.NextAfterRun = run
		var m InputManifest
		switch {
		case stored:
			report.AlreadyStored++
		case !reusable:
			report.NotReusable++
		case json.Unmarshal(body, &m) != nil:
			rows.Close()
			return report, ErrInvalid
		default:
			m.Reusable = true
			todo = append(todo, pending{run, m})
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return report, err
	}
	for _, item := range todo {
		if _, ok := EquivalenceFingerprint(item.manifest); !ok {
			report.NotReusable++
			continue
		}
		if err = s.recordEquivalence(ctx, item.run, item.manifest); err != nil {
			return report, err
		}
		report.Written++
	}
	report.Complete = report.Scanned < limit
	return report, nil
}
