package trip

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTrip_Finish(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		status  TripStatus
		wantErr error
	}{
		{"active trip finishes", TripStatusActive, nil},
		{"completed trip cannot finish again", TripStatusCompleted, ErrTripCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trip := Trip{Status: tt.status}

			err := trip.Finish(now)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if trip.Status != TripStatusCompleted {
					t.Errorf("status = %s, want completed", trip.Status)
				}
				if trip.FinishedAt == nil || !trip.FinishedAt.Equal(now) {
					t.Errorf("FinishedAt not set correctly")
				}
			}
		})
	}
}

func TestCoordinates_Validate(t *testing.T) {
	tests := []struct {
		name    string
		coords  Coordinates
		wantErr bool
	}{
		{"valid coordinates", Coordinates{Latitude: 40.7128, Longitude: -74.0060}, false},
		{"invalid latitude above max", Coordinates{Latitude: 91, Longitude: -74.0060}, true},
		{"invalid latitude below min", Coordinates{Latitude: -91, Longitude: -74.0060}, true},
		{"invalid longitude above max", Coordinates{Latitude: 40.7128, Longitude: 181}, true},
		{"invalid longitude below min", Coordinates{Latitude: 40.7128, Longitude: -181}, true},
		{"latitude 90", Coordinates{Latitude: 90, Longitude: -74.0060}, false},
		{"latitude -90", Coordinates{Latitude: -90, Longitude: -74.0060}, false},
		{"longitude 180", Coordinates{Latitude: 44.0, Longitude: 180}, false},
		{"longitude -180", Coordinates{Latitude: 44.0, Longitude: -180}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.coords.Validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func validTrip() Trip {
	return Trip{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		DriverID:   uuid.New(),
		StartPoint: Coordinates{Latitude: 59.9398, Longitude: 30.3146},
		EndPoint:   Coordinates{Latitude: 59.929, Longitude: 30.3626},
		Price:      1450,
		Status:     TripStatusActive,
	}
}

func TestTrip_Validate(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Trip)
		wantField string
	}{
		{"valid trip", func(tr *Trip) {}, ""},
		{"empty user id", func(tr *Trip) { tr.UserID = uuid.Nil }, "user_id"},
		{"empty driver id", func(tr *Trip) { tr.DriverID = uuid.Nil }, "driver_id"},
		{"negative price", func(tr *Trip) { tr.Price = -1 }, "price"},
		{"zero price is allowed", func(tr *Trip) { tr.Price = 0 }, ""},
		{"start latitude out of range", func(tr *Trip) { tr.StartPoint.Latitude = 100 }, "latitude"},
		{"end longitude out of range", func(tr *Trip) { tr.EndPoint.Longitude = 200 }, "longitude"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trip := validTrip()
			tt.mutate(&trip)

			err := trip.Validate()

			if tt.wantField == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected *ValidationError, got %v", err)
			}
			if ve.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", ve.Field, tt.wantField)
			}
		})
	}
}
