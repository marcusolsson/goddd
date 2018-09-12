package main

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/inmem"
	"github.com/marcusolsson/goddd/inspection"
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

	handlingEventFactory := shipping.HandlingEventFactory{
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
		origin      = shipping.CNHKG // Hongkong
		destination = shipping.SESTO // Stockholm
		deadline    = time.Date(2009, time.March, 18, 12, 00, 00, 00, time.UTC)
	)

	//
	// Use case 1: booking
	//

	id, err := bookingService.BookNewCargo(origin, destination, deadline)

	chk.Assert(err, IsNil)

	c, err := cargoRepository.Find(id)

	chk.Assert(err, IsNil)
	chk.Check(c.Delivery.TransportStatus, Equals, shipping.NotReceived)
	chk.Check(c.Delivery.RoutingStatus, Equals, shipping.NotRouted)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, true)
	chk.Check(c.Delivery.ETA, Equals, time.Time{})
	chk.Check(c.Delivery.NextExpectedActivity, Equals, shipping.HandlingActivity{})

	//
	// Use case 2: routing
	//

	itineraries := bookingService.RequestPossibleRoutesForCargo(id)
	itinerary := selectPreferredItinerary(itineraries)

	c.AssignToRoute(itinerary)

	cargoRepository.Store(c)

	chk.Check(c.Delivery.TransportStatus, Equals, shipping.NotReceived)
	chk.Check(c.Delivery.RoutingStatus, Equals, shipping.Routed)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)
	chk.Check(c.Delivery.ETA, Not(Equals), time.Time{})
	chk.Check(c.Delivery.NextExpectedActivity, Equals, shipping.HandlingActivity{Type: shipping.Receive, Location: shipping.CNHKG})

	//
	// Use case 3: handling
	//

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 1), id, "", shipping.CNHKG, shipping.Receive)
	chk.Check(err, IsNil)

	// Ensure we're not working with stale shipping.
	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.TransportStatus, Equals, shipping.InPort)
	chk.Check(c.Delivery.LastKnownLocation, Equals, shipping.CNHKG)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 3), id, shipping.V100.VoyageNumber, shipping.CNHKG, shipping.Load)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.TransportStatus, Equals, shipping.OnboardCarrier)
	chk.Check(c.Delivery.LastKnownLocation, Equals, shipping.CNHKG)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, shipping.V100.VoyageNumber)
	chk.Check(c.Delivery.NextExpectedActivity, Equals, shipping.HandlingActivity{Type: shipping.Unload, Location: shipping.USNYC, VoyageNumber: shipping.V100.VoyageNumber})

	//
	// Here's an attempt to register a handling event that's not valid
	// because there is no voyage with the specified voyage number,
	// and there's no location with the specified UN Locode either.
	//

	noSuchVoyageNumber := shipping.VoyageNumber("XX000")
	noSuchUNLocode := shipping.UNLocode("ZZZZZ")
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 5), id, noSuchVoyageNumber, noSuchUNLocode, shipping.Load)
	chk.Check(err, NotNil)

	//
	// Cargo is incorrectly unloaded in Tokyo
	//

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 5), id, shipping.V100.VoyageNumber, shipping.JNTKO, shipping.Unload)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, shipping.JNTKO)
	chk.Check(c.Delivery.TransportStatus, Equals, shipping.InPort)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, shipping.VoyageNumber(""))
	chk.Check(c.Delivery.NextExpectedActivity, Equals, shipping.HandlingActivity{})

	// Cargo is now misdirected
	chk.Check(c.Delivery.IsMisdirected, Equals, true)

	//
	// Cargo needs to be rerouted
	//

	rs := shipping.RouteSpecification{
		Origin:          shipping.JNTKO,
		Destination:     shipping.SESTO,
		ArrivalDeadline: deadline,
	}

	// Specify a new route, this time from Tokyo (where it was incorrectly unloaded) to Stockholm
	c.SpecifyNewRoute(rs)

	cargoRepository.Store(c)

	chk.Check(c.Delivery.RoutingStatus, Equals, shipping.Misrouted)
	chk.Check(c.Delivery.NextExpectedActivity, Equals, shipping.HandlingActivity{})

	// Repeat procedure of selecting one out of a number of possible routes satisfying the route spec
	newItineraries := bookingService.RequestPossibleRoutesForCargo(id)
	newItinerary := selectPreferredItinerary(newItineraries)

	c.AssignToRoute(newItinerary)

	cargoRepository.Store(c)

	chk.Check(c.Delivery.RoutingStatus, Equals, shipping.Routed)

	//
	// Cargo has been rerouted, shipping continues
	//

	// Load in Tokyo
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 8), id, shipping.V300.VoyageNumber, shipping.JNTKO, shipping.Load)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, shipping.JNTKO)
	chk.Check(c.Delivery.TransportStatus, Equals, shipping.OnboardCarrier)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, shipping.V300.VoyageNumber)
	chk.Check(c.Delivery.NextExpectedActivity, Equals, shipping.HandlingActivity{Type: shipping.Unload, Location: shipping.DEHAM, VoyageNumber: shipping.V300.VoyageNumber})

	// Unload in Hamburg
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 12), id, shipping.V300.VoyageNumber, shipping.DEHAM, shipping.Unload)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, shipping.DEHAM)
	chk.Check(c.Delivery.TransportStatus, Equals, shipping.InPort)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, shipping.VoyageNumber(""))
	chk.Check(c.Delivery.NextExpectedActivity, Equals, shipping.HandlingActivity{Type: shipping.Load, Location: shipping.DEHAM, VoyageNumber: shipping.V400.VoyageNumber})

	// Load in Hamburg
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 14), id, shipping.V400.VoyageNumber, shipping.DEHAM, shipping.Load)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, shipping.DEHAM)
	chk.Check(c.Delivery.TransportStatus, Equals, shipping.OnboardCarrier)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, shipping.V400.VoyageNumber)
	chk.Check(c.Delivery.NextExpectedActivity, Equals, shipping.HandlingActivity{Type: shipping.Unload, Location: shipping.SESTO, VoyageNumber: shipping.V400.VoyageNumber})

	// Unload in Stockholm
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 15), id, shipping.V400.VoyageNumber, shipping.SESTO, shipping.Unload)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, shipping.SESTO)
	chk.Check(c.Delivery.TransportStatus, Equals, shipping.InPort)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, shipping.VoyageNumber(""))
	chk.Check(c.Delivery.NextExpectedActivity, Equals, shipping.HandlingActivity{Type: shipping.Claim, Location: shipping.SESTO})

	// Finally, cargo is claimed in Stockholm. This ends the cargo lifecycle from our perspective.
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 16), id, shipping.V400.VoyageNumber, shipping.SESTO, shipping.Claim)
	chk.Check(err, IsNil)

	c, _ = cargoRepository.Find(id)

	chk.Check(c.Delivery.LastKnownLocation, Equals, shipping.SESTO)
	chk.Check(c.Delivery.TransportStatus, Equals, shipping.Claimed)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.CurrentVoyage, Equals, shipping.VoyageNumber(""))
	chk.Check(c.Delivery.NextExpectedActivity, Equals, shipping.HandlingActivity{})
}

func selectPreferredItinerary(itineraries []shipping.Itinerary) shipping.Itinerary {
	return itineraries[0]
}

func toDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 12, 00, 00, 00, time.UTC)
}

// Stub RoutingService
type stubRoutingService struct{}

func (s *stubRoutingService) FetchRoutesForSpecification(rs shipping.RouteSpecification) []shipping.Itinerary {
	if rs.Origin == shipping.CNHKG {
		return []shipping.Itinerary{
			{Legs: []shipping.Leg{
				shipping.MakeLeg("V100", shipping.CNHKG, shipping.USNYC, toDate(2009, time.March, 3), toDate(2009, time.March, 9)),
				shipping.MakeLeg("V200", shipping.USNYC, shipping.USCHI, toDate(2009, time.March, 10), toDate(2009, time.March, 14)),
				shipping.MakeLeg("V300", shipping.USCHI, shipping.SESTO, toDate(2009, time.March, 7), toDate(2009, time.March, 11)),
			}},
		}
	}

	return []shipping.Itinerary{
		{Legs: []shipping.Leg{
			shipping.MakeLeg("V300", shipping.JNTKO, shipping.DEHAM, toDate(2009, time.March, 8), toDate(2009, time.March, 12)),
			shipping.MakeLeg("V400", shipping.DEHAM, shipping.SESTO, toDate(2009, time.March, 14), toDate(2009, time.March, 15)),
		}},
	}
}

// Stub HandlingEventHandler
type stubHandlingEventHandler struct {
	InspectionService inspection.Service
}

func (h *stubHandlingEventHandler) CargoWasHandled(event shipping.HandlingEvent) {
	h.InspectionService.InspectCargo(event.TrackingID)
}

// Stub CargoEventHandler
type stubCargoEventHandler struct {
}

func (h *stubCargoEventHandler) CargoWasMisdirected(c *shipping.Cargo) {
}

func (h *stubCargoEventHandler) CargoHasArrived(c *shipping.Cargo) {
}
