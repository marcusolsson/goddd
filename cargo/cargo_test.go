package cargo

import (
	"container/list"
	"testing"
	"time"

	"bitbucket.org/marcus_olsson/goddd/location"

	. "gopkg.in/check.v1"
)

// Hook gocheck up to the "go test" runner
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

	NewCargo(trackingId, specification)
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

	c1 := NewCargo("ABC", spec1)
	c2 := NewCargo("CBA", spec1)
	c3 := NewCargo("ABC", spec2)
	c4 := NewCargo("ABC", spec1)

	c.Check(true, Equals, c1.Equal(c4))
	c.Check(true, Equals, c1.Equal(c3))
	c.Check(true, Equals, c3.Equal(c4))
	c.Check(false, Equals, c1.Equal(c2))
}

func (s *S) TestRepositoryFind(c *C) {
	c1 := NewCargo("ABC", RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	})

	r := NewCargoRepository()
	r.Store(*c1)

	c2, err := r.Find(TrackingId("ABC"))

	c.Assert(err, IsNil)
	c.Check(true, Equals, c1.Equal(&c2))
}

func (s *S) TestRepositoryFindAll(c *C) {
	c1 := NewCargo("ABC", RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	})

	c2 := NewCargo("CBA", RouteSpecification{
		Origin:      location.Hongkong,
		Destination: location.Stockholm,
	})

	r := NewCargoRepository()
	r.Store(*c1)
	r.Store(*c2)

	cargos := r.FindAll()

	c.Check(cargos, HasLen, 2)
}

func (s *S) TestRoutingStatus(c *C) {
	cargo := NewCargo("ABC", RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	})

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
	c.Check(NotRouted, Equals, cargo.Delivery.RoutingStatus)

	cargo.AssignToRoute(bad)
	c.Check(Misrouted, Equals, cargo.Delivery.RoutingStatus)

	cargo.AssignToRoute(good)
	c.Check(Routed, Equals, cargo.Delivery.RoutingStatus)
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
