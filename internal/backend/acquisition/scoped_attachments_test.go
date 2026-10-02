package acquisition

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func scopedFixture() *registry.AttachmentPolicy {
	e := registry.Evidence{URL: "https://municipal.example/news/order", Locator: "Reviewed official attachment", ObservedAt: time.Now()}
	return &registry.AttachmentPolicy{ContentClass: "article-content", ValidatePDF: true, External: []registry.AttachmentScope{{Origin: "https://assets.example", PathPrefix: "/s3/42/allegati/", Referral: e, Policy: registry.Policy{Evidence: &e, CollectionPermitted: true, RetentionPermitted: true, Conditions: "Reviewed municipal documents."}}}}
}

type boundedFixture struct{ fixtureCrawler }

func (f *boundedFixture) CrawlBounded(ctx context.Context, raw string, allowed func(string) bool) (Page, error) {
	if !allowed(raw) {
		return Page{}, ErrInvalidConfiguration
	}
	return f.Crawl(ctx, raw)
}

func syntheticPDF() []byte {
	var b strings.Builder
	b.WriteString("%PDF-1.4\n")
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [3 0 R] /Count 1 >>", "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 100 100] /Resources << >> /Contents 4 0 R >>", "<< /Length 0 >>\nstream\n\nendstream"}
	offsets := []int{0}
	for i, object := range objects {
		offsets = append(offsets, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 5\n0000000000 65535 f \n")
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xref)
	return []byte(b.String())
}

func TestScopedAttachmentsSharePreviewAndScheduledChecks(t *testing.T) {
	const section = "https://municipal.example/news"
	const parent = section + "/order"
	const pdf = "https://assets.example/s3/42/allegati/order.pdf"
	for _, invalid := range []bool{false, true} {
		for _, preview := range []bool{false, true} {
			t.Run(fmt.Sprintf("invalid=%v/preview=%v", invalid, preview), func(t *testing.T) {
				policy := scopedFixture()
				crawler := &fixtureCrawler{pages: map[string]Page{section: {StatusCode: 200, HTML: []byte("listing"), Links: []Link{{URL: parent}}}, parent: {URL: parent, StatusCode: 200, HTML: []byte(`<div class="article-content"><a href="` + pdf + `">Order</a><a href="` + pdf + `">Duplicate</a></div><footer><a href="https://generic.example/guide.pdf">Site guide</a></footer>`)}}}
				body := syntheticPDF()
				if invalid {
					body = []byte("<html>Access denied</html>")
				}
				resources := &boundedFixture{fixtureCrawler{pages: map[string]Page{pdf: {URL: pdf, StatusCode: 200, MediaType: "application/pdf", HTML: body}}}}
				retained := &fakeRetention{}
				reg := &fakeRegistry{version: registry.Version{Configuration: registry.Configuration{Attachments: policy, AccessMethod: "crawl4ai", Sections: []string{section}, Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true}, Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/news/"}, PaginationParameter: "disabled", MaxPagesPerSection: 1, MaxDocuments: 10, ListingContentMarkers: []string{"listing"}}}}}
				engine := Engine{Registry: reg, Crawler: crawler, Resources: resources, Retained: retained, Tracking: &fakeTracker{}, validatePDF: func(_ context.Context, p Page) error {
					if !strings.HasPrefix(string(p.HTML), "%PDF-") {
						return errInvalidPDF
					}
					return nil
				}}
				var err error
				if preview {
					_, err = engine.Preview(context.Background(), "source", 2, "operator")
				} else {
					_, err = engine.retainPlanned(context.Background(), "source", 2, PlannedDocument{URL: parent}, time.Now(), &acquisitionState{}, policy)
				}
				if (err != nil) != invalid {
					t.Fatalf("wrong acquisition result: %v", err)
				}
				if len(resources.calls) != 1 || resources.calls[0] != pdf {
					t.Fatalf("wrong dependencies: %v", resources.calls)
				}
				item := retained.items[len(retained.items)-1]
				if len(item.Resources) != 2 {
					t.Fatalf("footer retained or original omitted: %#v", item.Resources)
				}
				if invalid && item.Resources[1].Missing != "unavailable" {
					t.Fatalf("invalid PDF accepted: %#v", item.Resources[1])
				}
				if invalid && reg.recorded.URL != "" {
					t.Fatal("failed preview recorded as successful")
				}
			})
		}
	}
}

func TestScopedContentContainerCannotDisappearSilently(t *testing.T) {
	if _, err := documentPDFLinks("https://municipal.example/order", []byte(`<main>No configured container</main>`), "article-content"); err == nil {
		t.Fatal("missing content container accepted")
	}
	links, err := documentPDFLinks("https://municipal.example/order", []byte(`<div class="article-content">No attachments in this notice</div>`), "article-content")
	if err != nil || len(links) != 0 {
		t.Fatalf("notice without attachments rejected: %v %v", links, err)
	}
}

func TestPDFValidationUsesRealParser(t *testing.T) {
	for _, body := range [][]byte{[]byte("<html>error</html>"), []byte("%PDF-1.4 truncated")} {
		if validateAttachmentPDF(context.Background(), Page{MediaType: "application/pdf", HTML: body}) == nil {
			t.Fatal("non-PDF/truncated response accepted")
		}
	}
	if _, err := exec.LookPath("pdfinfo"); err != nil {
		t.Skip("real parser is verified in the application image")
	}
	if err := validateAttachmentPDF(context.Background(), Page{MediaType: "application/pdf", HTML: syntheticPDF()}); err != nil {
		t.Fatalf("valid scanned/blank PDF rejected: %v", err)
	}
	if validateAttachmentPDF(context.Background(), Page{MediaType: "application/pdf", HTML: []byte("%PDF-1.4\ninvalid structure\n%%EOF")}) == nil {
		t.Fatal("structurally invalid PDF accepted")
	}
}

func TestScopedDiscoveryPreservesExplicitDependencies(t *testing.T) {
	const parent = "https://municipal.example/news/order"
	const explicit = "https://unreviewed.example/missing.pdf"
	crawler := &fixtureCrawler{pages: map[string]Page{parent: {StatusCode: 200, HTML: []byte(`<div class="article-content">No implicit attachments</div>`)}}}
	retained := &fakeRetention{}
	engine := Engine{Crawler: crawler, Retained: retained, Tracking: &fakeTracker{}}
	_, err := engine.retainPlanned(context.Background(), "source", 2, PlannedDocument{URL: parent, Resources: []PlannedResource{{URL: explicit, Required: true}}}, time.Now(), &acquisitionState{}, scopedFixture())
	if err != ErrRequiredAttachment || len(retained.items) != 1 || retained.items[0].Resources[1].Missing != "forbidden" {
		t.Fatalf("explicit required dependency hidden: %v %#v", err, retained.items)
	}
}
