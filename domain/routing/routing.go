package routing

import (
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
)

// Our end of the routing service. This is basically a data model
// translation layer between our domain model and the API put forward
// by the routing team, which operates in a different context from us.
type RoutingService interface {
	FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary
}

type externalRoutingService struct {
	locationRepository location.LocationRepository
}

func (s *externalRoutingService) FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary {

	it := cargo.Itinerary{Legs: []cargo.Leg{cargo.Leg{
		LoadLocation:   routeSpecification.Origin,
		UnloadLocation: routeSpecification.Destination,
	}}}

	return []cargo.Itinerary{it}
}

func NewRoutingService(locationRepository location.LocationRepository) RoutingService {
	return &externalRoutingService{locationRepository: locationRepository}
}
