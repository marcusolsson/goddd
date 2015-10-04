package routing

import "github.com/marcusolsson/goddd/cargo"

var DefaultService Service = &service{}

// Service provides access to an external routing service.
type Service interface {
	// FetchRoutesForSpecification finds all possible routes that satisfy a
	// given specification.
	FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary
}

type service struct {
}

func (s *service) FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary {
	return nil
}
