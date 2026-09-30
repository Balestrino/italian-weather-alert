package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/publiccopy"
	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
	"github.com/Balestrino/italian-weather-alert/internal/publicview"
)

var (
	errCursorExpired  = errors.New("public query cursor expired")
	errCursorMismatch = errors.New("public query cursor mismatch")
)

type PublicViewStore interface {
	Create(context.Context, publicview.Create) (publicview.View, error)
	Get(context.Context, string) (publicview.View, error)
}

type PublicRuntime struct {
	Limits       PublicLimits
	Views        PublicViewStore
	ViewLifetime time.Duration
	CursorKey    []byte
	Copies       PublicCopyAccess
}

type PublicCopyAccess interface {
	Link(context.Context, string, string) (*string, error)
	Read(context.Context, string, string) (publiccopy.Content, error)
}

type cursorPayload struct {
	Version     int            `json:"v"`
	ViewID      string         `json:"view_id"`
	Operation   string         `json:"operation"`
	RequestHash string         `json:"request_hash"`
	Positions   map[string]int `json:"positions"`
	ExpiresAt   int64          `json:"expires_at"`
}

func (s *publicService) invoke(ctx context.Context, operation string, raw json.RawMessage) (publicResponse, int) {
	if s.views == nil {
		response, status := s.invokeUnviewed(ctx, operation, raw)
		if status == http.StatusOK {
			var err error
			response.Data, err = s.applyCopyLinks(ctx, response.Data)
			if err != nil {
				return s.failure(s.now().UTC(), err, nil)
			}
		}
		return response, status
	}
	servedAt := s.now().UTC()
	if err := validateInputShape(operation, raw); err != nil {
		return s.failure(servedAt, err, nil)
	}
	var common commonInput
	if err := json.Unmarshal(defaultObject(raw), &common); err != nil {
		return s.failure(servedAt, publicquery.ErrInvalidParameters, nil)
	}
	if common.Cursor == "" && common.DatasetVersion == "" {
		return s.createView(ctx, operation, raw, servedAt)
	}

	positions := map[string]int{}
	viewID := common.DatasetVersion
	var cursor cursorPayload
	if common.Cursor != "" {
		var err error
		cursor, err = decodeCursor(common.Cursor, s.cursorKey)
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		if cursor.Operation != operation || (viewID != "" && viewID != cursor.ViewID) {
			return s.failure(servedAt, errCursorMismatch, nil)
		}
		if !servedAt.Before(time.Unix(cursor.ExpiresAt, 0).UTC()) {
			return s.failure(servedAt, errCursorExpired, nil)
		}
		viewID, positions = cursor.ViewID, cursor.Positions
	}
	view, err := s.views.Get(ctx, viewID)
	if errors.Is(err, publicview.ErrNotFound) {
		return s.failure(servedAt, errCursorExpired, nil)
	}
	if err != nil {
		return s.failure(servedAt, err, nil)
	}
	if !servedAt.Before(view.ExpiresAt) {
		return s.failure(servedAt, errCursorExpired, nil)
	}
	if view.Operation != operation || (common.Cursor != "" && cursor.RequestHash != view.RequestHash) {
		return s.failure(servedAt, errCursorMismatch, nil)
	}
	persistedHash, err := requestDigest(view.Request)
	if err != nil || persistedHash != view.RequestHash {
		return s.failure(servedAt, errors.New("public view request hash mismatch"), nil)
	}
	if err := requestMatchesView(raw, view.Request); err != nil {
		return s.failure(servedAt, err, nil)
	}
	data, err := decodeStoredData(operation, view.Request, view.Data)
	if err != nil {
		return s.failure(servedAt, err, nil)
	}
	var history publicquery.History
	if err := json.Unmarshal(view.History, &history); err != nil {
		return s.failure(servedAt, err, nil)
	}
	pageSize, err := storedPageSize(view.Request, s.maxPageSize)
	if err != nil {
		return s.failure(servedAt, err, nil)
	}
	return s.page(ctx, view, data, history, positions, pageSize, servedAt)
}

func (s *publicService) createView(ctx context.Context, operation string, raw json.RawMessage, servedAt time.Time) (publicResponse, int) {
	canonical, sanitized, pageSize, qt, err := canonicalInitialRequest(operation, raw, servedAt, s.maxPageSize)
	if err != nil {
		return s.failure(servedAt, err, nil)
	}
	full, status := s.invokeUnviewed(ctx, operation, sanitized)
	if status != http.StatusOK || full.Meta == nil {
		return full, status
	}
	dataJSON, err := json.Marshal(full.Data)
	if err != nil {
		return s.failure(servedAt, err, nil)
	}
	history := publicquery.History{Start: full.Meta.HistoryStart, Gaps: full.Meta.HistoryGaps, Limitations: full.Meta.Limitations}
	historyJSON, err := json.Marshal(history)
	if err != nil {
		return s.failure(servedAt, err, nil)
	}
	requestHash, err := requestDigest(canonical)
	if err != nil {
		return s.failure(servedAt, err, nil)
	}
	view, err := s.views.Create(ctx, publicview.Create{Operation: operation, RequestHash: requestHash, Request: canonical, Data: dataJSON, History: historyJSON, EvaluationTime: qt.EvaluationTime, KnownAt: qt.KnownAt, CreatedAt: servedAt, ExpiresAt: servedAt.Add(s.viewLifetime)})
	if err != nil {
		return s.failure(servedAt, err, nil)
	}
	return s.page(ctx, view, full.Data, history, map[string]int{}, pageSize, servedAt)
}

func (s *publicService) page(ctx context.Context, view publicview.View, data any, history publicquery.History, positions map[string]int, pageSize int, servedAt time.Time) (publicResponse, int) {
	page, next, more, err := paginateData(data, positions, pageSize)
	if err != nil {
		return s.failure(servedAt, err, nil)
	}
	page = refreshData(page, servedAt)
	page, err = s.applyCopyLinks(ctx, page)
	if err != nil {
		return s.failure(servedAt, err, nil)
	}
	var nextCursor *string
	if more {
		token, err := encodeCursor(cursorPayload{Version: 1, ViewID: view.ID, Operation: view.Operation, RequestHash: view.RequestHash, Positions: next, ExpiresAt: view.ExpiresAt.Unix()}, s.cursorKey)
		if err != nil {
			return s.failure(servedAt, err, nil)
		}
		nextCursor = &token
	}
	gaps := history.Gaps
	if gaps == nil {
		gaps = []string{}
	}
	limitations := history.Limitations
	if limitations == nil {
		limitations = []string{}
	}
	return publicResponse{Data: page, Meta: &publicMeta{DatasetVersion: view.ID, EvaluationTime: view.EvaluationTime, ServedAt: servedAt, ViewExpiresAt: view.ExpiresAt, HistoryStart: history.Start, HistoryGaps: gaps, Limitations: limitations, NextCursor: nextCursor}}, http.StatusOK
}

func canonicalInitialRequest(operation string, raw json.RawMessage, servedAt time.Time, maximum int) (json.RawMessage, json.RawMessage, int, publicquery.QueryTime, error) {
	if err := validateInputShape(operation, raw); err != nil {
		return nil, nil, 0, publicquery.QueryTime{}, err
	}
	var common commonInput
	if err := json.Unmarshal(defaultObject(raw), &common); err != nil || common.Cursor != "" || common.DatasetVersion != "" {
		return nil, nil, 0, publicquery.QueryTime{}, publicquery.ErrInvalidParameters
	}
	pageSize := maximum
	if common.PageSize != nil {
		pageSize = *common.PageSize
	}
	if pageSize < 1 || pageSize > maximum {
		return nil, nil, 0, publicquery.QueryTime{}, publicquery.ErrInvalidParameters
	}
	evaluation, err := requiredOrDefaultTime(common.EvaluationTime, servedAt)
	if err != nil {
		return nil, nil, 0, publicquery.QueryTime{}, err
	}
	knownAt, err := requiredOrDefaultTime(common.KnownAt, evaluation)
	if err != nil || knownAt.After(evaluation) {
		return nil, nil, 0, publicquery.QueryTime{}, publicquery.ErrInvalidParameters
	}
	var fields map[string]any
	if err := json.Unmarshal(defaultObject(raw), &fields); err != nil {
		return nil, nil, 0, publicquery.QueryTime{}, publicquery.ErrInvalidParameters
	}
	delete(fields, "cursor")
	delete(fields, "dataset_version")
	fields["evaluation_time"] = evaluation.Format(time.RFC3339Nano)
	fields["known_at"] = knownAt.Format(time.RFC3339Nano)
	fields["page_size"] = pageSize
	canonical, err := json.Marshal(fields)
	if err != nil {
		return nil, nil, 0, publicquery.QueryTime{}, err
	}
	delete(fields, "page_size")
	sanitized, err := json.Marshal(fields)
	return canonical, sanitized, pageSize, publicquery.QueryTime{EvaluationTime: evaluation, KnownAt: knownAt}, err
}

func requestMatchesView(raw, stored json.RawMessage) error {
	var incoming, canonical map[string]json.RawMessage
	if json.Unmarshal(defaultObject(raw), &incoming) != nil || json.Unmarshal(stored, &canonical) != nil {
		return errCursorMismatch
	}
	delete(incoming, "cursor")
	delete(incoming, "dataset_version")
	for key, value := range incoming {
		expected, ok := canonical[key]
		if !ok || !equalJSON(value, expected) {
			return errCursorMismatch
		}
	}
	return nil
}

func equalJSON(left, right json.RawMessage) bool {
	var l, r any
	if json.Unmarshal(left, &l) != nil || json.Unmarshal(right, &r) != nil {
		return false
	}
	lb, _ := json.Marshal(l)
	rb, _ := json.Marshal(r)
	return bytes.Equal(lb, rb)
}

func requestDigest(raw json.RawMessage) (string, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(canonical)
	return hex.EncodeToString(hash[:]), nil
}

func validateInputShape(operation string, raw json.RawMessage) error {
	var destination any
	switch operation {
	case "discover_municipalities":
		destination = &discoveryInput{}
	case "get_municipality_situation":
		destination = &situationInput{}
	case "search_alerts":
		destination = &searchInput{}
	case "get_document":
		destination = &documentInput{}
	case "get_source_coverage":
		destination = &coverageInput{}
	default:
		return publicquery.ErrInvalidParameters
	}
	return decodeInput(raw, destination)
}

func defaultObject(raw json.RawMessage) json.RawMessage {
	if len(bytes.TrimSpace(raw)) == 0 {
		return json.RawMessage(`{}`)
	}
	return raw
}

func storedPageSize(raw json.RawMessage, maximum int) (int, error) {
	var fields struct {
		PageSize int `json:"page_size"`
	}
	if json.Unmarshal(raw, &fields) != nil || fields.PageSize < 1 || fields.PageSize > maximum {
		return 0, errCursorMismatch
	}
	return fields.PageSize, nil
}

func decodeStoredData(operation string, request, raw json.RawMessage) (any, error) {
	switch operation {
	case "discover_municipalities":
		var value []publicquery.Municipality
		err := json.Unmarshal(raw, &value)
		return value, err
	case "get_municipality_situation":
		var value publicquery.Situation
		err := json.Unmarshal(raw, &value)
		return value, err
	case "search_alerts":
		var input searchInput
		if json.Unmarshal(request, &input) != nil {
			return nil, errCursorMismatch
		}
		switch input.Kind {
		case "document":
			var value []publicquery.Document
			err := json.Unmarshal(raw, &value)
			return value, err
		case "measure":
			var value []publicquery.Measure
			err := json.Unmarshal(raw, &value)
			return value, err
		case "regional":
			var value []publicquery.RegionalWarning
			err := json.Unmarshal(raw, &value)
			return value, err
		default:
			return nil, errCursorMismatch
		}
	case "get_document":
		var value documentData
		err := json.Unmarshal(raw, &value)
		return value, err
	case "get_source_coverage":
		var value []publicquery.Coverage
		err := json.Unmarshal(raw, &value)
		return value, err
	default:
		return nil, errCursorMismatch
	}
}

func paginateData(data any, positions map[string]int, size int) (any, map[string]int, bool, error) {
	switch value := data.(type) {
	case []publicquery.Municipality:
		page, next, more, err := pageSlice(value, positions, "data", size)
		return page, next, more, err
	case []publicquery.Document:
		page, next, more, err := pageSlice(value, positions, "data", size)
		return page, next, more, err
	case []publicquery.Measure:
		page, next, more, err := pageSlice(value, positions, "data", size)
		return page, next, more, err
	case []publicquery.RegionalWarning:
		page, next, more, err := pageSlice(value, positions, "data", size)
		return page, next, more, err
	case []publicquery.Coverage:
		page, next, more, err := pageSlice(value, positions, "data", size)
		return page, next, more, err
	case documentData:
		if !onlyPositionKeys(positions, "versions") {
			return nil, nil, false, errCursorMismatch
		}
		versions, next, more, err := slicePage(value.Versions, positions["versions"], size)
		value.Versions = versions
		return value, map[string]int{"versions": next}, more, err
	case publicquery.Situation:
		keys := []string{"local_measures", "operational_phases", "regional_products", "documents_requiring_attention", "coverage"}
		if !onlyPositionKeys(positions, keys...) {
			return nil, nil, false, errCursorMismatch
		}
		next := map[string]int{}
		more := false
		var err error
		value.LocalMeasures, next["local_measures"], more, err = slicePage(value.LocalMeasures, positions["local_measures"], size)
		if err != nil {
			return nil, nil, false, err
		}
		var itemMore bool
		value.OperationalPhases, next["operational_phases"], itemMore, err = slicePage(value.OperationalPhases, positions["operational_phases"], size)
		more = more || itemMore
		if err != nil {
			return nil, nil, false, err
		}
		value.RegionalProducts, next["regional_products"], itemMore, err = slicePage(value.RegionalProducts, positions["regional_products"], size)
		more = more || itemMore
		if err != nil {
			return nil, nil, false, err
		}
		value.DocumentsRequiringAttention, next["documents_requiring_attention"], itemMore, err = slicePage(value.DocumentsRequiringAttention, positions["documents_requiring_attention"], size)
		more = more || itemMore
		if err != nil {
			return nil, nil, false, err
		}
		value.Coverage, next["coverage"], itemMore, err = slicePage(value.Coverage, positions["coverage"], size)
		more = more || itemMore
		return value, next, more, err
	default:
		return nil, nil, false, errCursorMismatch
	}
}

func pageSlice[T any](values []T, positions map[string]int, key string, size int) ([]T, map[string]int, bool, error) {
	if !onlyPositionKeys(positions, key) {
		return nil, nil, false, errCursorMismatch
	}
	page, next, more, err := slicePage(values, positions[key], size)
	return page, map[string]int{key: next}, more, err
}

func slicePage[T any](values []T, start, size int) ([]T, int, bool, error) {
	if start < 0 || start > len(values) || size < 1 {
		return nil, 0, false, errCursorMismatch
	}
	end := min(start+size, len(values))
	page := values[start:end]
	return page, end, end < len(values), nil
}

func onlyPositionKeys(positions map[string]int, allowed ...string) bool {
	for key := range positions {
		if !slices.Contains(allowed, key) {
			return false
		}
	}
	return true
}

func encodeCursor(payload cursorPayload, key []byte) (string, error) {
	if len(key) < 32 {
		return "", errors.New("public cursor key unavailable")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(encoded)
	return base64.RawURLEncoding.EncodeToString(encoded) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func decodeCursor(token string, key []byte) (cursorPayload, error) {
	if len(key) < 32 || len(token) > 4096 {
		return cursorPayload{}, errCursorMismatch
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return cursorPayload{}, errCursorMismatch
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return cursorPayload{}, errCursorMismatch
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return cursorPayload{}, errCursorMismatch
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(payloadJSON)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return cursorPayload{}, errCursorMismatch
	}
	var payload cursorPayload
	if json.Unmarshal(payloadJSON, &payload) != nil || payload.Version != 1 || payload.ViewID == "" || payload.Operation == "" || len(payload.RequestHash) != 64 || payload.ExpiresAt < 1 || payload.Positions == nil {
		return cursorPayload{}, errCursorMismatch
	}
	return payload, nil
}

func refreshData(data any, servedAt time.Time) any {
	switch value := data.(type) {
	case []publicquery.Document:
		for index := range value {
			refreshDocument(&value[index], servedAt)
		}
		return value
	case []publicquery.Measure:
		for index := range value {
			refreshQuality(&value[index].Quality, servedAt)
		}
		return value
	case []publicquery.RegionalWarning:
		for index := range value {
			refreshQuality(&value[index].Quality, servedAt)
		}
		return value
	case []publicquery.Coverage:
		for index := range value {
			refreshQuality(&value[index].Quality, servedAt)
		}
		return value
	case documentData:
		refreshDocument(&value.Document, servedAt)
		for index := range value.Versions {
			refreshDocument(&value.Versions[index], servedAt)
		}
		return value
	case publicquery.Situation:
		for index := range value.LocalMeasures {
			refreshQuality(&value.LocalMeasures[index].Quality, servedAt)
		}
		for index := range value.OperationalPhases {
			refreshQuality(&value.OperationalPhases[index].Quality, servedAt)
		}
		for index := range value.RegionalProducts {
			refreshQuality(&value.RegionalProducts[index].Quality, servedAt)
		}
		for index := range value.DocumentsRequiringAttention {
			refreshDocument(&value.DocumentsRequiringAttention[index], servedAt)
		}
		for index := range value.Coverage {
			refreshQuality(&value.Coverage[index].Quality, servedAt)
		}
		return value
	}
	return data
}

func refreshDocument(value *publicquery.Document, servedAt time.Time) {
	refreshQuality(&value.Quality, servedAt)
}

func refreshQuality(quality *publicquery.Quality, servedAt time.Time) {
	updating := &quality.Updating
	if updating.State != "ok" && updating.State != "delayed" || updating.LastCompleteAt == nil || updating.DelayThresholdSeconds < 1 {
		return
	}
	const delayLimitation = "complete-check delay threshold reached; fact validity is evaluated separately"
	limitations := updating.Limitations[:0]
	for _, limitation := range updating.Limitations {
		if limitation != delayLimitation {
			limitations = append(limitations, limitation)
		}
	}
	updating.Limitations = limitations
	if !servedAt.Before(updating.LastCompleteAt.Add(time.Duration(updating.DelayThresholdSeconds) * time.Second)) {
		updating.State = "delayed"
		updating.Limitations = append(updating.Limitations, delayLimitation)
	} else {
		updating.State = "ok"
	}
}
