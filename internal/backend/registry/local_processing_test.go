package registry

import (
	"testing"
	"time"
)

func localPolicy() *LocalProcessing {
	return &LocalProcessing{Version: LocalProcessingVersion, Evidence: Evidence{URL: "https://municipal.example/review", Locator: "synthetic article and text-only PDF review", ObservedAt: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)}, HTMLBodySelectors: []string{"article.body"}, TextPDFPathPrefixes: []string{"/ordinanze/"}}
}

func TestLocalProcessingPolicyBoundaries(t *testing.T) {
	p := localPolicy()
	if !p.valid(nil) || !(*LocalProcessing)(nil).valid(nil) {
		t.Fatal("valid optional policy rejected")
	}
	for _, selector := range []string{"article", "#notice", ".article-body", "main#notice", "div.article_body"} {
		if !ValidBodySelector(selector) {
			t.Fatal(selector)
		}
	}
	for _, selector := range []string{"", "*", "body div", "[hidden]", ".a.b", "script:first-child", "#"} {
		if ValidBodySelector(selector) {
			t.Fatal("unbounded selector", selector)
		}
	}
	for _, mutate := range []func(*LocalProcessing){
		func(p *LocalProcessing) { p.Version = "future" },
		func(p *LocalProcessing) { p.Evidence = Evidence{} },
		func(p *LocalProcessing) { p.HTMLBodySelectors = []string{".a", ".a"} },
		func(p *LocalProcessing) { p.TextPDFPathPrefixes = []string{"/"} },
		func(p *LocalProcessing) { p.TextPDFPathPrefixes = []string{"/a/../b/"} },
		func(p *LocalProcessing) { p.TextPDFPathPrefixes = []string{"/a/%2f/"} },
		func(p *LocalProcessing) { p.TextPDFPathPrefixes = []string{"/a/*/"} },
		func(p *LocalProcessing) { p.StructuredFormat = "cfr-criticality-v1" },
	} {
		q := localPolicy()
		mutate(q)
		if q.valid(nil) {
			t.Fatalf("invalid policy accepted: %+v", q)
		}
	}
	for _, raw := range []string{"https://municipal.example/ordinanze/order.pdf", "https://municipal.example/ordinanze/nested/order.pdf"} {
		if !p.TextPDF(raw) {
			t.Fatal(raw)
		}
	}
	for _, raw := range []string{"https://municipal.example/other/order.pdf", "https://municipal.example/ordinanze-other/order.pdf", "https://municipal.example/ordinanze/../maps/a.pdf", "https://municipal.example/ordinanze/%2e%2e/maps/a.pdf", "https://municipal.example/ordinanze/%61.pdf", "https://municipal.example/ordinanze/a.pdf#x"} {
		if p.TextPDF(raw) {
			t.Fatal("scope escaped", raw)
		}
	}
	id := p.Identity()
	p.HTMLBodySelectors = []string{"#notice"}
	if p.Identity() == id {
		t.Fatal("selector policy identities collide")
	}
	p = &LocalProcessing{Version: LocalProcessingVersion, Evidence: p.Evidence, StructuredFormat: "cfr-criticality-v1"}
	regional := &RegionalProductContract{Kind: "criticality"}
	if !p.valid(regional) || p.valid(&RegionalProductContract{Kind: "vigilance"}) {
		t.Fatal("regional policy scope")
	}
	p.TextPDFPathPrefixes = []string{"/prints/"}
	if p.valid(regional) {
		t.Fatal("graphical regional product declared text-only")
	}
}
