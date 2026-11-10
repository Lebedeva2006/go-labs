package finish_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
	"github.com/Lebedeva2006/go-labs/internal/http/trip/finish"
)

func TestFinishHandler(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name       string
		mockSetup  func(*Mockusecase)
		wantStatus int
		wantCode   string
	}{
		{
			name: "success",
			mockSetup: func(m *Mockusecase) {
				m.EXPECT().Execute(gomock.Any(), id).Return(domain.Trip{ID: id, Status: domain.TripStatusCompleted}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "trip completed",
			mockSetup: func(m *Mockusecase) {
				m.EXPECT().Execute(gomock.Any(), id).Return(domain.Trip{}, domain.ErrTripCompleted)
			},
			wantStatus: http.StatusConflict,
			wantCode:   "trip_completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockUC := NewMockusecase(ctrl)
			tt.mockSetup(mockUC)

			h := finish.New(mockUC, nil)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+id.String()+"/finish", nil)
			rr := httptest.NewRecorder()

			h.Handle(rr, req, id)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d. Body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}

			if tt.wantCode != "" && !strings.Contains(rr.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Errorf("body does not contain code %q: %s", tt.wantCode, rr.Body.String())
			}
		})
	}
}