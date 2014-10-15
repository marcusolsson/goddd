package routing

import "github.com/marcusolsson/goddd/domain/cargo"

// RoutingService is a provides access to an external routing service.
type RoutingService interface {
	// FetchRoutesForSpecification finds all possible routes that satisfy a
	// given specification.
	FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary
}
