package middleware

import (
	"time"

	"github.com/go-kit/kit/log"

	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
)

// LoggingMiddleware adds logging to the booking service.
type LoggingMiddleware struct {
	Logger log.Logger
	booking.Service
}

// BookNewCargo logs the booking request.
func (mw LoggingMiddleware) BookNewCargo(origin location.UNLocode, destination location.UNLocode, arrivalDeadline time.Time) (id cargo.TrackingID, err error) {
	id, err = mw.Service.BookNewCargo(origin, destination, arrivalDeadline)
	_ = mw.Logger.Log(
		"method", "book",
		"err", err,
		"origin", origin,
		"destination", destination,
		"arrival_deadline", arrivalDeadline,
		"tracking_id", id,
	)
	return
}
