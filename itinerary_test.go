package shipping

import (
	"reflect"
	"testing"
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
	{HandlingActivity{Type: Receive, Location: SESTO}, true},
	{HandlingActivity{Type: Receive, Location: AUMEL}, false},
	{HandlingActivity{Type: Load, Location: AUMEL, VoyageNumber: "001A"}, true},
	{HandlingActivity{Type: Load, Location: CNHKG, VoyageNumber: "001A"}, false},
	{HandlingActivity{Type: Unload, Location: CNHKG, VoyageNumber: "001A"}, true},
	{HandlingActivity{Type: Unload, Location: SESTO, VoyageNumber: "001A"}, false},
	{HandlingActivity{Type: Claim, Location: CNHKG}, true},
	{HandlingActivity{Type: Claim, Location: SESTO}, false},
}

func TestItinerary_IsExpected(t *testing.T) {
	i := Itinerary{Legs: []Leg{
		{
			VoyageNumber:   "001A",
			LoadLocation:   SESTO,
			UnloadLocation: AUMEL,
		},
		{
			VoyageNumber:   "001A",
			LoadLocation:   AUMEL,
			UnloadLocation: CNHKG,
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
