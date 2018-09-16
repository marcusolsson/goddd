package shipping

import (
	"time"
)

// Leg describes the transportation between two locations on a voyage.
type Leg struct {
	VoyageNumber   VoyageNumber `json:"voyage_number"`
	LoadLocation   UNLocode     `json:"from"`
	UnloadLocation UNLocode     `json:"to"`
	LoadTime       time.Time    `json:"load_time"`
	UnloadTime     time.Time    `json:"unload_time"`
}

// MakeLeg creates a new itinerary leg.
func MakeLeg(voyageNumber VoyageNumber, loadLocation, unloadLocation UNLocode, loadTime, unloadTime time.Time) Leg {
	return Leg{
		VoyageNumber:   voyageNumber,
		LoadLocation:   loadLocation,
		UnloadLocation: unloadLocation,
		LoadTime:       loadTime,
		UnloadTime:     unloadTime,
	}
}

// Itinerary specifies steps required to transport a cargo from its origin to
// destination.
type Itinerary struct {
	Legs []Leg `json:"legs"`
}

// InitialDepartureLocation returns the start of the itinerary.
func (i Itinerary) InitialDepartureLocation() UNLocode {
	if i.IsEmpty() {
		return UNLocode("")
	}
	return i.Legs[0].LoadLocation
}

// FinalArrivalLocation returns the end of the itinerary.
func (i Itinerary) FinalArrivalLocation() UNLocode {
	if i.IsEmpty() {
		return UNLocode("")
	}
	return i.Legs[len(i.Legs)-1].UnloadLocation
}

// FinalArrivalTime returns the expected arrival time at final destination.
func (i Itinerary) FinalArrivalTime() time.Time {
	return i.Legs[len(i.Legs)-1].UnloadTime
}

// IsEmpty checks if the itinerary contains at least one leg.
func (i Itinerary) IsEmpty() bool {
	return len(i.Legs) == 0
}

// IsExpected checks if the given handling event is expected when executing
// this itinerary.
func (i Itinerary) IsExpected(event HandlingEvent) bool {
	if i.IsEmpty() {
		return true
	}

	switch event.Activity.Type {
	case Receive:
		return i.InitialDepartureLocation() == event.Activity.Location
	case Load:
		for _, l := range i.Legs {
			if l.LoadLocation == event.Activity.Location && l.VoyageNumber == event.Activity.VoyageNumber {
				return true
			}
		}
		return false
	case Unload:
		for _, l := range i.Legs {
			if l.UnloadLocation == event.Activity.Location && l.VoyageNumber == event.Activity.VoyageNumber {
				return true
			}
		}
		return false
	case Claim:
		return i.FinalArrivalLocation() == event.Activity.Location
	}

	return true
}
