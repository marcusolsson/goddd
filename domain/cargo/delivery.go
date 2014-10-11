package cargo

import (
	"reflect"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"
)

type Delivery struct {
	LastEvent         HandlingEvent
	LastKnownLocation location.UNLocode
	Itinerary
	RouteSpecification
	RoutingStatus
	TransportStatus
	IsMisdirected           bool
	IsUnloadedAtDestination bool
}

func (d Delivery) UpdateOnRouting(routeSpecification RouteSpecification, itinerary Itinerary) Delivery {
	return newDelivery(d.LastEvent, itinerary, routeSpecification)
}

func (d Delivery) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(d, v.(Delivery))
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
	if event.Type == NotHandled {
		return false
	}

	return !itinerary.IsExpected(event)
}

func calculateUnloadedAtDestination(event HandlingEvent, routeSpecification RouteSpecification) bool {
	if event.Type == NotHandled {
		return false
	}

	return event.Type == Unload && routeSpecification.Destination == event.Location
}

func calculateTransportStatus(event HandlingEvent) TransportStatus {
	switch event.Type {
	case NotHandled:
		return NotReceived
	case Load:
		return OnboardCarrier
	case Unload:
	case Receive:
	case Customs:
		return InPort
	case Claim:
		return Claimed
	}
	return Unknown
}

func calculateLastKnownLocation(event HandlingEvent) location.UNLocode {
	return event.Location
}

func newDelivery(lastEvent HandlingEvent, itinerary Itinerary, routeSpecification RouteSpecification) Delivery {

	var (
		routingStatus           = calculateRoutingStatus(itinerary, routeSpecification)
		TransportStatus         = calculateTransportStatus(lastEvent)
		lastKnownLocation       = calculateLastKnownLocation(lastEvent)
		isMisdirected           = calculateMisdirectedStatus(lastEvent, itinerary)
		isUnloadedAtDestination = calculateUnloadedAtDestination(lastEvent, routeSpecification)
	)

	return Delivery{
		LastEvent:               lastEvent,
		Itinerary:               itinerary,
		RouteSpecification:      routeSpecification,
		RoutingStatus:           routingStatus,
		TransportStatus:         TransportStatus,
		LastKnownLocation:       lastKnownLocation,
		IsMisdirected:           isMisdirected,
		IsUnloadedAtDestination: isUnloadedAtDestination,
	}
}
