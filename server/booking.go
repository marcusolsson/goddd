package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	kitlog "github.com/go-kit/kit/log"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/booking"
)

type bookingHandler struct {
	s booking.Service

	logger kitlog.Logger
}

func (h *bookingHandler) router() chi.Router {
	r := chi.NewRouter()

	r.Route("/cargos", func(r chi.Router) {
		r.Post("/", h.bookCargo)
		r.Get("/", h.listCargos)
		r.Route("/{trackingID}", func(r chi.Router) {
			r.Get("/", h.loadCargo)
			r.Get("/request_routes", h.requestRoutes)
			r.Post("/assign_to_route", h.assignToRoute)
			r.Post("/change_destination", h.changeDestination)
		})

	})
	r.Get("/locations", h.listLocations)

	r.Method("GET", "/docs", http.StripPrefix("/booking/v1/docs", http.FileServer(http.Dir("booking/docs"))))

	return r
}

func (h *bookingHandler) bookCargo(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var request struct {
		Origin          shipping.UNLocode
		Destination     shipping.UNLocode
		ArrivalDeadline time.Time
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Log("error", err)
		encodeError(ctx, err, w)
		return
	}

	id, err := h.s.BookNewCargo(request.Origin, request.Destination, request.ArrivalDeadline)
	if err != nil {
		encodeError(ctx, err, w)
		return
	}

	var response = struct {
		ID shipping.TrackingID `json:"tracking_id"`
	}{
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Log("error", err)
		encodeError(ctx, err, w)
		return
	}
}

func (h *bookingHandler) loadCargo(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	trackingID := shipping.TrackingID(chi.URLParam(r, "trackingID"))

	c, err := h.s.LoadCargo(trackingID)
	if err != nil {
		encodeError(ctx, err, w)
		return
	}

	var response = struct {
		Cargo booking.Cargo `json:"cargo"`
	}{
		Cargo: c,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Log("error", err)
		encodeError(ctx, err, w)
		return
	}
}

func (h *bookingHandler) requestRoutes(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	trackingID := shipping.TrackingID(chi.URLParam(r, "trackingID"))

	itin := h.s.RequestPossibleRoutesForCargo(trackingID)

	var response = struct {
		Routes []shipping.Itinerary `json:"routes"`
	}{
		Routes: itin,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Log("error", err)
		encodeError(ctx, err, w)
		return
	}
}

func (h *bookingHandler) assignToRoute(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	trackingID := shipping.TrackingID(chi.URLParam(r, "trackingID"))

	var request struct {
		Itinerary shipping.Itinerary `json:"route"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Log("error", err)
		encodeError(ctx, err, w)
		return
	}

	err := h.s.AssignCargoToRoute(trackingID, request.Itinerary)
	if err != nil {
		encodeError(ctx, err, w)
		return
	}
}

func (h *bookingHandler) changeDestination(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	trackingID := shipping.TrackingID(chi.URLParam(r, "trackingID"))

	var request struct {
		Destination shipping.UNLocode `json:"destination"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Log("error", err)
		encodeError(ctx, err, w)
		return
	}

	err := h.s.ChangeDestination(trackingID, request.Destination)
	if err != nil {
		encodeError(ctx, err, w)
		return
	}
}

func (h *bookingHandler) listCargos(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	cs := h.s.Cargos()

	var response = struct {
		Cargos []booking.Cargo `json:"cargos"`
	}{
		Cargos: cs,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Log("error", err)
		encodeError(ctx, err, w)
		return
	}
}

func (h *bookingHandler) listLocations(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	ls := h.s.Locations()

	var response = struct {
		Locations []booking.Location `json:"locations"`
	}{
		Locations: ls,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Log("error", err)
		encodeError(ctx, err, w)
		return
	}
}
