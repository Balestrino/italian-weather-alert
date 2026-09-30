// Package workerdiag reports lifecycle failures without exposing database or
// transport error text, which may contain credentials and document content.
package workerdiag

import (
	"context"
	"errors"
	"net"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type OperationError struct {
	Operation string
	Err       error
}

// Retryable is deliberately narrow: configuration/schema/integrity failures
// remain fatal. A worker retries at most three consecutive transient failures.
func Retryable(err error) bool {
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		return pg.Code == "40001" || pg.Code == "40P01" || pg.Code == "57P01" || pg.Code == "57P02" || pg.Code == "57P03" || len(pg.Code) == 5 && pg.Code[:2] == "08"
	}
	var network net.Error
	return errors.As(err, &network) || errors.Is(err, context.DeadlineExceeded)
}

func Backoff(ctx context.Context, failures int) bool {
	timer := time.NewTimer(time.Duration(1<<uint(failures-1)) * 250 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (e *OperationError) Error() string { return "worker operation failed" }
func (e *OperationError) Unwrap() error { return e.Err }

func Wrap(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &OperationError{operation, err}
}

var safeOperation = regexp.MustCompile(`^[a-z_]{1,40}$`)
var sqlState = regexp.MustCompile(`^[0-9A-Z]{5}$`)

func Describe(err error) (operation, category, state string) {
	operation, category = "run", "internal"
	var op *OperationError
	if errors.As(err, &op) && safeOperation.MatchString(op.Operation) {
		operation = op.Operation
	}
	switch {
	case err == nil:
		category = "returned"
	case errors.Is(err, context.Canceled):
		category = "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		category = "timeout"
	default:
		var pg *pgconn.PgError
		var network net.Error
		if errors.As(err, &pg) {
			category = "database"
			if sqlState.MatchString(pg.Code) {
				state = pg.Code
			}
		} else if errors.As(err, &network) {
			category = "network"
		}
	}
	return
}
