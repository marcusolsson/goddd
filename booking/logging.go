package booking

import (
	"time"

	"github.com/go-kit/kit/log"

	shipping "github.com/marcusolsson/goddd"
)

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

// NewLoggingMiddleware returns a new instance of a logging middleware.
func NewLoggingMiddleware(logger log.Logger, next Service) Service {
	return &loggingMiddleware{
		logger: logger,
		next:   next,
	}
}

func (mw *loggingMiddleware) BookNewCargo(origin shipping.UNLocode, destination shipping.UNLocode, deadline time.Time) (id shipping.TrackingID, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "book",
			"origin", origin,
			"destination", destination,
			"arrival_deadline", deadline,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.BookNewCargo(origin, destination, deadline)
}

func (mw *loggingMiddleware) LoadCargo(id shipping.TrackingID) (c Cargo, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "load",
			"tracking_id", id,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.LoadCargo(id)
}

func (mw *loggingMiddleware) RequestPossibleRoutesForCargo(id shipping.TrackingID) []shipping.Itinerary {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "request_routes",
			"tracking_id", id,
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.RequestPossibleRoutesForCargo(id)
}

func (mw *loggingMiddleware) AssignCargoToRoute(id shipping.TrackingID, itinerary shipping.Itinerary) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "assign_to_route",
			"tracking_id", id,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.AssignCargoToRoute(id, itinerary)
}

func (mw *loggingMiddleware) ChangeDestination(id shipping.TrackingID, l shipping.UNLocode) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "change_destination",
			"tracking_id", id,
			"destination", l,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.ChangeDestination(id, l)
}

func (mw *loggingMiddleware) Cargos() []Cargo {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "list_cargos",
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.Cargos()
}

func (mw *loggingMiddleware) Locations() []Location {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "list_locations",
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.Locations()
}
