package cargo

import (
	"container/list"
	"testing"
	"time"

	"github.com/marcusolsson/goddd/domain/location"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestConstruction(c *C) {
	trackingId := TrackingId("XYZ")
	specification := RouteSpecification{
		Origin:          location.Stockholm,
		Destination:     location.Melbourne,
		ArrivalDeadline: time.Date(2009, time.March, 13, 0, 0, 0, 0, time.UTC),
	}

	cargo := NewCargo(trackingId, specification)

	c.Check(cargo.Delivery.RoutingStatus, Equals, NotRouted)
	c.Check(cargo.Delivery.TransportStatus, Equals, NotReceived)
	c.Check(cargo.Delivery.LastKnownLocation, Equals, location.UnknownLocation)
}

func (s *S) TestEquality(c *C) {
	spec1 := RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	}
	spec2 := RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Melbourne,
	}

	c.Check(spec1.SameValue(spec1), Equals, true)
	c.Check(spec1.SameValue(spec2), Equals, false)

	c1 := NewCargo("ABC", spec1)
	c2 := NewCargo("CBA", spec1)
	c3 := NewCargo("ABC", spec2)
	c4 := NewCargo("ABC", spec1)

	c.Check(c1.SameIdentity(c4), Equals, true)
	c.Check(c1.SameIdentity(c3), Equals, true)
	c.Check(c3.SameIdentity(c4), Equals, true)
	c.Check(c1.SameIdentity(c2), Equals, false)
}

func (s *S) TestItineraryEquality(c *C) {

	i1 := Itinerary{Legs: []Leg{
		Leg{LoadLocation: location.Stockholm, UnloadLocation: location.Melbourne},
		Leg{LoadLocation: location.Melbourne, UnloadLocation: location.Hongkong},
	}}

	i2 := Itinerary{Legs: []Leg{
		Leg{LoadLocation: location.Stockholm, UnloadLocation: location.Melbourne},
		Leg{LoadLocation: location.Melbourne, UnloadLocation: location.Hongkong},
	}}

	i3 := Itinerary{Legs: []Leg{
		Leg{LoadLocation: location.Hongkong, UnloadLocation: location.Melbourne},
		Leg{LoadLocation: location.Melbourne, UnloadLocation: location.Stockholm},
	}}

	c.Check(i1.SameValue(i1), Equals, true)
	c.Check(i1.SameValue(i2), Equals, true)
	c.Check(i1.SameValue(i3), Equals, false)
	c.Check(i2.SameValue(i3), Equals, false)
}

func (s *S) TestRoutingStatus(c *C) {
	cargo := NewCargo("ABC", RouteSpecification{})

	good := Itinerary{Legs: make([]Leg, 1)}
	good.Legs[0] = Leg{
		LoadLocation:   location.Stockholm,
		UnloadLocation: location.Melbourne,
	}

	bad := Itinerary{Legs: make([]Leg, 1)}
	bad.Legs[0] = Leg{
		LoadLocation:   location.Stockholm,
		UnloadLocation: location.Hongkong,
	}

	acceptOnlyGood := RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Melbourne,
	}

	cargo.SpecifyNewRoute(acceptOnlyGood)
	c.Check(cargo.Delivery.RoutingStatus, Equals, NotRouted)

	cargo.AssignToRoute(bad)
	c.Check(cargo.Delivery.RoutingStatus, Equals, Misrouted)

	cargo.AssignToRoute(good)
	c.Check(cargo.Delivery.RoutingStatus, Equals, Routed)
}

func (s *S) TestLastKnownLocationUnknownWhenNoEvents(c *C) {
	cargo := NewCargo("ABC", RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	})

	c.Check(location.UnknownLocation, Equals, cargo.Delivery.LastKnownLocation)
}

func (s *S) TestLastKnownLocationReceived(c *C) {
	cargo := populateCargoReceivedInStockholm()
	c.Check(location.Stockholm, Equals, cargo.Delivery.LastKnownLocation)
}

func (s *S) TestLastKnownLocationClaimed(c *C) {
	cargo := populateCargoReceivedInStockholm()
	c.Check(location.Stockholm, Equals, cargo.Delivery.LastKnownLocation)
}

func (s *S) TestItineraryIsExpected(c *C) {

	emptyItinerary := Itinerary{}
	emptyEvent := HandlingEvent{}
	c.Check(emptyItinerary.IsExpected(emptyEvent), Equals, true)

	i := Itinerary{[]Leg{
		Leg{LoadLocation: location.Stockholm, UnloadLocation: location.Melbourne},
		Leg{LoadLocation: location.Melbourne, UnloadLocation: location.Hongkong},
	}}
	c.Check(i.IsExpected(emptyEvent), Equals, true)

	var (
		receiveEvent              = HandlingEvent{Type: Receive, Location: location.Stockholm}
		receiveEventWrongLocation = HandlingEvent{Type: Receive, Location: location.Melbourne}
	)
	c.Check(i.IsExpected(receiveEvent), Equals, true)
	c.Check(i.IsExpected(receiveEventWrongLocation), Equals, false)

	var (
		claimEvent              = HandlingEvent{Type: Claim, Location: location.Hongkong}
		claimEventWrongLocation = HandlingEvent{Type: Claim, Location: location.Stockholm}
	)
	c.Check(i.IsExpected(claimEvent), Equals, true)
	c.Check(i.IsExpected(claimEventWrongLocation), Equals, false)
}

func populateCargoReceivedInStockholm() *Cargo {
	cargo := NewCargo("XYZ", RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Melbourne,
	})

	e := HandlingEvent{
		TrackingId: cargo.TrackingId,
		Type:       Receive,
		Location:   location.Stockholm,
	}

	history := HandlingHistory{HandlingEvents: list.New()}
	history.HandlingEvents.PushBack(e)

	cargo.DeriveDeliveryProgress(history)

	return cargo
}

func populateCargoClaimedInMelbourne() *Cargo {
	cargo := NewCargo("XYZ", RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Melbourne,
	})

	e := HandlingEvent{
		TrackingId: cargo.TrackingId,
		Type:       Claim,
		Location:   location.Melbourne,
	}

	history := HandlingHistory{HandlingEvents: list.New()}
	history.HandlingEvents.PushBack(e)

	cargo.DeriveDeliveryProgress(history)

	return cargo
}
