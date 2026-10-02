//go:build integration

package domain

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupMunicipalityLifecycle(t *testing.T, p *pgxpool.Pool) (*Store, string) {
	t.Helper()
	ctx := context.Background()
	s := New(p)
	meta, body := importFixture()
	preview, _ := PreviewMunicipalities(body, meta)
	id, err := s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 0, "importer")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ConfigureRegion(ctx, "09", 1, RegionConfiguration{MunicipalityDataset: id, Profiles: []string{"municipal-html", "toscana-cfr"}}, "configurator"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SetRegionEnabled(ctx, "09", 2, true, "operator"); err != nil {
		t.Fatal(err)
	}
	reg := registry.New(p)
	for _, err = range []error{
		reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "Fixture", OfficialURL: "https://example.test"}),
		reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "web", URL: "https://example.test"}),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	evidence := registry.Evidence{URL: "https://example.test", Locator: "synthetic preview", ObservedAt: time.Now()}
	for source, product := range map[string]string{"050004": "municipal", "048017": "municipal", "regional": "criticality"} {
		where := source
		if source == "regional" {
			where = "Toscana"
		}
		cfg := registry.Configuration{URL: "https://example.test", Sections: []string{"https://example.test"}, AccessMethod: "html", Attribution: "fixture", Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true, Evidence: &evidence}}
		if err = reg.CreateSource(ctx, registry.Source{ID: source, AuthorityID: "a", ChannelID: "c", ProductID: product, Territory: where}, cfg, "operator"); err != nil {
			t.Fatal(err)
		}
		if err = reg.RecordPreview(ctx, source, 1, "fixture", evidence); err != nil {
			t.Fatal(err)
		}
		if err = reg.EnableCollection(ctx, source, 1, "operator"); err != nil {
			t.Fatal(err)
		}
	}
	return s, id
}

func TestMunicipalityLifecycleGatesAndPreservesChoices(t *testing.T) {
	ctx := context.Background()
	p := territorialTestDB(t)
	s, _ := setupMunicipalityLifecycle(t, p)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{jobs.Migrate, acquisition.Migrate} {
		if err := m(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	state, err := s.MunicipalityState(ctx, "09", "050004")
	if err != nil || state.Enabled || state.Revision != 0 || state.Eligible || state.BlockedBy != "municipality_disabled" {
		t.Fatal(state, err)
	}
	if err = territory.CheckExecution(ctx, p, "050004"); !errors.Is(err, territory.ErrDisabled) {
		t.Fatal("disabled municipal execution", err)
	}
	for _, istat := range []string{"050004", "048017"} {
		if _, err = s.SetMunicipalityEnabled(ctx, "09", istat, 0, true, "operator"); err != nil {
			t.Fatal(err)
		}
	}
	queue := jobs.New(p)
	now := time.Now()
	add := func(key, source string) {
		t.Helper()
		_, err := queue.Enqueue(ctx, jobs.EnqueueRequest{Queue: "municipal-test", Kind: "test", IdempotencyKey: key, Payload: json.RawMessage(`{"source_id":"` + source + `"}`), MaxAttempts: 1, RetryBase: time.Second, AvailableAt: now})
		if err != nil {
			t.Fatal(err)
		}
	}
	add("running", "050004")
	running, err := queue.Claim(ctx, "municipal-test", "test", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	add("blocked", "050004")
	add("other", "048017")
	if _, err = s.SetMunicipalityEnabled(ctx, "09", "050004", 1, false, "operator"); err != nil {
		t.Fatal(err)
	}
	if err = queue.Complete(ctx, running, jobs.Result{Payload: json.RawMessage(`{}`)}, time.Now()); err != nil {
		t.Fatal("admitted work must finish", err)
	}
	for _, source := range []string{"regional", "048017"} {
		if err = territory.CheckExecution(ctx, p, source); err != nil {
			t.Fatal(source, err)
		}
	}
	if err = registry.New(p).CheckTerritorialExecution(ctx, "050004", 1); !errors.Is(err, territory.ErrDisabled) {
		t.Fatal("manual launch bypass", err)
	}
	claim, err := queue.Claim(ctx, "municipal-test", "test", time.Now(), time.Minute)
	if err != nil || claim.IdempotencyKey != "other" {
		t.Fatal(claim, err)
	}
	if _, err = queue.Claim(ctx, "municipal-test", "test", time.Now(), time.Minute); !errors.Is(err, jobs.ErrNoJob) {
		t.Fatal("disabled job claimed", err)
	}
	schedule := acquisition.NewScheduleStore(p)
	if err = schedule.SyncEnabled(ctx, now); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		check, err := schedule.ClaimDue(ctx, "test", time.Now(), time.Minute)
		if err != nil || check.SourceID == "050004" {
			t.Fatal(check, err)
		}
	}
	if _, err = schedule.ClaimDue(ctx, "test", time.Now(), time.Minute); !errors.Is(err, acquisition.ErrNoDueCheck) {
		t.Fatal("disabled acquisition claimed", err)
	}
	if _, err = s.SetRegionEnabled(ctx, "09", 3, false, "operator"); err != nil {
		t.Fatal(err)
	}
	state, err = s.MunicipalityState(ctx, "09", "048017")
	if err != nil || !state.Enabled || state.Eligible || state.BlockedBy != "region_disabled" {
		t.Fatal(state, err)
	}
	if _, err = s.SetRegionEnabled(ctx, "09", 4, true, "operator"); err != nil {
		t.Fatal(err)
	}
	if err = territory.CheckExecution(ctx, p, "050004"); !errors.Is(err, territory.ErrDisabled) {
		t.Fatal("parent changed child choice", err)
	}
	if _, err = s.SetMunicipalityEnabled(ctx, "09", "050004", 2, true, "operator"); err != nil {
		t.Fatal(err)
	}
	claim, err = queue.Claim(ctx, "municipal-test", "test", time.Now(), time.Minute)
	if err != nil || claim.IdempotencyKey != "blocked" {
		t.Fatal(claim, err)
	}
	var count int
	if err = p.QueryRow(ctx, `SELECT count(*) FROM processing_jobs`).Scan(&count); err != nil || count != 3 {
		t.Fatal("enablement scheduled catch-up", count, err)
	}
	st, err := registry.New(p).State(ctx, "050004")
	if err != nil || !st.CollectionEnabled || st.PublicEnabled {
		t.Fatal("source flags changed", st, err)
	}
	if _, err = s.SetMunicipalityEnabled(ctx, "09", "050004", 1, false, "stale"); !errors.Is(err, ErrConflict) {
		t.Fatal("stale edit", err)
	}
	if _, err = s.SetMunicipalityEnabled(ctx, "03", "050004", 0, true, "foreign"); !errors.Is(err, ErrTerritoryNotFound) {
		t.Fatal("foreign membership", err)
	}
	if _, err = s.SetMunicipalityEnabled(ctx, "09", "050004", 3, false, ""); !errors.Is(err, ErrInvalid) {
		t.Fatal("missing actor", err)
	}
	if err = p.QueryRow(ctx, `SELECT count(*) FROM territorial_municipality_events WHERE istat='050004' AND actor='operator'`).Scan(&count); err != nil || count != 3 {
		t.Fatal("audit", count, err)
	}
	if _, err = p.Exec(ctx, `DELETE FROM territorial_municipality_events`); err == nil {
		t.Fatal("mutable audit")
	}
}

func TestMunicipalityRegisterReplacementAndAdmissionLock(t *testing.T) {
	ctx := context.Background()
	p := territorialTestDB(t)
	s, _ := setupMunicipalityLifecycle(t, p)
	if _, err := s.SetMunicipalityEnabled(ctx, "09", "050004", 0, true, "operator"); err != nil {
		t.Fatal(err)
	}
	tx, err := p.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err = territory.CheckExecution(ctx, tx, "050004"); err != nil {
		t.Fatal(err)
	}
	blocked, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	_, err = s.SetMunicipalityEnabled(blocked, "09", "050004", 1, false, "operator")
	cancel()
	if err == nil {
		t.Fatal("disablement did not wait for admitted transaction")
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	state, err := s.MunicipalityState(ctx, "09", "050004")
	if err != nil || !state.Enabled || state.Revision != 1 {
		t.Fatal(state, err)
	}
	meta, body := importFixture()
	meta.Version = "replacement"
	body = []byte("istat,comune,codice_regione,sigla\n050004,Calcinaia,09,PI\n049001,Bibbona,09,LI\n")
	preview, _ := PreviewMunicipalities(body, meta)
	if _, err = s.AdoptMunicipalities(ctx, body, meta, preview.SHA256, 3, "importer"); err != nil {
		t.Fatal(err)
	}
	state, err = s.MunicipalityState(ctx, "09", "050004")
	if err != nil || !state.Enabled || state.Historical {
		t.Fatal("lost retained choice", state, err)
	}
	state, err = s.MunicipalityState(ctx, "09", "049001")
	if err != nil || state.Enabled || state.Revision != 0 {
		t.Fatal("new municipality enabled", state, err)
	}
	state, err = s.MunicipalityState(ctx, "09", "048017")
	if err != nil || !state.Historical {
		t.Fatal("retired history lost", state, err)
	}
	if _, err = s.SetMunicipalityEnabled(ctx, "09", "048017", 0, true, "operator"); !errors.Is(err, ErrTerritoryNotFound) {
		t.Fatal("retired enabled", err)
	}
	if _, err = s.SetMunicipalityEnabled(ctx, "09", "048017", 0, false, "operator"); err != nil {
		t.Fatal("retired disablement", err)
	}
}

func TestMunicipalityMigrationPreservesExistingEligibility(t *testing.T) {
	ctx := context.Background()
	p := domainTestDB(t)
	for _, m := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, Migrate, MigrateGeography, migrateTerritorialCatalog, migrateTerritorialScope, migrateTerritorialGates, migrateTerritorialGateHardening} {
		if err := m(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	_, _ = setupMunicipalityLifecycle(t, p)
	if err := migrateMunicipalityLifecycle(ctx, p); err != nil {
		t.Fatal(err)
	}
	s := New(p)
	state, err := s.MunicipalityState(ctx, "09", "050004")
	if err != nil || !state.Enabled || !state.Eligible || state.Revision != 1 {
		t.Fatal(state, err)
	}
	if _, err = s.SetMunicipalityEnabled(ctx, "09", "050004", 1, false, "operator"); err != nil {
		t.Fatal(err)
	}
	if err = MigrateTerritories(ctx, p); err != nil {
		t.Fatal(err)
	}
	state, err = s.MunicipalityState(ctx, "09", "050004")
	if err != nil || state.Enabled || state.Revision != 2 {
		t.Fatal("migration replaced operator choice", state, err)
	}
	var actor string
	if err = p.QueryRow(ctx, `SELECT actor FROM territorial_municipality_events WHERE istat='048017' AND kind='migration'`).Scan(&actor); err != nil || actor != "municipality-lifecycle-migration" {
		t.Fatal(actor, err)
	}
}
