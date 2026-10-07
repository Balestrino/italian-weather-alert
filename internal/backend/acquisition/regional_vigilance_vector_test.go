package acquisition

import (
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/testfixtures/cfrgraphics"
)

func vigilanceVectorFixture() ([]byte, VectorEvidence) {
	html, bbox, pages := cfrgraphics.Vigilance()
	return html, VectorEvidence{BBox: bbox, Pages: pages}
}

func TestVigilanceMapsKeepAmountsSymbolsAndSevenRiskSemantics(t *testing.T) {
	html, ev := vigilanceVectorFixture()
	p, err := ProjectVigilanceVector(html, "https://regional.example/vigilance.pdf", ev)
	if err != nil || len(p.Facts) != 364 || len(p.Limitations) != 0 {
		t.Fatalf("complete vigilance: facts=%d limits=%v err=%v", len(p.Facts), p.Limitations, err)
	}
	for _, f := range p.Facts {
		if f.Level != "not_applicable" || f.EvidenceURL == "" || f.Page != 1 || f.Weather == nil {
			t.Fatalf("vigilance became an alert or lost evidence: %+v", f)
		}
		w := f.Weather
		if w.Phenomenon == "rainfall" {
			want := "0 - 10"
			if f.Date.Day() == 4 {
				want = "10 - 20"
			}
			if w.RainfallBand != want || w.TotalRainfallBand != "40 - 60" || w.TotalPeriod != "TOTALE: dalle 12 di oggi alle 24 di domani" || w.Unit != "mm" || w.AmountScope != "area_average" {
				t.Fatalf("rainfall semantics: %+v", w)
			}
		} else {
			want := "not_depicted"
			if f.Zone == "A4" && f.Date.Day() == 3 {
				want = "depicted"
			}
			if w.GraphicalStatus != want {
				t.Fatalf("map or legend symbol misattributed: %s %s %s %+v", f.Risk, f.Zone, f.Date, w)
			}
		}
	}
}

func TestVigilanceUnknownCumulativeFillDoesNotEraseDailyBand(t *testing.T) {
	html, ev := vigilanceVectorFixture()
	// Change only map fills, retaining the distinct labelled legend swatch.
	ev.Pages[1] = []byte(strings.Replace(string(ev.Pages[1]), `fill="rgb(0%, 59.999084%, 100%)"`, `fill="rgb(27%, 70%, 100%)"`, 26))
	p, err := ProjectVigilanceVector(html, "print", ev)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range p.Facts {
		if f.Weather.Phenomenon == "rainfall" && (f.Weather.RainfallBand == "" || f.Weather.GraphicalStatus != "depicted" || f.Weather.TotalRainfallBand != "" || f.Weather.TotalGraphicalStatus != "unresolved") {
			t.Fatal("daily evidence lost or unsupported total inferred", f)
		}
	}
	if len(p.Limitations) < 2 {
		t.Fatal("unresolved component not explained", p.Limitations)
	}
}

func TestVigilanceRejectsMissingLegendPanelAndEdition(t *testing.T) {
	for _, mutate := range []func(*VectorEvidence){
		func(e *VectorEvidence) { e.BBox = []byte(strings.Replace(string(e.BBox), "ore 10.22", "ore 11.22", 1)) },
		func(e *VectorEvidence) {
			e.BBox = []byte(strings.Replace(string(e.BBox), ">A4</word>", ">Z9</word>", 1))
		},
		func(e *VectorEvidence) {
			e.BBox = []byte(strings.Replace(string(e.BBox), ">120</word>", ">121</word>", 1))
		},
		func(e *VectorEvidence) { e.Pages[1] = nil },
		func(e *VectorEvidence) {
			e.BBox = []byte(strings.Replace(string(e.BBox), "TOTALE: dalle 12", "TOTALE: dalle 06", 1))
		},
		func(e *VectorEvidence) {
			e.BBox = []byte(strings.Replace(string(e.BBox), "Domenica, 04", "Domenica, 05", 1))
		},
		func(e *VectorEvidence) {
			e.Pages[1] = []byte(strings.Replace(string(e.Pages[1]), "<svg>", `<svg><image href="uninterpreted-symbol.png"/>`, 1))
		},
		func(e *VectorEvidence) { e.Pages[1] = []byte(`<svg/>`) },
	} {
		html, ev := vigilanceVectorFixture()
		mutate(&ev)
		if _, err := ProjectVigilanceVector(html, "print", ev); err == nil {
			t.Fatal("unsupported vigilance accepted")
		}
	}
}

func TestVigilanceUnlistedBlackMarkRemainsUnresolved(t *testing.T) {
	html, ev := vigilanceVectorFixture()
	ev.Pages[1] = []byte(strings.Replace(string(ev.Pages[1]), `</svg>`, `<path fill="rgb(0%, 0%, 0%)" fill-opacity="1" d="M 41 135 L 45 135 L 43 141 Z"/></svg>`, 1))
	p, err := ProjectVigilanceVector(html, "print", ev)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range p.Facts {
		if f.Risk == "thunderstorms" && f.Zone == "A6" && f.Date.Day() == 3 {
			found = f.Weather.GraphicalStatus == "unresolved" && !f.Weather.UnderEvaluation
		}
	}
	if !found || len(p.Limitations) == 0 {
		t.Fatal("unlisted unrecognized black mark treated as absence")
	}
}

func TestVigilanceAmbiguousLegendDoesNotIdentifyAWeatherSymbol(t *testing.T) {
	html, ev := vigilanceVectorFixture()
	ev.Pages[1] = []byte(strings.ReplaceAll(string(ev.Pages[1]), "rgb(30%, 0%, 50%)", "rgb(10%, 0%, 50%)"))
	p, err := ProjectVigilanceVector(html, "print", ev)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range p.Facts {
		if f.Zone == "A4" && f.Date.Day() == 3 && (f.Risk == "wind" || f.Risk == "snow") && f.Weather.GraphicalStatus != "unresolved" {
			t.Fatal("ambiguous shared symbol classified", f)
		}
	}
	if len(p.Limitations) == 0 {
		t.Fatal("ambiguous legend hidden")
	}
}

func TestVigilanceUnresolvedSymbolDoesNotBecomeAbsence(t *testing.T) {
	html, ev := vigilanceVectorFixture()
	ev.Pages[1] = []byte(strings.Replace(string(ev.Pages[1]), `fill="rgb(100%, 100%, 0%)"`, `fill="rgb(30%, 30%, 30%)"`, 1))
	p, err := ProjectVigilanceVector(html, "print", ev)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range p.Facts {
		if f.Risk == "thunderstorms" && f.Zone == "A4" && f.Date.Day() == 3 && f.Weather.GraphicalStatus != "unresolved" {
			t.Fatal("missing evaluated phenomenon treated as absent", f.Weather)
		}
	}
	if len(p.Limitations) == 0 {
		t.Fatal("missing graphical support hidden")
	}
}
