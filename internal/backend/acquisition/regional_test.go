package acquisition

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func TestCriticalityUsesOnlyExplicitRowsForLevelsZonesAndValidity(t *testing.T) {
	body := []byte(`<html><body>
<h1>BOLLETTINO DI VALUTAZIONE DELLE CRITICITÀ</h1>
<p>Emissione di <b>Domenica, 12 Aprile 2026</b>, ore <b>12.48</b></p>
<p>Legenda: VERDE GIALLO ARANCIONE ROSSO; tutte le zone A1, A2, A3, A4, A5, A6, B, C, E1, E2, E3, F1, F2, I, L, M, O1, O2, O3, R1, R2, S1, S2, S3, T, V.</p>
<table><tr><th>ZONE DI ALLERTA</th><th>RISCHIO</th><th>TEMPI</th><th>CRITICITÀ</th></tr>
<tr><td>A4, A5, S1</td><td>RISCHIO IDRAULICO RETICOLO PRINCIPALE</td><td>dalle ore 12.00 alle ore 23.59 di Lunedì, 13 Aprile 2026</td><td>GIALLO</td></tr>
<tr><td>E2, E3</td><td>RISCHIO VENTO</td><td>dalle ore 08.00 alle ore 18.00 di Lunedì, 13 Aprile 2026</td><td>ARANCIONE</td></tr></table>
</body></html>`)
	observation, err := observeRegionalHTML("criticality", body, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(observation.AffectedZones, []string{"A4", "A5", "E2", "E3", "S1"}) || !slices.Equal(observation.ExplicitLevels, []string{"arancione", "giallo"}) || !slices.Equal(observation.RisksOrPhenomena, []string{"RISCHIO IDRAULICO RETICOLO PRINCIPALE", "RISCHIO VENTO"}) || !slices.Equal(observation.ValidityExpressions, []string{"dalle ore 08.00 alle ore 18.00 di Lunedì, 13 Aprile 2026", "dalle ore 12.00 alle ore 23.59 di Lunedì, 13 Aprile 2026"}) {
		t.Fatalf("legend or surrounding prose leaked into explicit facts: %#v", observation)
	}
}
func TestMonitoringNeverCreatesAnAlertLevel(t *testing.T) {
	body := []byte(`<html><body><h1>BOLLETTINO DI MONITORAGGIO E AGGIORNAMENTO EVENTO</h1>
<p>Emissione di Sabato, 15 Marzo 2025, ore 18.00</p>
<p>L'evento collegato all'allerta arancione evolve; prossimo aggiornamento alle 21.00.</p>
<table><tr><th>ZONE</th><th>RISCHIO</th><th>TEMPI</th><th>CRITICITÀ</th></tr>
<tr><td>A4</td><td>idrogeologico</td><td>in corso</td><td>ROSSO</td></tr></table></body></html>`)
	observation, err := observeRegionalHTML("monitoring", body, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(observation.ExplicitLevels) != 0 || len(observation.AffectedZones) != 0 || !strings.HasPrefix(observation.LevelSource, "none:") {
		t.Fatalf("monitoring was turned into a warning level: %#v", observation)
	}
}

func TestRegionalIssuanceInstantUsesEditionAndRome(t *testing.T) {
	for _, expression := range []string{"Mercoledí, 07 Ottobre 2026, ore 13.10", "07/10/2026, ore 13.10"} {
		got := regionalIssuanceInstant(expression)
		if got == nil || got.Format(time.RFC3339) != "2026-10-07T11:10:00Z" {
			t.Fatal(expression, got)
		}
	}
	for _, expression := range []string{"", "NESSUN AVVISO IN CORSO", "31/02/2026, ore 13.10", "07/10/2026, ore 25.00"} {
		if got := regionalIssuanceInstant(expression); got != nil {
			t.Fatal("invented publication instant", expression, got)
		}
	}
}
func TestRegionalPreviewRetainsHTMLPrintAndBoundedGraphics(t *testing.T) {
	const pageURL = "https://www.cfr.toscana.it/index.php?IDS=2&IDSS=71"
	const printURL = "https://www.cfr.toscana.it/bollettini/pdf_14.php?print=true"
	const mapURL = "https://www.cfr.toscana.it/tmp_cfr/map.png"
	const symbolURL = "https://www.cfr.toscana.it/supports/images/icon_lighting.png"
	body := []byte(`<html><body><h1>Bollettino di Vigilanza Meteorologica Regionale</h1>
<p>Emissione di Mercoledí, 16 Settembre 2026, ore 11.51</p>
<img src="/tmp_cfr/map.png"><img src="supports/images/icon_lighting.png"><img src="https://tracker.example/pixel.png">
<table><tr><th></th><th>Mercoledí, 16 Settembre 2026</th><th>Giovedì, 17 Settembre 2026</th></tr>
<tr><td>Pioggia</td><td></td><td>A4, A5</td></tr><tr><td>Temporali</td><td></td><td>A4</td></tr>
<tr><td>Vento</td><td></td><td></td></tr><tr><td>Mare</td><td></td><td></td></tr>
<tr><td>Neve</td><td></td><td></td></tr><tr><td>Ghiaccio</td><td></td><td></td></tr></table></body></html>`)
	crawler := &fixtureCrawler{pages: map[string]Page{
		pageURL: {URL: pageURL, HTML: body, MediaType: "text/html", StatusCode: 200},
	}}
	resources := &fixtureCrawler{pages: map[string]Page{
		printURL:  {URL: printURL, HTML: []byte("PDF with embedded maps"), MediaType: "application/pdf", StatusCode: 200},
		mapURL:    {URL: mapURL, HTML: []byte("map"), MediaType: "image/png", StatusCode: 200},
		symbolURL: {URL: symbolURL, HTML: []byte("symbol"), MediaType: "image/png", StatusCode: 200},
	}}
	evidence := registry.Evidence{URL: "https://www.cfr.toscana.it/", Locator: "operator decision: public CFR publication is sufficient for internal retention", ObservedAt: time.Date(2026, 9, 16, 19, 0, 0, 0, time.UTC)}
	reg := &fakeRegistry{version: registry.Version{SourceID: "cfr-vigilance", Revision: 1, Configuration: registry.Configuration{
		URL: pageURL, Sections: []string{pageURL}, AccessMethod: "crawl4ai-html-pdf", Attribution: "Centro Funzionale Regionale Toscana",
		Policy:          registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true, Conditions: "retained copies remain internal"},
		RegionalProduct: &registry.RegionalProductContract{Kind: "vigilance", ContentMarkers: []string{"Vigilanza Meteorologica Regionale"}, PrintURL: printURL, ResourcePathPrefixes: []string{"/tmp_cfr/", "/supports/images/"}, MinimumGraphicalAssets: 2, PrintContainsGraphics: true},
	}}}
	retained := &fakeRetention{}
	engine := Engine{Registry: reg, Retained: retained, Crawler: crawler, Resources: resources, Now: func() time.Time { return evidence.ObservedAt }}
	report, err := engine.Preview(context.Background(), "cfr-vigilance", 1, "operator")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Documents) != 1 || len(retained.items) != 1 || len(retained.items[0].Resources) != 4 {
		t.Fatalf("regional evidence graph incomplete: report=%#v retained=%#v", report, retained.items)
	}
	if !slices.Equal(crawler.calls, []string{pageURL}) || len(resources.calls) != 3 {
		t.Fatalf("HTML crawler and original-resource fetcher were not separated: html=%#v resources=%#v", crawler.calls, resources.calls)
	}
	for _, call := range append(slices.Clone(crawler.calls), resources.calls...) {
		if strings.Contains(call, "tracker.example") {
			t.Fatalf("unconfigured graphical origin acquired: %s", call)
		}
	}
	var metadata RegionalObservation
	if err = json.Unmarshal(retained.items[0].Metadata, &metadata); err != nil || metadata.Product != "vigilance" || !slices.Equal(metadata.AffectedZones, []string{"A4", "A5"}) || len(metadata.ExplicitLevels) != 0 {
		t.Fatalf("regional metadata mismatch: %#v, %v", metadata, err)
	}
}
func TestRegionalPreviewRejectsMissingRequiredGraphics(t *testing.T) {
	const pageURL = "https://www.cfr.toscana.it/vigilance"
	reg := &fakeRegistry{version: registry.Version{Configuration: registry.Configuration{
		URL: pageURL, AccessMethod: "crawl4ai-html-pdf", Sections: []string{pageURL}, Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true},
		RegionalProduct: &registry.RegionalProductContract{Kind: "vigilance", ContentMarkers: []string{"Vigilanza"}, ResourcePathPrefixes: []string{"/maps/"}, MinimumGraphicalAssets: 1},
	}}}
	crawler := &fixtureCrawler{pages: map[string]Page{pageURL: {URL: pageURL, HTML: []byte("<h1>Vigilanza</h1><p>Emissione di Oggi, ore 11.30</p>"), StatusCode: 200}}}
	engine := Engine{Registry: reg, Retained: &fakeRetention{}, Crawler: crawler, Resources: crawler}
	if _, err := engine.Preview(context.Background(), "cfr-vigilance", 1, "operator"); err == nil || !strings.Contains(err.Error(), "graphical resources") {
		t.Fatalf("missing graphical dependencies accepted: %v", err)
	}
}
func TestRegionalCheckTracksStableURLRevisions(t *testing.T) {
	const pageURL = "https://www.cfr.toscana.it/monitoring"
	crawler := &fixtureCrawler{pages: map[string]Page{pageURL: {
		URL: pageURL, HTML: []byte("<h1>Monitoraggio Evento</h1><p>Emissione di Giovedì, 17 Settembre 2026, ore 12.00</p>"), StatusCode: 200,
	}}}
	reg := &fakeRegistry{version: registry.Version{Configuration: registry.Configuration{
		URL: pageURL, AccessMethod: "crawl4ai-html-pdf", Sections: []string{pageURL}, Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true},
		RegionalProduct: &registry.RegionalProductContract{Kind: "monitoring", ContentMarkers: []string{"Monitoraggio Evento"}},
	}}}
	retained := &stableRetention{}
	tracker := &fakeTracker{}
	checkedAt := time.Date(2026, 9, 17, 12, 5, 0, 0, time.UTC)
	engine := Engine{Registry: reg, Retained: retained, Crawler: crawler, Resources: crawler, Tracking: tracker, Now: func() time.Time { return checkedAt }}
	first := engine.Check(context.Background(), "cfr-monitoring", 1, checkedAt.Add(-time.Minute))
	if !first.Complete || first.Revisions != 0 {
		t.Fatalf("first regional version mislabeled: %#v", first)
	}
	second := engine.Check(context.Background(), "cfr-monitoring", 1, checkedAt)
	if !second.Complete || second.Revisions != 0 {
		t.Fatalf("unchanged regional version mislabeled: %#v", second)
	}
	crawler.pages[pageURL] = Page{URL: pageURL, HTML: []byte("<h1>Monitoraggio Evento</h1><p>Emissione di Giovedì, 17 Settembre 2026, ore 15.00</p>"), StatusCode: 200}
	third := engine.Check(context.Background(), "cfr-monitoring", 1, checkedAt.Add(time.Minute))
	if !third.Complete || third.Revisions != 1 {
		t.Fatalf("changed stable regional URL not tracked: %#v", third)
	}
}
