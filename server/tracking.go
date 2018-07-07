package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	kitlog "github.com/go-kit/kit/log"

	"github.com/marcusolsson/goddd/tracking"
)

type trackingHandler struct {
	s tracking.Service

	logger kitlog.Logger
}

func (h *trackingHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Get("/cargos/{trackingID}", h.track)
	r.Method("GET", "/docs", http.StripPrefix("/tracking/v1/docs", http.FileServer(http.Dir("tracking/docs"))))
	return r
}

func (h *trackingHandler) track(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	trackingID := chi.URLParam(r, "trackingID")

	c, err := h.s.Track(trackingID)
	if err != nil {
		encodeError(ctx, err, w)
		return
	}

	var response = struct {
		Cargo *tracking.Cargo `json:"cargo"`
	}{Cargo: &c}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Log("error", err)
		encodeError(ctx, err, w)
		return
	}
}
