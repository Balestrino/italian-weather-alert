// Package evidenceguard coordinates evidence snapshots with destructive cleanup.
package evidenceguard

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// Lock uses a dedicated session. Close the connection if unlocking fails, so a
// cancelled operation cannot leave a session lock behind in the pool.
func Lock(ctx context.Context, pool *pgxpool.Pool, shared bool) (func(), error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	lock, unlock := "pg_advisory_lock", "pg_advisory_unlock"
	if shared {
		lock += "_shared"
		unlock += "_shared"
	}
	if _, err = conn.Exec(ctx, "SELECT "+lock+"(730026)"); err != nil {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = conn.Conn().Close(closeCtx)
		cancel()
		conn.Release()
		return nil, err
	}
	return func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := conn.Exec(c, "SELECT "+unlock+"(730026)"); err != nil {
			_ = conn.Conn().Close(c)
		}
		conn.Release()
	}, nil
}
