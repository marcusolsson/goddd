package application

import (
	"testing"
	"time"

	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/infrastructure"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestBookNewCargo(c *C) {

	cargoRepository := infrastructure.NewInMemCargoRepository()
	locationRepository := infrastructure.NewInMemLocationRepository()

	bookingService := &bookingService{
		cargoRepository:    cargoRepository,
		locationRepository: locationRepository,
	}

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

func (s *stubRoutingService) FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary {

	legs := []cargo.Leg{
		cargo.Leg{
			LoadLocation:   routeSpecification.Origin,
			UnloadLocation: routeSpecification.Destination,
		},
	}

	return []cargo.Itinerary{
		cargo.Itinerary{Legs: legs},
	}
}

func (s *S) TestRequestPossibleRoutesForCargo(c *C) {

	cargoRepository := infrastructure.NewInMemCargoRepository()
	locationRepository := infrastructure.NewInMemLocationRepository()
	routingService := &stubRoutingService{}

	var bookingService BookingService = &bookingService{
		cargoRepository:    cargoRepository,
		locationRepository: locationRepository,
		routingService:     routingService,
	}

	origin, destination := location.Stockholm, location.Melbourne
	arrivalDeadline := time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)

	trackingID, err := bookingService.BookNewCargo(origin.UNLocode, destination.UNLocode, arrivalDeadline)

	c.Assert(err, IsNil)

	itineraries := bookingService.RequestPossibleRoutesForCargo(trackingID)

	c.Check(itineraries, HasLen, 1)
}

func (s *S) TestAssignCargoToRoute(c *C) {
	cargoRepository := infrastructure.NewInMemCargoRepository()
	locationRepository := infrastructure.NewInMemLocationRepository()
	routingService := &stubRoutingService{}

	var bookingService BookingService = &bookingService{
		cargoRepository:    cargoRepository,
		locationRepository: locationRepository,
		routingService:     routingService,
	}

	origin, destination := location.Stockholm, location.Melbourne
	arrivalDeadline := time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)

	trackingID, err := bookingService.BookNewCargo(origin.UNLocode, destination.UNLocode, arrivalDeadline)

	c.Assert(err, IsNil)

	itineraries := bookingService.RequestPossibleRoutesForCargo(trackingID)

	c.Assert(itineraries, HasLen, 1)

	err = bookingService.AssignCargoToRoute(itineraries[0], trackingID)

	c.Assert(err, IsNil)
}

func (s *S) TestChangeCargoDestination(c *C) {

	var (
		cargoRepository    = infrastructure.NewInMemCargoRepository()
		locationRepository = infrastructure.NewInMemLocationRepository()
		routingService     = &stubRoutingService{}
	)

	var bookingService BookingService = &bookingService{
		cargoRepository:    cargoRepository,
		locationRepository: locationRepository,
		routingService:     routingService,
	}

	c1 := cargo.NewCargo("ABC", cargo.RouteSpecification{
		Origin:          location.Stockholm.UNLocode,
		Destination:     location.Hongkong.UNLocode,
		ArrivalDeadline: time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC),
	})

	var err error

	err = cargoRepository.Store(*c1)
	c.Assert(err, IsNil)
	c.Assert(c1.RouteSpecification.Destination, Equals, location.CNHKG)

	err = bookingService.ChangeDestination(c1.TrackingID, location.AUMEL)
	c.Assert(err, IsNil)

	updatedCargo, err := cargoRepository.Find(c1.TrackingID)

	c.Assert(err, IsNil)
	c.Assert(updatedCargo.RouteSpecification.Destination, Equals, location.AUMEL)
}
