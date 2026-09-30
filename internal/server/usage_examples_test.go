package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	contract "github.com/Balestrino/italian-weather-alert/docs/prerequisiti-mvp"
)

type publishedUsageExamples struct {
	Kind            string `json:"kind"`
	ProtocolVersion string `json:"protocol_version"`
	Examples        []struct {
		ID           string         `json:"id"`
		Description  string         `json:"description"`
		APIPath      string         `json:"api_path"`
		MCPTool      string         `json:"mcp_tool"`
		MCPArguments map[string]any `json:"mcp_arguments"`
	} `json:"examples"`
}

func TestPublishedUsageExamplesAgainstLocalService(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "deploy", "public-usage-examples.json"))
	if err != nil {
		t.Fatal(err)
	}
	var published publishedUsageExamples
	if err := json.Unmarshal(raw, &published); err != nil {
		t.Fatal(err)
	}
	if published.Kind != "synthetic_local_service_usage_examples" || published.ProtocolVersion != publicMCPProtocolVersion {
		t.Fatalf("unexpected published example metadata: %#v", published)
	}

	wantIDs := []string{"coverage_history", "discovery", "document", "search", "situation"}
	wantTools := []string{"discover_municipalities", "get_document", "get_municipality_situation", "get_source_coverage", "search_alerts"}
	var gotIDs, exampleTools []string
	for _, example := range published.Examples {
		gotIDs = append(gotIDs, example.ID)
		exampleTools = append(exampleTools, example.MCPTool)
		if example.Description == "" || !strings.HasPrefix(example.APIPath, "/v1/") || len(example.MCPArguments) == 0 {
			t.Fatalf("incomplete published example: %#v", example)
		}
	}
	sort.Strings(gotIDs)
	sort.Strings(exampleTools)
	if !reflect.DeepEqual(gotIDs, wantIDs) || !reflect.DeepEqual(exampleTools, wantTools) {
		t.Fatalf("published coverage ids=%v tools=%v", gotIDs, exampleTools)
	}

	fixedNow := time.Date(2026, 9, 17, 12, 5, 0, 0, time.UTC)
	runtime := PublicRuntime{
		Limits:       PublicLimits{Allowance: 120, Window: time.Minute, MaxPageSize: 100},
		Views:        &memoryViewStore{},
		ViewLifetime: 30 * time.Minute,
		CursorKey:    []byte(strings.Repeat("e", 32)),
	}
	testServer := httptest.NewServer(handlerWithRuntime("public", nil, func() time.Time { return fixedNow }, runtime, &publicQueriesFake{}))
	defer testServer.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "iwa-published-examples-test", Version: "1.0.0"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: testServer.URL + "/mcp", HTTPClient: testServer.Client(), DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var listedTools []string
	for _, tool := range listed.Tools {
		listedTools = append(listedTools, tool.Name)
	}
	sort.Strings(listedTools)
	if !reflect.DeepEqual(listedTools, wantTools) {
		t.Fatalf("public tools = %v", listedTools)
	}

	for _, example := range published.Examples {
		t.Run(example.ID, func(t *testing.T) {
			response, err := testServer.Client().Get(testServer.URL + example.APIPath)
			if err != nil {
				t.Fatal(err)
			}
			body, readErr := io.ReadAll(response.Body)
			response.Body.Close()
			if readErr != nil {
				t.Fatal(readErr)
			}
			if response.StatusCode != http.StatusOK {
				t.Fatalf("API status %d: %s", response.StatusCode, body)
			}
			var apiValue map[string]any
			if err := json.Unmarshal(body, &apiValue); err != nil {
				t.Fatal(err)
			}
			meta, ok := apiValue["meta"].(map[string]any)
			if !ok || meta["dataset_version"] == nil || meta["history_gaps"] == nil {
				t.Fatalf("view/history metadata missing: %#v", apiValue)
			}
			arguments := make(map[string]any, len(example.MCPArguments)+1)
			for key, value := range example.MCPArguments {
				arguments[key] = value
			}
			arguments["dataset_version"] = meta["dataset_version"]
			result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: example.MCPTool, Arguments: arguments})
			if err != nil {
				t.Fatal(err)
			}
			if result.IsError {
				t.Fatalf("MCP application error: %#v", result.Content)
			}
			mcpValue := jsonValue(t, result.StructuredContent)
			if !reflect.DeepEqual(any(apiValue), mcpValue) {
				t.Fatalf("published API and MCP examples returned different views")
			}
			assertNoGeneratedSafetyFields(t, apiValue)
			encoded, _ := json.Marshal(apiValue)
			if strings.Contains(strings.ToLower(string(encoded)), "dpc-internal") {
				t.Fatalf("internal DPC fixture escaped through public response: %s", encoded)
			}
		})
	}

	for _, path := range []string{"/admin/status", "/admin/sources", "/config", "/debug/vars"} {
		response, err := testServer.Client().Get(testServer.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusNotFound {
			t.Fatalf("public listener exposed %s with status %d", path, response.StatusCode)
		}
	}

	var descriptors any
	if err := json.Unmarshal(contract.MCPTools, &descriptors); err != nil {
		t.Fatal(err)
	}
	assertNoGeneratedSafetyFields(t, descriptors)
}

func assertNoGeneratedSafetyFields(t *testing.T, value any) {
	t.Helper()
	forbidden := map[string]bool{
		"advice": true, "recommendation": true, "safety_advice": true,
		"safety_judgment": true, "generated_judgment": true, "safety_status": true,
		"guidance": true, "instruction": true, "instructions": true, "generated_guidance": true,
	}
	var walk func(any)
	walk = func(current any) {
		switch item := current.(type) {
		case map[string]any:
			for key, child := range item {
				if forbidden[strings.ToLower(key)] {
					t.Errorf("public contract exposes generated safety field %q", key)
				}
				walk(child)
			}
		case []any:
			for _, child := range item {
				walk(child)
			}
		}
	}
	walk(value)
}
