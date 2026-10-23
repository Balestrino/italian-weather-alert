package registry

import (
	"encoding/json"
	"testing"
	"time"
)

func alboConfigurationFixture() Configuration {
	p := &JCityGovAlboContract{MunicipalityISTAT: "050004", ServesSource: "calcinaia-municipal", Origin: "https://albo.example.it", ListPath: "/web/trasparenza/papca-ap/-/papca/igrid/1/2",
		DetailPathPrefix: "/web/trasparenza/papca-ap/-/papca/display/", Categories: map[string]string{"ordinanza": "ORDINANZE", "decreto": "DECRETI"}, PageBound: 6}
	e := Evidence{URL: "https://comune.example.it/", Locator: "Synthetic footer referral to the albo pretorio; no license assertion", ObservedAt: time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)}
	c := Configuration{JCityGovAlbo: p, URL: p.ListURL(), Sections: []string{p.ListURL()}, AccessMethod: JCityGovAlboAccess, Attribution: "Comune, albo pretorio",
		Referral: &Referral{Evidence: e, Destination: p.ListURL(), Context: "synthetic municipal referral", Sections: []string{p.ListURL()}, ProductID: "municipal", Territory: "050004"},
		Policy:   Policy{Evidence: &e, CollectionPermitted: true, RetentionPermitted: true, PublicationPermitted: true}}
	c.defaults()
	return c
}

func TestJCityGovAlboScope(t *testing.T) {
	c := alboConfigurationFixture()
	if !c.valid() {
		t.Fatal("reviewed albo scope rejected")
	}
	for name, change := range map[string]func(*Configuration){
		"no contract":        func(c *Configuration) { c.JCityGovAlbo = nil },
		"wrong access":       func(c *Configuration) { c.AccessMethod = "crawl4ai" },
		"public copies":      func(c *Configuration) { c.Policy.CopiesPermitted = true },
		"no retention":       func(c *Configuration) { c.Policy.RetentionPermitted = false },
		"http origin":        func(c *Configuration) { c.JCityGovAlbo.Origin = "http://albo.example.it" },
		"origin with path":   func(c *Configuration) { c.JCityGovAlbo.Origin = "https://albo.example.it/web" },
		"list outside url":   func(c *Configuration) { c.URL = "https://albo.example.it/other" },
		"extra section":      func(c *Configuration) { c.Sections = append(c.Sections, "https://albo.example.it/web/") },
		"detail traversal":   func(c *Configuration) { c.JCityGovAlbo.DetailPathPrefix = "/web/../display/" },
		"detail no slash":    func(c *Configuration) { c.JCityGovAlbo.DetailPathPrefix = "/web/display" },
		"unknown kind":       func(c *Configuration) { c.JCityGovAlbo.Categories["avviso"] = "AVVISI" },
		"no categories":      func(c *Configuration) { c.JCityGovAlbo.Categories = nil },
		"unbounded pages":    func(c *Configuration) { c.JCityGovAlbo.PageBound = 0 },
		"too many pages":     func(c *Configuration) { c.JCityGovAlbo.PageBound = 21 },
		"no served source":   func(c *Configuration) { c.JCityGovAlbo.ServesSource = "" },
		"referral territory": func(c *Configuration) { c.Referral.Territory = "050008" },
		"no referral":        func(c *Configuration) { c.Referral = nil },
		"attachments":        func(c *Configuration) { c.Attachments = &AttachmentPolicy{} },
	} {
		t.Run(name, func(t *testing.T) {
			raw, _ := json.Marshal(c)
			var changed Configuration
			if err := json.Unmarshal(raw, &changed); err != nil {
				t.Fatal(err)
			}
			change(&changed)
			if changed.valid() {
				t.Fatal("unscoped albo configuration accepted")
			}
		})
	}
}

func TestJCityGovAlboDetailOwnership(t *testing.T) {
	p := alboConfigurationFixture().JCityGovAlbo
	for raw, want := range map[string]bool{
		"https://albo.example.it/web/trasparenza/papca-ap/-/papca/display/2744237":                  true,
		"https://albo.example.it/web/trasparenza/papca-ap/-/papca/display/2744237?p_p_state=pop_up": true,
		"https://albo.example.it/web/trasparenza/papca-ap/-/papca/display/":                         false,
		"https://albo.example.it/web/trasparenza/papca-ap/-/papca/display/27a":                      false,
		"https://other.example.it/web/trasparenza/papca-ap/-/papca/display/1":                       false,
		"http://albo.example.it/web/trasparenza/papca-ap/-/papca/display/1":                         false,
	} {
		if got := p.AllowsDetail(raw); got != want {
			t.Errorf("AllowsDetail(%q)=%v", raw, got)
		}
	}
}
