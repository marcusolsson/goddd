package booking

import (
	"errors"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"
)

type Service interface {
	// BookNewCargo registers a new cargo in the tracking system, not yet
	// routed.
	BookNewCargo(origin location.UNLocode, destination location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingID, error)

	// RequestPossibleRoutesForCargo requests a list of itineraries describing
	// possible routes for this cargo.
	RequestPossibleRoutesForCargo(trackingID cargo.TrackingID) []cargo.Itinerary

	// AssignCargoToRoute assigns a cargo to the route specified by the
	// itinerary.
	AssignCargoToRoute(itinerary cargo.Itinerary, trackingID cargo.TrackingID) error

	// ChangeDestination changes the destination of a cargo.
	ChangeDestination(trackingID cargo.TrackingID, unLocode location.UNLocode) error
}

type service struct {
	cargoRepository    cargo.Repository
	locationRepository location.Repository
	routingService     routing.Service
}

func (s *service) BookNewCargo(originLocode location.UNLocode, destinationLocode location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingID, error) {

	trackingID := cargo.NextTrackingID()
	routeSpecification := cargo.RouteSpecification{
		Origin:          originLocode,
		Destination:     destinationLocode,
		ArrivalDeadline: arrivalDeadline,
	}

	c := cargo.New(trackingID, routeSpecification)

	s.cargoRepository.Store(*c)

	return c.TrackingID, nil
}

func (s *service) RequestPossibleRoutesForCargo(trackingID cargo.TrackingID) []cargo.Itinerary {
	c, err := s.cargoRepository.Find(trackingID)

	if err != nil {
		return []cargo.Itinerary{}
	}

	return s.routingService.FetchRoutesForSpecification(c.RouteSpecification)
}

func (s *service) AssignCargoToRoute(itinerary cargo.Itinerary, trackingID cargo.TrackingID) error {
	var err error

	c, err := s.cargoRepository.Find(trackingID)

	if err != nil {
		return err
	}

	c.AssignToRoute(itinerary)

	if err := s.cargoRepository.Store(c); err != nil {
		return err
	}

	return nil
}

func (s *service) ChangeDestination(trackingID cargo.TrackingID, unLocode location.UNLocode) error {
	c, err := s.cargoRepository.Find(trackingID)

	if err != nil {
		return errors.New("Could not find cargo.")
	}

	l, err := s.locationRepository.Find(unLocode)

	if err != nil {
		return err
	}

	routeSpecification := cargo.RouteSpecification{
		Origin:          c.Origin,
		Destination:     l.UNLocode,
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
	}

	c.SpecifyNewRoute(routeSpecification)

	if err := s.cargoRepository.Store(c); err != nil {
		return err
	}

	return nil
}

// NewService creates a booking service with necessary dependencies.
func NewService(cr cargo.Repository, lr location.Repository, rs routing.Service) Service {
	return &service{
		cargoRepository:    cr,
		locationRepository: lr,
		routingService:     rs,
	}
}
