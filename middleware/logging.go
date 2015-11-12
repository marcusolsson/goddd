package middleware

import (
	"time"

	"github.com/go-kit/kit/log"

	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/tracking"
)

// LoggingBookingService adds logging to the booking service.
type LoggingBookingService struct {
	Logger log.Logger
	booking.Service
}

// BookNewCargo logs the booking request.
func (mw LoggingBookingService) BookNewCargo(origin location.UNLocode, destination location.UNLocode, arrivalDeadline time.Time) (id cargo.TrackingID, err error) {
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

// ChangeDestination logs the change destination request.
func (mw LoggingBookingService) ChangeDestination(id cargo.TrackingID, l location.UNLocode) (err error) {
	err = mw.Service.ChangeDestination(id, l)
	_ = mw.Logger.Log(
		"method", "change_destination",
		"err", err,
		"tracking_id", id,
		"destination", l,
	)
	return
}

// LoggingTrackingService adds logging to the tracking service.
type LoggingTrackingService struct {
	Logger log.Logger
	tracking.Service
}

// Track logs the track request.
func (mw LoggingTrackingService) Track(id string) (c tracking.Cargo, err error) {
	c, err = mw.Service.Track(id)
	_ = mw.Logger.Log(
		"method", "track",
		"err", err,
		"tracking_id", id,
	)
	return
}
