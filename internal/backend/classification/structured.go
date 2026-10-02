package classification

import (
	"context"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

const StructuredClassifierVersion = "structured-regional-v1"

// structuredDecision establishes only positive relevance. A regional product's
// contents, validity or graphical meaning still need their normal interpretation.
func structuredDecision(ctx context.Context, retained documentStore, version documents.Version, content Content) (Decision, bool, error) {
	if !content.Complete || len(content.Sections) == 0 {
		return Decision{}, false, nil
	}
	for _, ref := range version.Resources {
		policy := ref.LocalProcessing
		media := strings.ToLower(strings.TrimSpace(strings.Split(ref.MediaType, ";")[0]))
		if ref.Role != "original" || !ref.InferenceEligible() || ref.Missing != "" || policy == nil || policy.Version != registry.LocalProcessingVersion || len(policy.HTMLBodySelectors) != 0 || media != "text/html" && media != "application/xhtml+xml" {
			continue
		}
		product := ""
		switch policy.StructuredFormat {
		case "cfr-criticality-v1":
			product = "criticality"
		case "cfr-vigilance-v1":
			product = "vigilance"
		case "cfr-monitoring-v1":
			product = "monitoring"
		default:
			continue
		}
		body, err := retained.Read(ctx, version.ID, ref.URL)
		if err != nil {
			return Decision{}, false, err
		}
		projection, err := acquisition.ProjectRegionalHTML(product, body)
		if err != nil {
			return Decision{}, false, nil
		}
		var quote string
		if product == "monitoring" && projection.Statement == "NESSUN AVVISO IN CORSO DI VALIDITÀ O EVENTO IN CORSO" {
			quote = projection.Statement
		} else if len(projection.Facts) > 0 {
			// Use the recognized table row, including its risk/phenomenon, rather
			// than presenting a date alone as evidence of weather relevance.
			fact := projection.Facts[0]
			if strings.HasPrefix(fact.Locator, "HTML criticality table ") || strings.HasPrefix(fact.Locator, "HTML vigilance zone-summary table:") {
				_, row, _ := strings.Cut(fact.Locator, ": ")
				quote = strings.ReplaceAll(row, " | ", " ")
			} else if projection.Statement == "Criticità previste: NESSUNA" {
				quote = projection.Statement
			}
		} else if product == "vigilance" && projection.Statement == "Vigilance phenomenon zone-summary table" {
			// The strict parser recognized all six phenomenon rows and both
			// dates, even when no zones are listed. The product is still relevant;
			// its title establishes identity without asserting an all-clear.
			quote = "Bollettino di Vigilanza Meteorologica Regionale"
		}
		// Evidence must belong to the original's actual interpretation text,
		// rather than borrowing a coincidental quotation from an attachment.
		for _, section := range content.Sections {
			if section.ResourceURL == ref.URL && section.Role == "original" {
				if literal, ok := CanonicalLiteral(section.Text, quote); quote != "" && ok {
					return Decision{Relevant: true, ReasonCode: "regional_warning", EvidenceQuote: literal}, true, nil
				}
			}
		}
	}
	return Decision{}, false, nil
}
