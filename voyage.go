package shipping

import (
	"errors"
	"time"
)

// VoyageNumber uniquely identifies a particular Voyage.
type VoyageNumber string

// Voyage is a uniquely identifiable series of carrier movements.
type Voyage struct {
	VoyageNumber VoyageNumber
	Schedule     Schedule
}

// NewVoyage creates a voyage with a voyage number and a provided schedule.
func NewVoyage(n VoyageNumber, s Schedule) *Voyage {
	return &Voyage{VoyageNumber: n, Schedule: s}
}

// Schedule describes a voyage schedule.
type Schedule struct {
	CarrierMovements []CarrierMovement
}

// CarrierMovement is a vessel voyage from one location to another.
type CarrierMovement struct {
	DepartureLocation UNLocode
	ArrivalLocation   UNLocode
	DepartureTime     time.Time
	ArrivalTime       time.Time
}

// ErrUnknownVoyage is used when a voyage could not be found.
var ErrUnknownVoyage = errors.New("unknown voyage")

// VoyageRepository provides access a voyage store.
type VoyageRepository interface {
	Find(VoyageNumber) (*Voyage, error)
}
