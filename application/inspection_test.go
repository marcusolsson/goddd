package application

import (
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/voyage"
	"github.com/marcusolsson/goddd/infrastructure"
	. "gopkg.in/check.v1"
)

type stubCargoEventHandler struct {
	handledEvents []interface{}
}

func (h *stubCargoEventHandler) CargoWasMisdirected(c cargo.Cargo) {
	h.handledEvents = append(h.handledEvents, c)
}

func (h *stubCargoEventHandler) CargoHasArrived(c cargo.Cargo) {
	h.handledEvents = append(h.handledEvents, c)
}

func (s *S) TestInspectMisdirectedCargo(c *C) {

	var (
		cargoRepository         = infrastructure.NewInMemCargoRepository()
		handlingEventRepository = infrastructure.NewInMemHandlingEventRepository()
		cargoEventHandler       = &stubCargoEventHandler{make([]interface{}, 0)}

		inspectionService CargoInspectionService = &cargoInspectionService{
			cargoRepository:         cargoRepository,
			handlingEventRepository: handlingEventRepository,
			cargoEventHandler:       cargoEventHandler,
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

	storeEvent(handlingEventRepository, trackingId, voyageNumber, cargo.Receive, location.SESTO)
	storeEvent(handlingEventRepository, trackingId, voyageNumber, cargo.Load, location.SESTO)
	storeEvent(handlingEventRepository, trackingId, voyageNumber, cargo.Unload, location.USNYC)

	c.Check(len(cargoEventHandler.handledEvents), Equals, 0)

	inspectionService.InspectCargo(trackingId)

	c.Check(len(cargoEventHandler.handledEvents), Equals, 1)
}

func (s *S) TestInspectUnloadedCargo(c *C) {

	var (
		cargoRepository         = infrastructure.NewInMemCargoRepository()
		handlingEventRepository = infrastructure.NewInMemHandlingEventRepository()
		cargoEventHandler       = &stubCargoEventHandler{make([]interface{}, 0)}

		inspectionService CargoInspectionService = &cargoInspectionService{
			cargoRepository:         cargoRepository,
			handlingEventRepository: handlingEventRepository,
			cargoEventHandler:       cargoEventHandler,
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

	storeEvent(handlingEventRepository, trackingId, voyageNumber, cargo.Receive, location.SESTO)
	storeEvent(handlingEventRepository, trackingId, voyageNumber, cargo.Load, location.SESTO)
	storeEvent(handlingEventRepository, trackingId, voyageNumber, cargo.Unload, location.AUMEL)
	storeEvent(handlingEventRepository, trackingId, voyageNumber, cargo.Load, location.AUMEL)
	storeEvent(handlingEventRepository, trackingId, voyageNumber, cargo.Unload, location.CNHKG)

	c.Check(len(cargoEventHandler.handledEvents), Equals, 0)

	inspectionService.InspectCargo(trackingId)

	c.Check(len(cargoEventHandler.handledEvents), Equals, 1)
}

func storeEvent(repository cargo.HandlingEventRepository, trackingId cargo.TrackingId, voyage voyage.VoyageNumber, activityType cargo.HandlingEventType, location location.UNLocode) {
	repository.Store(cargo.HandlingEvent{
		TrackingId: trackingId,
		Activity:   cargo.HandlingActivity{VoyageNumber: voyage, Type: activityType, Location: location},
	})
}
