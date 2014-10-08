package voyage

import (
	"reflect"
	"time"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"
)

type VoyageNumber string

type Voyage struct {
	VoyageNumber
	Schedule
}

type Schedule struct {
	CarrierMovements []CarrierMovement
}

type CarrierMovement struct {
	DepartureLocation location.Location
	ArrivalLocation   location.Location
	DepartureTime     time.Time
	ArrivalTime       time.Time
}

type VoyageRepository interface {
	Find(VoyageNumber) Voyage
}

func New(n VoyageNumber, s Schedule) *Voyage {
	return &Voyage{VoyageNumber: n, Schedule: s}
}

func (v *Voyage) SameIdentity(vo shared.ValueObject) bool {
	return v.VoyageNumber == vo.(*Voyage).VoyageNumber
}

func (s Schedule) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(s, v.(Schedule))
}

func (m CarrierMovement) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(m, v.(CarrierMovement))
}

// Sample voyages for test purposes.
var (
	// Voyage number 0100S (by ship)
	MelbourneToStockholm = New("0100S", Schedule{
		[]CarrierMovement{
			CarrierMovement{DepartureLocation: location.Melbourne, ArrivalLocation: location.Hongkong},
			CarrierMovement{DepartureLocation: location.Hongkong, ArrivalLocation: location.Stockholm},
		},
	})
)
