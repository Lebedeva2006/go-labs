package create_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
	"github.com/Lebedeva2006/go-labs/internal/http/trip/create"
)

func TestCreateHandler(t *testing.T) {
	validJSON := `{
		"user_id": "550e8400-e29b-41d4-a716-446655440000",
		"driver_id": "550e8400-e29b-41d4-a716-446655440001",
		"start_point": {"latitude": 59.9, "longitude": 30.3},
		"end_point": {"latitude": 59.8, "longitude": 30.4},
		"price": 100
	}`

	tests := []struct {
		name       string
		body       string
		mockSetup  func(*Mockusecase)
		wantStatus int
		wantCode   string
	}{
		{
			name: "success",
			body: validJSON,
			mockSetup: func(m *Mockusecase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(domain.Trip{}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "driver busy",
			body: validJSON,
			mockSetup: func(m *Mockusecase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(domain.Trip{}, domain.ErrDriverBusy)
			},
			wantStatus: http.StatusConflict,
			wantCode:   "driver_busy",
		},
		{
			name: "malformed JSON",
			body: `{"user_id": "broken`,
			mockSetup: func(m *Mockusecase) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_request",
		},
		{
			name: "unknown fields",
			body: `{"unknown": "field"}`,
			mockSetup: func(m *Mockusecase) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockUC := NewMockusecase(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}
			h := create.New(mockUC, nil)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.Handle(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d. Body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}

			if tt.wantCode != "" {
				if !strings.Contains(rr.Body.String(), `"code":"`+tt.wantCode+`"`) {
					t.Errorf("body does not contain code %q: %s", tt.wantCode, rr.Body.String())
				}
			}
		})
	}
}
