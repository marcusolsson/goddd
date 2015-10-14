package routing

import (
	"time"

	"github.com/marcusolsson/goddd/cargo"
)

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

// Route is a read model for routing views.
type Route struct {
	Legs []Leg `json:"legs"`
}

// Leg is a read model for routing views.
type Leg struct {
	VoyageNumber string    `json:"voyageNumber"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	LoadTime     time.Time `json:"loadTime"`
	UnloadTime   time.Time `json:"unloadTime"`
}
