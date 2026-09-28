package repository

import (
 "context"
 "fmt"
 "time"

 sq "github.com/Masterminds/squirrel"
 "github.com/jackc/pgx/v5/pgxpool"

 "github.com/Lebedeva2006/go-labs/internal/db"
 "github.com/Lebedeva2006/go-labs/internal/domain"
)


type TripStatusHistoryRepository struct {
 pool *pgxpool.Pool
 queryTimeout time.Duration
}

func NewTripStatusHistoryRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *TripStatusHistoryRepository {
 return &TripStatusHistoryRepository{pool: pool, queryTimeout: queryTimeout}
}


func (r *TripStatusHistoryRepository) Create(ctx context.Context, tripID string, fromStatus *domain.TripStatus, toStatus domain.TripStatus, reason string) error {
 ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
 defer cancel()

 exec := db.ExecutorFromContext(ctx, r.pool)

 query, args, err := sq.
  Insert("trip_status_history").
  Columns("trip_id", "from_status", "to_status", "reason").
  Values(tripID, fromStatus, toStatus, reason).
  PlaceholderFormat(sq.Dollar).
  ToSql()
 if err != nil {
  return fmt.Errorf("repository: failed to build insert query: %w", err)
 }

 if _, err := exec.Exec(ctx, query, args...); err != nil {
  return fmt.Errorf("repository: failed to insert status history: %w", err)
 }

 return nil
}
