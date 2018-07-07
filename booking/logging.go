package booking

import (
	"time"

	"github.com/go-kit/kit/log"

	shipping "github.com/marcusolsson/goddd"
)

type loggingService struct {
	logger log.Logger
	next   Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) BookNewCargo(origin shipping.UNLocode, destination shipping.UNLocode, deadline time.Time) (id shipping.TrackingID, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "book",
			"origin", origin,
			"destination", destination,
			"arrival_deadline", deadline,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.BookNewCargo(origin, destination, deadline)
}

func (s *loggingService) LoadCargo(id shipping.TrackingID) (c Cargo, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "load",
			"tracking_id", id,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.LoadCargo(id)
}

func (s *loggingService) RequestPossibleRoutesForCargo(id shipping.TrackingID) []shipping.Itinerary {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "request_routes",
			"tracking_id", id,
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.next.RequestPossibleRoutesForCargo(id)
}

func (s *loggingService) AssignCargoToRoute(id shipping.TrackingID, itinerary shipping.Itinerary) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "assign_to_route",
			"tracking_id", id,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.AssignCargoToRoute(id, itinerary)
}

func (s *loggingService) ChangeDestination(id shipping.TrackingID, l shipping.UNLocode) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "change_destination",
			"tracking_id", id,
			"destination", l,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.ChangeDestination(id, l)
}

func (s *loggingService) Cargos() []Cargo {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "list_cargos",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.next.Cargos()
}

func (s *loggingService) Locations() []Location {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "list_locations",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.next.Locations()
}
