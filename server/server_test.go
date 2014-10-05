package server

import (
	"testing"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"

	. "gopkg.in/check.v1"
)

// Hook gocheck up to the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestAssembleDTO(c *C) {
	cargo := cargo.NewCargo("FTL456", cargo.RouteSpecification{
		Origin:      location.Melbourne,
		Destination: location.Stockholm,
	})

	dto := Assemble(*cargo)

	c.Check("FTL456", Equals, dto.TrackingId)
	c.Check("In port Melbourne", Equals, dto.StatusText)
	c.Check("SESTO", Equals, dto.Destination)
}
