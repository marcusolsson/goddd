package main

import (
	"testing"
	"time"

	"github.com/marcusolsson/goddd/application"
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/infrastructure"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

var (
	cargoRepository    = infrastructure.NewInMemCargoRepository()
	locationRepository = infrastructure.NewInMemLocationRepository()
	routingService     = infrastructure.NewExternalRoutingService(locationRepository)
	bookingService     = application.NewBookingService(cargoRepository, locationRepository, routingService)
)

var (
	handlingEventRepository = infrastructure.NewInMemHandlingEventRepository()
	handlingEventFactory    = cargo.HandlingEventFactory{cargoRepository}
	applicationEvents       = &stubEventHandler{}
	handlingEventService    = application.NewHandlingEventService(handlingEventRepository, handlingEventFactory, applicationEvents)
)

func (s *S) TestCargoFromHongkongToStockholm(chk *C) {

	var err error

	var (
		origin          = location.CNHKG // Hongkong
		destination     = location.SESTO // Stockholm
		arrivalDeadline = time.Date(2009, time.March, 18, 12, 00, 00, 00, time.UTC)
	)

	// Use case 1: booking
	trackingId, err := bookingService.BookNewCargo(origin, destination, arrivalDeadline)

	chk.Assert(err, IsNil)

	c, err := cargoRepository.Find(trackingId)

	chk.Assert(err, IsNil)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.NotReceived)
	chk.Check(c.Delivery.RoutingStatus, Equals, cargo.NotRouted)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	//chk.Check(c.Delivery.ETA, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	// Use case 2: routing
	itineraries := bookingService.RequestPossibleRoutesForCargo(trackingId)
	itinerary := selectPreferredItinerary(itineraries)

	c.AssignToRoute(itinerary)

	chk.Check(c.TransportStatus, Equals, cargo.NotReceived)
	chk.Check(c.RoutingStatus, Equals, cargo.Routed)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	//chk.Check(c.Delivery.ETA, Equals, ...)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)

	// Use case 3: handling
	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 1), trackingId, "", location.CNHKG, cargo.Receive)

	chk.Check(c.Delivery.TransportStatus, Equals, cargo.InPort)
	chk.Check(c.Delivery.LastKnownLocation, Equals, location.CNHKG)

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 3), trackingId, "V100", location.CNHKG, cargo.Load)

	//chk.Check(c.Delivery.CurrentVoyage, Equals, ...)
	chk.Check(c.Delivery.LastKnownLocation, Equals, location.CNHKG)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.OnboardCarrier)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)
}

func selectPreferredItinerary(itineraries []cargo.Itinerary) cargo.Itinerary {
	return cargo.Itinerary{}
}

func toDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 12, 00, 00, 00, time.UTC)
}

type stubEventHandler struct {
	handledEvents []interface{}
}

func (h *stubEventHandler) CargoWasHandled(event cargo.HandlingEvent) {
	h.handledEvents = append(h.handledEvents, event)
}

func (h *stubEventHandler) CargoWasMisdirected(c cargo.Cargo) {
	h.handledEvents = append(h.handledEvents, c)
}

func (h *stubEventHandler) CargoHasArrived(c cargo.Cargo) {
	h.handledEvents = append(h.handledEvents, c)
}
