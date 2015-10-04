package main

import (
	"time"

	"github.com/go-kit/kit/log"

	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
)

type loggingMiddleware struct {
	logger log.Logger
	booking.Service
}

func (mw loggingMiddleware) BookNewCargo(origin location.UNLocode, destination location.UNLocode, arrivalDeadline time.Time) (id cargo.TrackingID, err error) {
	id, err = mw.Service.BookNewCargo(origin, destination, arrivalDeadline)
	_ = mw.logger.Log(
		"method", "book",
		"err", err,
		"origin", origin,
		"destination", destination,
		"arrival_deadline", arrivalDeadline,
		"tracking_id", id,
	)
	return
}
