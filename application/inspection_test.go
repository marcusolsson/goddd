package application

import (
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/voyage"
	"github.com/marcusolsson/goddd/infrastructure"
	. "gopkg.in/check.v1"
)

func (s *S) TestInspectMisdirectedCargo(c *C) {

	var (
		cargoRepository         = infrastructure.NewInMemCargoRepository()
		handlingEventRepository = infrastructure.NewInMemHandlingEventRepository()
		eventHandler            = &stubEventHandler{make([]interface{}, 0)}

		inspectionService CargoInspectionService = &cargoInspectionService{
			cargoRepository:         cargoRepository,
			handlingEventRepository: handlingEventRepository,
			eventHandler:            eventHandler,
		}
	)

	trackingId := cargo.TrackingId("ABC123")
	misdirectedCargo := cargo.NewCargo(trackingId, cargo.RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	})

	var voyageNumber voyage.VoyageNumber = "001A"

	misdirectedCargo.AssignToRoute(cargo.Itinerary{Legs: []cargo.Leg{
		cargo.Leg{VoyageNumber: voyageNumber, LoadLocation: location.Stockholm, UnloadLocation: location.Melbourne},
		cargo.Leg{VoyageNumber: voyageNumber, LoadLocation: location.Melbourne, UnloadLocation: location.Hongkong},
	}})

	cargoRepository.Store(*misdirectedCargo)

	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Receive, Location: location.Stockholm})
	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Load, Location: location.Stockholm})
	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Unload, Location: location.NewYork})

	c.Check(len(eventHandler.handledEvents), Equals, 0)

	inspectionService.InspectCargo(trackingId)

	c.Check(len(eventHandler.handledEvents), Equals, 1)
}
