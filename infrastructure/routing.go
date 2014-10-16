package infrastructure

import (
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/routing"
)

type externalRoutingService struct {
	locationRepository location.Repository
}

func (s *externalRoutingService) FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary {

	it := cargo.Itinerary{Legs: []cargo.Leg{cargo.Leg{
		LoadLocation:   routeSpecification.Origin,
		UnloadLocation: routeSpecification.Destination,
	}}}

	return []cargo.Itinerary{it}
}

func NewExternalRoutingService(locationRepository location.Repository) routing.RoutingService {
	return &externalRoutingService{locationRepository: locationRepository}
}
