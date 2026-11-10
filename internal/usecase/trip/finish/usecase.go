package finish

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
)

type Usecase struct {
	trips   tripRepository
	history historyRepository
	tx      txManager
}

func New(trips tripRepository, history historyRepository, tx txManager) *Usecase {
	return &Usecase{trips: trips, history: history, tx: tx}
}

func (u *Usecase) Execute(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	now := time.Now().UTC()
	sid := id.String()
	err := u.tx.Do(ctx, func(ctx context.Context) error {
		trip, err := u.trips.GetByID(ctx, sid)
		if err != nil {
			return err
		}
		if err := trip.Finish(now); err != nil {
			return err
		}
		if err := u.trips.Finish(ctx, sid, now); err != nil {
			return err
		}
		active := domain.TripStatusActive
		return u.history.Create(ctx, sid, &active, domain.TripStatusCompleted, "trip finished")
	})
	if err != nil {
		return domain.Trip{}, err
	}
	return u.trips.GetByID(ctx, sid)
}
