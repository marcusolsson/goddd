package voyage

import (
	"errors"
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

// CarrierMovement is a vessel voyage from one location to another.
type CarrierMovement struct {
	DepartureLocation location.Location
	ArrivalLocation   location.Location
	DepartureTime     time.Time
	ArrivalTime       time.Time
}

// ErrUnknownVoyage is used when a voyage could not be found.
var ErrUnknownVoyage = errors.New("unknown voyage")

// Repository provides access a voyage store.
type Repository interface {
	Find(Number) (Voyage, error)
}
