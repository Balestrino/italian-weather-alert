package extraction

import (
	"regexp"
	"strings"
)

var retainedOperativeList = regexp.MustCompile(`(?i)(?:^|[.!?]\s+)(Restano(?: al momento)? in vigore[^.!?:]{0,500}:\s*([^.!?;]{1,600}))`)
var retainedProhibitionList = regexp.MustCompile(`(?i)(?:^|\s+e\s+)divieto di ([^,.!?;:]{1,100}?)\s+(?:nei|nelle|negli|nel|nella)\s+(.+)$`)
var repeatedPlacePreposition = regexp.MustCompile(`(?i)(?:,\s*|\s+e\s+)(?:nei|nelle|negli|nel|nella)\s+`)
var uncertainRetainedList = regexp.MustCompile(`(?i)\b(?:non|forse|se|qualora|eccetto|tranne|esclus[oaie])\b`)
var retainedRelativeEnd = regexp.MustCompile(`(?i)salvo successive modifiche`)

// Split only a complete, positive retained list with explicit repeated
// prepositions. A conjunction inside one place is not a second scope.
func supplementRetainedProhibitions(measures []Measure, window Window) []Measure {
	for _, indices := range retainedOperativeList.FindAllStringSubmatchIndex(window.Text, -1) {
		quote := window.Text[indices[2]:indices[3]]
		body := window.Text[indices[4]:indices[5]]
		if indices[3] == len(window.Text) || !strings.ContainsRune(".!?", rune(window.Text[indices[3]])) || len(quote) > 1000 || !window.ownsOperativeQuote(quote) || reopeningInsideQuotation(window.Text[:indices[2]]) || uncertainRetainedList.MatchString(quote) || len(operativeEvidenceOutsideHeading([]Evidence{{Quote: quote}}, window)) == 0 {
			continue
		}
		list := retainedProhibitionList.FindStringSubmatch(body)
		if list == nil {
			continue
		}
		subject := strings.TrimSpace(list[1])
		places := repeatedPlacePreposition.Split(list[2], -1)
		if len(places) < 2 || len(places) > 8 {
			continue
		}
		valid, seen := true, map[string]bool{}
		for i := range places {
			places[i] = strings.TrimSpace(places[i])
			key := strings.ToLower(places[i])
			if len(places[i]) == 0 || len(places[i]) > 160 || strings.ContainsAny(places[i], ",;:") || seen[key] {
				valid = false
			}
			seen[key] = true
		}
		if !valid {
			continue
		}
		// Replace only candidates literally bound to this prohibition, including
		// a first place embedded in a grouped subject. Unrelated facts survive.
		kept := make([]Measure, 0, len(measures)+len(places))
		preserveTemporalScope := false
		for _, m := range measures {
			bound := false
			literalSubject := strings.EqualFold(m.Subject, subject)
			for _, place := range places {
				for _, preposition := range []string{"nei", "nelle", "negli", "nel", "nella"} {
					literalSubject = literalSubject || strings.EqualFold(m.Subject, subject+" "+preposition+" "+place)
				}
			}
			if m.Kind == "prohibition" && literalSubject && (m.Place == nil || seen[strings.ToLower(*m.Place)]) {
				for _, e := range m.Evidence {
					if e.Field == "kind" && strings.Contains(strings.ToLower(e.Quote), "divieto di "+strings.ToLower(subject)) && (strings.Contains(quote, e.Quote) || strings.Contains(e.Quote, quote)) {
						bound = true
					}
				}
			}
			if !bound {
				kept = append(kept, m)
			} else {
				for _, candidate := range m.TemporalCandidates {
					if !strings.EqualFold(candidate.OriginalExpression, "al momento") && !strings.EqualFold(candidate.OriginalExpression, "salvo successive modifiche") {
						preserveTemporalScope = true
					}
				}
			}
		}
		if preserveTemporalScope {
			continue
		}
		for _, place := range places {
			kind := Evidence{Field: "kind", ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: quote, SegmentOrdinal: window.Ordinal}
			activity, location := kind, kind
			activity.Field, activity.Quote = "subject", subject
			location.Field, location.Quote = "place", place
			m := Measure{Kind: "prohibition", Subject: subject, Place: &place, Evidence: []Evidence{kind, activity, location}}
			// Relative expressions belong to this retained clause. They do not
			// acquire an absolute date from page metadata or a reopened measure.
			for _, field := range []struct{ name, expression string }{{"valid_from", "al momento"}, {"valid_until", retainedRelativeEnd.FindString(quote)}} {
				if field.expression == "" || !strings.Contains(quote, field.expression) {
					continue
				}
				e := kind
				e.Field = field.name
				m.TemporalCandidates = append(m.TemporalCandidates, TemporalCandidate{Field: field.name, OriginalExpression: field.expression, Evidence: []Evidence{e}})
				m.Evidence = append(m.Evidence, e)
			}
			if projectTemporalCandidates(&m) == nil {
				m.IndeterminateFields = indeterminateFields(m)
				kept = append(kept, m)
			}
		}
		measures = kept
	}
	for i := range measures {
		measures[i].Ordinal = i + 1
	}
	return measures
}
