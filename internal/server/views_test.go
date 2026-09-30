package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
	"github.com/Balestrino/italian-weather-alert/internal/publicview"
)

type memoryViewStore struct {
	mu                sync.Mutex
	views             map[string]publicview.View
	next              int
	createErr, getErr error
}

func (s *memoryViewStore) Create(_ context.Context, request publicview.Create) (publicview.View, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.createErr != nil {
		return publicview.View{}, s.createErr
	}
	s.next++
	view := publicview.View{ID: fmt.Sprintf("view-%d", s.next), Operation: request.Operation, RequestHash: request.RequestHash, Request: append(json.RawMessage(nil), request.Request...), Data: append(json.RawMessage(nil), request.Data...), History: append(json.RawMessage(nil), request.History...), EvaluationTime: request.EvaluationTime, KnownAt: request.KnownAt, CreatedAt: request.CreatedAt, ExpiresAt: request.ExpiresAt}
	if s.views == nil {
		s.views = map[string]publicview.View{}
	}
	s.views[view.ID] = view
	return view, nil
}

func (s *memoryViewStore) Get(_ context.Context, id string) (publicview.View, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.getErr != nil {
		return publicview.View{}, s.getErr
	}
	view, ok := s.views[id]
	if !ok {
		return publicview.View{}, publicview.ErrNotFound
	}
	return view, nil
}

type mutableCoverageQueries struct {
	*publicQueriesFake
	mu        sync.Mutex
	prefix    string
	calls     int
	checkedAt time.Time
}

func (q *mutableCoverageQueries) Coverage(_ context.Context, request publicquery.CoverageQuery) (publicquery.CoverageResult, error) {
	if request.MunicipalityISTAT != "050004" || !fixedQueryTime(request.QueryTime) {
		return publicquery.CoverageResult{}, publicquery.ErrInvalidParameters
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.calls++
	quality := fixtureQuality()
	checkedAt := q.checkedAt
	quality.Updating.LastAttemptAt = &checkedAt
	quality.Updating.LastCompleteAt = &checkedAt
	quality.Updating.DelayThresholdSeconds = 60
	values := make([]publicquery.Coverage, 3)
	for index := range values {
		values[index] = publicquery.Coverage{SourceID: fmt.Sprintf("%s-%d", q.prefix, index), Product: "municipal", Territory: "050004", DeclaredSections: []string{"https://example.test/source"}, PublicState: "enabled", CoverageStatus: "accepted_declared_scope", CoverageLimitations: []string{}, Quality: quality}
	}
	return publicquery.CoverageResult{Sources: values, History: fixtureHistory()}, nil
}

func (q *mutableCoverageQueries) setPrefix(value string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.prefix = value
}

func (q *mutableCoverageQueries) callCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.calls
}

type coverageEnvelope struct {
	Data  []publicquery.Coverage `json:"data"`
	Meta  *publicMeta            `json:"meta"`
	Error *publicError           `json:"error"`
}

func TestPersistentViewPaginationAndFreshness(t *testing.T) {
	clock := time.Date(2026, 9, 17, 12, 0, 30, 0, time.UTC)
	queries := &mutableCoverageQueries{publicQueriesFake: &publicQueriesFake{}, prefix: "old", checkedAt: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)}
	views := &memoryViewStore{}
	runtime := PublicRuntime{Limits: PublicLimits{Allowance: 120, Window: time.Minute, MaxPageSize: 10}, Views: views, ViewLifetime: 30 * time.Minute, CursorKey: []byte(strings.Repeat("k", 32))}
	testServer := httptest.NewServer(handlerWithRuntime("public", nil, func() time.Time { return clock }, runtime, queries))
	defer testServer.Close()
	baseQuery := "municipality_istat=050004&page_size=1&evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z"

	first := getCoverage(t, testServer.Client(), testServer.URL+"/v1/sources/coverage?"+baseQuery)
	if len(first.Data) != 1 || first.Data[0].SourceID != "old-0" || first.Data[0].Quality.Updating.State != "ok" || first.Meta == nil || first.Meta.NextCursor == nil || first.Meta.DatasetVersion != "view-1" || first.Meta.ViewExpiresAt.Sub(clock) != 30*time.Minute {
		t.Fatalf("unexpected first page: %#v", first)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "iwa-view-test", Version: "1.0.0"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: testServer.URL + "/mcp", HTTPClient: testServer.Client(), DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tool, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "get_source_coverage", Arguments: map[string]any{"municipality_istat": "050004", "page_size": 1, "evaluation_time": "2026-09-17T12:00:00Z", "known_at": "2026-09-17T11:30:00Z", "dataset_version": first.Meta.DatasetVersion}})
	if err != nil || tool.IsError {
		t.Fatalf("MCP fixed-view request: %#v %v", tool, err)
	}
	var firstValue any
	encodedFirst, _ := json.Marshal(publicResponse{Data: first.Data, Meta: first.Meta})
	if json.Unmarshal(encodedFirst, &firstValue) != nil || !reflect.DeepEqual(firstValue, jsonValue(t, tool.StructuredContent)) {
		t.Fatalf("API/MCP fixed-view mismatch: API=%s MCP=%#v", encodedFirst, tool.StructuredContent)
	}

	queries.setPrefix("new")
	clock = time.Date(2026, 9, 17, 12, 1, 1, 0, time.UTC)
	secondURL := testServer.URL + "/v1/sources/coverage?" + baseQuery + "&cursor=" + url.QueryEscape(*first.Meta.NextCursor)
	second := getCoverage(t, testServer.Client(), secondURL)
	if len(second.Data) != 1 || second.Data[0].SourceID != "old-1" || second.Data[0].Quality.Updating.State != "delayed" || second.Meta.DatasetVersion != first.Meta.DatasetVersion || second.Meta.EvaluationTime != first.Meta.EvaluationTime || queries.callCount() != 1 {
		t.Fatalf("second page mixed data or freshness: %#v calls=%d", second, queries.callCount())
	}
	if second.Data[0].Quality.Updating.LastCompleteAt == nil || !second.Data[0].Quality.Updating.LastCompleteAt.Equal(queries.checkedAt) || second.Meta.ServedAt != clock {
		t.Fatalf("cached freshness timestamps are untruthful: %#v", second)
	}

	tampered := *first.Meta.NextCursor
	if strings.HasSuffix(tampered, "A") {
		tampered = strings.TrimSuffix(tampered, "A") + "B"
	} else {
		tampered += "A"
	}
	assertPublicError(t, testServer.Client(), testServer.URL+"/v1/sources/coverage?"+baseQuery+"&cursor="+url.QueryEscape(tampered), http.StatusBadRequest, "cursor_mismatch")
	assertPublicError(t, testServer.Client(), secondURL+"&source_id=changed", http.StatusBadRequest, "cursor_mismatch")

	clock = time.Date(2026, 9, 17, 12, 31, 0, 0, time.UTC)
	assertPublicError(t, testServer.Client(), secondURL, http.StatusGone, "cursor_expired")
	fresh := getCoverage(t, testServer.Client(), testServer.URL+"/v1/sources/coverage?"+baseQuery)
	if fresh.Meta.DatasetVersion == first.Meta.DatasetVersion || fresh.Data[0].SourceID != "new-0" || queries.callCount() != 2 {
		t.Fatalf("restart after expiry did not create a fresh view: %#v", fresh)
	}
}

func TestViewStorageFailuresAreUnavailable(t *testing.T) {
	clock := time.Date(2026, 9, 17, 12, 0, 30, 0, time.UTC)
	queries := &mutableCoverageQueries{publicQueriesFake: &publicQueriesFake{}, prefix: "old", checkedAt: clock}
	views := &memoryViewStore{createErr: errors.New("storage down")}
	runtime := PublicRuntime{Limits: PublicLimits{Allowance: 20, Window: time.Minute, MaxPageSize: 10}, Views: views, ViewLifetime: 30 * time.Minute, CursorKey: []byte(strings.Repeat("k", 32))}
	handler := handlerWithRuntime("public", nil, func() time.Time { return clock }, runtime, queries)
	request := httptest.NewRequest(http.MethodGet, "/v1/sources/coverage?municipality_istat=050004&page_size=1&evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z", nil)
	request.RemoteAddr = "192.0.2.1:1000"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "service_unavailable") {
		t.Fatalf("create failure became data: %d %s", response.Code, response.Body.String())
	}

	views.createErr = nil
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("view recovery failed: %d %s", response.Code, response.Body.String())
	}
	var created coverageEnvelope
	if json.Unmarshal(response.Body.Bytes(), &created) != nil {
		t.Fatal("invalid recovery response")
	}
	views.getErr = errors.New("storage down")
	request = httptest.NewRequest(http.MethodGet, "/v1/sources/coverage?dataset_version="+url.QueryEscape(created.Meta.DatasetVersion), nil)
	request.RemoteAddr = "192.0.2.2:1000"
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "service_unavailable") {
		t.Fatalf("load failure became empty data: %d %s", response.Code, response.Body.String())
	}
}

func TestCompoundAndDocumentPaginationPositions(t *testing.T) {
	situation := publicquery.Situation{
		LocalMeasures:     []publicquery.Measure{{ID: "m1"}, {ID: "m2"}},
		OperationalPhases: []publicquery.OperationalPhase{{ID: "p1"}, {ID: "p2"}},
		RegionalProducts:  []publicquery.RegionalWarning{{ID: "r1"}},
		Coverage:          []publicquery.Coverage{{SourceID: "s1"}, {SourceID: "s2"}},
	}
	firstAny, positions, more, err := paginateData(situation, map[string]int{}, 1)
	first := firstAny.(publicquery.Situation)
	if err != nil || !more || first.LocalMeasures[0].ID != "m1" || first.OperationalPhases[0].ID != "p1" || first.RegionalProducts[0].ID != "r1" || first.Coverage[0].SourceID != "s1" {
		t.Fatalf("first compound page: %#v %#v %v", first, positions, err)
	}
	secondAny, _, more, err := paginateData(situation, positions, 1)
	second := secondAny.(publicquery.Situation)
	if err != nil || more || second.LocalMeasures[0].ID != "m2" || second.OperationalPhases[0].ID != "p2" || len(second.RegionalProducts) != 0 || second.Coverage[0].SourceID != "s2" {
		t.Fatalf("second compound page: %#v more=%v err=%v", second, more, err)
	}

	documents := documentData{Document: publicquery.Document{ID: "stable"}, Versions: []publicquery.Document{{VersionID: "v1"}, {VersionID: "v2"}}}
	firstDocumentAny, documentPositions, more, err := paginateData(documents, map[string]int{}, 1)
	firstDocument := firstDocumentAny.(documentData)
	if err != nil || !more || firstDocument.Document.ID != "stable" || firstDocument.Versions[0].VersionID != "v1" {
		t.Fatalf("first document page: %#v", firstDocument)
	}
	secondDocumentAny, _, more, err := paginateData(documents, documentPositions, 1)
	secondDocument := secondDocumentAny.(documentData)
	if err != nil || more || secondDocument.Document.ID != "stable" || secondDocument.Versions[0].VersionID != "v2" {
		t.Fatalf("second document page: %#v", secondDocument)
	}
}

func getCoverage(t *testing.T, client *http.Client, endpoint string) coverageEnvelope {
	t.Helper()
	response, err := client.Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var result coverageEnvelope
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status=%d result=%#v", endpoint, response.StatusCode, result)
	}
	return result
}

func assertPublicError(t *testing.T, client *http.Client, endpoint string, status int, code string) {
	t.Helper()
	response, err := client.Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var result coverageEnvelope
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != status || result.Error == nil || result.Error.Code != code {
		t.Fatalf("GET %s: status=%d error=%#v", endpoint, response.StatusCode, result.Error)
	}
}
