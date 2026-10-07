package acquisition

import (
	"fmt"
	"math"
	"strings"
)

const vigilanceRainHeading = "FENOMENI PIOGGIA e TEMPORALI"
const vigilanceOtherHeading = "FENOMENI VENTO, MARE, NEVE e GHIACCIO"

func vectorVigilancePage(p vectorPage) bool {
	for _, l := range vectorLines(p) {
		if vectorLineText(l) == vigilanceRainHeading {
			return true
		}
	}
	return false
}

func vectorEdition(pages []vectorPage, issuance string) bool {
	for _, p := range pages {
		var words []string
		for _, w := range p.Words {
			words = append(words, w.Text)
		}
		if strings.Contains(strings.Join(words, " "), issuance) {
			return true
		}
	}
	return false
}

func vectorBounds(p vectorPolygon) (float64, float64, float64, float64) {
	x0, y0, x1, y1 := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, q := range p.Points {
		x0, y0, x1, y1 = math.Min(x0, q.X), math.Min(y0, q.Y), math.Max(x1, q.X), math.Max(y1, q.Y)
	}
	return x0, y0, x1, y1
}

func vectorSameSymbol(a, b vectorPolygon) bool {
	if a.Fill != b.Fill || a.Stroke != b.Stroke || len(a.Points) != len(b.Points) {
		return false
	}
	ax, ay, ax1, ay1 := vectorBounds(a)
	bx, by, bx1, by1 := vectorBounds(b)
	if ax1 <= ax || ay1 <= ay || bx1 <= bx || by1 <= by {
		return false
	}
	for i, p := range a.Points {
		q := b.Points[i]
		if math.Abs((p.X-ax)/(ax1-ax)-(q.X-bx)/(bx1-bx)) > 0.15 || math.Abs((p.Y-ay)/(ay1-ay)-(q.Y-by)/(by1-by)) > 0.15 {
			return false
		}
	}
	return true
}

// ProjectVigilanceVector reads the five labelled panels and their own legends.
// Two rainfall days and their explicitly labelled total stay distinct. All seven
// risk categories are represented, but warning colors do not apply to vigilance.
func ProjectVigilanceVector(htmlBody []byte, pdfURL string, evidence VectorEvidence) (RegionalProjection, error) {
	out := RegionalProjection{Product: "vigilance", Facts: []RegionalFact{}, Limitations: []string{}}
	rows, err := ProjectRegionalHTML("vigilance", htmlBody)
	if err != nil {
		return out, err
	}
	obs, err := ObserveRegionalHTML("vigilance", htmlBody, nil)
	if err != nil {
		return out, err
	}
	pages, err := vectorPages(evidence.BBox)
	if err != nil || !vectorEdition(pages, obs.IssuanceExpression) {
		return out, fmt.Errorf("vigilance PDF/HTML issuance mismatch")
	}
	pageNumber := 0
	var page vectorPage
	for i, p := range pages {
		if vectorVigilancePage(p) {
			if pageNumber != 0 {
				return out, fmt.Errorf("duplicate vigilance map page")
			}
			pageNumber, page = i+1, p
		}
	}
	if pageNumber == 0 {
		return out, ErrUnrecognizedContent
	}
	shapes, err := vectorShapes(evidence.Pages[pageNumber])
	if err != nil {
		return out, err
	}
	polygons, err := vectorPolygons(evidence.Pages[pageNumber])
	if err != nil {
		return out, err
	}
	lines := vectorLines(page)
	// Maps occupy the printable content grid, not thirds of the physical page.
	// Derive that grid from the two dated column headings in this document.
	origin, columnWidth := 0.0, 0.0
	for _, line := range lines {
		if len(italianDate.FindAllString(vectorLineText(line), -1)) != 2 {
			continue
		}
		var starts []float64
		for i, w := range line.Words {
			var text []string
			for j := i; j < len(line.Words) && j < i+4; j++ {
				text = append(text, line.Words[j].Text)
			}
			if regionalDatePattern.MatchString(strings.Join(text, " ")) {
				starts = append(starts, w.XMin)
			}
		}
		if len(starts) == 2 {
			origin, columnWidth = starts[0], starts[1]-starts[0]
			break
		}
	}
	if columnWidth <= 0 {
		return out, fmt.Errorf("vigilance dated column grid unavailable")
	}
	column := func(x float64) int { return int(math.Floor((x - origin) / columnWidth)) }
	rainY, otherY, forecastY, legendY := -1.0, -1.0, -1.0, -1.0
	for _, line := range lines {
		switch vectorLineText(line) {
		case vigilanceRainHeading:
			rainY = line.Y
		case vigilanceOtherHeading:
			otherY = line.Y
		case "Previsione fino alle 24 di domani:":
			forecastY = line.Y
		case "Cumulato medio sull'area [mm]":
			legendY = line.Y
		}
	}
	if rainY < 0 || otherY <= legendY || legendY <= rainY || forecastY <= otherY {
		return out, fmt.Errorf("vigilance panel/legend layout unsupported")
	}
	rasters, err := vectorRasterRects(evidence.Pages[pageNumber])
	if err != nil {
		return out, err
	}
	for _, raster := range rasters {
		_, y0, _, y1 := vectorBounds(raster)
		if y1 > rainY && y0 < forecastY {
			return out, fmt.Errorf("unsupported raster inside vigilance panels or legends")
		}
	}
	// Read each literal band from the legend in this PDF. No criticality palette
	// is reused: white is 0–10 mm here, never green or no rainfall.
	bands := map[string]string{}
	for _, line := range lines {
		if line.Y <= legendY || line.Y >= otherY {
			continue
		}
		for i := 0; i < len(line.Words); {
			count := 3
			if line.Words[i].Text == ">" {
				count = 2
			}
			if i+count > len(line.Words) {
				return out, fmt.Errorf("rainfall legend labels unsupported")
			}
			var text []string
			for _, w := range line.Words[i : i+count] {
				text = append(text, w.Text)
			}
			band := strings.Join(text, " ")
			w := line.Words[i]
			w.XMax = line.Words[i+count-1].XMax
			color := "rgb(100%, 100%, 100%)"
			if p, ok := vectorLabelPolygon(polygons, w); ok {
				color = p.Fill
			} else if band != "0 - 10" {
				return out, fmt.Errorf("rainfall legend fill unavailable")
			}
			if _, exists := bands[color]; exists {
				return out, fmt.Errorf("ambiguous rainfall legend color")
			}
			bands[color] = band
			i += count
		}
	}
	if len(bands) != 8 {
		return out, fmt.Errorf("rainfall legend incomplete")
	}
	for _, expected := range []string{"0 - 10", "10 - 20", "20 - 40", "40 - 60", "60 - 80", "80 - 100", "100 - 120", "> 120"} {
		found := false
		for _, band := range bands {
			if band == expected {
				found = true
			}
		}
		if !found {
			return out, fmt.Errorf("rainfall legend bands changed")
		}
	}
	zonePolygons := vectorZonePolygons(polygons)
	// Symbol identities come from the same document's labelled legend, including
	// the lightning mark. A symbol in the legend itself is never a mapped event.
	symbolNames := map[string]string{"Temporali": "thunderstorms", "Vento": "wind", "Mare": "coastal_waves", "Neve": "snow", "Ghiaccio": "ice"}
	symbols := map[string][]vectorPolygon{}
	for _, line := range lines {
		if !((line.Y > rainY && line.Y < legendY) || (line.Y > otherY && line.Y < forecastY)) {
			continue
		}
		for _, w := range line.Words {
			name := symbolNames[w.Text]
			if name == "" {
				continue
			}
			for _, p := range shapes {
				x0, y0, x1, y1 := vectorBounds(p)
				if x0 >= w.XMax && x1 <= w.XMax+80 && y0 <= w.YMax && y1 >= w.YMin {
					symbols[name] = append(symbols[name], p)
				}
			}
		}
	}
	if len(symbols["thunderstorms"]) != 1 {
		return out, fmt.Errorf("thunderstorm legend unavailable or ambiguous")
	}
	underEvaluation := map[string]bool{}
	for _, f := range rows.Facts {
		underEvaluation[f.Risk+":"+f.Zone+":"+f.Date.Format("2006-01-02")] = true
	}
	type mapZone struct {
		word    vectorWord
		polygon vectorPolygon
		known   bool
	}
	var panels [5]map[string]mapZone
	var dates [2][2]string
	totalPeriod := ""
	for _, w := range page.Words {
		if w.YMin <= rainY || w.YMin >= forecastY {
			continue
		}
		group := 0
		end := legendY
		if w.YMin > otherY {
			group, end = 1, forecastY
		}
		if w.YMin >= end {
			continue
		}
		side := column(w.XMin)
		if side < 0 || side > 2 || (group == 1 && side == 2) {
			continue
		}
		if zonePattern.FindString(w.Text) == w.Text {
			index := side
			if group == 1 {
				index += 3
			}
			if panels[index] == nil {
				panels[index] = map[string]mapZone{}
			}
			if _, exists := panels[index][w.Text]; exists {
				return out, fmt.Errorf("duplicate vigilance zone label")
			}
			p, ok := vectorLabelPolygon(zonePolygons, w)
			panels[index][w.Text] = mapZone{w, p, ok}
		}
	}
	for group, heading := range []float64{rainY, otherY} {
		for _, line := range lines {
			if line.Y <= heading {
				continue
			}
			if len(italianDate.FindAllString(vectorLineText(line), -1)) != 2 {
				continue
			}
			for side := 0; side < 2; side++ {
				var words []string
				for _, w := range line.Words {
					if column(w.XMin) == side {
						words = append(words, w.Text)
					}
				}
				dates[group][side] = strings.Join(words, " ")
			}
			break
		}
	}
	for _, line := range lines {
		if line.Y > rainY && line.Y < legendY {
			var words []string
			for _, w := range line.Words {
				if column(w.XMin) == 2 && zonePattern.FindString(w.Text) != w.Text {
					words = append(words, w.Text)
				}
			}
			text := strings.Join(words, " ")
			if strings.HasPrefix(text, "TOTALE:") || strings.HasPrefix(text, "alle 24 di domani") {
				if totalPeriod != "" {
					totalPeriod += " "
				}
				totalPeriod += text
			}
		}
	}
	if totalPeriod != "TOTALE: dalle 12 di oggi alle 24 di domani" {
		return out, fmt.Errorf("vigilance cumulative period unsupported")
	}
	for _, panel := range panels {
		if len(panel) != 26 {
			return out, fmt.Errorf("vigilance zone scope incomplete")
		}
	}
	isZoneShape := func(p vectorPolygon) bool {
		for _, panel := range panels {
			for _, z := range panel {
				if !z.known || p.Fill != z.polygon.Fill || p.Stroke != z.polygon.Stroke || len(p.Points) != len(z.polygon.Points) {
					continue
				}
				equal := true
				for i, q := range p.Points {
					if q != z.polygon.Points[i] {
						equal = false
						break
					}
				}
				if equal {
					return true
				}
			}
		}
		return false
	}
	unresolved := 0
	dailyUnknown, totalUnknown, symbolsUnknown := 0, 0, 0
	for side := 0; side < 2; side++ {
		date, e := parseRegionalDay(dates[0][side])
		otherDate, otherErr := parseRegionalDay(dates[1][side])
		if e != nil || otherErr != nil || !date.Equal(*otherDate) {
			return out, fmt.Errorf("vigilance panel dates inconsistent")
		}
		issuanceDay, issuanceErr := parseRegionalDay(obs.IssuanceExpression)
		if issuanceErr != nil || !date.Equal(issuanceDay.AddDate(0, 0, side)) {
			return out, fmt.Errorf("vigilance map validity does not match its edition")
		}
		for _, zone := range strings.Fields("A1 A2 A3 A4 A5 A6 B C E1 E2 E3 F1 F2 I L M O1 O2 O3 R1 R2 S1 S2 S3 T V") {
			for risk, id := range riskIDs {
				panel := panels[side]
				phenomenon := "rainfall"
				if risk == 2 {
					phenomenon = "thunderstorms"
				}
				if risk > 2 {
					panel, phenomenon = panels[side+3], id
				}
				z := panel[zone]
				weather := &VigilanceWeather{Phenomenon: phenomenon, GraphicalStatus: "not_depicted", UnderEvaluation: underEvaluation[id+":"+zone+":"+date.Format("2006-01-02")]}
				if !z.known {
					weather.GraphicalStatus = "unresolved"
				}
				if risk < 2 {
					weather.Unit, weather.AmountScope, weather.TotalPeriod = "mm", "area_average", totalPeriod
					if z.known {
						weather.RainfallBand = bands[z.polygon.Fill]
					}
					if total := panels[2][zone]; total.known {
						weather.TotalRainfallBand = bands[total.polygon.Fill]
					}
					weather.TotalGraphicalStatus = "depicted"
					if weather.TotalRainfallBand == "" {
						weather.TotalGraphicalStatus = "unresolved"
						totalUnknown++
					}
					if weather.RainfallBand == "" {
						weather.GraphicalStatus = "unresolved"
						dailyUnknown++
					} else {
						weather.GraphicalStatus = "depicted"
					}
				} else if z.known {
					unknownMark := false
					for _, p := range shapes {
						if isZoneShape(p) {
							continue
						}
						x0, y0, x1, y1 := vectorBounds(p)
						if !vectorContains(z.polygon, vectorPoint{(x0 + x1) / 2, (y0 + y1) / 2}) {
							continue
						}
						matchedNames := map[string]bool{}
						for name, exemplars := range symbols {
							// Multi-component or otherwise ambiguous symbols need their
							// own reviewed decoder; one shared component is insufficient.
							if len(exemplars) != 1 {
								continue
							}
							for _, exemplar := range exemplars {
								if vectorSameSymbol(p, exemplar) {
									matchedNames[name] = true
								}
							}
						}
						if len(matchedNames) != 1 {
							unknownMark = true
						} else if matchedNames[phenomenon] {
							weather.GraphicalStatus = "depicted"
						}
					}
					if unknownMark {
						weather.GraphicalStatus = "unresolved"
					}
					if weather.UnderEvaluation && weather.GraphicalStatus != "depicted" {
						weather.GraphicalStatus = "unresolved"
					}
				}
				if weather.GraphicalStatus == "unresolved" || weather.TotalGraphicalStatus == "unresolved" {
					unresolved++
					if risk >= 2 {
						symbolsUnknown++
					}
				}
				label := []string{"Pioggia", "Pioggia", "Temporali", "Vento", "Mare", "Neve", "Ghiaccio"}[risk]
				out.Facts = append(out.Facts, RegionalFact{Risk: id, Label: label, Zone: zone, Level: "not_applicable", Original: dates[0][side], Precision: "date", Date: date, Locator: "Retained PDF vigilance labelled zone and product-specific map/legend; absence of a symbol is not absence of risk; " + evidence.PageMethods[pageNumber], EvidenceURL: pdfURL, Page: pageNumber, Weather: weather})
			}
		}
	}
	out.Statement = "Vigilance weather maps and area-average rainfall bands read from retained PDF; warning colors do not apply"
	if unresolved > 0 {
		out.Limitations = append(out.Limitations, fmt.Sprintf("Unresolved vigilance graphics: %d zone/phenomenon/day entries; no weather value or absence is inferred.", unresolved))
		out.Limitations = append(out.Limitations, fmt.Sprintf("Unresolved daily rainfall bands: %d; cumulative rainfall bands: %d; symbol or evaluation entries: %d. Bands require a fill present in this edition's labelled legend; daily and cumulative evidence remain separate.", dailyUnknown, totalUnknown, symbolsUnknown))
	}
	return out, nil
}
