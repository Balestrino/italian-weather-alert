package acquisition

import (
	"fmt"
	"golang.org/x/net/html"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

// RegionalFact is an evidence-bound row, not an OCR interpretation of a map.
type RegionalFact struct {
	Risk, Label, Zone, Level, Original, Precision, Locator string
	Date, Start, End                                       *time.Time
	EvidenceURL                                            string
	Page                                                   int
	Weather                                                *VigilanceWeather
}

// VigilanceWeather describes depicted meteorology, never a warning color or
// an assertion of absence of risk. Amounts are the literal area-average bands.
type VigilanceWeather struct {
	Phenomenon           string `json:"phenomenon"`
	GraphicalStatus      string `json:"graphical_status"`
	UnderEvaluation      bool   `json:"under_evaluation"`
	RainfallBand         string `json:"rainfall_band,omitempty"`
	TotalRainfallBand    string `json:"total_rainfall_band,omitempty"`
	TotalGraphicalStatus string `json:"total_graphical_status,omitempty"`
	TotalPeriod          string `json:"total_period,omitempty"`
	Unit                 string `json:"unit,omitempty"`
	AmountScope          string `json:"amount_scope,omitempty"`
}
type RegionalProjection struct {
	Product, Statement string
	Facts              []RegionalFact
	Limitations        []string
}

var riskIDs = []string{"minor_network_hydro", "main_network_hydraulic", "thunderstorms", "wind", "coastal_waves", "snow", "ice"}
var italianDate = regexp.MustCompile(`(?i)([0-9]{1,2})\s+(gennaio|febbraio|marzo|aprile|maggio|giugno|luglio|agosto|settembre|ottobre|novembre|dicembre)\s+([0-9]{4})`)
var clockDate = regexp.MustCompile(`(?i)([0-9]{1,2})[:.]([0-9]{2})\s+(?:del\s+)?([0-9]{1,2})/([0-9]{1,2})/([0-9]{4})`)

func parseRegionalDay(value string) (*time.Time, error) {
	m := italianDate.FindStringSubmatch(value)
	if len(m) != 4 {
		return nil, fmt.Errorf("regional date unrecognized")
	}
	months := strings.Fields("gennaio febbraio marzo aprile maggio giugno luglio agosto settembre ottobre novembre dicembre")
	month := 0
	for i, v := range months {
		if v == strings.ToLower(m[2]) {
			month = i + 1
		}
	}
	day, _ := strconv.Atoi(m[1])
	year, _ := strconv.Atoi(m[3])
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if d.Day() != day {
		return nil, fmt.Errorf("regional date invalid")
	}
	return &d, nil
}

// Dates are constrained to the complete explicit phrase, rather than matched
// from unrelated issuance text. Weekday spelling is evidence, not a date source.
var sharedItalianInterval = regexp.MustCompile(`(?i)^dalle ore ([0-9]{1,2})[:.]([0-9]{2}) alle ore ([0-9]{1,2})[:.]([0-9]{2}) di (?:[a-zàèéìíòù]+,? )?([0-9]{1,2} [a-z]+ [0-9]{4})$`)
var separateItalianInterval = regexp.MustCompile(`(?i)^dalle ore ([0-9]{1,2})[:.]([0-9]{2}) (?:di )?(?:[a-zàèéìíòù]+,? )?([0-9]{1,2} [a-z]+ [0-9]{4}) alle ore ([0-9]{1,2})[:.]([0-9]{2}) (?:di )?(?:[a-zàèéìíòù]+,? )?([0-9]{1,2} [a-z]+ [0-9]{4})$`)

func rowInterval(value string) (*time.Time, *time.Time, error) {
	value = strings.Join(strings.Fields(value), " ")
	if m := sharedItalianInterval.FindStringSubmatch(value); m != nil {
		day, err := parseRegionalDay(m[5])
		if err != nil {
			return nil, nil, err
		}
		value = fmt.Sprintf("%s:%s del %s %s:%s del %s", m[1], m[2], day.Format("02/01/2006"), m[3], m[4], day.Format("02/01/2006"))
	} else if m := separateItalianInterval.FindStringSubmatch(value); m != nil {
		start, err := parseRegionalDay(m[3])
		if err != nil {
			return nil, nil, err
		}
		end, err := parseRegionalDay(m[6])
		if err != nil {
			return nil, nil, err
		}
		value = fmt.Sprintf("%s:%s del %s %s:%s del %s", m[1], m[2], start.Format("02/01/2006"), m[4], m[5], end.Format("02/01/2006"))
	}

	matches := clockDate.FindAllStringSubmatch(value, -1)
	if len(matches) != 2 {
		return nil, nil, fmt.Errorf("regional interval unrecognized")
	}
	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		return nil, nil, err
	}
	var times []time.Time
	for _, m := range matches {
		n := make([]int, 5)
		for i := range n {
			n[i], _ = strconv.Atoi(m[i+1])
		}
		t := time.Date(n[4], time.Month(n[3]), n[2], n[0], n[1], 0, 0, loc)
		if t.Hour() != n[0] || t.Minute() != n[1] || t.Day() != n[2] || t.Month() != time.Month(n[3]) {
			return nil, nil, fmt.Errorf("invalid local interval")
		}
		for _, other := range []time.Time{t.Add(-time.Hour), t.Add(time.Hour)} {
			if other.Format("2006-01-02 15:04") == t.Format("2006-01-02 15:04") {
				return nil, nil, fmt.Errorf("ambiguous local interval")
			}
		}
		times = append(times, t.UTC())
	}
	if !times[1].After(times[0]) {
		return nil, nil, fmt.Errorf("reversed regional interval")
	}
	return &times[0], &times[1], nil
}

// ProjectRegionalHTML emits explicit table facts. An explicit no-criticality
// statement is retained as a statement with UNKNOWN per-zone levels, never
// expanded into invented green colors. Missing graphical interpretation stays
// visible. Unknown/new layouts fail closed instead of yielding an all-clear.
func ProjectRegionalHTML(product string, body []byte) (RegionalProjection, error) {
	out := RegionalProjection{Product: product, Facts: []RegionalFact{}, Limitations: []string{}}
	// Scripts and styles are not bulletin evidence.
	cleaned, cleanErr := html.Parse(strings.NewReader(string(body)))
	if cleanErr != nil {
		return out, cleanErr
	}
	var prune func(*html.Node)
	prune = func(n *html.Node) {
		for child := n.FirstChild; child != nil; {
			next := child.NextSibling
			if child.Type == html.ElementNode && (child.Data == "script" || child.Data == "style" || child.Data == "template") {
				n.RemoveChild(child)
			} else {
				prune(child)
			}
			child = next
		}
	}
	prune(cleaned)
	var rendered strings.Builder
	if err := html.Render(&rendered, cleaned); err != nil {
		return out, err
	}
	body = []byte(rendered.String())
	obs, err := ObserveRegionalHTML(product, body, nil)
	if err != nil {
		return out, err
	}
	root, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return out, err
	}
	if product == "monitoring" {
		if obs.MonitoringNoEvent {
			out.Statement = "NESSUN AVVISO IN CORSO DI VALIDITÀ O EVENTO IN CORSO"
		} else {
			out.Statement = obs.IssuanceExpression
			out.Limitations = append(out.Limitations, "Monitoring narrative has no verified coded zone/risk/validity projection; consult the retained original. No new alert color is inferred.")
		}
		return out, nil
	}
	tables := htmlTableRows(root)
	if product == "criticality" {
		for ti, rows := range tables {
			if len(rows) < 2 || len(rows[0]) < 4 {
				continue
			}
			header := strings.ToUpper(strings.Join(rows[0], " "))
			if !strings.Contains(header, "ZONE") || !strings.Contains(header, "RISCHIO") || !strings.Contains(header, "CRITIC") {
				continue
			}
			for ri, row := range rows[1:] {
				if len(row) != 4 {
					return out, fmt.Errorf("criticality row shape changed")
				}
				risk := ""
				label := strings.ToUpper(row[1])
				for i, l := range criticalityRiskLabels {
					if label == l || label == strings.TrimPrefix(l, "RISCHIO ") {
						risk = riskIDs[i]
					}
				}
				if risk == "" {
					switch label {
					case "IDROGEOLOGICO IDRAULICO RETICOLO MINORE", "IDROGEOLOGICO-IDRAULICO RETICOLO MINORE":
						risk = riskIDs[0]
					case "IDRAULICO RETICOLO PRINCIPALE":
						risk = riskIDs[1]
					case "TEMPORALI FORTI":
						risk = riskIDs[2]
					case "VENTO":
						risk = riskIDs[3]
					case "MAREGGIATE":
						risk = riskIDs[4]
					case "NEVE":
						risk = riskIDs[5]
					case "GHIACCIO":
						risk = riskIDs[6]
					}
				}
				level := map[string]string{"verde": "green", "giallo": "yellow", "arancione": "orange", "rosso": "red"}[strings.ToLower(row[3])]
				zones := zonePattern.FindAllString(row[0], -1)
				start, end, e := rowInterval(row[2])
				if risk == "" || level == "" || len(zones) == 0 || e != nil {
					return out, fmt.Errorf("unsupported criticality row: %w", ErrUnrecognizedContent)
				}
				for _, z := range uniqueSorted(zones) {
					out.Facts = append(out.Facts, RegionalFact{Risk: risk, Label: row[1], Zone: z, Level: level, Original: row[2], Precision: "interval", Start: start, End: end, Locator: fmt.Sprintf("HTML criticality table %d row %d: %s", ti+1, ri+2, strings.Join(row, " | "))})
				}
			}
		}
		if len(out.Facts) > 0 {
			if obs.TextSaysNoCriticality {
				return out, fmt.Errorf("conflicting criticality table and statement")
			}
			out.Limitations = append(out.Limitations, "Only explicit criticality table rows are projected; map-only colors and unlisted zones are not inferred.")
			return out, nil
		}
		if !obs.TextSaysNoCriticality || len(obs.ValidityExpressions) != 2 || len(obs.RisksOrPhenomena) != 7 {
			return out, fmt.Errorf("criticality has no supported rows or explicit no-criticality statement")
		}
		// The area key names the product's geographic scope; it does not supply colors.
		text := strings.Join(strings.Fields(nodeText(root)), " ")
		begin := strings.Index(text, "AREE INTERESSATE:")
		end := strings.Index(text, "Legenda criticità:")
		if begin < 0 || end <= begin {
			return out, fmt.Errorf("criticality area key missing")
		}
		zones := uniqueSorted(zonePattern.FindAllString(text[begin:end], -1))
		if len(zones) != 26 {
			return out, fmt.Errorf("criticality zone scope incomplete")
		}
		out.Statement = "Criticità previste: NESSUNA"
		out.Limitations = append(out.Limitations, "Official bulletin states 'Criticità previste: NESSUNA'. Individual risk/zone map colors have not been verified: level remains unknown, not green.")
		for _, expression := range obs.ValidityExpressions {
			date, e := parseRegionalDay(expression)
			if e != nil {
				return out, e
			}
			for i, risk := range riskIDs {
				for _, z := range zones {
					out.Facts = append(out.Facts, RegionalFact{Risk: risk, Label: criticalityRiskLabels[i], Zone: z, Level: "unknown", Original: expression, Precision: "date", Date: date, Locator: "HTML: Criticità previste: NESSUNA; area key; dated risk-map headings"})
				}
			}
		}
		return out, nil
	}
	if product == "vigilance" {
		phenomena := map[string][]string{"pioggia": {riskIDs[0], riskIDs[1]}, "temporali": {riskIDs[2]}, "vento": {riskIDs[3]}, "mare": {riskIDs[4]}, "neve": {riskIDs[5]}, "ghiaccio": {riskIDs[6]}}
		for _, rows := range tables {
			matched := 0
			for _, row := range rows {
				if len(row) > 0 && phenomena[strings.ToLower(row[0])] != nil {
					matched++
				}
			}
			if matched != 6 {
				continue
			}
			if len(rows[0]) != 3 {
				return out, fmt.Errorf("vigilance date columns changed")
			}
			for _, row := range rows[1:] {
				if len(row) != 3 {
					return out, fmt.Errorf("vigilance row columns changed")
				}
				risks := phenomena[strings.ToLower(row[0])]
				if risks == nil {
					return out, fmt.Errorf("unknown phenomenon")
				}
				for day := 1; day <= 2; day++ {
					date, e := parseRegionalDay(rows[0][day])
					if e != nil {
						return out, e
					}
					for _, z := range uniqueSorted(zonePattern.FindAllString(row[day], -1)) {
						for _, risk := range risks {
							out.Facts = append(out.Facts, RegionalFact{Risk: risk, Label: row[0], Zone: z, Level: "not_applicable", Original: rows[0][day], Precision: "date", Date: date, Locator: "HTML vigilance zone-summary table: " + strings.Join(row, " | ")})
						}
					}
				}
			}
			out.Statement = "Vigilance phenomenon zone-summary table"
			out.Limitations = append(out.Limitations, "Vigilance lists phenomena under evaluation, not alert colors. Rainfall bands and graphical details are retained but not interpreted; unlisted zones do not establish absence of risk.")
			return out, nil
		}
	}
	return out, fmt.Errorf("regional product layout unsupported")
}
