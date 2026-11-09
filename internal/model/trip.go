package model

import (
	"github.com/google/uuid"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
)

type CreateTripInput struct {
	UserID, DriverID     uuid.UUID
	StartPoint, EndPoint domain.Coordinates
	Price                int64
}