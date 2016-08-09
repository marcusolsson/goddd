package cargo

import (
	"testing"
	"time"

	"github.com/marcusolsson/goddd/location"
)

func TestConstruction(t *testing.T) {
	id := NextTrackingID()
	spec := RouteSpecification{
		Origin:          location.SESTO,
		Destination:     location.AUMEL,
		ArrivalDeadline: time.Date(2009, time.March, 13, 0, 0, 0, 0, time.UTC),
	}

	c := New(id, spec)

	if c.Delivery.RoutingStatus != NotRouted {
		t.Errorf("RoutingStatus = %v; want = %v",
			c.Delivery.RoutingStatus, NotRouted)
	}
	if c.Delivery.TransportStatus != NotReceived {
		t.Errorf("TransportStatus = %v; want = %v",
			c.Delivery.TransportStatus, NotReceived)
	}
	if c.Delivery.LastKnownLocation != "" {
		t.Errorf("LastKnownLocation = %s; want = %s",
			c.Delivery.LastKnownLocation, "")
	}
}

func TestRoutingStatus(t *testing.T) {
	good := Itinerary{
		Legs: []Leg{
			{LoadLocation: location.SESTO, UnloadLocation: location.AUMEL},
		},
	}

	bad := Itinerary{
		Legs: []Leg{
			{LoadLocation: location.SESTO, UnloadLocation: location.CNHKG},
		},
	}

	acceptOnlyGood := RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.AUMEL,
	}

	c := New("ABC", RouteSpecification{})

	c.SpecifyNewRoute(acceptOnlyGood)
	if c.Delivery.RoutingStatus != NotRouted {
		t.Errorf("RoutingStatus = %v; want = %v",
			c.Delivery.RoutingStatus, NotRouted)
	}

	c.AssignToRoute(bad)
	if c.Delivery.RoutingStatus != Misrouted {
		t.Errorf("RoutingStatus = %v; want = %v",
			c.Delivery.RoutingStatus, Misrouted)
	}

	c.AssignToRoute(good)
	if c.Delivery.RoutingStatus != Routed {
		t.Errorf("RoutingStatus = %v; want = %v",
			c.Delivery.RoutingStatus, Routed)
	}
}

func TestLastKnownLocation_WhenNoEvents(t *testing.T) {
	c := New("ABC", RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	})

	if c.Delivery.LastKnownLocation != "" {
		t.Errorf("should be equal")
	}
}

func TestLastKnownLocation_WhenReceived(t *testing.T) {
	c := populateCargoReceivedInStockholm()

	if c.Delivery.LastKnownLocation != location.SESTO {
		t.Errorf("LastKnownLocation = %s; want = %s",
			c.Delivery.LastKnownLocation, location.SESTO)
	}
}

func TestLastKnownLocation_WhenClaimed(t *testing.T) {
	c := populateCargoClaimedInMelbourne()

	if c.Delivery.LastKnownLocation != location.AUMEL {
		t.Errorf("LastKnownLocation = %s; want = %s",
			c.Delivery.LastKnownLocation, location.AUMEL)
	}
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

func TestRoutingStatus_Stringer(t *testing.T) {
	for _, tt := range routingStatusTests {
		if tt.routingStatus.String() != tt.expected {
			t.Errorf("routingStatus.String() = %s; want = %s",
				tt.routingStatus.String(), tt.expected)
		}
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

func TestTransportStatus_Stringer(t *testing.T) {
	for _, tt := range transportStatusTests {
		if tt.transportStatus.String() != tt.expected {
			t.Errorf("transportStatus.String() = %s; want = %s",
				tt.transportStatus.String(), tt.expected)
		}
	}
}

func populateCargoReceivedInStockholm() *Cargo {
	c := New("XYZ", RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.AUMEL,
	})

	e := HandlingEvent{
		TrackingID: c.TrackingID,
		Activity: HandlingActivity{
			Type:     Receive,
			Location: location.SESTO,
		},
	}

	hh := HandlingHistory{
		HandlingEvents: []HandlingEvent{e},
	}

	c.DeriveDeliveryProgress(hh)

	return c
}

func populateCargoClaimedInMelbourne() *Cargo {
	c := New("XYZ", RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.AUMEL,
	})

	e := HandlingEvent{
		TrackingID: c.TrackingID,
		Activity: HandlingActivity{
			Type:     Claim,
			Location: location.AUMEL,
		},
	}

	hh := HandlingHistory{
		HandlingEvents: []HandlingEvent{e},
	}

	c.DeriveDeliveryProgress(hh)

	return c
}
