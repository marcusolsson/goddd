package main

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"

	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/inmem"
	"github.com/marcusolsson/goddd/inspection"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestCargoFromHongkongToStockholm(chk *C) {
	var err error

	var (
		cargoRepository         = inmem.NewCargoRepository()
		locationRepository      = inmem.NewLocationRepository()
		voyageRepository        = inmem.NewVoyageRepository()
		handlingEventRepository = inmem.NewHandlingEventRepository()
	)

	handlingEventFactory := cargo.HandlingEventFactory{
		CargoRepository:    cargoRepository,
		VoyageRepository:   voyageRepository,
		LocationRepository: locationRepository,
	}

	routingService := &stubRoutingService{}

	cargoEventHandler := &stubCargoEventHandler{}
	cargoInspectionService := inspection.NewService(cargoRepository, handlingEventRepository, cargoEventHandler)
	handlingEventHandler := &stubHandlingEventHandler{cargoInspectionService}

	var (
		bookingService       = booking.NewService(cargoRepository, locationRepository, handlingEventRepository, routingService)
		handlingEventService = handling.NewService(handlingEventRepository, handlingEventFactory, handlingEventHandler)
	)

	var (
		origin      = location.CNHKG // Hongkong
		destination = location.SESTO // Stockholm
		deadline    = time.Date(2009, time.March, 18, 12, 00, 00, 00, time.UTC)
	)

	//
	// Use case 1: booking
	//

	id, err := bookingService.BookNewCargo(origin, destination, deadline)

	chk.Assert(err, IsNil)

	c, err := cargoRepository.Find(id)

	chk.Assert(err, IsNil)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.NotReceived)
	chk.Check(c.Delivery.RoutingStatus, Equals, cargo.NotRouted)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, true)
	chk.Check(c.Delivery.ETA, Equals, time.Time{})
	chk.Check(c.Delivery.NextExpectedActivity, Equals, cargo.HandlingActivity{})

	//
	// Use case 2: routing
	//

	itineraries := bookingService.RequestPossibleRoutesForCargo(id)
	itinerary := selectPreferredItinerary(itineraries)

	c.AssignToRoute(itinerary)

	cargoRepository.Store(c)

	chk.Check(c.Delivery.TransportStatus, Equals, cargo.NotReceived)
	chk.Check(c.Delivery.RoutingStatus, Equals, cargo.Routed)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)
	chk.Check(c.Delivery.ETA, Not(Equals), time.Time{})
	chk.Check(c.Delivery.NextExpectedActivity, Equals, cargo.HandlingActivity{Type: cargo.Receive, Location: location.CNHKG})

	//
	// Use case 3: handling
	//

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 1), id, "", location.CNHKG, cargo.Receive)
	chk.Check(err, IsNil)

	// Ensure we're not working with stale cargo.
	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.TransportStatus, Equals, cargo.InPort)
	chk.Check(c.Delivery.LastKnownLocation, Equals, location.CNHKG)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 3), id, voyage.V100.Number, location.CNHKG, cargo.Load)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.TransportStatus, Equals, cargo.OnboardCarrier)
	chk.Check(c.Delivery.LastKnownLocation, Equals, location.CNHKG)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, voyage.V100.Number)
	chk.Check(c.Delivery.NextExpectedActivity, Equals, cargo.HandlingActivity{Type: cargo.Unload, Location: location.USNYC, VoyageNumber: voyage.V100.Number})

	//
	// Here's an attempt to register a handling event that's not valid
	// because there is no voyage with the specified voyage number,
	// and there's no location with the specified UN Locode either.
	//

	noSuchVoyageNumber := voyage.Number("XX000")
	noSuchUNLocode := location.UNLocode("ZZZZZ")
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 5), id, noSuchVoyageNumber, noSuchUNLocode, cargo.Load)
	chk.Check(err, NotNil)

	//
	// Cargo is incorrectly unloaded in Tokyo
	//

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 5), id, voyage.V100.Number, location.JNTKO, cargo.Unload)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.JNTKO)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.InPort)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, voyage.Number(""))
	chk.Check(c.Delivery.NextExpectedActivity, Equals, cargo.HandlingActivity{})

	// Cargo is now misdirected
	chk.Check(c.Delivery.IsMisdirected, Equals, true)

	//
	// Cargo needs to be rerouted
	//

	rs := cargo.RouteSpecification{
		Origin:          location.JNTKO,
		Destination:     location.SESTO,
		ArrivalDeadline: deadline,
	}

	// Specify a new route, this time from Tokyo (where it was incorrectly unloaded) to Stockholm
	c.SpecifyNewRoute(rs)

	cargoRepository.Store(c)

	chk.Check(c.Delivery.RoutingStatus, Equals, cargo.Misrouted)
	chk.Check(c.Delivery.NextExpectedActivity, Equals, cargo.HandlingActivity{})

	// Repeat procedure of selecting one out of a number of possible routes satisfying the route spec
	newItineraries := bookingService.RequestPossibleRoutesForCargo(id)
	newItinerary := selectPreferredItinerary(newItineraries)

	c.AssignToRoute(newItinerary)

	cargoRepository.Store(c)

	chk.Check(c.Delivery.RoutingStatus, Equals, cargo.Routed)

	//
	// Cargo has been rerouted, shipping continues
	//

	// Load in Tokyo
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 8), id, voyage.V300.Number, location.JNTKO, cargo.Load)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.JNTKO)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.OnboardCarrier)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, voyage.V300.Number)
	chk.Check(c.Delivery.NextExpectedActivity, Equals, cargo.HandlingActivity{Type: cargo.Unload, Location: location.DEHAM, VoyageNumber: voyage.V300.Number})

	// Unload in Hamburg
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 12), id, voyage.V300.Number, location.DEHAM, cargo.Unload)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.DEHAM)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.InPort)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, voyage.Number(""))
	chk.Check(c.Delivery.NextExpectedActivity, Equals, cargo.HandlingActivity{Type: cargo.Load, Location: location.DEHAM, VoyageNumber: voyage.V400.Number})

	// Load in Hamburg
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 14), id, voyage.V400.Number, location.DEHAM, cargo.Load)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.DEHAM)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.OnboardCarrier)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, voyage.V400.Number)
	chk.Check(c.Delivery.NextExpectedActivity, Equals, cargo.HandlingActivity{Type: cargo.Unload, Location: location.SESTO, VoyageNumber: voyage.V400.Number})

	// Unload in Stockholm
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 15), id, voyage.V400.Number, location.SESTO, cargo.Unload)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.SESTO)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.InPort)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, voyage.Number(""))
	chk.Check(c.Delivery.NextExpectedActivity, Equals, cargo.HandlingActivity{Type: cargo.Claim, Location: location.SESTO})

	// Finally, cargo is claimed in Stockholm. This ends the cargo lifecycle from our perspective.
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 16), id, voyage.V400.Number, location.SESTO, cargo.Claim)
	chk.Check(err, IsNil)

	c, _ = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.SESTO)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.Claimed)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, voyage.Number(""))
	chk.Check(c.Delivery.NextExpectedActivity, Equals, cargo.HandlingActivity{})
}

func selectPreferredItinerary(itineraries []cargo.Itinerary) cargo.Itinerary {
	return itineraries[0]
}

func toDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 12, 00, 00, 00, time.UTC)
}

// Stub RoutingService
type stubRoutingService struct{}

func (s *stubRoutingService) FetchRoutesForSpecification(rs cargo.RouteSpecification) []cargo.Itinerary {
	if rs.Origin == location.CNHKG {
		return []cargo.Itinerary{
			{Legs: []cargo.Leg{
				cargo.NewLeg("V100", location.CNHKG, location.USNYC, toDate(2009, time.March, 3), toDate(2009, time.March, 9)),
				cargo.NewLeg("V200", location.USNYC, location.USCHI, toDate(2009, time.March, 10), toDate(2009, time.March, 14)),
				cargo.NewLeg("V300", location.USCHI, location.SESTO, toDate(2009, time.March, 7), toDate(2009, time.March, 11)),
			}},
		}
	}

	return []cargo.Itinerary{
		{Legs: []cargo.Leg{
			cargo.NewLeg("V300", location.JNTKO, location.DEHAM, toDate(2009, time.March, 8), toDate(2009, time.March, 12)),
			cargo.NewLeg("V400", location.DEHAM, location.SESTO, toDate(2009, time.March, 14), toDate(2009, time.March, 15)),
		}},
	}
}

// Stub HandlingEventHandler
type stubHandlingEventHandler struct {
	InspectionService inspection.Service
}

func (h *stubHandlingEventHandler) CargoWasHandled(event cargo.HandlingEvent) {
	h.InspectionService.InspectCargo(event.TrackingID)
}

// Stub CargoEventHandler
type stubCargoEventHandler struct {
}

func (h *stubCargoEventHandler) CargoWasMisdirected(c *cargo.Cargo) {
}

func (h *stubCargoEventHandler) CargoHasArrived(c *cargo.Cargo) {
}
