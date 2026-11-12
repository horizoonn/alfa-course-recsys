package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, errors.New("parse PostgreSQL connection configuration")
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return pool, nil
}
