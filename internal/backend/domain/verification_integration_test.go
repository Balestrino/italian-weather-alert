//go:build integration

package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/embedding"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/interpretation"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

type verificationFixture struct {
	ctx     context.Context
	t       *testing.T
	pool    *pgxpool.Pool
	store   *Store
	docs    *documents.Store
	objects *retentionObjects
	at      time.Time
	serial  int
}

func newVerificationFixture(t *testing.T) *verificationFixture {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	pool := domainTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, acquisition.Migrate, ocr.Migrate, classification.Migrate, extraction.Migrate, linking.Migrate, embedding.Migrate, interpretation.Migrate, Migrate, MigrateGeography, MigrateTemporal, MigrateQuality, MigrateRetention, func(ctx context.Context, pool *pgxpool.Pool) error {
		body, err := os.ReadFile("../publicquery/schema.sql")
		if err != nil {
			return err
		}
		_, err = pool.Exec(ctx, string(body))
		return err
	}, MigrateVerification, MigrateVerification} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	at := time.Now().UTC().Truncate(time.Microsecond)
	reg := registry.New(pool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, authority := range []registry.Authority{{ID: "municipality", Name: "Synthetic municipality", OfficialURL: "https://municipal.example"}, {ID: "cfr", Name: "Synthetic regional origin", OfficialURL: "https://regional.example"}} {
		must(reg.CreateAuthority(ctx, authority))
	}
	for _, channel := range []registry.Channel{{ID: "municipal", PublisherID: "municipality", Platform: "municipal-site", URL: "https://municipal.example"}, {ID: "platform", PublisherID: "municipality", Platform: "cittadino-informato", URL: "https://cittadinoinformato.it/calcinaia/", External: true}, {ID: "regional", PublisherID: "cfr", Platform: "regional-origin", URL: "https://regional.example"}} {
		must(reg.CreateChannel(ctx, channel))
	}
	proof := registry.Evidence{URL: "https://municipal.example/referral", Locator: "synthetic verification fixture", ObservedAt: at}
	policy := registry.Policy{Evidence: &proof, CollectionPermitted: true, RetentionPermitted: true}
	contract := &registry.CittadinoInformatoContract{MunicipalityISTAT: "050004", MunicipalitySlug: "calcinaia", Publisher: "comune_calcinaia", Updates: true, PageSize: 1}
	sections := []string{contract.BaseURL() + "aggiornamenti/"}
	platformCfg := registry.Configuration{URL: contract.BaseURL(), Sections: sections, AccessMethod: registry.CittadinoInformatoAccess, Attribution: "synthetic", Policy: policy, CittadinoInformato: contract, Discovery: registry.Discovery{MaxPagesPerSection: 2, MaxDocuments: 10, BootstrapDays: 30}, Referral: &registry.Referral{Evidence: proof, Destination: contract.BaseURL(), Territory: "050004", ProductID: "municipal", Sections: sections, Context: "synthetic municipal referral"}}
	must(reg.CreateSource(ctx, registry.Source{ID: "platform", AuthorityID: "municipality", ChannelID: "platform", ProductID: "municipal", Territory: "050004"}, platformCfg, "fixture"))
	for _, item := range []struct{ id, authority, channel, product, territory, url string }{{"municipal", "municipality", "municipal", "municipal", "050004", "https://municipal.example"}, {"regional", "cfr", "regional", "criticality", "toscana", "https://regional.example"}, {"monitoring", "cfr", "regional", "monitoring", "toscana", "https://regional.example"}, {"outside-regional", "cfr", "regional", "criticality", "Umbria", "https://regional.example"}} {
		must(reg.CreateSource(ctx, registry.Source{ID: item.id, AuthorityID: item.authority, ChannelID: item.channel, ProductID: item.product, Territory: item.territory}, registry.Configuration{URL: item.url, Sections: []string{item.url + "/notices"}, AccessMethod: "fixture", Attribution: "synthetic", Policy: policy}, "fixture"))
	}
	objects := &retentionObjects{domainObjects: &domainObjects{items: map[string][]byte{}}}
	return &verificationFixture{ctx: ctx, t: t, pool: pool, store: New(pool), docs: documents.New(pool, objects), objects: objects, at: at}
}

func (f *verificationFixture) retain(source, path string, values map[string]string) VerificationEvidence {
	f.t.Helper()
	f.serial++
	url := "https://municipal.example/notices/" + path
	if source == "platform" {
		url = "https://cittadinoinformato.it/calcinaia/wp-json/cittadino/v2/aggiornamenti/" + path + "?comune=calcinaia"
	} else if source == "regional" || source == "monitoring" {
		url = "https://regional.example/notices/" + path
	}
	body, _ := json.Marshal(values)
	v, err := f.docs.Retain(f.ctx, documents.Acquisition{ID: fmt.Sprintf("verification-fixture-%d", f.serial), SourceID: source, Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: source, Configuration: 1, MediaType: "application/json", Bytes: body}}})
	if err != nil {
		f.t.Fatal(err)
	}
	if now := time.Now().UTC().Truncate(time.Microsecond); now.After(f.at) {
		f.at = now
	}
	fields := map[string]EvidenceSelection{}
	for name, value := range values {
		fields[name] = EvidenceSelection{ResourceURL: url, JSONPointer: "/" + name, EndByte: len(cleanLiteral(value)), Locator: "synthetic scalar " + name}
	}
	return VerificationEvidence{SourceID: source, VersionID: v.ID, Fields: fields}
}

func (f *verificationFixture) check(role, state string, evidence VerificationEvidence) VerificationCheck {
	return VerificationCheck{Role: role, State: state, Reason: "synthetic_bounded_check", CheckedAt: f.at, VerificationEvidence: evidence}
}
func (f *verificationFixture) request(id, kind string, candidate VerificationEvidence, checks ...VerificationCheck) VerificationRequest {
	return VerificationRequest{RequestID: id, CandidateKey: "synthetic-notice", Kind: kind, MunicipalityISTAT: "050004", Candidate: candidate, Checks: checks}
}
func (f *verificationFixture) verify(r VerificationRequest) VerificationReceipt {
	f.t.Helper()
	v, err := f.store.VerifyMultiSource(f.ctx, f.docs, r, f.at)
	if err != nil {
		f.t.Fatal(err)
	}
	return v
}
func localVerificationValues() map[string]string {
	return map[string]string{"reference": "Ordinanza sintetica 42/2026", "edition": "2026-10-03", "kind": "chiusura", "subject": "Sottopasso sintetico", "place": "Via Sintetica", "validity": "fino a nuovo ordine"}
}

func TestVerificationPersistenceAdmissionHistoryAndRevisions(t *testing.T) {
	f := newVerificationFixture(t)
	values := localVerificationValues()
	candidate, primary := f.retain("platform", "1", values), f.retain("municipal", "act-42", values)
	checks := []VerificationCheck{f.check("municipal", "available", primary), f.check("regional", "not_applicable", VerificationEvidence{})}
	req := f.request("first", "local_measure", candidate, checks...)
	receipt := f.verify(req)
	if !receipt.Admitted || receipt.Outcome != "corroborated" || receipt.DomainRecordID == "" || len(receipt.Checks) != 3 {
		t.Fatalf("admission: %+v", receipt)
	}
	persisted, err := New(f.pool).VerificationReceipt(f.ctx, receipt.ID)
	if err != nil || persisted.Candidate.Fields["subject"].Passage != values["subject"] || persisted.Candidate.Hash == "" {
		t.Fatal("persisted proof missing", err)
	}
	again := f.verify(req)
	if again.ID != receipt.ID {
		t.Fatal("same request duplicated receipt")
	}
	altered := req
	altered.CandidateKey = "another"
	if _, err = f.store.VerifyMultiSource(f.ctx, f.docs, altered, f.at); !errors.Is(err, ErrConflict) {
		t.Fatal("request identity reused with different evidence", err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r := req
			r.RequestID = fmt.Sprint("repeat-", n)
			_, err := f.store.VerifyMultiSource(f.ctx, f.docs, r, f.at)
			errs <- err
		}(n)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err = f.pool.QueryRow(f.ctx, "SELECT count(*) FROM domain_local_measures").Scan(&count); err != nil || count != 1 {
		t.Fatal("concurrent verification duplicated measures", count, err)
	}
	history, err := f.store.VerificationHistory(f.ctx, "platform", "synthetic-notice", 0, 100)
	if err != nil || len(history) != 3 {
		t.Fatal("history lost", len(history), err)
	}
	if _, err = f.pool.Exec(f.ctx, "UPDATE domain_verification_receipts SET body='{}' WHERE id=$1", receipt.ID); err == nil {
		t.Fatal("immutable receipt edited")
	}
	// A changed platform claim conflicts with the original; it creates no local fact.
	values["kind"] = "riapertura"
	changed := f.retain("platform", "1", values)
	mismatch := f.verify(f.request("changed-platform", "local_measure", changed, checks...))
	if mismatch.Admitted || mismatch.Outcome != "conflict" || mismatch.PrimaryRecordID != receipt.PrimaryRecordID {
		t.Fatal("platform conflict replaced municipal act")
	}
	// A primary revision supersedes only its explicit act/scope, at the known time.
	f.at = f.at.Add(time.Minute)
	revised := f.retain("municipal", "act-42", values)
	latest := f.verify(f.request("changed-primary", "local_measure", changed, f.check("municipal", "available", revised), f.check("regional", "not_applicable", VerificationEvidence{})))
	if !latest.Admitted || latest.DomainRecordID == receipt.DomainRecordID {
		t.Fatal("primary revision not retained")
	}
	var wasCurrent, isCurrent bool
	if err = f.pool.QueryRow(f.ctx, "SELECT domain_verification_record_current('local_measure',$1,$2),domain_verification_record_current('local_measure',$1,$3)", receipt.DomainRecordID, receipt.VerifiedAt, f.at).Scan(&wasCurrent, &isCurrent); err != nil || !wasCurrent || isCurrent {
		t.Fatal("revision leaked into earlier knowledge or duplicated current facts", err)
	}
	original, ok, err := f.store.LocalMeasure(f.ctx, receipt.DomainRecordID)
	if err != nil || !ok || original.Kind != "closure" {
		t.Fatal("old domain facts overwritten")
	}
	// A later fetch can return an earlier content hash. Acquisition order, not
	// the immutable version ID, chooses the current primary while reusing facts.
	values["kind"] = "chiusura"
	revertedPrimary := f.retain("municipal", "act-42", values)
	revertedCandidate := f.retain("platform", "1", values)
	if revertedPrimary.VersionID != primary.VersionID {
		t.Fatal("unchanged content lost version reuse")
	}
	f.at = f.at.Add(time.Minute)
	reverted := f.verify(f.request("reverted-primary", "local_measure", revertedCandidate, f.check("municipal", "available", revertedPrimary), f.check("regional", "not_applicable", VerificationEvidence{})))
	if reverted.DomainRecordID != receipt.DomainRecordID {
		t.Fatal("returning primary duplicated original measure")
	}
	if err = f.pool.QueryRow(f.ctx, "SELECT domain_verification_record_current('local_measure',$1,$2),domain_verification_record_current('local_measure',$3,$2)", receipt.DomainRecordID, f.at, latest.DomainRecordID).Scan(&isCurrent, &wasCurrent); err != nil || !isCurrent || wasCurrent {
		t.Fatal("returning content left obsolete measure current", err)
	}
	// Direct platform insertion cannot bypass the evidence verifier.
	if err = f.store.PutLocalMeasure(f.ctx, LocalMeasure{ID: "bypass", DocumentVersionID: candidate.VersionID, SourceID: "platform", MunicipalityISTAT: "050004", Kind: "closure", Subject: "invented"}); !errors.Is(err, ErrInvalid) {
		t.Fatal("platform bypass admitted", err)
	}
}

func TestVerificationMissingUnavailableNonComparableAndPrimaryOnly(t *testing.T) {
	f := newVerificationFixture(t)
	values := localVerificationValues()
	candidate := f.retain("platform", "1", values)
	for _, state := range []string{"missing", "unavailable"} {
		r := f.verify(f.request(state, "local_measure", candidate, f.check("municipal", state, VerificationEvidence{SourceID: "municipal"}), f.check("regional", "not_applicable", VerificationEvidence{})))
		want := state
		if state == "missing" {
			want = "missing_evidence"
		}
		if r.Admitted || r.DomainRecordID != "" || r.PrimaryRecordID != "" || r.Outcome != want {
			t.Fatalf("%s fabricated primary or conflict: %+v", state, r)
		}
		if r.Checks[2].Outcome != want {
			t.Fatal("platform corroborated itself without required primary", r.Checks[2].Outcome)
		}
	}
	values["edition"] = "2026-10-04"
	primary := f.retain("municipal", "act-42", values)
	r := f.verify(f.request("non-comparable", "local_measure", candidate, f.check("municipal", "available", primary), f.check("regional", "not_applicable", VerificationEvidence{})))
	if r.Admitted || r.Outcome != "non_comparable" || r.PrimaryRecordID == "" {
		t.Fatal("different edition labeled conflict or primary lost")
	}
	primaryOnly := f.verify(f.request("primary-only", "local_measure", primary, f.check("platform", "missing", VerificationEvidence{SourceID: "platform"}), f.check("regional", "not_applicable", VerificationEvidence{})))
	if !primaryOnly.Admitted || primaryOnly.DomainRecordID != r.PrimaryRecordID {
		t.Fatal("primary collection depends on a platform copy")
	}
	bad := candidate
	bad.SourceID = "municipal"
	if _, err := f.store.VerifyMultiSource(f.ctx, f.docs, f.request("wrong-owner", "local_measure", bad, f.check("platform", "missing", VerificationEvidence{SourceID: "platform"}), f.check("regional", "not_applicable", VerificationEvidence{})), f.at); !errors.Is(err, ErrVerificationEvidence) {
		t.Fatal("foreign version accepted", err)
	}
	// Selection ownership and offsets are checked against the retained bytes.
	bad = candidate
	bad.Fields = map[string]EvidenceSelection{"kind": {ResourceURL: "https://municipal.example/foreign", EndByte: 8, Locator: "foreign"}}
	if _, err := f.store.VerifyMultiSource(f.ctx, f.docs, f.request("bad-span", "local_measure", bad, f.check("municipal", "missing", VerificationEvidence{SourceID: "municipal"}), f.check("regional", "not_applicable", VerificationEvidence{})), f.at); !errors.Is(err, ErrVerificationEvidence) {
		t.Fatal("foreign resource accepted", err)
	}
	// Storage loss is an unavailable check, not contrary evidence.
	var hash string
	if err := f.pool.QueryRow(f.ctx, "SELECT object_hash FROM retained_resources WHERE version_id=$1", primary.VersionID).Scan(&hash); err != nil {
		f.t.Fatal(err)
	}
	f.objects.mu.Lock()
	saved := f.objects.items[hash]
	delete(f.objects.items, hash)
	f.objects.mu.Unlock()
	unavailable := f.verify(f.request("storage-unavailable", "local_measure", candidate, f.check("municipal", "available", primary), f.check("regional", "not_applicable", VerificationEvidence{})))
	if unavailable.Admitted || unavailable.Outcome != "unavailable" {
		t.Fatal("storage loss called conflict or admitted")
	}
	if unavailable.Checks[1].Configuration != 1 || unavailable.Checks[1].VersionID != primary.VersionID || unavailable.Checks[1].Hash == "" {
		t.Fatal("unavailable check discarded known version/configuration identity")
	}
	f.objects.mu.Lock()
	f.objects.items[hash] = saved
	f.objects.mu.Unlock()
	if _, err := f.pool.Exec(f.ctx, "UPDATE registry_sources SET interpretation_suspended_at=$1 WHERE id='municipal'", f.at); err != nil {
		t.Fatal(err)
	}
	suspended := f.verify(f.request("suspended", "local_measure", candidate, f.check("municipal", "available", primary), f.check("regional", "not_applicable", VerificationEvidence{})))
	if suspended.Admitted || suspended.Outcome != "unavailable" {
		t.Fatal("suspended primary admitted")
	}
}

func TestVerificationRegionalOriginWinsWithoutMajorityAndMonitoringColor(t *testing.T) {
	f := newVerificationFixture(t)
	values := map[string]string{"product": "criticità", "risk": "vento", "zone": "H2", "issuance": "2026-10-03T12:00:00+02:00", "validity": "2026-10-03T13:00:00+02:00/2026-10-04T13:00:00+02:00", "level": "giallo"}
	primary := f.retain("regional", "bulletin", values)
	values["level"] = "rosso"
	candidate, municipal := f.retain("platform", "1", values), f.retain("municipal", "republication", values)
	missing := f.verify(f.request("regional-missing", "regional_record", candidate, f.check("regional", "missing", VerificationEvidence{SourceID: "regional"}), f.check("municipal", "available", municipal)))
	if missing.Admitted || missing.PrimaryRecordID != "" {
		t.Fatal("two republications admitted without CFR")
	}
	for _, check := range missing.Checks {
		if check.Outcome != "missing_evidence" {
			t.Fatal("republications established their own originating corroboration", check.Role, check.Outcome)
		}
	}
	if _, err := f.store.VerifyMultiSource(f.ctx, f.docs, f.request("foreign-region", "regional_record", candidate, f.check("regional", "unavailable", VerificationEvidence{SourceID: "outside-regional"}), f.check("municipal", "available", municipal)), f.at); !errors.Is(err, ErrVerificationEvidence) {
		t.Fatal("another region accepted as Toscana origin", err)
	}
	receipt := f.verify(f.request("regional-conflict", "regional_record", candidate, f.check("regional", "available", primary), f.check("municipal", "available", municipal)))
	if receipt.Admitted || receipt.Outcome != "conflict" || receipt.DomainRecordID != "" || receipt.PrimaryRecordID == "" {
		t.Fatal("secondary majority overrode CFR", receipt)
	}
	for _, c := range receipt.Checks {
		if c.Outcome != "conflict" {
			t.Fatal("all channels must retain originating comparison", c.Role, c.Outcome)
		}
		if c.ComparisonToSourceID != "regional" || c.ComparisonToVersionID != primary.VersionID {
			t.Fatal("comparison direction or version lost")
		}
	}
	record, ok, err := f.store.Regional(f.ctx, receipt.PrimaryRecordID)
	if err != nil || !ok || len(record.Facts) != 1 || *record.Facts[0].Level != "yellow" || record.Facts[0].Zone != "H2" || record.SourceID != "regional" {
		t.Fatal("originating product or color changed", err, record)
	}
	values["validity"] = "2026-10-04T13:00:00+02:00/2026-10-05T13:00:00+02:00"
	next := f.retain("platform", "1", values)
	result := f.verify(f.request("different-period", "regional_record", next, f.check("regional", "available", primary), f.check("municipal", "available", municipal)))
	if result.Outcome != "non_comparable" || result.Admitted || result.PrimaryRecordID != receipt.PrimaryRecordID {
		t.Fatal("different period called a conflict")
	}
	values["product"] = "monitoraggio"
	delete(values, "level")
	monitoring := f.retain("monitoring", "monitoring", values)
	monitor := f.verify(f.request("monitoring", "regional_record", monitoring, f.check("municipal", "missing", VerificationEvidence{SourceID: "municipal"}), f.check("platform", "missing", VerificationEvidence{SourceID: "platform"})))
	record, ok, err = f.store.Regional(f.ctx, monitor.DomainRecordID)
	if err != nil || !ok || !monitor.Admitted || record.Facts[0].Level != nil {
		t.Fatal("monitoring invented an alert color", err)
	}
	// Operational phases use municipal evidence independently of regional alerts.
	phaseValues := map[string]string{"reference": "Avviso sintetico 7", "edition": "2026-10-03", "phase": "attenzione", "validity": "2026-10-03"}
	phaseCandidate, phasePrimary := f.retain("platform", "2", phaseValues), f.retain("municipal", "phase", phaseValues)
	phase := f.verify(f.request("phase", "operational_phase", phaseCandidate, f.check("municipal", "available", phasePrimary), f.check("regional", "not_applicable", VerificationEvidence{})))
	if !phase.Admitted || phase.DomainRecordID == "" {
		t.Fatal("municipal phase demanded regional alert")
	}
}

func TestVerificationRetentionKeepsCounterpartAndPurgesExpiredGraph(t *testing.T) {
	f := newVerificationFixture(t)
	values := localVerificationValues()
	candidate, primary := f.retain("platform", "1", values), f.retain("municipal", "act-42", values)
	receipt := f.verify(f.request("retention", "local_measure", candidate, f.check("municipal", "available", primary), f.check("regional", "not_applicable", VerificationEvidence{})))
	old := f.at.AddDate(0, -4, 0)
	if _, err := f.pool.Exec(f.ctx, "UPDATE retained_versions SET first_acquired_at=$1", old); err != nil {
		t.Fatal(err)
	}
	retention := NewRetention(f.pool, f.objects)
	result, err := retention.Cleanup(f.ctx, f.at)
	if err != nil || result.DeletedVersions != 0 {
		t.Fatal("ongoing primary lost its counterpart", result, err)
	}
	start, end := old, f.at.Add(-time.Hour)
	if _, err = f.store.PutTemporal(f.ctx, TemporalValue{EntityKind: "local_measure", EntityID: receipt.DomainRecordID, Meaning: "validity", Original: "synthetic explicit interval", Precision: "interval", Instant: &start, EndInstant: &end, EvidenceDocumentVersionID: primary.VersionID, CreatedAt: f.at}); err != nil {
		t.Fatal(err)
	}
	result, err = retention.Cleanup(f.ctx, f.at)
	if err != nil || result.DeletedVersions != 2 {
		t.Fatal("expired circular graph not purged", result, err)
	}
	var count int
	for _, table := range []string{"domain_verification_receipts", "domain_verification_dependencies", "domain_verification_projections", "domain_local_measures"} {
		if err = f.pool.QueryRow(f.ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != 0 {
			t.Fatal("expired evidence dangling", table, count, err)
		}
	}
}

func TestVerificationReferencedPDFRequiresOwnedCompleteOCR(t *testing.T) {
	f := newVerificationFixture(t)
	values := localVerificationValues()
	candidate := f.retain("platform", "1", values)
	url := "https://municipal.example/notices/referenced-act.pdf"
	version, err := f.docs.Retain(f.ctx, documents.Acquisition{ID: "synthetic-pdf", SourceID: "municipal", Configuration: 1, URL: url, Metadata: json.RawMessage(`{}`), Resources: []documents.Resource{{URL: url, Role: "original", Required: true, SourceID: "municipal", Configuration: 1, MediaType: "application/pdf", Bytes: []byte("%PDF-1.4\nsynthetic verification evidence\n%%EOF")}}})
	if err != nil {
		t.Fatal(err)
	}
	process := processing.New(f.pool)
	catalog, err := ocr.RegisterCatalog(f.ctx, process, "fixture", "model", f.at)
	if err != nil {
		t.Fatal(err)
	}
	run, err := process.StartRun(f.ctx, processing.RunRequest{IdempotencyKey: "verification-ocr", Stage: "ocr", Workload: "evaluation", ConfigurationVersionID: catalog.ConfigurationVersionID, DocumentVersionID: &version.ID, Subject: json.RawMessage(`{}`), CreatedAt: f.at})
	if err != nil {
		t.Fatal(err)
	}
	text := values["reference"] + " | " + values["edition"] + " | " + values["kind"] + " | " + values["subject"] + " | " + values["place"] + " | " + values["validity"]
	hash := strings.Repeat("a", 64)
	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO ocr_resource_results(run_id,document_version_id,resource_url,status,page_count,created_at) VALUES($1,$2,$3,'complete',1,$4)`, []any{run.ID, version.ID, url, f.at}},
		{`INSERT INTO ocr_page_results(run_id,page_number,document_version_id,resource_url,status,media_type,input_sha256,output_sha256,extracted_text,provider_response_id,returned_model,created_at) VALUES($1,1,$2,$3,'complete','image/png',$4,$4,$5,'synthetic','fixture',$6)`, []any{run.ID, version.ID, url, hash, text, f.at}},
	} {
		if _, err = f.pool.Exec(f.ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	f.at = time.Now().UTC()
	// Previously inserted facts can lack a supported interpretation event or
	// carry an unsupported temporal assumption. Reuse the measure, append the
	// supported temporal assessment, and preserve its earlier interpretation.
	if err = f.store.PutLocalMeasure(f.ctx, LocalMeasure{ID: "preexisting-pdf-measure", DocumentVersionID: version.ID, SourceID: "municipal", MunicipalityISTAT: "050004", Kind: "closure", Subject: values["subject"], Place: stringPointer(values["place"])}); err != nil {
		t.Fatal(err)
	}
	assumed := f.at
	if _, err = f.store.PutTemporal(f.ctx, TemporalValue{EntityKind: "local_measure", EntityID: "preexisting-pdf-measure", Meaning: "validity", Original: values["validity"], Precision: "instant", Instant: &assumed, EvidenceDocumentVersionID: version.ID, CreatedAt: f.at}); err != nil {
		t.Fatal(err)
	}
	primary := VerificationEvidence{SourceID: "municipal", VersionID: version.ID, Fields: map[string]EvidenceSelection{}}
	for name, value := range values {
		start := strings.Index(text, value)
		primary.Fields[name] = EvidenceSelection{ResourceURL: url, OCRRunID: run.ID, Page: 1, StartByte: start, EndByte: start + len(value), Locator: "synthetic referenced PDF page 1"}
	}
	r := f.verify(f.request("pdf", "local_measure", candidate, f.check("municipal", "available", primary), f.check("regional", "not_applicable", VerificationEvidence{})))
	if !r.Admitted || r.Checks[1].Fields["subject"].OCRRunID != run.ID {
		t.Fatal("retained referenced-act proof not admitted")
	}
	var precision, state string
	if r.DomainRecordID != "preexisting-pdf-measure" {
		t.Fatal("preexisting primary measure duplicated")
	}
	if err = f.pool.QueryRow(f.ctx, "SELECT precision FROM domain_temporal_values WHERE entity_id=$1 ORDER BY id DESC LIMIT 1", r.DomainRecordID).Scan(&precision); err != nil || precision != "conditional" {
		t.Fatal("unsupported temporal assumption inherited", precision, err)
	}
	if err = f.pool.QueryRow(f.ctx, "SELECT state FROM domain_interpretation_events WHERE local_measure_id=$1 ORDER BY id DESC LIMIT 1", r.DomainRecordID).Scan(&state); err != nil || state != "supported" {
		t.Fatal("reused projection kept unprocessed state", state, err)
	}
	bad := primary
	bad.Fields = map[string]EvidenceSelection{"kind": {ResourceURL: url, OCRRunID: run.ID, Page: 2, EndByte: 8, Locator: "absent page"}}
	if _, err = f.store.VerifyMultiSource(f.ctx, f.docs, f.request("absent-page", "local_measure", candidate, f.check("municipal", "available", bad), f.check("regional", "not_applicable", VerificationEvidence{})), f.at); !errors.Is(err, ErrVerificationEvidence) {
		t.Fatal("unreadable or foreign OCR page accepted", err)
	}
	native := primary
	native.Fields = map[string]EvidenceSelection{"kind": {ResourceURL: url, EndByte: 8, Locator: "PDF bytes are not evidence text"}}
	if _, err = f.store.VerifyMultiSource(f.ctx, f.docs, f.request("pdf-bytes", "local_measure", candidate, f.check("municipal", "available", native), f.check("regional", "not_applicable", VerificationEvidence{})), f.at); !errors.Is(err, ErrVerificationEvidence) {
		t.Fatal("raw PDF interpreted without completed text evidence", err)
	}
}
