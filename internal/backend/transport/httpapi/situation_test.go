package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publicquery"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func summaryFixture() publicquery.Situation {
	quality := fixtureQuality()
	checked := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	quality.Updating.LastCompleteAt = &checked
	quality.Updating.DelayThresholdSeconds = 60
	evidence := []publicquery.Evidence{{DocumentID: "7", VersionID: "11", SourceURL: "https://example.test/bulletin.pdf"}}
	day := "2026-09-17"
	s := publicquery.Situation{Municipality: publicquery.Municipality{ISTAT: "050004", Name: "Calcinaia", Zones: []string{"A4"}}, History: fixtureHistory()}
	for i := 0; i < 7; i++ {
		level := "green"
		if i == 6 {
			level = "not_applicable"
		}
		s.RegionalProducts = append(s.RegionalProducts, publicquery.RegionalWarning{ID: fmt.Sprint(i), Product: "criticality", Risk: fmt.Sprint(i), Zone: "A4", Level: level, Status: "current", Validity: publicquery.Temporal{Precision: "date", Date: &day}, Quality: quality, Evidence: evidence})
	}
	s.LocalMeasures = []publicquery.Measure{{ID: "m1", Action: "closure", Subject: "parchi", Status: "undetermined", Validity: publicquery.Temporal{Precision: "unknown", Original: "data contraddittoria"}, Quality: quality, Evidence: evidence, NewerUninterpretedDocumentIDs: []string{"8"}}}
	s.Coverage = []publicquery.Coverage{{SourceID: "cfr", Product: "criticality", PublicState: "enabled", CoverageStatus: "pending", Quality: quality}, {SourceID: "town", Product: "municipal", PublicState: "enabled", CoverageStatus: "pending", Quality: quality}}
	s.DocumentsRequiringAttention = make([]publicquery.Document, 255)
	return s
}

func TestSituationConclusionsAndMissingChannel(t *testing.T) {
	s := summaryFixture()
	got := summarizeSituation(s)
	if !strings.Contains(got.ProcessedData.Summary, "6 livelli verdi e 1 rischio non applicabile") || !strings.Contains(got.ProcessedData.Summary, "1 con validità non determinabile") {
		t.Fatal(got.ProcessedData.Summary)
	}
	if len(got.ProcessedData.References) != 1 || len(got.ProcessedData.RegionalAlerts) != 7 {
		t.Fatal("evidence not deduplicated or risks lost")
	}
	if got.SourceSummaries.CittadinoInformato.Status != "not_collected" || got.SourceSummaries.Municipal.Products[0].CoverageStatus != "pending" {
		t.Fatal("missing channel or pending acceptance promoted")
	}
	if !strings.Contains(got.SourceSummaries.Municipal.Summary, "chiusura — parchi") || !strings.Contains(got.SourceSummaries.Municipal.Summary, "incompleta") {
		t.Fatal(got.SourceSummaries.Municipal)
	}
	limits := strings.Join(got.ProcessedData.LocalMeasures[0].Limitations, " ")
	if !strings.Contains(limits, "non considerare") || !strings.Contains(limits, "successive") {
		t.Fatal("operative uncertainty hidden", limits)
	}
	s.RegionalProducts[0].Level = "unknown"
	if !strings.Contains(summarizeSituation(s).ProcessedData.Summary, "solo in parte") {
		t.Fatal("unknown collapsed into green")
	}
	s.RegionalProducts[0].Level = "orange"
	s.RegionalProducts[0].Quality.Interpretation.State = "partial"
	if !strings.Contains(summarizeSituation(s).ProcessedData.Summary, "livello arancione") {
		t.Fatal("elevated level lost")
	}
	s.RegionalProducts[0].Status = "future"
	s.RegionalProducts[1].Quality.Provenance.State = "unverified"
	if !strings.Contains(summarizeSituation(s).ProcessedData.Summary, "solo in parte") {
		t.Fatal("unverified provenance ignored")
	}
}

func TestSituationPlatformIdentityAndNoAllClear(t *testing.T) {
	s := summaryFixture()
	c := s.Coverage[1]
	c.SourceID = "additional"
	c.Platform = "cittadino-informato"
	s.Coverage = append(s.Coverage, c)
	got := summarizeSituation(s)
	if len(got.SourceSummaries.Municipal.Products) != 1 || got.SourceSummaries.CittadinoInformato.Products[0].SourceID != "additional" {
		t.Fatal("platform mixed with municipality")
	}
	s = publicquery.Situation{}
	got = summarizeSituation(s)
	if !strings.Contains(got.ProcessedData.Summary, "non dimostra l'assenza") || got.SourceSummaries.Regional.Status != "not_collected" {
		t.Fatal("empty data implies all clear")
	}
	s = summaryFixture()
	s.Coverage[0].PublicState = "pending"
	s.Coverage[0].Quality.Interpretation.Evidence = []publicquery.Evidence{{DocumentID: "private", SourceURL: "https://private.invalid"}}
	got = summarizeSituation(s)
	if got.SourceSummaries.Regional.Status != "unavailable" || len(got.SourceSummaries.Regional.Products[0].ReferenceIDs) != 0 {
		t.Fatal("unpublished coverage evidence leaked")
	}
}

type summaryQueries struct {
	*publicQueriesFake
	value publicquery.Situation
}

func (q *summaryQueries) MunicipalitySituation(context.Context, publicquery.SituationQuery) (publicquery.Situation, error) {
	return q.value, nil
}

func TestSituationSummaryAPIAndMCPPinnedPagination(t *testing.T) {
	clock := time.Date(2026, 9, 17, 12, 0, 30, 0, time.UTC)
	queries := &summaryQueries{publicQueriesFake: &publicQueriesFake{}, value: summaryFixture()}
	queries.value.RegionalProducts[0].Product = "vigilance"
	queries.value.RegionalProducts[0].Level = "not_applicable"
	queries.value.RegionalProducts[0].Weather = &acquisition.VigilanceWeather{Phenomenon: "rainfall", GraphicalStatus: "depicted", RainfallBand: "0 - 10", TotalRainfallBand: "10 - 20", Unit: "mm", AmountScope: "area_average", TotalPeriod: "TOTALE: dalle 12 di oggi alle 24 di domani"}
	runtime := PublicRuntime{Limits: PublicLimits{Allowance: 120, Window: time.Minute, MaxPageSize: 100}, Views: &memoryViewStore{}, ViewLifetime: 30 * time.Minute, CursorKey: []byte(strings.Repeat("k", 32))}
	server := httptest.NewServer(handlerWithRuntime(nil, func() time.Time { return clock }, runtime, queries))
	defer server.Close()
	get := func(path string) publicResponse {
		t.Helper()
		r, err := server.Client().Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()
		var result publicResponse
		if err = json.NewDecoder(r.Body).Decode(&result); err != nil || r.StatusCode != 200 {
			t.Fatal(result, err)
		}
		return result
	}
	path := "/v1/municipalities/050004/situation?page_size=1"
	first := get(path)
	raw, _ := json.Marshal(first.Data)
	for _, noisy := range []string{"documents_requiring_attention", "regional_products", "quality", "sha256"} {
		if strings.Contains(string(raw), `"`+noisy+`"`) {
			t.Fatal("noisy raw fields exposed", noisy)
		}
	}
	var data situationData
	json.Unmarshal(raw, &data)
	if len(data.ProcessedData.RegionalAlerts) != 1 || first.Meta.NextCursor == nil || !strings.Contains(data.ProcessedData.Summary, "5 livelli verdi") {
		t.Fatal("summary computed from only first page", data)
	}
	if weather := data.ProcessedData.RegionalAlerts[0].Weather; weather == nil || weather.RainfallBand != "0 - 10" || weather.TotalRainfallBand != "10 - 20" {
		t.Fatal("vigilance amounts lost in situation", weather)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "summary-test", Version: "1"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: server.URL + "/mcp", HTTPClient: server.Client(), DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tool, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "get_municipality_situation", Arguments: map[string]any{"dataset_version": first.Meta.DatasetVersion}})
	if err != nil || tool.IsError || !reflect.DeepEqual(jsonValue(t, first), jsonValue(t, tool.StructuredContent)) {
		t.Fatal("API MCP summary mismatch", tool, err)
	}
	queries.value.RegionalProducts = nil
	clock = clock.Add(40 * time.Second)
	second := get("/v1/municipalities/050004/situation?cursor=" + url.QueryEscape(*first.Meta.NextCursor))
	raw, _ = json.Marshal(second.Data)
	json.Unmarshal(raw, &data)
	if data.ProcessedData.RegionalAlerts[0].ID != "1" || data.SourceSummaries.Regional.Products[0].Status != "delayed" || !strings.Contains(data.ProcessedData.Summary, "5 livelli verdi") {
		t.Fatal("snapshot or freshness lost", data)
	}
	cursor := second.Meta.NextCursor
	pages := 2
	for cursor != nil {
		page := get("/v1/municipalities/050004/situation?cursor=" + url.QueryEscape(*cursor))
		cursor = page.Meta.NextCursor
		pages++
		if pages > 7 {
			t.Fatal("attention documents drive pagination")
		}
	}
	if pages != 7 {
		t.Fatal("risks omitted", pages)
	}
}

func TestPartialOrangeSummaryPreservesRiskZoneAndRomeIntervals(t *testing.T) {
	value := summaryFixture()
	start := time.Date(2026, 10, 7, 15, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	zone := "Europe/Rome"
	warning := &value.RegionalProducts[0]
	warning.Level = "orange"
	warning.OfficialRiskLabel = "Temporali"
	warning.Quality.Interpretation.State = "partial"
	warning.Validity = publicquery.Temporal{Precision: "interval", Instant: &start, EndInstant: &end, Timezone: &zone}
	summary := summarizeSituation(value).ProcessedData.Summary
	for _, want := range []string{"Temporali: livello arancione nella zona A4", "dal 07/10/2026 17:00 al 08/10/2026 17:00 (Europe/Rome)"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("supported interval missing: %q", summary)
		}
	}
}
