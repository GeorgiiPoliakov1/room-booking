package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupDB(databaseURL string, maxOpen, maxIdle int, maxLifetime time.Duration) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}
	config.MaxConns = int32(maxOpen)
	config.MinConns = int32(maxIdle)
	config.MaxConnLifetime = maxLifetime

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgxpool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
