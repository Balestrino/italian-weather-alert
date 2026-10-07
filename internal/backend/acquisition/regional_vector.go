package acquisition

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// VectorEvidence is produced from the same retained PDF, never from an external
// map URL. Text coordinates identify zones; closed filled paths supply colors.
type VectorEvidence struct {
	BBox        []byte
	Pages       map[int][]byte
	PageMethods map[int]string
}
type VectorPDFReader interface {
	Read(context.Context, []byte) (VectorEvidence, error)
}
type PopplerVectorReader struct{}

func (PopplerVectorReader) Read(ctx context.Context, body []byte) (VectorEvidence, error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	out := VectorEvidence{Pages: map[int][]byte{}, PageMethods: map[int]string{}}
	if len(body) > 32<<20 || !bytes.HasPrefix(body, []byte("%PDF-")) {
		return out, ErrUnrecognizedContent
	}
	cache, err := os.MkdirTemp("", "iwa-vector-fonts-")
	if err != nil {
		return out, err
	}
	defer os.RemoveAll(cache)
	run := func(input []byte, name string, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdin = bytes.NewReader(input)
		cmd.Env = append(os.Environ(), "XDG_CACHE_HOME="+cache, "LC_ALL=C")
		var w boundedVectorOutput
		cmd.Stdout = &w
		err := cmd.Run()
		return w.Bytes(), err
	}
	out.BBox, err = run(body, "pdftotext", "-bbox", "-", "-")
	if err != nil {
		return out, err
	}
	pages, err := vectorPages(out.BBox)
	if err != nil || len(pages) > 12 {
		return out, ErrUnrecognizedContent
	}
	for i, p := range pages {
		if len(vectorMapHeadings(p)) == 0 && !vectorVigilancePage(p) {
			continue
		}
		out.Pages[i+1], err = run(body, "pdftocairo", "-svg", "-f", strconv.Itoa(i+1), "-l", strconv.Itoa(i+1), "-", "-")
		out.PageMethods[i+1] = "direct retained PDF vector rendering"
		if err != nil {
			// Some retained TCPDF editions trigger Cairo's SVG font-cache failure.
			// A bounded local vector-PDF rendering embeds substitute fonts before
			// converting that single page. Never accept partial crashed SVG output.
			pagePDF, e := run(body, "pdftocairo", "-pdf", "-f", strconv.Itoa(i+1), "-l", strconv.Itoa(i+1), "-", "-")
			if e != nil {
				return out, e
			}
			out.Pages[i+1], err = run(pagePDF, "pdftocairo", "-svg", "-f", "1", "-l", "1", "-", "-")
			if err != nil {
				return out, err
			}
			out.PageMethods[i+1] = "local single-page vector PDF font normalization, then SVG rendering; original physical page retained"
		}
	}
	return out, nil
}

type boundedVectorOutput struct{ bytes.Buffer }

func (w *boundedVectorOutput) Write(b []byte) (int, error) {
	if w.Len()+len(b) > 16<<20 {
		return 0, fmt.Errorf("vector output exceeds limit")
	}
	return w.Buffer.Write(b)
}

type vectorWord struct {
	XMin float64 `xml:"xMin,attr"`
	YMin float64 `xml:"yMin,attr"`
	XMax float64 `xml:"xMax,attr"`
	YMax float64 `xml:"yMax,attr"`
	Text string  `xml:",chardata"`
}
type vectorPage struct {
	Width  float64      `xml:"width,attr"`
	Height float64      `xml:"height,attr"`
	Words  []vectorWord `xml:"word"`
}

func vectorPages(body []byte) ([]vectorPage, error) {
	d := xml.NewDecoder(bytes.NewReader(body))
	var pages []vectorPage
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if e, ok := tok.(xml.StartElement); ok && e.Name.Local == "page" {
			var p vectorPage
			if err = d.DecodeElement(&p, &e); err != nil {
				return nil, err
			}
			if p.Width <= 0 || p.Height <= 0 {
				return nil, ErrUnrecognizedContent
			}
			pages = append(pages, p)
		}
	}
	if len(pages) == 0 {
		return nil, ErrUnrecognizedContent
	}
	return pages, nil
}

type vectorLine struct {
	Y     float64
	Words []vectorWord
}

func vectorLines(p vectorPage) []vectorLine {
	words := append([]vectorWord{}, p.Words...)
	sort.Slice(words, func(i, j int) bool { return words[i].YMin < words[j].YMin })
	var lines []vectorLine
	for _, w := range words {
		if len(lines) == 0 || math.Abs(lines[len(lines)-1].Y-w.YMin) > 1 {
			lines = append(lines, vectorLine{Y: w.YMin})
		}
		lines[len(lines)-1].Words = append(lines[len(lines)-1].Words, w)
	}
	for i := range lines {
		sort.Slice(lines[i].Words, func(a, b int) bool { return lines[i].Words[a].XMin < lines[i].Words[b].XMin })
	}
	return lines
}
func vectorLineText(l vectorLine) string {
	var words []string
	for _, w := range l.Words {
		words = append(words, w.Text)
	}
	return strings.Join(words, " ")
}

type vectorHeading struct {
	Y    float64
	Risk int
}

func vectorHeadings(p vectorPage) []vectorHeading {
	var out []vectorHeading
	for _, line := range vectorLines(p) {
		for i, label := range criticalityRiskLabels {
			if vectorLineText(line) == label {
				out = append(out, vectorHeading{line.Y, i})
			}
		}
	}
	return out
}

// Risk names can also occur in scenario prose before the graphical pages.
// A map heading must have its own following two-day row, within its section.
func vectorMapHeadings(p vectorPage) []vectorHeading {
	headings := vectorHeadings(p)
	var maps []vectorHeading
	for i, h := range headings {
		end := p.Height
		if i+1 < len(headings) {
			end = headings[i+1].Y
		}
		for _, line := range vectorLines(p) {
			if line.Y > h.Y && line.Y < end && len(italianDate.FindAllString(vectorLineText(line), -1)) == 2 {
				maps = append(maps, h)
				break
			}
		}
	}
	return maps
}

type vectorPoint struct{ X, Y float64 }
type vectorPolygon struct {
	Points []vectorPoint
	Fill   string
	Stroke string
}

var vectorNumbers = regexp.MustCompile(`[-+]?(?:[0-9]*\.)?[0-9]+(?:[eE][-+]?[0-9]+)?`)

func vectorShapes(body []byte) ([]vectorPolygon, error) {
	return vectorMapShapes(body, math.Inf(-1))
}

func vectorMapShapes(body []byte, firstMapY float64) ([]vectorPolygon, error) {
	d := xml.NewDecoder(bytes.NewReader(body))
	var out []vectorPolygon
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		e, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		attrs := map[string]string{}
		for _, a := range e.Attr {
			attrs[a.Name.Local] = a.Value
		}
		if e.Name.Local == "g" && (attrs["transform"] != "" || attrs["opacity"] != "" && attrs["opacity"] != "1") {
			return nil, ErrUnrecognizedContent
		}
		if e.Name.Local != "path" || attrs["fill"] == "none" || attrs["fill"] == "" {
			continue
		}
		if attrs["fill-opacity"] != "1" || attrs["opacity"] != "" && attrs["opacity"] != "1" {
			return nil, ErrUnrecognizedContent
		}
		if attrs["transform"] != "" && !strings.HasPrefix(attrs["transform"], "matrix(") {
			return nil, ErrUnrecognizedContent
		}
		m := [6]float64{1, 0, 0, 1, 0, 0}
		if attrs["transform"] != "" {
			ns := vectorNumbers.FindAllString(attrs["transform"], -1)
			if len(ns) != 6 {
				return nil, ErrUnrecognizedContent
			}
			for i, n := range ns {
				m[i], err = strconv.ParseFloat(n, 64)
				if err != nil || math.IsInf(m[i], 0) || math.IsNaN(m[i]) {
					return nil, ErrUnrecognizedContent
				}
			}
		}
		fields := strings.Fields(attrs["d"])
		// A cubic header symbol is irrelevant only when every transformed
		// control point lies strictly above all dated maps on this page.
		// The convex hull bounds the entire curve; map curves remain rejected.
		if headerCurveOutsideMaps(fields, m, firstMapY) {
			continue
		}
		p := vectorPolygon{Fill: attrs["fill"], Stroke: attrs["stroke"]}
		finish := func() {
			if len(p.Points) > 1 && p.Points[0] == p.Points[len(p.Points)-1] {
				p.Points = p.Points[:len(p.Points)-1]
			}
			if len(p.Points) > 2 {
				out = append(out, p)
			}
			p = vectorPolygon{Fill: attrs["fill"], Stroke: attrs["stroke"]}
		}
		for i := 0; i < len(fields); {
			command := fields[i]
			i++
			if command == "Z" || command == "z" {
				finish()
				continue
			}
			if command != "M" && command != "L" {
				return nil, ErrUnrecognizedContent
			}
			if command == "M" {
				finish()
			}
			if i+1 >= len(fields) {
				return nil, ErrUnrecognizedContent
			}
			x, e1 := strconv.ParseFloat(fields[i], 64)
			y, e2 := strconv.ParseFloat(fields[i+1], 64)
			i += 2
			if e1 != nil || e2 != nil || math.IsInf(x, 0) || math.IsInf(y, 0) || math.IsNaN(x) || math.IsNaN(y) {
				return nil, ErrUnrecognizedContent
			}
			p.Points = append(p.Points, vectorPoint{m[0]*x + m[2]*y + m[4], m[1]*x + m[3]*y + m[5]})
		}
		finish()
	}
	if len(out) == 0 {
		return nil, ErrUnrecognizedContent
	}
	return out, nil
}

func headerCurveOutsideMaps(fields []string, matrix [6]float64, firstMapY float64) bool {
	curve, points := false, 0
	for i := 0; i < len(fields); {
		command := fields[i]
		i++
		pairs := 0
		switch command {
		case "M", "L":
			pairs = 1
		case "C":
			pairs, curve = 3, true
		case "Z", "z":
			continue
		default:
			return false
		}
		for j := 0; j < pairs; j++ {
			if i+1 >= len(fields) {
				return false
			}
			x, ex := strconv.ParseFloat(fields[i], 64)
			y, ey := strconv.ParseFloat(fields[i+1], 64)
			i += 2
			mappedY := matrix[1]*x + matrix[3]*y + matrix[5]
			if ex != nil || ey != nil || math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(mappedY) || math.IsInf(mappedY, 0) || mappedY >= firstMapY {
				return false
			}
			points++
		}
	}
	return curve && points > 0
}

// Raster logos outside the map are harmless; a raster placed in its panels
// makes graphical interpretation unsupported. Resolve Poppler's image uses
// without reading the embedded image bytes.
func vectorRasterRects(body []byte) ([]vectorPolygon, error) {
	d := xml.NewDecoder(bytes.NewReader(body))
	images := map[string]vectorPolygon{}
	var uses []map[string]string
	var out []vectorPolygon
	definitions := 0
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if end, ok := tok.(xml.EndElement); ok && end.Name.Local == "defs" {
			definitions--
		}
		e, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if e.Name.Local == "defs" {
			definitions++
		}
		a := map[string]string{}
		for _, attr := range e.Attr {
			a[attr.Name.Local] = attr.Value
		}
		if e.Name.Local == "use" {
			uses = append(uses, a)
		}
		if e.Name.Local != "image" {
			continue
		}
		var values [4]float64
		for i, key := range []string{"x", "y", "width", "height"} {
			if a[key] == "" && i < 2 {
				continue
			}
			values[i], err = strconv.ParseFloat(a[key], 64)
			if err != nil || math.IsNaN(values[i]) || math.IsInf(values[i], 0) {
				return nil, ErrUnrecognizedContent
			}
		}
		if values[2] <= 0 || values[3] <= 0 || a["transform"] != "" {
			return nil, ErrUnrecognizedContent
		}
		x, y, width, height := values[0], values[1], values[2], values[3]
		p := vectorPolygon{Points: []vectorPoint{{x, y}, {x + width, y}, {x + width, y + height}, {x, y + height}}}
		if definitions > 0 {
			images[a["id"]] = p
		} else {
			out = append(out, p)
		}
	}
	for _, a := range uses {
		p, found := images[strings.TrimPrefix(a["href"], "#")]
		if !found {
			continue
		}
		if a["x"] != "" || a["y"] != "" || !strings.HasPrefix(a["transform"], "matrix(") {
			return nil, ErrUnrecognizedContent
		}
		numbers := vectorNumbers.FindAllString(a["transform"], -1)
		if len(numbers) != 6 {
			return nil, ErrUnrecognizedContent
		}
		var m [6]float64
		for i, number := range numbers {
			m[i], _ = strconv.ParseFloat(number, 64)
			if math.IsNaN(m[i]) || math.IsInf(m[i], 0) {
				return nil, ErrUnrecognizedContent
			}
		}
		for i, q := range p.Points {
			p.Points[i] = vectorPoint{m[0]*q.X + m[2]*q.Y + m[4], m[1]*q.X + m[3]*q.Y + m[5]}
		}
		out = append(out, p)
	}
	return out, nil
}

func vectorPolygons(body []byte) ([]vectorPolygon, error) {
	shapes, err := vectorShapes(body)
	if err != nil {
		return nil, err
	}
	var polygons []vectorPolygon
	for _, p := range shapes {
		// Poppler may emit a zone's fill in page coordinates and its outline
		// separately. Keep those fills; geographical association is checked at
		// the label, inside a complete dated map panel, never at the legend.
		if p.Stroke == "rgb(0%, 0%, 0%)" || p.Stroke == "" {
			polygons = append(polygons, p)
		}
	}
	return polygons, nil
}

func vectorZonePolygons(polygons []vectorPolygon) []vectorPolygon {
	var zones []vectorPolygon
	for _, p := range polygons {
		// A standalone rectangular fill may be the page background or legend,
		// not an outlined zone. CFR's separately filled L outline is irregular.
		if p.Stroke == "rgb(0%, 0%, 0%)" || p.Stroke == "" && len(p.Points) > 4 {
			zones = append(zones, p)
		}
	}
	return zones
}

// Prefer an unambiguous label center. The reviewed CFR A6 and island labels can
// sit just outside their outline: a bounded typographic offset is permitted only
// for those labels and only with a clearly separated nearest outline. This is
// not a general geographical nearest-zone fallback.
func vectorLabelPolygon(polygons []vectorPolygon, w vectorWord) (vectorPolygon, bool) {
	q := vectorPoint{(w.XMin + w.XMax) / 2, (w.YMin + w.YMax) / 2}
	var matches []int
	for i, p := range polygons {
		if vectorContains(p, q) {
			matches = append(matches, i)
		}
	}
	if len(matches) == 1 {
		return polygons[matches[0]], true
	}
	if len(matches) != 0 || (w.Text != "A6" && w.Text != "I") {
		return vectorPolygon{}, false
	}
	best, second, index := math.Inf(1), math.Inf(1), -1
	for i, p := range polygons {
		d := vectorDistance(p, q)
		if d < best {
			second, best, index = best, d, i
		} else if d < second {
			second = d
		}
	}
	limit := (w.YMax - w.YMin) * 0.6
	if w.Text == "A6" {
		limit *= 0.1
	}
	if index >= 0 && best <= limit && second-best > (w.YMax-w.YMin)*0.3 {
		return polygons[index], true
	}
	return vectorPolygon{}, false
}

func vectorDistance(p vectorPolygon, q vectorPoint) float64 {
	best := math.Inf(1)
	for i, a := range p.Points {
		b := p.Points[(i+1)%len(p.Points)]
		dx, dy := b.X-a.X, b.Y-a.Y
		t := 0.0
		if dx*dx+dy*dy > 0 {
			t = math.Max(0, math.Min(1, ((q.X-a.X)*dx+(q.Y-a.Y)*dy)/(dx*dx+dy*dy)))
		}
		best = math.Min(best, math.Hypot(q.X-a.X-t*dx, q.Y-a.Y-t*dy))
	}
	return best
}
func vectorContains(p vectorPolygon, q vectorPoint) bool {
	inside := false
	j := len(p.Points) - 1
	for i, a := range p.Points {
		b := p.Points[j]
		if (a.Y > q.Y) != (b.Y > q.Y) && q.X < (b.X-a.X)*(q.Y-a.Y)/(b.Y-a.Y)+a.X {
			inside = !inside
		}
		j = i
	}
	return inside
}
func vectorColor(fill string, risk int) string {
	if !strings.HasPrefix(fill, "rgb(") {
		return "unknown"
	}
	nums := vectorNumbers.FindAllString(fill, -1)
	if len(nums) != 3 {
		return "unknown"
	}
	var c [3]float64
	for i, n := range nums {
		c[i], _ = strconv.ParseFloat(n, 64)
		if strings.Contains(fill, "%") {
			c[i] *= 2.55
		}
	}
	palettes := []struct {
		Level string
		RGB   [3]float64
	}{{"green", [3]float64{153, 204, 51}}, {"yellow", [3]float64{255, 255, 0}}, {"orange", [3]float64{255, 165, 0}}, {"red", [3]float64{255, 0, 0}}}
	for _, p := range palettes {
		if math.Abs(c[0]-p.RGB[0]) < 4 && math.Abs(c[1]-p.RGB[1]) < 4 && math.Abs(c[2]-p.RGB[2]) < 4 {
			return p.Level
		}
	}
	if (risk == 1 || risk == 4) && c[0] > 254 && c[1] > 254 && c[2] > 254 {
		return "not_applicable"
	}
	return "unknown"
}

// ProjectCriticalityVector reads each labelled risk/day map independently. A
// white *zone polygon* is excluded only in product masks for main rivers/coast;
// white page background, missing polygons and unknown colors stay unknown.
func ProjectCriticalityVector(htmlBody []byte, pdfURL string, evidence VectorEvidence) (RegionalProjection, error) {
	out := RegionalProjection{Product: "criticality", Facts: []RegionalFact{}, Limitations: []string{}}
	obs, err := ObserveRegionalHTML("criticality", htmlBody, nil)
	if err != nil {
		return out, err
	}
	pages, err := vectorPages(evidence.BBox)
	if err != nil {
		return out, err
	}
	// Match the edition, including issuance time, before combining HTML and PDF.
	if !vectorEdition(pages, obs.IssuanceExpression) {
		return out, fmt.Errorf("PDF/HTML issuance mismatch")
	}
	seen := map[string]bool{}
	riskDays := map[string]bool{}
	days := map[string]bool{}
	panels := map[string]int{}
	unresolved := 0
	for pageIndex, p := range pages {
		headings := vectorMapHeadings(p)
		if len(headings) == 0 {
			continue
		}
		polygons, e := vectorMapShapes(evidence.Pages[pageIndex+1], headings[0].Y)
		if e != nil {
			return out, e
		}
		polygons = vectorZonePolygons(polygons)
		rasters, e := vectorRasterRects(evidence.Pages[pageIndex+1])
		if e != nil {
			return out, e
		}
		lines := vectorLines(p)
		for hi, h := range headings {
			end := p.Height
			if hi+1 < len(headings) {
				end = headings[hi+1].Y
			}
			var dateLine *vectorLine
			for i := range lines {
				if lines[i].Y > h.Y && lines[i].Y < end && len(italianDate.FindAllString(vectorLineText(lines[i]), -1)) == 2 {
					dateLine = &lines[i]
					break
				}
			}
			if dateLine == nil {
				return out, fmt.Errorf("risk map dates missing")
			}
			for _, raster := range rasters {
				_, y0, _, y1 := vectorBounds(raster)
				if y1 > dateLine.Y && y0 < end {
					return out, fmt.Errorf("unsupported raster inside criticality map")
				}
			}
			var dates [2]*time.Time
			var expressions [2]string
			for side := 0; side < 2; side++ {
				var words []string
				for _, w := range dateLine.Words {
					if (w.XMin < p.Width/2) == (side == 0) {
						words = append(words, w.Text)
					}
				}
				expressions[side] = strings.Join(words, " ")
				dates[side], e = parseRegionalDay(expressions[side])
				if e != nil {
					return out, e
				}
				issuanceDay, issuanceErr := parseRegionalDay(obs.IssuanceExpression)
				if issuanceErr != nil || !dates[side].Equal(issuanceDay.AddDate(0, 0, side)) {
					return out, fmt.Errorf("criticality map validity does not match its edition")
				}
				riskDays[fmt.Sprintf("%d:%s", h.Risk, dates[side].Format(time.DateOnly))] = true
				days[dates[side].Format(time.DateOnly)] = true
			}
			for _, w := range p.Words {
				if w.YMin <= dateLine.Y || w.YMin >= end || zonePattern.FindString(w.Text) != w.Text {
					continue
				}
				side := 0
				if w.XMin >= p.Width/2 {
					side = 1
				}
				key := fmt.Sprintf("%d:%s:%s", h.Risk, dates[side].Format(time.DateOnly), w.Text)
				if seen[key] {
					return out, fmt.Errorf("duplicate risk/day/zone label")
				}
				seen[key] = true
				panels[fmt.Sprintf("%d:%s", h.Risk, dates[side].Format(time.DateOnly))]++
				level := "unknown"
				if polygon, ok := vectorLabelPolygon(polygons, w); ok {
					level = vectorColor(polygon.Fill, h.Risk)
				}
				if obs.TextSaysNoCriticality && (level == "yellow" || level == "orange" || level == "red") {
					return out, fmt.Errorf("map conflicts with no-criticality statement")
				}
				if level == "unknown" {
					unresolved++
				}
				out.Facts = append(out.Facts, RegionalFact{Risk: riskIDs[h.Risk], Label: criticalityRiskLabels[h.Risk], Zone: w.Text, Level: level, Original: expressions[side], Precision: "date", Date: dates[side], Locator: fmt.Sprintf("Retained PDF labelled zone %s; filled vector polygon; risk/day heading; %s", w.Text, evidence.PageMethods[pageIndex+1]), EvidenceURL: pdfURL, Page: pageIndex + 1})
			}
		}
	}
	for _, count := range panels {
		if count != 26 {
			return out, fmt.Errorf("risk/day zone labels incomplete")
		}
	}
	if len(days) != 2 || len(riskDays) != 14 || len(out.Facts) != 364 {
		return out, fmt.Errorf("incomplete seven-risk/two-day/26-zone map set")
	}
	out.Statement = "Criticality risk/day maps read from retained PDF vector polygons"
	if obs.TextSaysNoCriticality {
		out.Statement = "Criticità previste: NESSUNA; risk/day map colors verified"
	}
	if unresolved > 0 {
		out.Limitations = append(out.Limitations, fmt.Sprintf("Unresolved map labels: %d zone/risk/day entries lack an unambiguous labelled polygon; their levels remain unknown.", unresolved))
	}
	return out, nil
}

// Explicit alert rows retain their precise validity. Daily maps must not turn
// an afternoon warning into an alert already effective at midnight.
func MergeCriticalityMaps(maps, rows RegionalProjection) (RegionalProjection, error) {
	if rows.Product != "criticality" {
		return maps, ErrUnrecognizedContent
	}
	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		return maps, err
	}
	merged := []RegionalFact{}
	seen := map[string]bool{}
	rank := map[string]int{"green": 0, "yellow": 1, "orange": 2, "red": 3}
	for _, f := range maps.Facts {
		var intervals []RegionalFact
		peak := -1
		for _, row := range rows.Facts {
			if row.Risk != f.Risk || row.Zone != f.Zone || row.Precision != "interval" || row.Start == nil || row.End == nil || f.Date == nil {
				continue
			}
			day := time.Date(f.Date.Year(), f.Date.Month(), f.Date.Day(), 0, 0, 0, 0, loc)
			if !row.End.After(day) || !row.Start.Before(day.AddDate(0, 0, 1)) {
				continue
			}
			for _, previous := range intervals {
				if row.Level != previous.Level && row.Start.Before(*previous.End) && previous.Start.Before(*row.End) {
					return maps, fmt.Errorf("conflicting overlapping explicit criticality intervals")
				}
			}
			intervals = append(intervals, row)
			if value, ok := rank[row.Level]; ok && value > peak {
				peak = value
			}
		}
		if len(intervals) > 0 {
			if value, ok := rank[f.Level]; f.Level != "unknown" && (!ok || value != peak) {
				return maps, fmt.Errorf("explicit table/map level conflict")
			}
		} else {
			intervals = append(intervals, f)
		}
		for _, fact := range intervals {
			if fact.Precision == "interval" {
				fact.Locator = f.Locator + "; precise validity from explicit HTML criticality row: " + fact.Original
			}
			key := fmt.Sprintf("%s:%s:%s:%s:%v:%v:%v", fact.Risk, fact.Zone, fact.Level, fact.Precision, fact.Start, fact.End, fact.Date)
			if seen[key] {
				continue
			}
			seen[key] = true
			merged = append(merged, fact)
		}
	}
	maps.Facts = merged
	return maps, nil
}
