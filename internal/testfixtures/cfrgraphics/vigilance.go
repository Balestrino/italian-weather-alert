package cfrgraphics

import (
	"fmt"
	"strings"
)

// Synthetic rectangular geography and invented symbols are redistributable.
// The three rainfall panels deliberately carry different daily/total bands.
func Vigilance() ([]byte, []byte, map[int][]byte) {
	var html, box, svg strings.Builder
	html.WriteString(`<h1>Bollettino di Vigilanza Meteorologica Regionale</h1><p>Emissione di Sabato, 03 Ottobre 2026, ore 10.22</p><table><tr><th></th><th>Sabato, 03 Ottobre 2026</th><th>Domenica, 04 Ottobre 2026</th></tr>`)
	for _, phenomenon := range []string{"Pioggia", "Temporali", "Vento", "Mare", "Neve", "Ghiaccio"} {
		zones := ""
		if phenomenon != "Pioggia" {
			zones = "A4"
		}
		fmt.Fprintf(&html, `<tr><td>%s</td><td>%s</td><td></td></tr>`, phenomenon, zones)
	}
	html.WriteString(`</table>`)
	box.WriteString(`<doc><page width="620" height="1000">`)
	svg.WriteString(`<svg>`)
	word := func(x, y float64, text string) {
		fmt.Fprintf(&box, `<word xMin="%g" yMin="%g" xMax="%g" yMax="%g">%s</word>`, x, y, x+12, y+6, text)
	}
	path := func(fill, stroke string, x, y, width, height float64) {
		fmt.Fprintf(&svg, `<path fill="%s" stroke="%s" fill-opacity="1" d="M %g %g L %g %g L %g %g L %g %g Z"/>`, fill, stroke, x, y, x+width, y, x+width, y+height, x, y+height)
	}
	icon := func(fill, stroke string, x, y float64) {
		fmt.Fprintf(&svg, `<path fill="%s" stroke="%s" fill-opacity="1" d="M %g %g L %g %g L %g %g Z"/>`, fill, stroke, x, y, x+6, y, x+3, y+6)
	}
	word(20, 10, "Sabato, 03 Ottobre 2026, ore 10.22")
	word(20, 40, "FENOMENI PIOGGIA e TEMPORALI")
	word(20, 60, "Sabato, 03 Ottobre 2026")
	word(220, 60, "Domenica, 04 Ottobre 2026")
	word(420, 50, "TOTALE: dalle 12 di oggi")
	word(420, 60, "alle 24 di domani")
	zones := strings.Fields("A1 A2 A3 A4 A5 A6 B C E1 E2 E3 F1 F2 I L M O1 O2 O3 R1 R2 S1 S2 S3 T V")
	for panel := 0; panel < 5; panel++ {
		side, top := panel, 80.0
		if panel > 2 {
			side, top = panel-3, 550
		}
		for i, zone := range zones {
			x, y := float64(22+side*200+(i%5)*36), top+float64(i/5)*36
			word(x+3, y+3, zone)
			fill := "rgb(100%, 100%, 100%)"
			if panel == 1 {
				fill = "rgb(0%, 100%, 100%)"
			}
			if panel == 2 {
				fill = "rgb(0%, 59.999084%, 100%)"
			}
			path(fill, "rgb(0%, 0%, 0%)", x, y, 30, 30)
			if zone == "A4" && panel == 0 {
				icon("rgb(100%, 100%, 0%)", "rgb(100%, 33.332825%, 0%)", x+18, y+18)
			}
			if zone == "A4" && panel == 3 {
				for n := 0; n < 4; n++ {
					icon(fmt.Sprintf("rgb(%d%%, 0%%, 50%%)", 10+n*10), "rgb(20%, 20%, 20%)", x+17+float64(n%2)*7, y+14+float64(n/2)*7)
				}
			}
		}
	}
	word(20, 320, "Temporali")
	icon("rgb(100%, 100%, 0%)", "rgb(100%, 33.332825%, 0%)", 36, 320)
	word(20, 350, "Cumulato medio sull'area [mm]")
	bands := []string{"0 - 10", "10 - 20", "20 - 40", "40 - 60", "60 - 80", "80 - 100", "100 - 120", "> 120"}
	colors := []string{"rgb(100%, 100%, 100%)", "rgb(0%, 100%, 100%)", "rgb(59.999084%, 79.998779%, 100%)", "rgb(0%, 59.999084%, 100%)", "rgb(0%, 27.842712%, 100%)", "rgb(15.686035%, 0%, 59.999084%)", "rgb(59.999084%, 39.99939%, 79.998779%)", "rgb(58.03833%, 0%, 41.960144%)"}
	for i, band := range bands {
		x := float64(20 + i*70)
		path(colors[i], "", x, 370, 66, 10)
		for j, text := range strings.Fields(band) {
			word(x+float64(j)*15+2, 372, text)
		}
	}
	word(20, 500, "FENOMENI VENTO, MARE, NEVE e GHIACCIO")
	word(20, 525, "Sabato, 03 Ottobre 2026")
	word(220, 525, "Domenica, 04 Ottobre 2026")
	for i, name := range []string{"Vento", "Mare", "Neve", "Ghiaccio"} {
		x := float64(20 + i*100)
		word(x, 790, name)
		icon(fmt.Sprintf("rgb(%d%%, 0%%, 50%%)", 10+i*10), "rgb(20%, 20%, 20%)", x+16, 790)
	}
	word(20, 850, "Previsione fino alle 24 di domani:")
	box.WriteString(`</page></doc>`)
	svg.WriteString(`</svg>`)
	return []byte(html.String()), []byte(box.String()), map[int][]byte{1: []byte(svg.String())}
}
