package middleware

import (
	"time"

	"github.com/go-kit/kit/log"

	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/tracking"
)

// LoggingBooking adds logging to the booking service.
type LoggingBooking struct {
	Logger log.Logger
	booking.Service
}

// BookNewCargo logs the booking request.
func (mw LoggingBooking) BookNewCargo(origin location.UNLocode, destination location.UNLocode, arrivalDeadline time.Time) (id cargo.TrackingID, err error) {
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

// LoggingTracking adds logging to the tracking service.
type LoggingTracking struct {
	Logger log.Logger
	tracking.Service
}

// Track logs the track request.
func (mw LoggingTracking) Track(id string) (c tracking.Cargo, err error) {
	c, err = mw.Service.Track(id)
	_ = mw.Logger.Log(
		"method", "track",
		"err", err,
		"tracking_id", id,
	)
	return
}
