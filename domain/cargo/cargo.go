package cargo

import (
	"container/list"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"

	"code.google.com/p/go-uuid/uuid"
)

type TrackingId string

type Cargo struct {
	TrackingId         TrackingId
	Origin             location.Location
	RouteSpecification RouteSpecification
	Itinerary          Itinerary
	Delivery
}

func (c *Cargo) SpecifyNewRoute(routeSpecification RouteSpecification) {
	c.RouteSpecification = routeSpecification
	c.Delivery = c.Delivery.UpdateOnRouting(c.RouteSpecification, c.Itinerary)
}

func (c *Cargo) AssignToRoute(itinerary Itinerary) {
	c.Itinerary = itinerary
	c.Delivery = c.Delivery.UpdateOnRouting(c.RouteSpecification, c.Itinerary)
}

func (c *Cargo) DeriveDeliveryProgress(history HandlingHistory) {
	c.Delivery = DeriveDeliveryFrom(c.RouteSpecification, c.Itinerary, history)
}

func (c *Cargo) SameIdentity(e shared.Entity) bool {
	return c.TrackingId == e.(*Cargo).TrackingId
}

func NewCargo(trackingId TrackingId, routeSpecification RouteSpecification) *Cargo {
	emptyItinerary := Itinerary{}

	emptyHistory := HandlingHistory{}
	emptyHistory.HandlingEvents = list.New()

	return &Cargo{
		TrackingId:         trackingId,
		Origin:             routeSpecification.Origin,
		RouteSpecification: routeSpecification,
		Delivery:           DeriveDeliveryFrom(routeSpecification, emptyItinerary, emptyHistory),
	}
}

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
	case InPort:
		return "In port"
	}
	return ""
}

type RouteSpecification struct {
	Origin          location.Location
	Destination     location.Location
	ArrivalDeadline time.Time
}

func (s RouteSpecification) IsSatisfiedBy(itinerary Itinerary) bool {
	return itinerary.Legs != nil &&
		s.Origin.UNLocode == itinerary.InitialDepartureLocation().UNLocode &&
		s.Destination.UNLocode == itinerary.FinalArrivalLocation().UNLocode
}

func (s RouteSpecification) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(s, v.(RouteSpecification))
}

type Leg struct {
	Voyage         string
	LoadLocation   location.Location
	UnloadLocation location.Location
}

func (l Leg) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(l, v.(Leg))
}

type Itinerary struct {
	Legs []Leg
}

func (i Itinerary) InitialDepartureLocation() location.Location {
	if i.IsEmpty() {
		return location.UnknownLocation
	}
	return i.Legs[0].LoadLocation
}

func (i Itinerary) FinalArrivalLocation() location.Location {
	if i.IsEmpty() {
		return location.UnknownLocation
	}
	return i.Legs[len(i.Legs)-1].UnloadLocation
}

func (i Itinerary) IsEmpty() bool {
	return i.Legs == nil || len(i.Legs) == 0
}

func (i Itinerary) IsExpected(event HandlingEvent) bool {
	if i.IsEmpty() {
		return true
	}

	switch event.Type {
	case Receive:
		return i.InitialDepartureLocation().SameIdentity(event.Location)
	case Claim:
		return i.FinalArrivalLocation().SameIdentity(event.Location)
	}

	return true
}

func (i Itinerary) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(i, v.(Itinerary))
}

type CargoRepository interface {
	Store(cargo Cargo) error
	Find(trackingId TrackingId) (Cargo, error)
	FindAll() []Cargo
}

var ErrUnknownCargo = errors.New("Unknown cargo")

func NextTrackingId() TrackingId {
	return TrackingId(strings.Split(strings.ToUpper(uuid.New()), "-")[0])
}
