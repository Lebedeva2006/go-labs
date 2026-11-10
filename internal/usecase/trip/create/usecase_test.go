package create_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
	"github.com/Lebedeva2006/go-labs/internal/model"
	"github.com/Lebedeva2006/go-labs/internal/usecase/trip/create"
)

func runTx(_ context.Context, fn func(context.Context) error) error { return fn(context.Background()) }

func input() model.CreateTripInput {
	return model.CreateTripInput{
		UserID: uuid.New(), DriverID: uuid.New(),
		StartPoint: domain.Coordinates{Latitude: 59.9, Longitude: 30.3},
		EndPoint:   domain.Coordinates{Latitude: 59.8, Longitude: 30.4}, Price: 100,
	}
}

func TestExecute(t *testing.T) {
	boom := errors.New("db down")
	tests := []struct {
		name    string
		in      func() model.CreateTripInput
		repoErr error
		useRepo bool
		wantErr error
	}{
		{"success", input, nil, true, nil},
		{"driver busy", input, domain.ErrDriverBusy, true, domain.ErrDriverBusy},
		{"repo error", input, boom, true, boom},
		{"invalid input", func() model.CreateTripInput { i := input(); i.Price = -1; return i }, nil, false, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			trips := NewMocktripRepository(ctrl)
			hist := NewMockhistoryRepository(ctrl)
			tx := NewMocktxManager(ctrl)
			if tt.useRepo {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(runTx)
				trips.EXPECT().Create(gomock.Any(), gomock.Any()).Return(tt.repoErr)
				if tt.repoErr == nil {
					hist.EXPECT().Create(gomock.Any(), gomock.Any(), nil, domain.TripStatusActive, gomock.Any()).Return(nil)
				}
			}
			_, err := create.New(trips, hist, tx).Execute(context.Background(), tt.in())
			if tt.name == "invalid input" {
				var ve *domain.ValidationError
				if !errors.As(err, &ve) {
					t.Fatalf("want ValidationError, got %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}
