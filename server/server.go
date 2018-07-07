package server

import (
	"net/http"

	kitlog "github.com/go-kit/kit/log"
	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/tracking"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server holds the dependencies for a HTTP server.
type Server struct {
	Booking  booking.Service
	Tracking tracking.Service
	Handling handling.Service

	Logger kitlog.Logger

	mux *http.ServeMux
}

// New returns a new HTTP server.
func New(bs booking.Service, ts tracking.Service, hs handling.Service, logger kitlog.Logger) *Server {
	s := &Server{
		Booking:  bs,
		Tracking: ts,
		Handling: hs,
		Logger:   logger,
	}

	mux := http.NewServeMux()

	mux.Handle("/booking/v1/", makeBookingHandler(s.Booking, s.Logger))
	mux.Handle("/tracking/v1/", makeTrackingHandler(s.Tracking, s.Logger))
	mux.Handle("/handling/v1/", makeHandlingHandler(s.Handling, s.Logger))
	mux.Handle("/metrics", promhttp.Handler())

	mux.Handle("/", accessControl(mux))

	s.mux = mux

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
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
