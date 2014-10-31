package cargo

import (
	"github.com/marcusolsson/goddd/domain/location"

	. "gopkg.in/check.v1"
)

func (s *S) TestCreateEmptyItinerary(c *C) {
	emptyItinerary := Itinerary{}

	var emptyLegs []Leg

	c.Check(emptyItinerary.Legs, DeepEquals, emptyLegs)
	c.Check(emptyItinerary.InitialDepartureLocation(), Equals, location.UNLocode(""))
	c.Check(emptyItinerary.FinalArrivalLocation(), Equals, location.UNLocode(""))
}

func (s *S) TestItineraryEquality(c *C) {

	i1 := Itinerary{Legs: []Leg{
		Leg{LoadLocation: location.SESTO, UnloadLocation: location.AUMEL},
		Leg{LoadLocation: location.AUMEL, UnloadLocation: location.CNHKG},
	}}

	i2 := Itinerary{Legs: []Leg{
		Leg{LoadLocation: location.SESTO, UnloadLocation: location.AUMEL},
		Leg{LoadLocation: location.AUMEL, UnloadLocation: location.CNHKG},
	}}

	i3 := Itinerary{Legs: []Leg{
		Leg{LoadLocation: location.CNHKG, UnloadLocation: location.AUMEL},
		Leg{LoadLocation: location.AUMEL, UnloadLocation: location.SESTO},
	}}

	c.Check(i1, DeepEquals, i1)
	c.Check(i1, DeepEquals, i2)
	c.Check(i1, Not(DeepEquals), i3)
	c.Check(i2, Not(DeepEquals), i3)
}

func (s *S) TestLegEquality(c *C) {
	l1 := Leg{LoadLocation: location.SESTO, UnloadLocation: location.AUMEL}
	l2 := Leg{LoadLocation: location.SESTO, UnloadLocation: location.AUMEL}
	l3 := Leg{LoadLocation: location.AUMEL, UnloadLocation: location.CNHKG}

	c.Check(l1, DeepEquals, l1)
	c.Check(l1, DeepEquals, l2)
	c.Check(l1, Not(DeepEquals), l3)
	c.Check(l2, Not(DeepEquals), l3)
}

func (s *S) TestItineraryIsExpected(c *C) {

	emptyItinerary := Itinerary{}
	emptyEvent := HandlingEvent{}

	c.Check(emptyItinerary.IsExpected(emptyEvent), Equals, true)

	i := Itinerary{Legs: []Leg{
		Leg{VoyageNumber: "001A", LoadLocation: location.SESTO, UnloadLocation: location.AUMEL},
		Leg{VoyageNumber: "001A", LoadLocation: location.AUMEL, UnloadLocation: location.CNHKG},
	}}

	c.Check(i.IsExpected(emptyEvent), Equals, true)

	var (
		receiveEvent              = HandlingEvent{Activity: HandlingActivity{Type: Receive, Location: location.SESTO}}
		receiveEventWrongLocation = HandlingEvent{Activity: HandlingActivity{Type: Receive, Location: location.AUMEL}}
	)
	c.Check(i.IsExpected(receiveEvent), Equals, true)
	c.Check(i.IsExpected(receiveEventWrongLocation), Equals, false)

	var (
		loadEvent              = HandlingEvent{Activity: HandlingActivity{VoyageNumber: "001A", Type: Load, Location: location.AUMEL}}
		loadEventWrongLocation = HandlingEvent{Activity: HandlingActivity{VoyageNumber: "001A", Type: Load, Location: location.CNHKG}}
	)
	c.Check(i.IsExpected(loadEvent), Equals, true)
	c.Check(i.IsExpected(loadEventWrongLocation), Equals, false)

	var (
		unloadEvent              = HandlingEvent{Activity: HandlingActivity{VoyageNumber: "001A", Type: Unload, Location: location.CNHKG}}
		unloadEventWrongLocation = HandlingEvent{Activity: HandlingActivity{VoyageNumber: "001A", Type: Unload, Location: location.SESTO}}
	)
	c.Check(i.IsExpected(unloadEvent), Equals, true)
	c.Check(i.IsExpected(unloadEventWrongLocation), Equals, false)

	var (
		claimEvent              = HandlingEvent{Activity: HandlingActivity{Type: Claim, Location: location.CNHKG}}
		claimEventWrongLocation = HandlingEvent{Activity: HandlingActivity{Type: Claim, Location: location.SESTO}}
	)
	c.Check(i.IsExpected(claimEvent), Equals, true)
	c.Check(i.IsExpected(claimEventWrongLocation), Equals, false)
}
