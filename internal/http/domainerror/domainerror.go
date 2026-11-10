package domainerror

import (
	"errors"
	"net/http"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
	api "github.com/Lebedeva2006/go-labs/internal/generated"
)

func Problem(err error) (int, api.Problem) {
	var ve *domain.ValidationError
	switch {
	case errors.As(err, &ve):
		return build(http.StatusBadRequest, "invalid_request", "Invalid request", ve.Error())
	case errors.Is(err, domain.ErrTripNotFound):
		return build(http.StatusNotFound, "trip_not_found", "Trip not found", "Trip was not found")
	case errors.Is(err, domain.ErrTripCompleted):
		return build(http.StatusConflict, "trip_completed", "Trip completed", "Operation is not allowed for a completed trip")
	case errors.Is(err, domain.ErrDriverBusy):
		return build(http.StatusConflict, "driver_busy", "Driver busy", "Driver already has an active trip")
	default:
		return build(http.StatusInternalServerError, "internal_error", "Internal Server Error", "Internal server error")
	}
}

func build(status int, code, title, detail string) (int, api.Problem) {
	return status, api.Problem{
		Type:   "https://tripgo.example/problems/" + code,
		Title:  title,
		Status: int32(status), //nolint:gosec // HTTP-статусы всегда в пределах int32
		Detail: &detail,
		Code:   code,
	}
}
