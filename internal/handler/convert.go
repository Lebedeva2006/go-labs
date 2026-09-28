package handler

import (
 api "github.com/Lebedeva2006/go-labs/internal/generated"
 "github.com/Lebedeva2006/go-labs/internal/domain"
)

func toDomainCoordinates(c api.Coordinates) domain.Coordinates {
 return domain.Coordinates{Latitude: c.Latitude, Longitude: c.Longitude}
}

func toAPICoordinates(c domain.Coordinates) api.Coordinates {
 return api.Coordinates{Latitude: c.Latitude, Longitude: c.Longitude}
}

func toAPITrip(t domain.Trip) api.Trip {
 trip := api.Trip{
  Id:         t.ID,
  UserId:     t.UserID,
  DriverId:   t.DriverID,
  StartPoint: toAPICoordinates(t.StartPoint),
  EndPoint:   toAPICoordinates(t.EndPoint),
  Price:      t.Price,
  Status:     api.TripStatus(t.Status),
  StartedAt:  t.StartedAt,
  FinishedAt: t.FinishedAt,
 }
 trip.LastPositionAt = t.LastPositionAt
 return trip
}
