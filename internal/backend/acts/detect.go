package acts

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
)

// DetectionVersion identifies the reviewed reference patterns below.
const DetectionVersion = "act-reference-detection-v1"

// MaxReferences bounds the references recorded for one notice version.
const MaxReferences = 3

// Reference is an explicit act reference found in a notice's own text.
type Reference struct {
	Ordinal     int
	Kind        string
	Number      int
	Year        int
	Date        *time.Time
	Passage     string
	ResourceURL string
	Page        int
}

const kindWords = `(ordinanz[ae]|decret[oi]|delibera(?:zion[ei])?|determin[ae](?:zion[ei])?)`

var (
	numberedReference = regexp.MustCompile(`(?i)\b` + kindWords + `(?:\s+(?:sindacale|comunale|dirigenziale|del sindaco|della polizia locale|di giunta(?: comunale)?|del consiglio comunale|contingibile e urgente))?\s*(?:n\.|nr\.?|n°|numero)\s*0*([0-9]{1,5})(?:\s*/\s*([0-9]{4}))?(?:\s+del(?:l['’])?\s+([0-9]{1,2})[/.]([0-9]{1,2})[/.]([0-9]{4}))?`)
	// Unnumbered forms only in operative sentences of the notice.
	unnumberedReference = regexp.MustCompile(`(?i)\b(?:emess[ao]|emanat[ao]|adottat[ao]|firmat[ao]|apposit[ao]|mediante(?:\s+apposit[ao])?|con(?:\s+apposit[ao])?)\s+` + kindWords + `\b|\b` + kindWords + `\s+(?:sindacale|comunale|del sindaco)\b`)
	negatedReference    = regexp.MustCompile(`(?i)\b(?:non|nessun[ao]?|alcun[ao]?|eventual[ei]|potr[àa]|potranno|sar[àa] emess[ao]|qualora|se necessari[ao])\b[^.;:]{0,60}$`)
	sentenceBoundary    = regexp.MustCompile(`[.;!?]\s|\n`)
)

func canonicalKind(word string) string {
	switch w := strings.ToLower(word); {
	case strings.HasPrefix(w, "ordinanz"):
		return "ordinanza"
	case strings.HasPrefix(w, "decret"):
		return "decreto"
	case strings.HasPrefix(w, "delibera"):
		return "delibera"
	default:
		return "determina"
	}
}

// Detect returns the explicit act references of a notice's original sections,
// in order of appearance. Attachments are candidates for resolution, not
// sources of references, so citations inside an attached act are ignored.
func Detect(content classification.Content) []Reference {
	var numbered, unnumbered []Reference
	seen := map[string]bool{}
	for _, section := range content.Sections {
		if section.Role != "original" {
			continue
		}
		text := section.Text
		for _, m := range numberedReference.FindAllStringSubmatchIndex(text, -1) {
			if negatedReference.MatchString(text[sentenceStart(text, m[0]):m[0]]) {
				continue
			}
			r := Reference{Kind: canonicalKind(text[m[2]:m[3]]), Passage: passage(text, m[0], m[1]), ResourceURL: section.ResourceURL, Page: section.Page}
			r.Number, _ = strconv.Atoi(text[m[4]:m[5]])
			if m[6] >= 0 {
				r.Year, _ = strconv.Atoi(text[m[6]:m[7]])
			}
			if m[8] >= 0 {
				day, _ := strconv.Atoi(text[m[8]:m[9]])
				month, _ := strconv.Atoi(text[m[10]:m[11]])
				year, _ := strconv.Atoi(text[m[12]:m[13]])
				date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
				if date.Day() == day && int(date.Month()) == month {
					r.Date = &date
					if r.Year == 0 {
						r.Year = year
					}
				}
			}
			key := r.Kind + ":" + strconv.Itoa(r.Number) + "/" + strconv.Itoa(r.Year)
			if !seen[key] {
				seen[key] = true
				numbered = append(numbered, r)
			}
		}
		for _, m := range unnumberedReference.FindAllStringSubmatchIndex(text, -1) {
			word, wordEnd := m[2], m[3]
			if word < 0 {
				word, wordEnd = m[4], m[5]
			}
			if negatedReference.MatchString(text[sentenceStart(text, m[0]):m[0]]) || numberedReference.MatchString(text[word:min(len(text), m[1]+40)]) {
				continue
			}
			kind := canonicalKind(text[word:wordEnd])
			if !seen["unnumbered:"+kind] {
				seen["unnumbered:"+kind] = true
				unnumbered = append(unnumbered, Reference{Kind: kind, Passage: passage(text, m[0], m[1]), ResourceURL: section.ResourceURL, Page: section.Page})
			}
		}
	}
	// A numbered reference identifies the act an unnumbered one of the same kind names.
	var result []Reference
	result = append(result, numbered...)
	for _, r := range unnumbered {
		covered := false
		for _, n := range numbered {
			covered = covered || n.Kind == r.Kind
		}
		if !covered {
			result = append(result, r)
		}
	}
	if len(result) > MaxReferences {
		result = result[:MaxReferences]
	}
	for i := range result {
		result[i].Ordinal = i + 1
	}
	return result
}

func sentenceStart(text string, at int) int {
	start := 0
	for _, m := range sentenceBoundary.FindAllStringIndex(text[:at], -1) {
		start = m[1]
	}
	return start
}

// passage returns the sentence around a match, bounded to 400 bytes.
func passage(text string, from, to int) string {
	start := sentenceStart(text, from)
	end := len(text)
	if m := sentenceBoundary.FindStringIndex(text[to:]); m != nil {
		end = to + m[0] + 1
	}
	if from-start > 200 {
		start = from - 200
	}
	if end-to > 200 {
		end = to + 200
	}
	for start > 0 && !utf8Start(text[start]) {
		start--
	}
	for end < len(text) && !utf8Start(text[end]) {
		end++
	}
	return strings.TrimSpace(text[start:end])
}

func utf8Start(b byte) bool { return b&0xC0 != 0x80 }
