package convert

import (
    domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
    api    "github.com/Lebedeva2006/go-labs/internal/generated"
)

func ToDomainCoordinates(c api.Coordinates) domain.Coordinates {
	return domain.Coordinates{Latitude: c.Latitude, Longitude: c.Longitude}
}

func ToAPICoordinates(c domain.Coordinates) api.Coordinates {
	return api.Coordinates{Latitude: c.Latitude, Longitude: c.Longitude}
}

func ToAPITrip(t domain.Trip) api.Trip {
	trip := api.Trip{
		Id:         t.ID,
		UserId:     t.UserID,
		DriverId:   t.DriverID,
		StartPoint: ToAPICoordinates(t.StartPoint),
		EndPoint:   ToAPICoordinates(t.EndPoint),
		Price:      t.Price,
		Status:     api.TripStatus(t.Status),
		StartedAt:  t.StartedAt,
		FinishedAt: t.FinishedAt,
	}
	trip.LastPositionAt = t.LastPositionAt
	return trip
}
