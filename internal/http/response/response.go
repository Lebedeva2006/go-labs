package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Lebedeva2006/go-labs/internal/http/domainerror"
	trip "github.com/Lebedeva2006/go-labs/internal/domain/trip"
)

func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func BadRequest(w http.ResponseWriter, r *http.Request, detail string) {
	Error(w, r, slog.Default(), &trip.ValidationError{Field: "request", Reason: detail})
}

func Error(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	status, p := domainerror.Problem(err)
	if status == http.StatusInternalServerError {
		log.ErrorContext(r.Context(), "internal error", "error", err)
	}
	inst := r.URL.Path
	p.Instance = &inst
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}