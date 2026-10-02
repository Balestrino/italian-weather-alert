package classification

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
)

func TestCascinaCSRFPolicyAndManifest(t *testing.T) {
	const source = "cascina-municipal"
	const url = "https://municipal.example/it/news/42/avviso"
	one := `<meta name="csrf-token" content="` + strings.Repeat("a", 40) + `"><main><a href="/ordinanza">Chiusura 8 ottobre, ore 18:00</a></main><form id="ricerca"><input type="hidden" name="_token" value="` + strings.Repeat("a", 40) + `"></form>`
	two := strings.ReplaceAll(one, strings.Repeat("a", 40), strings.Repeat("b", 40))
	build := func(html, namedSource string, complete bool) InputManifest {
		t.Helper()
		body := []byte(html)
		docs := &fakeDocuments{bodies: map[string][]byte{url: body}}
		v := documents.Version{ID: 1, DocumentID: 1, Metadata: json.RawMessage(`{}`), Complete: complete, Resources: []documents.Reference{{URL: url, SourceID: namedSource, Role: "original", Required: true, MediaType: "text/html", Hash: manifestHash(body)}}}
		content, err := GatherContent(context.Background(), docs, fakeOCR{}, v)
		if err != nil {
			t.Fatal(err)
		}
		m, err := BuildManifest(context.Background(), docs, fakeOCR{}, v, content, "cfg")
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	a, b := build(one, source, true), build(two, source, true)
	if a.Hash != b.Hash || !a.Reusable || !strings.HasSuffix(a.Resources[0].Policy, ":"+CascinaHTMLPolicy) {
		t.Fatal("CSRF-only change not reusable under the versioned policy")
	}
	if bytes.Equal([]byte(one), CanonicalHTML(source, []byte(one))) {
		t.Fatal("fixture not normalized")
	}
	for _, s := range []string{"unknown", "livorno-municipal", "calcinaia-municipal"} {
		if build(one, s, true).Hash == build(two, s, true).Hash {
			t.Fatalf("normalization escaped Cascina: %s", s)
		}
	}
	for _, changed := range []string{strings.Replace(two, "18:00", "19:00", 1), strings.Replace(two, "8 ottobre", "9 ottobre", 1), strings.Replace(two, "Chiusura", "Revoca", 1), strings.Replace(two, "/ordinanza", "/revoca", 1), strings.Replace(two, `type="hidden"`, `type="text"`, 1)} {
		if build(changed, source, true).Hash == a.Hash {
			t.Fatal("meaningful change normalized away")
		}
	}
	if m := build(two, source, false); m.Reusable || m.Hash == a.Hash {
		t.Fatal("incomplete content reused")
	}
	if !bytes.Equal(CanonicalHTML(source, []byte(one)), CanonicalHTML(source, CanonicalHTML(source, []byte(one)))) {
		t.Fatal("normalization not idempotent")
	}
}

func TestManifestInvalidatesMeaningfulChangesAndIncompleteGraphics(t *testing.T) {
	ctx := context.Background()
	url := "https://source.example/notice"
	version := documents.Version{ID: 1, DocumentID: 1, Complete: true, Metadata: json.RawMessage(`{"issued":"2026-09-18","valid_until":"2026-09-19"}`), Resources: []documents.Reference{{URL: url, SourceID: "source", Role: "original", Required: true, MediaType: "text/html", Hash: strings.Repeat("a", 64)}, {URL: url + "/map.png", Role: "resource", Required: true, MediaType: "image/png", Hash: strings.Repeat("b", 64)}}}
	docs := &fakeDocuments{bodies: map[string][]byte{url: []byte("<main>Chiusura parco dal 18 settembre</main>")}}
	extracted := fakeOCR{pages: []ocr.PageResult{{ResourceURL: url + "/map.png", PageNumber: 1, Status: "complete", InputSHA256: strings.Repeat("c", 64), OutputSHA256: strings.Repeat("d", 64), ExtractedText: "allerta", ReturnedModel: "fixture"}}, resources: []ocr.ResourceResult{{ResourceURL: url + "/map.png", Status: "complete", PageCount: 1}}}
	build := func(v documents.Version, config string) InputManifest {
		t.Helper()
		content, err := GatherContent(ctx, docs, extracted, v)
		if err != nil {
			t.Fatal(err)
		}
		m, err := BuildManifest(ctx, docs, extracted, v, content, config)
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	base := build(version, "v1")
	assertChanged := func(m InputManifest) {
		t.Helper()
		if m.Hash == base.Hash {
			t.Fatal("meaningful change considered equivalent")
		}
	}
	assertChanged(build(version, "v2"))
	issuer := "different-authority"
	changedIssuer := version
	changedIssuer.IssuerID = &issuer
	assertChanged(build(changedIssuer, "v1"))
	attachment := version
	attachment.Resources = append(append([]documents.Reference(nil), version.Resources...), documents.Reference{URL: url + "/ordinanza.txt", Role: "attachment", Required: true, MediaType: "text/plain", Hash: manifestHash([]byte("Nuova ordinanza"))})
	docs.bodies[url+"/ordinanza.txt"] = []byte("Nuova ordinanza")
	assertChanged(build(attachment, "v1"))
	changed := version
	changed.Metadata = json.RawMessage(`{"issued":"2026-09-19","valid_until":"2026-09-20"}`)
	assertChanged(build(changed, "v1"))
	docs.bodies[url] = []byte("<main>Chiusura parco dal 19 settembre</main>")
	assertChanged(build(version, "v1"))
	docs.bodies[url] = []byte("<main>Chiusura parco dal 18 settembre</main>")
	extracted.pages[0].InputSHA256 = strings.Repeat("e", 64)
	assertChanged(build(version, "v1"))
	extracted.pages = nil
	m := build(version, "v1")
	assertChanged(m)
	if m.Reusable {
		t.Fatal("missing graphic provenance reusable")
	}
	version.Resources[1].Missing = "unavailable"
	m = build(version, "v1")
	if m.Reusable || m.Complete {
		t.Fatal("missing required resource concealed")
	}
}
func TestReviewedMunicipalDivClassNoise(t *testing.T) {
	a := []byte(`<div class="js-view-dom-id-50da3f0ff410e5d886a921ac5cee6ad43ee017ecd29414d4bcfed96ce551f212"><a href="/ordinanza">Divieto ore 18:00</a></div>`)
	b := []byte(`<div class="js-view-dom-id-8f11bd75b3a219fe7e8bc1bec28a49b051f470ce2bc2e95a5024f15e065bce9f"><a href="/ordinanza">Divieto ore 18:00</a></div>`)
	if string(CanonicalHTML("calcinaia-municipal", a)) != string(CanonicalHTML("calcinaia-municipal", b)) {
		t.Fatal("reviewed class noise differs")
	}
	if string(CanonicalHTML("unknown", a)) == string(CanonicalHTML("unknown", b)) {
		t.Fatal("unreviewed source changed")
	}
	for _, changed := range []string{strings.Replace(string(b), "18:00", "19:00", 1), strings.Replace(string(b), "/ordinanza", "/revoca", 1), strings.Replace(string(b), "Divieto", "Revoca", 1)} {
		if string(CanonicalHTML("calcinaia-municipal", a)) == string(CanonicalHTML("calcinaia-municipal", []byte(changed))) {
			t.Fatal("meaningful update normalized away")
		}
	}
}
