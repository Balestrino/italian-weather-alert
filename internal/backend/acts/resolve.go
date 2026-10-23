package acts

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

// Albo is the bounded albo access a resolution needs.
type Albo interface {
	Window(ctx context.Context, from, to time.Time, bound int) ([]acquisition.AlboRecord, error)
	Detail(ctx context.Context, recordID string) (acquisition.AlboDetail, []byte, error)
}

// Notice is what resolution knows about the citing notice version.
type Notice struct {
	PublicationDate time.Time // date only; zero when the source states none
	Content         classification.Content
	Measures        []NoticeMeasure
}

type NoticeMeasure struct {
	ID, Kind, Subject string
}

type Candidate struct {
	RecordID       string `json:"record_id"`
	Category       string `json:"category"`
	ActNumber      int    `json:"act_number,omitempty"`
	ActYear        int    `json:"act_year,omitempty"`
	ActDate        string `json:"act_date,omitempty"`
	RegistryYear   int    `json:"registry_year"`
	RegistryNumber int    `json:"registry_number"`
	Object         string `json:"object"`
}

type Resolution struct {
	Outcome          string
	AlboRecordID     string
	AttachedResource string
	Candidates       []Candidate
	Reason           string
	Detail           *acquisition.AlboDetail
	DetailMarkup     []byte
}

const (
	unnumberedWindowDays = 7
	numberedLookbackDays = 30
	numberedDateDays     = 10
	maxCandidateDetails  = 3
)

var attachedAct = regexp.MustCompile(`(?i)\b` + kindWords + `\s*(?:n\.|nr\.?|n°|numero)\s*0*([0-9]{1,5})\s*/\s*([0-9]{4})`)

// Resolve applies the order attached, numbered, unique_candidate. contract
// and albo are nil when the municipality has no configured albo.
func Resolve(ctx context.Context, ref Reference, notice Notice, contract *registry.JCityGovAlboContract, albo Albo) Resolution {
	if r, ok := resolveAttached(ref, notice.Content); ok {
		return r
	}
	if contract == nil || albo == nil {
		return Resolution{Outcome: "not_found", Reason: "no_configured_albo"}
	}
	category, ok := contract.Categories[ref.Kind]
	if !ok {
		return Resolution{Outcome: "not_found", Reason: "kind_not_published_in_albo"}
	}
	if ref.Number > 0 {
		return resolveNumbered(ctx, ref, notice, contract, albo, category)
	}
	return resolveUnnumbered(ctx, notice, contract, albo, category)
}

func resolveAttached(ref Reference, content classification.Content) (Resolution, bool) {
	var matches []string
	for _, section := range content.Sections {
		if section.Role != "attachment" || slices.Contains(matches, section.ResourceURL) {
			continue
		}
		head := section.Text
		if len(head) > 1200 {
			head = head[:1200]
		}
		m := attachedAct.FindStringSubmatch(head)
		if m == nil || canonicalKind(m[1]) != ref.Kind {
			continue
		}
		number, _ := strconv.Atoi(m[2])
		year, _ := strconv.Atoi(m[3])
		if ref.Number > 0 && (number != ref.Number || ref.Year > 0 && year != ref.Year) {
			continue
		}
		matches = append(matches, section.ResourceURL)
	}
	if len(matches) != 1 {
		return Resolution{}, false
	}
	return Resolution{Outcome: "attached", AttachedResource: matches[0]}, true
}

func unavailable(err error) Resolution {
	reason := "albo_unavailable"
	switch {
	case errors.Is(err, acquisition.ErrAlboMarkup):
		reason = "albo_markup_not_recognized"
	case errors.Is(err, acquisition.ErrAlboBound):
		reason = "albo_page_bound_exceeded"
	}
	return Resolution{Outcome: "unavailable", Reason: reason + ": " + err.Error()}
}

func resolveNumbered(ctx context.Context, ref Reference, notice Notice, contract *registry.JCityGovAlboContract, albo Albo, category string) Resolution {
	var from, to time.Time
	switch {
	case ref.Date != nil:
		from, to = *ref.Date, ref.Date.AddDate(0, 0, numberedDateDays)
	case !notice.PublicationDate.IsZero():
		from, to = notice.PublicationDate.AddDate(0, 0, -numberedLookbackDays), notice.PublicationDate
	default:
		return Resolution{Outcome: "not_found", Reason: "no_date_to_bound_the_search"}
	}
	records, err := albo.Window(ctx, from, to, contract.PageBound)
	if err != nil {
		return unavailable(err)
	}
	var found []acquisition.AlboRecord
	for _, r := range records {
		if r.Category == category && r.ActNumber == ref.Number && (ref.Year == 0 || r.ActYear == ref.Year) {
			found = append(found, r)
		}
	}
	return choose(ctx, albo, found, func(acquisition.AlboDetail) bool { return true }, "numbered")
}

func resolveUnnumbered(ctx context.Context, notice Notice, contract *registry.JCityGovAlboContract, albo Albo, category string) Resolution {
	if notice.PublicationDate.IsZero() {
		return Resolution{Outcome: "not_found", Reason: "notice_publication_date_unknown"}
	}
	terms := DispositionTerms(notice.Measures)
	if len(terms) == 0 {
		return Resolution{Outcome: "not_found", Reason: "no_disposition_terms"}
	}
	from, to := notice.PublicationDate.AddDate(0, 0, -unnumberedWindowDays), notice.PublicationDate
	// Acts are registered on or after their date; read publications up to the
	// numbered-date margin after the notice, then filter on the act date.
	records, err := albo.Window(ctx, from, to.AddDate(0, 0, numberedDateDays), contract.PageBound)
	if err != nil {
		return unavailable(err)
	}
	var found []acquisition.AlboRecord
	for _, r := range records {
		if r.Category == category && objectMatches(r.Object, terms) {
			found = append(found, r)
		}
	}
	inWindow := func(d acquisition.AlboDetail) bool {
		return !d.ActDate.IsZero() && !d.ActDate.Before(from) && !d.ActDate.After(to)
	}
	return choose(ctx, albo, found, inWindow, "unique_candidate")
}

// choose confirms candidates through their details and requires exactly one.
func choose(ctx context.Context, albo Albo, found []acquisition.AlboRecord, accept func(acquisition.AlboDetail) bool, outcome string) Resolution {
	if len(found) == 0 {
		return Resolution{Outcome: "not_found", Reason: "no_matching_albo_record"}
	}
	if len(found) > maxCandidateDetails {
		r := Resolution{Outcome: "ambiguous", Reason: "too_many_matching_albo_records"}
		for _, f := range found {
			r.Candidates = append(r.Candidates, candidate(f))
		}
		return r
	}
	var accepted []acquisition.AlboDetail
	var markup [][]byte
	var candidates []Candidate
	for _, f := range found {
		detail, raw, err := albo.Detail(ctx, f.RecordID)
		if err != nil {
			return unavailable(err)
		}
		if accept(detail) {
			accepted = append(accepted, detail)
			markup = append(markup, raw)
			candidates = append(candidates, candidate(detail.AlboRecord))
		}
	}
	switch len(accepted) {
	case 0:
		return Resolution{Outcome: "not_found", Reason: "no_albo_record_in_act_date_window"}
	case 1:
		return Resolution{Outcome: outcome, AlboRecordID: accepted[0].RecordID, Candidates: candidates, Detail: &accepted[0], DetailMarkup: markup[0]}
	default:
		return Resolution{Outcome: "ambiguous", Reason: "several_matching_albo_records", Candidates: candidates}
	}
}

func candidate(r acquisition.AlboRecord) Candidate {
	c := Candidate{RecordID: r.RecordID, Category: r.Category, ActNumber: r.ActNumber, ActYear: r.ActYear, RegistryYear: r.RegistryYear, RegistryNumber: r.RegistryNumber, Object: r.Object}
	if !r.ActDate.IsZero() {
		c.ActDate = r.ActDate.Format(time.DateOnly)
	}
	return c
}

// kindKeywords are already stems (see stem) of each kind's Italian noun.
var kindKeywords = map[string]string{"closure": "chiusu", "reopening": "riaper", "restriction": "limita", "prohibition": "diviet", "suspension": "sospen", "activation": "attiva"}

var termStopwords = map[string]bool{"della": true, "delle": true, "degli": true, "dello": true, "nelle": true, "negli": true, "tutte": true, "tutti": true, "ogni": true, "grado": true, "ordine": true, "comunale": true, "comunali": true, "pubblico": true, "pubblici": true, "pubbliche": true, "pubblica": true, "territorio": true}

// DispositionTerms are lower-case word stems of the notice's projected
// measures: each kind's keyword and the content words of each subject.
func DispositionTerms(measures []NoticeMeasure) []string {
	var terms []string
	add := func(t string) {
		if t != "" && !slices.Contains(terms, t) {
			terms = append(terms, t)
		}
	}
	for _, m := range measures {
		add(kindKeywords[m.Kind])
		for _, word := range strings.FieldsFunc(strings.ToLower(m.Subject), func(r rune) bool { return !unicode.IsLetter(r) }) {
			if len([]rune(word)) >= 5 && !termStopwords[word] {
				add(stem(word))
			}
		}
	}
	return terms
}

func stem(word string) string {
	r := []rune(word)
	if len(r) > 6 {
		r = r[:6]
	}
	return string(r)
}

func objectMatches(object string, terms []string) bool {
	for _, word := range strings.FieldsFunc(strings.ToLower(object), func(r rune) bool { return !unicode.IsLetter(r) }) {
		if len([]rune(word)) >= 5 && slices.Contains(terms, stem(word)) {
			return true
		}
	}
	return false
}
