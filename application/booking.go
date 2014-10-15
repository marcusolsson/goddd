package application

import (
	"errors"
	"time"

	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/routing"
)

type BookingService interface {
	// BookNewCargo registers a new cargo in the tracking system, not yet
	// routed.
	BookNewCargo(origin location.UNLocode, destination location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingId, error)

	// RequestPossibleRoutesForCargo requests a list of itineraries describing
	// possible routes for this cargo.
	RequestPossibleRoutesForCargo(trackingId cargo.TrackingId) []cargo.Itinerary

	// AssignCargoToRoute assigns a cargo to the route specified by the
	// itinerary.
	AssignCargoToRoute(itinerary cargo.Itinerary, trackingId cargo.TrackingId) error

	// ChangeDestination changes the destination of a cargo.
	ChangeDestination(trackingId cargo.TrackingId, unLocode location.UNLocode) error
}

type bookingService struct {
	cargoRepository    cargo.CargoRepository
	locationRepository location.LocationRepository
	routingService     routing.RoutingService
}

func (s *bookingService) BookNewCargo(originLocode location.UNLocode, destinationLocode location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingId, error) {

	trackingId := cargo.NextTrackingId()
	routeSpecification := cargo.RouteSpecification{
		Origin:          originLocode,
		Destination:     destinationLocode,
		ArrivalDeadline: arrivalDeadline,
	}

	c := cargo.NewCargo(trackingId, routeSpecification)

	s.cargoRepository.Store(*c)

	return c.TrackingId, nil
}

func (s *bookingService) RequestPossibleRoutesForCargo(trackingId cargo.TrackingId) []cargo.Itinerary {
	c, err := s.cargoRepository.Find(trackingId)

	if err != nil {
		return []cargo.Itinerary{}
	}

	return s.routingService.FetchRoutesForSpecification(c.RouteSpecification)
}

func (s *bookingService) AssignCargoToRoute(itinerary cargo.Itinerary, trackingId cargo.TrackingId) error {
	var err error

	c, err := s.cargoRepository.Find(trackingId)

	if err != nil {
		return err
	}

	c.AssignToRoute(itinerary)

	if err := s.cargoRepository.Store(c); err != nil {
		return err
	}

	return nil
}

func (s *bookingService) ChangeDestination(trackingId cargo.TrackingId, unLocode location.UNLocode) error {
	c, err := s.cargoRepository.Find(trackingId)

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
func NewBookingService(cr cargo.CargoRepository, lr location.LocationRepository, rs routing.RoutingService) BookingService {
	return &bookingService{
		cargoRepository:    cr,
		locationRepository: lr,
		routingService:     rs,
	}
}
