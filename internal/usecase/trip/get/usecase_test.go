package get_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
	"github.com/Lebedeva2006/go-labs/internal/usecase/trip/get"
)

func TestExecute(t *testing.T) {
	id := uuid.New()
	trip := domain.Trip{ID: id}

	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{"success", nil, nil},
		{"not found", domain.ErrTripNotFound, domain.ErrTripNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			trips := NewMocktripRepository(ctrl)

			trips.EXPECT().GetByID(gomock.Any(), id.String()).Return(trip, tt.repoErr)

			uc := get.New(trips)
			got, err := uc.Execute(context.Background(), id)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got.ID != id {
				t.Errorf("got trip ID %v, want %v", got.ID, id)
			}
		})
	}
}
