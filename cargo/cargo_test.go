package cargo

import (
	"testing"
	"time"

	"bitbucket.org/marcus_olsson/goddd/location"

	. "gopkg.in/check.v1"
)

// Hook gocheck up to the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestConstruction(c *C) {
	trackingId := TrackingId("XYZ")
	specification := RouteSpecification{
		Origin:          location.Stockholm,
		Destination:     location.Melbourne,
		ArrivalDeadline: time.Date(2009, time.March, 13, 0, 0, 0, 0, time.UTC),
	}

	NewCargo(trackingId, specification)
}

func (s *S) TestEquality(c *C) {
	spec1 := RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	}
	spec2 := RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Melbourne,
	}

	c1 := NewCargo("ABC", spec1)
	c2 := NewCargo("CBA", spec1)
	c3 := NewCargo("ABC", spec2)
	c4 := NewCargo("ABC", spec1)

	c.Check(true, Equals, c1.Equal(c4))
	c.Check(true, Equals, c1.Equal(c3))
	c.Check(true, Equals, c3.Equal(c4))
	c.Check(false, Equals, c1.Equal(c2))
}

func (s *S) TestRepositoryFind(c *C) {
	c1 := NewCargo("ABC", RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	})

	r := NewCargoRepository()
	r.Store(*c1)

	c2, err := r.Find(TrackingId("ABC"))

	c.Assert(err, IsNil)
	c.Check(true, Equals, c1.Equal(c2))
}

func (s *S) TestRepositoryFindAll(c *C) {
	c1 := NewCargo("ABC", RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	})

	c2 := NewCargo("CBA", RouteSpecification{
		Origin:      location.Hongkong,
		Destination: location.Stockholm,
	})

	r := NewCargoRepository()
	r.Store(*c1)
	r.Store(*c2)

	cargos, err := r.FindAll()

	c.Assert(err, IsNil)
	c.Check(cargos, HasLen, 2)
}
