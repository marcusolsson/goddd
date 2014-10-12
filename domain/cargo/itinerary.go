package cargo

import (
	"reflect"
	"time"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"
	"github.com/marcusolsson/goddd/domain/voyage"
)

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
