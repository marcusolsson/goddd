package booking

import (
	"testing"
	"time"

	"bitbucket.org/marcus_olsson/goddd/cargo"
	"bitbucket.org/marcus_olsson/goddd/location"

	. "gopkg.in/check.v1"
)

// Hook gocheck up to the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestBookNewCargo(c *C) {

	cargoRepository := cargo.NewCargoRepository()
	locationRepository := location.NewLocationRepository()

	bookingService := &bookingService{
		cargoRepository:    cargoRepository,
		locationRepository: locationRepository,
	}

	origin, destination := location.Stockholm, location.Melbourne
	arrivalDeadline := time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)

	trackingId, err := bookingService.BookNewCargo(origin.UNLocode, destination.UNLocode, arrivalDeadline)

	c.Assert(err, IsNil)

	cargo, err := cargoRepository.Find(trackingId)

	c.Assert(err, IsNil)

	c.Check(trackingId, Equals, cargo.TrackingId)
	c.Check(location.Stockholm, Equals, cargo.Origin)
	c.Check(location.Melbourne, Equals, cargo.RouteSpecification.Destination)
	c.Check(arrivalDeadline, Equals, cargo.RouteSpecification.ArrivalDeadline)
}

type stubRoutingService struct{}

// Create an itinerary with a single leg from origin to destination.
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

	cargoRepository := cargo.NewCargoRepository()
	locationRepository := location.NewLocationRepository()
	routingService := &stubRoutingService{}

	var bookingService BookingService = &bookingService{
		cargoRepository:    cargoRepository,
		locationRepository: locationRepository,
		routingService:     routingService,
	}

	origin, destination := location.Stockholm, location.Melbourne
	arrivalDeadline := time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)

	trackingId, err := bookingService.BookNewCargo(origin.UNLocode, destination.UNLocode, arrivalDeadline)

	c.Assert(err, IsNil)

	itineraries := bookingService.RequestPossibleRoutesForCargo(trackingId)

	c.Check(itineraries, HasLen, 1)
}

func (s *S) TestAssignCargoToRoute(c *C) {
	cargoRepository := cargo.NewCargoRepository()
	locationRepository := location.NewLocationRepository()
	routingService := &stubRoutingService{}

	var bookingService BookingService = &bookingService{
		cargoRepository:    cargoRepository,
		locationRepository: locationRepository,
		routingService:     routingService,
	}

	origin, destination := location.Stockholm, location.Melbourne
	arrivalDeadline := time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)

	trackingId, err := bookingService.BookNewCargo(origin.UNLocode, destination.UNLocode, arrivalDeadline)

	c.Assert(err, IsNil)

	itineraries := bookingService.RequestPossibleRoutesForCargo(trackingId)

	c.Assert(itineraries, HasLen, 1)

	err = bookingService.AssignCargoToRoute(itineraries[0], trackingId)

	c.Assert(err, IsNil)
}
