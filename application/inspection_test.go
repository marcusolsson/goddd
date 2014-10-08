package application

import (
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/infrastructure"
	. "gopkg.in/check.v1"
)

func (s *S) TestInspectMisdirectedCargo(c *C) {

	var (
		cargoRepository = infrastructure.NewInMemCargoRepository()
		eventHandler    = &stubEventHandler{make([]interface{}, 0)}

		inspectionService CargoInspectionService = &cargoInspectionService{
			cargoRepository: cargoRepository,
			eventHandler:    eventHandler,
		}
	)

	trackingId := cargo.TrackingId("ABC123")
	misdirectedCargo := cargo.NewCargo(trackingId, cargo.RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	})

	cargoRepository.Store(*misdirectedCargo)

	inspectionService.InspectCargo(trackingId)

	c.Check(len(eventHandler.handledEvents), Equals, 1)
}
