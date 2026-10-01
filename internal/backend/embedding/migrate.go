package embedding

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return linking.MigrateSemantic(ctx, pool)
}
