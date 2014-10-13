package cargo

import (
	"reflect"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"
)

type Delivery struct {
	Itinerary
	RouteSpecification
	RoutingStatus
	TransportStatus
	IsMisdirected           bool
	IsUnloadedAtDestination bool
	NextExpectedActivity    HandlingActivity
	LastEvent               HandlingEvent
	LastKnownLocation       location.UNLocode
}

func (d Delivery) UpdateOnRouting(routeSpecification RouteSpecification, itinerary Itinerary) Delivery {
	return newDelivery(d.LastEvent, itinerary, routeSpecification)
}

func (d Delivery) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(d, v.(Delivery))
}

func (d Delivery) IsOnTrack() bool {
	return d.RoutingStatus == Routed && !d.IsMisdirected
}

func DeriveDeliveryFrom(routeSpecification RouteSpecification, itinerary Itinerary, history HandlingHistory) Delivery {
	lastEvent, _ := history.MostRecentlyCompletedEvent()
	return newDelivery(lastEvent, itinerary, routeSpecification)
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

func calculateMisdirectedStatus(event HandlingEvent, itinerary Itinerary) bool {
	if event.Activity.Type == NotHandled {
		return false
	}

	return !itinerary.IsExpected(event)
}

func calculateUnloadedAtDestination(event HandlingEvent, routeSpecification RouteSpecification) bool {
	if event.Activity.Type == NotHandled {
		return false
	}

	return event.Activity.Type == Unload && routeSpecification.Destination == event.Activity.Location
}

func calculateTransportStatus(event HandlingEvent) TransportStatus {
	switch event.Activity.Type {
	case NotHandled:
		return NotReceived
	case Load:
		return OnboardCarrier
	case Unload:
		return InPort
	case Receive:
		return InPort
	case Customs:
		return InPort
	case Claim:
		return Claimed
	}
	return Unknown
}

func calculateLastKnownLocation(event HandlingEvent) location.UNLocode {
	return event.Activity.Location
}

func calculateNextExpectedActivity(d Delivery) HandlingActivity {
	if !d.IsOnTrack() {
		return HandlingActivity{}
	}

	switch d.LastEvent.Activity.Type {
	case NotHandled:
		return HandlingActivity{Type: Receive, Location: d.RouteSpecification.Origin}
	case Load:
		for _, l := range d.Itinerary.Legs {
			if l.LoadLocation == d.LastEvent.Activity.Location {
				return HandlingActivity{Type: Unload, Location: l.UnloadLocation, VoyageNumber: l.VoyageNumber}
			}
		}
	case Unload:
		for i, l := range d.Itinerary.Legs {
			if l.UnloadLocation == d.LastEvent.Activity.Location {
				if i < len(d.Itinerary.Legs)-1 {
					return HandlingActivity{Type: Load, Location: d.Itinerary.Legs[i+1].LoadLocation, VoyageNumber: d.Itinerary.Legs[i+1].VoyageNumber}
				} else {
					return HandlingActivity{Type: Claim, Location: l.UnloadLocation}
				}
			}
		}
	}

	return HandlingActivity{}
}

func newDelivery(lastEvent HandlingEvent, itinerary Itinerary, routeSpecification RouteSpecification) Delivery {

	var (
		routingStatus           = calculateRoutingStatus(itinerary, routeSpecification)
		transportStatus         = calculateTransportStatus(lastEvent)
		lastKnownLocation       = calculateLastKnownLocation(lastEvent)
		isMisdirected           = calculateMisdirectedStatus(lastEvent, itinerary)
		isUnloadedAtDestination = calculateUnloadedAtDestination(lastEvent, routeSpecification)
	)

	d := Delivery{
		LastEvent:               lastEvent,
		Itinerary:               itinerary,
		RouteSpecification:      routeSpecification,
		RoutingStatus:           routingStatus,
		TransportStatus:         transportStatus,
		LastKnownLocation:       lastKnownLocation,
		IsMisdirected:           isMisdirected,
		IsUnloadedAtDestination: isUnloadedAtDestination,
	}

	d.NextExpectedActivity = calculateNextExpectedActivity(d)

	return d
}
