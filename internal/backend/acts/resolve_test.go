package acts

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

type fakeAlbo struct {
	records []acquisition.AlboRecord
	dates   map[string]time.Time
	err     error
	windows [][2]time.Time
	details int
}

func (f *fakeAlbo) Window(_ context.Context, from, to time.Time, bound int) ([]acquisition.AlboRecord, error) {
	f.windows = append(f.windows, [2]time.Time{from, to})
	if f.err != nil {
		return nil, f.err
	}
	var result []acquisition.AlboRecord
	for _, r := range f.records {
		if !r.PublishedFrom.Before(from) && !r.PublishedFrom.After(to) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (f *fakeAlbo) Detail(_ context.Context, id string) (acquisition.AlboDetail, []byte, error) {
	f.details++
	for _, r := range f.records {
		if r.RecordID == id {
			r.ActDate = f.dates[id]
			return acquisition.AlboDetail{AlboRecord: r}, []byte("<html>" + id + "</html>"), nil
		}
	}
	return acquisition.AlboDetail{}, nil, fmt.Errorf("%w: missing", acquisition.ErrAlboUnavailable)
}

func day(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }

const object38 = "EMERGENZA DI PROTEZIONE CIVILE - ORDINANZA CONTINGIBILE E URGENTE DI CHIUSURE TEMPORANEE DEI PARCHI, GIARDINI PUBBLICI, AREE DI SGAMBAMENTO, SOTTOPASSI, CIMITERI, DAL 10 SETTEMBRE 2026"

func septemberAlbo() *fakeAlbo {
	return &fakeAlbo{
		records: []acquisition.AlboRecord{
			{RecordID: "660", Category: "ORDINANZE", Subcategory: "ORDINANZA", ActYear: 2026, ActNumber: 38, RegistryYear: 2026, RegistryNumber: 660, Object: object38, PublishedFrom: day(9, 9)},
			{RecordID: "658", Category: "DECRETI", Subcategory: "DECRETO", ActYear: 2026, ActNumber: 9, RegistryYear: 2026, RegistryNumber: 658, Object: object38, PublishedFrom: day(9, 9)},
			{RecordID: "655", Category: "DELIBERE", ActYear: 2026, ActNumber: 101, RegistryYear: 2026, RegistryNumber: 655, Object: "APPROVAZIONE PROGETTO DI MANUTENZIONE", PublishedFrom: day(9, 7)},
			{RecordID: "640", Category: "ORDINANZE", ActYear: 2026, ActNumber: 36, RegistryYear: 2026, RegistryNumber: 640, Object: "REGOLAMENTAZIONE DELLA CIRCOLAZIONE PER LAVORI", PublishedFrom: day(9, 3)},
		},
		dates: map[string]time.Time{"660": day(9, 9), "658": day(9, 9), "640": day(9, 3)},
	}
}

var septemberMeasures = []NoticeMeasure{{Kind: "closure", Subject: "cimiteri comunali"}, {Kind: "closure", Subject: "area sgambamento cani"}, {Kind: "suspension", Subject: "attività educative di tutte le scuole"}}

func contractFixture() *registry.JCityGovAlboContract {
	return &registry.JCityGovAlboContract{Categories: map[string]string{"ordinanza": "ORDINANZE", "decreto": "DECRETI"}, PageBound: 12}
}

func TestResolveOutcomes(t *testing.T) {
	ctx := context.Background()
	unnumbered := Reference{Kind: "ordinanza", Passage: "mediante apposita ordinanza"}
	september := Notice{PublicationDate: day(9, 9), Content: notice("Emessa ordinanza di chiusura."), Measures: septemberMeasures}

	t.Run("unique ordinance beside a decree with the same object", func(t *testing.T) {
		albo := septemberAlbo()
		r := Resolve(ctx, unnumbered, september, contractFixture(), albo)
		if r.Outcome != "unique_candidate" || r.AlboRecordID != "660" || r.Detail == nil || r.Detail.ActNumber != 38 || len(r.Candidates) != 1 || albo.details != 1 {
			t.Fatalf("%+v details=%d", r, albo.details)
		}
		if w := albo.windows[0]; !w[0].Equal(day(9, 2)) || !w[1].Equal(day(9, 19)) {
			t.Fatalf("window %v", w)
		}
	})
	t.Run("two matching ordinances", func(t *testing.T) {
		albo := septemberAlbo()
		albo.records = append(albo.records, acquisition.AlboRecord{RecordID: "661", Category: "ORDINANZE", ActYear: 2026, ActNumber: 39, Object: "CHIUSURA DEI CIMITERI PER LAVORI", PublishedFrom: day(9, 9)})
		albo.dates["661"] = day(9, 8)
		r := Resolve(ctx, unnumbered, september, contractFixture(), albo)
		if r.Outcome != "ambiguous" || r.AlboRecordID != "" || len(r.Candidates) != 2 || r.Reason == "" {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("act dated outside the window", func(t *testing.T) {
		albo := septemberAlbo()
		albo.dates["660"] = day(8, 20)
		if r := Resolve(ctx, unnumbered, september, contractFixture(), albo); r.Outcome != "not_found" {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("no disposition term in any object", func(t *testing.T) {
		notice := september
		notice.Measures = []NoticeMeasure{{Kind: "observation", Subject: "rami caduti"}}
		if r := Resolve(ctx, unnumbered, notice, contractFixture(), septemberAlbo()); r.Outcome != "not_found" {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("numbered", func(t *testing.T) {
		albo := septemberAlbo()
		r := Resolve(ctx, Reference{Kind: "ordinanza", Number: 38, Year: 2026}, Notice{PublicationDate: day(9, 10)}, contractFixture(), albo)
		if r.Outcome != "numbered" || r.AlboRecordID != "660" {
			t.Fatalf("%+v", r)
		}
		if w := albo.windows[0]; !w[0].Equal(day(8, 11)) || !w[1].Equal(day(9, 10)) {
			t.Fatalf("numbered window %v", w)
		}
	})
	t.Run("numbered with date", func(t *testing.T) {
		date := day(9, 9)
		albo := septemberAlbo()
		if r := Resolve(ctx, Reference{Kind: "decreto", Number: 9, Year: 2026, Date: &date}, Notice{}, contractFixture(), albo); r.Outcome != "numbered" || r.AlboRecordID != "658" {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("attached act", func(t *testing.T) {
		albo := septemberAlbo()
		withAct := Notice{PublicationDate: day(8, 20), Measures: septemberMeasures, Content: notice("È stata emessa apposita ordinanza.", "COMUNE DI CALCINAIA Provincia di Pisa ORDINANZA N. 34 / 2026 Oggetto: EMERGENZA DI PROTEZIONE CIVILE")}
		r := Resolve(ctx, unnumbered, withAct, contractFixture(), albo)
		if r.Outcome != "attached" || r.AttachedResource != "https://comune.example.it/allegato.pdf" || len(albo.windows) != 0 {
			t.Fatalf("%+v windows=%d", r, len(albo.windows))
		}
		if r = Resolve(ctx, Reference{Kind: "ordinanza", Number: 38, Year: 2026}, withAct, contractFixture(), albo); r.Outcome == "attached" {
			t.Fatalf("attachment of another number accepted: %+v", r)
		}
	})
	t.Run("no albo", func(t *testing.T) {
		if r := Resolve(ctx, unnumbered, september, nil, nil); r.Outcome != "not_found" || r.Reason != "no_configured_albo" {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("albo unreachable", func(t *testing.T) {
		albo := septemberAlbo()
		albo.err = fmt.Errorf("%w: HTTP 503", acquisition.ErrAlboUnavailable)
		r := Resolve(ctx, unnumbered, september, contractFixture(), albo)
		if r.Outcome != "unavailable" || r.Reason == "" {
			t.Fatalf("%+v", r)
		}
		albo.err = fmt.Errorf("%w", acquisition.ErrAlboBound)
		if r = Resolve(ctx, unnumbered, september, contractFixture(), albo); r.Outcome != "unavailable" || !errors.Is(albo.err, acquisition.ErrAlboBound) {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("publication date unknown", func(t *testing.T) {
		notice := september
		notice.PublicationDate = time.Time{}
		if r := Resolve(ctx, unnumbered, notice, contractFixture(), septemberAlbo()); r.Outcome != "not_found" {
			t.Fatalf("%+v", r)
		}
	})
}

func TestDispositionTerms(t *testing.T) {
	terms := DispositionTerms(septemberMeasures)
	for _, want := range []string{"chiusu", "cimite", "sgamba", "sospen", "scuole"} {
		found := false
		for _, term := range terms {
			found = found || term == want
		}
		if !found {
			t.Fatalf("missing %q in %v", want, terms)
		}
	}
	if objectMatches("APPROVAZIONE PROGETTO DI MANUTENZIONE", terms) || !objectMatches(object38, terms) {
		t.Fatal("object matching")
	}
}
