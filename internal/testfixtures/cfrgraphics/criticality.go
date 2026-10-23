package cfrgraphics

import (
	"fmt"
	"strings"
)

// Criticality is a synthetic edition shaped like the 8 October 2026 case: zone
// A4 thunderstorms is orange "in corso" until 17.00, then yellow until the next
// day at 14.00. Rectangles stand in for zone outlines; every other zone and risk
// is green, coastal waves are a not-applicable white mask.
func Criticality() ([]byte, []byte, map[int][]byte) {
	const thunderstorms = "TEMPORALI FORTI CON RISCHIO IDROGEOLOGICO-IDRAULICO RETICOLO MINORE"
	html := []byte(`<h1>BOLLETTINO DI VALUTAZIONE DELLE CRITICITÀ</h1><p>Emissione di Giovedì, 08 Ottobre 2026, ore 13.17</p>
<table><tr><th>ZONE DI ALLERTA</th><th>RISCHIO</th><th>TEMPI</th><th>CRITICITÀ</th></tr>
<tr><td>A4</td><td>` + thunderstorms + `</td><td>in corso fino alle ore 17.00 di Giovedì, 08 Ottobre 2026</td><td>ARANCIONE</td></tr>
<tr><td>A4</td><td>` + thunderstorms + `</td><td>dalle ore 17.00 di Giovedì, 08 Ottobre 2026 alle ore 14.00 di Venerdì, 09 Ottobre 2026</td><td>GIALLO</td></tr></table>`)
	labels := []string{"RISCHIO IDROGEOLOGICO-IDRAULICO RETICOLO MINORE", "RISCHIO IDRAULICO RETICOLO PRINCIPALE", thunderstorms, "RISCHIO VENTO", "RISCHIO MAREGGIATE", "RISCHIO NEVE", "RISCHIO GHIACCIO"}
	var box, svg strings.Builder
	box.WriteString(`<doc><page width="600" height="4000">`)
	word := func(x, y float64, text string) {
		fmt.Fprintf(&box, `<word xMin="%g" yMin="%g" xMax="%g" yMax="%g">%s</word>`, x, y, x+6, y+6, text)
	}
	word(10, 10, "Giovedì, 08 Ottobre 2026, ore 13.17")
	svg.WriteString(`<svg>`)
	zones := strings.Fields("A1 A2 A3 A4 A5 A6 B C E1 E2 E3 F1 F2 I L M O1 O2 O3 R1 R2 S1 S2 S3 T V")
	for risk, label := range labels {
		y := float64(40 + risk*400)
		word(100, y, label)
		word(10, y+12, "Giovedì, 08 Ottobre 2026")
		word(310, y+12, "Venerdì, 09 Ottobre 2026")
		for side := 0; side < 2; side++ {
			for i, z := range zones {
				x := float64(15 + side*300 + (i%5)*48)
				top := y + 25 + float64(i/5)*35
				word(x+5, top+5, z)
				fill := "rgb(60%, 80%, 20%)"
				switch {
				case risk == 4:
					fill = "rgb(100%, 100%, 100%)"
				case risk == 2 && z == "A4" && side == 0:
					fill = "rgb(100%, 64.7%, 0%)"
				case risk == 2 && z == "A4":
					fill = "rgb(100%, 100%, 0%)"
				}
				fmt.Fprintf(&svg, `<path fill="%s" fill-opacity="1" stroke="rgb(0%%, 0%%, 0%%)" transform="matrix(1,0,0,1,0,0)" d="M %g %g L %g %g L %g %g L %g %g Z"/>`, fill, x, top, x+30, top, x+30, top+25, x, top+25)
			}
		}
	}
	box.WriteString(`</page></doc>`)
	svg.WriteString(`</svg>`)
	return html, []byte(box.String()), map[int][]byte{1: []byte(svg.String())}
}
