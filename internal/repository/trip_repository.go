package repository

import (
 "context"
 "errors"
 "fmt"

 sq "github.com/Masterminds/squirrel"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgconn"
 "github.com/jackc/pgx/v5/pgxpool"

 "github.com/Lebedeva2006/go-labs/internal/db"
 "github.com/Lebedeva2006/go-labs/internal/domain"
)

const uniqueViolationCode = "23505"


type TripRepository struct {
 pool *pgxpool.Pool
}

func NewTripRepository(pool *pgxpool.Pool) *TripRepository {
 return &TripRepository{pool: pool}
}


func (r *TripRepository) Create(ctx context.Context, trip domain.Trip) error {
 exec := db.ExecutorFromContext(ctx, r.pool)

 query, args, err := sq.
  Insert("trips").
  Columns(
   "id", "user_id", "driver_id",
   "start_latitude", "start_longitude",
   "end_latitude", "end_longitude",
   "price", "status", "started_at",
  ).
  Values(
   trip.ID, trip.UserID, trip.DriverID,
   trip.StartPoint.Latitude, trip.StartPoint.Longitude,
   trip.EndPoint.Latitude, trip.EndPoint.Longitude,
   trip.Price, trip.Status, trip.StartedAt,
  ).
  PlaceholderFormat(sq.Dollar).
  ToSql()
 if err != nil {
  return fmt.Errorf("repository: failed to build insert query: %w", err)
 }

 if _, err := exec.Exec(ctx, query, args...); err != nil {
  var pgErr *pgconn.PgError
  if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
   return domain.ErrDriverBusy
  }
  return fmt.Errorf("repository: failed to insert trip: %w", err)
 }

 return nil
}


func (r *TripRepository) GetByID(ctx context.Context, id string) (domain.Trip, error) {
 exec := db.ExecutorFromContext(ctx, r.pool)

 query, args, err := sq.
  Select(
   "id", "user_id", "driver_id",
   "start_latitude", "start_longitude",
   "end_latitude", "end_longitude",
   "price", "status", "started_at", "finished_at", "last_position_at",
  ).
  From("trips").
  Where(sq.Eq{"id": id}).
  PlaceholderFormat(sq.Dollar).
  ToSql()
 if err != nil {
  return domain.Trip{}, fmt.Errorf("repository: failed to build select query: %w", err)
 }

 var trip domain.Trip
 row := exec.QueryRow(ctx, query, args...)
 err = row.Scan(
  &trip.ID, &trip.UserID, &trip.DriverID,
  &trip.StartPoint.Latitude, &trip.StartPoint.Longitude,
  &trip.EndPoint.Latitude, &trip.EndPoint.Longitude,
  &trip.Price, &trip.Status, &trip.StartedAt, &trip.FinishedAt, &trip.LastPositionAt,
 )
 if err != nil {
  if errors.Is(err, pgx.ErrNoRows) {
   return domain.Trip{}, domain.ErrTripNotFound
  }
  return domain.Trip{}, fmt.Errorf("repository: failed to scan trip: %w", err)
 }

 return trip, nil
}


func (r *TripRepository) Finish(ctx context.Context, id string, finishedAt interface{}) error {
 exec := db.ExecutorFromContext(ctx, r.pool)

 query, args, err := sq.
  Update("trips").
  Set("status", domain.TripStatusCompleted).
  Set("finished_at", finishedAt).
  Set("updated_at", sq.Expr("now()")).
  Where(sq.Eq{"id": id, "status": domain.TripStatusActive}).
  PlaceholderFormat(sq.Dollar).
  ToSql()
 if err != nil {
  return fmt.Errorf("repository: failed to build update query: %w", err)
 }

 tag, err := exec.Exec(ctx, query, args...)
 if err != nil {
  return fmt.Errorf("repository: failed to update trip: %w", err)
 }

 if tag.RowsAffected() == 0 {
   _, getErr := r.GetByID(ctx, id)
  if getErr != nil {
   return getErr
  }
  return domain.ErrTripCompleted
 }

 return nil
}
