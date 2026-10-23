package acts

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
)

func notice(text string, attachments ...string) classification.Content {
	c := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: "https://comune.example.it/avviso", Role: "original", Text: text}}}
	for i, a := range attachments {
		c.Sections = append(c.Sections, classification.ContentSection{ResourceURL: "https://comune.example.it/allegato.pdf", Role: "attachment", Page: i + 1, Text: a})
	}
	return c
}

const footer = " Albo pretorio Meccanismo di feedback Seguici su Facebook"

func TestDetectReferences(t *testing.T) {
	for name, test := range map[string]struct {
		text        string
		attachments []string
		want        []string
	}{
		"unnumbered criticality notice": {
			text: "Emessa ordinanza di chiusura cimiteri e parchi a partire dalle 0.00 di giovedì 10 settembre. A scopo precauzionale, il Sindaco ha predisposto mediante apposita ordinanza la chiusura al pubblico dei cimiteri." + footer,
			want: []string{"ordinanza -/0"},
		},
		"monitoring citing municipal ordinance": {
			text: "Restano in vigore le disposizioni adottate con ordinanza comunale: chiusura al pubblico dei cimiteri.",
			want: []string{"ordinanza -/0"},
		},
		"numbered with year": {
			text: "Con Ordinanza n. 38/2026 il Sindaco dispone la chiusura dei parchi. È stata emessa apposita ordinanza.",
			want: []string{"ordinanza 38/2026"},
		},
		"numbered with date": {
			text: "Si comunica che con ordinanza sindacale n° 0034 del 20/08/2026 sono chiusi i cimiteri.",
			want: []string{"ordinanza 34/2026 2026-08-20"},
		},
		"ordinance and decree": {
			text: "Emessa ordinanza di chiusura dei parchi. Con decreto n. 9/2026 è attivato il centro operativo.",
			want: []string{"decreto 9/2026", "ordinanza -/0"},
		},
		"footer only": {
			text: "Allerta meteo gialla per temporali." + footer,
		},
		"negated": {
			text: "Al momento non è stata emessa alcuna ordinanza di chiusura.",
		},
		"hypothetical": {
			text: "Qualora necessario potrà essere emessa apposita ordinanza di chiusura.",
		},
		"citations inside an attached act": {
			text:        "È stata emessa apposita ordinanza per la chiusura dei parchi.",
			attachments: []string{"Visto il Piano comunale adottato con deliberazione di Giunta comunale n. 129 del 24/10/2023; ORDINA la chiusura."},
			want:        []string{"ordinanza -/0"},
		},
		"bounded": {
			text: "Ordinanza n. 1/2026. Ordinanza n. 2/2026. Ordinanza n. 3/2026. Ordinanza n. 4/2026.",
			want: []string{"ordinanza 1/2026", "ordinanza 2/2026", "ordinanza 3/2026"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := []string{}
			for i, r := range Detect(notice(test.text, test.attachments...)) {
				if r.Ordinal != i+1 || r.Passage == "" || r.ResourceURL != "https://comune.example.it/avviso" || !strings.Contains(strings.ToLower(r.Passage), r.Kind[:5]) {
					t.Fatalf("reference %+v", r)
				}
				label := r.Kind + " -/0"
				if r.Number > 0 {
					label = r.Kind + " " + strconv.Itoa(r.Number) + "/" + strconv.Itoa(r.Year)
				}
				if r.Date != nil {
					label += " " + r.Date.Format(time.DateOnly)
				}
				got = append(got, label)
			}
			if strings.Join(got, ";") != strings.Join(test.want, ";") {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}
