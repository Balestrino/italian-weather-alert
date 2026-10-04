package registry

import (
	"encoding/json"
	"testing"
	"time"
)

func cittadinoConfigurationFixture() Configuration {
	p := &CittadinoInformatoContract{MunicipalityISTAT: "050004", MunicipalitySlug: "calcinaia", Publisher: "comune_calcinaia", Updates: true, Risks: true, PageSize: 2}
	e := Evidence{URL: "https://comune.calcinaia.pi.it/", Locator: "Synthetic operator selection for multi-source verification; no license assertion", ObservedAt: time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)}
	c := Configuration{CittadinoInformato: p, URL: p.BaseURL(), Sections: []string{p.BaseURL(), p.BaseURL() + "aggiornamenti/"}, AccessMethod: CittadinoInformatoAccess, Attribution: "Comune di Calcinaia, platform Cittadino Informato",
		Referral: &Referral{Evidence: e, Destination: p.BaseURL(), Context: "synthetic municipal referral", Sections: []string{p.BaseURL(), p.BaseURL() + "aggiornamenti/"}, ProductID: "municipal", Territory: "050004"},
		Policy:   Policy{Evidence: &e, CollectionPermitted: true, RetentionPermitted: true}, Discovery: Discovery{MaxPagesPerSection: 3, MaxDocuments: 10}}
	c.defaults()
	return c
}

func TestCittadinoInformatoExplicitScope(t *testing.T) {
	c := cittadinoConfigurationFixture()
	if !c.valid() {
		t.Fatal("explicit multi-source operator choice rejected")
	}
	for name, change := range map[string]func(*Configuration){
		"no opt in":               func(c *Configuration) { c.CittadinoInformato = nil },
		"wrong access":            func(c *Configuration) { c.AccessMethod = "crawl4ai" },
		"unselected municipality": func(c *Configuration) { c.CittadinoInformato.MunicipalitySlug = "cascina" },
		"referral territory":      func(c *Configuration) { c.Referral.Territory = "050008" },
		"extra section":           func(c *Configuration) { c.Sections = append(c.Sections, "https://cittadinoinformato.it/cascina/") },
		"public copies":           func(c *Configuration) { c.Policy.PublicationPermitted = true; c.Policy.CopiesPermitted = true },
		"unbounded pages":         func(c *Configuration) { c.Discovery.MaxPagesPerSection = 0 },
		"other municipality PDF":  func(c *Configuration) { c.CittadinoInformato.AttachmentPaths = []string{"/cascina/"} },
		"all shared uploads":      func(c *Configuration) { c.CittadinoInformato.AttachmentPaths = []string{"/app/uploads/"} },
		"bad filter date":         func(c *Configuration) { c.CittadinoInformato.UpdatesSince = "03/10/2026" },
	} {
		t.Run(name, func(t *testing.T) {
			raw, _ := json.Marshal(c)
			var changed Configuration
			json.Unmarshal(raw, &changed)
			change(&changed)
			if changed.valid() {
				t.Fatal("unscoped contract accepted")
			}
		})
	}
}

func TestCittadinoInformatoResourceOwnership(t *testing.T) {
	c := cittadinoConfigurationFixture()
	p := c.CittadinoInformato
	p.AttachmentPaths = []string{"/app/uploads/2026/10/"}
	for raw, allowed := range map[string]bool{
		p.APIBase() + "aggiornamenti/10?comune=calcinaia":                                                     true,
		p.APIBase() + "aggiornamenti/10?comune=cascina":                                                       false,
		p.APIBase() + "rischi/oggi":                                                                           true,
		p.APIBase() + "rischi/oggi?comune=cascina":                                                            false,
		p.APIBase() + "aggiornamenti?comune=calcinaia&ente=comune_calcinaia&page=1&per_page=2":                true,
		p.APIBase() + "aggiornamenti?comune=calcinaia&comune=cascina&ente=comune_calcinaia&page=1&per_page=2": false,
		"https://cittadinoinformato.it/cascina/wp-json/cittadino/v2/rischi/oggi":                              false,
	} {
		if p.AllowsOriginal(raw) != allowed {
			t.Fatalf("original scope: %s", raw)
		}
	}
	for raw, allowed := range map[string]bool{
		"https://cittadinoinformato.it/app/uploads/2026/10/ordinanza.pdf": true,
		"https://cittadinoinformato.it/app/uploads/2026/09/ordinanza.pdf": false,
		"https://cittadinoinformato.it/app/uploads/2026/10/../other.pdf":  false,
		"https://other.example/app/uploads/2026/10/ordinanza.pdf":         false,
		"https://cittadinoinformato.it/cascina/ordinanza.pdf":             false,
	} {
		if p.AllowsAttachment(raw) != allowed {
			t.Fatalf("attachment scope: %s", raw)
		}
	}
}

func TestCittadinoInformatoExternalDependencies(t *testing.T) {
	c := cittadinoConfigurationFixture()
	e := *c.Policy.Evidence
	c.CittadinoInformato.ExternalAttachments = []AttachmentScope{{Origin: "https://municipal.example", PathPrefix: "/sites/default/files/", Referral: e, Policy: Policy{Evidence: &e, CollectionPermitted: true, RetentionPermitted: true, Conditions: "Synthetic reviewed linked PDFs; link-only"}}}
	if !c.valid() {
		t.Fatal("reviewed external dependency rejected")
	}
	for target, want := range map[string]bool{
		"https://municipal.example/sites/default/files/2026-10/atto%20sintetico.pdf": true,
		"https://municipal.example/sites/default/files/2026-10/atto%C3%A0.pdf":       true,
		"https://municipal.example/sites/default/files/../other/atto.pdf":            false,
		"https://municipal.example/sites/default/files/%2e%2e/atto.pdf":              false,
		"https://municipal.example/sites/default/files/x%2f..%2fatto.pdf":            false,
		"https://municipal.example/sites/default/files/x%5catto.pdf":                 false,
		"https://municipal.example/sites/default/files/atto.pdf?foreign=1":           false,
		"https://municipal.example/sites/default/files/atto.html":                    false,
		"https://municipal.example/other/atto.pdf":                                   false,
		"https://other.example/sites/default/files/atto.pdf":                         false,
		"http://municipal.example/sites/default/files/atto.pdf":                      false,
	} {
		if c.CittadinoInformato.AllowsAttachment(target) != want {
			t.Fatalf("dependency boundary %s", target)
		}
	}
	c.CittadinoInformato.ExternalAttachments[0].Policy.CopiesPermitted = true
	if c.valid() {
		t.Fatal("external platform copies accepted")
	}
}
