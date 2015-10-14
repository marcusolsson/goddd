// Package booking provides the use-case of booking a cargo. Used by views
// facing an administrator.
package booking

import (
	"errors"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"
)

// Service is the interface that provides booking methods.
type Service interface {
	// BookNewCargo registers a new cargo in the tracking system, not yet
	// routed.
	BookNewCargo(origin location.UNLocode, destination location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingID, error)

	// RequestPossibleRoutesForCargo requests a list of itineraries describing
	// possible routes for this cargo.
	RequestPossibleRoutesForCargo(trackingID cargo.TrackingID) []cargo.Itinerary

	// AssignCargoToRoute assigns a cargo to the route specified by the
	// itinerary.
	AssignCargoToRoute(trackingID cargo.TrackingID, itinerary cargo.Itinerary) error

	// ChangeDestination changes the destination of a cargo.
	ChangeDestination(trackingID cargo.TrackingID, unLocode location.UNLocode) error

	// Cargos returns a list of all cargos that have been booked.
	Cargos() []Cargo

	// Locations returns a list of registered locations.
	Locations() []Location
}

type service struct {
	cargoRepository         cargo.Repository
	locationRepository      location.Repository
	routingService          routing.Service
	handlingEventRepository cargo.HandlingEventRepository
}

func (s *service) AssignCargoToRoute(trackingID cargo.TrackingID, itinerary cargo.Itinerary) error {
	var err error

	c, err := s.cargoRepository.Find(trackingID)
	if err != nil {
		return err
	}

	c.AssignToRoute(itinerary)

	if err := s.cargoRepository.Store(c); err != nil {
		return err
	}

	return nil
}

func (s *service) BookNewCargo(origin, destination location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingID, error) {
	trackingID := cargo.NextTrackingID()
	routeSpecification := cargo.RouteSpecification{
		Origin:          origin,
		Destination:     destination,
		ArrivalDeadline: arrivalDeadline,
	}

	c := cargo.New(trackingID, routeSpecification)

	s.cargoRepository.Store(*c)

	return c.TrackingID, nil
}

func (s *service) ChangeDestination(trackingID cargo.TrackingID, destination location.UNLocode) error {
	c, err := s.cargoRepository.Find(trackingID)
	if err != nil {
		return errors.New("Could not find cargo.")
	}

	l, err := s.locationRepository.Find(destination)
	if err != nil {
		return err
	}

	routeSpecification := cargo.RouteSpecification{
		Origin:          c.Origin,
		Destination:     l.UNLocode,
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
	}

	c.SpecifyNewRoute(routeSpecification)

	if err := s.cargoRepository.Store(c); err != nil {
		return err
	}

	return nil
}

func (s *service) RequestPossibleRoutesForCargo(trackingID cargo.TrackingID) []cargo.Itinerary {
	c, err := s.cargoRepository.Find(trackingID)
	if err != nil {
		return []cargo.Itinerary{}
	}

	return s.routingService.FetchRoutesForSpecification(c.RouteSpecification)
}

func (s *service) Cargos() []Cargo {
	cargos := s.cargoRepository.FindAll()
	var result []Cargo
	for _, c := range cargos {
		result = append(result, assemble(c, s.handlingEventRepository))
	}
	return result
}

func (s *service) Locations() []Location {
	var result []Location
	for _, v := range s.locationRepository.FindAll() {
		result = append(result, Location{
			UNLocode: string(v.UNLocode),
			Name:     v.Name,
		})
	}
	return result
}

// NewService creates a booking service with necessary dependencies.
func NewService(cr cargo.Repository, lr location.Repository, her cargo.HandlingEventRepository, rs routing.Service) Service {
	return &service{
		cargoRepository:         cr,
		locationRepository:      lr,
		handlingEventRepository: her,
		routingService:          rs,
	}
}

// Location is a read model for booking views.
type Location struct {
	UNLocode string `json:"locode"`
	Name     string `json:"name"`
}

// Cargo is a read model for booking views.
type Cargo struct {
	ArrivalDeadline time.Time `json:"arrivalDeadline"`
	Destination     string    `json:"destination"`
	Legs            []Leg     `json:"legs,omitempty"`
	Misrouted       bool      `json:"misrouted"`
	Origin          string    `json:"origin"`
	Routed          bool      `json:"routed"`
	TrackingID      string    `json:"trackingId"`
}

// Leg is a read model for booking views.
type Leg struct {
	VoyageNumber string    `json:"voyageNumber"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	LoadTime     time.Time `json:"loadTime"`
	UnloadTime   time.Time `json:"unloadTime"`
}

func assemble(c cargo.Cargo, her cargo.HandlingEventRepository) Cargo {
	return Cargo{
		TrackingID:      string(c.TrackingID),
		Origin:          string(c.Origin),
		Destination:     string(c.RouteSpecification.Destination),
		Misrouted:       c.Delivery.RoutingStatus == cargo.Misrouted,
		Routed:          !c.Itinerary.IsEmpty(),
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
		Legs:            assembleLegs(c),
	}
}

func assembleLegs(c cargo.Cargo) []Leg {
	var legs []Leg
	for _, l := range c.Itinerary.Legs {
		legs = append(legs, Leg{
			VoyageNumber: string(l.VoyageNumber),
			From:         string(l.LoadLocation),
			To:           string(l.UnloadLocation),
			LoadTime:     l.LoadTime,
			UnloadTime:   l.UnloadTime,
		})
	}
	return legs
}
