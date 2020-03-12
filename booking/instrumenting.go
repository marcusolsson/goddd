package booking

import (
	"time"

	"github.com/go-kit/kit/metrics"

	shipping "github.com/marcusolsson/goddd"
)

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

// NewInstrumentingMiddleware returns an instance of an instrumenting middleware.
func NewInstrumentingMiddleware(counter metrics.Counter, latency metrics.Histogram, next Service) Service {
	return &instrumentingMiddleware{
		requestCount:   counter,
		requestLatency: latency,
		next:           next,
	}
}

func (mw *instrumentingMiddleware) BookNewCargo(origin, destination shipping.UNLocode, deadline time.Time) (shipping.TrackingID, error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "book").Add(1)
		mw.requestLatency.With("method", "book").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.BookNewCargo(origin, destination, deadline)
}

func (mw *instrumentingMiddleware) LoadCargo(id shipping.TrackingID) (c Cargo, err error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "load").Add(1)
		mw.requestLatency.With("method", "load").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.LoadCargo(id)
}

func (mw *instrumentingMiddleware) RequestPossibleRoutesForCargo(id shipping.TrackingID) []shipping.Itinerary {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "request_routes").Add(1)
		mw.requestLatency.With("method", "request_routes").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.RequestPossibleRoutesForCargo(id)
}

func (mw *instrumentingMiddleware) AssignCargoToRoute(id shipping.TrackingID, itinerary shipping.Itinerary) (err error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "assign_to_route").Add(1)
		mw.requestLatency.With("method", "assign_to_route").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.AssignCargoToRoute(id, itinerary)
}

func (mw *instrumentingMiddleware) ChangeDestination(id shipping.TrackingID, l shipping.UNLocode) (err error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "change_destination").Add(1)
		mw.requestLatency.With("method", "change_destination").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.ChangeDestination(id, l)
}

func (mw *instrumentingMiddleware) Cargos() []Cargo {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "list_cargos").Add(1)
		mw.requestLatency.With("method", "list_cargos").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.Cargos()
}

func (mw *instrumentingMiddleware) Locations() []Location {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "list_locations").Add(1)
		mw.requestLatency.With("method", "list_locations").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.Locations()
}
