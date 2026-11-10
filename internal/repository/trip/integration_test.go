//go:build integration

package trip_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
	"github.com/Lebedeva2006/go-labs/internal/infra/postgres/transaction"
	repository "github.com/Lebedeva2006/go-labs/internal/repository/trip"
)

func setup(t *testing.T) (*pgxpool.Pool, *repository.TripRepository, *repository.TripStatusHistoryRepository, transaction.TxManager) {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool,
		repository.NewTripRepository(pool, 5*time.Second),
		repository.NewTripStatusHistoryRepository(pool, 5*time.Second),
		transaction.NewTxManager(pool)
}

func newTrip() domain.Trip {
	return domain.Trip{
		ID: uuid.New(), UserID: uuid.New(), DriverID: uuid.New(),
		StartPoint: domain.Coordinates{Latitude: 59.9, Longitude: 30.3},
		EndPoint:   domain.Coordinates{Latitude: 59.8, Longitude: 30.4},
		Price:      100, Status: domain.TripStatusActive, StartedAt: time.Now().UTC(),
	}
}

func cleanup(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM trips WHERE id = $1`, id)
	})
}

func TestCreateAndGet(t *testing.T) {
	pool, trips, _, _ := setup(t)
	tr := newTrip()
	cleanup(t, pool, tr.ID)

	if err := trips.Create(context.Background(), tr); err != nil {
		t.Fatal(err)
	}
	got, err := trips.GetByID(context.Background(), tr.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != tr.ID || got.Price != tr.Price || got.Status != domain.TripStatusActive {
		t.Fatalf("unexpected trip: %+v", got)
	}
}

func TestGetNotFound(t *testing.T) {
	_, trips, _, _ := setup(t)
	_, err := trips.GetByID(context.Background(), uuid.NewString())
	if !errors.Is(err, domain.ErrTripNotFound) {
		t.Fatalf("got %v, want ErrTripNotFound", err)
	}
}

func TestDriverBusy(t *testing.T) {
	pool, trips, _, _ := setup(t)
	first := newTrip()
	second := newTrip()
	second.DriverID = first.DriverID
	cleanup(t, pool, first.ID)
	cleanup(t, pool, second.ID)

	if err := trips.Create(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if err := trips.Create(context.Background(), second); !errors.Is(err, domain.ErrDriverBusy) {
		t.Fatalf("got %v, want ErrDriverBusy", err)
	}
}

func TestFinish(t *testing.T) {
	pool, trips, _, _ := setup(t)
	tr := newTrip()
	cleanup(t, pool, tr.ID)
	ctx := context.Background()

	if err := trips.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}
	if err := trips.Finish(ctx, tr.ID.String(), time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := trips.Finish(ctx, tr.ID.String(), time.Now().UTC()); !errors.Is(err, domain.ErrTripCompleted) {
		t.Fatalf("got %v, want ErrTripCompleted", err)
	}
	if err := trips.Finish(ctx, uuid.NewString(), time.Now().UTC()); !errors.Is(err, domain.ErrTripNotFound) {
		t.Fatalf("got %v, want ErrTripNotFound", err)
	}
}

func TestTransactionRollback(t *testing.T) {
	pool, trips, history, tx := setup(t)
	tr := newTrip()
	cleanup(t, pool, tr.ID)
	boom := errors.New("boom")

	err := tx.Do(context.Background(), func(ctx context.Context) error {
		if err := trips.Create(ctx, tr); err != nil {
			return err
		}
		if err := history.Create(ctx, tr.ID.String(), nil, domain.TripStatusActive, "created"); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("got %v, want boom", err)
	}

	var trCount, histCount int
	_ = pool.QueryRow(context.Background(), `SELECT count(*) FROM trips WHERE id=$1`, tr.ID).Scan(&trCount)
	_ = pool.QueryRow(context.Background(), `SELECT count(*) FROM trip_status_history WHERE trip_id=$1`, tr.ID).Scan(&histCount)
	if trCount != 0 || histCount != 0 {
		t.Fatalf("rollback failed: trips=%d history=%d", trCount, histCount)
	}
}
