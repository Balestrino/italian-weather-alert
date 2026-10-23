package acquisition

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// Redistributable synthetic vector maps deliberately use rectangles, not a
// captured regional outline, to exercise geometry, dates and all seven risks.
func vectorFixture() ([]byte, VectorEvidence) {
	html := []byte(`<p>Emissione di Sabato, 03 Ottobre 2026, ore 10.52</p>Criticità previste: NESSUNA`)
	var box, svg strings.Builder
	box.WriteString(`<doc><page width="600" height="4000">`)
	word := func(x, y float64, text string) {
		fmt.Fprintf(&box, `<word xMin="%g" yMin="%g" xMax="%g" yMax="%g">%s</word>`, x, y, x+6, y+6, text)
	}
	word(10, 10, "Sabato, 03 Ottobre 2026, ore 10.52")
	svg.WriteString(`<svg>`)
	zones := strings.Fields("A1 A2 A3 A4 A5 A6 B C E1 E2 E3 F1 F2 I L M O1 O2 O3 R1 R2 S1 S2 S3 T V")
	for risk, label := range criticalityRiskLabels {
		y := float64(40 + risk*400)
		word(100, y, label)
		word(10, y+12, "Sabato, 03 Ottobre 2026")
		word(310, y+12, "Domenica, 04 Ottobre 2026")
		for side := 0; side < 2; side++ {
			for i, z := range zones {
				x := float64(15 + side*300 + (i%5)*48)
				top := y + 25 + float64(i/5)*35
				word(x+5, top+5, z)
				fill := "rgb(60%, 80%, 20%)"
				if risk == 4 {
					fill = "rgb(100%, 100%, 100%)"
				}
				fmt.Fprintf(&svg, `<path fill="%s" fill-opacity="1" stroke="rgb(0%%, 0%%, 0%%)" transform="matrix(1,0,0,1,0,0)" d="M %g %g L %g %g L %g %g L %g %g Z"/>`, fill, x, top, x+30, top, x+30, top+25, x, top+25)
			}
		}
	}
	box.WriteString(`</page></doc>`)
	svg.WriteString(`</svg>`)
	return html, VectorEvidence{BBox: []byte(box.String()), Pages: map[int][]byte{1: []byte(svg.String())}}
}
func TestVectorCriticalityUsesLabelledPolygons(t *testing.T) {
	html, ev := vectorFixture()
	p, e := ProjectCriticalityVector(html, "https://regional.example/print.pdf", ev)
	if e != nil || len(p.Facts) != 364 || len(p.Limitations) != 0 {
		t.Fatal("complete maps", len(p.Facts), e, p.Limitations)
	}
	for _, f := range p.Facts {
		want := "green"
		if f.Risk == "coastal_waves" {
			want = "not_applicable"
		}
		if f.Level != want || f.EvidenceURL == "" || f.Page != 1 {
			t.Fatal("risk/date/zone attribution", f)
		}
	}
	// A filled outline is required: whitespace outside a missing zone is unknown.
	ev.Pages[1] = []byte(strings.Replace(string(ev.Pages[1]), `fill="rgb(60%, 80%, 20%)"`, `fill="rgb(20%, 20%, 90%)"`, 1))
	p, e = ProjectCriticalityVector(html, "print", ev)
	if e != nil || p.Facts[0].Level != "unknown" {
		t.Fatal("unknown color inferred", e)
	}
	_, original := vectorFixture()
	ev = original
	ev.Pages[1] = []byte(strings.Replace(string(ev.Pages[1]), `fill="rgb(60%, 80%, 20%)"`, `fill="rgb(100%, 100%, 0%)"`, 1))
	if _, e = ProjectCriticalityVector(html, "print", ev); e == nil {
		t.Fatal("yellow map accepted against no-criticality statement")
	}
	for _, level := range []struct{ color, want string }{{"rgb(100%, 100%, 0%)", "yellow"}, {"rgb(100%, 64.7%, 0%)", "orange"}, {"rgb(100%, 0%, 0%)", "red"}} {
		if got := vectorColor(level.color, 0); got != level.want {
			t.Fatal(got, level.want)
		}
	}
	if vectorColor("rgb(100%, 100%, 100%)", 0) != "unknown" {
		t.Fatal("white on applicable risk inferred")
	}
}
func TestVectorRejectsEditionAndIncompletePanels(t *testing.T) {
	html, ev := vectorFixture()
	ev.BBox = []byte(strings.Replace(string(ev.BBox), "ore 10.52", "ore 11.52", 1))
	if _, e := ProjectCriticalityVector(html, "print", ev); e == nil {
		t.Fatal("different issuance accepted")
	}
	_, ev = vectorFixture()
	ev.BBox = []byte(strings.Replace(string(ev.BBox), ">A4</word>", ">Z9</word>", 1))
	if _, e := ProjectCriticalityVector(html, "print", ev); e == nil {
		t.Fatal("missing zone accepted")
	}
	_, ev = vectorFixture()
	ev.Pages[1] = nil
	if _, e := ProjectCriticalityVector(html, "print", ev); e == nil {
		t.Fatal("missing map accepted")
	}
}

func TestVectorMapsKeepExplicitAlertIntervals(t *testing.T) {
	start := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Hour)
	day := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	f := RegionalFact{Risk: "wind", Zone: "A4", Level: "yellow", Precision: "date", Date: &day, Original: "Sabato, 03 Ottobre 2026"}
	next := day.Add(24 * time.Hour)
	second := f
	second.Date = &next
	second.Original = "Domenica, 04 Ottobre 2026"
	maps := RegionalProjection{Product: "criticality", Facts: []RegionalFact{f, second}}
	rows := RegionalProjection{Product: "criticality", Facts: []RegionalFact{{Risk: "wind", Zone: "A4", Level: "yellow", Precision: "interval", Start: &start, End: &end, Original: "explicit afternoon validity"}}}
	merged, e := MergeCriticalityMaps(maps, rows)
	if e != nil || len(merged.Facts) != 3 || merged.Facts[0].Precision != "interval" || !merged.Facts[0].Start.Equal(start) {
		t.Fatal("precise validity replaced or duplicated", merged, e)
	}
	// The rest of each mapped local day is green, derived and labelled as such.
	for i, want := range [][2]string{{"2026-10-02T22:00:00Z", "2026-10-03T10:00:00Z"}, {"2026-10-04T16:00:00Z", "2026-10-04T22:00:00Z"}} {
		g := merged.Facts[i+1]
		if g.Level != "green" || g.Precision != "interval" || g.Start.Format(time.RFC3339) != want[0] || g.End.Format(time.RFC3339) != want[1] || g.Assumption != DerivedGreenAssumption || !strings.HasSuffix(g.Locator, "derived: no listed alert for this period") {
			t.Fatal("rest of the day not derived green", g)
		}
	}
	maps.Facts[0].Level = "green"
	if _, e = MergeCriticalityMaps(maps, rows); e == nil {
		t.Fatal("explicit map/table conflict accepted")
	}
}

func TestDailyPeakPreservesSequentialIntervals(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Rome")
	day := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 10, 7, 15, 0, 0, 0, loc)
	transition := start.Add(2 * time.Hour)
	end := transition.Add(24 * time.Hour)
	finish := time.Date(2026, 10, 9, 0, 0, 0, 0, loc)
	row := func(level string, from, to *time.Time) RegionalFact {
		return RegionalFact{Risk: "thunderstorms", Zone: "A4", Level: level, Precision: "interval", Start: from, End: to, Original: level}
	}
	next := day.AddDate(0, 0, 1)
	maps := RegionalProjection{Product: "criticality", Facts: []RegionalFact{
		{Risk: "thunderstorms", Zone: "A4", Level: "orange", Precision: "date", Date: &day},
		{Risk: "thunderstorms", Zone: "A4", Level: "orange", Precision: "date", Date: &next},
	}}
	rows := RegionalProjection{Product: "criticality", Facts: []RegionalFact{row("yellow", &start, &transition), row("orange", &transition, &end), row("yellow", &end, &finish)}}
	got, err := MergeCriticalityMaps(maps, rows)
	if err != nil || len(got.Facts) != 4 || !got.Facts[0].Start.Equal(start) || !got.Facts[3].End.Equal(finish) {
		t.Fatal("sequential levels or exact intervals lost", got, err)
	}
	// Only the morning before the first alert is green; the next day is fully covered.
	if g := got.Facts[2]; g.Level != "green" || !g.End.Equal(start) || g.Start.Format(time.RFC3339) != "2026-10-06T22:00:00Z" {
		t.Fatal("morning remainder not derived green", g)
	}
	third := next.AddDate(0, 0, 1)
	maps.Facts = append(maps.Facts, RegionalFact{Risk: "thunderstorms", Zone: "A4", Level: "green", Precision: "date", Date: &third})
	got, err = MergeCriticalityMaps(maps, rows)
	if err != nil || len(got.Facts) != 5 || got.Facts[4].Precision != "date" {
		t.Fatal("midnight end leaked into following day", got, err)
	}
	rows.Facts[0].End = &end
	if _, err = MergeCriticalityMaps(maps, rows); err == nil {
		t.Fatal("contradictory overlapping intervals accepted")
	}
}

func TestVectorSeparatedFillAndScopedOutsideLabels(t *testing.T) {
	shapes, err := vectorPolygons([]byte(`<svg><path fill="rgb(60%, 80%, 20%)" fill-opacity="1" d="M 0 0 L 20 0 L 20 20 L 0 20 Z"/></svg>`))
	if err != nil || len(shapes) != 1 {
		t.Fatal("separated fill discarded", shapes, err)
	}
	for _, tc := range []struct {
		label      string
		xmin, ymin float64
		want       bool
	}{
		{"L", 5, 5, true}, {"A6", -2.1, 5, true}, {"I", 5, -5, true},
		{"L", 5, -5, false}, {"I", 5, -20, false},
	} {
		_, ok := vectorLabelPolygon(shapes, vectorWord{XMin: tc.xmin, XMax: tc.xmin + 4, YMin: tc.ymin, YMax: tc.ymin + 6, Text: tc.label})
		if ok != tc.want {
			t.Fatalf("unsupported offset %s: %v", tc.label, ok)
		}
	}
	_, ok := vectorLabelPolygon(append(shapes, shapes[0]), vectorWord{XMin: 5, XMax: 9, YMin: 5, YMax: 11, Text: "A6"})
	if ok {
		t.Fatal("overlapping shapes accepted")
	}
	parts, err := vectorPolygons([]byte(`<svg><path fill="rgb(60%, 80%, 20%)" stroke="rgb(0%, 0%, 0%)" fill-opacity="1" transform="matrix(2,0,0,2,10,20)" d="M 0 0 L 2 0 L 2 2 L 0 2 Z M 5 0 L 7 0 L 7 2 L 5 2 Z"/></svg>`))
	if err != nil || len(parts) != 2 || vectorContains(parts[0], vectorPoint{18, 22}) || !vectorContains(parts[1], vectorPoint{22, 22}) {
		t.Fatal("compound paths or affine coordinates conflated", parts, err)
	}
}

func TestVectorHistoricalPrefaceAndNonGreenMaps(t *testing.T) {
	html, ev := vectorFixture()
	html = []byte(strings.Replace(string(html), "Criticità previste: NESSUNA", "", 1))
	page := strings.TrimSuffix(strings.TrimPrefix(string(ev.BBox), "<doc>"), "</doc>")
	ev.BBox = []byte(`<doc><page width="600" height="800"><word xMin="10" yMin="10" xMax="30" yMax="20">Idraulico reticolo principale</word></page>` + page + `</doc>`)
	ev.Pages[2] = []byte(strings.Replace(string(ev.Pages[1]), `fill="rgb(60%, 80%, 20%)"`, `fill="rgb(100%, 100%, 0%)"`, 1))
	delete(ev.Pages, 1)
	p, err := ProjectCriticalityVector(html, "print", ev)
	if err != nil || len(p.Facts) != 364 || p.Facts[0].Level != "yellow" || p.Facts[0].Page != 2 {
		t.Fatal("preface mistaken for map or original physical page lost", err)
	}
	ev.BBox = []byte(strings.Replace(string(ev.BBox), "Domenica, 04", "Domenica, 05", 1))
	if _, err := ProjectCriticalityVector(html, "print", ev); err == nil {
		t.Fatal("unrelated map date accepted")
	}
}

func TestVectorBackgroundAndRasterCannotSupplyLevels(t *testing.T) {
	html, ev := vectorFixture()
	// A page background cannot substitute for a missing outlined map zone.
	ev.Pages[1] = []byte(strings.Replace(string(ev.Pages[1]), `stroke="rgb(0%, 0%, 0%)"`, `stroke=""`, 1))
	p, err := ProjectCriticalityVector(html, "print", ev)
	if err != nil || p.Facts[0].Level != "unknown" {
		t.Fatal("standalone rectangle supplied a zone", err)
	}
	_, ev = vectorFixture()
	ev.Pages[1] = []byte(strings.Replace(string(ev.Pages[1]), `<svg>`, `<svg><defs><image id="raster" width="10" height="10"/></defs><use href="#raster" transform="matrix(1,0,0,1,20,80)"/>`, 1))
	if _, err := ProjectCriticalityVector(html, "print", ev); err == nil {
		t.Fatal("raster inside map ignored")
	}
	ev.Pages[1] = []byte(strings.Replace(string(ev.Pages[1]), "1,20,80)", "1,20,0)", 1))
	if _, err := ProjectCriticalityVector(html, "print", ev); err != nil {
		t.Fatal("logo outside maps rejected", err)
	}
}

func TestCriticalityHeaderCurvesCannotHideMapGeometry(t *testing.T) {
	html, ev := vectorFixture()
	header := `<path fill="rgb(100%, 100%, 100%)" fill-opacity="1" d="M 10 10 C 10 5 15 5 15 10 C 15 15 10 15 10 10 Z"/>`
	ev.Pages[1] = []byte(strings.Replace(string(ev.Pages[1]), "<svg>", "<svg>"+header, 1))
	p, err := ProjectCriticalityVector(html, "retained.pdf", ev)
	if err != nil || len(p.Facts) != 364 {
		t.Fatal("header curve discarded maps", len(p.Facts), err)
	}
	for _, modified := range []string{
		strings.Replace(header, `d="`, `transform="matrix(1,0,0,1,0,50)" d="`, 1),
		strings.Replace(header, "10 5 15 5", "10 70 15 5", 1),
		strings.Replace(header, "C 10 5", "Q 10 5", 1),
	} {
		_, ev = vectorFixture()
		ev.Pages[1] = []byte(strings.Replace(string(ev.Pages[1]), "<svg>", "<svg>"+modified, 1))
		if _, err = ProjectCriticalityVector(html, "retained.pdf", ev); err == nil {
			t.Fatal("unsupported curve inside map or unknown command accepted", modified)
		}
	}
	if _, err := vectorShapes([]byte("<svg>" + header + "</svg>")); err == nil {
		t.Fatal("generic shape parser became permissive")
	}
}

// The 8 October 2026 shape: an ongoing orange row followed by yellow.
func TestOngoingRowsMergeWithDailyMaps(t *testing.T) {
	rows, err := ProjectRegionalHTML("criticality", ongoingCriticalityFixture(`<p>Emissione di <b>Giovedì, 08 Ottobre 2026</b>, ore <b>13.17</b></p>`,
		`<tr><td>A4</td><td>TEMPORALI FORTI CON RISCHIO IDROGEOLOGICO-IDRAULICO RETICOLO MINORE</td><td>in corso fino alle ore 17.00 di Giovedì, 08 Ottobre 2026</td><td>ARANCIONE</td></tr>
<tr><td>A4</td><td>TEMPORALI FORTI CON RISCHIO IDROGEOLOGICO-IDRAULICO RETICOLO MINORE</td><td>dalle ore 17.00 di Giovedì, 08 Ottobre 2026 alle ore 14.00 di Venerdì, 09 Ottobre 2026</td><td>GIALLO</td></tr>`))
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	next := day.AddDate(0, 0, 1)
	maps := RegionalProjection{Product: "criticality", Facts: []RegionalFact{
		{Risk: "thunderstorms", Zone: "A4", Level: "orange", Precision: "date", Date: &day, Original: "Giovedì, 08 Ottobre 2026"},
		{Risk: "thunderstorms", Zone: "A4", Level: "yellow", Precision: "date", Date: &next, Original: "Venerdì, 09 Ottobre 2026"},
	}}
	merged, err := MergeCriticalityMaps(maps, rows)
	if err != nil {
		t.Fatal("ongoing peak raised a false conflict", err)
	}
	var orange, yellow int
	for _, f := range merged.Facts {
		if f.Precision == "date" {
			t.Fatal("daily map replaced an explicit interval", f)
		}
		switch f.Level {
		case "orange":
			orange++
			if f.End.Format(time.RFC3339) != "2026-10-08T15:00:00Z" || f.Assumption != OngoingRowAssumption {
				t.Fatal("orange outlasts its stated end", f)
			}
		case "yellow":
			yellow++
		}
	}
	if orange != 1 || yellow != 1 {
		t.Fatal("explicit intervals lost", merged.Facts)
	}
	// A contradictory overlap still fails closed.
	clash := *rows.Facts[1].Start
	clashEnd := clash.Add(time.Hour)
	rows.Facts = append(rows.Facts, RegionalFact{Risk: "thunderstorms", Zone: "A4", Level: "orange", Precision: "interval", Start: &clash, End: &clashEnd})
	if _, err = MergeCriticalityMaps(maps, rows); err == nil {
		t.Fatal("contradictory overlapping intervals accepted")
	}
}

func TestRemainderGreenOutsideListedIntervals(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Rome")
	at := func(y int, m time.Month, d, h, min int) *time.Time {
		v := time.Date(y, m, d, h, min, 0, 0, loc)
		return &v
	}
	date := func(y int, m time.Month, d int) *time.Time {
		v := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
		return &v
	}
	interval := func(level string, from, to *time.Time) RegionalFact {
		return RegionalFact{Risk: "thunderstorms", Zone: "A4", Level: level, Precision: "interval", Start: from, End: to}
	}
	merge := func(mapLevel string, day *time.Time, rows ...RegionalFact) []RegionalFact {
		t.Helper()
		got, err := MergeCriticalityMaps(RegionalProjection{Product: "criticality", Facts: []RegionalFact{{Risk: "thunderstorms", Zone: "A4", Level: mapLevel, Precision: "date", Date: day, Original: "heading"}}},
			RegionalProjection{Product: "criticality", Facts: rows})
		if err != nil {
			t.Fatal(err)
		}
		return got.Facts
	}
	levelAt := func(facts []RegionalFact, instant time.Time) string {
		for _, f := range facts {
			if f.Precision == "interval" && !instant.Before(*f.Start) && instant.Before(*f.End) {
				return f.Level
			}
		}
		return ""
	}
	// Evening-only alert ending at 23.59: green morning and afternoon, no one-minute green at midnight.
	evening := merge("yellow", date(2026, 10, 20), interval("yellow", at(2026, 10, 20, 18, 0), at(2026, 10, 20, 23, 59)))
	if len(evening) != 2 || levelAt(evening, *at(2026, 10, 20, 15, 0)) != "green" || levelAt(evening, *at(2026, 10, 20, 19, 0)) != "yellow" {
		t.Fatal("evening alert remainder", evening)
	}
	// Ongoing morning alert: green after its end.
	morning := merge("orange", date(2026, 10, 8), RegionalFact{Risk: "thunderstorms", Zone: "A4", Level: "orange", Precision: "interval", Start: at(2026, 10, 8, 13, 17), End: at(2026, 10, 8, 17, 0), Assumption: OngoingRowAssumption})
	if levelAt(morning, *at(2026, 10, 8, 17, 1)) != "green" || levelAt(morning, *at(2026, 10, 8, 16, 59)) != "orange" || levelAt(morning, *at(2026, 10, 8, 12, 0)) != "green" {
		t.Fatal("ongoing alert remainder", morning)
	}
	// 25-hour local day at the end of daylight saving time.
	dst := merge("yellow", date(2026, 10, 25), interval("yellow", at(2026, 10, 25, 18, 0), at(2026, 10, 25, 23, 59)))
	if g := dst[1]; g.Start.Format(time.RFC3339) != "2026-10-24T22:00:00Z" || g.End.Format(time.RFC3339) != "2026-10-25T17:00:00Z" {
		t.Fatal("DST day remainder", g)
	}
	// Unresolved map: intervals are kept, no green is derived.
	unknown := merge("unknown", date(2026, 10, 20), interval("yellow", at(2026, 10, 20, 18, 0), at(2026, 10, 20, 23, 59)))
	if len(unknown) != 1 || unknown[0].Level != "yellow" {
		t.Fatal("green derived from an unresolved map", unknown)
	}
	// Not-applicable mask without intervals keeps its daily fact; with an interval it is a conflict.
	if na := merge("not_applicable", date(2026, 10, 20)); len(na) != 1 || na[0].Level != "not_applicable" || na[0].Precision != "date" {
		t.Fatal("not-applicable mask changed", na)
	}
	if _, err := MergeCriticalityMaps(RegionalProjection{Product: "criticality", Facts: []RegionalFact{{Risk: "thunderstorms", Zone: "A4", Level: "not_applicable", Precision: "date", Date: date(2026, 10, 20)}}},
		RegionalProjection{Product: "criticality", Facts: []RegionalFact{interval("yellow", at(2026, 10, 20, 18, 0), at(2026, 10, 20, 23, 59))}}); err == nil {
		t.Fatal("alert on a not-applicable mask accepted")
	}
}
