package classification

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

type equivalenceFixture struct {
	html, configuration, metadata, ocrOutput, ocrModel, ocrText, policy string
	issuer                                                              *string
	attachmentRequired, complete                                        bool
}

func baseEquivalenceFixture() equivalenceFixture {
	return equivalenceFixture{
		html:          `<main class="v1"><h1>Allerta meteo</h1><p>Il Comune dispone la chiusura del sottopasso dalle ore 18.</p></main>`,
		configuration: "classification-config-a", metadata: `{"source_publication_date":"2026-10-07"}`,
		ocrOutput: strings.Repeat("b", 64), ocrModel: "ocr-model-a", ocrText: "ORDINANZA chiusura sottopasso",
		policy: registry.EligibilityVersion + ":default", attachmentRequired: true, complete: true,
	}
}

func buildEquivalenceManifest(t *testing.T, f equivalenceFixture) InputManifest {
	t.Helper()
	const page, pdf = "https://comune.example/avviso", "https://comune.example/ordinanza.pdf"
	body := []byte(f.html)
	docs := &fakeDocuments{bodies: map[string][]byte{page: body}}
	v := documents.Version{ID: 7, DocumentID: 3, IssuerID: f.issuer, Metadata: json.RawMessage(f.metadata), Complete: f.complete, Resources: []documents.Reference{
		{URL: page, SourceID: "calcinaia-municipal", Role: "original", Required: true, MediaType: "text/html", Hash: manifestHash(body)},
		{URL: pdf, SourceID: "calcinaia-municipal", Role: "attachment", Required: f.attachmentRequired, MediaType: "application/pdf", Hash: strings.Repeat("c", 64), Inference: &registry.EligibilityDecision{Eligible: true, Policy: f.policy}},
	}}
	extracted := fakeOCR{
		pages:     []ocr.PageResult{{PageNumber: 1, ResourceURL: pdf, Status: "complete", InputSHA256: strings.Repeat("a", 64), OutputSHA256: f.ocrOutput, ExtractedText: f.ocrText, ReturnedModel: f.ocrModel}},
		resources: []ocr.ResourceResult{{ResourceURL: pdf, Status: "complete", PageCount: 1}},
	}
	content, err := GatherContent(context.Background(), docs, extracted, v)
	if err != nil {
		t.Fatal(err)
	}
	m, err := BuildManifest(context.Background(), docs, extracted, v, content, f.configuration)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestEquivalenceIgnoresConfigurationAndMarkupOutsideText(t *testing.T) {
	base := baseEquivalenceFixture()
	other := base
	other.html = strings.Replace(base.html, `class="v1"`, `class="v2" data-template="2026-10-21"`, 1)
	other.configuration = "classification-config-b"
	other.ocrModel = "ocr-model-b"
	a, b := buildEquivalenceManifest(t, base), buildEquivalenceManifest(t, other)
	if a.Hash == b.Hash {
		t.Fatal("fixture must differ in the configuration-bound manifest")
	}
	fa, okA := EquivalenceFingerprint(a)
	fb, okB := EquivalenceFingerprint(b)
	if !okA || !okB || fa != fb || len(fa) != 64 {
		t.Fatal("markup-only versions under two configurations are not equivalent", fa, fb)
	}
	// The stored jsonb body is re-spaced and drops the in-memory hash/reusable fields.
	encoded, _ := json.Marshal(a)
	var spaced bytes.Buffer
	if err := json.Indent(&spaced, encoded, "", "  "); err != nil {
		t.Fatal(err)
	}
	var stored InputManifest
	if err := json.Unmarshal(spaced.Bytes(), &stored); err != nil {
		t.Fatal(err)
	}
	stored.Reusable = true
	if fs, ok := EquivalenceFingerprint(stored); !ok || fs != fa {
		t.Fatal("stored manifest fingerprint differs from the in-memory one")
	}
}

func TestEquivalenceDetectsInterpretedChanges(t *testing.T) {
	base := baseEquivalenceFixture()
	want, ok := EquivalenceFingerprint(buildEquivalenceManifest(t, base))
	if !ok {
		t.Fatal("base fixture has no fingerprint")
	}
	issuer := "comune-calcinaia"
	for name, change := range map[string]func(*equivalenceFixture){
		"text":       func(f *equivalenceFixture) { f.html = strings.Replace(f.html, "ore 18", "ore 19", 1) },
		"metadata":   func(f *equivalenceFixture) { f.metadata = `{"source_publication_date":"2026-10-08"}` },
		"issuer":     func(f *equivalenceFixture) { f.issuer = &issuer },
		"required":   func(f *equivalenceFixture) { f.attachmentRequired = false },
		"policy":     func(f *equivalenceFixture) { f.policy = "resource-eligibility-v1:reviewed" },
		"ocr output": func(f *equivalenceFixture) { f.ocrOutput = strings.Repeat("d", 64) },
	} {
		f := base
		change(&f)
		got, ok := EquivalenceFingerprint(buildEquivalenceManifest(t, f))
		if !ok || got == want {
			t.Fatalf("%s change kept equivalence", name)
		}
	}
}

func TestEquivalenceRequiresCompleteReusableInput(t *testing.T) {
	f := baseEquivalenceFixture()
	f.complete = false
	if _, ok := EquivalenceFingerprint(buildEquivalenceManifest(t, f)); ok {
		t.Fatal("incomplete input has a fingerprint")
	}
	m := buildEquivalenceManifest(t, baseEquivalenceFixture())
	m.Reusable = false
	if _, ok := EquivalenceFingerprint(m); ok {
		t.Fatal("non-reusable input has a fingerprint")
	}
	m = buildEquivalenceManifest(t, baseEquivalenceFixture())
	m.Resources[0].Missing, m.Reusable = "unavailable", false
	if _, ok := EquivalenceFingerprint(m); ok {
		t.Fatal("missing resource input has a fingerprint")
	}
}
