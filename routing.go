package shipping

// RoutingService is a domain service for routing cargos.
type RoutingService interface {
	// FetchRoutesForSpecification finds all possible routes that satisfy a
	// given specification.
	FetchRoutesForSpecification(rs RouteSpecification) []Itinerary
}
