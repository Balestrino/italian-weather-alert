//go:build integration

package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
)

func automaticFixture(t *testing.T) *verificationFixture {
	f := newVerificationFixture(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(MigrateTerritories(f.ctx, f.pool))
	meta, body := importFixture()
	preview, err := PreviewMunicipalities(body, meta)
	must(err)
	dataset, err := f.store.AdoptMunicipalities(f.ctx, body, meta, preview.SHA256, 0, "fixture")
	must(err)
	_, err = f.store.ConfigureRegion(f.ctx, "09", 1, RegionConfiguration{MunicipalityDataset: dataset, Profiles: []string{"municipal-html", "toscana-cfr"}}, "fixture")
	must(err)
	_, err = f.store.SetRegionEnabled(f.ctx, "09", 2, true, "fixture")
	must(err)
	_, err = f.store.SetMunicipalityEnabled(f.ctx, "09", "050004", 0, true, "fixture")
	must(err)
	tx, err := f.pool.Begin(f.ctx)
	must(err)
	defer tx.Rollback(f.ctx)
	for _, id := range []string{"municipal", "platform"} {
		must(territory.Associate(f.ctx, tx, id, "municipal", "050004", "municipal-html", "fixture"))
	}
	must(territory.Associate(f.ctx, tx, "regional", "criticality", "09", "toscana-cfr", "fixture"))
	must(tx.Commit(f.ctx))
	reg := registry.New(f.pool)
	versions, err := reg.Versions(f.ctx, "platform")
	must(err)
	cfg := versions[0].Configuration
	proof := *cfg.Policy.Evidence
	cfg.CittadinoInformato.ExternalAttachments = []registry.AttachmentScope{{Origin: "https://municipal.example", PathPrefix: "/acts/", Referral: proof, Policy: registry.Policy{Evidence: &proof, CollectionPermitted: true, RetentionPermitted: true, Conditions: "Synthetic linked act scope"}}}
	_, err = reg.AppendConfiguration(f.ctx, "platform", 1, cfg, "fixture")
	must(err)
	for _, id := range []string{"municipal", "platform", "regional"} {
		revision := 1
		if id == "platform" {
			revision = 2
		}
		must(reg.RecordPreview(f.ctx, id, revision, "fixture", registry.Evidence{URL: "https://municipal.example", Locator: "synthetic preview", ObservedAt: f.at}))
		must(reg.EnableCollection(f.ctx, id, revision, "fixture"))
	}
	must(acquisition.NewScheduleStore(f.pool).SyncEnabled(f.ctx, time.Now()))
	_, err = f.pool.Exec(f.ctx, "UPDATE acquisition_source_status SET last_complete_at=$1", time.Now())
	must(err)
	return f
}

func automaticRetain(f *verificationFixture, source, path, action string) (documents.Version, string) {
	f.t.Helper()
	f.serial++
	revision := 1
	if source == "platform" {
		revision = 2
	}
	u := "https://municipal.example/notices/" + path
	if source == "platform" {
		u = "https://cittadinoinformato.it/calcinaia/wp-json/cittadino/v2/aggiornamenti/" + path + "?comune=calcinaia"
	}
	text := "Ordinanza 42/2026 del 03/10/2026: " + action + " del Sottopasso sintetico, per lavori urgenti alla viabilità."
	body, _ := json.Marshal(map[string]string{"contenuto": text})
	v, err := f.docs.Retain(f.ctx, documents.Acquisition{ID: fmt.Sprintf("automatic-fixture-%d", f.serial), SourceID: source, Configuration: revision, URL: u, Metadata: json.RawMessage(`{"kind":"notice"}`), Resources: []documents.Resource{{URL: u, Role: "original", Required: true, SourceID: source, Configuration: revision, MediaType: "application/json", Bytes: body}, {URL: "https://municipal.example/acts/synthetic.pdf", Role: "attachment", SourceID: source, Configuration: revision, MediaType: "application/pdf", Bytes: []byte("%PDF-1.4 synthetic act")}}})
	if err != nil {
		f.t.Fatal(err)
	}
	return v, text
}

func automaticExtraction(f *verificationFixture, source string, v documents.Version, text, action string) {
	f.t.Helper()
	process := processing.New(f.pool)
	at := time.Now().UTC()
	must := func(err error) {
		f.t.Helper()
		if err != nil {
			f.t.Fatal(err)
		}
	}
	for _, stage := range []string{"classification", "extraction"} {
		must(process.RegisterConfiguration(f.ctx, processing.ConfigurationVersion{ID: "automatic-" + stage, Name: "synthetic", Stage: stage, Revision: "v1", LogicVersion: "synthetic", Settings: json.RawMessage(`{}`), CreatedAt: at}))
	}
	start := func(stage string) processing.Run {
		r, err := process.StartRun(f.ctx, processing.RunRequest{IdempotencyKey: fmt.Sprintf("auto-%s-%d", stage, v.ID), Stage: stage, Workload: "ordinary", ConfigurationVersionID: "automatic-" + stage, SourceID: &source, DocumentVersionID: &v.ID, Subject: json.RawMessage(`{}`), CreatedAt: at})
		must(err)
		return r
	}
	c := start("classification")
	relevant := true
	must(classification.NewStore(f.pool).Put(f.ctx, classification.Result{RunID: c.ID, DocumentVersionID: v.ID, Status: "classified", Relevant: &relevant, ReasonCode: "local_weather_measure", EvidenceQuote: text, ContentSHA256: strings.Repeat("a", 64), ContentComplete: true, ProviderResponseID: "synthetic", ReturnedModel: "synthetic", CreatedAt: at}))
	e := start("extraction")
	kind := map[string]string{"chiusura": "closure", "riapertura": "reopening"}[action]
	url := v.Resources[0].URL
	for _, r := range v.Resources {
		if r.Role == "original" {
			url = r.URL
		}
	}
	must(extraction.NewStore(f.pool).Put(f.ctx, extraction.Result{RunID: e.ID, DocumentVersionID: v.ID, ClassificationRunID: c.ID, Status: "extracted", ReasonCode: "measures_extracted", ContentSHA256: strings.Repeat("b", 64), ContentComplete: true, ProviderResponseID: "synthetic", ReturnedModel: "synthetic", CreatedAt: at, Measures: []extraction.Measure{{Ordinal: 1, Kind: kind, Subject: "Sottopasso sintetico", IndeterminateFields: []string{"place", "valid_from", "valid_until"}, Evidence: []extraction.Evidence{{Field: "kind", ResourceURL: url, Quote: action + " del Sottopasso sintetico"}, {Field: "subject", ResourceURL: url, Quote: action + " del Sottopasso sintetico"}}}}}))
}

func TestAutomaticVerificationRevisitsPrimaryAndPreservesHistory(t *testing.T) {
	f := automaticFixture(t)
	c, text := automaticRetain(f, "platform", "1", "chiusura")
	automaticExtraction(f, "platform", c, text, "chiusura")
	p, ptext := automaticRetain(f, "municipal", "act", "chiusura")
	automaticExtraction(f, "municipal", p, ptext, "chiusura")
	// Archiving interpretation does not revoke retained primary evidence.
	if _, err := f.pool.Exec(f.ctx, "INSERT INTO interpretation_archives(document_version_id,archived_at,cutoff,actor) VALUES($1,$2,$2,'fixture')", p.ID, time.Now()); err != nil {
		t.Fatal(err)
	}
	w := VerificationWorker{Store: f.store, Documents: f.docs}
	run := func(want int) {
		t.Helper()
		n, err := w.RunOnce(f.ctx, time.Now())
		if err != nil || n != want {
			t.Fatalf("receipts=%d want=%d err=%v", n, want, err)
		}
	}
	run(1)
	run(0)
	var first []byte
	if err := f.pool.QueryRow(f.ctx, "SELECT body FROM domain_verification_receipts ORDER BY id LIMIT 1").Scan(&first); err != nil {
		t.Fatal(err)
	}
	var receipt VerificationReceipt
	if json.Unmarshal(first, &receipt) != nil || receipt.Outcome != "corroborated" || receipt.Admitted || receipt.PrimaryRecordID != "" {
		t.Fatalf("comparison changed domain admission: %s", first)
	}
	// A title-only or partial-body match cannot establish identity. Change the
	// primary's action while retaining its explicit act/edition and subject.
	p, ptext = automaticRetain(f, "municipal", "act", "riapertura")
	automaticExtraction(f, "municipal", p, ptext, "riapertura")
	// The independently acquired primary and platform share the linked act URL.
	var err error
	run(1)
	var outcome string
	if err := f.pool.QueryRow(f.ctx, "SELECT body->>'outcome' FROM domain_verification_receipts ORDER BY id DESC LIMIT 1").Scan(&outcome); err != nil || outcome != "conflict" {
		t.Fatal(outcome, err)
	}
	_, err = f.pool.Exec(f.ctx, "UPDATE acquisition_source_status SET last_error_code='synthetic_unavailable',last_error_at=$1 WHERE source_id='municipal'", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	run(1)
	_, err = f.pool.Exec(f.ctx, "UPDATE acquisition_source_status SET last_error_code=NULL,last_error_at=NULL,last_complete_at=$1 WHERE source_id='municipal'", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	run(1)
	var old []byte
	if err := f.pool.QueryRow(f.ctx, "SELECT body FROM domain_verification_receipts ORDER BY id LIMIT 1").Scan(&old); err != nil || string(old) != string(first) {
		t.Fatal("history changed", err)
	}
	var measures int
	if err := f.pool.QueryRow(f.ctx, "SELECT count(*) FROM domain_local_measures").Scan(&measures); err != nil || measures != 0 {
		t.Fatal("comparison projected candidates", measures, err)
	}
	if _, err := f.pool.Exec(f.ctx, "INSERT INTO interpretation_archives(document_version_id,archived_at,cutoff,actor) VALUES($1,$2,$2,'fixture')", c.ID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := f.pool.Exec(f.ctx, "UPDATE acquisition_source_status SET last_error_code='synthetic_unavailable',last_error_at=$1 WHERE source_id='municipal'", time.Now()); err != nil {
		t.Fatal(err)
	}
	run(0)
	if err := registry.New(f.pool).Disable(f.ctx, "platform", 2, "fixture", false); err != nil {
		t.Fatal(err)
	}
	run(0)
}

func TestAutomaticRegionalDisplayDatesStayNonComparable(t *testing.T) {
	f := automaticFixture(t)
	f.retain("regional", "criticality", map[string]string{"product": "criticità", "risk": "vento", "zone": "A4", "level": "giallo"})
	url := "https://cittadinoinformato.it/calcinaia/wp-json/cittadino/v2/rischi/oggi"
	versions, err := registry.New(f.pool).Versions(f.ctx, "platform")
	if err != nil {
		t.Fatal(err)
	}
	cfg := versions[len(versions)-1].Configuration
	cfg.CittadinoInformato.Risks = true
	cfg.Sections = append(cfg.Sections, cfg.CittadinoInformato.BaseURL())
	cfg.Referral.Sections = append([]string(nil), cfg.Sections...)
	revision, err := registry.New(f.pool).AppendConfiguration(f.ctx, "platform", 2, cfg, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	reg := registry.New(f.pool)
	if err := reg.RecordPreview(f.ctx, "platform", revision, "fixture", *cfg.Policy.Evidence); err != nil {
		t.Fatal(err)
	}
	if err := reg.EnableCollection(f.ctx, "platform", revision, "fixture"); err != nil {
		t.Fatal(err)
	}
	_, err = f.docs.Retain(f.ctx, documents.Acquisition{ID: "automatic-risk", SourceID: "platform", Configuration: revision, URL: url, Metadata: json.RawMessage(`{"kind":"regional_republication"}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "platform", Configuration: revision, MediaType: "application/json", Bytes: []byte(`{"stati_allerta":[{"tipologia":"Vento","oggi":{"allerta":"giallo","data_bollettino":"04/10/2026","validita_cfr":null}}]}`)}}})
	if err != nil {
		t.Fatal(err)
	}
	w := VerificationWorker{Store: f.store, Documents: f.docs}
	if n, err := w.RunOnce(f.ctx, time.Now()); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	var body []byte
	if err := f.pool.QueryRow(f.ctx, "SELECT body FROM domain_verification_receipts ORDER BY id DESC LIMIT 1").Scan(&body); err != nil {
		t.Fatal(err)
	}
	var r VerificationReceipt
	if json.Unmarshal(body, &r) != nil || r.Outcome != "non_comparable" || r.Candidate.Fields["level"].Value != "yellow" || r.Candidate.Fields["risk"].Value != "wind" || r.Candidate.Fields["issuance"].Value != "" || r.Candidate.Fields["validity"].Value != "" || r.Admitted {
		t.Fatalf("display date or primary context inferred: %s", body)
	}
	if _, err := f.store.SetMunicipalityEnabled(f.ctx, "09", "050004", 1, false, "fixture"); err != nil {
		t.Fatal(err)
	}
	if n, err := w.RunOnce(f.ctx, time.Now()); err != nil || n != 0 {
		t.Fatal("disabled municipality executed", n, err)
	}
}
