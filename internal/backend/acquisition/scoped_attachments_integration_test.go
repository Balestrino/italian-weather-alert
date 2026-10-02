//go:build integration

package acquisition

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

type scopedObjects map[string][]byte

func (s scopedObjects) Ensure(ctx context.Context, hash string, body []byte) error {
	if fmt.Sprintf("%x", sha256.Sum256(body)) != hash {
		return documents.ErrStorage
	}
	s[hash] = append([]byte(nil), body...)
	_, err := s.Read(ctx, hash, int64(len(body)))
	return err
}

func (s scopedObjects) Read(_ context.Context, hash string, size int64) ([]byte, error) {
	body, found := s[hash]
	if !found || int64(len(body)) != size || fmt.Sprintf("%x", sha256.Sum256(body)) != hash {
		return nil, documents.ErrStorage
	}
	return body, nil
}

func TestScopedAttachmentsPersistFailureAndRecover(t *testing.T) {
	ctx := context.Background()
	pool := acquisitionTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(registry.Migrate(ctx, pool))
	must(documents.Migrate(ctx, pool))
	reg := registry.New(pool)
	policy := scopedFixture()
	const parent = "https://municipal.example/news/order"
	const pdf = "https://assets.example/s3/42/allegati/order.pdf"
	must(reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "Fixture municipality", OfficialURL: "https://municipal.example"}))
	must(reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "fixture", URL: "https://municipal.example"}))
	cfg := registry.Configuration{URL: parent, Sections: []string{parent}, AccessMethod: "crawl4ai", Attribution: "Fixture municipality", Attachments: policy, Policy: policy.External[0].Policy}
	must(reg.CreateSource(ctx, registry.Source{ID: "source", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "050004"}, cfg, "fixture"))
	objects := scopedObjects{}
	store := documents.New(pool, objects)
	crawler := &fixtureCrawler{pages: map[string]Page{parent: {StatusCode: 200, HTML: []byte(`<div class="article-content"><a href="` + pdf + `">Order</a></div>`)}}}
	resources := &boundedFixture{fixtureCrawler{pages: map[string]Page{pdf: {StatusCode: 200, MediaType: "application/pdf", HTML: []byte("<html>upstream error</html>")}}}}
	engine := Engine{Retained: store, Crawler: crawler, Resources: resources, Tracking: &fakeTracker{}, validatePDF: func(_ context.Context, p Page) error {
		if len(p.HTML) < 5 || string(p.HTML[:5]) != "%PDF-" {
			return errInvalidPDF
		}
		return nil
	}}
	state := &acquisitionState{}
	failed, err := engine.retainPlanned(ctx, "source", 1, PlannedDocument{URL: parent}, time.Now(), state, policy)
	if err != ErrRequiredAttachment || state.errorCode != "invalid_attachment_pdf" {
		t.Fatalf("invalid PDF did not fail visibly: %v %s", err, state.errorCode)
	}
	v, err := store.Version(ctx, failed.VersionID)
	must(err)
	if v.Complete || len(v.Resources) != 2 {
		t.Fatalf("failure evidence missing: %#v", v)
	}
	for _, r := range v.Resources {
		if r.URL == pdf && (r.Missing != "unavailable" || r.Hash != "") {
			t.Fatalf("bad response retained as a PDF: %#v", r)
		}
	}
	resources.pages[pdf] = Page{StatusCode: 200, MediaType: "application/pdf", HTML: syntheticPDF()}
	recovered, err := engine.retainPlanned(ctx, "source", 1, PlannedDocument{URL: parent}, time.Now(), &acquisitionState{}, policy)
	must(err)
	v, err = store.Version(ctx, recovered.VersionID)
	must(err)
	if !v.Complete || recovered.VersionID == failed.VersionID {
		t.Fatal("recovery overwrote failed evidence or stayed incomplete")
	}
	_, err = store.Read(ctx, recovered.VersionID, pdf)
	must(err)
	published := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	withDate, err := engine.retainPlanned(ctx, "source", 1, PlannedDocument{URL: parent, PublicationDate: &published}, time.Now(), &acquisitionState{}, policy)
	must(err)
	if withDate.VersionID == recovered.VersionID {
		t.Fatal("dated collection reused a preview acquisition with different metadata")
	}
	_, err = store.Retain(ctx, documents.Acquisition{ID: "outside", SourceID: "source", Configuration: 1, URL: parent, Resources: []documents.Resource{
		{URL: parent, Role: "original", Required: true, SourceID: "source", Configuration: 1, MediaType: "text/html", Bytes: []byte("fixture")},
		{URL: "https://assets.example/s3/43/allegati/outside.pdf", Role: "attachment", Required: true, SourceID: "source", Configuration: 1, MediaType: "application/pdf", Bytes: syntheticPDF()},
	}})
	if err != documents.ErrPolicy {
		t.Fatalf("storage accepted outside-scope attachment: %v", err)
	}
}
