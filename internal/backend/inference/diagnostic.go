package inference

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Diagnostic contains only bounded identifiers and allowlisted categories.
// It deliberately has no free-form provider message or request/response body.
type Diagnostic struct {
	HTTPStatus                        int
	Category, ProviderCode, RequestID string
}

var opaqueID = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,160}$`)

func safeIdentifier(value, secret string) string {
	if !opaqueID.MatchString(value) || (secret != "" && strings.Contains(value, secret)) {
		return ""
	}
	return value
}

func responseDiagnostic(response *http.Response, secret string) Diagnostic {
	id := response.Header.Get("X-Request-ID")
	if id == "" {
		id = response.Header.Get("Request-ID")
	}
	return Diagnostic{HTTPStatus: response.StatusCode, RequestID: safeIdentifier(id, secret)}
}

func rejection(response *http.Response, secret string) *CallError {
	err := statusError(response)
	err.Diagnostic = responseDiagnostic(response, secret)
	var wire struct {
		Error struct {
			Code string `json:"code"`
			Type string `json:"type"`
		} `json:"error"`
		Usage *wireUsage `json:"usage"`
	}
	// Unknown error shapes/messages are intentionally discarded.
	if json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&wire) == nil {
		for _, code := range []string{wire.Error.Code, wire.Error.Type} {
			switch code {
			case "invalid_api_key", "authentication_error", "insufficient_quota", "quota_exceeded", "budget_exceeded", "trial_expired", "rate_limit_exceeded", "model_not_found", "invalid_request_error", "context_length_exceeded", "server_error", "permission_denied":
				err.Diagnostic.ProviderCode = code
			}
			if err.Diagnostic.ProviderCode != "" {
				break
			}
		}
		err.Usage = wire.Usage.usage()
	}
	switch response.StatusCode {
	case 401:
		err.Diagnostic.Category = "authentication"
	case 403:
		err.Diagnostic.Category = "permission"
	case 402:
		err.Diagnostic.Category = "quota"
	case 429:
		err.Diagnostic.Category = "rate_limit"
	case 400, 404, 413, 415, 422:
		err.Diagnostic.Category = "request"
	default:
		if err.Temporary {
			err.Diagnostic.Category = "availability"
		} else {
			err.Diagnostic.Category = "rejected"
		}
	}
	switch err.Diagnostic.ProviderCode {
	case "invalid_api_key", "authentication_error":
		err.Diagnostic.Category = "authentication"
	case "insufficient_quota", "quota_exceeded", "budget_exceeded", "trial_expired":
		err.Diagnostic.Category = "quota"
	}
	return err
}

type wireUsage struct {
	PromptTokens     *int64 `json:"prompt_tokens"`
	CompletionTokens *int64 `json:"completion_tokens"`
	PromptDetails    *struct {
		CachedTokens *int64 `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
}

func nonnegative(value *int64) *int64 {
	if value != nil && *value < 0 {
		return nil
	}
	return value
}

func (u *wireUsage) usage() Usage {
	if u == nil {
		return Usage{}
	}
	result := Usage{InputTokens: nonnegative(u.PromptTokens), OutputTokens: nonnegative(u.CompletionTokens)}
	if u.PromptDetails != nil {
		result.CacheReadTokens = nonnegative(u.PromptDetails.CachedTokens)
	}
	return result
}
