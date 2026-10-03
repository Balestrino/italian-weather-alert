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
	if e != nil || len(merged.Facts) != 1 || merged.Facts[0].Precision != "interval" || !merged.Facts[0].Start.Equal(start) {
		t.Fatal("precise validity replaced or duplicated", merged, e)
	}
	maps.Facts[0].Level = "green"
	if _, e = MergeCriticalityMaps(maps, rows); e == nil {
		t.Fatal("explicit map/table conflict accepted")
	}
}
