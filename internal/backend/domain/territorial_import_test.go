package domain

import (
	"strings"
	"testing"
	"time"
)

func importFixture() (MunicipalityImport, []byte) {
	return MunicipalityImport{Region: "09", OfficialURL: "https://www.istat.it/fixture", Version: "test", VerifiedAt: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC), ExpectedCount: 2, Complete: true, CompletenessEvidence: "Attributed fixture of two municipalities"}, []byte("istat,comune,codice_regione,sigla\n050004,Calcinaia,09,PI\n048017,Firenze,09,FI\n")
}
func TestMunicipalityImportValidation(t *testing.T) {
	meta, body := importFixture()
	p, err := PreviewMunicipalities(body, meta)
	if err != nil || !p.Complete || len(p.Municipalities) != 2 {
		t.Fatal(p, err)
	}
	for _, bad := range []string{string(body) + "050004,Duplicate,09,PI\n", strings.Replace(string(body), "050004", "015146", 1), strings.Replace(string(body), "09,PI", "03,PI", 1), "istat,comune,codice_regione,sigla\n050004,Name\n"} {
		if _, err := PreviewMunicipalities([]byte(bad), meta); err == nil {
			t.Fatal("invalid import accepted")
		}
	}
	meta.ExpectedCount = 3
	p, err = PreviewMunicipalities(body, meta)
	if err != nil || p.Complete {
		t.Fatal("partial register labeled complete", err)
	}
	if _, err = PreviewMunicipalities(make([]byte, MaxMunicipalityImportBytes+1), meta); err == nil {
		t.Fatal("oversize accepted")
	}
}
