package cargo

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"
	"github.com/marcusolsson/goddd/domain/voyage"

	"code.google.com/p/go-uuid/uuid"
)

type TrackingId string

type Cargo struct {
	TrackingId         TrackingId
	Origin             location.UNLocode
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

	emptyHistory := HandlingHistory{make([]HandlingEvent, 0)}

	return &Cargo{
		TrackingId:         trackingId,
		Origin:             routeSpecification.Origin,
		RouteSpecification: routeSpecification,
		Delivery:           DeriveDeliveryFrom(routeSpecification, emptyItinerary, emptyHistory),
	}
}

type RouteSpecification struct {
	Origin          location.UNLocode
	Destination     location.UNLocode
	ArrivalDeadline time.Time
}

func (s RouteSpecification) IsSatisfiedBy(itinerary Itinerary) bool {
	return itinerary.Legs != nil &&
		s.Origin == itinerary.InitialDepartureLocation() &&
		s.Destination == itinerary.FinalArrivalLocation()
}

func (s RouteSpecification) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(s, v.(RouteSpecification))
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

type Leg struct {
	VoyageNumber   voyage.VoyageNumber
	LoadLocation   location.UNLocode
	UnloadLocation location.UNLocode
	LoadTime       time.Time
	UnloadTime     time.Time
}

func (l Leg) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(l, v.(Leg))
}

type Itinerary struct {
	Legs []Leg
}

func (i Itinerary) InitialDepartureLocation() location.UNLocode {
	if i.IsEmpty() {
		return location.UNLocode("")
	}
	return i.Legs[0].LoadLocation
}

func (i Itinerary) FinalArrivalLocation() location.UNLocode {
	if i.IsEmpty() {
		return location.UNLocode("")
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
		return i.InitialDepartureLocation() == event.Location
	case Load:
		for _, l := range i.Legs {
			if l.LoadLocation == event.Location && l.VoyageNumber == event.VoyageNumber {
				return true
			}
		}
		return false
	case Unload:
		for _, l := range i.Legs {
			if l.UnloadLocation == event.Location && l.VoyageNumber == event.VoyageNumber {
				return true
			}
		}
		return false
	case Claim:
		return i.FinalArrivalLocation() == event.Location
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

var ErrUnknownCargo = errors.New("unknown cargo")

func NextTrackingId() TrackingId {
	return TrackingId(strings.Split(strings.ToUpper(uuid.New()), "-")[0])
}
