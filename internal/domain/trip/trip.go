package trip

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TripStatus string

const (
	TripStatusActive    TripStatus = "active"
	TripStatusCompleted TripStatus = "completed"
)

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type Trip struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartPoint     Coordinates
	EndPoint       Coordinates
	Price          int64
	Status         TripStatus
	StartedAt      time.Time
	FinishedAt     *time.Time
	LastPositionAt *time.Time
}

func (c Coordinates) Validate() error {
	if c.Latitude < -90 || c.Latitude > 90 {
		return &ValidationError{Field: "latitude", Reason: "must be between -90 and 90"}
	}
	if c.Longitude < -180 || c.Longitude > 180 {
		return &ValidationError{Field: "longitude", Reason: "must be between -180 and 180"}
	}
	return nil
}

func (t Trip) Validate() error {
	if t.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Reason: "must not be empty"}
	}
	if t.DriverID == uuid.Nil {
		return &ValidationError{Field: "driver_id", Reason: "must not be empty"}
	}
	if err := t.StartPoint.Validate(); err != nil {
		return fmt.Errorf("start_point: %w", err)
	}
	if err := t.EndPoint.Validate(); err != nil {
		return fmt.Errorf("end_point: %w", err)
	}
	if t.Price < 0 {
		return &ValidationError{Field: "price", Reason: "must not be negative"}
	}
	return nil
}

func (t *Trip) Finish(at time.Time) error {
	if t.Status != TripStatusActive {
		return ErrTripCompleted
	}
	t.Status = TripStatusCompleted
	t.FinishedAt = &at
	return nil
}
