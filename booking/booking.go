package booking

import (
	"time"

	"bitbucket.org/marcus_olsson/goddd/cargo"
	"bitbucket.org/marcus_olsson/goddd/location"
	"bitbucket.org/marcus_olsson/goddd/routing"
)

// Cargo booking service.
type BookingService interface {
	// BookNewCargo registers a new cargo in the tracking system,
	// not yet routed.
	BookNewCargo(origin location.UNLocode, destination location.UNLocode, arrivalDeadline time.Time) cargo.TrackingId
	// Requests a list of itineraries describing possible routes
	// for this cargo.
	RequestPossibleRoutesForCargo(trackingId cargo.TrackingId) []cargo.Itinerary
	// Assigns the cargo to a route.
	AssignCargoToRoute(itinerary cargo.Itinerary, trackingId cargo.TrackingId)
	// Changes the destination of a cargo.
	ChangeDestination(trackingId cargo.TrackingId, unLocode location.UNLocode)
}

type bookingService struct {
	cargoRepository    cargo.CargoRepository
	locationRepository location.LocationRepository
	routingService     routing.RoutingService
}

func (s *bookingService) BookNewCargo(originLocode location.UNLocode, destinationLocode location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingId, error) {

	trackingId, err := cargo.NewTrackingId()
	if err != nil {
		return "", err
	}
	origin := s.locationRepository.Find(originLocode)
	destination := s.locationRepository.Find(destinationLocode)

	routeSpecification := cargo.RouteSpecification{
		Origin:          origin,
		Destination:     destination,
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
