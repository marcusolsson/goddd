package routing

import "bitbucket.org/marcus_olsson/goddd/cargo"

type RoutingService interface {
	FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary
}

// Our end of the routing service. This is basically a data model
// translation layer between our domain model and the API put forward
// by the routing team, which operates in a different context from us.
type externalRoutingService struct{}

func (s *externalRoutingService) FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary {
	// TODO: Port pathfinder
	return []cargo.Itinerary{}
}

func NewRoutingService() RoutingService {
	return &externalRoutingService{}
}
