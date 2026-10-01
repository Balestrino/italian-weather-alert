// iwa-output-audit replays retained responses locally; it has no provider client.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
)

type auditCall struct {
	CaseID       string `json:"case_id"`
	Stage        string `json:"stage"`
	Segment      int    `json:"segment"`
	Output       string `json:"output"`
	FinishReason string `json:"finish_reason"`
	ErrorCode    string `json:"error_code"`
}
type auditDocument struct {
	Input struct {
		ID string `json:"id"`
	} `json:"input"`
	Content  classification.Content   `json:"content"`
	Segments []classification.Segment `json:"segments"`
	Windows  []extraction.Window      `json:"operational_windows"`
}
type finding struct {
	File, Case, Stage, HistoricalError, CurrentReason string
	Segment                                           int
	FinishReason                                      string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: iwa-output-audit <retained-evidence-directory>")
		os.Exit(2)
	}
	if err := run(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "retained output audit failed")
		os.Exit(1)
	}
}
func run(root string) error {
	paths, err := filepath.Glob(filepath.Join(root, "regression*.json"))
	if err != nil {
		return err
	}
	report := struct {
		Counts            map[string]int
		Failures          []finding
		ProductionUnknown int
		Limitations       []string
	}{Counts: map[string]int{}, Limitations: []string{"Historical regression responses cover different processing versions and are not the 74 production failures.", "Production raw responses were not retained; their detailed causes remain unknown.", "No provider calls or repairs are performed."}}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var run struct {
			Documents []auditDocument `json:"documents"`
			Calls     []auditCall     `json:"calls"`
		}
		if json.Unmarshal(body, &run) != nil {
			return fmt.Errorf("invalid retained file")
		}
		docs := map[string]auditDocument{}
		for _, doc := range run.Documents {
			for _, section := range doc.Content.Sections {
				doc.Content.Text += section.Text + "\n"
			}
			docs[doc.Input.ID] = doc
		}
		for _, call := range run.Calls {
			if call.Stage != "classification" && call.Stage != "extraction" {
				continue
			}
			report.Counts["reviewed_calls"]++
			reason := "response_unavailable"
			if call.Output != "" {
				doc, ok := docs[call.CaseID]
				if !ok {
					reason = "input_unavailable"
				} else {
					content := doc.Content
					for _, segment := range doc.Segments {
						if segment.Ordinal == call.Segment {
							content = segment.Content()
						}
					}
					if call.Stage == "classification" {
						_, err = classification.ParseDecision(call.Output, content)
						if err != nil {
							reason = classification.InvalidOutputReason(call.Output, call.FinishReason, content)
						} else {
							reason = "accepted_by_current_parser"
						}
					} else {
						var window *extraction.Window
						for i := range doc.Windows {
							if doc.Windows[i].Ordinal == call.Segment {
								window = &doc.Windows[i]
								break
							}
						}
						if window != nil {
							_, err = extraction.ParseWindow(call.Output, *window)
							reason = extraction.InvalidOutputReason(call.Output, call.FinishReason, *window, err)
						} else {
							_, err = extraction.Parse(call.Output, content)
							reason = "legacy_schema_or_evidence"
						}
						if err == nil {
							reason = "accepted_by_current_parser"
						}
					}
				}
			}
			report.Counts[reason]++
			if call.ErrorCode != "" || reason != "accepted_by_current_parser" {
				report.Failures = append(report.Failures, finding{File: filepath.Base(path), Case: call.CaseID, Stage: call.Stage, HistoricalError: call.ErrorCode, CurrentReason: reason, Segment: call.Segment, FinishReason: call.FinishReason})
			}
		}
	}
	production, err := os.ReadFile(filepath.Join(root, "regolo-efficiency-2026-09-22/corpus/manifest.json"))
	if err != nil {
		return err
	}
	var manifest struct {
		Failures []json.RawMessage `json:"validation_failures"`
	}
	if json.Unmarshal(production, &manifest) != nil {
		return fmt.Errorf("invalid production manifest")
	}
	report.ProductionUnknown = len(manifest.Failures)
	if len(paths) == 0 || report.ProductionUnknown == 0 {
		return fmt.Errorf("missing corpus")
	}
	// Avoid accidental absolute filesystem paths in a shareable report.
	for _, f := range report.Failures {
		if strings.Contains(f.File, "/") {
			return fmt.Errorf("invalid report path")
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
