package workerdiag

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestSafeDiagnosticPreservesCause(t *testing.T) {
	for _, tc := range []struct {
		err             error
		category, state string
	}{
		{&pgconn.PgError{Code: "40001", Message: "password=secret", Detail: "private document"}, "database", "40001"},
		{context.DeadlineExceeded, "timeout", ""},
		{context.Canceled, "canceled", ""},
		{errors.New("https://secret@host/private"), "internal", ""},
	} {
		err := Wrap("heartbeat", tc.err)
		op, category, state := Describe(err)
		if op != "heartbeat" || category != tc.category || state != tc.state || !errors.Is(err, tc.err) || err.Error() != "worker operation failed" {
			t.Fatalf("unsafe or incorrect diagnostic: %q %q %q", op, category, state)
		}
	}
	op, _, state := Describe(Wrap("password=secret", &pgconn.PgError{Code: "private"}))
	if op != "run" || state != "" {
		t.Fatal("untrusted identifiers leaked")
	}
}
