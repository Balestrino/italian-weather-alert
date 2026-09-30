//go:build integration

package trialcost

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/observation"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
)

func TestTrialReportSeparatesCostsAndGatesBudgetProposal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := trialCostTestDB(t)
	for _, migrate := range []func(context.Context, *pgxpool.Pool) error{registry.Migrate, documents.Migrate, jobs.Migrate, processing.Migrate, acquisition.Migrate, observation.Migrate, Migrate, Migrate} {
		if err := migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Truncate(time.Second)
	started, through := now.Add(-8*24*time.Hour), now.Add(-24*time.Hour)
	evidence := registry.Evidence{URL: "https://evidence.example/trial", Locator: "retained fixture", ObservedAt: now}
	reg := registry.New(pool)
	if err := reg.CreateAuthority(ctx, registry.Authority{ID: "calcinaia", Name: "Comune", OfficialURL: "https://source.example"}); err != nil {
		t.Fatal(err)
	}
	if err := reg.CreateChannel(ctx, registry.Channel{ID: "municipal", PublisherID: "calcinaia", Platform: "fixture", URL: "https://source.example"}); err != nil {
		t.Fatal(err)
	}
	configuration := registry.Configuration{URL: "https://source.example", Sections: []string{"https://source.example/notices"}, AccessMethod: "fixture", Attribution: "Comune", Provenance: &evidence, Policy: registry.Policy{Evidence: &evidence, CollectionPermitted: true, RetentionPermitted: true}, CheckSeconds: 600, DelaySeconds: 1800, Limitations: []string{"fixture"}}
	for _, err := range []error{
		reg.CreateSource(ctx, registry.Source{ID: "calcinaia", AuthorityID: "calcinaia", ChannelID: "municipal", ProductID: "municipal", Territory: "050004"}, configuration, "operator"),
		reg.RecordPreview(ctx, "calcinaia", 1, "operator", evidence),
		reg.EnableCollection(ctx, "calcinaia", 1, "operator"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	campaign, err := observation.New(pool).Start(ctx, observation.StartRequest{ID: "trial", Actor: "operator", StartedAt: started, SourceIDs: []string{"calcinaia"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO observation_assessments(campaign_id,actor,through_at,status,evidence,report)
 VALUES($1,'reviewer',$2,'complete','{}','{}')`, campaign.ID, through); err != nil {
		t.Fatal(err)
	}
	proc := processing.New(pool)
	type catalog struct {
		model, config, price string
		rates                []processing.Rate
	}
	catalogs := []catalog{
		{model: "qwen", config: "qwen-classification", price: "qwen-price", rates: []processing.Rate{{Metric: "input_tokens", UnitSize: 1, PriceMicrounits: 1}, {Metric: "output_tokens", UnitSize: 1, PriceMicrounits: 2}}},
		{model: "ocr", config: "ocr-config", price: "ocr-price", rates: []processing.Rate{{Metric: "requests", UnitSize: 1, PriceMicrounits: 50}}},
		{model: "embedding", config: "embedding-config", price: "embedding-price", rates: []processing.Rate{{Metric: "input_tokens", UnitSize: 1, PriceMicrounits: 3}}},
	}
	priceStart, priceEnd := started.Add(-time.Hour), through.Add(time.Hour)
	for _, item := range catalogs {
		model := processing.ModelVersion{ID: item.model, Provider: "fixture", Model: item.model, Revision: "1", Capabilities: json.RawMessage(`{}`), CreatedAt: priceStart}
		if err = proc.RegisterModel(ctx, model); err != nil {
			t.Fatal(err)
		}
		stage := "classification"
		if item.model == "ocr" {
			stage = "ocr"
		}
		if item.model == "embedding" {
			stage = "embedding"
		}
		if err = proc.RegisterConfiguration(ctx, processing.ConfigurationVersion{ID: item.config, Name: item.config, Stage: stage, Revision: "1", ModelVersionID: &model.ID, LogicVersion: "fixture", Settings: json.RawMessage(`{}`), CreatedAt: priceStart}); err != nil {
			t.Fatal(err)
		}
		if err = proc.RegisterPrice(ctx, processing.PriceVersion{ID: item.price, ModelVersionID: model.ID, Currency: "EUR", ProvenanceURL: "https://pricing.example/" + item.model, ObservedAt: priceStart, EffectiveFrom: &priceStart, EffectiveThrough: &priceEnd, Details: json.RawMessage(`{"plan":"fixture"}`), Rates: item.rates, CreatedAt: priceStart}); err != nil {
			t.Fatal(err)
		}
	}
	source := "calcinaia"
	addRun := func(key, workload, stage, config, price string, attempts []processing.Usage) {
		t.Helper()
		run, runErr := proc.StartRun(ctx, processing.RunRequest{IdempotencyKey: key, Workload: workload, Stage: stage, ConfigurationVersionID: config, SourceID: &source, Subject: json.RawMessage(`{}`), CreatedAt: started.Add(time.Hour)})
		if runErr != nil {
			t.Fatal(runErr)
		}
		for index, usage := range attempts {
			at := started.Add(time.Duration(index+1) * time.Hour)
			attempt, startErr := proc.StartAttempt(ctx, processing.AttemptStart{RunID: run.ID, StartedAt: at})
			if startErr != nil {
				t.Fatal(startErr)
			}
			outcome := "succeeded"
			errorCode := ""
			if len(attempts) > 1 && index == 0 {
				outcome = "failed"
				errorCode = "provider_temporary"
			}
			if _, finishErr := proc.FinishAttempt(ctx, processing.AttemptFinish{RunID: run.ID, Number: attempt.Number, FinishedAt: at.Add(time.Second), Outcome: outcome, Usage: usage, PriceVersionID: &price, ErrorCode: errorCode}); finishErr != nil {
				t.Fatal(finishErr)
			}
		}
	}
	n := func(value int64) *int64 { return &value }
	addRun("bootstrap", "bootstrap", "classification", "qwen-classification", "qwen-price", []processing.Usage{{Status: "reported", InputTokens: n(100), OutputTokens: n(10)}})
	addRun("ordinary-qwen", "ordinary", "classification", "qwen-classification", "qwen-price", []processing.Usage{{Status: "reported", InputTokens: n(20), OutputTokens: n(5)}, {Status: "reported", InputTokens: n(30), OutputTokens: n(5)}})
	addRun("ordinary-ocr", "ordinary", "ocr", "ocr-config", "ocr-price", []processing.Usage{{Status: "reported", OtherUnits: map[string]int64{"requests": 2}}})
	addRun("ordinary-embedding", "ordinary", "embedding", "embedding-config", "embedding-price", []processing.Usage{{Status: "reported", InputTokens: n(10)}})
	store := New(pool)
	record := func(category, workload, component, status string, quantity, unitSize, price *int64) Input {
		t.Helper()
		input := Input{Category: category, Workload: workload, Component: component, Status: status, Evidence: evidence, Notes: "retained trial invoice or explicit applicability review"}
		if status == "reported" {
			input.Basis, input.Quantity, input.Unit, input.UnitSize, input.PriceMicrounits, input.Currency = "rate", quantity, "trial-unit", unitSize, price, "EUR"
			input.EffectiveFrom, input.EffectiveThrough = &priceStart, &priceEnd
		}
		result, recordErr := store.RecordInput(ctx, campaign.ID, "operator", input)
		if recordErr != nil {
			t.Fatal(recordErr)
		}
		return result
	}
	one := int64(1)
	record("hosting", "bootstrap", "vm-bootstrap", "not_applicable", nil, nil, nil)
	record("storage", "bootstrap", "bootstrap-space", "reported", n(1), &one, n(20))
	record("hosting", "ordinary", "vm-hours", "reported", n(168), &one, n(10))
	incomplete, err := store.Report(ctx, campaign.ID, through)
	if err != nil || incomplete.BudgetReady || !contains(incomplete.MissingMetrics, "infrastructure_metric_missing:storage:ordinary") {
		t.Fatalf("missing storage was not explicit: %#v %v", incomplete, err)
	}
	if _, err = store.Propose(ctx, campaign.ID, "operator", through, evidence); !errors.Is(err, ErrIncomplete) {
		t.Fatalf("incomplete report proposed a budget: %v", err)
	}
	last := record("storage", "ordinary", "object-gb-days", "reported", n(14), &one, n(5))
	report, err := store.Report(ctx, campaign.ID, through)
	if err != nil || !report.BudgetReady || len(report.MissingMetrics) != 0 || len(report.Currencies) != 1 || report.Currencies[0] != "EUR" || len(report.Prices) != 3 {
		t.Fatalf("complete report: %#v %v", report, err)
	}
	stages := map[string]Breakdown{}
	for _, item := range report.Breakdown {
		stages[item.Workload+":"+item.Stage] = item
	}
	if stages["ordinary:classification"].Retries != 1 || *stages["ordinary:classification"].KnownRetryCostMicrounits != 40 || stages["ordinary:ocr"].OtherUnits["requests"] != 2 || *stages["ordinary:embedding"].InputTokens != 10 {
		t.Fatalf("processing dimensions collapsed: %#v", stages)
	}
	proposal, err := store.Propose(ctx, campaign.ID, "operator", through, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if proposal.BootstrapCostMicrounits != 140 || proposal.OrdinaryTrialCostMicrounits != 1950 || proposal.OrdinaryMonthlyCostMicrounits != 8358 || proposal.RetryTrialCostMicrounits != 40 || proposal.Currency != "EUR" {
		t.Fatalf("unexpected proposal: %#v", proposal)
	}
	if _, err = store.Propose(ctx, campaign.ID, "operator", through, evidence); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate proposal was not rejected: %v", err)
	}
	proposals, err := store.Proposals(ctx, campaign.ID)
	if err != nil || len(proposals) != 1 || proposals[0].Report.BudgetReady != true {
		t.Fatalf("proposal history: %#v %v", proposals, err)
	}
	if _, err = pool.Exec(ctx, `UPDATE trial_cost_inputs SET notes='tampered' WHERE campaign_id=$1 AND category=$2 AND workload=$3 AND component=$4 AND revision=$5`, campaign.ID, last.Category, last.Workload, last.Component, last.Revision); err == nil {
		t.Fatal("cost evidence was mutable")
	}
	unknownAt := through.Add(30 * time.Minute)
	unknownRun, err := proc.StartRun(ctx, processing.RunRequest{IdempotencyKey: "unknown-after-proposal", Workload: "ordinary", Stage: "classification", ConfigurationVersionID: "qwen-classification", SourceID: &source, Subject: json.RawMessage(`{}`), CreatedAt: unknownAt})
	if err != nil {
		t.Fatal(err)
	}
	unknownAttempt, err := proc.StartAttempt(ctx, processing.AttemptStart{RunID: unknownRun.ID, StartedAt: unknownAt})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = proc.FinishAttempt(ctx, processing.AttemptFinish{RunID: unknownRun.ID, Number: unknownAttempt.Number, FinishedAt: unknownAt.Add(time.Second), Outcome: "failed", Usage: processing.Usage{Status: "unavailable"}, ErrorCode: "provider_usage_missing"}); err != nil {
		t.Fatal(err)
	}
	unknownReport, err := store.Report(ctx, campaign.ID, through.Add(time.Hour))
	if err != nil || !contains(unknownReport.MissingMetrics, "usage_unavailable:ordinary:classification") || !contains(unknownReport.MissingMetrics, "provider_cost_unavailable:ordinary:classification") {
		t.Fatalf("unknown provider metrics were not explicit: %#v %v", unknownReport.MissingMetrics, err)
	}
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func trialCostTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	name := "iwa-trial-cost-test-" + hex.EncodeToString(random[:6])
	password := hex.EncodeToString(random)
	passwordFile := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		t.Helper()
		commandCtx, commandCancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer commandCancel()
		output, err := exec.CommandContext(commandCtx, "docker", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("isolated PostgreSQL operation failed: %v: %s", err, strings.ReplaceAll(string(output), password, "[redacted]"))
		}
		return strings.TrimSpace(string(output))
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	binding := listener.Addr().String() + ":5432"
	_ = listener.Close()
	run("run", "--pull=never", "--detach", "--name", name, "--publish", binding, "--mount", "type=bind,src="+passwordFile+",dst=/run/secrets/password,readonly", "--env", "POSTGRES_PASSWORD_FILE=/run/secrets/password", "--env", "POSTGRES_USER=iwa", "--env", "POSTGRES_DB=iwa", "postgres:16-alpine")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", name).Run() })
	address := run("port", name, "5432/tcp")
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatal(err)
	}
	portNumber, _ := strconv.Atoi(port)
	configuration, err := pgxpool.ParseConfig("postgres://iwa@localhost/iwa?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	configuration.ConnConfig.Host, configuration.ConnConfig.Port, configuration.ConnConfig.Password = host, uint16(portNumber), password
	pool, err := pgxpool.NewWithConfig(context.Background(), configuration)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	readyCtx, readyCancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer readyCancel()
	for pool.Ping(readyCtx) != nil {
		if readyCtx.Err() != nil {
			t.Fatal("test PostgreSQL did not become ready")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return pool
}
