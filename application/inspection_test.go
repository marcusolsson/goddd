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
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	})

	var voyageNumber voyage.VoyageNumber = "001A"

	misdirectedCargo.AssignToRoute(cargo.Itinerary{Legs: []cargo.Leg{
		cargo.Leg{VoyageNumber: voyageNumber, LoadLocation: location.SESTO, UnloadLocation: location.AUMEL},
		cargo.Leg{VoyageNumber: voyageNumber, LoadLocation: location.AUMEL, UnloadLocation: location.CNHKG},
	}})

	cargoRepository.Store(*misdirectedCargo)

	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Receive, Location: location.SESTO})
	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Load, Location: location.SESTO})
	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Unload, Location: location.USNYC})

	c.Check(len(eventHandler.handledEvents), Equals, 0)

	inspectionService.InspectCargo(trackingId)

	c.Check(len(eventHandler.handledEvents), Equals, 1)
}

func (s *S) TestInspectUnloadedCargo(c *C) {

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
	unloadedCargo := cargo.NewCargo(trackingId, cargo.RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	})

	var voyageNumber voyage.VoyageNumber = "001A"

	unloadedCargo.AssignToRoute(cargo.Itinerary{Legs: []cargo.Leg{
		cargo.Leg{VoyageNumber: voyageNumber, LoadLocation: location.SESTO, UnloadLocation: location.AUMEL},
		cargo.Leg{VoyageNumber: voyageNumber, LoadLocation: location.AUMEL, UnloadLocation: location.CNHKG},
	}})

	cargoRepository.Store(*unloadedCargo)

	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Receive, Location: location.SESTO})
	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Load, Location: location.SESTO})
	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Unload, Location: location.AUMEL})
	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Load, Location: location.AUMEL})
	handlingEventRepository.Store(cargo.HandlingEvent{TrackingId: trackingId, VoyageNumber: voyageNumber, Type: cargo.Unload, Location: location.CNHKG})

	c.Check(len(eventHandler.handledEvents), Equals, 0)

	inspectionService.InspectCargo(trackingId)

	c.Check(len(eventHandler.handledEvents), Equals, 1)
}
