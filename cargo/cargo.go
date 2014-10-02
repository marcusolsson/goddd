package cargo

import (
	"container/list"
	"errors"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/marcusolsson/goddd/location"
)

func NextTrackingId() TrackingId {
	return TrackingId(strings.Split(strings.ToUpper(uuid.New()), "-")[0])
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
	if i.Legs == nil || len(i.Legs) == 0 {
		return location.UnknownLocation
	}
	return i.Legs[0].LoadLocation
}

func (i *Itinerary) FinalArrivalLocation() location.Location {
	if i.Legs == nil || len(i.Legs) == 0 {
		return location.UnknownLocation
	}
	return i.Legs[len(i.Legs)-1].UnloadLocation
}

// Delivery is the actual transportation of the cargo, as opposed to
// the customer requirement (RouteSpecification) and the plan
// (Itinerary).
type Delivery struct {
	LastEvent         HandlingEvent
	LastKnownLocation location.Location
	Itinerary
	RouteSpecification
	RoutingStatus
	TransportStatus
}

// UpdateOnRouting creates a new delivery snapshot to reflect changes
// in routing, i.e.  when the route specification or the itinerary has
// changed but no additional handling of the cargo has been performed.
func (d *Delivery) UpdateOnRouting(routeSpecification RouteSpecification, itinerary Itinerary) Delivery {
	return newDelivery(d.LastEvent, itinerary, routeSpecification)
}

// DerivedFrom creates a new delivery snapshot based on the complete
// handling history of a cargo, as well as its route specification and
// itinerary.
func DeriveDeliveryFrom(routeSpecification RouteSpecification, itinerary Itinerary, history HandlingHistory) Delivery {
	lastEvent, _ := history.MostRecentlyCompletedEvent()
	return newDelivery(lastEvent, itinerary, routeSpecification)
}

func newDelivery(lastEvent HandlingEvent, itinerary Itinerary, routeSpecification RouteSpecification) Delivery {

	var (
		routingStatus     = calculateRoutingStatus(itinerary, routeSpecification)
		TransportStatus   = calculateTransportStatus(lastEvent)
		lastKnownLocation = calculateLastKnownLocation(lastEvent)
	)

	return Delivery{
		LastEvent:          lastEvent,
		Itinerary:          itinerary,
		RouteSpecification: routeSpecification,
		RoutingStatus:      routingStatus,
		TransportStatus:    TransportStatus,
		LastKnownLocation:  lastKnownLocation,
	}
}

func calculateRoutingStatus(itinerary Itinerary, routeSpecification RouteSpecification) RoutingStatus {
	if itinerary.Legs == nil {
		return NotRouted
	} else {
		if routeSpecification.IsSatisfiedBy(itinerary) {
			return Routed
		} else {
			return Misrouted
		}
	}

}

func calculateTransportStatus(event HandlingEvent) TransportStatus {
	zero := HandlingEvent{}
	if event == zero {
		return NotReceived
	}

	switch event.Type {
	case Load:
		return OnboardCarrier
	case Unload:
	case Receive:
	case Customs:
		return InPort
	case Claim:
		return Claimed
	default:
		return Unknown
	}

	return Unknown
}

func calculateLastKnownLocation(event HandlingEvent) location.Location {
	zero := HandlingEvent{}
	if event != zero {
		return event.Location
	} else {
		return location.UnknownLocation
	}
}

// TrackingId uniquely identifies a particular cargo.
type TrackingId string

type Equaler interface {
	Equal(Equaler) bool
}

// Cargo is the central class in the domain model,
type Cargo struct {
	TrackingId         TrackingId
	Origin             location.Location
	RouteSpecification RouteSpecification
	Itinerary          Itinerary
	Delivery
}

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

// Assert that Cargo implements the Equaler interface.
var _ Equaler = &Cargo{}

func (c *Cargo) Equal(e Equaler) bool {
	return c.TrackingId == e.(*Cargo).TrackingId
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
// When either of those three changes, i.e. when a new route is
// specified for the cargo, the cargo is assigned to a route or when
// the cargo is handled, the status must be re-calculated.
//
// RouteSpecification and Itinerary are both inside the Cargo
// aggregate, so changes to them cause the status to be updated
// synchronously, but changes to the delivery history (when a cargo is
// handled) cause the status update to happen asynchronously since
// HandlingEvent is in a different aggregate.
func (c *Cargo) DeriveDeliveryProgress(history HandlingHistory) {
	c.Delivery = DeriveDeliveryFrom(c.RouteSpecification, c.Itinerary, history)
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
