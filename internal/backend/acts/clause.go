package acts

import (
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
)

// ParserVersion identifies the reviewed dispositive-clause patterns below.
const ParserVersion = "act-dispositive-clause-v1"

var ErrClauseUnrecognized = errors.New("dispositive clause not recognized")

// Clause is the dispositive part of a municipal emergency ordinance: one
// start, an end instant or ending condition, and the listed dispositions.
type Clause struct {
	Start        time.Time
	End          *time.Time
	Condition    string
	Passage      string
	ResourceURL  string
	Page         int
	Dispositions []Disposition
}

type Disposition struct {
	Ordinal int
	Kind    string
	Text    string
}

const italianMonths = `(gennaio|febbraio|marzo|aprile|maggio|giugno|luglio|agosto|settembre|ottobre|novembre|dicembre)`

var (
	// "ORDINA [in via contingibile e urgente,] a partire dalle ore 00.00 del giorno 10 settembre 2026, fino al perdurare dell'emergenza:"
	clauseHeader     = regexp.MustCompile(`(?s)\bORDINA\b\s*(?:in via contingibile e urgente\s*,?\s*)?a partire dalle ore ([0-9]{1,2})[.:]([0-9]{2}) del giorno ([0-9]{1,2}) ` + italianMonths + ` ([0-9]{4})\s*,?\s*(fino (?:al|alle)(?:[^:]|[0-9]:[0-9]){1,160}):\s`)
	clauseEndInstant = regexp.MustCompile(`^fino alle ore ([0-9]{1,2})[.:]([0-9]{2}) del giorno ([0-9]{1,2}) ` + italianMonths + ` ([0-9]{4})$`)
	clauseCondition  = regexp.MustCompile(`(?i)^fino al (?:perdurare|termine|cessare|cessazione) dell['’]emergenza$|^fino a (?:nuova disposizione|revoca)$`)
	clauseEnd        = regexp.MustCompile(`\b(?:AVVERTE|DISPONE|DEMANDA|INFORMA|AVVISA|Adempimenti)\b`)
	dispositionStart = regexp.MustCompile(`(?i)\b(?:la|il|lo|l['’])\s*(chiusura|divieto|sospensione|attivazione|limitazion[ei]|riapertura)\b`)
)

var dispositionKinds = map[string]string{"chiusura": "closure", "divieto": "prohibition", "sospensione": "suspension", "attivazione": "activation", "limitazione": "restriction", "limitazioni": "restriction", "riapertura": "reopening"}

// ParseClause reads the dispositive clause from the act's sections. Only the
// part after ORDINA counts: recitals repeat the dispositions without force.
func ParseClause(sections []classification.ContentSection) (Clause, error) {
	var text strings.Builder
	type span struct {
		start   int
		section classification.ContentSection
	}
	var spans []span
	for _, section := range sections {
		spans = append(spans, span{text.Len(), section})
		text.WriteString(section.Text)
		text.WriteByte('\n')
	}
	all := text.String()
	m := clauseHeader.FindStringSubmatchIndex(all)
	if m == nil {
		return Clause{}, ErrClauseUnrecognized
	}
	at := func(i int) string { return all[m[2*i]:m[2*i+1]] }
	start, ok := romeInstant(at(1), at(2), at(3), at(4), at(5))
	if !ok {
		return Clause{}, ErrClauseUnrecognized
	}
	clause := Clause{Start: start, Passage: strings.Join(strings.Fields(all[m[0]:m[1]]), " ")}
	for _, s := range spans {
		if s.start <= m[0] {
			clause.ResourceURL, clause.Page = s.section.ResourceURL, s.section.Page
		}
	}
	end := strings.Join(strings.Fields(at(6)), " ")
	switch {
	case clauseEndInstant.MatchString(end):
		e := clauseEndInstant.FindStringSubmatch(end)
		instant, ok := romeInstant(e[1], e[2], e[3], e[4], e[5])
		if !ok || instant.Before(start) {
			return Clause{}, ErrClauseUnrecognized
		}
		clause.End = &instant
	case clauseCondition.MatchString(end):
		clause.Condition = end
	default:
		return Clause{}, ErrClauseUnrecognized
	}
	body := all[m[1]:]
	if stop := clauseEnd.FindStringIndex(body); stop != nil {
		body = body[:stop[0]]
	}
	starts := dispositionStart.FindAllStringSubmatchIndex(body, -1)
	for i, d := range starts {
		to := len(body)
		if i+1 < len(starts) {
			to = starts[i+1][0]
		}
		clause.Dispositions = append(clause.Dispositions, Disposition{Ordinal: i + 1, Kind: dispositionKinds[strings.ToLower(body[d[2]:d[3]])], Text: strings.Join(strings.Fields(body[d[0]:to]), " ")})
	}
	if len(clause.Dispositions) == 0 {
		return Clause{}, ErrClauseUnrecognized
	}
	return clause, nil
}

func romeInstant(hour, minute, day, month, year string) (time.Time, bool) {
	h, _ := strconv.Atoi(hour)
	mi, _ := strconv.Atoi(minute)
	d, _ := strconv.Atoi(day)
	y, _ := strconv.Atoi(year)
	mo := slices.Index(strings.Fields("gennaio febbraio marzo aprile maggio giugno luglio agosto settembre ottobre novembre dicembre"), strings.ToLower(month)) + 1
	location, err := time.LoadLocation("Europe/Rome")
	if err != nil || h > 23 || mi > 59 || mo < 1 {
		return time.Time{}, false
	}
	t := time.Date(y, time.Month(mo), d, h, mi, 0, 0, location)
	return t, t.Day() == d && int(t.Month()) == mo
}

var headStopwords = map[string]bool{"della": true, "delle": true, "dello": true, "degli": true, "dell": true, "dei": true, "del": true, "nelle": true, "negli": true, "nella": true, "tutte": true, "tutti": true, "qualunque": true, "qualsiasi": true, "ogni": true}

// HeadTerm is the stem of the first content word of a measure subject.
func HeadTerm(subject string) string {
	for _, word := range strings.FieldsFunc(strings.ToLower(subject), func(r rune) bool { return !unicode.IsLetter(r) }) {
		if len([]rune(word)) >= 4 && !headStopwords[word] {
			return stem(word)
		}
	}
	return ""
}

// Match returns the disposition of the measure's kind whose text contains
// the measure subject's head term, or nil when there is none or several.
func (c Clause) Match(kind, subject string) *Disposition {
	head := HeadTerm(subject)
	if head == "" {
		return nil
	}
	var found *Disposition
	for i, d := range c.Dispositions {
		if d.Kind != kind {
			continue
		}
		for _, word := range strings.FieldsFunc(strings.ToLower(d.Text), func(r rune) bool { return !unicode.IsLetter(r) }) {
			if len([]rune(word)) >= 4 && stem(word) == head {
				if found != nil && found.Ordinal != d.Ordinal {
					return nil
				}
				found = &c.Dispositions[i]
				break
			}
		}
	}
	return found
}
