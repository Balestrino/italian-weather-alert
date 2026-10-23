package acts

import (
	"errors"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
)

// Synthetic ordinance pages modelled on two reviewed municipal emergency
// ordinances: recitals repeat the dispositions before ORDINA.
const ordinance38 = `COMUNE DI ESEMPIO ORDINANZA N. 38 / 2026 Oggetto: EMERGENZA DI PROTEZIONE CIVILE IL SINDACO Visto il D. Lgs. n. 267 del 18/8/2000;
Ritenuto che debbano essere assunte le sotto elencate misure, disponendo: • la chiusura al pubblico: ◦ dell'area di sgambamento cani; ◦ dei cimiteri comunali;`
const ordinance38Page2 = `• il divieto di qualunque attività all'aperto: ◦ nei giardini/parchi pubblici/impianti sportivi; ◦ nelle aree verdi; • la sospensione delle attività didattiche educative di tutte le scuole di ogni ordine e grado e dei servizi educativo-scolastici pubblici e privati;
ORDINA in via contingibile e urgente, a partire dalle ore 00.00 del giorno 10 settembre 2026, fino al perdurare dell'emergenza: • la chiusura al pubblico: ◦ dell'area di sgambamento cani; ◦ dei cimiteri comunali; • il divieto di qualunque attività all'aperto: ◦ nei giardini/parchi pubblici/impianti sportivi; ◦ nelle aree verdi; ◦ negli orti sociali; • la sospensione delle attività didattiche educative di tutte le scuole di ogni ordine e grado e dei servizi educativo-scolastici pubblici e privati;
AVVERTE che si riserva di valutare ulteriori misure; DISPONE l'attivazione delle seguenti funzioni di supporto: volontariato;`

const ordinance34 = `ORDINANZA N. 34 / 2026 ORDINA in via contingibile e urgente, a partire dalle ore 18.00 del giorno 20 agosto 2026, fino al perdurare dell'emergenza: - la chiusura al pubblico: ◦ dell'area di sgambamento cani; ◦ dei cimiteri comunali; - il divieto di qualunque attività all'aperto: ◦ nei giardini/parchi pubblici; - la sospensione delle attività didattiche educative di tutte le scuole; AVVERTE che si riserva`

func pages(texts ...string) []classification.ContentSection {
	var sections []classification.ContentSection
	for i, text := range texts {
		sections = append(sections, classification.ContentSection{ResourceURL: "https://albo.example.it/ordinanza.pdf", Role: "attachment", Page: i + 1, Text: text})
	}
	return sections
}

func TestParseClauseConditionalOrdinance(t *testing.T) {
	clause, err := ParseClause(pages(ordinance38, ordinance38Page2))
	if err != nil {
		t.Fatal(err)
	}
	rome, _ := time.LoadLocation("Europe/Rome")
	if !clause.Start.Equal(time.Date(2026, 9, 10, 0, 0, 0, 0, rome)) || clause.End != nil || clause.Condition != "fino al perdurare dell'emergenza" || clause.Page != 2 {
		t.Fatalf("clause %+v", clause)
	}
	if len(clause.Dispositions) != 3 || clause.Dispositions[0].Kind != "closure" || clause.Dispositions[1].Kind != "prohibition" || clause.Dispositions[2].Kind != "suspension" {
		t.Fatalf("dispositions %+v", clause.Dispositions)
	}
	for _, test := range []struct {
		kind, subject string
		want          int
	}{
		{"closure", "cimiteri comunali (Calcinaia e Fornacette)", 1},
		{"closure", "area sgambamento cani", 1},
		{"prohibition", "attività all'aperto", 2},
		{"suspension", "attività educative di tutte le scuole di ogni ordine e grado", 3},
		{"suspension", "servizi educativo scolastici pubblici e privati", 3},
		{"closure", "sottopasso di via Maremmana", 0},
		{"activation", "funzioni di supporto", 0},
		{"prohibition", "cimiteri comunali", 0},
	} {
		got := clause.Match(test.kind, test.subject)
		if (test.want == 0) != (got == nil) || got != nil && got.Ordinal != test.want {
			t.Errorf("Match(%s, %q) = %+v, want %d", test.kind, test.subject, got, test.want)
		}
	}
	other, err := ParseClause(pages(ordinance34))
	if err != nil || !other.Start.Equal(time.Date(2026, 8, 20, 18, 0, 0, 0, rome)) || len(other.Dispositions) != 3 || other.Match("closure", "Cimiteri") == nil {
		t.Fatalf("ordinance 34: %+v %v", other, err)
	}
}

func TestParseClauseInterval(t *testing.T) {
	clause, err := ParseClause(pages("ORDINA a partire dalle ore 08:00 del giorno 3 ottobre 2026, fino alle ore 18:00 del giorno 4 ottobre 2026: la chiusura del sottopasso di via Roma; AVVERTE"))
	rome, _ := time.LoadLocation("Europe/Rome")
	if err != nil || clause.End == nil || !clause.End.Equal(time.Date(2026, 10, 4, 18, 0, 0, 0, rome)) || clause.Condition != "" || clause.Match("closure", "sottopasso") == nil {
		t.Fatalf("interval clause %+v %v", clause, err)
	}
}

func TestParseClauseUnrecognized(t *testing.T) {
	for name, text := range map[string]string{
		"no ORDINA":          "Si dispone la chiusura dei cimiteri dalle ore 00.00 del 10 settembre.",
		"recitals only":      "Ritenuto che debbano essere assunte misure, disponendo: • la chiusura al pubblico dei cimiteri;",
		"open-ended wording": "ORDINA a partire dalle ore 00.00 del giorno 10 settembre 2026, fino a quando sarà opportuno: la chiusura dei cimiteri;",
		"end before start":   "ORDINA a partire dalle ore 18.00 del giorno 10 settembre 2026, fino alle ore 08.00 del giorno 9 settembre 2026: la chiusura dei cimiteri;",
		"impossible date":    "ORDINA a partire dalle ore 00.00 del giorno 31 settembre 2026, fino al perdurare dell'emergenza: la chiusura dei cimiteri;",
		"no dispositions":    "ORDINA a partire dalle ore 00.00 del giorno 10 settembre 2026, fino al perdurare dell'emergenza: quanto in premessa; AVVERTE",
	} {
		if _, err := ParseClause(pages(text)); !errors.Is(err, ErrClauseUnrecognized) {
			t.Errorf("%s: accepted %v", name, err)
		}
	}
}
