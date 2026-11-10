package finish_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
	"github.com/Lebedeva2006/go-labs/internal/usecase/trip/finish"
)

const testSuccess = "success"

func runTx(_ context.Context, fn func(context.Context) error) error { return fn(context.Background()) }

func TestExecute(t *testing.T) {
	id := uuid.New()

	activeTrip := domain.Trip{ID: id, Status: domain.TripStatusActive}
	completedTrip := domain.Trip{ID: id, Status: domain.TripStatusCompleted}

	tests := []struct {
		name    string
		repoErr error
	}{
		{testSuccess, nil},
		{"already completed", domain.ErrTripCompleted},
		{"not found", domain.ErrTripNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			trips := NewMocktripRepository(ctrl)
			hist := NewMockhistoryRepository(ctrl)
			tx := NewMocktxManager(ctrl)
			m := NewMockmetrics(ctrl)

			m.EXPECT().TripCompleted(gomock.Any()).Return().AnyTimes()

			tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(runTx)

			if tt.name == testSuccess {
				gomock.InOrder(
					trips.EXPECT().GetByID(gomock.Any(), id.String()).Return(activeTrip, nil),
					trips.EXPECT().Finish(gomock.Any(), id.String(), gomock.Any()).Return(nil),
					hist.EXPECT().Create(gomock.Any(), id.String(), gomock.Any(), domain.TripStatusCompleted, gomock.Any()).Return(nil),
					trips.EXPECT().GetByID(gomock.Any(), id.String()).Return(completedTrip, nil),
				)
			} else {
				if tt.repoErr == domain.ErrTripCompleted {
					trips.EXPECT().GetByID(gomock.Any(), id.String()).Return(completedTrip, nil)
				} else {
					trips.EXPECT().GetByID(gomock.Any(), id.String()).Return(domain.Trip{}, tt.repoErr)
				}
			}

			uc := finish.New(trips, hist, tx, m)
			_, err := uc.Execute(context.Background(), id)

			if tt.name == "success" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if !errors.Is(err, tt.repoErr) {
					t.Fatalf("got error %v, want %v", err, tt.repoErr)
				}
			}
		})
	}
}
