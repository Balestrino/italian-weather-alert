package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/publiccopy"
)

type copyAccessFake struct {
	mu      sync.Mutex
	allowed bool
	readErr error
	lastDoc string
	lastVer string
}

func (f *copyAccessFake) Link(_ context.Context, documentID, versionID string) (*string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.allowed {
		return nil, nil
	}
	value := "https://alerts.example/v1/documents/" + url.PathEscape(documentID) + "/versions/" + url.PathEscape(versionID) + "/content"
	return &value, nil
}

func (f *copyAccessFake) Read(_ context.Context, documentID, versionID string) (publiccopy.Content, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastDoc, f.lastVer = documentID, versionID
	if f.readErr != nil {
		return publiccopy.Content{}, f.readErr
	}
	return publiccopy.Content{Bytes: []byte("exact retained version"), MediaType: "application/pdf", SHA256: strings.Repeat("a", 64)}, nil
}

func (f *copyAccessFake) setAllowed(value bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.allowed = value
}

func TestRetainedCopyEndpointIsExactAndMediated(t *testing.T) {
	copyAccess := &copyAccessFake{allowed: true}
	handler := handlerWithRuntime("public", nil, time.Now, PublicRuntime{Limits: PublicLimits{Allowance: 10, Window: time.Minute, MaxPageSize: 10}, Copies: copyAccess}, &publicQueriesFake{})
	request := httptest.NewRequest(http.MethodGet, "/v1/documents/7/versions/11/content", nil)
	request.RemoteAddr = "192.0.2.1:1000"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "exact retained version" || response.Header().Get("Content-Type") != "application/pdf" || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Content-Type-Options") != "nosniff" || response.Header().Get("ETag") != `"sha256:`+strings.Repeat("a", 64)+`"` {
		t.Fatalf("unexpected copy response: %d %#v %q", response.Code, response.Header(), response.Body.String())
	}
	if copyAccess.lastDoc != "7" || copyAccess.lastVer != "11" || response.Header().Get("RateLimit") == "" {
		t.Fatalf("copy was not exact or budgeted: doc=%q version=%q headers=%#v", copyAccess.lastDoc, copyAccess.lastVer, response.Header())
	}
}

func TestRetainedCopyErrorsUseContractEnvelope(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   string
	}{
		{publiccopy.ErrDisabled, http.StatusForbidden, "copy_access_disabled"},
		{publiccopy.ErrRestricted, http.StatusForbidden, "copy_access_restricted"},
		{publiccopy.ErrUnknown, http.StatusNotFound, "unknown_identifier"},
		{publiccopy.ErrStorage, http.StatusServiceUnavailable, "service_unavailable"},
	}
	for _, test := range tests {
		t.Run(test.code+test.err.Error(), func(t *testing.T) {
			copyAccess := &copyAccessFake{readErr: test.err}
			handler := handlerWithRuntime("public", nil, time.Now, PublicRuntime{Limits: PublicLimits{Allowance: 10, Window: time.Minute, MaxPageSize: 10}, Copies: copyAccess}, &publicQueriesFake{})
			request := httptest.NewRequest(http.MethodGet, "/v1/documents/7/versions/11/content", nil)
			request.RemoteAddr = "192.0.2.1:1000"
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			var envelope publicResponse
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil || response.Code != test.status || envelope.Error == nil || envelope.Error.Code != test.code {
				t.Fatalf("status=%d envelope=%#v decode=%v", response.Code, envelope, err)
			}
		})
	}
}

func TestStoredViewReevaluatesCurrentCopyPolicy(t *testing.T) {
	fixedNow := time.Date(2026, 9, 17, 12, 5, 0, 0, time.UTC)
	queries := &publicQueriesFake{}
	copies := &copyAccessFake{allowed: true}
	runtime := PublicRuntime{Limits: PublicLimits{Allowance: 20, Window: time.Minute, MaxPageSize: 10}, Views: &memoryViewStore{}, ViewLifetime: 30 * time.Minute, CursorKey: []byte(strings.Repeat("k", 32)), Copies: copies}
	testServer := httptest.NewServer(handlerWithRuntime("public", nil, func() time.Time { return fixedNow }, runtime, queries))
	defer testServer.Close()
	endpoint := testServer.URL + "/v1/documents/7?include_versions=true&evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z&page_size=1"
	response, err := testServer.Client().Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	var first struct {
		Data documentData `json:"data"`
		Meta publicMeta   `json:"meta"`
	}
	if json.Unmarshal(body, &first) != nil || first.Data.Document.CopyURL == nil || first.Data.Versions[0].CopyURL == nil {
		t.Fatalf("enabled link missing: %s", body)
	}
	copies.setAllowed(false)
	response, err = testServer.Client().Get(testServer.URL + "/v1/documents/7?dataset_version=" + url.QueryEscape(first.Meta.DatasetVersion))
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(response.Body)
	response.Body.Close()
	var restricted struct {
		Data documentData `json:"data"`
	}
	if json.Unmarshal(body, &restricted) != nil || restricted.Data.Document.CopyURL != nil || restricted.Data.Versions[0].CopyURL != nil {
		t.Fatalf("old view retained revoked link: %s", body)
	}
	queries.mu.Lock()
	defer queries.mu.Unlock()
	if queries.calls["get_document"] != 1 {
		t.Fatalf("policy recheck reran domain query: %#v", queries.calls)
	}
}

func TestNilCopyAccessIsDisabled(t *testing.T) {
	handler := handlerWithRuntime("public", nil, time.Now, PublicRuntime{Limits: PublicLimits{Allowance: 2, Window: time.Minute, MaxPageSize: 10}}, &publicQueriesFake{})
	request := httptest.NewRequest(http.MethodGet, "/v1/documents/7/versions/11/content", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "copy_access_disabled") {
		t.Fatalf("nil copy policy was not disabled: %d %s", response.Code, response.Body.String())
	}
}
