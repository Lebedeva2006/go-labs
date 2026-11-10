package domainerror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
	"github.com/Lebedeva2006/go-labs/internal/http/domainerror"
)

func TestProblem(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "validation error",
			err:        &domain.ValidationError{Field: "price", Reason: "must be positive"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_request",
		},
		{
			name:       "trip not found",
			err:        domain.ErrTripNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   "trip_not_found",
		},
		{
			name:       "trip not found wrapped",
			err:        fmt.Errorf("get trip: %w", domain.ErrTripNotFound),
			wantStatus: http.StatusNotFound,
			wantCode:   "trip_not_found",
		},
		{
			name:       "trip completed",
			err:        domain.ErrTripCompleted,
			wantStatus: http.StatusConflict,
			wantCode:   "trip_completed",
		},
		{
			name:       "driver busy",
			err:        domain.ErrDriverBusy,
			wantStatus: http.StatusConflict,
			wantCode:   "driver_busy",
		},
		{
			name:       "unknown error",
			err:        errors.New("db connection lost"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, problem := domainerror.Problem(tt.err)
			if status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
			if problem.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", problem.Code, tt.wantCode)
			}

			if tt.wantStatus == http.StatusInternalServerError {
				if problem.Detail != nil && *problem.Detail == tt.err.Error() {
					t.Errorf("internal error details leaked to response")
				}
			}
		})
	}
}
