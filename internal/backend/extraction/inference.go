package extraction

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
)

const (
	CompactEnvelopeVersion    = "compact-evidence-v1"
	CompactEnvelopeVersionV2  = "compact-evidence-v2"
	MaxWindowCompletionTokens = 4096
	MaxCompactResponseBytes   = 16 << 10
)

var legacyResponseFormat = json.RawMessage(`{"type":"json_schema","json_schema":{"name":"measure_extraction","strict":true,"schema":{"type":"object","properties":{"measures":{"type":"array","maxItems":100,"items":{"type":"object","properties":{"kind":{"type":"string","enum":["closure","reopening","restriction","prohibition","suspension","activation","deactivation","operational_update","observation"]},"subject":{"type":"string"},"place":{"type":["string","null"]},"valid_from":{"type":["string","null"]},"valid_until":{"type":["string","null"]},"indeterminate_fields":{"type":"array","items":{"type":"string","enum":["place","valid_from","valid_until"]}},"evidence":{"type":"array","minItems":2,"items":{"type":"object","properties":{"field":{"type":"string","enum":["kind","subject","place","valid_from","valid_until"]},"resource_url":{"type":"string"},"page":{"type":["integer","null"],"minimum":1},"quote":{"type":"string"}},"required":["field","resource_url","page","quote"],"additionalProperties":false}}},"required":["kind","subject","place","valid_from","valid_until","indeterminate_fields","evidence"],"additionalProperties":false}}},"required":["measures"],"additionalProperties":false}}}`)

var compactResponseFormat = json.RawMessage(`{"type":"json_schema","json_schema":{"name":"compact_measure_extraction_v1","strict":true,"schema":{"type":"object","properties":{"envelope_version":{"type":"string","enum":["compact-evidence-v1"]},"window_ordinal":{"type":"integer","minimum":1},"evidence":{"type":"array","maxItems":100,"items":{"type":"string"}},"measures":{"type":"array","maxItems":100,"items":{"type":"object","properties":{"kind":{"type":"string","enum":["closure","reopening","restriction","prohibition","suspension","activation","deactivation","operational_update","observation"]},"subject":{"type":"string"},"place":{"type":["string","null"]},"valid_from":{"type":["string","null"]},"valid_until":{"type":["string","null"]},"evidence_refs":{"type":"object","properties":{"kind":{"type":"array","minItems":1,"maxItems":100,"items":{"type":"integer","minimum":0,"maximum":99}},"subject":{"type":"array","minItems":1,"maxItems":100,"items":{"type":"integer","minimum":0,"maximum":99}},"place":{"type":"array","maxItems":100,"items":{"type":"integer","minimum":0,"maximum":99}},"valid_from":{"type":"array","maxItems":100,"items":{"type":"integer","minimum":0,"maximum":99}},"valid_until":{"type":"array","maxItems":100,"items":{"type":"integer","minimum":0,"maximum":99}}},"required":["kind","subject","place","valid_from","valid_until"],"additionalProperties":false}},"required":["kind","subject","place","valid_from","valid_until","evidence_refs"],"additionalProperties":false}}},"required":["envelope_version","window_ordinal","evidence","measures"],"additionalProperties":false}}}`)

var compactResponseFormatV2 = json.RawMessage(`{"type":"json_schema","json_schema":{"name":"compact_measure_extraction_v2","strict":true,"schema":{"type":"object","properties":{"envelope_version":{"type":"string","enum":["compact-evidence-v2"]},"window_ordinal":{"type":"integer","minimum":1},"evidence":{"type":"array","maxItems":100,"items":{"type":"string"}},"measures":{"type":"array","maxItems":100,"items":{"type":"object","properties":{"kind":{"type":"string","enum":["closure","reopening","restriction","prohibition","suspension","activation","deactivation","operational_update","observation"]},"subject":{"type":"string"},"place":{"type":["string","null"]},"evidence_refs":{"type":"object","properties":{"kind":{"type":"array","minItems":1,"maxItems":100,"items":{"type":"integer","minimum":0,"maximum":99}},"subject":{"type":"array","minItems":1,"maxItems":100,"items":{"type":"integer","minimum":0,"maximum":99}},"place":{"type":"array","maxItems":100,"items":{"type":"integer","minimum":0,"maximum":99}}},"required":["kind","subject","place"],"additionalProperties":false},"temporal_candidates":{"type":"array","maxItems":100,"items":{"type":"object","properties":{"field":{"type":"string","enum":["valid_from","valid_until"]},"original_expression":{"type":"string"},"evidence_refs":{"type":"array","minItems":1,"maxItems":100,"items":{"type":"integer","minimum":0,"maximum":99}}},"required":["field","original_expression","evidence_refs"],"additionalProperties":false}}},"required":["kind","subject","place","evidence_refs","temporal_candidates"],"additionalProperties":false}}},"required":["envelope_version","window_ordinal","evidence","measures"],"additionalProperties":false}}}`)

var localStartExpression = regexp.MustCompile(`(?i)a partire dalle(?: ore)? [0-9]{1,2}[.:][0-9]{2}(?: del giorno)?(?: di)? [[:alpha:]àèéìòù]+ [0-9]{1,2} [[:alpha:]àèéìòù]+`)

type wireEvidence struct {
	Field       string `json:"field"`
	ResourceURL string `json:"resource_url"`
	Page        *int   `json:"page"`
	Quote       string `json:"quote"`
}

type wireMeasure struct {
	Kind                string         `json:"kind"`
	Subject             string         `json:"subject"`
	Place               *string        `json:"place"`
	ValidFrom           *string        `json:"valid_from"`
	ValidUntil          *string        `json:"valid_until"`
	IndeterminateFields []string       `json:"indeterminate_fields"`
	Evidence            []wireEvidence `json:"evidence"`
}

type response struct {
	Measures []wireMeasure `json:"measures"`
}

type compactEvidenceRefs struct {
	Kind       []int `json:"kind"`
	Subject    []int `json:"subject"`
	Place      []int `json:"place"`
	ValidFrom  []int `json:"valid_from"`
	ValidUntil []int `json:"valid_until"`
}

type compactWireMeasure struct {
	Kind       string              `json:"kind"`
	Subject    string              `json:"subject"`
	Place      *string             `json:"place"`
	ValidFrom  *string             `json:"valid_from"`
	ValidUntil *string             `json:"valid_until"`
	Evidence   compactEvidenceRefs `json:"evidence_refs"`
}

type compactResponse struct {
	EnvelopeVersion string               `json:"envelope_version"`
	WindowOrdinal   int                  `json:"window_ordinal"`
	Evidence        []string             `json:"evidence"`
	Measures        []compactWireMeasure `json:"measures"`
}

type compactTemporalCandidateV2 struct {
	Field              string `json:"field"`
	OriginalExpression string `json:"original_expression"`
	Evidence           []int  `json:"evidence_refs"`
}

type compactEvidenceRefsV2 struct {
	Kind    []int `json:"kind"`
	Subject []int `json:"subject"`
	Place   []int `json:"place"`
}

type compactWireMeasureV2 struct {
	Kind               string                       `json:"kind"`
	Subject            string                       `json:"subject"`
	Place              *string                      `json:"place"`
	Evidence           compactEvidenceRefsV2        `json:"evidence_refs"`
	TemporalCandidates []compactTemporalCandidateV2 `json:"temporal_candidates"`
}

type compactResponseV2 struct {
	EnvelopeVersion string                 `json:"envelope_version"`
	WindowOrdinal   int                    `json:"window_ordinal"`
	Evidence        []string               `json:"evidence"`
	Measures        []compactWireMeasureV2 `json:"measures"`
}

func Request(model string, versionID int64, content classification.Content) (inference.Request, error) {
	if model == "" || versionID < 1 || !content.Complete || len(content.Sections) == 0 {
		return inference.Request{}, ErrInvalid
	}
	body, err := json.Marshal(struct {
		DocumentVersionID int64                           `json:"document_version_id"`
		ContentComplete   bool                            `json:"content_complete"`
		Sections          []classification.ContentSection `json:"sections"`
	}{versionID, content.Complete, content.Sections})
	if err != nil {
		return inference.Request{}, ErrInvalid
	}
	input, _ := json.Marshal(string(body))
	system, _ := json.Marshal(LegacyPromptBody)
	return inference.Request{Model: model, Messages: []inference.Message{{Role: "system", Content: system}, {Role: "user", Content: input}}, ResponseFormat: legacyResponseFormat, MaxCompletionTokens: 12288}, nil
}

func SegmentRequest(model string, segment classification.Segment) (inference.Request, error) {
	if model == "" || segment.DocumentVersionID < 1 || segment.Ordinal < 1 || segment.Total < segment.Ordinal || segment.ResourceURL == "" || segment.StartByte < 0 || segment.EndByte <= segment.StartByte || segment.Text == "" || len(segment.Text) > classification.MaxSegmentTextBytes {
		return inference.Request{}, ErrInvalid
	}
	body, err := json.Marshal(struct {
		DocumentVersionID int64                  `json:"document_version_id"`
		Segment           extractionSegmentInput `json:"segment"`
	}{segment.DocumentVersionID, extractionSegmentInput{segment.Ordinal, segment.Total, segment.ResourceURL, segment.Role, segment.Page, segment.StartByte, segment.EndByte, segment.Text}})
	if err != nil {
		return inference.Request{}, ErrInvalid
	}
	input, _ := json.Marshal(string(body))
	system, _ := json.Marshal(LegacyPromptBody)
	return inference.Request{Model: model, Messages: []inference.Message{{Role: "system", Content: system}, {Role: "user", Content: input}}, ResponseFormat: legacyResponseFormat, MaxCompletionTokens: MaxWindowCompletionTokens}, nil
}

func WindowRequest(model string, window Window) (inference.Request, error) {
	if model == "" || window.DocumentVersionID < 1 || window.Ordinal < 1 || window.Total < window.Ordinal || window.ResourceURL == "" || window.ContextStartByte < 0 || window.CoreStartByte < window.ContextStartByte || window.CoreEndByte <= window.CoreStartByte || window.ContextEndByte < window.CoreEndByte || window.ContextEndByte-window.ContextStartByte != len(window.Text) || window.Text == "" || window.CoreEndByte-window.CoreStartByte > MaxWindowCoreBytes || window.CoreStartByte-window.ContextStartByte > MaxWindowContextBytes || window.ContextEndByte-window.CoreEndByte > MaxWindowContextBytes || len(window.Text) > MaxWindowCoreBytes+2*MaxWindowContextBytes {
		return inference.Request{}, ErrInvalid
	}
	body, err := json.Marshal(struct {
		DocumentVersionID int64                 `json:"document_version_id"`
		Window            extractionWindowInput `json:"window"`
	}{window.DocumentVersionID, extractionWindowInput{
		window.Ordinal, window.Total, window.ResourceURL, window.Role, window.Page,
		window.ContextStartByte, window.ContextEndByte, window.CoreStartByte, window.CoreEndByte, window.Text,
	}})
	if err != nil {
		return inference.Request{}, ErrInvalid
	}
	input, _ := json.Marshal(string(body))
	system, _ := json.Marshal(PromptBody)
	return inference.Request{Model: model, Messages: []inference.Message{{Role: "system", Content: system}, {Role: "user", Content: input}}, ResponseFormat: compactResponseFormatV2, MaxCompletionTokens: MaxWindowCompletionTokens}, nil
}

type extractionWindowInput struct {
	Ordinal          int    `json:"ordinal"`
	Total            int    `json:"total"`
	ResourceURL      string `json:"resource_url"`
	Role             string `json:"role"`
	Page             int    `json:"page,omitempty"`
	ContextStartByte int    `json:"context_start_byte"`
	ContextEndByte   int    `json:"context_end_byte"`
	CoreStartByte    int    `json:"core_start_byte"`
	CoreEndByte      int    `json:"core_end_byte"`
	Text             string `json:"text"`
}

type extractionSegmentInput struct {
	Ordinal     int    `json:"ordinal"`
	Total       int    `json:"total"`
	ResourceURL string `json:"resource_url"`
	Role        string `json:"role"`
	Page        int    `json:"page,omitempty"`
	StartByte   int    `json:"start_byte"`
	EndByte     int    `json:"end_byte"`
	Text        string `json:"text"`
}

func Parse(raw string, content classification.Content) ([]Measure, error) {
	return parse(raw, content, 0)
}

func ParseSegment(raw string, segment classification.Segment) ([]Measure, error) {
	return parse(raw, segment.Content(), segment.Ordinal)
}

func ParseWindow(raw string, window Window) ([]Measure, error) {
	measures, err := parseCompactWindow(raw, window)
	if err != nil {
		return nil, err
	}
	for _, measure := range measures {
		owned := false
		support := measure.Evidence
		if !window.legacyLiteral {
			support = operativeEvidenceOutsideHeading(support, window)
		}
		for _, evidence := range support {
			if evidence.Field == "kind" && window.ownsOperativeQuote(evidence.Quote) {
				owned = true
				break
			}
		}
		if !owned {
			return nil, ErrEvidence
		}
	}
	return measures, nil
}

func parseCompactWindow(raw string, window Window) ([]Measure, error) {
	if len(raw) > MaxCompactResponseBytes {
		return nil, ErrInvalid
	}
	var header struct {
		EnvelopeVersion string `json:"envelope_version"`
	}
	if json.Unmarshal([]byte(raw), &header) != nil {
		return nil, ErrInvalid
	}
	if header.EnvelopeVersion == CompactEnvelopeVersionV2 {
		return parseCompactWindowV2(raw, window)
	}
	if header.EnvelopeVersion != CompactEnvelopeVersion {
		return nil, ErrInvalid
	}
	return parseCompactWindowV1(raw, window)
}

func parseCompactWindowV1(raw string, window Window) ([]Measure, error) {
	var decoded compactResponse
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&decoded) != nil || !errors.Is(decoder.Decode(new(any)), io.EOF) || decoded.EnvelopeVersion != CompactEnvelopeVersion || decoded.WindowOrdinal != window.Ordinal || decoded.Evidence == nil || decoded.Measures == nil || len(decoded.Evidence) > 100 || len(decoded.Measures) > 100 {
		return nil, ErrInvalid
	}
	quotes := make([]string, len(decoded.Evidence))
	seenQuotes := make(map[string]bool, len(decoded.Evidence))
	for index, rawQuote := range decoded.Evidence {
		quote, ok := canonicalLiteralReference(window.Content(), window.ResourceURL, windowPage(window), classification.Normalize(rawQuote))
		if window.legacyLiteral {
			quote = classification.Normalize(rawQuote)
		}
		key := strings.ToLower(quote)
		if quote == "" || len(quote) > 1000 || seenQuotes[key] || !ok {
			return nil, ErrEvidence
		}
		seenQuotes[key] = true
		quotes[index] = quote
	}
	used := make([]bool, len(quotes))
	measures := make([]Measure, 0, len(decoded.Measures))
	for index, candidate := range decoded.Measures {
		measure, err := expandCompactMeasure(index+1, candidate, window, quotes, used)
		if err != nil {
			return nil, err
		}
		measures = append(measures, measure)
	}
	// Extra literal table rows do not assert a fact. Ignore them while retaining
	// the raw response for evaluation instead of discarding every supported
	// measure in the window. Duplicate, non-literal and out-of-range references
	// remain hard failures above and in expandCompactMeasure.
	return measures, nil
}

func parseCompactWindowV2(raw string, window Window) ([]Measure, error) {
	var decoded compactResponseV2
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&decoded) != nil || !errors.Is(decoder.Decode(new(any)), io.EOF) || decoded.EnvelopeVersion != CompactEnvelopeVersionV2 || decoded.WindowOrdinal != window.Ordinal || decoded.Evidence == nil || decoded.Measures == nil || len(decoded.Evidence) > 100 || len(decoded.Measures) > 100 {
		return nil, ErrInvalid
	}
	quotes := compactEvidenceTableV2(decoded.Evidence, window)
	measures := make([]Measure, 0, len(decoded.Measures))
	headingOnly := 0
	for index, candidate := range decoded.Measures {
		if !window.legacyLiteral && len(candidate.Evidence.Kind) > 0 {
			var support []Evidence
			for _, reference := range candidate.Evidence.Kind {
				if reference >= 0 && reference < len(quotes) && quotes[reference] != "" {
					support = append(support, Evidence{Quote: quotes[reference]})
				}
			}
			if len(support) == len(candidate.Evidence.Kind) && windowKindSupportedByEvidence(candidate.Kind, support, window) && !windowKindSupportedByEvidence(candidate.Kind, operativeEvidenceOutsideHeading(support, window), window) {
				headingOnly++
				continue
			}
		}
		measure, expandErr := expandCompactMeasureV2(index+1, candidate, window, quotes)
		if expandErr != nil {
			// A malformed or unsupported candidate is not a supported fact. Reject
			// that candidate without erasing independently validated measures from
			// the same well-formed window; the retained raw response remains
			// available to the evaluation report.
			continue
		}
		measures = append(measures, measure)
	}
	for index := range measures {
		measures[index].Ordinal = index + 1
	}
	measures = supplementExplicitOperativeMeasures(measures, window)
	if !window.legacyLiteral {
		measures = supplementRetainedClosures(measures, window)
		measures = supplementRetainedProhibitions(measures, window)
		measures = supplementRoadClearance(measures, window)
		measures = supplementDirectiveClosures(measures, window)
		measures = supplementNamedReopenings(measures, window)
	}
	if len(decoded.Measures) > 0 && len(measures) == 0 && headingOnly != len(decoded.Measures) {
		return nil, ErrEvidence
	}
	return measures, nil
}

var directiveClosures = regexp.MustCompile(`(?i)(?:^|[.;]\s+|,\s+)(ORDINA in via (?:contingibile|contingenti) e urgente[^:]{0,300}):\s*-\s*la chiusura al pubblico:\s*((?:[-○•]\s*(?:dell['’]|della|delle|degli|dei|del)\s*[^;:.]{1,180};\s*){1,5})`)
var closureListArticle = regexp.MustCompile(`(?i)^[-○•]\s*(?:dell['’]|della|delle|degli|dei|del)\s*`)
var directiveStart = regexp.MustCompile(`(?i)a partire dalle ore [0-9]{1,2}[.:][0-9]{2} del giorno [0-9]{1,2} [[:alpha:]àèéìòù]+ [0-9]{4}`)
var directiveConditionalEnd = regexp.MustCompile(`(?i)fino al perdurare dell['’]emergenza`)

// Complete a literal dispositive list only: headings and preamble lists cannot
// supply closures, and the following prohibition/activation has a separate scope.
func supplementDirectiveClosures(measures []Measure, window Window) []Measure {
	start, end := window.CoreStartByte-window.ContextStartByte, window.CoreEndByte-window.ContextStartByte
	if start < 0 || end > len(window.Text) || end <= start {
		return measures
	}
	core := window.Text[start:end]
	for _, indices := range directiveClosures.FindAllStringSubmatchIndex(core, -1) {
		prefix := strings.ToLower(core[:indices[2]])
		if boundary := strings.LastIndexAny(prefix, ".;"); boundary >= 0 {
			prefix = prefix[boundary+1:]
		}
		if strings.Contains(" "+prefix, " non ") || strings.Contains(" "+prefix, " se ") || strings.Contains(" "+prefix, " qualora ") || strings.Contains(prefix, "\"") || strings.Contains(prefix, "«") {
			continue
		}
		header, list := core[indices[2]:indices[3]], core[indices[4]:indices[5]]
		quote := core[indices[2]:indices[5]]
		if len(quote) > 1000 {
			continue
		}
		for _, item := range strings.Split(list, ";") {
			subject := strings.TrimSpace(closureListArticle.ReplaceAllString(strings.TrimSpace(item), ""))
			if subject == "" {
				continue
			}
			present := false
			for _, measure := range measures {
				if measure.Kind == "closure" && strings.EqualFold(measure.Subject, subject) {
					present = true
				}
			}
			if present {
				continue
			}
			kind := Evidence{Field: "kind", ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: quote, SegmentOrdinal: window.Ordinal}
			support := kind
			support.Field = "subject"
			measure := Measure{Kind: "closure", Subject: subject, Evidence: []Evidence{kind, support}}
			for _, field := range []struct {
				name       string
				expression *regexp.Regexp
			}{{"valid_from", directiveStart}, {"valid_until", directiveConditionalEnd}} {
				if expression := field.expression.FindString(header); expression != "" {
					e := kind
					e.Field, e.Quote = field.name, header
					measure.TemporalCandidates = append(measure.TemporalCandidates, TemporalCandidate{Field: field.name, OriginalExpression: expression, Evidence: []Evidence{e}})
					measure.Evidence = append(measure.Evidence, e)
				}
			}
			if projectTemporalCandidates(&measure) == nil {
				measures = append(measures, measure)
			}
		}
		bodyEnd := len(core)
		for _, marker := range []string{"avverte", "dispone"} {
			if offset := strings.Index(strings.ToLower(core[indices[5]:]), marker); offset >= 0 && indices[5]+offset < bodyEnd {
				bodyEnd = indices[5] + offset
			}
		}
		body := core[indices[2]:bodyEnd]
		for index := range measures {
			measure := &measures[index]
			if measure.Kind != "closure" && measure.Kind != "prohibition" && measure.Kind != "suspension" {
				continue
			}
			if !strings.Contains(strings.ToLower(body), strings.ToLower(measure.Subject)) || !windowKindSupportedByEvidence(measure.Kind, []Evidence{{Quote: body}}, window) {
				continue
			}
			for _, field := range []struct {
				name       string
				expression *regexp.Regexp
			}{{"valid_from", directiveStart}, {"valid_until", directiveConditionalEnd}} {
				expression := field.expression.FindString(header)
				if expression == "" {
					continue
				}
				present := false
				for _, candidate := range measure.TemporalCandidates {
					if candidate.Field == field.name && candidate.OriginalExpression == expression {
						present = true
					}
				}
				if present {
					continue
				}
				e := Evidence{Field: field.name, ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: body, SegmentOrdinal: window.Ordinal}
				measure.TemporalCandidates = append(measure.TemporalCandidates, TemporalCandidate{Field: field.name, OriginalExpression: expression, Evidence: []Evidence{e}})
				measure.Evidence = append(measure.Evidence, e)
			}
			_ = projectTemporalCandidates(measure)
		}
	}
	for index := range measures {
		measures[index].Ordinal = index + 1
	}
	return measures
}

// A positive, completed clearance report is an operational update. It does not
// revoke a separately stated closure or establish a formal reopening time.
var completedRoadClearance = regexp.MustCompile(`(?i)(?:^|[.!?]\s+)(?:aggiornamento:\s*)?si informa la cittadinanza che (tutte le strade[^.!?;:]{0,400}) sono state liberate[^.!?;:]{0,300}(?:[.!?]|$)`)

func supplementRoadClearance(measures []Measure, window Window) []Measure {
	start, end := window.CoreStartByte-window.ContextStartByte, window.CoreEndByte-window.ContextStartByte
	if start < 0 || end > len(window.Text) || end <= start {
		return measures
	}
	core := window.Text[start:end]
	for _, match := range completedRoadClearance.FindAllStringSubmatch(core, -1) {
		quote, subject := strings.TrimSpace(match[0]), strings.TrimSpace(match[1])
		// Refuse embedded negation, conditions and future/prospective wording.
		lower := " " + strings.ToLower(quote) + " "
		if strings.Contains(lower, " non ") || strings.Contains(lower, " se ") || strings.Contains(lower, " qualora ") {
			continue
		}
		present := false
		for _, measure := range measures {
			if measure.Kind != "operational_update" {
				continue
			}
			for _, evidence := range measure.Evidence {
				if evidence.Field == "kind" && strings.Contains(strings.ToLower(evidence.Quote), strings.ToLower(subject)+" sono state liberate") {
					present = true
				}
			}
		}
		if present {
			continue
		}
		kind := Evidence{Field: "kind", ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: quote, SegmentOrdinal: window.Ordinal}
		subjectEvidence := kind
		subjectEvidence.Field = "subject"
		measure := Measure{Kind: "operational_update", Subject: subject, Evidence: []Evidence{kind, subjectEvidence}}
		if projectTemporalCandidates(&measure) != nil {
			continue
		}
		measures = append(measures, measure)
	}
	for index := range measures {
		measures[index].Ordinal = index + 1
	}
	return measures
}

func supplementExplicitOperativeMeasures(measures []Measure, window Window) []Measure {
	present := map[string]bool{}
	for _, measure := range measures {
		present[measure.Kind] = true
	}
	coreStart, coreEnd := window.CoreStartByte-window.ContextStartByte, window.CoreEndByte-window.ContextStartByte
	if coreStart < 0 || coreEnd > len(window.Text) || coreEnd <= coreStart {
		return measures
	}
	core := window.Text[coreStart:coreEnd]
	type formula struct {
		kind     string
		markers  []string
		trimLead []string
		ends     []string
	}
	formulas := []formula{
		{"prohibition", []string{"il divieto di qualunque attività all'aperto", "il divieto di qualsiasi attività all’aperto", "il divieto di qualsiasi attività all'aperto"}, []string{"il divieto di "}, []string{"; - la sospensione", "; - la esposizione", " e’ stata infine", " è stata infine", "; avverte"}},
		{"suspension", []string{"la sospensione delle attività didattiche educative", "la sospensione delle attività educative", "la esposizione delle attività didattiche educative"}, []string{"la sospensione delle ", "la esposizione delle "}, []string{";", "."}},
		{"activation", []string{"dispone l'attivazione delle seguenti funzioni di supporto"}, []string{"dispone l'attivazione delle "}, []string{"."}},
	}
	for _, candidate := range formulas {
		if present[candidate.kind] {
			continue
		}
		quote, ok := operativeClause(core, candidate.markers, candidate.ends)
		if !ok {
			continue
		}
		subject := quote
		lowerSubject := strings.ToLower(subject)
		for _, prefix := range candidate.trimLead {
			if strings.HasPrefix(lowerSubject, prefix) {
				subject = subject[len(prefix):]
				break
			}
		}
		subject = strings.TrimSpace(strings.TrimRight(subject, ";."))
		if subject == "" {
			continue
		}
		kindEvidence := Evidence{Field: "kind", ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: quote, SegmentOrdinal: window.Ordinal}
		subjectEvidence := kindEvidence
		subjectEvidence.Field = "subject"
		measure := Measure{Kind: candidate.kind, Subject: subject, Evidence: []Evidence{kindEvidence, subjectEvidence}}
		supplementLocalStartCandidates(&measure, window)
		measure.Evidence = mergeEvidence(nil, measure.Evidence)
		if projectTemporalCandidates(&measure) != nil {
			continue
		}
		measures = append(measures, measure)
		present[candidate.kind] = true
	}
	for index := range measures {
		measures[index].Ordinal = index + 1
	}
	return measures
}

var ordinanceHeading = regexp.MustCompile(`(?i)\boggetto\s*:`)
var ordinanceBody = regexp.MustCompile(`(?i)\b(?:visto|vista|visti|viste|premesso|considerato|richiamato|ordina|dispone)\b`)

// Heading evidence can qualify an independently operative fact (including a
// temporal conflict), but cannot by itself establish that fact.
func operativeEvidenceOutsideHeading(evidence []Evidence, window Window) []Evidence {
	var operative []Evidence
	for _, item := range evidence {
		supported := false
		for offset := 0; offset <= len(window.Text)-len(item.Quote); {
			relative := strings.Index(window.Text[offset:], item.Quote)
			if relative < 0 {
				break
			}
			start := offset + relative
			headingOnly := false
			for _, heading := range ordinanceHeading.FindAllStringIndex(window.Text, -1) {
				end := len(window.Text)
				if body := ordinanceBody.FindStringIndex(window.Text[heading[1]:]); body != nil {
					end = heading[1] + body[0]
				}
				if start >= heading[0] && start+len(item.Quote) <= end {
					headingOnly = true
				}
			}
			if !headingOnly {
				supported = true
				break
			}
			offset = start + 1
		}
		if supported {
			operative = append(operative, item)
		}
	}
	return operative
}

var retainedClosure = regexp.MustCompile(`(?i)(?:^|[.!?]\s+)(Restano(?: al momento)? in vigore[^.!?:]{0,500}:\s*chiusura al pubblico (?:dell['’]|della|delle|degli|dei|del)\s*([^.;:]{1,500}))`)
var coordinatedSubject = regexp.MustCompile(`(?i)\s+e\s+(?:dell['’]|della|delle|degli|dei|del)\s*`)
var nextRetainedKind = regexp.MustCompile(`(?i)\s+e\s+(?:divieto|sospensione|attivazione)\b`)

// Complete only positive, explicitly retained coordinated closures. Subject
// coverage, rather than kind alone, determines whether a fact is missing.
func supplementRetainedClosures(measures []Measure, window Window) []Measure {
	for _, match := range retainedClosure.FindAllStringSubmatchIndex(window.Text, -1) {
		quote := window.Text[match[2]:match[3]]
		list := window.Text[match[4]:match[5]]
		if boundary := nextRetainedKind.FindStringIndex(list); boundary != nil {
			quote = quote[:len(quote)-len(list)+boundary[0]]
			list = list[:boundary[0]]
		}
		if len(quote) > 1000 || !window.ownsOperativeQuote(quote) {
			continue
		}
		subjects := coordinatedSubject.Split(list, -1)
		if len(subjects) < 2 {
			continue
		}
		for _, subject := range subjects {
			subject = strings.TrimSpace(subject)
			present := subject == ""
			for _, measure := range measures {
				if measure.Kind == "closure" && strings.Contains(strings.ToLower(measure.Subject), strings.ToLower(subject)) {
					present = true
				}
			}
			if present {
				continue
			}
			kind := Evidence{Field: "kind", ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: quote, SegmentOrdinal: window.Ordinal}
			support := kind
			support.Field, support.Quote = "subject", subject
			measure := Measure{Ordinal: len(measures) + 1, Kind: "closure", Subject: subject, Evidence: []Evidence{kind, support}}
			measure.IndeterminateFields = indeterminateFields(measure)
			measures = append(measures, measure)
		}
	}
	return measures
}

func operativeClause(core string, markers, endMarkers []string) (string, bool) {
	lower := strings.ToLower(core)
	start := -1
	for _, marker := range markers {
		if index := strings.Index(lower, marker); index >= 0 && (start < 0 || index < start) {
			start = index
		}
	}
	if start < 0 {
		return "", false
	}
	end := len(core)
	for _, marker := range endMarkers {
		if relative := strings.Index(lower[start:], marker); relative > 0 && start+relative < end {
			end = start + relative
		}
	}
	quote := strings.TrimSpace(core[start:end])
	return quote, quote != "" && len(quote) <= 1000
}

// compactEvidenceTableV2 preserves table indices while marking malformed rows
// unusable. A candidate that references one of those rows is rejected by
// expandCompactMeasureV2, but unrelated candidates can still be accepted.
func compactEvidenceTableV2(raw []string, window Window) []string {
	quotes := make([]string, len(raw))
	seen := make(map[string]bool, len(raw))
	for index, value := range raw {
		quote := classification.Normalize(value)
		key := strings.ToLower(quote)
		literal, ok := canonicalLiteralReference(window.Content(), window.ResourceURL, windowPage(window), quote)
		if quote == "" || len(quote) > 1000 || seen[key] || !ok {
			continue
		}
		seen[key] = true
		quotes[index] = literal
	}
	return quotes
}

func expandCompactMeasure(ordinal int, candidate compactWireMeasure, window Window, quotes []string, used []bool) (Measure, error) {
	candidate.Subject = classification.Normalize(candidate.Subject)
	validKind := map[string]bool{"closure": true, "reopening": true, "restriction": true, "prohibition": true, "suspension": true, "activation": true, "deactivation": true, "operational_update": true, "observation": true}
	if !validKind[candidate.Kind] || candidate.Subject == "" || len(candidate.Subject) > 1000 {
		return Measure{}, ErrInvalid
	}
	optional := map[string]*string{"place": cleanPointer(candidate.Place), "valid_from": cleanPointer(candidate.ValidFrom), "valid_until": cleanPointer(candidate.ValidUntil)}
	if candidate.Place != nil && optional["place"] == nil || candidate.ValidFrom != nil && optional["valid_from"] == nil || candidate.ValidUntil != nil && optional["valid_until"] == nil {
		return Measure{}, ErrInvalid
	}
	refs := map[string][]int{
		"kind": candidate.Evidence.Kind, "subject": candidate.Evidence.Subject, "place": candidate.Evidence.Place,
		"valid_from": candidate.Evidence.ValidFrom, "valid_until": candidate.Evidence.ValidUntil,
	}
	// Optional text without a reference is not an evidence-backed assertion. Treat
	// it as unknown instead of rejecting otherwise usable measures; the raw model
	// response remains retained by the caller for unsupported-assertion review.
	for _, field := range []string{"place", "valid_from", "valid_until"} {
		if optional[field] != nil && len(refs[field]) == 0 {
			optional[field] = nil
		}
	}
	for field, indices := range refs {
		asserted := field == "kind" || field == "subject" || optional[field] != nil
		if indices == nil || asserted != (len(indices) > 0) || len(indices) > 100 {
			return Measure{}, ErrInvalid
		}
	}
	page := windowPage(window)
	evidence := make([]Evidence, 0)
	for _, field := range []string{"kind", "subject", "place", "valid_from", "valid_until"} {
		seen := map[int]bool{}
		for _, reference := range refs[field] {
			if reference < 0 || reference >= len(quotes) || seen[reference] {
				return Measure{}, ErrEvidence
			}
			seen[reference], used[reference] = true, true
			evidence = append(evidence, Evidence{Field: field, ResourceURL: window.ResourceURL, Page: page, Quote: quotes[reference], SegmentOrdinal: window.Ordinal})
		}
	}
	measure := Measure{Ordinal: ordinal, Kind: candidate.Kind, Subject: candidate.Subject, Place: optional["place"], ValidFrom: optional["valid_from"], ValidUntil: optional["valid_until"], Evidence: evidence}
	measure.TemporalCandidates = temporalCandidatesFromResolved(measure)
	measure.IndeterminateFields = indeterminateFields(measure)
	return measure, nil
}

func expandCompactMeasureV2(ordinal int, candidate compactWireMeasureV2, window Window, quotes []string) (Measure, error) {
	candidate.Subject = classification.Normalize(candidate.Subject)
	validKind := map[string]bool{"closure": true, "reopening": true, "restriction": true, "prohibition": true, "suspension": true, "activation": true, "deactivation": true, "operational_update": true, "observation": true}
	if !validKind[candidate.Kind] || candidate.Subject == "" || len(candidate.Subject) > 1000 || candidate.TemporalCandidates == nil {
		return Measure{}, ErrInvalid
	}
	place := cleanPointer(candidate.Place)
	if candidate.Place != nil && place == nil {
		return Measure{}, ErrInvalid
	}
	refs := map[string][]int{"kind": candidate.Evidence.Kind, "subject": candidate.Evidence.Subject, "place": candidate.Evidence.Place}
	if place != nil && len(refs["place"]) == 0 {
		place = nil
	}
	// A null optional field makes no assertion. Validate reference indices, but
	// do not discard supported kind/subject facts for redundant place refs.
	if !window.legacyLiteral && place == nil && refs["place"] != nil {
		if len(refs["place"]) > 100 {
			return Measure{}, ErrInvalid
		}
		seen := map[int]bool{}
		for _, index := range refs["place"] {
			if index < 0 || index >= len(quotes) || seen[index] {
				return Measure{}, ErrEvidence
			}
			seen[index] = true
		}
		refs["place"] = []int{}
	}
	for field, indices := range refs {
		asserted := field == "kind" || field == "subject" || place != nil
		if indices == nil || asserted != (len(indices) > 0) || len(indices) > 100 {
			return Measure{}, ErrInvalid
		}
	}
	page := windowPage(window)
	evidence := make([]Evidence, 0)
	appendEvidence := func(field string, indices []int) ([]Evidence, error) {
		selected := make([]Evidence, 0, len(indices))
		seen := map[int]bool{}
		for _, reference := range indices {
			if reference < 0 || reference >= len(quotes) || seen[reference] {
				return nil, ErrEvidence
			}
			seen[reference] = true
			if quotes[reference] == "" {
				continue
			}
			item := Evidence{Field: field, ResourceURL: window.ResourceURL, Page: page, Quote: quotes[reference], SegmentOrdinal: window.Ordinal}
			selected = append(selected, item)
			evidence = append(evidence, item)
		}
		return selected, nil
	}
	selectedByField := map[string][]Evidence{}
	for _, field := range []string{"kind", "subject", "place"} {
		selected, err := appendEvidence(field, refs[field])
		if err != nil {
			return Measure{}, err
		}
		selectedByField[field] = selected
	}
	if len(selectedByField["kind"]) == 0 {
		return Measure{}, ErrEvidence
	}
	kindEvidence := selectedByField["kind"]
	if !window.legacyLiteral {
		kindEvidence = operativeEvidenceOutsideHeading(kindEvidence, window)
	}
	if !windowKindSupportedByEvidence(candidate.Kind, kindEvidence, window) {
		return Measure{}, ErrEvidence
	}
	if !valueInEvidence(candidate.Subject, selectedByField["subject"]) {
		contextual, ok := contextualFieldEvidence("subject", candidate.Subject, window, selectedByField["kind"])
		if !ok {
			return Measure{}, ErrEvidence
		}
		evidence = append(evidence, contextual)
		selectedByField["subject"] = append(selectedByField["subject"], contextual)
	}
	if place != nil && !valueInEvidence(*place, selectedByField["place"]) {
		contextual, ok := contextualFieldEvidence("place", *place, window, selectedByField["kind"])
		if ok {
			evidence = append(evidence, contextual)
			selectedByField["place"] = append(selectedByField["place"], contextual)
			if !window.legacyLiteral {
				// Scope validation must use the retained spelling as the evidence
				// table does, including reviewed apostrophe presentation changes.
				place = &contextual.Quote
			}
		} else {
			place = nil
			evidence = evidenceWithoutField(evidence, "place")
		}
	}
	if place != nil && !placeScopedToSubject(*place, candidate.Subject, selectedByField) {
		// A place named in one item of a list cannot scope the list's
		// general predicate. Keep the supported measure with unknown place.
		place = nil
		evidence = evidenceWithoutField(evidence, "place")
	}
	if subject, scopeEvidence, ok := listedSubjectScope(candidate.Kind, candidate.Subject, selectedByField["kind"], window); ok {
		candidate.Subject = subject
		evidence = append(evidence, scopeEvidence)
	}
	measure := Measure{Ordinal: ordinal, Kind: candidate.Kind, Subject: candidate.Subject, Place: place, Evidence: evidence}
	for _, raw := range candidate.TemporalCandidates {
		expression := classification.Normalize(raw.OriginalExpression)
		if (raw.Field != "valid_from" && raw.Field != "valid_until") || expression == "" || len(expression) > 1000 || raw.Evidence == nil || len(raw.Evidence) == 0 || len(raw.Evidence) > 100 || !valueInReferencedQuotes(expression, raw.Evidence, quotes) {
			continue
		}
		if temporalReferencesOnlyMetadata(raw.Evidence, quotes) {
			continue
		}
		if candidate.Kind == "activation" && !referencesContainActivation(raw.Evidence, quotes, window) {
			continue
		}
		if !window.legacyLiteral && candidate.Kind != "activation" {
			supported := false
			for _, reference := range raw.Evidence {
				if reference >= 0 && reference < len(quotes) && strings.Contains(strings.ToLower(quotes[reference]), strings.ToLower(expression)) {
					anchor := temporalSubjectAnchor(candidate.Subject)
					if (anchor == "" || strings.Contains(strings.ToLower(quotes[reference]), anchor)) && windowKindSupportedByEvidence(candidate.Kind, []Evidence{{Quote: quotes[reference]}}, window) {
						supported = true
					}
				}
			}
			if !supported {
				continue
			}
		}
		selected, err := appendEvidence(raw.Field, raw.Evidence)
		if err != nil {
			return Measure{}, err
		}
		measure.TemporalCandidates = append(measure.TemporalCandidates, TemporalCandidate{Field: raw.Field, OriginalExpression: expression, Evidence: selected})
	}
	measure.Evidence = evidence
	supplementLocalStartCandidates(&measure, window)
	measure.Evidence = mergeEvidence(nil, measure.Evidence)
	if err := projectTemporalCandidates(&measure); err != nil {
		return Measure{}, err
	}
	return measure, nil
}

func temporalReferencesOnlyMetadata(references []int, quotes []string) bool {
	for _, reference := range references {
		if reference < 0 || reference >= len(quotes) {
			return false
		}
		quote := strings.ToLower(quotes[reference])
		metadata := strings.Contains(quote, "allerta meteo") || strings.Contains(quote, "validità dell'allerta") || strings.Contains(quote, "validità dell’allerta") || strings.HasPrefix(quote, "data:") || strings.HasPrefix(quote, "data scadenza:") || strings.HasPrefix(quote, "ultimo aggiornamento")
		operative := strings.Contains(quote, "ordina") || strings.Contains(quote, "chiusur") || strings.Contains(quote, "riapert") || strings.Contains(quote, "divieto") || strings.Contains(quote, "sospension") || strings.Contains(quote, "attivazion")
		if !metadata || operative {
			return false
		}
	}
	return len(references) > 0
}

func referencesContainActivation(references []int, quotes []string, window Window) bool {
	for _, reference := range references {
		if reference >= 0 && reference < len(quotes) && windowKindSupportedByEvidence("activation", []Evidence{{Quote: quotes[reference]}}, window) {
			return true
		}
	}
	return false
}

var operativeTransition = regexp.MustCompile(`(?i)(?:e[’']|è) stat[oa] (?:inoltre|infine)`)

func supplementLocalStartCandidates(measure *Measure, window Window) {
	anchor, predicate := temporalSubjectAnchor(measure.Subject), temporalKindPredicate(measure.Kind)
	if anchor == "" || predicate == "" {
		return
	}
	text, lower := window.Text, strings.ToLower(window.Text)
	for _, location := range localStartExpression.FindAllStringIndex(text, -1) {
		start := 0
		if boundary := strings.LastIndex(text[:location[0]], ". "); boundary >= 0 {
			start = boundary + 2
		}
		end := len(text)
		if boundary := strings.Index(text[location[1]:], ". "); boundary >= 0 {
			end = location[1] + boundary + 1
		}
		if !window.legacyLiteral {
			// Flattened HTML lists often have no full stop before the next
			// operative clause. Its predicate/subject must not borrow this date.
			for _, boundary := range operativeTransition.FindAllStringIndex(text, -1) {
				if boundary[0] > location[1] && boundary[0] < end {
					end = boundary[0]
				}
				if boundary[0] < location[0] && boundary[0] > start {
					start = boundary[0]
				}
			}
		}
		sentence := lower[start:end]
		if !strings.Contains(sentence, anchor) || !strings.Contains(sentence, predicate) {
			continue
		}
		expression := text[location[0]:location[1]]
		evidence := Evidence{Field: "valid_from", ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: expression, SegmentOrdinal: window.Ordinal}
		measure.TemporalCandidates = append(measure.TemporalCandidates, TemporalCandidate{Field: "valid_from", OriginalExpression: expression, Evidence: []Evidence{evidence}})
		measure.Evidence = append(measure.Evidence, evidence)
	}
}

func temporalSubjectAnchor(subject string) string {
	lower := strings.ToLower(subject)
	for _, anchor := range []string{"cimiter", "sgamb", "sottopass", "ciclopist", "giardin", "parch", "scuol", "attività", "attivita", "orti"} {
		if strings.Contains(lower, anchor) {
			return anchor
		}
	}
	return ""
}

func temporalKindPredicate(kind string) string {
	switch kind {
	case "closure":
		return "chius"
	case "reopening":
		return "riapert"
	case "restriction":
		return "restrizion"
	case "prohibition":
		return "diviet"
	case "suspension":
		return "sospens"
	case "activation":
		return "attiv"
	default:
		return ""
	}
}

func valueInReferencedQuotes(value string, references []int, quotes []string) bool {
	value = strings.ToLower(classification.Normalize(value))
	for _, reference := range references {
		if reference >= 0 && reference < len(quotes) && strings.Contains(strings.ToLower(quotes[reference]), value) {
			return true
		}
	}
	return false
}

func valueInEvidence(value string, evidence []Evidence) bool {
	value = strings.ToLower(classification.Normalize(value))
	for _, item := range evidence {
		if strings.Contains(strings.ToLower(item.Quote), value) {
			return true
		}
	}
	return false
}

func placeScopedToSubject(place, subject string, evidence map[string][]Evidence) bool {
	place = strings.ToLower(classification.Normalize(place))
	subject = strings.ToLower(classification.Normalize(subject))
	for _, item := range evidence["subject"] {
		quote := strings.ToLower(classification.Normalize(item.Quote))
		if sameListedPlaceSubject(quote, place, subject) {
			return true
		}
	}
	for _, item := range evidence["place"] {
		quote := strings.ToLower(classification.Normalize(item.Quote))
		if !strings.Contains(quote, place) {
			continue
		}
		if !strings.Contains(quote, ";") {
			return true
		}
		if sameListedPlaceSubject(quote, place, subject) {
			return true
		}
	}
	return false
}

func sameListedPlaceSubject(quote, place, subject string) bool {
	for _, part := range strings.Split(quote, ";") {
		if strings.Contains(part, place) && strings.Contains(part, subject) {
			return true
		}
	}
	return false
}

func listedSubjectScope(kind, subject string, predicates []Evidence, window Window) (string, Evidence, bool) {
	if kind != "prohibition" || !strings.Contains(strings.ToLower(subject), "attivit") {
		return "", Evidence{}, false
	}
	for _, predicate := range predicates {
		quote := predicate.Quote
		subjectOffset := strings.Index(quote, subject)
		if subjectOffset < 0 {
			continue
		}
		if strings.Contains(quote[subjectOffset:], ";") {
			bounded := quote[subjectOffset:]
			if transition := operativeTransition.FindStringIndex(bounded); transition != nil {
				bounded = bounded[:transition[0]]
			}
			lower := strings.ToLower(bounded)
			for _, marker := range []string{"; - la sospensione", "; - la esposizione", ". "} {
				if at := strings.Index(lower, marker); at >= 0 {
					bounded = bounded[:at]
					lower = lower[:at]
				}
			}
			bounded = strings.TrimSpace(bounded)
			if len(bounded) <= 1000 && bounded != subject && strings.Contains(bounded, ";") {
				return bounded, Evidence{Field: "subject", ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: bounded, SegmentOrdinal: window.Ordinal}, true
			}
		}
		if !strings.HasSuffix(strings.TrimSpace(quote), ":") {
			continue
		}
		start := strings.Index(window.Text, quote)
		if start < 0 {
			continue
		}
		end := len(window.Text)
		remaining := window.Text[start+len(quote):]
		for _, boundary := range operativeTransition.FindAllStringIndex(remaining, -1) {
			if start+len(quote)+boundary[0] < end {
				end = start + len(quote) + boundary[0]
			}
		}
		lower := strings.ToLower(remaining)
		for _, marker := range []string{"; - la sospensione", "; - la esposizione", ". "} {
			if at := strings.Index(lower, marker); at >= 0 && start+len(quote)+at < end {
				end = start + len(quote) + at
			}
		}
		list := strings.TrimSpace(window.Text[start+len(quote) : end])
		if !strings.Contains(list, ";") {
			continue
		}
		bounded := strings.TrimSpace(window.Text[start+subjectOffset : end])
		if len(bounded) > 1000 || bounded == subject {
			continue
		}
		return bounded, Evidence{Field: "subject", ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: bounded, SegmentOrdinal: window.Ordinal}, true
	}
	return "", Evidence{}, false
}

func evidenceWithoutField(evidence []Evidence, field string) []Evidence {
	filtered := evidence[:0]
	for _, item := range evidence {
		if item.Field != field {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func kindSupportedByEvidence(kind string, evidence []Evidence) bool {
	var text strings.Builder
	for _, item := range evidence {
		text.WriteByte(' ')
		text.WriteString(strings.ToLower(item.Quote))
	}
	value := text.String()
	switch kind {
	case "closure":
		return strings.Contains(value, "chius")
	case "reopening":
		return strings.Contains(value, "riapert")
	case "restriction":
		return strings.Contains(value, "restrizion") || strings.Contains(value, "limitazion")
	case "prohibition":
		return strings.Contains(value, "diviet")
	case "suspension":
		return strings.Contains(value, "sospens") || strings.Contains(value, "esposizione delle attività didattiche")
	case "activation":
		return activationPredicate(value)
	case "deactivation":
		return strings.Contains(value, "disattiv") || strings.Contains(value, "cessa")
	case "operational_update":
		return strings.Contains(value, "aggiorn") || strings.Contains(value, "restano") || strings.Contains(value, "liberat") || strings.Contains(value, "ripristinat")
	case "observation":
		return strings.Contains(value, "monitor") || strings.Contains(value, "osserv")
	default:
		return false
	}
}

// Opening an explicitly named civil-protection centre is an operational
// activation; ordinary office openings and negated statements are not.
var civilProtectionOpening = regexp.MustCompile(`(?i)\b(?:coc\s*\(centro operativo comunale\)|centro operativo comunale)\s+(?:sarà|sara|è|resta|rimane)\s+aperto\b`)

func windowKindSupportedByEvidence(kind string, evidence []Evidence, window Window) bool {
	if kind == "activation" && !window.legacyLiteral {
		for _, item := range evidence {
			if civilProtectionOpening.MatchString(item.Quote) {
				return true
			}
		}
	}
	return kindSupportedByEvidence(kind, evidence)
}

func activationPredicate(value string) bool {
	return strings.Contains(value, "attivazion") || strings.Contains(value, "resta attivo") || strings.Contains(value, "resta attiva") || strings.Contains(value, "attivato") || strings.Contains(value, "attivata")
}

func contextualFieldEvidence(field, value string, window Window, kindEvidence []Evidence) (Evidence, bool) {
	canonical, ok := canonicalLiteralReference(window.Content(), window.ResourceURL, windowPage(window), classification.Normalize(value))
	if !ok {
		return Evidence{}, false
	}
	text := window.Text
	for valueStart := 0; valueStart <= len(text)-len(canonical); {
		relative := strings.Index(text[valueStart:], canonical)
		if relative < 0 {
			break
		}
		valueStart += relative
		valueEnd := valueStart + len(canonical)
		for _, operative := range kindEvidence {
			for kindStart := 0; kindStart <= len(text)-len(operative.Quote); {
				relativeKind := strings.Index(text[kindStart:], operative.Quote)
				if relativeKind < 0 {
					break
				}
				kindStart += relativeKind
				kindEnd := kindStart + len(operative.Quote)
				gap := max(kindStart-valueEnd, valueStart-kindEnd)
				if gap <= 512 {
					return Evidence{Field: field, ResourceURL: window.ResourceURL, Page: windowPage(window), Quote: canonical, SegmentOrdinal: window.Ordinal}, true
				}
				kindStart++
			}
		}
		valueStart++
	}
	return Evidence{}, false
}

func windowPage(window Window) *int {
	if window.Page == 0 {
		return nil
	}
	page := window.Page
	return &page
}

func parse(raw string, content classification.Content, segmentOrdinal int) ([]Measure, error) {
	var decoded response
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&decoded) != nil || !errors.Is(decoder.Decode(new(any)), io.EOF) || decoded.Measures == nil || len(decoded.Measures) > 100 {
		return nil, ErrInvalid
	}
	measures := make([]Measure, 0, len(decoded.Measures))
	for index, candidate := range decoded.Measures {
		measure, err := validateMeasure(index+1, candidate, content)
		if err != nil {
			return nil, err
		}
		for evidenceIndex := range measure.Evidence {
			measure.Evidence[evidenceIndex].SegmentOrdinal = segmentOrdinal
		}
		measures = append(measures, measure)
	}
	return measures, nil
}

// Merge combines independently evidence-validated window outputs. Exact
// duplicate facts still reject the merge. Incompatible supported temporal
// values are preserved as candidates under one conflict identity while only
// the affected resolved field becomes indeterminate.
func Merge(groups [][]Measure, content classification.Content, segmentCount int) ([]Measure, error) {
	if !content.Complete || segmentCount < 1 {
		return nil, ErrInvalid
	}
	var merged []Measure
	byIdentity := map[string]int{}
	placesByBase := map[string]map[string]bool{}
	for _, group := range groups {
		for _, rawCandidate := range group {
			candidate := rawCandidate
			if err := normalizeTemporalProjection(&candidate); err != nil {
				return nil, err
			}
			base := candidate.Kind + "\x00" + strings.ToLower(candidate.Subject)
			place := pointerKey(candidate.Place)
			if seen := placesByBase[base]; seen != nil {
				if (place == "\x00" && len(seen) > 0 && !seen[place]) || (place != "\x00" && seen["\x00"]) {
					return nil, ErrMergeConflict
				}
			} else {
				placesByBase[base] = map[string]bool{}
			}
			placesByBase[base][place] = true
			key := base + "\x00" + place
			index, exists := byIdentity[key]
			if !exists {
				candidate.Ordinal = len(merged) + 1
				merged = append(merged, candidate)
				byIdentity[key] = len(merged) - 1
				continue
			}
			current := &merged[index]
			if samePointer(current.ValidFrom, candidate.ValidFrom) && samePointer(current.ValidUntil, candidate.ValidUntil) && sameTemporalValues(current.TemporalCandidates, candidate.TemporalCandidates) {
				// Repeated wording in one operational window can yield the same
				// supported fact more than once. Collapse it and retain all distinct
				// evidence rather than making the complete document uninterpretable.
				current.Evidence = mergeEvidence(current.Evidence, candidate.Evidence)
				current.TemporalCandidates = mergeTemporalCandidates(current.TemporalCandidates, candidate.TemporalCandidates)
				continue
			}
			current.Evidence = mergeEvidence(current.Evidence, candidate.Evidence)
			current.TemporalCandidates = mergeTemporalCandidates(current.TemporalCandidates, candidate.TemporalCandidates)
			if err := projectTemporalCandidates(current); err != nil {
				return nil, err
			}
		}
	}
	if len(merged) > 100 {
		return nil, ErrInvalid
	}
	for index := range merged {
		merged[index].Ordinal = index + 1
		merged[index].IndeterminateFields = indeterminateFields(merged[index])
		if err := validateMergedMeasure(merged[index], content, segmentCount); err != nil {
			return nil, err
		}
	}
	return merged, nil
}

func validateMergedMeasure(measure Measure, content classification.Content, segmentCount int) error {
	if err := validateTemporalProjection(measure); err != nil {
		return err
	}
	temporal := temporalCandidateFields(measure.TemporalCandidates)
	asserted := map[string]bool{"kind": true, "subject": true, "place": measure.Place != nil, "valid_from": temporal["valid_from"], "valid_until": temporal["valid_until"]}
	evidenced := map[string]bool{}
	for _, evidence := range measure.Evidence {
		if evidence.SegmentOrdinal < 1 || evidence.SegmentOrdinal > segmentCount || !asserted[evidence.Field] || !literalReference(content, evidence.ResourceURL, evidence.Page, evidence.Quote) {
			return ErrEvidence
		}
		evidenced[evidence.Field] = true
	}
	for field, value := range asserted {
		if value && !evidenced[field] {
			return ErrEvidence
		}
	}
	for _, candidate := range measure.TemporalCandidates {
		for _, evidence := range candidate.Evidence {
			if evidence.Field != candidate.Field || evidence.SegmentOrdinal < 1 || evidence.SegmentOrdinal > segmentCount || !literalReference(content, evidence.ResourceURL, evidence.Page, evidence.Quote) || !containsEvidence(measure.Evidence, evidence) {
				return ErrEvidence
			}
		}
	}
	return nil
}

func samePointer(left, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func pointerKey(value *string) string {
	if value == nil {
		return "\x00"
	}
	return strings.ToLower(*value)
}

func mergeEvidence(left, right []Evidence) []Evidence {
	result := append([]Evidence(nil), left...)
	seen := map[string]bool{}
	for _, evidence := range result {
		seen[evidenceKey(evidence)] = true
	}
	for _, evidence := range right {
		if key := evidenceKey(evidence); !seen[key] {
			seen[key] = true
			result = append(result, evidence)
		}
	}
	return result
}

func evidenceKey(value Evidence) string {
	page := 0
	if value.Page != nil {
		page = *value.Page
	}
	return value.Field + "\x00" + value.ResourceURL + "\x00" + strconv.Itoa(page) + "\x00" + value.Quote + "\x00" + strconv.Itoa(value.SegmentOrdinal)
}

func temporalCandidatesFromResolved(measure Measure) []TemporalCandidate {
	var candidates []TemporalCandidate
	for _, value := range []struct {
		field string
		value *string
	}{{"valid_from", measure.ValidFrom}, {"valid_until", measure.ValidUntil}} {
		if value.value == nil {
			continue
		}
		candidate := TemporalCandidate{Field: value.field, OriginalExpression: *value.value}
		for _, evidence := range measure.Evidence {
			if evidence.Field == value.field {
				candidate.Evidence = append(candidate.Evidence, evidence)
			}
		}
		candidates = append(candidates, candidate)
	}
	for index := range candidates {
		candidates[index].Ordinal = index + 1
	}
	return candidates
}

func normalizeTemporalProjection(measure *Measure) error {
	if len(measure.TemporalCandidates) == 0 {
		measure.TemporalCandidates = temporalCandidatesFromResolved(*measure)
	}
	return projectTemporalCandidates(measure)
}

func projectTemporalCandidates(measure *Measure) error {
	if measure == nil {
		return ErrInvalid
	}
	if len(measure.TemporalCandidates) > 100 {
		return ErrInvalid
	}
	merged := make([]TemporalCandidate, 0, len(measure.TemporalCandidates))
	byKey := map[string]int{}
	for _, raw := range measure.TemporalCandidates {
		raw.OriginalExpression = classification.Normalize(raw.OriginalExpression)
		if (raw.Field != "valid_from" && raw.Field != "valid_until") || raw.OriginalExpression == "" || len(raw.OriginalExpression) > 1000 || len(raw.Evidence) == 0 || len(raw.Evidence) > 100 {
			return ErrInvalid
		}
		for _, evidence := range raw.Evidence {
			if evidence.Field != raw.Field {
				return ErrEvidence
			}
		}
		key := raw.Field + "\x00" + strings.ToLower(raw.OriginalExpression)
		if index, exists := byKey[key]; exists {
			merged[index].Evidence = mergeEvidence(merged[index].Evidence, raw.Evidence)
			continue
		}
		raw.ConflictIdentity = nil
		byKey[key] = len(merged)
		merged = append(merged, raw)
	}
	measure.TemporalCandidates = merged
	for _, field := range []string{"valid_from", "valid_until"} {
		indices := temporalCandidateIndices(measure.TemporalCandidates, field)
		switch {
		case len(indices) == 0:
			setResolvedTemporal(measure, field, nil)
		case len(indices) == 1:
			value := measure.TemporalCandidates[indices[0]].OriginalExpression
			setResolvedTemporal(measure, field, &value)
		case len(indices) <= 100:
			if field == "valid_from" {
				shortest := measure.TemporalCandidates[indices[0]].OriginalExpression
				compatible := true
				for _, index := range indices[1:] {
					value := measure.TemporalCandidates[index].OriginalExpression
					if len(value) < len(shortest) {
						shortest = value
					}
				}
				for _, index := range indices {
					if !sameStartWithDuration(shortest, measure.TemporalCandidates[index].OriginalExpression) {
						compatible = false
						break
					}
				}
				if compatible {
					setResolvedTemporal(measure, field, &shortest)
					break
				}
			}
			setResolvedTemporal(measure, field, nil)
			identity := temporalConflictIdentity(*measure, field, indices)
			for _, index := range indices {
				value := identity
				measure.TemporalCandidates[index].ConflictIdentity = &value
			}
		default:
			return ErrInvalid
		}
	}
	for index := range measure.TemporalCandidates {
		measure.TemporalCandidates[index].Ordinal = index + 1
	}
	measure.IndeterminateFields = indeterminateFields(*measure)
	return nil
}

func sameStartWithDuration(start, expression string) bool {
	start = strings.ToLower(classification.Normalize(start))
	expression = strings.ToLower(classification.Normalize(expression))
	if start == expression {
		return true
	}
	if !strings.HasPrefix(expression, start) {
		return false
	}
	suffix := strings.TrimSpace(expression[len(start):])
	return strings.HasPrefix(suffix, "e fino ") || strings.HasPrefix(suffix, ", fino ") || strings.HasPrefix(suffix, "fino ")
}

func validateTemporalProjection(measure Measure) error {
	copy := measure
	if err := projectTemporalCandidates(&copy); err != nil {
		return err
	}
	if !samePointer(copy.ValidFrom, measure.ValidFrom) || !samePointer(copy.ValidUntil, measure.ValidUntil) || !sameTemporalCandidates(copy.TemporalCandidates, measure.TemporalCandidates) {
		return ErrInvalid
	}
	return nil
}

func temporalCandidateIndices(values []TemporalCandidate, field string) []int {
	var indices []int
	for index, candidate := range values {
		if candidate.Field == field {
			indices = append(indices, index)
		}
	}
	return indices
}

func temporalCandidateFields(values []TemporalCandidate) map[string]bool {
	result := map[string]bool{}
	for _, candidate := range values {
		result[candidate.Field] = true
	}
	return result
}

func setResolvedTemporal(measure *Measure, field string, value *string) {
	if field == "valid_from" {
		measure.ValidFrom = value
	} else {
		measure.ValidUntil = value
	}
}

func temporalConflictIdentity(measure Measure, field string, indices []int) string {
	values := make([]string, 0, len(indices))
	for _, index := range indices {
		values = append(values, strings.ToLower(measure.TemporalCandidates[index].OriginalExpression))
	}
	sort.Strings(values)
	hash := sha256.New()
	for _, value := range append([]string{measure.Kind, strings.ToLower(measure.Subject), pointerKey(measure.Place), field}, values...) {
		hash.Write([]byte(value))
		hash.Write([]byte{0})
	}
	return "temporal-conflict-" + hex.EncodeToString(hash.Sum(nil))[:24]
}

func mergeTemporalCandidates(left, right []TemporalCandidate) []TemporalCandidate {
	result := append([]TemporalCandidate(nil), left...)
	for _, candidate := range right {
		key := candidate.Field + "\x00" + strings.ToLower(candidate.OriginalExpression)
		found := false
		for index := range result {
			if result[index].Field+"\x00"+strings.ToLower(result[index].OriginalExpression) == key {
				result[index].Evidence = mergeEvidence(result[index].Evidence, candidate.Evidence)
				found = true
				break
			}
		}
		if !found {
			result = append(result, candidate)
		}
	}
	return result
}

func sameTemporalCandidates(left, right []TemporalCandidate) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Ordinal != right[index].Ordinal || left[index].Field != right[index].Field || left[index].OriginalExpression != right[index].OriginalExpression || !samePointer(left[index].ConflictIdentity, right[index].ConflictIdentity) || len(left[index].Evidence) != len(right[index].Evidence) {
			return false
		}
		for evidenceIndex := range left[index].Evidence {
			if evidenceKey(left[index].Evidence[evidenceIndex]) != evidenceKey(right[index].Evidence[evidenceIndex]) {
				return false
			}
		}
	}
	return true
}

func sameTemporalValues(left, right []TemporalCandidate) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Field != right[index].Field || !strings.EqualFold(left[index].OriginalExpression, right[index].OriginalExpression) {
			return false
		}
	}
	return true
}

func containsEvidence(values []Evidence, wanted Evidence) bool {
	key := evidenceKey(wanted)
	for _, value := range values {
		if evidenceKey(value) == key {
			return true
		}
	}
	return false
}

func indeterminateFields(value Measure) []string {
	var fields []string
	if value.Place == nil {
		fields = append(fields, "place")
	}
	if value.ValidFrom == nil {
		fields = append(fields, "valid_from")
	}
	if value.ValidUntil == nil {
		fields = append(fields, "valid_until")
	}
	return fields
}

func validateMeasure(ordinal int, candidate wireMeasure, content classification.Content) (Measure, error) {
	validKind := map[string]bool{"closure": true, "reopening": true, "restriction": true, "prohibition": true, "suspension": true, "activation": true, "deactivation": true, "operational_update": true, "observation": true}
	candidate.Subject = classification.Normalize(candidate.Subject)
	if !validKind[candidate.Kind] || candidate.Subject == "" || len(candidate.Subject) > 1000 || len(candidate.Evidence) < 2 || len(candidate.Evidence) > 100 {
		return Measure{}, ErrInvalid
	}
	optional := map[string]*string{"place": cleanPointer(candidate.Place), "valid_from": cleanPointer(candidate.ValidFrom), "valid_until": cleanPointer(candidate.ValidUntil)}
	unknown := map[string]bool{}
	for _, field := range candidate.IndeterminateFields {
		if _, ok := optional[field]; !ok || unknown[field] {
			return Measure{}, ErrInvalid
		}
		unknown[field] = true
	}
	for field, value := range optional {
		if (value == nil) == !unknown[field] {
			return Measure{}, ErrInvalid
		}
	}
	asserted := map[string]bool{"kind": true, "subject": true}
	for field, value := range optional {
		asserted[field] = value != nil
	}
	evidenced := map[string]bool{}
	evidence := make([]Evidence, 0, len(candidate.Evidence))
	for _, reference := range candidate.Evidence {
		quote, ok := canonicalLiteralReference(content, reference.ResourceURL, reference.Page, classification.Normalize(reference.Quote))
		if content.LegacyLiteral {
			quote = classification.Normalize(reference.Quote)
		}
		if !asserted[reference.Field] || quote == "" || reference.ResourceURL == "" || !ok {
			return Measure{}, ErrEvidence
		}
		evidenced[reference.Field] = true
		evidence = append(evidence, Evidence{Field: reference.Field, ResourceURL: reference.ResourceURL, Page: reference.Page, Quote: quote})
	}
	for field, value := range asserted {
		if value && !evidenced[field] {
			return Measure{}, ErrEvidence
		}
	}
	measure := Measure{Ordinal: ordinal, Kind: candidate.Kind, Subject: candidate.Subject, Place: optional["place"], ValidFrom: optional["valid_from"], ValidUntil: optional["valid_until"], IndeterminateFields: append([]string(nil), candidate.IndeterminateFields...), Evidence: evidence}
	measure.TemporalCandidates = temporalCandidatesFromResolved(measure)
	return measure, nil
}

func literalReference(content classification.Content, raw string, page *int, quote string) bool {
	_, ok := canonicalLiteralReference(content, raw, page, quote)
	return ok
}

func canonicalLiteralReference(content classification.Content, raw string, page *int, quote string) (string, bool) {
	for _, section := range content.Sections {
		if section.ResourceURL != raw || page == nil && section.Page != 0 || page != nil && section.Page != *page {
			continue
		}
		if content.LegacyLiteral {
			text := classification.Normalize(section.Text)
			if index := strings.Index(strings.ToLower(text), strings.ToLower(quote)); index >= 0 && index+len(quote) <= len(text) {
				return text[index : index+len(quote)], true
			}
			continue
		}
		if literal, ok := classification.CanonicalLiteral(section.Text, quote); ok {
			return literal, true
		}
	}
	return "", false
}

func cleanPointer(value *string) *string {
	if value == nil {
		return nil
	}
	clean := classification.Normalize(*value)
	if clean == "" || len(clean) > 1000 {
		return nil
	}
	return &clean
}

// LocalWindowRequest records the evaluated local instructions before hashing.
func LocalWindowRequest(model string, window Window) (inference.Request, error) {
	request, err := WindowRequest(model, window)
	if err != nil {
		return inference.Request{}, err
	}
	request.Messages[0].Content, _ = json.Marshal(LocalPromptBody)
	return inference.LocalChatRequest(request), nil
}
