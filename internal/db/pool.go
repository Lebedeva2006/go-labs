package db

import (
 "context"
 "fmt"
 "time"

 "github.com/jackc/pgx/v5/pgxpool"

 "github.com/Lebedeva2006/go-labs/internal/config"
)

func NewPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
 poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
 if err != nil {
  return nil, fmt.Errorf("db: failed to parse database url: %w", err)
 }

 poolConfig.MaxConns = int32(cfg.DatabaseMaxConns)
 poolConfig.MinConns = int32(cfg.DatabaseMinConns)
 poolConfig.MaxConnLifetime = cfg.DatabaseMaxConnLifetime

 connectCtx, cancel := context.WithTimeout(ctx, cfg.DatabaseConnectTimeout)
 defer cancel()

 pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
 if err != nil {
  return nil, fmt.Errorf("db: failed to create pool: %w", err)
 }

 pingCtx, cancel := context.WithTimeout(ctx, cfg.DatabaseConnectTimeout)
 defer cancel()

 if err := pool.Ping(pingCtx); err != nil {
  pool.Close()
  return nil, fmt.Errorf("db: failed to ping database: %w", err)
 }

 return pool, nil
}

func Ping(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) error {
 pingCtx, cancel := context.WithTimeout(ctx, timeout)
 defer cancel()

 if err := pool.Ping(pingCtx); err != nil {
  return fmt.Errorf("db: ping failed: %w", err)
 }
 return nil
}
