package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/backoffice"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/publicquery"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type publicQueriesFake struct {
	mu            sync.Mutex
	calls         map[string]int
	coverageCount int
}

func (f *publicQueriesFake) called(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.calls == nil {
		f.calls = map[string]int{}
	}
	f.calls[name]++
}

func (f *publicQueriesFake) DiscoverMunicipalities(_ context.Context, query publicquery.DiscoveryQuery) (publicquery.DiscoveryResult, error) {
	f.called("discover_municipalities")
	if query.ISTAT != "050004" || !fixedQueryTime(query.QueryTime) {
		return publicquery.DiscoveryResult{}, publicquery.ErrInvalidParameters
	}
	return publicquery.DiscoveryResult{Municipalities: []publicquery.Municipality{{ISTAT: "050004", Name: "Calcinaia", Zones: []string{"A4"}, MappingLimitations: []string{}, LocalCoverage: "enabled"}}, History: fixtureHistory()}, nil
}

func (f *publicQueriesFake) MunicipalitySituation(_ context.Context, query publicquery.SituationQuery) (publicquery.Situation, error) {
	f.called("get_municipality_situation")
	if query.MunicipalityISTAT != "050004" || !fixedQueryTime(query.QueryTime) {
		return publicquery.Situation{}, publicquery.ErrInvalidParameters
	}
	return publicquery.Situation{Municipality: publicquery.Municipality{ISTAT: "050004", Name: "Calcinaia", Zones: []string{"A4"}, MappingLimitations: []string{}, LocalCoverage: "enabled"}, LocalMeasures: []publicquery.Measure{}, OperationalPhases: []publicquery.OperationalPhase{}, RegionalProducts: []publicquery.RegionalWarning{}, DocumentsRequiringAttention: []publicquery.Document{}, Coverage: []publicquery.Coverage{}, History: fixtureHistory()}, nil
}

func (f *publicQueriesFake) Search(_ context.Context, query publicquery.SearchQuery) (publicquery.SearchResult, error) {
	f.called("search_alerts")
	if query.Kind != "regional" || query.Product != "vigilance" || !fixedQueryTime(query.QueryTime) {
		return publicquery.SearchResult{}, publicquery.ErrInvalidParameters
	}
	return publicquery.SearchResult{Kind: "regional", Documents: []publicquery.Document{}, Measures: []publicquery.Measure{}, Regional: []publicquery.RegionalWarning{}, History: fixtureHistory()}, nil
}

func (f *publicQueriesFake) Document(_ context.Context, query publicquery.DocumentQuery) (publicquery.DocumentResult, error) {
	f.called("get_document")
	if query.DocumentID != "7" || !query.IncludeVersions || !fixedQueryTime(query.QueryTime) {
		return publicquery.DocumentResult{}, publicquery.ErrInvalidParameters
	}
	document := publicquery.Document{ID: "7", VersionID: "11", SourceID: "calcinaia", Publisher: "Comune di Calcinaia", Kind: "ordinance", OfficialURL: "https://example.test/7", SHA256: strings.Repeat("a", 64), Publication: publicquery.Temporal{Precision: "unknown"}, Modification: publicquery.Temporal{Precision: "unknown"}, AcquiredAt: time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC), Quality: fixtureQuality(), Attachments: []publicquery.Attachment{}}
	return publicquery.DocumentResult{Document: document, Versions: []publicquery.Document{document}, History: fixtureHistory()}, nil
}

func (f *publicQueriesFake) Coverage(_ context.Context, query publicquery.CoverageQuery) (publicquery.CoverageResult, error) {
	f.called("get_source_coverage")
	if query.MunicipalityISTAT != "050004" || !fixedQueryTime(query.QueryTime) {
		return publicquery.CoverageResult{}, publicquery.ErrInvalidParameters
	}
	count := f.coverageCount
	if count == 0 {
		count = 1
	}
	sources := make([]publicquery.Coverage, count)
	for index := range sources {
		sources[index] = publicquery.Coverage{SourceID: fmt.Sprintf("source-%03d", index), Product: "municipal", Territory: "050004", DeclaredSections: []string{"https://example.test/albo"}, PublicState: "enabled", CoverageStatus: "accepted_with_limitations", CoverageLimitations: []string{"fixture scope only"}, Quality: fixtureQuality()}
	}
	return publicquery.CoverageResult{Sources: sources, History: fixtureHistory()}, nil
}

func BenchmarkPublicCoverageHTTP(b *testing.B) {
	fake := &publicQueriesFake{coverageCount: 100}
	limits := PublicLimits{Allowance: 1_000_000, Window: time.Minute, MaxPageSize: 100}
	testServer := httptest.NewServer(handlerWithLimits(nil, time.Now, limits, fake))
	defer testServer.Close()
	client := testServer.Client()
	endpoint := testServer.URL + "/v1/sources/coverage?municipality_istat=050004&evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z"
	probe, err := client.Get(endpoint)
	if err != nil {
		b.Fatal(err)
	}
	probeBody, err := io.ReadAll(probe.Body)
	probe.Body.Close()
	if err != nil || probe.StatusCode != http.StatusOK {
		b.Fatalf("probe status=%d err=%v", probe.StatusCode, err)
	}
	responseBytes := len(probeBody)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			response, err := client.Get(endpoint)
			if err != nil {
				b.Error(err)
				return
			}
			_, err = io.Copy(io.Discard, response.Body)
			response.Body.Close()
			if err != nil || response.StatusCode != http.StatusOK {
				b.Errorf("response status=%d err=%v", response.StatusCode, err)
				return
			}
		}
	})
	b.ReportMetric(float64(responseBytes), "response-B")
}

func fixtureHistory() publicquery.History {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	return publicquery.History{Start: &start, Gaps: []string{}, Limitations: []string{"retained history starts after source history"}}
}

func fixtureQuality() publicquery.Quality {
	return publicquery.Quality{Provenance: publicquery.Dimension{State: "verified", Limitations: []string{}, Evidence: []publicquery.Evidence{}}, Interpretation: publicquery.Dimension{State: "supported", Limitations: []string{}, Evidence: []publicquery.Evidence{}}, Updating: publicquery.Updating{State: "ok", DelayThresholdSeconds: 3600, Limitations: []string{}}}
}

func fixedQueryTime(value publicquery.QueryTime) bool {
	return value.EvaluationTime.Equal(time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)) && value.KnownAt.Equal(time.Date(2026, 9, 17, 11, 30, 0, 0, time.UTC))
}

func TestPublicAPIAndRealMCPClientAreEquivalent(t *testing.T) {
	fixedNow := time.Date(2026, 9, 17, 12, 5, 0, 0, time.UTC)
	fake := &publicQueriesFake{}
	runtime := PublicRuntime{Limits: PublicLimits{Allowance: 120, Window: time.Minute, MaxPageSize: 100}, Views: &memoryViewStore{}, ViewLifetime: 30 * time.Minute, CursorKey: []byte(strings.Repeat("k", 32))}
	testServer := httptest.NewServer(handlerWithRuntime(nil, func() time.Time { return fixedNow }, runtime, fake))
	defer testServer.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "iwa-contract-test", Version: "1.0.0"}, &mcp.ClientOptions{})
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: testServer.URL + "/mcp", HTTPClient: testServer.Client(), DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if got := session.InitializeResult().ProtocolVersion; got != publicMCPProtocolVersion {
		t.Fatalf("protocol version = %q", got)
	}
	if instructions := session.InitializeResult().Instructions; !strings.Contains(instructions, "120 requests") || !strings.Contains(instructions, "Maximum page size is 100") {
		t.Fatalf("limits not published through MCP: %q", instructions)
	}

	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	wantTools := []string{"discover_municipalities", "get_document", "get_municipality_situation", "get_source_coverage", "search_alerts"}
	var gotTools []string
	for _, tool := range listed.Tools {
		gotTools = append(gotTools, tool.Name)
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint || !tool.Annotations.IdempotentHint || tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint || tool.Annotations.OpenWorldHint == nil || *tool.Annotations.OpenWorldHint {
			t.Fatalf("unsafe annotations for %s: %#v", tool.Name, tool.Annotations)
		}
		if tool.InputSchema == nil || tool.OutputSchema == nil {
			t.Fatalf("contract schema missing for %s", tool.Name)
		}
	}
	sort.Strings(gotTools)
	if !reflect.DeepEqual(gotTools, wantTools) {
		t.Fatalf("tools = %v", gotTools)
	}

	commonQuery := "evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z"
	commonArguments := map[string]any{"evaluation_time": "2026-09-17T12:00:00Z", "known_at": "2026-09-17T11:30:00Z"}
	tests := []struct {
		name, path string
		arguments  map[string]any
	}{
		{"discover_municipalities", "/v1/municipalities?istat=050004&" + commonQuery, with(commonArguments, "istat", "050004")},
		{"get_municipality_situation", "/v1/municipalities/050004/situation?" + commonQuery, with(commonArguments, "municipality_istat", "050004")},
		{"search_alerts", "/v1/search?kind=regional&product=vigilance&" + commonQuery, with(with(commonArguments, "kind", "regional"), "product", "vigilance")},
		{"get_document", "/v1/documents/7?include_versions=true&" + commonQuery, with(with(commonArguments, "document_id", "7"), "include_versions", true)},
		{"get_source_coverage", "/v1/sources/coverage?municipality_istat=050004&" + commonQuery, with(commonArguments, "municipality_istat", "050004")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := testServer.Client().Get(testServer.URL + test.path)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			apiBody, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != http.StatusOK {
				t.Fatalf("API status %d: %s", response.StatusCode, apiBody)
			}
			var apiValue any
			if err := json.Unmarshal(apiBody, &apiValue); err != nil {
				t.Fatal(err)
			}

			apiObject, ok := apiValue.(map[string]any)
			if !ok {
				t.Fatalf("API envelope type = %T", apiValue)
			}
			meta, ok := apiObject["meta"].(map[string]any)
			if !ok || meta["dataset_version"] == nil {
				t.Fatalf("API view metadata missing: %#v", apiObject)
			}
			arguments := with(test.arguments, "dataset_version", meta["dataset_version"])
			toolResult, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: test.name, Arguments: arguments})
			if err != nil {
				t.Fatal(err)
			}
			if toolResult.IsError {
				t.Fatalf("MCP application error: %#v", toolResult.Content)
			}
			accounting, ok := toolResult.Meta["iwa.dev/rateLimit"].(map[string]any)
			if !ok || accounting["partition"] != "client_ip" || accounting["max_page_size"] == nil {
				t.Fatalf("MCP accounting metadata missing: %#v", toolResult.Meta)
			}
			mcpValue := jsonValue(t, toolResult.StructuredContent)
			if !reflect.DeepEqual(apiValue, mcpValue) {
				apiJSON, _ := json.Marshal(apiValue)
				mcpJSON, _ := json.Marshal(mcpValue)
				t.Fatalf("API/MCP mismatch\nAPI %s\nMCP %s", apiJSON, mcpJSON)
			}
		})
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	for _, name := range wantTools {
		if fake.calls[name] != 1 {
			t.Errorf("%s calls = %d, want one snapshot query shared by API and MCP", name, fake.calls[name])
		}
	}
}

func TestPublicTransportErrorsAndIsolation(t *testing.T) {
	fixedNow := time.Date(2026, 9, 17, 12, 5, 0, 0, time.UTC)
	fake := &publicQueriesFake{}
	public := handler(nil, func() time.Time { return fixedNow }, fake)

	request := httptest.NewRequest(http.MethodGet, "/v1/search?kind=regional&unknown=value", nil)
	recorder := httptest.NewRecorder()
	public.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"code":"invalid_parameters"`) {
		t.Fatalf("unexpected invalid-parameter response: %d %s", recorder.Code, recorder.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Mcp-Protocol-Version", "2025-06-18")
	recorder = httptest.NewRecorder()
	public.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("legacy protocol status = %d", recorder.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "http://iwa.test/mcp", strings.NewReader(`{}`))
	request.Header.Set("Origin", "https://attacker.example")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Mcp-Protocol-Version", publicMCPProtocolVersion)
	recorder = httptest.NewRecorder()
	public.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("cross-origin status = %d", recorder.Code)
	}

	admin := backoffice.Handler(nil)
	for _, path := range []string{"/v1/municipalities", "/v1/search", "/mcp"} {
		recorder = httptest.NewRecorder()
		admin.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8081"+path, nil))
		if recorder.Code != http.StatusNotFound {
			t.Errorf("admin listener exposes %s: %d", path, recorder.Code)
		}
	}
}

func with(source map[string]any, key string, value any) map[string]any {
	result := make(map[string]any, len(source)+1)
	for existingKey, existingValue := range source {
		result[existingKey] = existingValue
	}
	result[key] = value
	return result
}

func jsonValue(t *testing.T, value any) any {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func TestPublicApplicationErrorsRemainStructuredInMCP(t *testing.T) {
	fixedNow := time.Date(2026, 9, 17, 12, 5, 0, 0, time.UTC)
	testServer := httptest.NewServer(handler(nil, func() time.Time { return fixedNow }, &publicQueriesFake{}))
	defer testServer.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "iwa-error-test", Version: "1.0.0"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: testServer.URL + "/mcp", HTTPClient: testServer.Client(), DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "search_alerts", Arguments: map[string]any{"kind": "invalid"}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("application error was reported as success")
	}
	value := jsonValue(t, result.StructuredContent)
	encoded, _ := json.Marshal(value)
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !strings.Contains(string(encoded), `"code":"invalid_parameters"`) || !ok || !strings.Contains(textContent.Text, "invalid_parameters") {
		t.Fatalf("structured error missing: %s %#v", encoded, result.Content)
	}
}
