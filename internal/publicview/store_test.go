package publicview

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCreateRejectsInvalidViewBeforeStorage(t *testing.T) {
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	valid := Create{Operation: "get_source_coverage", RequestHash: strings.Repeat("a", 64), Request: json.RawMessage(`{}`), Data: json.RawMessage(`[]`), History: json.RawMessage(`{}`), EvaluationTime: now, KnownAt: now, CreatedAt: now, ExpiresAt: now.Add(30 * time.Minute)}
	tests := []Create{
		{},
		func() Create { value := valid; value.Operation = "admin"; return value }(),
		func() Create { value := valid; value.RequestHash = strings.Repeat("z", 64); return value }(),
		func() Create { value := valid; value.Request = json.RawMessage(`[]`); return value }(),
		func() Create { value := valid; value.Data = json.RawMessage(`no`); return value }(),
		func() Create { value := valid; value.ExpiresAt = now; return value }(),
	}
	store := &Store{}
	for index, request := range tests {
		if _, err := store.Create(context.Background(), request); !errors.Is(err, ErrInvalid) {
			t.Errorf("case %d: err=%v", index, err)
		}
	}
}
