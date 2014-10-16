package cargo

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"

	"code.google.com/p/go-uuid/uuid"
)

// TrackingId uniquely identifies a particular cargo.
type TrackingId string

// Cargo is the central class in the domain model.
type Cargo struct {
	TrackingId         TrackingId
	Origin             location.UNLocode
	RouteSpecification RouteSpecification
	Itinerary          Itinerary
	Delivery           Delivery
}

// SpecifyNewRoute specifies a new route for this cargo.
func (c *Cargo) SpecifyNewRoute(routeSpecification RouteSpecification) {
	c.RouteSpecification = routeSpecification
	c.Delivery = c.Delivery.UpdateOnRouting(c.RouteSpecification, c.Itinerary)
}

// AssignToRoute attaches a new itinerary to this cargo.
func (c *Cargo) AssignToRoute(itinerary Itinerary) {
	c.Itinerary = itinerary
	c.Delivery = c.Delivery.UpdateOnRouting(c.RouteSpecification, c.Itinerary)
}

// DeriveDeliveryProgress updates all aspects of the cargo aggregate status
// based on the current route specification, itinerary and handling of the cargo.
func (c *Cargo) DeriveDeliveryProgress(history HandlingHistory) {
	c.Delivery = DeriveDeliveryFrom(c.RouteSpecification, c.Itinerary, history)
}

func (c *Cargo) SameIdentity(e shared.Entity) bool {
	return c.TrackingId == e.(*Cargo).TrackingId
}

// NewCargo creates a new, unrouted cargo.
func NewCargo(trackingId TrackingId, routeSpecification RouteSpecification) *Cargo {
	emptyItinerary := Itinerary{}
	emptyHistory := HandlingHistory{make([]HandlingEvent, 0)}

	return &Cargo{
		TrackingId:         trackingId,
		Origin:             routeSpecification.Origin,
		RouteSpecification: routeSpecification,
		Delivery:           DeriveDeliveryFrom(routeSpecification, emptyItinerary, emptyHistory),
	}
}

// CargoRepository provides access a cargo store.
type CargoRepository interface {
	Store(cargo Cargo) error
	Find(trackingId TrackingId) (Cargo, error)
	FindAll() []Cargo
}

// ErrUnknownCargo is used when a cargo could not be found.
var ErrUnknownCargo = errors.New("unknown cargo")

// TODO: Move to infrastructure
func NextTrackingId() TrackingId {
	return TrackingId(strings.Split(strings.ToUpper(uuid.New()), "-")[0])
}

// RouteSpecification Contains information about a route: its origin,
// destination and arrival deadline.
type RouteSpecification struct {
	Origin          location.UNLocode
	Destination     location.UNLocode
	ArrivalDeadline time.Time
}

// IsSatisfiedBy checks whether provided itinerary satisfies this
// specification.
func (s RouteSpecification) IsSatisfiedBy(itinerary Itinerary) bool {
	return itinerary.Legs != nil &&
		s.Origin == itinerary.InitialDepartureLocation() &&
		s.Destination == itinerary.FinalArrivalLocation()
}

func (s RouteSpecification) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(s, v.(RouteSpecification))
}

// RoutingStatus describes status of cargo routing.
type RoutingStatus int

const (
	NotRouted RoutingStatus = iota
	Misrouted
	Routed
)

func (s RoutingStatus) String() string {
	switch s {
	case NotRouted:
		return "Not routed"
	case Misrouted:
		return "Misrouted"
	case Routed:
		return "Routed"
	}
	return ""
}

// TransportStatus describes status of cargo transportation.
type TransportStatus int

const (
	NotReceived TransportStatus = iota
	InPort
	OnboardCarrier
	Claimed
	Unknown
)

func (s TransportStatus) String() string {
	switch s {
	case NotReceived:
		return "Not received"
	case InPort:
		return "In port"
	case OnboardCarrier:
		return "Onboard carrier"
	case Claimed:
		return "Claimed"
	case Unknown:
		return "Unknown"
	}
	return ""
}
