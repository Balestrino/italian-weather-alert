package classification

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func TestArticleBodySelectionAndLayoutFallback(t *testing.T) {
	body := []byte(`<nav>Allerta meteo: notizia precedente</nav><article class="body"><h1 id="title">Notizia ordinaria</h1><div id="notice">Testo integrale della notizia.<p>Ultima frase: per rischio idraulico si dispone la chiusura del ponte.</p></div></article><footer>Footer e navigazione</footer>`)
	want := "Notizia ordinaria Testo integrale della notizia. Ultima frase: per rischio idraulico si dispone la chiusura del ponte."
	for _, selectors := range [][]string{{"article.body"}, {"#title", "#notice"}, {"#notice", "#title"}, {"article.body", "#notice"}} {
		if got := articleHTML(body, selectors); got != want {
			t.Fatalf("body order/coverage %v: %q", selectors, got)
		}
	}
	for _, fixture := range []struct {
		body      string
		selectors []string
	}{
		{string(body), []string{"#missing"}},
		{string(body), []string{"#notice", "#missing"}},
		{string(body) + `<div id="notice">Other notice</div>`, []string{"#notice"}},
		{string(body) + `<div id="empty"><script>hidden</script></div>`, []string{"#empty"}},
		{string(body), []string{"article div"}},
	} {
		if got := articleHTML([]byte(fixture.body), fixture.selectors); got != visibleHTML([]byte(fixture.body)) {
			t.Fatal("layout drift discarded page evidence", got)
		}
	}
}

func TestArticleScopePreservesAttachmentsAndSeparatesIdentity(t *testing.T) {
	url, pdf := "https://municipal.example/news/notice", "https://municipal.example/orders/order.pdf"
	policy := &registry.LocalProcessing{Version: registry.LocalProcessingVersion, HTMLBodySelectors: []string{"#notice"}}
	version := documents.Version{ID: 8, Complete: true, Resources: []documents.Reference{
		{URL: url, Role: "original", MediaType: "text/html", Required: true, LocalProcessing: policy},
		{URL: pdf, Role: "attachment", MediaType: "application/pdf", Required: true},
	}}
	docs := &fakeDocuments{version: version, bodies: map[string][]byte{url: []byte(`<nav>Weather navigation</nav><div id="notice">Article body.</div><footer>Old warning</footer>`)}}
	content, err := GatherContent(context.Background(), docs, fakeOCR{}, version)
	if err != nil || content.Complete {
		t.Fatal("missing attachment became complete", err)
	}
	// A textual attachment is always retained even though only the parent has selectors.
	version.Resources[1].MediaType = "text/plain"
	docs.bodies[pdf] = []byte("Operative attachment: chiusura per rischio idraulico.")
	content, err = GatherContent(context.Background(), docs, fakeOCR{}, version)
	if err != nil || !content.Complete || len(content.Sections) != 2 || !strings.Contains(content.Text, "Operative attachment") || strings.Contains(content.Text, "Weather navigation") {
		t.Fatal("article/attachment scope", content, err)
	}
	// HTML attachments preserve all their text even if they happen to contain
	// the parent's configured container selector.
	version.Resources[1].MediaType = "text/html"
	version.Resources[1].LocalProcessing = policy
	docs.bodies[pdf] = []byte(`<div id="notice">Attachment heading.</div><p>Operative attachment: chiusura per rischio idraulico.</p>`)
	content, err = GatherContent(context.Background(), docs, fakeOCR{}, version)
	if err != nil || !strings.Contains(content.Text, "Operative attachment") {
		t.Fatal("HTML attachment truncated", content, err)
	}
	old := content.Hash
	policy.HTMLBodySelectors = []string{"div#notice"}
	next, _ := GatherContent(context.Background(), docs, fakeOCR{}, version)
	if next.Hash == old || next.Text != content.Text {
		t.Fatal("equivalent text conflated different reviewed policies")
	}
}

const regionalVigilanceFixture = `<p>Emissione di Lunedì, 13 Aprile 2026, ore 12.48</p><table><tr><th>Fenomeno</th><th>13 Aprile 2026</th><th>14 Aprile 2026</th></tr><tr><td>pioggia</td><td>A4</td><td>A5</td></tr><tr><td>temporali</td><td>A4</td><td>A5</td></tr><tr><td>vento</td><td>A4</td><td>A5</td></tr><tr><td>mare</td><td>A4</td><td>A5</td></tr><tr><td>neve</td><td>A4</td><td>A5</td></tr><tr><td>ghiaccio</td><td>A4</td><td>A5</td></tr></table>`

var emptyZoneVigilanceFixture = `<h1>Bollettino di Vigilanza Meteorologica Regionale</h1>` + strings.NewReplacer("A4", "", "A5", "").Replace(regionalVigilanceFixture)

const regionalTableFixture = `<p>Emissione di Lunedì, 13 Aprile 2026, ore 12.48</p><table><tr><th>ZONE</th><th>RISCHIO</th><th>VALIDITÀ</th><th>CRITICITÀ</th></tr><tr><td>A4 A5</td><td>VENTO</td><td>dalle 12:00 del 13/04/2026 alle 23:59 del 13/04/2026</td><td>GIALLO</td></tr></table>`

func TestStructuredClassificationUsesNoProviderAndKeepsFallback(t *testing.T) {
	for _, tc := range []struct {
		name, body, format string
		complete, local    bool
	}{
		{"regional table", regionalTableFixture, "cfr-criticality-v1", true, true},
		{"vigilance table", regionalVigilanceFixture, "cfr-vigilance-v1", true, true},
		{"vigilance without listed zones", emptyZoneVigilanceFixture, "cfr-vigilance-v1", true, true},
		{"empty vigilance table without title", strings.NewReplacer("A4", "", "A5", "").Replace(regionalVigilanceFixture), "cfr-vigilance-v1", true, false},
		{"vigilance title without table", `<h1>Bollettino di Vigilanza Meteorologica Regionale</h1><p>Emissione di Lunedì, 13 Aprile 2026, ore 12.48</p>`, "cfr-vigilance-v1", true, false},
		{"vigilance missing phenomenon", strings.Replace(emptyZoneVigilanceFixture, "ghiaccio", "altro", 1), "cfr-vigilance-v1", true, false},
		{"no event", `<b>NESSUN AVVISO IN CORSO DI VALIDITÀ O EVENTO IN CORSO</b>`, "cfr-monitoring-v1", true, true},
		{"unknown level", strings.Replace(regionalTableFixture, "GIALLO", "BLU", 1), "cfr-criticality-v1", true, false},
		{"ordinary news", `<p>La mostra Allerta meteo apre domani.</p>`, "cfr-criticality-v1", true, false},
		{"missing policy", regionalTableFixture, "", true, false},
		{"incomplete", regionalTableFixture, "cfr-criticality-v1", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			url := "https://regional.example/bulletin"
			policy := &registry.LocalProcessing{Version: registry.LocalProcessingVersion, StructuredFormat: tc.format}
			if tc.format == "" {
				policy = nil
			}
			docs := &fakeDocuments{version: documents.Version{ID: 8, Complete: tc.complete, Resources: []documents.Reference{{URL: url, Role: "original", Required: true, MediaType: "text/html", LocalProcessing: policy}}}, bodies: map[string][]byte{url: []byte(tc.body)}}
			adapter := &fakeAdapter{response: inference.Response{Content: `{"relevant":false,"reason_code":"not_relevant","evidence_quote":"` + visibleHTML([]byte(tc.body)) + `"}`, Model: "fixture"}}
			results, process := &memoryResults{}, &fakeProcessing{}
			runner := &Runner{Documents: docs, OCR: fakeOCR{}, Processing: process, Results: results, Adapter: adapter, Model: "fixture", ConfigurationVersion: "base", LocalConfigurationVersion: "local-first"}
			payload, _ := json.Marshal(Payload{DocumentVersionID: 8, Workload: "evaluation"})
			if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 2, Attempt: 1, Payload: payload}); err != nil {
				t.Fatal(err)
			}
			if tc.local {
				if len(adapter.requests) != 0 || results.value.ReturnedModel != StructuredClassifierVersion || results.value.Relevant == nil || !*results.value.Relevant || process.finished[0].Usage.Status != "not_applicable" || results.value.InputTokens != nil {
					t.Fatal("local result misattributed", results.value)
				}
				if tc.format == "cfr-criticality-v1" && !strings.Contains(results.value.EvidenceQuote, "VENTO") {
					t.Fatal("regional evidence lost weather risk", results.value.EvidenceQuote)
				}
				content, _ := GatherContent(context.Background(), docs, fakeOCR{}, docs.version)
				segments, _ := SegmentContent(8, content)
				if _, ok := RemapResult(*results.value, segments, content, 9, 9, time.Now()); !ok {
					t.Fatal("validated local result not reusable")
				}
				content.Text = "changed evidence"
				if _, ok := RemapResult(*results.value, segments, content, 9, 9, time.Now()); ok {
					t.Fatal("unsupported local quotation reused")
				}
			} else if tc.complete && len(adapter.requests) == 0 {
				t.Fatal("unrecognized content bypassed classifier")
			}
			if !tc.complete && (results.value.Status != "undetermined" || len(adapter.requests) != 0) {
				t.Fatal("incomplete content classified")
			}
			if policy != nil && process.lastConfiguration != "local-first" {
				t.Fatal("local-first processing config lost")
			}
		})
	}
}

func TestManifestSeparatesLocalPoliciesWithIdenticalArticleText(t *testing.T) {
	url := "https://municipal.example/notice"
	body := []byte(`<main id="notice">Ordina la chiusura del ponte per rischio idraulico.</main>`)
	policy := &registry.LocalProcessing{Version: registry.LocalProcessingVersion, HTMLBodySelectors: []string{"main"}}
	v := documents.Version{ID: 1, DocumentID: 1, Complete: true, Metadata: json.RawMessage(`{}`), Resources: []documents.Reference{{URL: url, SourceID: "municipal", Role: "original", MediaType: "text/html", Hash: manifestHash(body), LocalProcessing: policy, Required: true}}}
	docs := &fakeDocuments{bodies: map[string][]byte{url: body}}
	build := func() InputManifest {
		t.Helper()
		content, err := GatherContent(context.Background(), docs, fakeOCR{}, v)
		if err != nil {
			t.Fatal(err)
		}
		m, err := BuildManifest(context.Background(), docs, fakeOCR{}, v, content, "local-first")
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	first := build()
	policy.HTMLBodySelectors = []string{"#notice"}
	second := build()
	if !first.Reusable || !second.Reusable || first.Sections[0].Text != second.Sections[0].Text || first.Hash == second.Hash || first.Resources[0].Identity != second.Resources[0].Identity {
		t.Fatal("policy partition lost")
	}
}
