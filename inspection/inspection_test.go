package inspection

import (
	"testing"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/repository"
	"github.com/marcusolsson/goddd/voyage"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

type stubEventHandler struct {
	handledEvents []interface{}
}

func (h *stubEventHandler) CargoWasMisdirected(c cargo.Cargo) {
	h.handledEvents = append(h.handledEvents, c)
}

func (h *stubEventHandler) CargoHasArrived(c cargo.Cargo) {
	h.handledEvents = append(h.handledEvents, c)
}

func (s *S) TestInspectMisdirectedCargo(c *C) {

	var (
		cargoRepository         = repository.NewCargo()
		handlingEventRepository = repository.NewHandlingEvent()
		cargoEventHandler       = &stubEventHandler{make([]interface{}, 0)}

		inspectionService = &service{
			cargoRepository:         cargoRepository,
			handlingEventRepository: handlingEventRepository,
			cargoEventHandler:       cargoEventHandler,
		}
	)

	trackingID := cargo.TrackingID("ABC123")
	misdirectedCargo := cargo.New(trackingID, cargo.RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	})

	var voyageNumber voyage.Number = "001A"

	misdirectedCargo.AssignToRoute(cargo.Itinerary{Legs: []cargo.Leg{
		{VoyageNumber: voyageNumber, LoadLocation: location.SESTO, UnloadLocation: location.AUMEL},
		{VoyageNumber: voyageNumber, LoadLocation: location.AUMEL, UnloadLocation: location.CNHKG},
	}})

	cargoRepository.Store(*misdirectedCargo)

	storeEvent(handlingEventRepository, trackingID, voyageNumber, cargo.Receive, location.SESTO)
	storeEvent(handlingEventRepository, trackingID, voyageNumber, cargo.Load, location.SESTO)
	storeEvent(handlingEventRepository, trackingID, voyageNumber, cargo.Unload, location.USNYC)

	c.Check(len(cargoEventHandler.handledEvents), Equals, 0)

	inspectionService.InspectCargo(trackingID)

	c.Check(len(cargoEventHandler.handledEvents), Equals, 1)
}

func (s *S) TestInspectUnloadedCargo(c *C) {

	var (
		cargoRepository         = repository.NewCargo()
		handlingEventRepository = repository.NewHandlingEvent()
		cargoEventHandler       = &stubEventHandler{make([]interface{}, 0)}

		inspectionService = &service{
			cargoRepository:         cargoRepository,
			handlingEventRepository: handlingEventRepository,
			cargoEventHandler:       cargoEventHandler,
		}
	)

	trackingID := cargo.TrackingID("ABC123")
	unloadedCargo := cargo.New(trackingID, cargo.RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	})

	var voyageNumber voyage.Number = "001A"

	unloadedCargo.AssignToRoute(cargo.Itinerary{Legs: []cargo.Leg{
		{VoyageNumber: voyageNumber, LoadLocation: location.SESTO, UnloadLocation: location.AUMEL},
		{VoyageNumber: voyageNumber, LoadLocation: location.AUMEL, UnloadLocation: location.CNHKG},
	}})

	cargoRepository.Store(*unloadedCargo)

	storeEvent(handlingEventRepository, trackingID, voyageNumber, cargo.Receive, location.SESTO)
	storeEvent(handlingEventRepository, trackingID, voyageNumber, cargo.Load, location.SESTO)
	storeEvent(handlingEventRepository, trackingID, voyageNumber, cargo.Unload, location.AUMEL)
	storeEvent(handlingEventRepository, trackingID, voyageNumber, cargo.Load, location.AUMEL)
	storeEvent(handlingEventRepository, trackingID, voyageNumber, cargo.Unload, location.CNHKG)

	c.Check(len(cargoEventHandler.handledEvents), Equals, 0)

	inspectionService.InspectCargo(trackingID)

	c.Check(len(cargoEventHandler.handledEvents), Equals, 1)
}

func storeEvent(repository cargo.HandlingEventRepository, trackingID cargo.TrackingID, voyage voyage.Number, activityType cargo.HandlingEventType, location location.UNLocode) {
	repository.Store(cargo.HandlingEvent{
		TrackingID: trackingID,
		Activity:   cargo.HandlingActivity{VoyageNumber: voyage, Type: activityType, Location: location},
	})
}
