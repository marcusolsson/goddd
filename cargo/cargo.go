package cargo

import (
	"errors"
	"os/exec"
	"time"

	"bitbucket.org/marcus_olsson/goddd/location"
)

func NewTrackingId() (TrackingId, error) {
	out, err := exec.Command("uuidgen").Output()
	return TrackingId(out), err
}

// RoutingStatus
type RoutingStatus int

const (
	NotRouted RoutingStatus = iota
	Misrouted
	Routed
)

// TransportStatus
type TransportStatus int

const (
	NotReceived TransportStatus = iota
	InPort
	OnboardCarrier
	Claimed
	Unknown
)

// RouteSpecification describes where a cargo origin and destination
// is, and the arrival deadline.
type RouteSpecification struct {
	Origin          location.Location
	Destination     location.Location
	ArrivalDeadline time.Time
}

func (s RouteSpecification) IsSatisfiedBy(itinerary Itinerary) bool {
	return true
}

type Leg struct {
	LoadLocation   location.Location
	UnloadLocation location.Location
}

// Itinerary
type Itinerary struct {
	Legs []Leg
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
	itinerary          Itinerary
}

// NewCargo creates a new cargo in a consistent state.
func NewCargo(trackingId TrackingId, routeSpecification RouteSpecification) *Cargo {
	return &Cargo{
		TrackingId:         trackingId,
		Origin:             routeSpecification.Origin,
		RouteSpecification: routeSpecification,
	}
}

// Assert that Cargo implements the Equaler interface.
var _ Equaler = &Cargo{}

func (c *Cargo) Equal(e Equaler) bool {
	return c.TrackingId == e.(*Cargo).TrackingId
}

func (c *Cargo) SpecifyNewRoute(routeSpecification RouteSpecification) {
	// TODO: Decide how to port the Delivery entity.
}

func (c *Cargo) AssignToRoute(itinerary Itinerary) {
	// TODO: Decide how to port the Delivery entity.
}

// CargoRepository
type CargoRepository interface {
	Store(cargo Cargo) error
	// Finds a cargo using given id.
	Find(trackingId TrackingId) (*Cargo, error)
	FindAll() ([]Cargo, error)
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

func (r *cargoRepository) Find(trackingId TrackingId) (*Cargo, error) {

	if val, ok := r.cargos[trackingId]; ok {
		return &val, nil
	}

	return nil, ErrUnknownCargo
}

func (r *cargoRepository) FindAll() ([]Cargo, error) {
	c := make([]Cargo, 0, len(r.cargos))
	for _, val := range r.cargos {
		c = append(c, val)
	}
	return c, nil
}

func NewCargoRepository() CargoRepository {
	return &cargoRepository{
		cargos: make(map[TrackingId]Cargo),
	}
}
