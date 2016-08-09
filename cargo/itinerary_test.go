package cargo

import (
	"reflect"
	"testing"

	"github.com/marcusolsson/goddd/location"
)

func TestItinerary_CreateEmpty(t *testing.T) {
	i := Itinerary{}

	var legs []Leg

	if !reflect.DeepEqual(i.Legs, legs) {
		t.Errorf("should be equal")
	}
	if i.InitialDepartureLocation() != "" {
		t.Errorf("InitialDepartureLocation() = %s; want = %s",
			i.InitialDepartureLocation(), "")
	}
	if i.FinalArrivalLocation() != "" {
		t.Errorf("FinalArrivalLocation() = %s; want = %s",
			i.FinalArrivalLocation(), "")
	}
}

func TestItinerary_IsExpected_EmptyItinerary(t *testing.T) {
	i := Itinerary{}
	e := HandlingEvent{}

	if got, want := i.IsExpected(e), true; got != want {
		t.Errorf("IsExpected() = %v; want = %v", got, want)
	}
}

type eventExpectedTest struct {
	act HandlingActivity
	exp bool
}

var eventExpectedTests = []eventExpectedTest{
	{HandlingActivity{}, true},
	{HandlingActivity{Type: Receive, Location: location.SESTO}, true},
	{HandlingActivity{Type: Receive, Location: location.AUMEL}, false},
	{HandlingActivity{Type: Load, Location: location.AUMEL, VoyageNumber: "001A"}, true},
	{HandlingActivity{Type: Load, Location: location.CNHKG, VoyageNumber: "001A"}, false},
	{HandlingActivity{Type: Unload, Location: location.CNHKG, VoyageNumber: "001A"}, true},
	{HandlingActivity{Type: Unload, Location: location.SESTO, VoyageNumber: "001A"}, false},
	{HandlingActivity{Type: Claim, Location: location.CNHKG}, true},
	{HandlingActivity{Type: Claim, Location: location.SESTO}, false},
}

func TestItinerary_IsExpected(t *testing.T) {
	i := Itinerary{Legs: []Leg{
		{
			VoyageNumber:   "001A",
			LoadLocation:   location.SESTO,
			UnloadLocation: location.AUMEL,
		},
		{
			VoyageNumber:   "001A",
			LoadLocation:   location.AUMEL,
			UnloadLocation: location.CNHKG,
		},
	}}

	for _, tt := range eventExpectedTests {
		e := HandlingEvent{
			Activity: tt.act,
		}

		if got := i.IsExpected(e); got != tt.exp {
			t.Errorf("IsExpected() = %v; want = %v", got, tt.exp)
		}
	}
}
