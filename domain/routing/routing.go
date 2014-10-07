package routing

import "github.com/marcusolsson/goddd/domain/cargo"

type RoutingService interface {
	FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary
}
