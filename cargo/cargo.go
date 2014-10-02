package cargo

import (
	"container/list"
	"errors"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/marcusolsson/goddd/location"
)

// TrackingId uniquely identifies a particular cargo.
type TrackingId string

// Cargo is the central class in the domain model,
type Cargo struct {
	TrackingId         TrackingId
	Origin             location.Location
	RouteSpecification RouteSpecification
	Itinerary          Itinerary
	Delivery
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

// Updates all aspects of the cargo aggregate status based on the
// current route specification, itinerary and handling of the cargo.
func (c *Cargo) DeriveDeliveryProgress(history HandlingHistory) {
	c.Delivery = DeriveDeliveryFrom(c.RouteSpecification, c.Itinerary, history)
}

// Equaler is used to compare entities by identity.
type Equaler interface {
	Equal(Equaler) bool
}

func (c *Cargo) Equal(e Equaler) bool {
	return c.TrackingId == e.(*Cargo).TrackingId
}

// Assert that Cargo implements the Equaler interface.
var _ Equaler = &Cargo{}

// NewCargo creates a new cargo in a consistent state.
func NewCargo(trackingId TrackingId, routeSpecification RouteSpecification) *Cargo {
	emptyItinerary := Itinerary{}

	emptyHistory := HandlingHistory{}
	emptyHistory.HandlingEvents = list.New()

	// Cargo origin never changes, even if the route specification
	// changes. However, at creation, cargo orgin can be derived
	// from the initial route specification.
	return &Cargo{
		TrackingId:         trackingId,
		Origin:             routeSpecification.Origin,
		RouteSpecification: routeSpecification,
		Delivery:           DeriveDeliveryFrom(routeSpecification, emptyItinerary, emptyHistory),
	}
}

// RoutingStatus
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
	default:
		return ""
	}
}

// TransportStatus
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
	default:
		return ""
	}
}

// RouteSpecification describes where a cargo origin and destination
// is, and the arrival deadline.
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

type Leg struct {
	LoadLocation   location.Location
	UnloadLocation location.Location
}

// Itinerary
type Itinerary struct {
	Legs []Leg
}

func (i *Itinerary) InitialDepartureLocation() location.Location {
	if i.IsEmpty() {
		return location.UnknownLocation
	}
	return i.Legs[0].LoadLocation
}

func (i *Itinerary) FinalArrivalLocation() location.Location {
	if i.IsEmpty() {
		return location.UnknownLocation
	}
	return i.Legs[len(i.Legs)-1].UnloadLocation
}

func (i *Itinerary) IsEmpty() bool {
	return i.Legs == nil || len(i.Legs) == 0
}

// CargoRepository
type CargoRepository interface {
	Store(cargo Cargo) error
	// Finds a cargo using given id.
	Find(trackingId TrackingId) (Cargo, error)
	FindAll() []Cargo
}

var (
	ErrUnknownCargo = errors.New("Unknown cargo")
)

type cargoRepository struct {
	cargos map[TrackingId]Cargo
}

func (r *cargoRepository) Store(cargo Cargo) error {
	r.cargos[cargo.TrackingId] = cargo

	return nil
}

func (r *cargoRepository) Find(trackingId TrackingId) (Cargo, error) {

	if val, ok := r.cargos[trackingId]; ok {
		return val, nil
	}

	return Cargo{}, ErrUnknownCargo
}

func (r *cargoRepository) FindAll() []Cargo {
	c := make([]Cargo, 0, len(r.cargos))
	for _, val := range r.cargos {
		c = append(c, val)
	}
	return c
}

func NewCargoRepository() CargoRepository {
	return &cargoRepository{
		cargos: make(map[TrackingId]Cargo),
	}
}

func NextTrackingId() TrackingId {
	return TrackingId(strings.Split(strings.ToUpper(uuid.New()), "-")[0])
}
