package extraction

import (
	"regexp"
	"strings"
)

// Recognize only a completed circulation reopening in a standalone sentence.
// A preceding closure clause may carry its own time; it never dates reopening.
var namedReopening = regexp.MustCompile(`(?:^|[.!?]\s+)((?:Il|La) ((?:sottopasso|ponte|tunnel|strada) [^,.!?;:]{1,160})(?:, chius[oa] [^,.!?;:]{1,180},)? è stat[oa] riapert[oa] alla circolazione)`)
var uncertainReopeningSubject = regexp.MustCompile(`(?i)\b(?:non|mai|forse|se|qualora|quando|potrebbe|sarebbe|sarà|sara)\b`)

func supplementNamedReopenings(measures []Measure, window Window) []Measure {
	for _, match := range namedReopening.FindAllStringSubmatchIndex(window.Text, -1) {
		quote, subject := window.Text[match[2]:match[3]], strings.TrimSpace(window.Text[match[4]:match[5]])
		// Leave the terminal punctuation available as the next sentence's boundary.
		if match[3] < len(window.Text) && !strings.ContainsRune(".!?", rune(window.Text[match[3]])) {
			continue
		}
		if reopeningInsideQuotation(window.Text[:match[2]]) || uncertainReopeningSubject.MatchString(subject) || !window.ownsOperativeQuote(quote) || len(operativeEvidenceOutsideHeading([]Evidence{{Quote: quote}}, window)) == 0 {
			continue
		}
		present := false
		for index := range measures {
			m := &measures[index]
			if !strings.EqualFold(m.Subject, subject) {
				continue
			}
			if m.Kind != "operational_update" && m.Kind != "reopening" {
				continue
			}
			for _, e := range m.Evidence {
				if e.Field != "kind" || !(strings.Contains(e.Quote, quote) || strings.Contains(quote, e.Quote)) {
					continue
				}
				m.Kind = "reopening"
				m.Evidence = evidenceWithoutField(evidenceWithoutField(m.Evidence, "valid_from"), "valid_until")
				m.Evidence = evidenceWithoutField(m.Evidence, "kind")
				m.Evidence = append(m.Evidence, Evidence{Field: "kind", ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: quote, SegmentOrdinal: window.Ordinal})
				m.ValidFrom, m.ValidUntil, m.TemporalCandidates = nil, nil, nil
				m.IndeterminateFields = indeterminateFields(*m)
				present = true
				break
			}
		}
		if present {
			continue
		}
		kind := Evidence{Field: "kind", ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: quote, SegmentOrdinal: window.Ordinal}
		support := kind
		support.Field, support.Quote = "subject", subject
		m := Measure{Ordinal: len(measures) + 1, Kind: "reopening", Subject: subject, Evidence: []Evidence{kind, support}}
		m.IndeterminateFields = indeterminateFields(m)
		measures = append(measures, m)
	}
	return measures
}

func reopeningInsideQuotation(prefix string) bool {
	return strings.Count(prefix, "«") > strings.Count(prefix, "»") ||
		strings.Count(prefix, "“") > strings.Count(prefix, "”") ||
		strings.Count(prefix, `"`)%2 != 0
}
