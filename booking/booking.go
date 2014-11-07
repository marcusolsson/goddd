package booking

import (
	"errors"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"
)

type BookingService interface {
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

type bookingService struct {
	cargoRepository    cargo.Repository
	locationRepository location.Repository
	routingService     routing.RoutingService
}

func (s *bookingService) BookNewCargo(originLocode location.UNLocode, destinationLocode location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingID, error) {

	trackingID := cargo.NextTrackingID()
	routeSpecification := cargo.RouteSpecification{
		Origin:          originLocode,
		Destination:     destinationLocode,
		ArrivalDeadline: arrivalDeadline,
	}

	c := cargo.NewCargo(trackingID, routeSpecification)

	s.cargoRepository.Store(*c)

	return c.TrackingID, nil
}

func (s *bookingService) RequestPossibleRoutesForCargo(trackingID cargo.TrackingID) []cargo.Itinerary {
	c, err := s.cargoRepository.Find(trackingID)

	if err != nil {
		return []cargo.Itinerary{}
	}

	return s.routingService.FetchRoutesForSpecification(c.RouteSpecification)
}

func (s *bookingService) AssignCargoToRoute(itinerary cargo.Itinerary, trackingID cargo.TrackingID) error {
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

func (s *bookingService) ChangeDestination(trackingID cargo.TrackingID, unLocode location.UNLocode) error {
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

// NewBookingService creates a booking service with necessary dependencies.
func NewBookingService(cr cargo.Repository, lr location.Repository, rs routing.RoutingService) BookingService {
	return &bookingService{
		cargoRepository:    cr,
		locationRepository: lr,
		routingService:     rs,
	}
}
