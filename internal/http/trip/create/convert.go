package create

import (
	"github.com/google/uuid"
	api "github.com/Lebedeva2006/go-labs/internal/generated"
	"github.com/Lebedeva2006/go-labs/internal/model"
	"github.com/Lebedeva2006/go-labs/internal/http/trip/convert" 
)

func toInput(body api.TripData) model.CreateTripInput {
	return model.CreateTripInput{
		UserID:     uuid.UUID(body.UserId),
		DriverID:   uuid.UUID(body.DriverId),
		StartPoint: convert.ToDomainCoordinates(body.StartPoint),
		EndPoint:   convert.ToDomainCoordinates(body.EndPoint),
		Price:      body.Price,
	}
}