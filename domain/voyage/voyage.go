package voyage

import (
	"errors"
	"reflect"
	"time"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"
)

// Number uniquely identifies a particular Voyage.
type Number string

// Voyage is a uniquely identifiable series of carrier movements.
type Voyage struct {
	Number   Number
	Schedule Schedule
}

// SameIdentity returns whether two voyages have the same voyage number.
func (v *Voyage) SameIdentity(e shared.Entity) bool {
	return v.Number == e.(*Voyage).Number
}

// New creates a voyage with a voyage number and a provided schedule.
func New(n Number, s Schedule) *Voyage {
	return &Voyage{Number: n, Schedule: s}
}

// Schedule describes a voyage schedule.
type Schedule struct {
	CarrierMovements []CarrierMovement
}

// SameValue returns whether two schedules have the same value.
func (s Schedule) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(s, v.(Schedule))
}

// CarrierMovement is a vessel voyage from one location to another.
type CarrierMovement struct {
	DepartureLocation location.Location
	ArrivalLocation   location.Location
	DepartureTime     time.Time
	ArrivalTime       time.Time
}

// SameValue returns whether two carrier movements have the same value.
func (m CarrierMovement) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(m, v.(CarrierMovement))
}

// ErrUnknownVoyage is used when a voyage could not be found.
var ErrUnknownVoyage = errors.New("unknown voyage")

// Repository provides access a voyage store.
type Repository interface {
	Find(Number) (Voyage, error)
}
