package create

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
	"github.com/Lebedeva2006/go-labs/internal/model"
)

type Usecase struct {
	trips   tripRepository
	history historyRepository
	tx      txManager
}

func New(trips tripRepository, history historyRepository, tx txManager) *Usecase {
	return &Usecase{trips: trips, history: history, tx: tx}
}

func (u *Usecase) Execute(ctx context.Context, in model.CreateTripInput) (domain.Trip, error) {
	trip := domain.Trip{
		ID: uuid.New(), UserID: in.UserID, DriverID: in.DriverID,
		StartPoint: in.StartPoint, EndPoint: in.EndPoint, Price: in.Price,
		Status: domain.TripStatusActive, StartedAt: time.Now().UTC(),
	}
	if err := trip.Validate(); err != nil {
		return domain.Trip{}, err
	}
	err := u.tx.Do(ctx, func(ctx context.Context) error {
		if err := u.trips.Create(ctx, trip); err != nil {
			return err
		}
		return u.history.Create(ctx, trip.ID.String(), nil, domain.TripStatusActive, "trip created")
	})
	if err != nil {
		return domain.Trip{}, err
	}
	return trip, nil
}