//go:build integration

package domain

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

// Exercises a real pg_dump/pg_restore boundary. It never connects to deployment data.
func TestTerritorialMigrationBackupRestoreRehearsal(t *testing.T) {
	ctx := context.Background()
	p, container := domainTestDBContainer(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, Migrate, MigrateGeography} {
		if err := m(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	// Recreate the schema boundary preceding the additive public guard migration.
	if _, err := p.Exec(ctx, `DROP VIEW registry_public_sources; DROP FUNCTION registry_source_in_public_scope(text); DROP FUNCTION registry_dataset_in_public_scope(text); DELETE FROM iwa_migrations WHERE name='053_public_territorial_scope'`); err != nil {
		t.Fatal(err)
	}
	s := New(p)
	dataset := Dataset{ID: "legacy-rehearsal", Kind: DatasetMunicipalities, Authority: "ISTAT", Publisher: "ISTAT", OfficialURL: "https://example.test/istat", VersionLabel: "fixture", SourceSHA256: strings.Repeat("a", 64), VerifiedAt: time.Now(), Applicability: "verified", Limitations: []string{}, Usable: true}
	if err := s.RegisterMunicipalities(ctx, dataset, readMunicipalities(t)); err != nil {
		t.Fatal(err)
	}
	if err := s.SelectDataset(ctx, DatasetMunicipalities, dataset.ID, "fixture", time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Exec(ctx, `INSERT INTO registry_authorities VALUES('legacy','Legacy','https://example.test'); INSERT INTO registry_channels VALUES('legacy','legacy','web','https://example.test',false); INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES('legacy','legacy','legacy','municipal','050004'); INSERT INTO registry_configurations(source_id,revision,body,actor) VALUES('legacy',1,'{}','fixture'); UPDATE registry_sources SET latest_revision=1,active_revision=1,collection_enabled=true,public_enabled=true WHERE id='legacy'`); err != nil {
		t.Fatal(err)
	}
	fingerprint := func(db *pgxpool.Pool) string {
		t.Helper()
		var out string
		err := db.QueryRow(ctx, `SELECT jsonb_build_object('sources',(SELECT jsonb_agg(to_jsonb(s) ORDER BY id) FROM registry_sources s),'configurations',(SELECT jsonb_agg(to_jsonb(c) ORDER BY source_id,revision) FROM registry_configurations c),'selections',(SELECT jsonb_agg(to_jsonb(g) ORDER BY id) FROM geography_dataset_selections g),'municipalities',(SELECT jsonb_agg(to_jsonb(m) ORDER BY dataset_id,istat) FROM geography_municipalities m))::text`).Scan(&out)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	before := fingerprint(p)
	dump, err := exec.Command("docker", "exec", container, "pg_dump", "-U", "iwa", "-d", "iwa", "-Fc").Output()
	if err != nil {
		t.Fatal("dump", err)
	}
	migrate := func(db *pgxpool.Pool) {
		t.Helper()
		for i := 0; i < 2; i++ {
			if err := registry.Migrate(ctx, db); err != nil {
				t.Fatal(err)
			}
			if err := MigrateTerritories(ctx, db); err != nil {
				t.Fatal(err)
			}
			if err := New(db).BackfillToscana(ctx); err != nil {
				t.Fatal(err)
			}
		}
	}
	migrate(p)
	if fingerprint(p) != before {
		t.Fatal("legacy state changed during migration")
	}
	report, err := s.TerritoryMigrationReport(ctx)
	if err != nil || len(report.UnassociatedSources) != 0 || len(report.UnassociatedDatasets) != 0 {
		t.Fatal(report, err)
	}
	meta, _ := importFixture()
	meta.Region = "03"
	meta.ExpectedCount = 1
	body := []byte("istat,comune,codice_regione,sigla\n015146,Milano,03,MI\n")
	preview, _ := PreviewMunicipalities(body, meta)
	id, err := s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 0, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ConfigureRegion(ctx, "03", 1, RegionConfiguration{MunicipalityDataset: id, Profiles: []string{"municipal-html"}}, "fixture"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SetRegionEnabled(ctx, "03", 2, true, "fixture"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SetRegionEnabled(ctx, "03", 3, false, "rollback-fixture"); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Exec(ctx, `CREATE DATABASE iwa_restore`); err != nil {
		t.Fatal(err)
	}
	restore := exec.Command("docker", "exec", "-i", container, "pg_restore", "-U", "iwa", "-d", "iwa_restore", "--exit-on-error")
	restore.Stdin = bytes.NewReader(dump)
	if out, err := restore.CombinedOutput(); err != nil {
		t.Fatalf("restore: %v %s", err, out)
	}
	cfg := p.Config()
	cfg.ConnConfig.Database = "iwa_restore"
	restored, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	if fingerprint(restored) != before {
		t.Fatal("restored legacy snapshot differs")
	}
	var absent bool
	if err = restored.QueryRow(ctx, `SELECT to_regclass('territorial_regions') IS NULL`).Scan(&absent); err != nil || !absent {
		t.Fatal("rollback retained new territory", err)
	}
	migrate(restored)
	if fingerprint(restored) != before {
		t.Fatal("migration after restore changed legacy state")
	}
	t.Logf("PostgreSQL dump (%d bytes), additive migration twice, second-region enable/disable, restore to separate database and remigration passed; source/configuration/geography fingerprints unchanged", len(dump))
}
