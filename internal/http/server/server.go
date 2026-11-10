package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	api "github.com/Lebedeva2006/go-labs/internal/generated"
)

type (
	Handler       interface{ Handle(http.ResponseWriter, *http.Request) }
	IDHandler     interface{ Handle(http.ResponseWriter, *http.Request, uuid.UUID) }
	PingPool      interface{ Ping(ctx context.Context, timeout time.Duration) error }
)

type Server struct {
	Create       Handler
	Get          IDHandler
	Finish       IDHandler
	Pool         PingPool
	QueryTimeout time.Duration
	Log          *slog.Logger
}

func (s *Server) CreateTrip(w http.ResponseWriter, r *http.Request, _ api.CreateTripParams) {
	s.Create.Handle(w, r)
}
func (s *Server) GetTrip(w http.ResponseWriter, r *http.Request, id api.TripId) { s.Get.Handle(w, r, id) }
func (s *Server) FinishTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	s.Finish.Handle(w, r, id)
}

func (s *Server) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) Ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := s.Pool.Ping(r.Context(), s.QueryTimeout); err != nil {
		s.Log.ErrorContext(r.Context(), "readiness check failed", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"unavailable"}`))
		return
	}
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}