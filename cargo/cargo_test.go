package cargo

import (
	"testing"
	"time"

	"github.com/marcusolsson/goddd/location"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestConstruction(c *C) {
	trackingID := TrackingID("XYZ")
	specification := RouteSpecification{
		Origin:          location.SESTO,
		Destination:     location.AUMEL,
		ArrivalDeadline: time.Date(2009, time.March, 13, 0, 0, 0, 0, time.UTC),
	}

	cargo := NewCargo(trackingID, specification)

	c.Check(cargo.Delivery.RoutingStatus, Equals, NotRouted)
	c.Check(cargo.Delivery.TransportStatus, Equals, NotReceived)
	c.Check(cargo.Delivery.LastKnownLocation, Equals, location.UNLocode(""))
}

func (s *S) TestEquality(c *C) {
	spec1 := RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	}
	spec2 := RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.AUMEL,
	}

	c.Check(spec1, DeepEquals, spec1)
	c.Check(spec1, Not(DeepEquals), spec2)

	c1 := NewCargo("ABC", spec1)
	c2 := NewCargo("CBA", spec1)
	c3 := NewCargo("ABC", spec2)
	c4 := NewCargo("ABC", spec1)

	c.Check(c1.TrackingID, Equals, c4.TrackingID)
	c.Check(c1.TrackingID, Equals, c3.TrackingID)
	c.Check(c3.TrackingID, Equals, c4.TrackingID)
	c.Check(c1.TrackingID, Not(Equals), c2.TrackingID)
}

func (s *S) TestRoutingStatus(c *C) {
	cargo := NewCargo("ABC", RouteSpecification{})

	good := Itinerary{Legs: make([]Leg, 1)}
	good.Legs[0] = Leg{
		LoadLocation:   location.SESTO,
		UnloadLocation: location.AUMEL,
	}

	bad := Itinerary{Legs: make([]Leg, 1)}
	bad.Legs[0] = Leg{
		LoadLocation:   location.SESTO,
		UnloadLocation: location.CNHKG,
	}

	acceptOnlyGood := RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.AUMEL,
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
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	})

	c.Check(location.UNLocode(""), Equals, cargo.Delivery.LastKnownLocation)
}

func (s *S) TestLastKnownLocationReceived(c *C) {
	cargo := populateCargoReceivedInStockholm()
	c.Check(location.SESTO, Equals, cargo.Delivery.LastKnownLocation)
}

func (s *S) TestLastKnownLocationClaimed(c *C) {
	cargo := populateCargoReceivedInStockholm()
	c.Check(location.SESTO, Equals, cargo.Delivery.LastKnownLocation)
}

var routingStatusTests = []struct {
	routingStatus RoutingStatus
	expected      string
}{
	{NotRouted, "Not routed"},
	{Misrouted, "Misrouted"},
	{Routed, "Routed"},
	{1000, ""},
}

func (s *S) TestRoutingStatusStringer(c *C) {
	for _, tt := range routingStatusTests {
		c.Check(tt.routingStatus.String(), Equals, tt.expected)
	}
}

var transportStatusTests = []struct {
	transportStatus TransportStatus
	expected        string
}{
	{NotReceived, "Not received"},
	{InPort, "In port"},
	{OnboardCarrier, "Onboard carrier"},
	{Claimed, "Claimed"},
	{Unknown, "Unknown"},
	{1000, ""},
}

func (s *S) TestTransportStatusStringer(c *C) {
	for _, tt := range transportStatusTests {
		c.Check(tt.transportStatus.String(), Equals, tt.expected)
	}
}

func populateCargoReceivedInStockholm() *Cargo {
	cargo := NewCargo("XYZ", RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.AUMEL,
	})

	e := HandlingEvent{
		TrackingID: cargo.TrackingID,
		Activity: HandlingActivity{
			Type:     Receive,
			Location: location.SESTO,
		},
	}

	history := HandlingHistory{HandlingEvents: make([]HandlingEvent, 0)}
	history.HandlingEvents = append(history.HandlingEvents, e)

	cargo.DeriveDeliveryProgress(history)

	return cargo
}

func populateCargoClaimedInMelbourne() *Cargo {
	cargo := NewCargo("XYZ", RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.AUMEL,
	})

	e := HandlingEvent{
		TrackingID: cargo.TrackingID,
		Activity: HandlingActivity{
			Type:     Claim,
			Location: location.AUMEL,
		},
	}

	history := HandlingHistory{HandlingEvents: make([]HandlingEvent, 0)}
	history.HandlingEvents = append(history.HandlingEvents, e)

	cargo.DeriveDeliveryProgress(history)

	return cargo
}
