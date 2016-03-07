package booking

import (
	"testing"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/repository"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestBookNewCargo(c *C) {

	var (
		cargoRepository    = repository.NewCargo()
		locationRepository = repository.NewLocation()
		handlingEvents     = repository.NewHandlingEvent()
	)

	var bookingService = NewService(cargoRepository, locationRepository, handlingEvents, nil)

	origin, destination := location.SESTO, location.AUMEL
	arrivalDeadline := time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)

	trackingID, err := bookingService.BookNewCargo(origin, destination, arrivalDeadline)

	c.Assert(err, IsNil)

	cargo, err := cargoRepository.Find(trackingID)

	c.Assert(err, IsNil)

	c.Check(trackingID, Equals, cargo.TrackingID)
	c.Check(location.SESTO, Equals, cargo.Origin)
	c.Check(location.AUMEL, Equals, cargo.RouteSpecification.Destination)
	c.Check(arrivalDeadline, Equals, cargo.RouteSpecification.ArrivalDeadline)
}

type stubRoutingService struct{}

func (s *stubRoutingService) FetchRoutesForSpecification(rs cargo.RouteSpecification) []cargo.Itinerary {

	legs := []cargo.Leg{
		{
			LoadLocation:   rs.Origin,
			UnloadLocation: rs.Destination,
		},
	}

	return []cargo.Itinerary{
		{Legs: legs},
	}
}

func (s *S) TestRequestPossibleRoutesForCargo(c *C) {

	var (
		cargoRepository    = repository.NewCargo()
		locationRepository = repository.NewLocation()
		handlingEvents     = repository.NewHandlingEvent()
		routingService     = &stubRoutingService{}
	)

	var bookingService = NewService(cargoRepository, locationRepository, handlingEvents, routingService)

	origin, destination := location.Stockholm, location.Melbourne
	arrivalDeadline := time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)

	c.Check(bookingService.RequestPossibleRoutesForCargo("no_such_id"), Not(IsNil))

	trackingID, err := bookingService.BookNewCargo(origin.UNLocode, destination.UNLocode, arrivalDeadline)

	c.Assert(err, IsNil)

	itineraries := bookingService.RequestPossibleRoutesForCargo(trackingID)

	c.Check(itineraries, HasLen, 1)
}

func (s *S) TestAssignCargoToRoute(c *C) {
	var (
		cargoRepository    = repository.NewCargo()
		locationRepository = repository.NewLocation()
		handlingEvents     = repository.NewHandlingEvent()
		routingService     = &stubRoutingService{}
	)

	var bookingService = NewService(cargoRepository, locationRepository, handlingEvents, routingService)

	origin, destination := location.Stockholm, location.Melbourne
	arrivalDeadline := time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)

	trackingID, err := bookingService.BookNewCargo(origin.UNLocode, destination.UNLocode, arrivalDeadline)

	c.Assert(err, IsNil)

	itineraries := bookingService.RequestPossibleRoutesForCargo(trackingID)

	c.Assert(itineraries, HasLen, 1)

	err = bookingService.AssignCargoToRoute(trackingID, itineraries[0])

	c.Assert(err, IsNil)

	c.Assert(bookingService.AssignCargoToRoute("no_such_id", cargo.Itinerary{}), Not(IsNil))
}

func (s *S) TestChangeCargoDestination(c *C) {
	var (
		cargoRepository    = repository.NewCargo()
		locationRepository = repository.NewLocation()
		handlingEvents     = repository.NewHandlingEvent()
		routingService     = &stubRoutingService{}
	)

	var bookingService = NewService(cargoRepository, locationRepository, handlingEvents, routingService)

	c1 := cargo.New("ABC", cargo.RouteSpecification{
		Origin:          location.Stockholm.UNLocode,
		Destination:     location.Hongkong.UNLocode,
		ArrivalDeadline: time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC),
	})

	c.Check(bookingService.ChangeDestination("no_such_id", location.Stockholm.UNLocode), Not(IsNil))

	var err error

	err = cargoRepository.Store(c1)
	c.Assert(err, IsNil)

	c.Check(bookingService.ChangeDestination(c1.TrackingID, "no_such_unlocode"), Not(IsNil))

	c.Assert(c1.RouteSpecification.Destination, Equals, location.CNHKG)

	err = bookingService.ChangeDestination(c1.TrackingID, location.AUMEL)
	c.Assert(err, IsNil)

	updatedCargo, err := cargoRepository.Find(c1.TrackingID)

	c.Assert(err, IsNil)
	c.Assert(updatedCargo.RouteSpecification.Destination, Equals, location.AUMEL)
}
