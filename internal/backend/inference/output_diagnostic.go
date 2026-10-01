package inference

import (
	"encoding/json"
	"strings"
)

// OutputEnvelopeReason contains no response text or dynamic provider field names.
func OutputEnvelopeReason(raw, finish string) string {
	if finish == "length" {
		return "truncated"
	}
	if finish == "content_filter" {
		return "refused"
	}
	if strings.TrimSpace(raw) == "" {
		return "empty"
	}
	if !json.Valid([]byte(raw)) {
		return "json_syntax"
	}
	var object map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &object) != nil || object == nil {
		return "schema"
	}
	return ""
}
