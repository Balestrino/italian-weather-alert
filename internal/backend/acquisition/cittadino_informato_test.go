package acquisition

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

type cittadinoFixtureCrawler struct {
	fixtureCrawler
	failures map[string]error
}

func (f *cittadinoFixtureCrawler) CrawlBounded(ctx context.Context, raw string, allowed func(string) bool) (Page, error) {
	if !allowed(raw) {
		return Page{}, ErrInvalidConfiguration
	}
	if err := f.failures[raw]; err != nil {
		f.calls = append(f.calls, raw)
		var failure *CrawlFailure
		if errors.As(err, &failure) {
			return Page{URL: raw, StatusCode: failure.StatusCode}, err
		}
		return Page{}, err
	}
	page, err := f.Crawl(ctx, raw)
	if err == nil && !allowed(page.URL) {
		return Page{}, ErrInvalidConfiguration
	}
	return page, err
}

func cittadinoFixture() (registry.Configuration, *cittadinoFixtureCrawler) {
	p := &registry.CittadinoInformatoContract{MunicipalityISTAT: "050004", MunicipalitySlug: "calcinaia", Publisher: "comune_calcinaia", Updates: true, Risks: true, PageSize: 1}
	e := registry.Evidence{URL: "https://comune.calcinaia.pi.it/", Locator: "Synthetic operator multi-source selection", ObservedAt: time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)}
	sections := []string{p.BaseURL() + "aggiornamenti/", p.BaseURL()}
	c := registry.Configuration{CittadinoInformato: p, URL: p.BaseURL(), Sections: sections, AccessMethod: registry.CittadinoInformatoAccess, Attribution: "Synthetic Comune / Cittadino Informato",
		Referral: &registry.Referral{Evidence: e, Destination: p.BaseURL(), Context: "synthetic referral", Sections: sections, ProductID: "municipal", Territory: "050004"},
		Policy:   registry.Policy{Evidence: &e, CollectionPermitted: true, RetentionPermitted: true}, Discovery: registry.Discovery{MaxPagesPerSection: 2, MaxDocuments: 10, BootstrapDays: 30}}
	f := &cittadinoFixtureCrawler{fixtureCrawler: fixtureCrawler{pages: map[string]Page{}}, failures: map[string]error{}}
	f.pages[p.APIBase()+"comune?nome=calcinaia"] = cittadinoJSON(p.APIBase()+"comune?nome=calcinaia", `{"id":109,"slug":"calcinaia"}`)
	for number := 1; number <= 2; number++ {
		q := url.Values{"comune": {"calcinaia"}, "ente": {"comune_calcinaia"}, "per_page": {"1"}, "page": {strconv.Itoa(number)}}
		raw := p.APIBase() + "aggiornamenti?" + q.Encode()
		n := cittadinoNotice{ID: int64(number), Publisher: p.Publisher, URL: p.BaseURL() + "2026/10/02/synthetic-" + strconv.Itoa(number) + "/", Title: "Avviso sintetico", Content: "<p>Provvedimento sintetico.</p>", Published: "02/10/2026", DisplayFrom: "03/10/2026", DisplayTo: "04/10/2026"}
		listing, _ := json.Marshal(map[string]any{"count": 1, "total_count": 2, "pages": 2, "page": number, "posts": []cittadinoNotice{n}})
		f.pages[raw] = cittadinoJSON(raw, string(listing))
		detail := cittadinoDetailURL(p, n.ID)
		data, _ := json.Marshal(n)
		f.pages[detail] = cittadinoJSON(detail, string(data))
	}
	for _, day := range []string{"oggi", "domani"} {
		raw := p.APIBase() + "rischi/" + day
		f.pages[raw] = cittadinoJSON(raw, `{"stati_allerta":[{"tipologia":"Vento","`+day+`":{"allerta":"giallo","data_bollettino":"02/10/2026","validita_cfr":null}}]}`)
	}
	return c, f
}

func cittadinoJSON(raw, body string) Page {
	return Page{URL: raw, StatusCode: 200, MediaType: "application/json", HTML: []byte(body)}
}

func TestCittadinoInformatoPreviewRetainsScopedEvidenceAndDateMeanings(t *testing.T) {
	cfg, f := cittadinoFixture()
	retained := &fakeRetention{}
	reg := &fakeRegistry{version: registry.Version{Configuration: cfg}}
	e := Engine{Registry: reg, Retained: retained, Crawler: f, Resources: f}
	report, err := e.Preview(context.Background(), "calcinaia-cittadino-informato", 1, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Sections) != 2 || len(report.Sections[0].Listings) != 2 || len(report.Documents) != 4 {
		t.Fatalf("unexpected report: %#v", report)
	}
	var notices int
	for _, a := range retained.items {
		if a.SourceID != "calcinaia-cittadino-informato" {
			t.Fatal("primary source identity used")
		}
		var meta map[string]any
		json.Unmarshal(a.Metadata, &meta)
		if meta["kind"] == "notice" {
			notices++
			if meta["source_publication_date"] != "2026-10-02" || meta["platform_display_start"] != "03/10/2026" || meta["platform_display_end"] != "04/10/2026" || meta["verification_state"] != "pending" {
				t.Fatalf("date meanings conflated: %s", a.Metadata)
			}
			if a.Resources[0].MediaType != "application/json" || !json.Valid(a.Resources[0].Bytes) {
				t.Fatal("original API bytes not retained")
			}
		}
	}
	if notices != 2 {
		t.Fatal("notices missing")
	}
}

func TestCittadinoInformatoBoundsAndFailures(t *testing.T) {
	for name, change := range map[string]func(*registry.Configuration, *cittadinoFixtureCrawler){
		"page bound":     func(c *registry.Configuration, f *cittadinoFixtureCrawler) { c.Discovery.MaxPagesPerSection = 1 },
		"document bound": func(c *registry.Configuration, f *cittadinoFixtureCrawler) { c.Discovery.MaxDocuments = 1 },
		"wrong municipality": func(c *registry.Configuration, f *cittadinoFixtureCrawler) {
			raw := c.CittadinoInformato.APIBase() + "comune?nome=calcinaia"
			f.pages[raw] = cittadinoJSON(raw, `{"id":110,"slug":"cascina"}`)
		},
		"outside detail": func(c *registry.Configuration, f *cittadinoFixtureCrawler) {
			raw := cittadinoDetailURL(c.CittadinoInformato, 1)
			p := f.pages[raw]
			p.HTML = []byte(strings.ReplaceAll(string(p.HTML), "/calcinaia/2026/", "/cascina/2026/"))
			f.pages[raw] = p
		},
		"duplicate listing": func(c *registry.Configuration, f *cittadinoFixtureCrawler) {
			for raw, p := range f.pages {
				if strings.Contains(raw, "page=2&") {
					p.HTML = []byte(strings.ReplaceAll(string(p.HTML), `"id":2`, `"id":1`))
					f.pages[raw] = p
				}
			}
		},
		"unrecognized response": func(c *registry.Configuration, f *cittadinoFixtureCrawler) {
			raw := c.CittadinoInformato.APIBase() + "comune?nome=calcinaia"
			f.pages[raw] = cittadinoJSON(raw, `{}`)
		},
		"unknown risk level": func(c *registry.Configuration, f *cittadinoFixtureCrawler) {
			raw := c.CittadinoInformato.APIBase() + "rischi/oggi"
			f.pages[raw] = cittadinoJSON(raw, `{"stati_allerta":[{"tipologia":"Vento","oggi":{"allerta":"sconosciuto","data_bollettino":null}}]}`)
		},
	} {
		t.Run(name, func(t *testing.T) {
			cfg, f := cittadinoFixture()
			change(&cfg, f)
			reg := &fakeRegistry{version: registry.Version{Configuration: cfg}}
			e := Engine{Registry: reg, Retained: &fakeRetention{}, Crawler: f, Resources: f}
			out := e.Check(context.Background(), "platform", 1, time.Now())
			if out.Complete || out.ErrorCode == "" {
				t.Fatalf("failure hidden: %#v", out)
			}
			if strings.Contains(strings.Join(f.calls, " "), "/cascina/") {
				t.Fatal("unselected municipality requested")
			}
		})
	}
}

func TestCittadinoInformatoPreservesRetryAfter(t *testing.T) {
	cfg, f := cittadinoFixture()
	retry := time.Now().Add(time.Hour)
	raw := cfg.CittadinoInformato.APIBase() + "comune?nome=calcinaia"
	f.failures[raw] = &CrawlFailure{Code: "source_rate_limited", Reachable: true, StatusCode: 429, RetryAfter: &retry}
	e := Engine{Registry: &fakeRegistry{version: registry.Version{Configuration: cfg}}, Retained: &fakeRetention{}, Crawler: f, Resources: f}
	out := e.Check(context.Background(), "platform", 1, time.Now())
	if out.Complete || out.ErrorCode != "source_rate_limited" || out.RetryAfter == nil || !out.RetryAfter.Equal(retry) {
		t.Fatalf("retry instruction lost: %#v", out)
	}
}

func TestCittadinoInformatoForbiddenDependencyIsRetainedAsMissing(t *testing.T) {
	cfg, f := cittadinoFixture()
	raw := cittadinoDetailURL(cfg.CittadinoInformato, 1)
	var n cittadinoNotice
	json.Unmarshal(f.pages[raw].HTML, &n)
	n.Content = `<p>Durata nell’atto <a href="https://unselected.example/ordinanza.pdf">PDF</a>.</p>`
	data, _ := json.Marshal(n)
	f.pages[raw] = cittadinoJSON(raw, string(data))
	store := &fakeRetention{}
	e := Engine{Registry: &fakeRegistry{version: registry.Version{Configuration: cfg}}, Retained: store, Crawler: f, Resources: f}
	_, err := e.Preview(context.Background(), "platform", 1, "fixture")
	if !errors.Is(err, ErrRequiredAttachment) {
		t.Fatalf("dependency silently ignored: %v", err)
	}
	last := store.items[len(store.items)-1]
	if len(last.Resources) != 2 || last.Resources[1].Missing != "forbidden" || len(last.Resources[1].Bytes) != 0 {
		t.Fatalf("missing dependency not retained: %#v", last.Resources)
	}
}
