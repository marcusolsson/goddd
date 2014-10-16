package voyage

import (
	"errors"
	"reflect"
	"time"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"
)

// VoyageNumber uniquely identifies a particular Voyage.
type VoyageNumber string

// Voyage
type Voyage struct {
	VoyageNumber VoyageNumber
	Schedule     Schedule
}

func (v *Voyage) SameIdentity(e shared.Entity) bool {
	return v.VoyageNumber == e.(*Voyage).VoyageNumber
}

// New creates a voyage with a voyage number and a provided schedule.
func New(n VoyageNumber, s Schedule) *Voyage {
	return &Voyage{VoyageNumber: n, Schedule: s}
}

// Schedule describes a voyage schedule.
type Schedule struct {
	CarrierMovements []CarrierMovement
}

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

func (m CarrierMovement) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(m, v.(CarrierMovement))
}

// ErrUnknownVoyage is used when a voyage could not be found.
var ErrUnknownVoyage = errors.New("unknown voyage")

// VoyageRepository provides access a voyage store.
type Repository interface {
	Find(VoyageNumber) (Voyage, error)
}
