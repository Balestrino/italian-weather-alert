package acquisition

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"math"
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
	BBox  []byte
	Pages map[int][]byte
}
type VectorPDFReader interface {
	Read(context.Context, []byte) (VectorEvidence, error)
}
type PopplerVectorReader struct{}

func (PopplerVectorReader) Read(ctx context.Context, body []byte) (VectorEvidence, error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	out := VectorEvidence{Pages: map[int][]byte{}}
	if len(body) > 32<<20 || !bytes.HasPrefix(body, []byte("%PDF-")) {
		return out, ErrUnrecognizedContent
	}
	run := func(name string, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdin = bytes.NewReader(body)
		var w boundedVectorOutput
		cmd.Stdout = &w
		err := cmd.Run()
		return w.Bytes(), err
	}
	var err error
	out.BBox, err = run("pdftotext", "-bbox", "-", "-")
	if err != nil {
		return out, err
	}
	pages, err := vectorPages(out.BBox)
	if err != nil || len(pages) > 12 {
		return out, ErrUnrecognizedContent
	}
	for i, p := range pages {
		if len(vectorHeadings(p)) == 0 {
			continue
		}
		out.Pages[i+1], err = run("pdftocairo", "-svg", "-f", strconv.Itoa(i+1), "-l", strconv.Itoa(i+1), "-", "-")
		if err != nil {
			return out, err
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

type vectorPoint struct{ X, Y float64 }
type vectorPolygon struct {
	Points []vectorPoint
	Fill   string
}

var vectorNumbers = regexp.MustCompile(`[-+]?(?:[0-9]*\.)?[0-9]+(?:[eE][-+]?[0-9]+)?`)

func vectorPolygons(body []byte) ([]vectorPolygon, error) {
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
		if e.Name.Local == "g" && attrs["transform"] != "" {
			return nil, ErrUnrecognizedContent
		}
		if e.Name.Local != "path" || attrs["stroke"] != "rgb(0%, 0%, 0%)" || attrs["fill"] == "none" || attrs["fill"] == "" || attrs["fill-opacity"] != "1" {
			continue
		}
		if !strings.HasPrefix(attrs["transform"], "matrix(") {
			return nil, ErrUnrecognizedContent
		}
		ns := vectorNumbers.FindAllString(attrs["transform"], -1)
		if len(ns) != 6 {
			return nil, ErrUnrecognizedContent
		}
		var m [6]float64
		for i, n := range ns {
			m[i], err = strconv.ParseFloat(n, 64)
			if err != nil {
				return nil, err
			}
		}
		fields := strings.Fields(attrs["d"])
		p := vectorPolygon{Fill: attrs["fill"]}
		for i := 0; i < len(fields); {
			command := fields[i]
			i++
			if command == "Z" || command == "z" {
				continue
			}
			if command != "M" && command != "L" {
				return nil, ErrUnrecognizedContent
			}
			if i+1 >= len(fields) {
				return nil, ErrUnrecognizedContent
			}
			x, e1 := strconv.ParseFloat(fields[i], 64)
			y, e2 := strconv.ParseFloat(fields[i+1], 64)
			i += 2
			if e1 != nil || e2 != nil {
				return nil, ErrUnrecognizedContent
			}
			p.Points = append(p.Points, vectorPoint{m[0]*x + m[2]*y + m[4], m[1]*x + m[3]*y + m[5]})
		}
		if len(p.Points) > 2 {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, ErrUnrecognizedContent
	}
	return out, nil
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
	var first []string
	for _, w := range pages[0].Words {
		first = append(first, w.Text)
	}
	if !strings.Contains(strings.Join(first, " "), obs.IssuanceExpression) {
		return out, fmt.Errorf("PDF/HTML issuance mismatch")
	}
	seen := map[string]bool{}
	riskDays := map[string]bool{}
	days := map[string]bool{}
	panels := map[string]int{}
	unresolved := 0
	for pageIndex, p := range pages {
		headings := vectorHeadings(p)
		if len(headings) == 0 {
			continue
		}
		polygons, e := vectorPolygons(evidence.Pages[pageIndex+1])
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
				point := vectorPoint{(w.XMin + w.XMax) / 2, (w.YMin + w.YMax) / 2}
				level := "unknown"
				matches := 0
				for _, polygon := range polygons {
					if vectorContains(polygon, point) {
						matches++
						level = vectorColor(polygon.Fill, h.Risk)
					}
				}
				if matches != 1 {
					level = "unknown"
				}
				if obs.TextSaysNoCriticality && (level == "yellow" || level == "orange" || level == "red") {
					return out, fmt.Errorf("map conflicts with no-criticality statement")
				}
				if level == "unknown" {
					unresolved++
				}
				out.Facts = append(out.Facts, RegionalFact{Risk: riskIDs[h.Risk], Label: criticalityRiskLabels[h.Risk], Zone: w.Text, Level: level, Original: expressions[side], Precision: "date", Date: dates[side], Locator: fmt.Sprintf("Retained PDF labelled zone %s; filled vector polygon; risk/day heading", w.Text), EvidenceURL: pdfURL, Page: pageIndex + 1})
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
	for _, f := range maps.Facts {
		for _, row := range rows.Facts {
			if row.Risk != f.Risk || row.Zone != f.Zone || row.Precision != "interval" || row.Start == nil || row.End == nil || f.Date == nil {
				continue
			}
			day := f.Date.Format(time.DateOnly)
			if day < row.Start.In(loc).Format(time.DateOnly) || day > row.End.In(loc).Format(time.DateOnly) {
				continue
			}
			if f.Level != "unknown" && f.Level != row.Level {
				return maps, fmt.Errorf("explicit table/map level conflict")
			}
			f.Level = row.Level
			f.Original = row.Original
			f.Precision = row.Precision
			f.Start = row.Start
			f.End = row.End
			f.Date = nil
			f.Locator += "; precise validity from explicit HTML criticality row: " + row.Original
		}
		key := fmt.Sprintf("%s:%s:%s:%s:%s", f.Risk, f.Zone, f.Level, f.Precision, f.Original)
		if seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, f)
	}
	maps.Facts = merged
	return maps, nil
}
