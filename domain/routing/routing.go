package routing

import (
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
)

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
