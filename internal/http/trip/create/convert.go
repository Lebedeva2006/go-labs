package create

import (
	api "github.com/Lebedeva2006/go-labs/internal/generated"
	"github.com/Lebedeva2006/go-labs/internal/http/trip/convert"
	"github.com/Lebedeva2006/go-labs/internal/model"
)

func toInput(body api.TripData) model.CreateTripInput {
	return model.CreateTripInput{
		UserID:     body.UserId,
		DriverID:   body.DriverId,
		StartPoint: convert.ToDomainCoordinates(body.StartPoint),
		EndPoint:   convert.ToDomainCoordinates(body.EndPoint),
		Price:      body.Price,
	}
}
