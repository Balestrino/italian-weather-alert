package interpretation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
)

type preflightFixture struct{ bodies map[string][]byte }

func (f preflightFixture) Version(context.Context, int64) (documents.Version, error) {
	return documents.Version{}, nil
}
func (f preflightFixture) Read(_ context.Context, _ int64, url string) ([]byte, error) {
	return f.bodies[url], nil
}

type preflightRenderer struct{}

func (preflightRenderer) Render(_ context.Context, b []byte) ([]ocr.PageImage, error) {
	return []ocr.PageImage{{Number: 1, Bytes: []byte(strings.Split(string(b), ";metadata=")[0])}}, nil
}

func TestPreflightMeaningfulChanges(t *testing.T) {
	ctx := context.Background()
	build := func(source, html, pdf, metadata string, missing bool) (string, bool) {
		t.Helper()
		f := preflightFixture{map[string][]byte{"https://example/page": []byte(html), "https://example/file": []byte(pdf)}}
		v := documents.Version{DocumentID: 1, Complete: !missing, Metadata: json.RawMessage(metadata), Resources: []documents.Reference{{URL: "https://example/page", SourceID: source, Role: "original", MediaType: "text/html", Hash: preflightHash([]byte(html)), Required: true}, {URL: "https://example/file", SourceID: source, Role: "attachment", MediaType: "application/pdf", Hash: preflightHash([]byte(pdf)), Required: true}}}
		if missing {
			v.Resources[1].Missing = "unavailable"
		}
		p := Preflight{Documents: f, Renderer: preflightRenderer{}, Configuration: "cfg", RendererIdentity: "renderer"}
		m, err := p.Build(ctx, v)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := json.Marshal(m)
		return preflightHash(b), m.Complete
	}
	a, _ := build("cfr-criticality", "Warning 23 September", "map-orange;metadata=one", `{"issued":"2026-09-23"}`, false)
	b, _ := build("cfr-criticality", "Warning 23 September", "map-orange;metadata=two", `{"issued":"2026-09-23"}`, false)
	if a != b {
		t.Fatal("metadata-only PDF differs")
	}
	for _, tc := range []struct {
		html, pdf, metadata string
		missing             bool
	}{
		{"Warning 24 September", "map-orange;metadata=two", `{"issued":"2026-09-23"}`, false},
		{"Warning 23 September", "map-red;metadata=two", `{"issued":"2026-09-23"}`, false},
		{"Warning 23 September", "map-orange;metadata=two", `{"issued":"2026-09-24"}`, false},
		{"Warning 23 September", "map-orange;metadata=two", `{"issued":"2026-09-23"}`, true},
	} {
		h, complete := build("cfr-criticality", tc.html, tc.pdf, tc.metadata, tc.missing)
		if h == a || (tc.missing && complete) {
			t.Fatal("meaningful change lost")
		}
	}
	one := "<main>Ordinanza 34, ore 18:00</main><!-- js-view-dom-id-" + strings.Repeat("a", 64) + " -->"
	two := strings.ReplaceAll(one, strings.Repeat("a", 64), strings.Repeat("b", 64))
	a, _ = build("calcinaia-municipal", one, "unchanged", `{}`, false)
	b, _ = build("calcinaia-municipal", two, "unchanged", `{}`, false)
	if a != b {
		t.Fatal("reviewed comment differs")
	}
	b, _ = build("calcinaia-municipal", strings.Replace(two, "18:00", "19:00", 1), "unchanged", `{}`, false)
	if a == b {
		t.Fatal("operative time lost")
	}
	a, _ = build("unknown-source", one, "unchanged", `{}`, false)
	b, _ = build("unknown-source", two, "unchanged", `{}`, false)
	if a == b {
		t.Fatal("unreviewed normalization")
	}
	one = `<meta name="csrf-token" content="` + strings.Repeat("a", 40) + `"><main>Chiusura 8 ottobre ore 18:00</main><input type="hidden" name="_token" value="` + strings.Repeat("a", 40) + `">`
	two = strings.ReplaceAll(one, strings.Repeat("a", 40), strings.Repeat("b", 40))
	a, _ = build("cascina-municipal", one, "unchanged", `{}`, false)
	b, _ = build("cascina-municipal", two, "unchanged", `{}`, false)
	if a != b {
		t.Fatal("Cascina CSRF changes not equivalent")
	}
	for _, changed := range []string{strings.Replace(two, "18:00", "19:00", 1), strings.Replace(two, "8 ottobre", "9 ottobre", 1), strings.Replace(two, "Chiusura", "Revoca", 1)} {
		b, _ = build("cascina-municipal", changed, "unchanged", `{}`, false)
		if a == b {
			t.Fatal("Cascina meaningful change lost")
		}
	}
	b, complete := build("cascina-municipal", two, "unchanged", `{}`, true)
	if complete || a == b {
		t.Fatal("Cascina missing resource concealed")
	}
}
