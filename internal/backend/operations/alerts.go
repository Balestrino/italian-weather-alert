package operations

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publicquery"
)

type AlertBulletin struct {
	Source, Product, URL string
	Version              int64
	AcquiredAt           time.Time
	LastChecked          *time.Time
	Issued               string
	Validity             []string
	NoCriticalityText    bool
}
type AlertDay struct {
	Label, Date string
	Start, End  time.Time
	Bulletins   []AlertBulletin
	Regional    []publicquery.RegionalWarning
	Measures    []publicquery.Measure
}
type Alerts struct {
	ObservedAt       time.Time
	Days             []AlertDay
	UndatedBulletins []AlertBulletin
	UndatedRegional  []publicquery.RegionalWarning
	UndatedMeasures  []publicquery.Measure
	FactsUnavailable bool
}

var rome, _ = time.LoadLocation("Europe/Rome")
var italianDate = regexp.MustCompile(`(?i)^(?:(?:luned[ìíi]|marted[ìíi]|mercoled[ìíi]|gioved[ìíi]|venerd[ìíi]|sabato|domenica),?\s+)?([0-9]{1,2})\s+([\p{L}]+)\s+([0-9]{4})$`)

func bulletinDate(text string) (string, bool) {
	m := italianDate.FindStringSubmatch(strings.TrimSpace(text))
	if len(m) != 4 {
		return "", false
	}
	months := map[string]time.Month{"gennaio": 1, "febbraio": 2, "marzo": 3, "aprile": 4, "maggio": 5, "giugno": 6, "luglio": 7, "agosto": 8, "settembre": 9, "ottobre": 10, "novembre": 11, "dicembre": 12}
	month, ok := months[strings.ToLower(m[2])]
	if !ok {
		return "", false
	}
	d, _ := strconv.Atoi(m[1])
	y, _ := strconv.Atoi(m[3])
	t := time.Date(y, month, d, 0, 0, 0, 0, rome)
	if t.Day() != d || t.Month() != month {
		return "", false
	}
	return t.Format(time.DateOnly), true
}
func newAlerts(at time.Time) Alerts {
	local := at.In(rome)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, rome)
	r := Alerts{ObservedAt: local}
	for i, label := range []string{"Oggi", "Domani"} {
		s := start.AddDate(0, 0, i)
		r.Days = append(r.Days, AlertDay{Label: label, Date: s.Format("02/01/2006"), Start: s, End: s.AddDate(0, 0, 1)})
	}
	return r
}

// Only explicit dates/bounded intervals are assigned. Unknown and conditional
// validity are not claims of current or future applicability.
func temporalDay(v publicquery.Temporal, d AlertDay) (matches, known bool) {
	if v.Precision == "date" && v.Date != nil {
		_, e := time.Parse(time.DateOnly, *v.Date)
		if e != nil {
			return false, false
		}
		return *v.Date == d.Start.Format(time.DateOnly), true
	}
	if v.Precision == "interval" && v.Instant != nil && v.EndInstant != nil && !v.EndInstant.Before(*v.Instant) {
		return v.Instant.Before(d.End) && v.EndInstant.After(d.Start), true
	}
	if v.Precision == "instant" && v.Instant != nil {
		return !v.Instant.Before(d.Start) && v.Instant.Before(d.End), true
	}
	return false, false
}
func (s *Store) Alerts(ctx context.Context, at time.Time) (Alerts, error) {
	result := newAlerts(at)
	rows, err := s.pool.Query(ctx, `SELECT s.id,s.product_id,d.official_url,v.id,v.first_acquired_at,v.metadata,st.last_complete_at FROM registry_sources s JOIN retained_documents d ON d.source_id=s.id JOIN LATERAL(SELECT id,first_acquired_at,metadata FROM retained_versions WHERE document_id=d.id AND first_acquired_at<=$1 ORDER BY first_acquired_at DESC,id DESC LIMIT 1)v ON true LEFT JOIN acquisition_source_status st ON st.source_id=s.id WHERE s.product_id IN ('criticality','vigilance','monitoring') AND s.collection_enabled ORDER BY s.id,d.id`, at)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var b AlertBulletin
		var raw []byte
		if err = rows.Scan(&b.Source, &b.Product, &b.URL, &b.Version, &b.AcquiredAt, &raw, &b.LastChecked); err != nil {
			rows.Close()
			return result, err
		}
		var meta acquisition.RegionalObservation
		if json.Unmarshal(raw, &meta) != nil {
			rows.Close()
			return result, ErrInvalid
		}
		b.Issued = meta.IssuanceExpression
		b.Validity = meta.ValidityExpressions
		b.NoCriticalityText = meta.TextSaysNoCriticality
		b.AcquiredAt = b.AcquiredAt.In(rome)
		if b.LastChecked != nil {
			v := b.LastChecked.In(rome)
			b.LastChecked = &v
		}
		known := false
		for i := range result.Days {
			matched := false
			for _, expression := range b.Validity {
				date, ok := bulletinDate(expression)
				known = known || ok
				if ok && date == result.Days[i].Start.Format(time.DateOnly) {
					matched = true
				}
			}
			if matched {
				result.Days[i].Bulletins = append(result.Days[i].Bulletins, b)
			}
		}
		if !known {
			result.UndatedBulletins = append(result.UndatedBulletins, b)
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	query := publicquery.New(s.pool)
	for _, kind := range []string{"regional", "measure"} {
		values, e := query.Search(ctx, publicquery.SearchQuery{QueryTime: publicquery.QueryTime{EvaluationTime: at, KnownAt: at}, Kind: kind, From: &result.Days[0].Start, To: &result.Days[1].End})
		if e != nil {
			result.FactsUnavailable = true
			continue
		}
		for _, v := range values.Regional {
			if v.Status == "cancelled" || v.Status == "superseded" {
				continue
			}
			known := false
			for i := range result.Days {
				match, k := temporalDay(v.Validity, result.Days[i])
				known = known || k
				if match {
					result.Days[i].Regional = append(result.Days[i].Regional, v)
				}
			}
			if !known {
				result.UndatedRegional = append(result.UndatedRegional, v)
			}
		}
		for _, v := range values.Measures {
			if v.Status == "cancelled" || v.Status == "superseded" {
				continue
			}
			known := false
			for i := range result.Days {
				match, k := temporalDay(v.Validity, result.Days[i])
				known = known || k
				if match {
					result.Days[i].Measures = append(result.Days[i].Measures, v)
				}
			}
			if !known {
				result.UndatedMeasures = append(result.UndatedMeasures, v)
			}
		}
	}
	return result, nil
}
