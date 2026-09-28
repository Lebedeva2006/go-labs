package handler

import (
 "encoding/json"
 "log"
 "net/http"

 "github.com/Lebedeva2006/go-labs/internal/domain"
)

type problem struct {
 Type     string `json:"type"`
 Title    string `json:"title"`
 Status   int    `json:"status"`
 Detail   string `json:"detail"`
 Instance string `json:"instance"`
 Code     string `json:"code"`
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, title, detail string) {
 w.Header().Set("Content-Type", "application/problem+json")
 w.WriteHeader(status)
 _ = json.NewEncoder(w).Encode(problem{
  Type:     "https://tripgo.example/problems/" + code,
  Title:    title,
  Status:   status,
  Detail:   detail,
  Instance: r.URL.Path,
  Code:     code,
 })
}

func writeDomainError(w http.ResponseWriter, r *http.Request, err error) {
 switch {
 case err == domain.ErrTripNotFound:
  writeProblem(w, r, http.StatusNotFound, "trip_not_found", "Trip not found", "Trip was not found")
 case err == domain.ErrTripCompleted:
  writeProblem(w, r, http.StatusConflict, "trip_completed", "Trip completed", "Operation is not allowed for a completed trip")
 case err == domain.ErrDriverBusy:
  writeProblem(w, r, http.StatusConflict, "driver_busy", "Driver busy", "Driver already has an active trip")
 default:
  log.Printf("internal error: %v", err)
  writeProblem(w, r, http.StatusInternalServerError, "internal_error", "Internal Server Error", "Internal server error")
 }
}

func WriteBadRequest(w http.ResponseWriter, r *http.Request, detail string) {
 writeProblem(w, r, http.StatusBadRequest, "invalid_request", "Invalid request", detail)
}
