package cargo

import "github.com/marcusolsson/goddd/location"

// Delivery is the actual transportation of the cargo, as opposed to
// the customer requirement (RouteSpecification) and the plan
// (Itinerary).
type Delivery struct {
	LastEvent         HandlingEvent
	LastKnownLocation location.Location
	Itinerary
	RouteSpecification
	RoutingStatus
	TransportStatus
}

// UpdateOnRouting creates a new delivery snapshot to reflect changes
// in routing, i.e.  when the route specification or the itinerary has
// changed but no additional handling of the cargo has been performed.
func (d *Delivery) UpdateOnRouting(routeSpecification RouteSpecification, itinerary Itinerary) Delivery {
	return newDelivery(d.LastEvent, itinerary, routeSpecification)
}

// DerivedFrom creates a new delivery snapshot based on the complete
// handling history of a cargo, as well as its route specification and
// itinerary.
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

func calculateLastKnownLocation(event HandlingEvent) location.Location {
	return event.Location
}

func newDelivery(lastEvent HandlingEvent, itinerary Itinerary, routeSpecification RouteSpecification) Delivery {

	var (
		routingStatus     = calculateRoutingStatus(itinerary, routeSpecification)
		TransportStatus   = calculateTransportStatus(lastEvent)
		lastKnownLocation = calculateLastKnownLocation(lastEvent)
	)

	return Delivery{
		LastEvent:          lastEvent,
		Itinerary:          itinerary,
		RouteSpecification: routeSpecification,
		RoutingStatus:      routingStatus,
		TransportStatus:    TransportStatus,
		LastKnownLocation:  lastKnownLocation,
	}
}
