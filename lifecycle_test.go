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
	routingService     = &stubRoutingService{}
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

	err = handlingEventService.RegisterHandlingEvent(toDate(2009, time.March, 3), trackingId, "v100", location.CNHKG, cargo.Load)

	//chk.Check(c.Delivery.CurrentVoyage, Equals, ...)
	chk.Check(c.Delivery.LastKnownLocation, Equals, location.CNHKG)
	chk.Check(c.Delivery.TransportStatus, Equals, cargo.OnboardCarrier)
	chk.Check(c.Delivery.IsMisdirected, Equals, false)
	//chk.Check(c.Delivery.NextExpectedActivity, Equals, ...)
}

func selectPreferredItinerary(itineraries []cargo.Itinerary) cargo.Itinerary {
	return itineraries[0]
}

func toDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 12, 00, 00, 00, time.UTC)
}

// Stub EventHandler
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

// Stub RoutingService
type stubRoutingService struct{}

func (s *stubRoutingService) FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary {
	if routeSpecification.Origin == location.CNHKG {
		return []cargo.Itinerary{
			cargo.Itinerary{[]cargo.Leg{
				cargo.Leg{"v100", location.CNHKG, location.USNYC, toDate(2009, time.March, 3), toDate(2009, time.March, 9)},
				cargo.Leg{"v200", location.USNYC, location.USCHI, toDate(2009, time.March, 10), toDate(2009, time.March, 14)},
				cargo.Leg{"v300", location.USCHI, location.SESTO, toDate(2009, time.March, 7), toDate(2009, time.March, 11)},
			}},
		}
	} else {
		return []cargo.Itinerary{
			cargo.Itinerary{[]cargo.Leg{
				cargo.Leg{"v300", location.JNTKO, location.DEHAM, toDate(2009, time.March, 8), toDate(2009, time.March, 12)},
				cargo.Leg{"v400", location.DEHAM, location.SESTO, toDate(2009, time.March, 14), toDate(2009, time.March, 15)},
			}},
		}
	}

}
