package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Pool(ctx context.Context, conn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, conn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("fail while testing connection to database: %w", err)
	}

	return pool, nil
}
