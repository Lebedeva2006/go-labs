package handler

import (
 "github.com/google/uuid"

 api "github.com/Lebedeva2006/go-labs/internal/generated"
)

func validateCoordinates(c api.Coordinates) bool {
 return c.Latitude >= -90 && c.Latitude <= 90 &&
  c.Longitude >= -180 && c.Longitude <= 180
}

func validateTripData(body api.TripData) string {
 if body.UserId == uuid.Nil {
  return "user_id must not be empty"
 }
 if body.DriverId == uuid.Nil {
  return "driver_id must not be empty"
 }
 if !validateCoordinates(body.StartPoint) {
  return "start_point coordinates out of range"
 }
 if !validateCoordinates(body.EndPoint) {
  return "end_point coordinates out of range"
 }
 if body.Price < 0 {
  return "price must not be negative"
 }
 return ""
}
