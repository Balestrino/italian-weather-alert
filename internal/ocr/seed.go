package ocr

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/evidenceguard"
)

type SeedOptions struct {
	AfterRun                                      int64
	Limit                                         int
	Scope, Model, Configuration, RendererIdentity string
}
type SeedReport struct {
	NextRun                             int64 `json:"next_run"`
	Examined, Seeded, Existing, Skipped int
	Reasons                             map[string]int
}

// SeedHistorical has no inference adapter. It only reads retained resources,
// renders locally and reuses unambiguous output with byte-identical inputs.
func (s *Store) SeedHistorical(ctx context.Context, docs documentStore, renderer Renderer, options SeedOptions) (SeedReport, error) {
	report := SeedReport{NextRun: options.AfterRun, Reasons: map[string]int{}}
	if docs == nil || renderer == nil || options.AfterRun < 0 || options.Limit < 1 || options.Limit > 20 || options.Scope == "" || options.Model == "" || options.Configuration == "" || options.RendererIdentity == "" {
		return report, ErrInvalid
	}
	unlock, err := evidenceguard.Lock(ctx, s.pool, true)
	if err != nil {
		return report, err
	}
	defer unlock()
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT p.run_id FROM ocr_page_results p JOIN processing_runs r ON r.id=p.run_id WHERE p.run_id>$1 AND r.configuration_version_id=$2 ORDER BY p.run_id LIMIT $3`, options.AfterRun, options.Configuration, options.Limit)
	if err != nil {
		return report, err
	}
	var runs []int64
	for rows.Next() {
		var run int64
		if err = rows.Scan(&run); err != nil {
			rows.Close()
			return report, err
		}
		runs = append(runs, run)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return report, err
	}
	skip := func(reason string) { report.Skipped++; report.Reasons[reason]++ }
	for _, run := range runs {
		if ctx.Err() != nil {
			return report, ctx.Err()
		}
		resource, ok, err := s.Resource(ctx, run)
		if err != nil {
			return report, err
		}
		if !ok || resource.Status != "complete" || resource.PageCount < 1 || resource.PageCount > 500 {
			skip("incomplete_resource")
			report.NextRun = run
			continue
		}
		version, err := docs.Version(ctx, resource.DocumentVersionID)
		if err != nil {
			return report, err
		}
		media := ""
		for _, ref := range version.Resources {
			if ref.URL == resource.ResourceURL && ref.Missing == "" {
				media = ref.MediaType
			}
		}
		if media == "" {
			skip("missing_resource")
			report.NextRun = run
			continue
		}
		body, err := docs.Read(ctx, version.ID, resource.ResourceURL)
		if err != nil {
			return report, err
		}
		runner := Runner{Renderer: renderer}
		pages, err := runner.pages(ctx, media, body)
		if err != nil {
			skip("render_unavailable")
			report.NextRun = run
			continue
		}
		if len(pages) != resource.PageCount {
			skip("page_count_changed")
			report.NextRun = run
			continue
		}
		for _, image := range pages {
			report.Examined++
			page, ok, err := s.Page(ctx, run, image.Number)
			if err != nil {
				return report, err
			}
			if !ok || page.Status != "complete" || page.ReturnedModel != options.Model || page.InputSHA256 != digest(image.Bytes) || page.OutputSHA256 != digest([]byte(page.ExtractedText)) {
				skip("unverified_provenance")
				continue
			}
			var conflicts bool
			err = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ocr_page_results p JOIN processing_runs r ON r.id=p.run_id WHERE r.configuration_version_id=$1 AND p.input_sha256=$2 AND (p.output_sha256<>$3 OR p.status<>$4 OR p.returned_model<>$5))`, options.Configuration, page.InputSHA256, page.OutputSHA256, page.Status, options.Model).Scan(&conflicts)
			if err != nil {
				return report, err
			}
			if conflicts {
				skip("conflicting_historical_outputs")
				continue
			}
			request, _ := json.Marshal(pageRequest(options.Model, image))
			identity := ArtifactIdentity{Scope: options.Scope, Model: options.Model, Configuration: options.Configuration, Renderer: options.RendererIdentity, InputSHA256: page.InputSHA256, RequestSHA256: digest(request)}
			a, err := s.ClaimArtifact(ctx, identity, time.Now(), time.Minute)
			if err != nil {
				return report, err
			}
			if a.Ready {
				if a.Page.OutputSHA256 != page.OutputSHA256 {
					skip("existing_artifact_conflict")
				} else {
					report.Existing++
				}
				continue
			}
			if !a.Owned {
				skip("active_owner")
				continue
			}
			if err = s.CompleteArtifact(ctx, a, page, time.Now()); err != nil {
				return report, err
			}
			report.Seeded++
		}
		report.NextRun = run
	}
	return report, nil
}
