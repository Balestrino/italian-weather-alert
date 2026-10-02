package registry

import (
	"testing"
	"time"
)

func attachmentFixture() *AttachmentPolicy {
	e := Evidence{URL: "https://municipal.example/news/order", Locator: "Linked municipal ordinance", ObservedAt: time.Now()}
	return &AttachmentPolicy{ContentClass: "article-content", ValidatePDF: true, External: []AttachmentScope{{Origin: "https://assets.example", PathPrefix: "/s3/42/allegati/", Referral: e, Policy: Policy{Evidence: &e, CollectionPermitted: true, RetentionPermitted: true, Conditions: "Reviewed municipal attachments; attribute the municipality."}}}}
}

func TestAttachmentScopeBoundaries(t *testing.T) {
	p := attachmentFixture()
	const parent = "https://municipal.example/news/order"
	for target, want := range map[string]bool{
		"https://municipal.example/order.pdf":                              true,
		"https://assets.example/s3/42/allegati/order.pdf":                  true,
		"https://ASSETS.example/s3/42/allegati/order.pdf":                  true,
		"https://assets.example/s3/43/allegati/order.pdf":                  false,
		"https://assets.example/s3/42/allegati-other/order.pdf":            false,
		"https://assets.example/s3/42/allegati/../private/order.pdf":       false,
		"https://assets.example/s3/42/allegati/%2e%2e/private/order.pdf":   false,
		"https://assets.example/s3/42/allegati/x%2forder.pdf":              false,
		"https://assets.example/s3/42/allegati/order.pdf?redirect=other":   false,
		"https://assets.example/s3/42/allegati/order.pdf#fragment":         false,
		"http://assets.example/s3/42/allegati/order.pdf":                   false,
		"https://assets.example:444/s3/42/allegati/order.pdf":              false,
		"https://assets.example.attacker.example/s3/42/allegati/order.pdf": false,
		"https://user@assets.example/s3/42/allegati/order.pdf":             false,
	} {
		if got := p.Allows(parent, target); got != want {
			t.Errorf("%s allowed=%v, want %v", target, got, want)
		}
	}
	if (*AttachmentPolicy)(nil).Allows(parent, "https://assets.example/s3/42/allegati/order.pdf") {
		t.Fatal("another revision inherited external permission")
	}
}

func TestAttachmentScopeRequiresNarrowReviewedPolicy(t *testing.T) {
	for name, mutate := range map[string]func(*AttachmentPolicy){
		"wildcard":              func(p *AttachmentPolicy) { p.External[0].Origin = "https://*.example" },
		"whole host":            func(p *AttachmentPolicy) { p.External[0].PathPrefix = "/" },
		"traversal":             func(p *AttachmentPolicy) { p.External[0].PathPrefix = "/s3/42/../" },
		"ambiguous path":        func(p *AttachmentPolicy) { p.External[0].PathPrefix = "/s3//42/" },
		"missing referral":      func(p *AttachmentPolicy) { p.External[0].Referral = Evidence{} },
		"missing reuse":         func(p *AttachmentPolicy) { p.External[0].Policy.Evidence = nil },
		"retention denied":      func(p *AttachmentPolicy) { p.External[0].Policy.RetentionPermitted = false },
		"invalid content class": func(p *AttachmentPolicy) { p.ContentClass = "article content" },
	} {
		t.Run(name, func(t *testing.T) {
			p := attachmentFixture()
			mutate(p)
			if p.Valid() {
				t.Fatal("unreviewed/broad scope accepted")
			}
		})
	}
}
