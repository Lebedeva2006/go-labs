package get

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/Lebedeva2006/go-labs/internal/http/response"
	"github.com/Lebedeva2006/go-labs/internal/http/trip/convert"
)

type Handler struct {
	uc  usecase
	log *slog.Logger
}

func New(uc usecase, log *slog.Logger) *Handler { return &Handler{uc: uc, log: log} }

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	t, err := h.uc.Execute(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.JSON(w, http.StatusOK, convert.ToAPITrip(t))
}