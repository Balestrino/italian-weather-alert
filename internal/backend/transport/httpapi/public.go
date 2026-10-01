package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/platform/httpserver"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	contract "github.com/Balestrino/italian-weather-alert/api/public"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publiccopy"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publicquery"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const publicMCPProtocolVersion = "2026-07-28"

var errPaginationUnavailable = errors.New("result requires unavailable pagination")

// PublicQueries is deliberately read-only. Public transports cannot reach
// collection, inference, source configuration, or administrative operations.
type PublicQueries interface {
	DiscoverMunicipalities(context.Context, publicquery.DiscoveryQuery) (publicquery.DiscoveryResult, error)
	MunicipalitySituation(context.Context, publicquery.SituationQuery) (publicquery.Situation, error)
	Search(context.Context, publicquery.SearchQuery) (publicquery.SearchResult, error)
	Document(context.Context, publicquery.DocumentQuery) (publicquery.DocumentResult, error)
	Coverage(context.Context, publicquery.CoverageQuery) (publicquery.CoverageResult, error)
}

type publicService struct {
	queries      PublicQueries
	now          func() time.Time
	maxPageSize  int
	views        PublicViewStore
	viewLifetime time.Duration
	cursorKey    []byte
	copies       PublicCopyAccess
}

type publicMeta struct {
	DatasetVersion string     `json:"dataset_version"`
	EvaluationTime time.Time  `json:"evaluation_time"`
	ServedAt       time.Time  `json:"served_at"`
	ViewExpiresAt  time.Time  `json:"view_expires_at"`
	HistoryStart   *time.Time `json:"history_start"`
	HistoryGaps    []string   `json:"history_gaps"`
	Limitations    []string   `json:"limitations"`
	NextCursor     *string    `json:"next_cursor"`
}

type publicError struct {
	Code              string                     `json:"code"`
	Message           string                     `json:"message"`
	RetryAfterSeconds *int                       `json:"retry_after_seconds"`
	Candidates        []publicquery.Municipality `json:"candidates"`
}

type publicResponse struct {
	Data     any          `json:"data,omitempty"`
	Meta     *publicMeta  `json:"meta,omitempty"`
	Error    *publicError `json:"error,omitempty"`
	ServedAt *time.Time   `json:"served_at,omitempty"`
}

type commonInput struct {
	DatasetVersion string `json:"dataset_version"`
	EvaluationTime string `json:"evaluation_time"`
	KnownAt        string `json:"known_at"`
	PageSize       *int   `json:"page_size"`
	Cursor         string `json:"cursor"`
}

type discoveryInput struct {
	commonInput
	Name       string `json:"name"`
	ISTAT      string `json:"istat"`
	PostalCode string `json:"postal_code"`
}

type situationInput struct {
	commonInput
	MunicipalityISTAT string `json:"municipality_istat"`
}

type searchInput struct {
	commonInput
	MunicipalityISTAT string `json:"municipality_istat"`
	Zone              string `json:"zone"`
	Kind              string `json:"kind"`
	Product           string `json:"product"`
	Risk              string `json:"risk"`
	Status            string `json:"status"`
	SourceID          string `json:"source_id"`
	From              string `json:"from"`
	To                string `json:"to"`
}

type documentInput struct {
	commonInput
	DocumentID      string `json:"document_id"`
	VersionID       string `json:"version_id"`
	IncludeVersions bool   `json:"include_versions"`
}

type coverageInput struct {
	commonInput
	MunicipalityISTAT string `json:"municipality_istat"`
	SourceID          string `json:"source_id"`
	Product           string `json:"product"`
}

type documentData struct {
	Document publicquery.Document   `json:"document"`
	Versions []publicquery.Document `json:"versions"`
}

func (s *publicService) invokeUnviewed(ctx context.Context, operation string, raw json.RawMessage) (publicResponse, int) {
	servedAt := s.now().UTC()
	switch operation {
	case "discover_municipalities":
		var input discoveryInput
		if err := decodeInput(raw, &input); err != nil {
			return s.failure(servedAt, err, nil)
		}
		qt, err := input.commonInput.queryTime(servedAt, s.maxPageSize)
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		result, err := s.queries.DiscoverMunicipalities(ctx, publicquery.DiscoveryQuery{QueryTime: qt, Name: input.Name, ISTAT: input.ISTAT, PostalCode: input.PostalCode})
		if err != nil {
			return s.failure(servedAt, err, result.Municipalities)
		}
		return s.success(operation, input.commonInput, servedAt, qt, result.Municipalities, result.History)
	case "get_municipality_situation":
		var input situationInput
		if err := decodeInput(raw, &input); err != nil {
			return s.failure(servedAt, err, nil)
		}
		qt, err := input.commonInput.queryTime(servedAt, s.maxPageSize)
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		result, err := s.queries.MunicipalitySituation(ctx, publicquery.SituationQuery{QueryTime: qt, MunicipalityISTAT: input.MunicipalityISTAT})
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		return s.success(operation, input.commonInput, servedAt, qt, result, result.History)
	case "search_alerts":
		var input searchInput
		if err := decodeInput(raw, &input); err != nil {
			return s.failure(servedAt, err, nil)
		}
		qt, err := input.commonInput.queryTime(servedAt, s.maxPageSize)
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		from, err := optionalTime(input.From)
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		to, err := optionalTime(input.To)
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		result, err := s.queries.Search(ctx, publicquery.SearchQuery{QueryTime: qt, Kind: input.Kind, MunicipalityISTAT: input.MunicipalityISTAT, Zone: input.Zone, Product: input.Product, Risk: input.Risk, Status: input.Status, SourceID: input.SourceID, From: from, To: to})
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		var data any
		switch result.Kind {
		case "document":
			data = result.Documents
		case "measure":
			data = result.Measures
		case "regional":
			data = result.Regional
		default:
			return s.failure(servedAt, publicquery.ErrInvalidParameters, nil)
		}
		return s.success(operation, input.commonInput, servedAt, qt, data, result.History)
	case "get_document":
		var input documentInput
		if err := decodeInput(raw, &input); err != nil {
			return s.failure(servedAt, err, nil)
		}
		qt, err := input.commonInput.queryTime(servedAt, s.maxPageSize)
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		result, err := s.queries.Document(ctx, publicquery.DocumentQuery{QueryTime: qt, DocumentID: input.DocumentID, VersionID: input.VersionID, IncludeVersions: input.IncludeVersions})
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		data := documentData{Document: result.Document, Versions: result.Versions}
		return s.success(operation, input.commonInput, servedAt, qt, data, result.History)
	case "get_source_coverage":
		var input coverageInput
		if err := decodeInput(raw, &input); err != nil {
			return s.failure(servedAt, err, nil)
		}
		qt, err := input.commonInput.queryTime(servedAt, s.maxPageSize)
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		result, err := s.queries.Coverage(ctx, publicquery.CoverageQuery{QueryTime: qt, MunicipalityISTAT: input.MunicipalityISTAT, SourceID: input.SourceID, Product: input.Product})
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		return s.success(operation, input.commonInput, servedAt, qt, result.Sources, result.History)
	default:
		return s.failure(servedAt, publicquery.ErrInvalidParameters, nil)
	}
}

func decodeInput(raw json.RawMessage, destination any) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		raw = json.RawMessage(`{}`)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return publicquery.ErrInvalidParameters
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return publicquery.ErrInvalidParameters
	}
	return nil
}

func (input commonInput) queryTime(defaultTime time.Time, maxPageSize int) (publicquery.QueryTime, error) {
	// Durable views and cursors are implemented by task 6.4. Rejecting them is
	// explicit and prevents a caller from assuming an unpinned result is stable.
	if input.PageSize != nil && (*input.PageSize < 1 || *input.PageSize > maxPageSize) {
		return publicquery.QueryTime{}, publicquery.ErrInvalidParameters
	}
	if input.DatasetVersion != "" || input.Cursor != "" || input.PageSize != nil {
		return publicquery.QueryTime{}, publicquery.ErrInvalidParameters
	}
	evaluation, err := requiredOrDefaultTime(input.EvaluationTime, defaultTime)
	if err != nil {
		return publicquery.QueryTime{}, err
	}
	knownAt, err := requiredOrDefaultTime(input.KnownAt, evaluation)
	if err != nil || knownAt.After(evaluation) {
		return publicquery.QueryTime{}, publicquery.ErrInvalidParameters
	}
	return publicquery.QueryTime{EvaluationTime: evaluation, KnownAt: knownAt}, nil
}

func requiredOrDefaultTime(value string, fallback time.Time) (time.Time, error) {
	if value == "" {
		return fallback.UTC(), nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, publicquery.ErrInvalidParameters
	}
	return parsed.UTC(), nil
}

func optionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, publicquery.ErrInvalidParameters
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func (s *publicService) success(operation string, input commonInput, servedAt time.Time, qt publicquery.QueryTime, data any, history publicquery.History) (publicResponse, int) {
	if s.views == nil && !withinPageLimit(data, s.maxPageSize) {
		return s.failure(servedAt, errPaginationUnavailable, nil)
	}
	versionMaterial, _ := json.Marshal(struct {
		Operation string    `json:"operation"`
		KnownAt   time.Time `json:"known_at"`
		Data      any       `json:"data"`
	}{operation, qt.KnownAt, data})
	digest := sha256.Sum256(versionMaterial)
	datasetVersion := "sha256:" + hex.EncodeToString(digest[:])
	if input.DatasetVersion != "" && input.DatasetVersion != datasetVersion {
		return s.failure(servedAt, publicquery.ErrInvalidParameters, nil)
	}
	gaps := history.Gaps
	if gaps == nil {
		gaps = []string{}
	}
	limitations := history.Limitations
	if limitations == nil {
		limitations = []string{}
	}
	if s.views == nil {
		limitations = append(limitations, "consistent pagination views and cursors are not yet available")
	}
	response := publicResponse{Data: data, Meta: &publicMeta{
		DatasetVersion: datasetVersion,
		EvaluationTime: qt.EvaluationTime,
		ServedAt:       servedAt,
		ViewExpiresAt:  servedAt,
		HistoryStart:   history.Start,
		HistoryGaps:    gaps,
		Limitations:    limitations,
		NextCursor:     nil,
	}}
	return response, http.StatusOK
}

func (s *publicService) failure(servedAt time.Time, err error, candidates []publicquery.Municipality) (publicResponse, int) {
	code, message, status := "service_unavailable", "public query storage is unavailable", http.StatusServiceUnavailable
	switch {
	case errors.Is(err, publiccopy.ErrDisabled):
		code, message, status = "copy_access_disabled", "public retained-copy access is disabled", http.StatusForbidden
	case errors.Is(err, publiccopy.ErrRestricted):
		code, message, status = "copy_access_restricted", "current source policy restricts retained-copy access", http.StatusForbidden
	case errors.Is(err, publiccopy.ErrUnknown):
		code, message, status = "unknown_identifier", "the requested document version is unknown", http.StatusNotFound
	case errors.Is(err, publiccopy.ErrStorage):
		code, message, status = "service_unavailable", "retained-copy storage is unavailable", http.StatusServiceUnavailable
	case errors.Is(err, errCursorExpired):
		code, message, status = "cursor_expired", "the query view expired or is no longer available; restart the query", http.StatusGone
	case errors.Is(err, errCursorMismatch):
		code, message, status = "cursor_mismatch", "the cursor does not match this operation or its filters", http.StatusBadRequest
	case errors.Is(err, errPaginationUnavailable):
		code, message, status = "service_unavailable", "the result exceeds the configured page size and consistent pagination is not yet available", http.StatusServiceUnavailable
	case errors.Is(err, publicquery.ErrInvalidParameters):
		code, message, status = "invalid_parameters", "request parameters are invalid or not yet supported", http.StatusBadRequest
	case errors.Is(err, publicquery.ErrAmbiguousMunicipality):
		code, message, status = "ambiguous_municipality", "municipality selection is ambiguous", http.StatusConflict
	case errors.Is(err, publicquery.ErrUnknownIdentifier):
		code, message, status = "unknown_identifier", "the requested identifier is unknown", http.StatusNotFound
	case errors.Is(err, publicquery.ErrUnsupportedArea):
		code, message, status = "unsupported_area", "the requested area is outside the supported scope", http.StatusUnprocessableEntity
	case errors.Is(err, publicquery.ErrMappingUnavailable):
		code, message, status = "mapping_unavailable", "the requested geographic mapping is unavailable", http.StatusUnprocessableEntity
	}
	if candidates == nil {
		candidates = []publicquery.Municipality{}
	}
	return publicResponse{Error: &publicError{Code: code, Message: message, Candidates: candidates}, ServedAt: &servedAt}, status
}

func withinPageLimit(data any, maximum int) bool {
	switch value := data.(type) {
	case []publicquery.Municipality:
		return len(value) <= maximum
	case publicquery.Situation:
		return len(value.LocalMeasures) <= maximum && len(value.OperationalPhases) <= maximum && len(value.RegionalProducts) <= maximum && len(value.DocumentsRequiringAttention) <= maximum && len(value.Coverage) <= maximum
	case []publicquery.Document:
		return len(value) <= maximum
	case []publicquery.Measure:
		return len(value) <= maximum
	case []publicquery.RegionalWarning:
		return len(value) <= maximum
	case documentData:
		return len(value.Versions) <= maximum
	case []publicquery.Coverage:
		return len(value) <= maximum
	default:
		return false
	}
}

func publicRoutes(mux *http.ServeMux, queries PublicQueries, now func() time.Time, runtime PublicRuntime) {
	limits := runtime.Limits
	if runtime.Views != nil && (runtime.ViewLifetime <= 0 || len(runtime.CursorKey) < 32) {
		panic("invalid public view configuration")
	}
	service := &publicService{queries: queries, now: now, maxPageSize: limits.MaxPageSize, views: runtime.Views, viewLifetime: runtime.ViewLifetime, cursorKey: append([]byte(nil), runtime.CursorKey...), copies: runtime.Copies}
	budget := newSlidingBudget(limits, now)
	mux.Handle("GET /v1/municipalities", budget.middleware(service.httpHandler("discover_municipalities", nil)))
	mux.Handle("GET /v1/municipalities/{municipality_istat}/situation", budget.middleware(service.httpHandler("get_municipality_situation", map[string]string{"municipality_istat": "municipality_istat"})))
	mux.Handle("GET /v1/search", budget.middleware(service.httpHandler("search_alerts", nil)))
	mux.Handle("GET /v1/documents/{document_id}", budget.middleware(service.httpHandler("get_document", map[string]string{"document_id": "document_id"})))
	mux.Handle("GET /v1/documents/{document_id}/versions/{version_id}/content", budget.middleware(service.copyHandler()))
	mux.Handle("GET /v1/sources/coverage", budget.middleware(service.httpHandler("get_source_coverage", nil)))
	// Count malformed paths and unsupported methods that reached the public API.
	mux.Handle("/v1/", budget.middleware(http.NotFoundHandler()))
	mux.Handle("/mcp", budget.middleware(publicMCPHandler(service, limits)))
}

func (s *publicService) httpHandler(operation string, pathFields map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := httpArguments(r.URL.Query(), r, pathFields)
		if err != nil {
			response, status := s.failure(s.now().UTC(), err, nil)
			httpserver.Reply(w, status, response)
			return
		}
		response, status := s.invoke(r.Context(), operation, raw)
		httpserver.Reply(w, status, response)
	}
}

func httpArguments(values url.Values, request *http.Request, pathFields map[string]string) (json.RawMessage, error) {
	allowed := map[string]bool{
		"name": true, "istat": true, "postal_code": true, "municipality_istat": true,
		"zone": true, "kind": true, "product": true, "risk": true, "status": true,
		"source_id": true, "from": true, "to": true, "version_id": true,
		"include_versions": true, "dataset_version": true, "evaluation_time": true,
		"known_at": true, "page_size": true, "cursor": true,
	}
	arguments := map[string]any{}
	for key, items := range values {
		if !allowed[key] || len(items) != 1 {
			return nil, publicquery.ErrInvalidParameters
		}
		switch key {
		case "include_versions":
			value, err := strconv.ParseBool(items[0])
			if err != nil {
				return nil, publicquery.ErrInvalidParameters
			}
			arguments[key] = value
		case "page_size":
			value, err := strconv.Atoi(items[0])
			if err != nil || value < 1 {
				return nil, publicquery.ErrInvalidParameters
			}
			arguments[key] = value
		default:
			arguments[key] = items[0]
		}
	}
	for argument, pathName := range pathFields {
		arguments[argument] = request.PathValue(pathName)
	}
	encoded, err := json.Marshal(arguments)
	return encoded, err
}

type mcpDescriptorFile struct {
	Tools []struct {
		Name         string              `json:"name"`
		Description  string              `json:"description"`
		InputSchema  json.RawMessage     `json:"inputSchema"`
		OutputSchema json.RawMessage     `json:"outputSchema"`
		Annotations  mcp.ToolAnnotations `json:"annotations"`
	} `json:"tools"`
}

func publicMCPHandler(service *publicService, limits PublicLimits) http.Handler {
	instructions := fmt.Sprintf("Anonymous API and MCP requests share a per-IP sliding budget of %d requests every %s; callers behind one IP share it. Maximum page size is %d.", limits.Allowance, limits.Window, limits.MaxPageSize)
	server := mcp.NewServer(&mcp.Implementation{Name: "iwa-public", Title: "Toscana public alert access", Version: "0.1.1"}, &mcp.ServerOptions{Capabilities: &mcp.ServerCapabilities{}, Instructions: instructions})
	var descriptors mcpDescriptorFile
	if err := json.Unmarshal(contract.MCPTools, &descriptors); err != nil {
		panic(fmt.Sprintf("invalid embedded MCP contract: %v", err))
	}
	for _, descriptor := range descriptors.Tools {
		operation := descriptor.Name
		server.AddTool(&mcp.Tool{Name: descriptor.Name, Description: descriptor.Description, InputSchema: descriptor.InputSchema, OutputSchema: descriptor.OutputSchema, Annotations: &descriptor.Annotations}, func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if request == nil || request.Params == nil {
				response, _ := service.failure(service.now().UTC(), publicquery.ErrInvalidParameters, nil)
				encoded, _ := json.Marshal(response)
				return &mcp.CallToolResult{StructuredContent: response, Content: []mcp.Content{&mcp.TextContent{Text: string(encoded)}}, IsError: true}, nil
			}
			response, status := service.invoke(ctx, operation, request.Params.Arguments)
			encoded, err := json.Marshal(response)
			if err != nil {
				return nil, err
			}
			result := &mcp.CallToolResult{StructuredContent: response, Content: []mcp.Content{&mcp.TextContent{Text: string(encoded)}}, IsError: status != http.StatusOK}
			if budget, ok := budgetFromContext(ctx); ok {
				result.Meta = mcp.Meta{"iwa.dev/rateLimit": map[string]any{"limit": budget.Limit, "remaining": budget.Remaining, "reset_seconds": budget.ResetSeconds, "max_page_size": budget.MaxPageSize, "partition": "client_ip"}}
			}
			return result, nil
		})
	}
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true, PropagateRequestCancellation: true, MaxRequestBodyBytes: 1 << 20})
	protocolOnly := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodPost && r.Header.Get("Mcp-Protocol-Version") != publicMCPProtocolVersion {
			http.Error(w, "unsupported MCP protocol version", http.StatusBadRequest)
			return
		}
		mcpHandler.ServeHTTP(w, r)
	})
	return http.NewCrossOriginProtection().Handler(protocolOnly)
}
