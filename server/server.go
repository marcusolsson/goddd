package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	kitlog "github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/tracking"
)

// Server holds the dependencies for a HTTP server.
type Server struct {
	Booking  booking.Service
	Tracking tracking.Service
	Handling handling.Service

	
	
	
	Logger kitlog.Logger

	router chi.Router
}

// New returns a new HTTP server.
func New(bs booking.Service, ts tracking.Service, hs handling.Service, logger kitlog.Logger) *Server {
	s := &Server{
		Booking:  bs,
		Tracking: ts,
		Handling: hs,
		Logger:   logger,
	}

	r := chi.NewRouter()

	r.Use(accessControl)

	r.Route("/booking", func(r chi.Router) {
		h := bookingHandler{s.Booking, s.Logger}
		r.Mount("/v1", h.router())
	})
	r.Route("/tracking", func(r chi.Router) {
		h := trackingHandler{s.Tracking, s.Logger}
		r.Mount("/v1", h.router())
	})
	r.Route("/handling", func(r chi.Router) {
		h := handlingHandler{s.Handling, s.Logger}
		r.Mount("/v1", h.router())
	})

	r.Method("GET", "/metrics", promhttp.Handler())

	s.router = r

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case shipping.ErrUnknownCargo:
		w.WriteHeader(http.StatusNotFound)
	case tracking.ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
