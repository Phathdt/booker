package shared

import (
	"context"
	"fmt"

	"booker/config"
	"booker/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// InitLogger creates and initializes the application logger.
func InitLogger(cfg *config.Config) logger.Logger {
	return logger.New(cfg.Logger.Format, cfg.Logger.Level)
}

// InitDatabase creates and initializes the database connection pool.
func InitDatabase(ctx context.Context, uri string, log logger.Logger) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.ConnConfig.Tracer = NewPgxTracer(log)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// InitRedis creates and initializes the Redis client.
func InitRedis(ctx context.Context, uri string) (*redis.Client, error) {
	opts, err := redis.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
