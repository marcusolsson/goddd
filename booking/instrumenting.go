package booking

import (
	"time"

	"github.com/go-kit/kit/metrics"

	shipping "github.com/marcusolsson/goddd"
)

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

// NewInstrumentingService returns an instance of an instrumenting Service.
func NewInstrumentingService(counter metrics.Counter, latency metrics.Histogram, s Service) Service {
	return &instrumentingService{
		requestCount:   counter,
		requestLatency: latency,
		next:           s,
	}
}

func (s *instrumentingService) BookNewCargo(origin, destination shipping.UNLocode, deadline time.Time) (shipping.TrackingID, error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "book").Add(1)
		s.requestLatency.With("method", "book").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.BookNewCargo(origin, destination, deadline)
}

func (s *instrumentingService) LoadCargo(id shipping.TrackingID) (c Cargo, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "load").Add(1)
		s.requestLatency.With("method", "load").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.LoadCargo(id)
}

func (s *instrumentingService) RequestPossibleRoutesForCargo(id shipping.TrackingID) []shipping.Itinerary {
	defer func(begin time.Time) {
		s.requestCount.With("method", "request_routes").Add(1)
		s.requestLatency.With("method", "request_routes").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.RequestPossibleRoutesForCargo(id)
}

func (s *instrumentingService) AssignCargoToRoute(id shipping.TrackingID, itinerary shipping.Itinerary) (err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "assign_to_route").Add(1)
		s.requestLatency.With("method", "assign_to_route").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.AssignCargoToRoute(id, itinerary)
}

func (s *instrumentingService) ChangeDestination(id shipping.TrackingID, l shipping.UNLocode) (err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "change_destination").Add(1)
		s.requestLatency.With("method", "change_destination").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.ChangeDestination(id, l)
}

func (s *instrumentingService) Cargos() []Cargo {
	defer func(begin time.Time) {
		s.requestCount.With("method", "list_cargos").Add(1)
		s.requestLatency.With("method", "list_cargos").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Cargos()
}

func (s *instrumentingService) Locations() []Location {
	defer func(begin time.Time) {
		s.requestCount.With("method", "list_locations").Add(1)
		s.requestLatency.With("method", "list_locations").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Locations()
}
