package registry

import (
	"testing"
	"time"
)

func TestConfigurationEvidenceValidation(t *testing.T) {
	e := Evidence{URL: "https://official.example/reuse", Locator: "reuse conditions", ObservedAt: time.Now()}
	c := Configuration{URL: "https://official.example", Sections: []string{"https://official.example/news"}, AccessMethod: "html", Attribution: "official publisher"}
	c.defaults()
	if c.CheckSeconds != 600 || c.DelaySeconds != 1800 || c.BackoffBaseSeconds != 60 || c.BackoffMaxSeconds != 3600 || c.Discovery.BootstrapDays != 30 {
		t.Fatalf("unexpected scheduling defaults: %#v", c)
	}
	if !c.valid() {
		t.Fatal("pending configuration rejected")
	}
	c.Policy.PublicationPermitted = true
	if c.valid() {
		t.Fatal("permission without evidence accepted")
	}
	c.Policy.Evidence = &e
	if !c.valid() {
		t.Fatal("permission evidence rejected")
	}
	c.Policy.CopiesPermitted = true
	if c.valid() {
		t.Fatal("copy permission without retention accepted")
	}
	c.Policy.RetentionPermitted = true
	if !c.valid() {
		t.Fatal("complete copy policy rejected")
	}
	c.Sections = append(c.Sections, c.Sections[0])
	if c.valid() {
		t.Fatal("duplicate sections accepted")
	}
	c.Sections = c.Sections[:1]
	c.URL = "https://user:secret@official.example"
	if c.valid() {
		t.Fatal("credentials in URL accepted")
	}
	c.URL = "https://official.example"
	c.DelaySeconds = -1
	if c.valid() {
		t.Fatal("invalid interval accepted")
	}
}

func TestExpectedPublicationRequiresDocumentedCadence(t *testing.T) {
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	c := Configuration{URL: "https://official.example", Sections: []string{"https://official.example/news"}, AccessMethod: "html", Attribution: "official publisher"}
	c.defaults()
	c.ExpectedPublication = &ExpectedPublication{Anchor: now, IntervalSeconds: 86400, ToleranceSeconds: 3600}
	if c.valid() {
		t.Fatal("undocumented publication cadence accepted")
	}
	c.ExpectedPublication.Evidence = Evidence{URL: "https://official.example/publication-calendar", Locator: "daily bulletin cadence", ObservedAt: now}
	if !c.valid() {
		t.Fatal("documented publication cadence rejected")
	}
	c.ExpectedPublication.ToleranceSeconds = c.ExpectedPublication.IntervalSeconds + 1
	if c.valid() {
		t.Fatal("tolerance longer than publication interval accepted")
	}
}

func TestRegionalProductContractValidation(t *testing.T) {
	evidence := Evidence{URL: "https://official.example/publication-policy", Locator: "public CFR products", ObservedAt: time.Now()}
	c := Configuration{
		URL:          "https://official.example/vigilance",
		Sections:     []string{"https://official.example/vigilance"},
		AccessMethod: "html-and-pdf",
		Attribution:  "regional authority",
		Policy:       Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true},
		RegionalProduct: &RegionalProductContract{
			Kind:                   "vigilance",
			ContentMarkers:         []string{"Vigilanza Meteorologica"},
			PrintURL:               "https://official.example/vigilance.pdf",
			ResourcePathPrefixes:   []string{"/maps/", "/symbols/"},
			MinimumGraphicalAssets: 2,
			PrintContainsGraphics:  true,
		},
	}
	c.defaults()
	if !c.valid() {
		t.Fatal("valid regional product contract rejected")
	}
	c.RegionalProduct.Kind = "alert"
	if c.valid() {
		t.Fatal("unknown regional product accepted")
	}
	c.RegionalProduct.Kind = "vigilance"
	c.RegionalProduct.PrintURL = "https://other.example/vigilance.pdf"
	if c.valid() {
		t.Fatal("cross-origin regional print resource accepted")
	}
	c.RegionalProduct.Kind = "monitoring"
	c.RegionalProduct.PrintURL = ""
	if c.valid() {
		t.Fatal("embedded graphics asserted without a print document")
	}
	c.RegionalProduct.PrintContainsGraphics = false
	c.RegionalProduct.ResourcePathPrefixes = nil
	c.RegionalProduct.MinimumGraphicalAssets = 0
	if !c.valid() {
		t.Fatal("event-driven monitoring contract rejected")
	}
	c.ExpectedPublication = &ExpectedPublication{Evidence: evidence, Anchor: time.Now(), IntervalSeconds: 3600, ToleranceSeconds: 60}
	if c.valid() {
		t.Fatal("monitoring cadence could raise a false missing-publication alarm")
	}
	c.ExpectedPublication = nil
	c.RegionalProduct.ResourcePathPrefixes = []string{"https://unbounded.example/maps"}
	if c.valid() {
		t.Fatal("unbounded regional resource prefix accepted")
	}
}
