package embedding

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/linking"
)

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return linking.MigrateSemantic(ctx, pool)
}
