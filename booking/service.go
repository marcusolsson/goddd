// Package booking provides the use-case of booking a shipping. Used by views
// facing an administrator.
package booking

import (
	"errors"
	"time"

	shipping "github.com/marcusolsson/goddd"
)

// ErrInvalidArgument is returned when one or more arguments are invalid.
var ErrInvalidArgument = errors.New("invalid argument")

// Service is the interface that provides booking methods.
type Service interface {
	// BookNewCargo registers a new cargo in the tracking system, not yet
	// routed.
	BookNewCargo(origin shipping.UNLocode, destination shipping.UNLocode, deadline time.Time) (shipping.TrackingID, error)

	// LoadCargo returns a read model of a shipping.
	LoadCargo(id shipping.TrackingID) (Cargo, error)

	// RequestPossibleRoutesForCargo requests a list of itineraries describing
	// possible routes for this shipping.
	RequestPossibleRoutesForCargo(id shipping.TrackingID) []shipping.Itinerary

	// AssignCargoToRoute assigns a cargo to the route specified by the
	// itinerary.
	AssignCargoToRoute(id shipping.TrackingID, itinerary shipping.Itinerary) error

	// ChangeDestination changes the destination of a shipping.
	ChangeDestination(id shipping.TrackingID, destination shipping.UNLocode) error

	// Cargos returns a list of all cargos that have been booked.
	Cargos() []Cargo

	// Locations returns a list of registered locations.
	Locations() []Location
}

type service struct {
	cargos         shipping.CargoRepository
	locations      shipping.LocationRepository
	handlingEvents shipping.HandlingEventRepository
	routingService shipping.RoutingService
}

// NewService creates a booking service with necessary dependencies.
func NewService(cargos shipping.CargoRepository, locations shipping.LocationRepository, events shipping.HandlingEventRepository, rs shipping.RoutingService) Service {
	return &service{
		cargos:         cargos,
		locations:      locations,
		handlingEvents: events,
		routingService: rs,
	}
}

func (s *service) AssignCargoToRoute(id shipping.TrackingID, itinerary shipping.Itinerary) error {
	if id == "" || len(itinerary.Legs) == 0 {
		return ErrInvalidArgument
	}

	c, err := s.cargos.Find(id)
	if err != nil {
		return err
	}

	c.AssignToRoute(itinerary)

	return s.cargos.Store(c)
}

func (s *service) BookNewCargo(origin, destination shipping.UNLocode, deadline time.Time) (shipping.TrackingID, error) {
	if origin == "" || destination == "" || deadline.IsZero() {
		return "", ErrInvalidArgument
	}

	id := shipping.MakeTrackingID()
	rs := shipping.RouteSpecification{
		Origin:          origin,
		Destination:     destination,
		ArrivalDeadline: deadline,
	}

	c := shipping.NewCargo(id, rs)

	if err := s.cargos.Store(c); err != nil {
		return "", err
	}

	return c.TrackingID, nil
}

func (s *service) LoadCargo(id shipping.TrackingID) (Cargo, error) {
	if id == "" {
		return Cargo{}, ErrInvalidArgument
	}

	c, err := s.cargos.Find(id)
	if err != nil {
		return Cargo{}, err
	}

	return assemble(c, s.handlingEvents), nil
}

func (s *service) ChangeDestination(id shipping.TrackingID, destination shipping.UNLocode) error {
	if id == "" || destination == "" {
		return ErrInvalidArgument
	}

	c, err := s.cargos.Find(id)
	if err != nil {
		return err
	}

	l, err := s.locations.Find(destination)
	if err != nil {
		return err
	}

	c.SpecifyNewRoute(shipping.RouteSpecification{
		Origin:          c.Origin,
		Destination:     l.UNLocode,
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
	})

	if err := s.cargos.Store(c); err != nil {
		return err
	}

	return nil
}

func (s *service) RequestPossibleRoutesForCargo(id shipping.TrackingID) []shipping.Itinerary {
	if id == "" {
		return nil
	}

	c, err := s.cargos.Find(id)
	if err != nil {
		return []shipping.Itinerary{}
	}

	return s.routingService.FetchRoutesForSpecification(c.RouteSpecification)
}

func (s *service) Cargos() []Cargo {
	cargos := s.cargos.FindAll()
	result := make([]Cargo, 0, len(cargos))
	for _, c := range cargos {
		result = append(result, assemble(c, s.handlingEvents))
	}
	return result
}

func (s *service) Locations() []Location {
	locations := s.locations.FindAll()
	result := make([]Location, 0, len(locations))
	for _, v := range locations {
		result = append(result, Location{
			UNLocode: string(v.UNLocode),
			Name:     v.Name,
		})
	}
	return result
}

// Location is a read model for booking views.
type Location struct {
	UNLocode string `json:"locode"`
	Name     string `json:"name"`
}

// Cargo is a read model for booking views.
type Cargo struct {
	ArrivalDeadline time.Time      `json:"arrival_deadline"`
	Destination     string         `json:"destination"`
	Legs            []shipping.Leg `json:"legs,omitempty"`
	Misrouted       bool           `json:"misrouted"`
	Origin          string         `json:"origin"`
	Routed          bool           `json:"routed"`
	TrackingID      string         `json:"tracking_id"`
}

func assemble(c *shipping.Cargo, events shipping.HandlingEventRepository) Cargo {
	return Cargo{
		TrackingID:      string(c.TrackingID),
		Origin:          string(c.Origin),
		Destination:     string(c.RouteSpecification.Destination),
		Misrouted:       c.Delivery.RoutingStatus == shipping.Misrouted,
		Routed:          !c.Itinerary.IsEmpty(),
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
		Legs:            c.Itinerary.Legs,
	}
}
