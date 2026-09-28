package handler

import (
 "context"
 "encoding/json"
 "log"
 "net/http"
 "time"

 "github.com/google/uuid"

 api "github.com/Lebedeva2006/go-labs/internal/generated"
 "github.com/Lebedeva2006/go-labs/internal/db"
 "github.com/Lebedeva2006/go-labs/internal/domain"
 "github.com/Lebedeva2006/go-labs/internal/repository"
)

type Server struct {
 Pool         PingPool
 TxManager    db.TxManager
 TripRepo     *repository.TripRepository
 HistoryRepo  *repository.TripStatusHistoryRepository
 QueryTimeout time.Duration
}

type PingPool interface {
 Ping(ctx context.Context, timeout time.Duration) error
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
 w.Header().Set("Content-Type", "application/json")
 w.WriteHeader(http.StatusOK)
 w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) Ready(w http.ResponseWriter, r *http.Request) {
 w.Header().Set("Content-Type", "application/json")
 if err := s.Pool.Ping(r.Context(), s.QueryTimeout); err != nil {
  log.Printf("readiness check failed: %v", err)
  w.WriteHeader(http.StatusServiceUnavailable)
  w.Write([]byte(`{"status":"unavailable"}`))
  return
 }
 w.WriteHeader(http.StatusOK)
 w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
 var body api.TripData
 decoder := json.NewDecoder(r.Body)
 decoder.DisallowUnknownFields()
 if err := decoder.Decode(&body); err != nil {
  writeProblem(w, r, http.StatusBadRequest, "invalid_request", "Invalid request", "Malformed JSON body or unknown fields")
  return
 }

 if msg := validateTripData(body); msg != "" {
  writeProblem(w, r, http.StatusBadRequest, "invalid_request", "Invalid request", msg)
  return
 }

 trip := domain.Trip{
  ID:         uuid.New(),
  UserID:     body.UserId,
  DriverID:   body.DriverId,
  StartPoint: toDomainCoordinates(body.StartPoint),
  EndPoint:   toDomainCoordinates(body.EndPoint),
  Price:      body.Price,
  Status:     domain.TripStatusActive,
  StartedAt:  time.Now().UTC(),
 }

 err := s.TxManager.Do(r.Context(), func(ctx context.Context) error {
  if err := s.TripRepo.Create(ctx, trip); err != nil {
   return err
  }
  return s.HistoryRepo.Create(ctx, trip.ID.String(), nil, domain.TripStatusActive, "trip created")
 })
 if err != nil {
  writeDomainError(w, r, err)
  return
 }

 w.Header().Set("Content-Type", "application/json")
 w.Header().Set("Location", "/api/v1/trips/"+trip.ID.String())
 w.WriteHeader(http.StatusCreated)
 json.NewEncoder(w).Encode(toAPITrip(trip))
}

func (s *Server) GetTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
 trip, err := s.TripRepo.GetByID(r.Context(), tripId.String())
 if err != nil {
  writeDomainError(w, r, err)
  return
 }

 w.Header().Set("Content-Type", "application/json")
 w.WriteHeader(http.StatusOK)
 json.NewEncoder(w).Encode(toAPITrip(trip))
}

func (s *Server) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
 now := time.Now().UTC()

 err := s.TxManager.Do(r.Context(), func(ctx context.Context) error {
  if err := s.TripRepo.Finish(ctx, tripId.String(), now); err != nil {
   return err
  }
  active := domain.TripStatusActive
  return s.HistoryRepo.Create(ctx, tripId.String(), &active, domain.TripStatusCompleted, "trip finished")
 })
 if err != nil {
  writeDomainError(w, r, err)
  return
 }

 trip, err := s.TripRepo.GetByID(r.Context(), tripId.String())
 if err != nil {
  writeDomainError(w, r, err)
  return
 }

 w.Header().Set("Content-Type", "application/json")
 w.WriteHeader(http.StatusOK)
 json.NewEncoder(w).Encode(toAPITrip(trip))
}
