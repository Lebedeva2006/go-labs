package create

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Lebedeva2006/go-labs/internal/domain/trip"
	api "github.com/Lebedeva2006/go-labs/internal/generated"
	"github.com/Lebedeva2006/go-labs/internal/http/response"
	"github.com/Lebedeva2006/go-labs/internal/http/trip/convert"
)

type Handler struct {
	uc  usecase
	log *slog.Logger
}

func New(uc usecase, log *slog.Logger) *Handler { return &Handler{uc: uc, log: log} }

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	var body api.TripData
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		response.Error(w, r, h.log, &trip.ValidationError{Field: "body", Reason: "malformed JSON or unknown fields"})
		return
	}
	t, err := h.uc.Execute(r.Context(), toInput(body))
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	w.Header().Set("Location", "/api/v1/trips/"+t.ID.String())
	response.JSON(w, http.StatusCreated, convert.ToAPITrip(t))
}