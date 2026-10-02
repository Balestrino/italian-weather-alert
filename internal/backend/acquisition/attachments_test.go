package acquisition

import (
	"context"
	"testing"
	"time"
)

func TestPlannedCheckDiscoversPDFsAndRetainsUnavailableExternalReference(t *testing.T) {
	const parent = "https://source.example/novita/ordinanza"
	const local = "https://source.example/files/order.pdf"
	const external = "https://other.example/order.pdf"
	crawler := &fixtureCrawler{pages: map[string]Page{parent: {URL: parent, StatusCode: 200, MediaType: "text/html", HTML: []byte(`<main><a href="/files/order.pdf#page=2">Ordinanza</a><a href="/files/order.pdf">Copia</a><a href="https://other.example/order.pdf">Allegato esterno</a><a href="/news">Altra notizia</a></main>`)}}}
	resources := &fixtureCrawler{pages: map[string]Page{local: {URL: local, StatusCode: 200, MediaType: "application/pdf", HTML: []byte("%PDF retained fixture")}}}
	retained := &fakeRetention{}
	engine := Engine{Crawler: crawler, Resources: resources, Retained: retained, Tracking: &fakeTracker{}}
	state := acquisitionState{}
	page, err := engine.retainPlanned(context.Background(), "source", 1, PlannedDocument{URL: parent}, time.Now(), &state, nil)
	if err != ErrRequiredAttachment || page.VersionID == 0 || state.errorCode != "required_attachment_unavailable" {
		t.Fatalf("external required PDF silently omitted: %#v %v", state, err)
	}
	if len(resources.calls) != 1 || resources.calls[0] != local || len(crawler.calls) != 1 {
		t.Fatalf("wrong transport or external fetch: resources=%v html=%v", resources.calls, crawler.calls)
	}
	if len(retained.items) != 1 || len(retained.items[0].Resources) != 3 {
		t.Fatalf("attachment graph missing or duplicated: %#v", retained.items)
	}
	for _, resource := range retained.items[0].Resources {
		if resource.URL == external && (resource.Missing != "forbidden" || !resource.Required) {
			t.Fatalf("external restriction not retained: %#v", resource)
		}
	}
}

func TestLinkedPDFChangeCreatesNewParentVersion(t *testing.T) {
	const parent = "https://source.example/notice"
	const pdf = "https://source.example/order.pdf"
	crawler := &fixtureCrawler{pages: map[string]Page{parent: {URL: parent, StatusCode: 200, MediaType: "text/html", HTML: []byte(`<a href="/order.pdf">Ordinanza</a>`)}}}
	resources := &fixtureCrawler{pages: map[string]Page{pdf: {URL: pdf, StatusCode: 200, MediaType: "application/pdf", HTML: []byte("version one")}}}
	engine := Engine{Crawler: crawler, Resources: resources, Retained: &stableRetention{}, Tracking: &fakeTracker{}}
	first, err := engine.retainPlanned(context.Background(), "source", 1, PlannedDocument{URL: parent}, time.Now(), &acquisitionState{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	resources.pages[pdf] = Page{URL: pdf, StatusCode: 200, MediaType: "application/pdf", HTML: []byte("version two")}
	second, err := engine.retainPlanned(context.Background(), "source", 1, PlannedDocument{URL: parent}, time.Now(), &acquisitionState{}, nil)
	if err != nil || first.VersionID == second.VersionID || !second.Changed {
		t.Fatalf("changed attachment lost: %#v %#v %v", first, second, err)
	}
}
