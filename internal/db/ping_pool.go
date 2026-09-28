package db

import (
 "context"
 "time"

 "github.com/jackc/pgx/v5/pgxpool"
)

type PoolPinger struct {
 Pool *pgxpool.Pool
}

func (p PoolPinger) Ping(ctx context.Context, timeout time.Duration) error {
 return Ping(ctx, p.Pool, timeout)
}
