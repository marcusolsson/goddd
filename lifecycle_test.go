package main

import (
	"testing"
	"time"

	"github.com/marcusolsson/goddd/application"
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/voyage"
	"github.com/marcusolsson/goddd/infrastructure"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

var (
	cargoRepository         = infrastructure.NewInMemCargoRepository()
	locationRepository      = infrastructure.NewInMemLocationRepository()
	voyageRepository        = infrastructure.NewInMemVoyageRepository()
	handlingEventRepository = infrastructure.NewInMemHandlingEventRepository()
)

var (
	handlingEventFactory = cargo.HandlingEventFactory{cargoRepository, voyageRepository, locationRepository}
	handlingEventHandler = &stubHandlingEventHandler{cargoInspectionService}
	cargoEventHandler    = &stubCargoEventHandler{}
)

var (
	routingService         = &stubRoutingService{}
	bookingService         = application.NewBookingService(cargoRepository, locationRepository, routingService)
	cargoInspectionService = application.NewCargoInspectionService(cargoRepository, handlingEventRepository, cargoEventHandler)
	handlingEventService   = application.NewHandlingEventService(handlingEventRepository, handlingEventFactory, handlingEventHandler)
)

func (s *S) TestCargoFromHongkongToStockholm(chk *C) {

	var err error

	var (
		origin          = location.CNHKG // Hongkong
		destination     = location.SESTO // Stockholm
		arrivalDeadline = time.Date(2009, time.March, 18, 12, 00, 00, 00, time.UTC)
	)

	//
	// Use case 1: booking
	//

	trackingId, err := bookingService.BookNewCargo(origin, destination, arrivalDeadline)

	chk.Assert(err, IsNil)

	c, err := cargoRepository.Find(trackingId)

	chk.Assert(err, IsNil)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.NotReceived)
	chk.Check(c.Delivery.RoutingStatus, Equals, cargo.NotRouted)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, true)
	//chk.Check(c.Delivery.ETA, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	//
	// Use case 2: routing
	//

	itineraries := bookingService.RequestPossibleRoutesForCargo(trackingId)
	itinerary := selectPreferredItinerary(itineraries)

	c.AssignToRoute(itinerary)

	cargoRepository.Store(c)

	chk.Check(c.TransportStatus, Equals, cargo.NotReceived)
	chk.Check(c.RoutingStatus, Equals, cargo.Routed)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)
	//chk.Check(c.Delivery.ETA, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	//
	// Use case 3: handling
	//

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 1), trackingId, "", location.CNHKG, cargo.Receive)
	chk.Check(err, IsNil)

	// Ensure we're not working with stale cargo.
	c, err = cargoRepository.Find(trackingId)

	chk.Check(c.Delivery.TransportStatus, Equals, cargo.InPort)
	chk.Check(c.Delivery.LastKnownLocation, Equals, location.CNHKG)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 3), trackingId, voyage.V100.VoyageNumber, location.CNHKG, cargo.Load)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(trackingId)

	chk.Check(c.Delivery.TransportStatus, Equals, cargo.OnboardCarrier)
	chk.Check(c.Delivery.LastKnownLocation, Equals, location.CNHKG)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)
	//chk.Check(c.Delivery.CurrentVoyage, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	//
	// Here's an attempt to register a handling event that's not valid
	// because there is no voyage with the specified voyage number,
	// and there's no location with the specified UN Locode either.
	//

	noSuchVoyageNumber := voyage.VoyageNumber("XX000")
	noSuchUNLocode := location.UNLocode("ZZZZZ")
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 5), trackingId, noSuchVoyageNumber, noSuchUNLocode, cargo.Load)
	chk.Check(err, NotNil)

	//
	// Cargo is incorrectly unloaded in Tokyo
	//

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 5), trackingId, voyage.V100.VoyageNumber, location.JNTKO, cargo.Unload)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(trackingId)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.JNTKO)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.InPort)
	chk.Check(c.Delivery.Itinerary.IsEmpty(), Equals, false)
	//chk.Check(c.Delivery.CurrentVoyage, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	// Cargo is now misdirected
	chk.Check(c.Delivery.IsMisdirected, Equals, true)

	//
	// Cargo needs to be rerouted
	//

	routeSpecification := cargo.RouteSpecification{
		Origin:          location.JNTKO,
		Destination:     location.SESTO,
		ArrivalDeadline: arrivalDeadline,
	}

	// Specify a new route, this time from Tokyo (where it was incorrectly unloaded) to Stockholm
	c.SpecifyNewRoute(routeSpecification)

	cargoRepository.Store(c)

	chk.Check(c.RoutingStatus, Equals, cargo.Misrouted)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	// Repeat procedure of selecting one out of a number of possible routes satisfying the route spec
	newItineraries := bookingService.RequestPossibleRoutesForCargo(trackingId)
	newItinerary := selectPreferredItinerary(newItineraries)

	c.AssignToRoute(newItinerary)

	cargoRepository.Store(c)

	chk.Check(c.Delivery.RoutingStatus, Equals, cargo.Routed)

	//
	// Cargo has been rerouted, shipping continues
	//

	// Load in Tokyo
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 8), trackingId, voyage.V300.VoyageNumber, location.JNTKO, cargo.Load)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(trackingId)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.JNTKO)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.OnboardCarrier)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	//chk.Check(c.Delivery.CurrentVoyage, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	// Unload in Hamburg
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 12), trackingId, voyage.V300.VoyageNumber, location.DEHAM, cargo.Unload)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(trackingId)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.DEHAM)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.InPort)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	//chk.Check(c.Delivery.CurrentVoyage, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	// Load in Hamburg
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 14), trackingId, voyage.V400.VoyageNumber, location.DEHAM, cargo.Load)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(trackingId)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.DEHAM)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.OnboardCarrier)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	//chk.Check(c.Delivery.CurrentVoyage, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	// Unload in Stockholm
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 15), trackingId, voyage.V400.VoyageNumber, location.SESTO, cargo.Unload)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(trackingId)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.SESTO)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.InPort)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	//chk.Check(c.Delivery.CurrentVoyage, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	// Finally, cargo is claimed in Stockholm. This ends the cargo lifecycle from our perspective.
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 16), trackingId, voyage.V400.VoyageNumber, location.SESTO, cargo.Claim)
	chk.Check(err, IsNil)

	c, err = cargoRepository.Find(trackingId)

	chk.Check(c.Delivery.LastKnownLocation, Equals, location.SESTO)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.Claimed)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	//chk.Check(c.Delivery.CurrentVoyage, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)
}

func selectPreferredItinerary(itineraries []cargo.Itinerary) cargo.Itinerary {
	return itineraries[0]
}

func toDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 12, 00, 00, 00, time.UTC)
}

// Stub RoutingService
type stubRoutingService struct{}

func (s *stubRoutingService) FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary {
	if routeSpecification.Origin == location.CNHKG {
		return []cargo.Itinerary{
			cargo.Itinerary{[]cargo.Leg{
				cargo.Leg{"V100", location.CNHKG, location.USNYC, toDate(2009, time.March, 3), toDate(2009, time.March, 9)},
				cargo.Leg{"V200", location.USNYC, location.USCHI, toDate(2009, time.March, 10), toDate(2009, time.March, 14)},
				cargo.Leg{"V300", location.USCHI, location.SESTO, toDate(2009, time.March, 7), toDate(2009, time.March, 11)},
			}},
		}
	} else {
		return []cargo.Itinerary{
			cargo.Itinerary{[]cargo.Leg{
				cargo.Leg{"V300", location.JNTKO, location.DEHAM, toDate(2009, time.March, 8), toDate(2009, time.March, 12)},
				cargo.Leg{"V400", location.DEHAM, location.SESTO, toDate(2009, time.March, 14), toDate(2009, time.March, 15)},
			}},
		}
	}

}

// Stub HandlingEventHandler
type stubHandlingEventHandler struct {
	application.CargoInspectionService
}

func (h *stubHandlingEventHandler) CargoWasHandled(event cargo.HandlingEvent) {
	h.CargoInspectionService.InspectCargo(event.TrackingId)
}

// Stub CargoEventHandler
type stubCargoEventHandler struct {
}

func (h *stubCargoEventHandler) CargoWasMisdirected(c cargo.Cargo) {
}

func (h *stubCargoEventHandler) CargoHasArrived(c cargo.Cargo) {
}
