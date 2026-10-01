package registry

import (
	"strings"
	"testing"
	"time"
)

func TestReviewedDecorationIsExactAndConservative(t *testing.T) {
	url := "https://source.example/site-brand.png"
	hash := strings.Repeat("a", 64)
	p := &InferenceEligibility{Version: EligibilityVersion, Decorations: []DecorationRule{{URL: url, SHA256: hash, Reason: "reviewed_decoration", Evidence: Evidence{URL: "https://source.example/review", Locator: "fixture: branding outside bulletin content", ObservedAt: time.Now()}}}}
	if !p.valid("https://source.example") {
		t.Fatal("reviewed policy rejected")
	}
	if p.Decide(url, hash, "resource", "").Eligible {
		t.Fatal("reviewed decoration not excluded")
	}
	for _, tc := range []struct{ url, hash, role, missing string }{
		{url, strings.Repeat("b", 64), "resource", ""},
		{url, hash, "original", ""}, {url, hash, "attachment", ""}, {url, hash, "resource", "unavailable"},
		{"https://source.example/risk-map.png", hash, "resource", ""},
		{"https://source.example/legend.png", hash, "resource", ""},
		{"https://source.example/icon_lighting.png", hash, "resource", ""},
		{"https://source.example/image-only-notice.png", hash, "original", ""},
	} {
		if !p.Decide(tc.url, tc.hash, tc.role, tc.missing).Eligible {
			t.Fatalf("meaningful/unknown evidence excluded: %s", tc.url)
		}
	}
	previous := p.Decide(url, hash, "resource", "").Policy
	p.Decorations[0].SHA256 = strings.Repeat("c", 64)
	if p.Decide(url, hash, "resource", "").Policy == previous {
		t.Fatal("policy revision not reflected")
	}
	p.Decorations[0].Evidence = Evidence{}
	if p.valid("https://source.example") || !p.Decide(url, p.Decorations[0].SHA256, "resource", "").Eligible {
		t.Fatal("unreviewed exclusion accepted")
	}
}
