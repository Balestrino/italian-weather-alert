package domain

import (
	"testing"
	"time"
)

func TestVerificationTextAcceptsParameterizedMediaTypes(t *testing.T) {
	for _, test := range []struct{ media, body, pointer, want string }{
		{"text/html; charset=utf-8", "<p>Criticità <b>arancione</b></p>", "", "Criticità arancione"},
		{"application/json; charset=utf-8", `{"claim":"<b>Temporali</b>"}`, "/claim", "Temporali"},
		{"text/plain; charset=utf-8", "Emissione ufficiale", "", "Emissione ufficiale"},
	} {
		text, err := verificationText(test.media, []byte(test.body), test.pointer)
		if err != nil || text != test.want {
			t.Fatalf("%s: text=%q error=%v", test.media, text, err)
		}
	}
	for _, media := range []string{"text/html; invalid", "application/pdf"} {
		if _, err := verificationText(media, []byte("untrusted"), ""); err == nil {
			t.Fatalf("invalid or unsupported media accepted: %s", media)
		}
	}
}

func evidenceFields(role string, values map[string]string) ChannelEvidence {
	result := ChannelEvidence{Role: role, SourceID: role, VersionID: 1, Complete: true, Fields: map[string]VerifiedField{}}
	for name, value := range values {
		result.Fields[name] = VerifiedField{Value: normalizedField(name, value), Passage: value}
	}
	return result
}

func TestVerificationComparabilityAndFieldSupport(t *testing.T) {
	local := map[string]string{"reference": "Ordinanza 42/2026", "edition": "03/10/2026", "kind": "chiusura", "subject": "Sottopasso", "validity": "fino a nuovo ordine"}
	platform := evidenceFields("platform", local)
	primary := evidenceFields("municipal", local)
	outcome, _, fields := compareEvidence("local_measure", platform, primary)
	if !admitsFields("local_measure", ChannelVerification{Outcome: outcome, ComparedFields: fields}, platform) {
		t.Fatal("comparable municipal fields rejected")
	}
	delete(primary.Fields, "validity")
	outcome, _, fields = compareEvidence("local_measure", platform, primary)
	if outcome != "corroborated" {
		t.Fatal("missing optional validity undid supported core")
	}
	for _, f := range fields {
		if f.Field == "validity" && f.Outcome != "missing_evidence" {
			t.Fatal("unsupported validity confirmed")
		}
	}
	primary.Fields["kind"] = VerifiedField{Value: "reopening", Passage: "riapertura"}
	outcome, _, _ = compareEvidence("local_measure", platform, primary)
	if outcome != "conflict" {
		t.Fatal("comparable disagreement not retained")
	}
	primary.Fields["edition"] = VerifiedField{Value: "2026-10-04", Passage: "04/10/2026"}
	outcome, _, _ = compareEvidence("local_measure", platform, primary)
	if outcome != "non_comparable" {
		t.Fatal("different edition called a conflict")
	}
	regional := map[string]string{"product": "criticità", "risk": "vento", "zone": "H2", "issuance": "2026-10-03T12:00:00+02:00", "validity": "2026-10-03T13:00:00+02:00/2026-10-04T13:00:00+02:00", "level": "giallo"}
	a, b := evidenceFields("platform", regional), evidenceFields("regional", regional)
	b.Fields["issuance"] = VerifiedField{Value: "2026-10-03T10:00:00Z", Passage: "2026-10-03T10:00:00Z"}
	outcome, _, _ = compareEvidence("regional_record", a, b)
	if outcome != "corroborated" {
		t.Fatal("equivalent absolute issuance rejected")
	}
	b.Fields["level"] = VerifiedField{Value: "unknown", Passage: "sconosciuto"}
	outcome, _, _ = compareEvidence("regional_record", a, b)
	if outcome != "non_comparable" {
		t.Fatal("unknown treated as confirmed color or conflict")
	}
	for _, role := range []*ChannelEvidence{&a, &b} {
		role.Fields["validity"] = VerifiedField{Value: "oggi", Passage: "oggi"}
	}
	outcome, _, _ = compareEvidence("regional_record", a, b)
	if outcome != "non_comparable" {
		t.Fatal("relative display date admitted as operative validity")
	}
}

func TestVerificationEvidenceTextAndTemporalPrecision(t *testing.T) {
	text, err := verificationText("application/json", []byte(`{"items":[{"content":"<script>rosso</script><p>Giallo &amp; vento</p>"}],"a/b":"sintetico"}`), "/items/0/content")
	if err != nil || text != "Giallo & vento" {
		t.Fatalf("text: %q %v", text, err)
	}
	for _, pointer := range []string{"/items/00/content", "/items/-1/content", "/items/0", "/a~2b", "/absent"} {
		if _, err = verificationText("application/json", []byte(`{"items":[{"content":"value"}],"a~2b":"value"}`), pointer); err == nil {
			t.Fatalf("invalid selector accepted: %s", pointer)
		}
	}
	if verificationWordBounds("greenish", 0, 5) || verificationWordBounds("preverde", 3, 8) || !verificationWordBounds("colore verde.", 7, 12) {
		t.Fatal("partial enum word accepted")
	}
	at := time.Now().UTC()
	date := verificationTemporal("local_measure", "fixture", 1, VerifiedField{Passage: "03/10/2026"}, at)
	if date.Precision != "date" || date.Instant != nil || date.Date == nil {
		t.Fatal("date precision invented an instant")
	}
	interval := verificationTemporal("regional_record", "fixture", 1, VerifiedField{Passage: "2026-10-03T12:00:00+02:00/2026-10-04T12:00:00+02:00"}, at)
	if interval.Precision != "interval" || interval.Instant.Hour() != 10 || interval.EndInstant.Hour() != 10 {
		t.Fatal("explicit offset not retained")
	}
	unknown := verificationTemporal("local_measure", "fixture", 1, VerifiedField{Passage: "domani pomeriggio"}, at)
	if unknown.Precision != "unknown" || unknown.Instant != nil {
		t.Fatal("ambiguous validity invented an operative date")
	}
}
