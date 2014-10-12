package application

import (
	"time"

	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/voyage"
	"github.com/marcusolsson/goddd/infrastructure"
	. "gopkg.in/check.v1"
)

type stubHandlingEventHandler struct {
	handledEvents []interface{}
}

func (h *stubHandlingEventHandler) CargoWasHandled(event cargo.HandlingEvent) {
	h.handledEvents = append(h.handledEvents, event)
}

func (s *S) TestRegisterHandlingEvent(c *C) {

	var (
		cargoRepository         = infrastructure.NewInMemCargoRepository()
		voyageRepository        = infrastructure.NewInMemVoyageRepository()
		locationRepository      = infrastructure.NewInMemLocationRepository()
		handlingEventRepository = infrastructure.NewInMemHandlingEventRepository()
	)

	var (
		handlingEventHandler = &stubHandlingEventHandler{make([]interface{}, 0)}
		handlingEventFactory = cargo.HandlingEventFactory{
			CargoRepository:    cargoRepository,
			VoyageRepository:   voyageRepository,
			LocationRepository: locationRepository,
		}
		handlingEventService = &handlingEventService{
			handlingEventRepository: handlingEventRepository,
			handlingEventFactory:    handlingEventFactory,
			handlingEventHandler:    handlingEventHandler,
		}
	)

	var (
		completionTime = time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)
		trackingId     = cargo.TrackingId("ABC123")
		voyageNumber   = voyage.VoyageNumber("V100")
		unLocode       = location.Stockholm.UNLocode
		eventType      = cargo.Load
	)

	cargoRepository.Store(*cargo.NewCargo(trackingId, cargo.RouteSpecification{}))

	err := service.RegisterHandlingEvent(completionTime, trackingId, voyageNumber, unLocode, eventType)

	c.Check(err, IsNil)
	c.Check(len(handlingEventHandler.handledEvents), Equals, 1)
}
