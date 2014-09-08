package booking

import (
	"testing"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"

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
